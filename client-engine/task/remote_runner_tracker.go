package task

import (
	"client-engine/chunker"
	pb "client-engine/grpc_node_protocol"
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	REQUEST_POOLING_INTERVAL = 200 * time.Millisecond
)

type Node struct {
	nodeID  string
	NodeURI string
	conn    *grpc.ClientConn
	client  pb.NodeClient
}

func (n *Node) connect() error {
	err := error(nil)
	n.conn, err = grpc.NewClient(n.NodeURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	n.client = pb.NewNodeClient(n.conn)
	return nil
}

func (n *Node) disconnect() error {
	if n.conn != nil {
		err := n.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) RegisterCallback(CallbackURI string) (string, error) {
	callbackRegistration := pb.CallbackRegistration{
		CallbackURI: CallbackURI,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := n.client.RegisterCallback(ctx, &callbackRegistration)
	if err != nil {
		log.Printf("Failed to register callback: %v", err)
		return "", err
	}
	return res.CallbackID, nil
}

func (n *Node) UnregisterCallback(CallbackID string) error {
	callbackUnregistration := pb.CallbackHandler{
		CallbackID: CallbackID,
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

func (n *Node) RegisterTask(task *TaskDefinition) (string, error) {
	// Establish a connection to the node
	context := context.Background()
	stream, err := n.client.RegisterCommand(context)
	if err != nil {
		return "", err
	}

	// Send the task to the node
	chunkIndexCallableBinary := 0
	chunkIndexArgumentsBinary := map[uint32]*int{}
	chunkIndexKeywordArgumentsBinary := map[string]*int{}

	doneStreamArguments := map[uint32]bool{}
	doneStreamKeywordArguments := map[string]bool{}

	streamDone := false
	for {
		argumentsBinaryChunk := make(map[uint32][]byte)
		keywordArgumentsBinaryChunk := make(map[string][]byte)

		chunkCallableBinary := chunker.ChunkBytes(task.CallableBinary, &chunkIndexCallableBinary)
		for index, argumentsBinary := range task.ArgumentsBinary {
			idx := uint32(index)
			_, exists := chunkIndexArgumentsBinary[idx]
			if !exists {
				// We then initialize the index in the map
				chunkIndexArgumentsBinary[idx] = new(int)
				doneStreamArguments[idx] = false
			}
			argumentsBinaryChunk[idx] = chunker.ChunkBytes(argumentsBinary, chunkIndexArgumentsBinary[idx])
			if len(argumentsBinaryChunk[idx]) == 0 {
				doneStreamArguments[idx] = true
			}
		}
		for key, keywordArgumentsBinary := range task.KwargsBinary {
			_, exists := chunkIndexKeywordArgumentsBinary[key]
			if !exists {
				chunkIndexKeywordArgumentsBinary[key] = new(int)
				doneStreamKeywordArguments[key] = false
			}
			keywordArgumentsBinaryChunk[key] = chunker.ChunkBytes(keywordArgumentsBinary, chunkIndexKeywordArgumentsBinary[key])
			if len(keywordArgumentsBinaryChunk[key]) == 0 {
				doneStreamKeywordArguments[key] = true
			}
		}

		commandChunk := pb.CommandRequestChunk{
			CallableBinaryChunk:  chunkCallableBinary,
			ArgumentsBinaryChunk: argumentsBinaryChunk,
			KwargsBinaryChunk:    keywordArgumentsBinaryChunk,
		}
		err := stream.Send(&commandChunk)
		if err != nil {
			return "", err
		}
		doneStreamingArguments := true
		for _, done := range doneStreamArguments {
			doneStreamingArguments = doneStreamingArguments && done
		}
		doneStreamingKeywordArguments := true
		for _, done := range doneStreamKeywordArguments {
			doneStreamingKeywordArguments = doneStreamingKeywordArguments && done
		}
		streamDone = doneStreamingArguments && doneStreamingKeywordArguments && len(chunkCallableBinary) == 0
		if streamDone {
			break
		}
	}

	// Get the response from the node
	response, err := stream.CloseAndRecv()
	if err != nil {
		return "", err
	}
	return response.CommandID, nil
}

func (n *Node) RunCommand(commandID string, callbackID string) error {
	commandHandler := pb.CommandHandler{CommandID: commandID}
	callbackHandler := pb.CallbackHandler{CallbackID: callbackID}
	commandRequest := pb.CommandRequest{CommandHandler: &commandHandler, CallbackHandler: &callbackHandler}
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
	//    rpc GetCommandOutput (CommandHandler) returns (stream CommandOutputChunk) {}
	commandHandler := pb.CommandHandler{CommandID: commandID}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	stream, err := n.client.GetCommandOutput(ctx, &commandHandler)
	if err != nil {
		return err
	}

	objectReturnBinary := []byte{}
	stdoutBinary := []byte{}
	stderrBinary := []byte{}
	success := false

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		if chunk.ObjectReturnBinaryChunk != nil {
			objectReturnBinary = append(objectReturnBinary, chunk.ObjectReturnBinaryChunk...)
		}
		if chunk.StdoutBinaryChunk != nil {
			stdoutBinary = append(stdoutBinary, chunk.StdoutBinaryChunk...)
		}
		if chunk.StderrBinaryChunk != nil {
			stderrBinary = append(stderrBinary, chunk.StderrBinaryChunk...)
		}
		if chunk.Success {
			success = true
		}
	}

	taskRun.SetResult(objectReturnBinary, stdoutBinary, stderrBinary, success)
	return nil
}

type NodeTracker struct {
	NodesList               []*Node
	RoundRobinIndex         int
	ResourceManager         *NodeRequestClient
	ResourceRequestResponse *NodeRequestResponse
}

func NewNodeTracker() (*NodeTracker, error) {
	// Later these will come from the session configuration
	nodeType := "small-4cpu-8gb"
	nodeCount := 1

	// We will need to get the nodes from the resource manager
	resourceManager := NewNodeResourceManagerClient("localhost:50050")
	resourceManager.connect()
	requestResponse, err := resourceManager.RequestNode(nodeType, nodeCount)

	if err != nil {
		return nil, err
	}

	// Now we pool for the nodes to be ready
	for {
		// We will need to get the nodes from the resource manager
		requestResponse, err = resourceManager.GetNodeRequestStatus(requestResponse)
		if err != nil {
			return nil, err
		}

		if requestResponse.RequestStatus == REQUEST_ALLOCATED {
			break
		} else if requestResponse.RequestStatus == REQUEST_ERROR {
			resourceManager.ReleaseNodes(requestResponse) // We need to release the nodes
			return nil, errors.New("Failed to get nodes")
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
		nodeTracker.AddNode(liveNode.NodeGRPCAddress)
	}
	return nodeTracker, nil
}

func (n *NodeTracker) AddNode(nodeURI string) {
	node := &Node{
		nodeID:  uuid.New().String(),
		NodeURI: nodeURI,
	}
	err := node.connect()
	if err != nil {
		panic(err)
	}
	n.NodesList = append(n.NodesList, node)
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
}
