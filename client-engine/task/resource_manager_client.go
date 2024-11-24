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
	"fmt"
	"time"

	"github.com/google/uuid"
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

type ResourceManagerClient struct {
	ResourceManagerURI       string
	conn                     *grpc.ClientConn
	nodeRequestServiceClient pb.NodeRequestingServiceClient
	eventLogsServiceClient   pb.EventLogServiceClient
}

type NodeRequest struct {
	RequestID string
	NodeCount int
}

type LiveNode struct {
	NodeID          string
	NodeGRPCAddress string
	NodeStatus      NodeStatus
}

type NodeRequestResponse struct {
	RequestID     string
	NodeCount     int
	RequestStatus RequestStatus
	Nodes         []LiveNode
}

// EventLogs
type EventLog struct {
	EventID   string
	EventTime time.Time
	EventType string
	EventJSON string
}

func NewEventLog(eventType string, eventJSON string) *EventLog {
	eventID := fmt.Sprintf("evt-%s", uuid.New().String())

	return &EventLog{
		EventID:   eventID,
		EventTime: time.Now(),
		EventType: eventType,
		EventJSON: eventJSON,
	}
}

func NewNodeResourceManagerClient(resourceManagerURI string) *ResourceManagerClient {
	return &ResourceManagerClient{
		ResourceManagerURI: resourceManagerURI,
	}
}

func (n *ResourceManagerClient) connect() error {
	err := error(nil)
	n.conn, err = grpc.NewClient(n.ResourceManagerURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	n.nodeRequestServiceClient = pb.NewNodeRequestingServiceClient(n.conn)
	n.eventLogsServiceClient = pb.NewEventLogServiceClient(n.conn)
	return nil
}

func (n *ResourceManagerClient) disconnect() error {
	if n.conn != nil {
		err := n.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ResourceManagerClient *ResourceManagerClient) RequestNode(nodeCount int) (*NodeRequestResponse, error) {
	nodeRequest := &pb.NodeRequest{
		NodeCount: uint32(nodeCount),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := ResourceManagerClient.nodeRequestServiceClient.RequestNodes(ctx, nodeRequest)
	if err != nil {
		logger.Error("Failed to request node ", "CLIENT", err)
		return nil, err
	}
	if !res.Success {
		return nil, errors.New("failed to request node")
	}
	return &NodeRequestResponse{
		RequestID:     res.RequestHandler.RequestID,
		NodeCount:     nodeCount,
		RequestStatus: REQUEST_REQUESTED,
		Nodes:         []LiveNode{},
	}, nil
}

func (ResourceManagerClient *ResourceManagerClient) SendHeartbeat(requestID string) error {
	requestHandler := &pb.RequestHandler{
		RequestID: requestID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := ResourceManagerClient.nodeRequestServiceClient.SendRequestHeartbeat(ctx, requestHandler)
	if err != nil {
		logger.Error("Failed to send heartbeat", "CLIENT", err)
		return err
	}
	if !res.Success {
		logger.Error("Failed to send heartbeat", "CLIENT", errors.New("failed to send heartbeat"))
		return errors.New("failed to send heartbeat")
	}
	return nil
}

func (ResourceManagerClient *ResourceManagerClient) GetNodeRequestStatus(request *NodeRequestResponse) (*NodeRequestResponse, error) {
	nodeRequest := &pb.RequestHandler{
		RequestID: request.RequestID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := ResourceManagerClient.nodeRequestServiceClient.NodeRequestStatus(ctx, nodeRequest)
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
		NodeCount:     request.NodeCount,
		RequestStatus: RequestStatus(res.ServingStatus),
		Nodes:         nodes,
	}, nil
}

func (ResourceManagerClient *ResourceManagerClient) ReleaseNodes(request *NodeRequestResponse) error {
	nodeRequest := &pb.RequestHandler{
		RequestID: request.RequestID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := ResourceManagerClient.nodeRequestServiceClient.ReleaseNodes(ctx, nodeRequest)
	if err != nil {
		logger.Error("Failed to release nodes", "CLIENT", err)
		return err
	}
	if !res.Success {
		return errors.New("failed to release nodes")
	}
	return nil
}

func (ResourceManagerClient *ResourceManagerClient) LogEvent(event *EventLog) error {
	eventLog := &pb.EventLog{
		EventID:       event.EventID,
		EventTimeISO:  event.EventTime.Format(time.RFC3339),
		EventTypeName: event.EventType,
		EventJSON:     event.EventJSON,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res, err := ResourceManagerClient.eventLogsServiceClient.LogEvent(ctx, eventLog)
	if err != nil {
		logger.Error("Failed to log event", "CLIENT", err)
		return err
	}
	if !res.Success {
		return errors.New("failed to log event")
	}
	return nil
}
