// Package task implements any task related operations
//
// This file contains the implementation of the remote runner tracker (RRT)
// RRT is used to manage the nodes that are allocated to a session
// This is an abstraction layer that manages individual node connections
// A task execution will just request a Node via GetNextNode() and use that node to execute the task
// RRT does the round-robin scheduling of the nodes.
// RRT also is responsible for interfacing with the resource manager to get and release the nodes
//
// IMPORTANT: 	RRT is not controll if a node is getting more tasks than it can handle
// 				The node itself is responsible for throttle the task execution if it is overloaded
//				This can be revisited in the future.
//
// By design, the round-robin is designed to load balance the tasks across the nodes
// If a user wants lots of tasks, than there should be more nodes allocated
// By contrast a larger client-engine will also be required to handle the load
//
// Author: Caio Cominato

package task

import (
	"client-engine/chunker"
	pb "client-engine/grpc_node_protocol"
	"client-engine/logger"
	"context"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	REQUEST_POOLING_INTERVAL   = 200 * time.Millisecond
	REQUEST_HEARTBEAT_INTERVAL = 30 * time.Second
)

type BlockInternalReference struct {
	BlockID         string
	BlockIdentifier string
}

type Node struct {
	nodeID      string
	NodeURI     string
	CallbackURI string
	CallbackID  string
	SessionID   string
	Blocks      []BlockInternalReference
	conn        *grpc.ClientConn
	client      pb.NodeClient
}

func (n *Node) connect() error {
	err := error(nil)
	n.conn, err = grpc.NewClient(n.NodeURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	n.client = pb.NewNodeClient(n.conn)

	// Register the callback
	err = n.RegisterCallback()
	if err != nil {
		return err
	}
	// We also need to make sure that the node can isolate the enviroment for the task
	err = n.InitIsolatedEnv()
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) disconnect() error {
	n.UnregisterCallback()
	n.DestroyIsolatedEnv()
	if n.conn != nil {
		err := n.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) InitIsolatedEnv() error {
	// We need to make sure that the node can isolate the enviroment for the task
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300) // Higher timeout because this can take a while
	defer cancel()
	res, err := n.client.IsolatedEnvInit(ctx, &pb.IsolatedEnv{IsolatedEnvID: n.SessionID})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New("Failed to initialize isolated environment [" + res.ErrorMessage + "]")
	}
	return nil
}

func (n *Node) DestroyIsolatedEnv() error {
	// We need to make sure that the node can isolate the enviroment for the task
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300) // Higher timeout because this can take a while
	defer cancel()
	res, err := n.client.IsolatedEnvDestroy(ctx, &pb.IsolatedEnv{IsolatedEnvID: n.SessionID})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New("Failed to destroy isolated environment [" + res.ErrorMessage + "]")
	}
	return nil
}

func (n *Node) RegisterCallback() error {
	callbackRegistration := pb.CallbackRegistration{
		CallbackURI: n.CallbackURI,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := n.client.RegisterCallback(ctx, &callbackRegistration)
	if err != nil {
		logger.Error("Failed to register callback", "NODE", err)
		return err
	}
	n.CallbackID = res.CallbackID
	return nil
}

func (n *Node) UnregisterCallback() error {
	callbackUnregistration := pb.CallbackHandler{
		CallbackID: n.CallbackID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ack, err := n.client.UnregisterCallback(ctx, &callbackUnregistration)
	if err != nil {
		return err
	}
	if !ack.Success {
		return errors.New("Failed to unregister callback [" + ack.ErrorMessage + "]")
	}
	return nil
}

func (n *Node) StreamOutBlock(task *[]byte) (string, error) {
	// This is reversed from the server side
	stream, err := n.client.StreamInBlock(context.Background())
	if err != nil {
		return "", err
	}
	chunkIndex := 0
	for {
		chunk := chunker.ChunkBytes(*task, &chunkIndex)
		err := stream.Send(&pb.BlockChunk{BlockChunk: chunk})
		if err != nil {
			logger.Error("Failed to send block chunk", "NODE", err)
			return "", err
		}
		if len(chunk) == 0 {
			break
		}
	}
	// Send the EOF
	response, err := stream.CloseAndRecv()
	if err != nil {
		logger.Error("Failed to close stream", "NODE", err)
		return "", err
	}
	return response.BlockID, nil
}

func (n *Node) StreamInBlock(blockID string) ([]byte, error) {
	// The client will receive the block that is streamed
	buffer := []byte{}
	stream, err := n.client.StreamOutBlock(context.Background(), &pb.BlockHandler{BlockID: blockID})
	if err != nil {
		logger.Error("Failed to initiate stream", "NODE", err)
		return nil, err
	}
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Error("Failed to receive block chunk", "NODE", err)
			return nil, err
		}
		buffer = append(buffer, chunk.BlockChunk...)
	}
	return buffer, nil
}

func (n *Node) RegisterTask(task *TaskDefinition) (string, error) {
	// We start by streaming each block
	functionBlockID, err := n.StreamOutBlock(&task.CallableBinary)
	if err != nil {
		return "", err
	}
	n.Blocks = append(n.Blocks, BlockInternalReference{BlockID: functionBlockID, BlockIdentifier: "function-" + task.Name})
	argumentsBlockIDs := make(map[uint32]*pb.BlockHandler)
	for index, argument := range task.ArgumentsBinary {
		argumentBlockID, err := n.StreamOutBlock(&argument)
		if err != nil {
			return "", err
		}
		argumentsBlockIDs[uint32(index)] = &pb.BlockHandler{BlockID: argumentBlockID}
		idAsString := strconv.Itoa(index)
		n.Blocks = append(n.Blocks, BlockInternalReference{BlockID: argumentBlockID, BlockIdentifier: "argument-" + idAsString + "-" + task.Name})
	}
	kwargsBockIDs := make(map[string]*pb.BlockHandler)
	for key, value := range task.KwargsBinary {
		kwargsBlockID, err := n.StreamOutBlock(&value)
		if err != nil {
			return "", err
		}
		kwargsBockIDs[key] = &pb.BlockHandler{BlockID: kwargsBlockID}
		n.Blocks = append(n.Blocks, BlockInternalReference{BlockID: kwargsBlockID, BlockIdentifier: "kwargs-" + key + "-" + task.Name})
	}

	// Now we can register the command
	commandRegistration := pb.CommandRegistration{
		CallableBlockHandler:    &pb.BlockHandler{BlockID: functionBlockID},
		ArgumentsBlocksHandlers: argumentsBlockIDs,
		KwargsBlocksHandlers:    kwargsBockIDs,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := n.client.RegisterCommand(ctx, &commandRegistration)
	if err != nil {
		return "", err
	}
	return res.CommandID, nil
}

func (n *Node) RunCommand(commandID string) error {
	commandHandler := pb.CommandHandler{CommandID: commandID}
	callbackHandler := pb.CallbackHandler{CallbackID: n.CallbackID}
	isolatedEnv := pb.IsolatedEnv{IsolatedEnvID: n.SessionID}
	commandRequest := pb.CommandRequest{CommandHandler: &commandHandler, CallbackHandler: &callbackHandler, IsolatedEnv: &isolatedEnv}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := n.client.RunCommand(ctx, &commandRequest)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New("Failed to run command [" + res.ErrorMessage + "]")
	}
	return nil
}

func (n *Node) GetTaskOutput(commandID string, taskRun *TaskRun) error {
	//    rpc GetCommandOutput (CommandHandler) returns (CommandOutputChunk) {}
	commandHandler := pb.CommandHandler{CommandID: commandID}
	// We need to get the block handlers for the output
	context, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	output, err := n.client.GetCommandOutput(context, &commandHandler)
	if err != nil {
		return err
	}
	OutputBlockID := output.ObjectReturnBlockHandler.BlockID
	StdoutBlockHandler := output.StdoutBlockHandler
	StdErrBlockHandler := output.StderrBlockHandler
	// We need to stream the output
	outputBuffer, err := n.StreamInBlock(OutputBlockID)
	if err != nil {
		return err
	}
	// We need to stream the stdout
	stdoutBuffer, err := n.StreamInBlock(StdoutBlockHandler.BlockID)
	if err != nil {
		return err
	}
	// We need to stream the stderr
	stderrBuffer, err := n.StreamInBlock(StdErrBlockHandler.BlockID)
	if err != nil {
		return err
	}
	taskRun.SetResult(outputBuffer, stdoutBuffer, stderrBuffer, output.Success)
	return nil
}

type NodeTracker struct {
	NodesList                     []*Node
	RoundRobinIndex               int
	ResourceManager               *NodeRequestClient
	ResourceRequestResponse       *NodeRequestResponse
	NodeTrackerHeartbeatExit      chan bool
	NodeTrackerHeartbeatWaitGroup *sync.WaitGroup
}

func (n *NodeTracker) HeartbeatRoutine() {
	// Send the first heartbeat
	n.ResourceManager.SendHeartbeat(n.ResourceRequestResponse.RequestID)
	// Send a heartbeat every REQUEST_HEARTBEAT_INTERVAL
	for {
		select {
		case <-n.NodeTrackerHeartbeatExit:
			defer n.NodeTrackerHeartbeatWaitGroup.Done()
			return
		case <-time.After(REQUEST_HEARTBEAT_INTERVAL):
			n.ResourceManager.SendHeartbeat(n.ResourceRequestResponse.RequestID)
		}
	}
}

func NewNodeTracker(CallbackURI string, ResourceManagerURI string, SessionID string) (*NodeTracker, error) {
	// Later these will come from the session configuration
	nodeType := "small-4cpu-8gb"
	nodeCount := 1

	// We will need to get the nodes from the resource manager
	resourceManager := NewNodeResourceManagerClient(ResourceManagerURI)
	resourceManager.connect()
	requestResponse, err := resourceManager.RequestNode(nodeType, nodeCount)
	logger.Debug("Requesting node", "NODE-TRACKER", logrus.Fields{"RequestID": requestResponse.RequestID, "nodeCount": requestResponse.RequestStatus})

	if err != nil {
		logger.Error("Failed to request node", "NODE-TRACKER", err)
		return nil, err
	}

	// Now we pool for the nodes to be ready
	for {
		// We will need to get the nodes from the resource manager
		requestResponse, err = resourceManager.GetNodeRequestStatus(requestResponse)
		if err != nil {
			logger.Error("Failed to get request status", "NODE-TRACKER", err)
			return nil, err
		}

		if requestResponse.RequestStatus == REQUEST_ALLOCATED {
			logger.Debug("Nodes allocated", "NODE-TRACKER", logrus.Fields{"nodeType": requestResponse.RequestID})
			break
		} else if requestResponse.RequestStatus == REQUEST_ERROR {
			resourceManager.ReleaseNodes(requestResponse)
			resourceManager.disconnect()
			return nil, errors.New("failed to get nodes")
		}
		time.Sleep(REQUEST_POOLING_INTERVAL)
	}

	// We now add the nodes to the tracker
	nodeTracker := &NodeTracker{
		NodesList:               make([]*Node, 0),
		RoundRobinIndex:         0,
		ResourceManager:         resourceManager,
		ResourceRequestResponse: requestResponse,
	}

	for _, liveNode := range requestResponse.Nodes {
		// Add the node to the tracker (this will connect the node, register the callback and initialize the isolated environment)
		err = nodeTracker.AddNode(liveNode.NodeGRPCAddress, CallbackURI, SessionID)
		if err != nil {
			logger.Error("Failed to add node", "NODE-TRACKER", err, logrus.Fields{"nodeID": liveNode.NodeID, "nodeGRPCAddress": liveNode.NodeGRPCAddress})
			resourceManager.ReleaseNodes(requestResponse)
			resourceManager.disconnect()
			return nil, err
		}
	}

	// Everything is ready, now we can start the heartbeat
	nodeTracker.NodeTrackerHeartbeatExit = make(chan bool)
	nodeTracker.NodeTrackerHeartbeatWaitGroup = &sync.WaitGroup{}
	nodeTracker.NodeTrackerHeartbeatWaitGroup.Add(1)
	go nodeTracker.HeartbeatRoutine()
	return nodeTracker, nil
}

func (n *NodeTracker) AddNode(nodeURI string, callbackURI string, SessionID string) error {
	node := &Node{
		nodeID:      uuid.New().String(),
		NodeURI:     nodeURI,
		CallbackURI: callbackURI,
		SessionID:   SessionID,
	}
	err := node.connect()
	if err != nil {
		return err
	}
	n.NodesList = append(n.NodesList, node) // Add the node to the list
	return nil
}

func (n *NodeTracker) GetNextNode() *Node {
	// If we have no nodes then we should return nil
	if len(n.NodesList) == 0 {
		return nil
	}
	// Get the next node
	node := n.NodesList[n.RoundRobinIndex]
	// Increment the index
	n.RoundRobinIndex++
	// If the index is out of bounds then we need to reset it
	if n.RoundRobinIndex >= len(n.NodesList) {
		n.RoundRobinIndex = 0
	}
	return node
}

func (n *NodeTracker) GetNode(nodeID string) *Node {
	for _, node := range n.NodesList {
		if node.nodeID == nodeID {
			return node
		}
	}
	return nil
}

func (n *NodeTracker) Close() {
	for _, node := range n.NodesList {
		node.disconnect()
	}
	n.ResourceManager.ReleaseNodes(n.ResourceRequestResponse)
	n.ResourceManager.disconnect()
	n.NodeTrackerHeartbeatExit <- true
	n.NodeTrackerHeartbeatWaitGroup.Wait() // Wait for the heartbeat to finish
}
