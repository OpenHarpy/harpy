// Package task implements any task related operations
//
// This file contains the implementation of the resource manager client
// This client is used by the RRT to request nodes from the resource manager
// The session with the resource manager is kept alive until the session is closed
//
// Author: Caio Cominato

package task

import (
	pb "client-engine/grpc_resource_alloc_procotol"
	"client-engine/logger"
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NodeStatus pb.NodeStatus
type RequestStatus pb.ServingStatus

const (
	NODE_UNKNOWN       NodeStatus = NodeStatus(pb.NodeStatus_NODE_STATUS_UNKNOWN)
	NODE_BOOTING       NodeStatus = NodeStatus(pb.NodeStatus_NODE_STATUS_BOOTING)
	NODE_READY         NodeStatus = NodeStatus(pb.NodeStatus_NODE_STATUS_READY)
	NODE_SHUTTING_DOWN NodeStatus = NodeStatus(pb.NodeStatus_NODE_STATUS_SHUTTING_DOWN)
	NODE_SHUTDOWN      NodeStatus = NodeStatus(pb.NodeStatus_NODE_STATUS_SHUTDOWN)
	NODE_ERROR         NodeStatus = NodeStatus(pb.NodeStatus_NODE_STATUS_ERROR)
)

const (
	REQUEST_UNKNOWN           RequestStatus = RequestStatus(pb.ServingStatus_SERVING_STATUS_UNKNOWN)
	REQUEST_REQUESTED         RequestStatus = RequestStatus(pb.ServingStatus_SERVING_STATUS_REQUESTED)
	REQUEST_ALLOCATING        RequestStatus = RequestStatus(pb.ServingStatus_SERVING_STATUS_ALLOCATING)
	REQUEST_ALLOCATED         RequestStatus = RequestStatus(pb.ServingStatus_SERVING_STATUS_ALLOCATED)
	REQUEST_RELEASE_REQUESTED RequestStatus = RequestStatus(pb.ServingStatus_SERVING_STATUS_RELEASE_REQUESTED)
	REQUEST_RELEASED          RequestStatus = RequestStatus(pb.ServingStatus_SERVING_STATUS_RELEASED)
	REQUEST_ERROR             RequestStatus = RequestStatus(pb.ServingStatus_SERVING_STATUS_ERROR)
)

type NodeRequestClient struct {
	ResourceManagerURI string
	conn               *grpc.ClientConn
	client             pb.NodeRequestingServiceClient
}

type NodeRequest struct {
	RequestID string
	NodeType  string
	NodeCount int
}

type LiveNode struct {
	NodeID          string
	NodeGRPCAddress string
	NodeStatus      NodeStatus
}

type NodeRequestResponse struct {
	RequestID     string
	NodeType      string
	NodeCount     int
	RequestStatus RequestStatus
	Nodes         []LiveNode
}

func NewNodeResourceManagerClient(resourceManagerURI string) *NodeRequestClient {
	return &NodeRequestClient{
		ResourceManagerURI: resourceManagerURI,
	}
}

func (n *NodeRequestClient) connect() error {
	err := error(nil)
	n.conn, err = grpc.NewClient(n.ResourceManagerURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	n.client = pb.NewNodeRequestingServiceClient(n.conn)
	return nil
}

func (n *NodeRequestClient) disconnect() error {
	if n.conn != nil {
		err := n.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (nodeRequestClient *NodeRequestClient) RequestNode(nodeType string, nodeCount int) (*NodeRequestResponse, error) {
	nodeRequest := &pb.NodeRequest{
		NodeType:  nodeType,
		NodeCount: uint32(nodeCount),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := nodeRequestClient.client.RequestNodes(ctx, nodeRequest)
	if err != nil {
		logger.Error("Failed to request node ", "CLIENT", err)
		return nil, err
	}
	if !res.Success {
		return nil, errors.New("failed to request node")
	}
	return &NodeRequestResponse{
		RequestID:     res.RequestHandler.RequestID,
		NodeType:      nodeType,
		NodeCount:     nodeCount,
		RequestStatus: REQUEST_REQUESTED,
		Nodes:         []LiveNode{},
	}, nil
}

func (nodeRequestClient *NodeRequestClient) GetNodeRequestStatus(request *NodeRequestResponse) (*NodeRequestResponse, error) {
	nodeRequest := &pb.RequestHandler{
		RequestID: request.RequestID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := nodeRequestClient.client.NodeRequestStatus(ctx, nodeRequest)
	if err != nil {
		logger.Error("Failed to get request status", "CLIENT", err)
		return nil, err
	}
	if !res.Success {
		return nil, errors.New("failed to get request status")
	}
	nodes := []LiveNode{}
	for _, node := range res.Nodes {
		nodes = append(nodes, LiveNode{
			NodeID:          node.NodeID,
			NodeGRPCAddress: node.NodeGRPCAddress,
			NodeStatus:      NodeStatus(node.NodeStatus),
		})
	}

	return &NodeRequestResponse{
		RequestID:     request.RequestID,
		NodeType:      request.NodeType,
		NodeCount:     request.NodeCount,
		RequestStatus: RequestStatus(res.ServingStatus),
		Nodes:         nodes,
	}, nil
}

func (nodeRequestClient *NodeRequestClient) ReleaseNodes(request *NodeRequestResponse) error {
	nodeRequest := &pb.RequestHandler{
		RequestID: request.RequestID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := nodeRequestClient.client.ReleaseNodes(ctx, nodeRequest)
	if err != nil {
		logger.Error("Failed to release nodes", "CLIENT", err)
		return err
	}
	if !res.Success {
		return errors.New("failed to release nodes")
	}
	return nil
}
