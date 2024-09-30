package main

import (
	"context"
	pb "remote-runner/grpc_node_protocol"
	"remote-runner/logger"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Callback struct {
	CallbackID  string
	CallbackURI string
}

type CallbackClient struct {
	ContollerURI string
	conn         *grpc.ClientConn
	client       pb.NodeControllerClient
}

func (n *CallbackClient) connect() error {
	err := error(nil)
	n.conn, err = grpc.NewClient(n.ContollerURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	n.client = pb.NewNodeControllerClient(n.conn)
	return nil
}

func (n *CallbackClient) disconnect() error {
	if n.conn != nil {
		err := n.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *CallbackClient) Callback(commandID string, status string) error {
	commandStatus := pb.CommandStatusEnum_COMMAND_DEFINED
	switch status {
	case PROCESS_DEFINED:
		commandStatus = pb.CommandStatusEnum_COMMAND_DEFINED
	case PROCESS_QUEUED:
		commandStatus = pb.CommandStatusEnum_COMMAND_QUEUED
	case PROCESS_RUNNING:
		commandStatus = pb.CommandStatusEnum_COMMAND_RUNNING
	case PROCESS_KILLING:
		commandStatus = pb.CommandStatusEnum_COMMAND_KILLING
	case PROCESS_DONE:
		commandStatus = pb.CommandStatusEnum_COMMAND_DONE
	case PROCESS_PANIC:
		commandStatus = pb.CommandStatusEnum_COMMAND_PANIC
	}
	callbackStatus := pb.CommandStatus{
		CommandID:     commandID,
		CommandStatus: commandStatus,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := n.client.Callback(ctx, &callbackStatus)
	if err != nil {
		return err
	}
	logger.Info("Callback sent to the controller", "CALLBACK", logrus.Fields{"command_id": commandID, "status": status})
	return nil
}

func NewCallbackClient(controllerURI string) *CallbackClient {
	return &CallbackClient{
		ContollerURI: controllerURI,
	}
}
