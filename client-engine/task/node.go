// Package task implements any task related operations
//
// This file contains the implementation of the Node structure
// Nodes are responsible for executing the tasks and managing the blocks
//

package task

import (
	pb "client-engine/grpc_node_protocol"
	"client-engine/logger"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	nodeID       string
	NodeURI      string
	CallbackURI  string
	CallbackID   string
	SessionID    string
	BlockTracker *BlockTracker
	conn         *grpc.ClientConn
	client       pb.NodeClient
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

func (n *Node) StreamMetadataJson(metadata map[string]string) (string, error) {
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	BlockStreamWriter := NewBlockStreamingWriter(n)
	err = BlockStreamWriter.Write(metadataJson)
	if err != nil {
		return "", err
	}
	err = BlockStreamWriter.Close()
	if err != nil {
		return "", err
	}
	return BlockStreamWriter.BlockID, nil
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

func (n *Node) RegisterTask(task *TaskDefinition, TaskSet *TaskSet) (string, error) {
	// We start by streaming each block
	logger.Info("Registering task", "NODE", logrus.Fields{"BlockGroup": TaskSet.CurrentBlockGroupID})
	n.BlockTracker.AddBlock(&BlockInternalReference{BlockID: string(task.CallableBlockID), BlockType: "function", BlockGroup: *TaskSet.CurrentBlockGroupID})
	argumentsBlockIDs := make(map[uint32]*pb.BlockHandler)
	for index, argument := range task.ArgumentsBlockIDs {
		argumentsBlockIDs[uint32(index)] = &pb.BlockHandler{BlockID: string(argument)}
		n.BlockTracker.AddBlock(&BlockInternalReference{BlockID: string(argument), BlockType: "argument", BlockGroup: *TaskSet.CurrentBlockGroupID})
	}
	kwargsBockIDs := make(map[string]*pb.BlockHandler)
	for key, value := range task.KwargsBlockIDs {
		kwargsBockIDs[key] = &pb.BlockHandler{BlockID: string(value)}
		n.BlockTracker.AddBlock(&BlockInternalReference{BlockID: string(value), BlockType: "kwargs", BlockGroup: *TaskSet.CurrentBlockGroupID})
	}
	var MetadataBlockHandler *pb.BlockHandler = nil
	if task.Metadata != nil {
		metadataBlockID, err := n.StreamMetadataJson(task.Metadata)
		if err != nil {
			return "", err
		}
		n.BlockTracker.AddBlock(&BlockInternalReference{BlockID: metadataBlockID, BlockType: "metadata", BlockGroup: *TaskSet.CurrentBlockGroupID})
		MetadataBlockHandler = &pb.BlockHandler{BlockID: metadataBlockID}
	}

	// Now we can register the command
	commandRegistration := pb.CommandRegistration{
		CallableBlockHandler:    &pb.BlockHandler{BlockID: string(task.CallableBlockID)},
		ArgumentsBlocksHandlers: argumentsBlockIDs,
		KwargsBlocksHandlers:    kwargsBockIDs,
		MetadataBlockHandler:    MetadataBlockHandler,
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
	commandRequest := pb.CommandRequest{
		CommandHandler:  &commandHandler,
		CallbackHandler: &callbackHandler,
		IsolatedEnv:     &isolatedEnv,
	}
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
	if taskRun.TaskSet == nil {
		panic(errors.New("taskSet cannot be nil"))
	}
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

	// Track the blocks
	n.BlockTracker.AddBlock(&BlockInternalReference{BlockID: string(OutputBlockID), BlockType: "output", BlockGroup: *taskRun.TaskSet.CurrentBlockGroupID})
	n.BlockTracker.AddBlock(&BlockInternalReference{BlockID: string(StdoutBlockHandler.BlockID), BlockType: "stdout", BlockGroup: *taskRun.TaskSet.CurrentBlockGroupID})
	n.BlockTracker.AddBlock(&BlockInternalReference{BlockID: string(StdErrBlockHandler.BlockID), BlockType: "stderr", BlockGroup: *taskRun.TaskSet.CurrentBlockGroupID})
	taskRun.SetResult(BlockID(OutputBlockID), BlockID(StdoutBlockHandler.BlockID), BlockID(StdErrBlockHandler.BlockID), output.Success)
	return nil
}

func (n *Node) FlushAllBlocks() error {
	isolatedEnv := pb.IsolatedEnv{IsolatedEnvID: n.SessionID}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ack, err := n.client.ClearBlocks(ctx, &isolatedEnv)
	if err != nil {
		return err
	}
	if !ack.Success {
		return errors.New("Failed to clear blocks [" + ack.ErrorMessage + "]")
	}
	return nil
}

func (n *Node) FlushBlock(blockID string) error {
	blockHandler := pb.BlockHandler{BlockID: blockID}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ack, err := n.client.DestroyBlock(ctx, &blockHandler)
	if err != nil {
		return err
	}
	if !ack.Success {
		return errors.New("Failed to clear block [" + ack.ErrorMessage + "]")
	}
	return nil
}
