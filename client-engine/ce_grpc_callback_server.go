// Package main
//
// This file contains the implementation of the callback server
// the callback server is spawned by the client-engine when a task-group is created
// the server is terminated when the task-group is completed
// the server listens for the callback from the node when any command is updated
//
// Author: Caio Cominato
package main

import (
	pb "client-engine/grpc_node_protocol"
	"client-engine/logger"
	"client-engine/task"
	"context"
	"errors"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type CallbackServer struct {
	lm *LiveMemory
	pb.UnimplementedNodeControllerServer
}

func (s *CallbackServer) Callback(ctx context.Context, in *pb.CommandStatus) (*pb.Ack, error) {
	status := task.STATUS_UNDEFINED
	switch in.CommandStatus {
	case pb.CommandStatusEnum_COMMAND_DEFINED:
		status = task.STATUS_DEFINED
	case pb.CommandStatusEnum_COMMAND_QUEUED:
		status = task.STATUS_QUEUED
	case pb.CommandStatusEnum_COMMAND_RUNNING:
		status = task.STATUS_RUNNING
	case pb.CommandStatusEnum_COMMAND_KILLING:
		status = task.STATUS_KILLING
	case pb.CommandStatusEnum_COMMAND_DONE:
		status = task.STATUS_DONE
	case pb.CommandStatusEnum_COMMAND_PANIC:
		status = task.STATUS_PANIC
	}
	logger.Debug("Received callback", "CALLBACK_SERVER", logrus.Fields{"commandID": in.CommandID, "status": status})
	// Get the callback pointer
	callbackPointerID := s.lm.CommandCallbackPointer[in.CommandID]
	callbackPointer := s.lm.CallbackPointers[callbackPointerID]
	if callbackPointer == nil {
		logger.Error("Callback pointer not found", "CALLBACK_SERVER", errors.New("Callback pointer not found"))
		return &pb.Ack{Success: false}, nil
	}
	// Call the callback
	err := callbackPointer(in.CommandID, status)
	if err != nil {
		return &pb.Ack{Success: false}, err
	}
	return &pb.Ack{Success: true}, nil
}

func StartCallbackServer(exit chan bool, wg *sync.WaitGroup, lm *LiveMemory, port string) error {
	wg.Add(1)
	port = ":" + port
	logger.Info("Starting callback server", "CALLBACK_SERVER", logrus.Fields{"host": port})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("failed to listen", "CALLBACK_SERVER", err)
		defer wg.Done()
		return err
	} else {
		logger.Info("Callback server listening", "CALLBACK_SERVER", logrus.Fields{"host": port})
		s := grpc.NewServer()
		pb.RegisterNodeControllerServer(s, &CallbackServer{
			lm: lm,
		})
		// Goroutine for the server
		go func() {
			if err := s.Serve(lis); err != nil {
				logger.Error("failed to serve", "CALLBACK_SERVER", err)
				defer wg.Done()
			}
		}()
		go func() {
			<-exit
			logger.Info("Stopping callback server", "CALLBACK_SERVER", nil)
			s.GracefulStop()
			logger.Info("Callback server stopped", "CALLBACK_SERVER", nil)
			defer wg.Done()
		}()
	}
	return nil
}
