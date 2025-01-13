package main

import (
	"context"
	pb "remote-runner/grpc_resource_alloc_procotol"
	"remote-runner/logger"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NodeStatusUpdateClient struct {
	NodeID             string
	NodeGRPCAddress    string
	ResourceManagerURI string
	conn               *grpc.ClientConn
	client             pb.NodeStatusUpdateServiceClient
}

func NewNodeStatusUpdateClient(nodeID, nodeGRPCAddress, resourceManagerURI string) *NodeStatusUpdateClient {
	return &NodeStatusUpdateClient{
		NodeID:             nodeID,
		NodeGRPCAddress:    nodeGRPCAddress,
		ResourceManagerURI: resourceManagerURI,
	}
}

func (n *NodeStatusUpdateClient) connect() error {
	err := error(nil)
	n.conn, err = grpc.NewClient(n.ResourceManagerURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	n.client = pb.NewNodeStatusUpdateServiceClient(n.conn)
	return nil
}

func (n *NodeStatusUpdateClient) disconnect() error {
	if n.conn != nil {
		err := n.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func SendNewStatus(nodeStatusUpdateClient *NodeStatusUpdateClient, status pb.NodeStatus) error {
	nodeStatus := &pb.LiveNode{
		NodeID:          nodeStatusUpdateClient.NodeID,
		NodeGRPCAddress: nodeStatusUpdateClient.NodeGRPCAddress,
		NodeStatus:      status,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := nodeStatusUpdateClient.client.UpdateNodeStatus(ctx, nodeStatus)
	if err != nil {
		logger.Error("Failed to update node status", "NODE_STATUS_UPDATE_CLIENT", err)
		return err
	}
	return nil
}

func (n *NodeStatusUpdateClient) SetNodeBooting() error {
	return SendNewStatus(n, pb.NodeStatus_NODE_STATUS_BOOTING)
}

func (n *NodeStatusUpdateClient) SetNodeReady() error {
	return SendNewStatus(n, pb.NodeStatus_NODE_STATUS_READY)
}

func (n *NodeStatusUpdateClient) SetNodeShuttingDown() error {
	return SendNewStatus(n, pb.NodeStatus_NODE_STATUS_SHUTTING_DOWN)
}

func (n *NodeStatusUpdateClient) SetNodeShutdown() error {
	return SendNewStatus(n, pb.NodeStatus_NODE_STATUS_SHUTDOWN)
}

func (n *NodeStatusUpdateClient) SetNodePanic() error {
	return SendNewStatus(n, pb.NodeStatus_NODE_STATUS_ERROR)
}

func (n *NodeStatusUpdateClient) SendNodeHeartbeat() error {
	nodeHeartbeatRequest := &pb.NodeHeartbeatRequest{
		NodeID: n.NodeID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := n.client.NodeHeartbeat(ctx, nodeHeartbeatRequest)
	if err != nil {
		logger.Error("Failed to update node heartbeat", "NODE_STATUS_UPDATE_CLIENT", err)
		return err
	}
	return nil
}
