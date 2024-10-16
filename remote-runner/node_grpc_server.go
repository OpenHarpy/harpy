package main

import (
	"context"
	"errors"
	"io"
	"net"
	pb "remote-runner/grpc_node_protocol"
	"remote-runner/logger"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// TODO: Make these configurable via the configs
	port                       = ":50053"
	processPoolingInterval     = 200 * time.Millisecond
	timeoutAfterGettingResult  = 30 * time.Second
	timeoutProcessNotTriggered = 120 * time.Second
	heartbeatInterval          = 10 * time.Second
	allowParallelProcesses     = 4
)

type NodeServer struct {
	lm *LiveMemory
	pb.UnimplementedNodeServer
}

// Isolated enviroment
func (s *NodeServer) IsolatedEnvInit(ctx context.Context, in *pb.IsolatedEnv) (*pb.Ack, error) {
	// For now we are going to simply return the an error
	isolatedEnv := NewIsolatedEnvironment(in.SessionID)
	err := isolatedEnv.Begin()
	if err != nil {
		return &pb.Ack{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}
	s.lm.IsolatedEnvironment[in.SessionID] = isolatedEnv
	return &pb.Ack{
		Success:      true,
		ErrorMessage: "",
	}, nil
}
func (s *NodeServer) IsolatedEnvDestroy(ctx context.Context, in *pb.IsolatedEnv) (*pb.Ack, error) {
	// For now we are going to simply return the an error
	isolatedEnv, ok := s.lm.IsolatedEnvironment[in.SessionID]
	if !ok {
		return &pb.Ack{
			Success:      false,
			ErrorMessage: "Isolated environment not found",
		}, nil
	}
	err := isolatedEnv.Cleanup()
	if err != nil {
		return &pb.Ack{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}
	delete(s.lm.IsolatedEnvironment, in.SessionID)
	return &pb.Ack{
		Success:      true,
		ErrorMessage: "",
	}, nil
}
func (s *NodeServer) IsolatedEnvInstallPackage(ctx context.Context, in *pb.IsolatedEnvPackageRequest) (*pb.IsolatedEnvPackageResponse, error) {
	// For now we are going to simply return the an error
	return &pb.IsolatedEnvPackageResponse{
		Success:      false,
		ErrorMessage: "Not implemented",
	}, nil
}
func (s *NodeServer) IsolatedEnvUninstallPackage(ctx context.Context, in *pb.IsolatedEnvPackageRequest) (*pb.IsolatedEnvPackageResponse, error) {
	// For now we are going to simply return the an error
	return &pb.IsolatedEnvPackageResponse{
		Success:      false,
		ErrorMessage: "Not implemented",
	}, nil
}
func (s *NodeServer) IsolatedEnvListPackages(ctx context.Context, in *pb.IsolatedEnv) (*pb.IsolatedEnvPackageList, error) {
	// For now we are going to simply return the an error
	return &pb.IsolatedEnvPackageList{
		PackageURIs: []string{},
	}, nil
}

// Callbacks
func (s *NodeServer) RegisterCallback(ctx context.Context, in *pb.CallbackRegistration) (*pb.CallbackHandler, error) {
	// Create a new callback
	callback := &Callback{
		CallbackID:  uuid.New().String(),
		CallbackURI: in.CallbackURI,
	}
	// Create new callback client
	callbackClient := NewCallbackClient(in.CallbackURI)
	err := callbackClient.connect()
	if err != nil {
		return &pb.CallbackHandler{
			CallbackID: "",
		}, err
	}
	// Add the callback client to the live memory
	s.lm.CallbackClients[callback.CallbackID] = callbackClient

	// Add the callback to the live memory
	s.lm.Callback[callback.CallbackID] = callback
	logger.Info("Callback registered", "SERVER", logrus.Fields{"callback_id": callback.CallbackID})
	return &pb.CallbackHandler{
		CallbackID: callback.CallbackID,
	}, nil
}

func (s *NodeServer) UnregisterCallback(ctx context.Context, in *pb.CallbackHandler) (*pb.Ack, error) {
	// Remove the callback from the live memory
	_, ok := s.lm.Callback[in.CallbackID]
	if !ok {
		return &pb.Ack{
			Success:      false,
			ErrorMessage: "Callback not found",
		}, nil
	}

	// Disconnect the callback client
	callbackClient, ok := s.lm.CallbackClients[in.CallbackID]
	if ok {
		err := callbackClient.disconnect()
		if err != nil {
			return &pb.Ack{
				Success:      false,
				ErrorMessage: err.Error(),
			}, nil
		}
	}
	// Remove the callback from the live memory
	delete(s.lm.Callback, in.CallbackID)
	return &pb.Ack{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

func (s *NodeServer) RegisterCommand(ctx context.Context, in *pb.CommandRegistration) (*pb.CommandHandler, error) {
	// Validate the input
	callableBlock := s.lm.Blocks[in.CallableBlockHandler.BlockID]
	if callableBlock == nil {
		return &pb.CommandHandler{
			CommandID: "",
		}, status.Errorf(404, "Callable block not found")
	}
	argumentsBlocks := []*Block{}
	for _, block := range in.ArgumentsBlocksHandlers {
		argumentsBlock := s.lm.Blocks[block.BlockID]
		if argumentsBlock == nil {
			return &pb.CommandHandler{
				CommandID: "",
			}, status.Errorf(404, "Arguments block not found")
		}
		argumentsBlocks = append(argumentsBlocks, argumentsBlock)
	}
	keywordArgumentsBlocks := map[string]*Block{}
	for key, block := range in.KwargsBlocksHandlers {
		keywordArgumentsBlock := s.lm.Blocks[block.BlockID]
		if keywordArgumentsBlock == nil {
			return &pb.CommandHandler{
				CommandID: "",
			}, status.Errorf(404, "Keyword arguments block not found")
		}
		keywordArgumentsBlocks[key] = keywordArgumentsBlock
	}
	// Create a new process
	processID := uuid.New().String()
	// Build blocks for the process
	StdOutBlock := NewBlock(uuid.New().String())
	StdErrBlock := NewBlock(uuid.New().String())
	OutputBlock := NewBlock(uuid.New().String())

	s.lm.Blocks[StdOutBlock.BlockID] = StdOutBlock
	s.lm.Blocks[StdErrBlock.BlockID] = StdErrBlock
	s.lm.Blocks[OutputBlock.BlockID] = OutputBlock

	// Create a new process
	process := NewProcess(
		processID,
		callableBlock,
		argumentsBlocks,
		keywordArgumentsBlocks,
		OutputBlock,
		StdOutBlock,
		StdErrBlock,
	)
	// Add the process to the live memory
	s.lm.Process[processID] = process
	logger.Info("Process registered", "SERVER", logrus.Fields{"process_id": processID})
	return &pb.CommandHandler{
		CommandID: processID,
	}, nil
}

// Commands
func (s *NodeServer) RunCommand(ctx context.Context, in *pb.CommandRequest) (*pb.CommandRequestResponse, error) {
	// Cannot run a command if the callback is not set
	if s.lm.Callback == nil {
		return &pb.CommandRequestResponse{
			Success:      false,
			ErrorMessage: "Callback not found",
		}, nil
	}
	// Get the process from the live memory
	callback, ok := s.lm.Callback[in.CallbackHandler.CallbackID]
	if !ok {
		return &pb.CommandRequestResponse{
			Success:      false,
			ErrorMessage: "Callback not found",
		}, nil
	}
	process, ok := s.lm.Process[in.CommandHandler.CommandID]
	if !ok {
		return &pb.CommandRequestResponse{
			Success:      false,
			ErrorMessage: "Process not found",
		}, nil
	}
	// Mark the process for queuing
	process.QueueProcess(callback.CallbackID)
	s.lm.Process[in.CommandHandler.CommandID] = process
	// Isolated environment ref
	isolatedRef, ok := s.lm.IsolatedEnvironment[in.IsolatedEnv.SessionID]
	if !ok {
		return &pb.CommandRequestResponse{
			Success:      false,
			ErrorMessage: "Isolated environment not found",
		}, nil
	}
	// Reference the isolated environment
	s.lm.Process[in.CommandHandler.CommandID].EnvRef = isolatedRef
	// We keep pooling the process until it is done
	logger.Info("Process queued - Streaming process status", "SERVER", logrus.Fields{"process_id": in.CommandHandler.CommandID})
	ReportToCallbackClient(s.lm, in.CommandHandler.CommandID, PROCESS_QUEUED)
	return &pb.CommandRequestResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

func (s *NodeServer) KillCommand(ctx context.Context, in *pb.CommandRequest) (*pb.CommandRequestResponse, error) {
	// Cannot kill a command if the callback is not set
	if s.lm.Callback == nil {
		return &pb.CommandRequestResponse{
			Success:      false,
			ErrorMessage: "Callback not found",
		}, nil
	}
	// Get the callback from the live memory
	callback, ok := s.lm.Callback[in.CallbackHandler.CallbackID]
	if !ok {
		return &pb.CommandRequestResponse{
			Success:      false,
			ErrorMessage: "Callback not found",
		}, nil
	}
	// Get the process from the live memory
	process, ok := s.lm.Process[in.CommandHandler.CommandID]
	if !ok {
		return &pb.CommandRequestResponse{
			Success:      false,
			ErrorMessage: "Process not found",
		}, nil
	}
	// Set the process to killing
	process.KillProcess(callback.CallbackID)
	s.lm.Process[in.CommandHandler.CommandID] = process
	return &pb.CommandRequestResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

/*
message CommandOutput {
    string CommandID = 1;
    BlockHandler ObjectReturnBlockHandler = 2;
    BlockHandler StdoutBlockHandler = 3;
    BlockHandler StderrBlockHandler = 4;
    bool Success = 5;
    bool Panic = 6;
    string ErrorMessage = 7;
}
*/

func (s *NodeServer) GetCommandOutput(ctx context.Context, in *pb.CommandHandler) (*pb.CommandOutput, error) {
	// Get the process from the live memory
	process, ok := s.lm.Process[in.CommandID]
	if !ok {
		return nil, status.Errorf(404, "Process not found")
	}
	// Send the output blocks and the status
	return &pb.CommandOutput{
		CommandID:                process.ProcessID,
		ObjectReturnBlockHandler: &pb.BlockHandler{BlockID: process.OutputBlock.BlockID},
		StdoutBlockHandler:       &pb.BlockHandler{BlockID: process.StdoutBlock.BlockID},
		StderrBlockHandler:       &pb.BlockHandler{BlockID: process.StderrBlock.BlockID},
		Success:                  process.Success,
	}, nil
}

// Blocks
func (s *NodeServer) StreamInBlock(stream pb.Node_StreamInBlockServer) error {
	// Start by creating a new block
	blockID := uuid.New().String()
	block := NewBlock(blockID)
	blockWriter, err := NewBlockWriter(block)
	if err != nil {
		logger.Error("Failed to create block", "SERVER", err)
		return status.Errorf(500, "Failed to create block")
	}
	// Stream the block
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("Failed to receive block chunk", "SERVER", err)
			return status.Errorf(500, "Failed to receive block chunk")
		}
		blockWriter.AppendChunk(chunk.BlockChunk)
	}
	// Seal the block
	err = blockWriter.SealBlock()
	if err != nil {
		return err
	}
	// Add the block to the live memory
	s.lm.Blocks[blockID] = block
	// Return the block ID
	blockHandler := &pb.BlockHandler{BlockID: blockID}
	return stream.SendAndClose(blockHandler)
}

func (s *NodeServer) StreamOutBlock(in *pb.BlockHandler, stream pb.Node_StreamOutBlockServer) error {
	// Get the block from the live memory
	block := s.lm.Blocks[in.BlockID]
	if block == nil {
		logger.Error("Block not found", "SERVER", errors.New("Block not found"), logrus.Fields{"block_id": in.BlockID})
		return status.Errorf(404, "Block not found")
	}
	// Stream the block
	blockReader, err := NewBlockReader(block)
	hasReader := true
	if err != nil {
		logger.Error("Failed to create block reader", "SERVER", err)
		logger.Warn("This usually means the block is not found under the file system - Streaming back an empty block", "SERVER")
		hasReader = false
	}
	if !hasReader {
		chunk := []byte{}
		stream.Send(&pb.BlockChunk{BlockChunk: chunk})
	} else {
		for {
			if blockReader.State == BlockReaderStateStreamEnd {
				break
			}
			chunk, err := blockReader.ReadChunk()
			if err != nil {
				logger.Error("Failed to read block chunk", "SERVER", err)
				return status.Errorf(500, "Failed to read block")
			}
			stream.Send(&pb.BlockChunk{BlockChunk: chunk})
		}
	}
	// We are done now we can close the stream (EOF)
	return nil
}

func ReportToCallbackClient(lm *LiveMemory, CommandID string, status string) {
	// Report the status to the callback client
	process, ok := lm.Process[CommandID]
	if !ok {
		return
	}
	callbackID := process.CallbackID
	callbackClient, ok := lm.CallbackClients[callbackID]
	if !ok {
		return
	}
	err := callbackClient.Callback(CommandID, status)
	if err != nil {
		logger.Error("Failed to report to callback client", "SERVER", err)
	}
}

func NewNodeServer(exit chan bool, panicMainServer chan bool, waitServerExitChan chan bool, port string, lm *LiveMemory) {
	logger.Info("gRPC server started", "SERVER", logrus.Fields{"host": port})

	port = ":" + port
	s := grpc.NewServer()

	nodeServer := &NodeServer{
		lm: lm,
	}
	pb.RegisterNodeServer(s, nodeServer)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("failed to listen", "SERVER", err)
		panicMainServer <- true
		waitServerExitChan <- true
	}

	// Goroutine to listen for exit signal
	go func() {
		<-exit
		logger.Info("Received exit signal, stopping server...", "SERVER")
		s.GracefulStop()
		waitServerExitChan <- true
	}()

	if err := s.Serve(lis); err != nil {
		logger.Error("failed to serve", "SERVER", err)
		panicMainServer <- true
		waitServerExitChan <- true
	}
}
