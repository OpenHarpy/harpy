package main

import (
	"context"
	"io"
	"net"
	pb "remote-runner/grpc_node_protocol"
	"remote-runner/logger"
	"time"

	"remote-runner/chunker"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *NodeServer) InstallPackage(ctx context.Context, in *pb.PackageRequest) (*pb.PackageResponse, error) {
	// For now we are going to simply return the an error
	return &pb.PackageResponse{
		Success:      false,
		ErrorMessage: "Not implemented",
	}, nil
}

func (s *NodeServer) UninstallPackage(ctx context.Context, in *pb.PackageRequest) (*pb.PackageResponse, error) {
	// For now we are going to simply return the an error
	return &pb.PackageResponse{
		Success:      false,
		ErrorMessage: "Not implemented",
	}, nil
}

func (s *NodeServer) ListPackages(ctx context.Context, in *emptypb.Empty) (*pb.PackageList, error) {
	// For now we are going to simply return the an error
	return &pb.PackageList{
		PackageURIs: []string{},
	}, nil
}

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

func (s *NodeServer) RegisterCommand(stream pb.Node_RegisterCommandServer) error {
	callableBinary := []byte{}
	argumentsBinary := map[uint32][]byte{}
	kwargsBinary := map[string][]byte{}
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if chunk.CallableBinaryChunk != nil {
			callableBinary = append(callableBinary, chunk.CallableBinaryChunk...)
		}
		if chunk.ArgumentsBinaryChunk != nil {
			for k, v := range chunk.ArgumentsBinaryChunk {
				_, ok := argumentsBinary[k]
				if ok {
					argumentsBinary[k] = append(argumentsBinary[k], v...)
				} else {
					argumentsBinary[k] = v
				}
			}
		}
		if chunk.KwargsBinaryChunk != nil {
			for k, v := range chunk.KwargsBinaryChunk {
				_, ok := kwargsBinary[k]
				if ok {
					kwargsBinary[k] = append(kwargsBinary[k], v...)
				} else {
					kwargsBinary[k] = v
				}
			}
		}
	}
	// Convert the argumentsBinary to a slice
	argumentsSlice := make([][]byte, len(argumentsBinary))
	for key, value := range argumentsBinary {
		argumentsSlice[key] = value
	}
	processId := uuid.New().String()
	// Create a new process
	process := NewProcess(processId, callableBinary, argumentsSlice, kwargsBinary)
	// Add the process to the live memory
	s.lm.Process[process.ProcessID] = process
	return stream.SendAndClose(&pb.CommandHandler{
		CommandID: process.ProcessID,
	})
}

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

func (s *NodeServer) GetCommandOutput(in *pb.CommandHandler, stream pb.Node_GetCommandOutputServer) error {
	process, ok := s.lm.Process[in.CommandID]
	if !ok {
		stream.Send(&pb.CommandOutputChunk{
			CommandID:               in.CommandID,
			ObjectReturnBinaryChunk: nil,
			StdoutBinaryChunk:       nil,
			StderrBinaryChunk:       nil,
			Success:                 false,
			Panic:                   true,
			ErrorMessage:            "Process not found",
		})
	}
	streamedAll := false
	objectChunkIndex := 0
	stdoutChunkIndex := 0
	stderrChunkIndex := 0
	for {
		objectOutputChunk := chunker.ChunkBytes(process.ObjectReturnBinary, &objectChunkIndex)
		stdoutOutputChunk := chunker.ChunkBytes(process.StdoutBinary, &stdoutChunkIndex)
		stderrOutputChunk := chunker.ChunkBytes(process.StderrBinary, &stderrChunkIndex)
		if len(objectOutputChunk) == 0 && len(stdoutOutputChunk) == 0 && len(stderrOutputChunk) == 0 {
			streamedAll = true
		}
		stream.Send(
			&pb.CommandOutputChunk{
				CommandID:               in.CommandID,
				ObjectReturnBinaryChunk: objectOutputChunk,
				StdoutBinaryChunk:       stdoutOutputChunk,
				StderrBinaryChunk:       stderrOutputChunk,
				Success:                 process.Success,
				Panic:                   false,
				ErrorMessage:            "",
			},
		)
		if streamedAll {
			break
		}
	}
	// Process was streamed back to the client we mark the time served
	// We will keep the process in memory for a while before we clean it up
	// The event loop will clean up the process
	process.SetResultFetchTime()
	s.lm.Process[in.CommandID] = process
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
