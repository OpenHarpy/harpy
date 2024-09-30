package task

import (
	pb "client-engine/grpc_node_protocol"
	"client-engine/logger"
	"context"
	"net"

	"google.golang.org/grpc"
)

type CallbackServer struct {
	TaskGroup *TaskGroup
	pb.UnimplementedNodeControllerServer
}

func (s *CallbackServer) Callback(ctx context.Context, in *pb.CommandStatus) (*pb.Ack, error) {
	status := ""
	switch in.CommandStatus {
	case pb.CommandStatusEnum_COMMAND_DEFINED:
		status = "defined"
	case pb.CommandStatusEnum_COMMAND_QUEUED:
		status = "queued"
	case pb.CommandStatusEnum_COMMAND_RUNNING:
		status = "running"
	case pb.CommandStatusEnum_COMMAND_KILLING:
		status = "killing"
	case pb.CommandStatusEnum_COMMAND_DONE:
		status = "done"
	case pb.CommandStatusEnum_COMMAND_PANIC:
		status = "panic"
	}
	s.TaskGroup.CommandIDTaskMapping[in.CommandID].SetStatus(status)
	return &pb.Ack{Success: true}, nil
}

func StartCallbackServer(exit chan bool, taskGroup *TaskGroup, port string, waitServerChan chan bool) {
	// TODO: Check if we need a panic chan here
	port = ":" + port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("failed to listen", "CALLBACK_SERVER", err)
	}
	s := grpc.NewServer()
	pb.RegisterNodeControllerServer(s, &CallbackServer{
		TaskGroup: taskGroup,
	})

	// Goroutine to listen for exit signal
	go func() {
		<-exit
		logger.Info("Stopping callback server", "CALLBACK_SERVER", nil)
		s.Stop()
		waitServerChan <- true
	}()

	if err := s.Serve(lis); err != nil {
		logger.Error("failed to serve", "CALLBACK_SERVER", err)
	}
}
