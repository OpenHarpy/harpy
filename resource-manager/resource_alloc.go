package main

import (
	"context"
	"net"
	pb "resource-manager/grpc_resource_alloc_procotol"
	"resource-manager/logger"
	obj "resource-manager/objects"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type ResourceAllocServer struct {
	EventLoggingEnabled bool
	EventLoggingMaxSize int
	pb.UnimplementedNodeRequestingServiceServer
	pb.UnimplementedNodeStatusUpdateServiceServer
	pb.UnimplementedEventLogServiceServer
}

func (s *ResourceAllocServer) RequestNodes(ctx context.Context, in *pb.NodeRequest) (*pb.NodeAllocationResponse, error) {
	// This function is used to handle the request for a node
	requestId := uuid.New().String()
	logger.Info("Request for node received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"request_id": requestId, "node_count": in.NodeCount})

	// Create new assigment for the request
	now := time.Now()
	assignment := &obj.ResourceAssignment{
		RequestID:             requestId,
		NodeCount:             in.NodeCount,
		ServingStatus:         obj.ResourceAssignmentStatusEnum_RESOURCE_REQUESTED,
		RequestCreatedAt:      now.Unix(),
		LastHeartbeatReceived: now.Unix(),
	}
	assignment.Sync()

	reqHandler := &pb.RequestHandler{
		RequestID: requestId,
	}
	response := &pb.NodeAllocationResponse{
		Success:        true,
		ErrorMessage:   "",
		RequestHandler: reqHandler,
	}

	return response, nil
}

func (s *ResourceAllocServer) ReleaseNodes(ctx context.Context, in *pb.RequestHandler) (*pb.NodeReleaseResponse, error) {
	// This function is used to handle the release of nodes
	// Check if the request ID is valid
	logger.Info("Request to release nodes received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"request_id": in.RequestID})
	resourceAssignment, ok := obj.GetResourceAssignment(in.RequestID)
	if !ok {
		response := &pb.NodeReleaseResponse{
			Success:      false,
			ErrorMessage: "Invalid request ID",
		}
		return response, nil
	}

	// Mark the assigment as RELEASE_REQUESTED
	resourceAssignment.ServingStatus = obj.ResourceAssignmentStatusEnum_RESOURCE_RELEASE_REQUESTED
	resourceAssignment.Sync()

	response := &pb.NodeReleaseResponse{
		Success:      true,
		ErrorMessage: "",
	}

	return response, nil
}

func (s *ResourceAllocServer) NodeRequestStatus(ctx context.Context, in *pb.RequestHandler) (*pb.NodeRequestStatusResponse, error) {
	// This function is used to get the status of the request
	// Check if the request ID is valid
	logger.Debug("Request for node status received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"request_id": in.RequestID})
	resourceAssignment, ok := obj.GetResourceAssignment(in.RequestID)
	if !ok {
		response := &pb.NodeRequestStatusResponse{
			Success:      false,
			ErrorMessage: "Invalid request ID",
		}
		return response, nil
	}

	// Make NODES payload
	nodes := obj.GetLiveNodesServingRequest(resourceAssignment.RequestID)
	nodesReturn := []*pb.LiveNode{}
	for _, node := range nodes {
		nodesReturn = append(nodesReturn, &pb.LiveNode{
			NodeID:          node.NodeID,
			NodeGRPCAddress: node.NodeGRPCAddress,
			NodeStatus:      pb.NodeStatus(node.NodeStatus),
		})
	}

	response := &pb.NodeRequestStatusResponse{
		Success:       true,
		ErrorMessage:  "",
		ServingStatus: pb.ServingStatus(resourceAssignment.ServingStatus),
		Nodes:         nodesReturn,
	}

	return response, nil
}

func (s *ResourceAllocServer) SendRequestHeartbeat(ctx context.Context, in *pb.RequestHandler) (*pb.UpdateOk, error) {
	// This function is used to handle the heartbeat request
	logger.Info("Request heartbeat received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"request_id": in.RequestID})
	resourceAssignment, ok := obj.GetResourceAssignment(in.RequestID)
	if !ok {
		return nil, status.Errorf(400, "Invalid request ID")
	}

	now := time.Now().Unix()
	// Update the last heartbeat received
	resourceAssignment.LastHeartbeatReceived = now
	resourceAssignment.Sync()

	response := &pb.UpdateOk{
		Success: true,
	}

	return response, nil
}

func (s *ResourceAllocServer) UpdateNodeStatus(ctx context.Context, in *pb.LiveNode) (*pb.UpdateOk, error) {
	// Get the node from the live memory
	logger.Info("Request to update node status received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"node_id": in.NodeID})
	node, ok := obj.GetLiveNode(in.NodeID)
	if !ok {
		return nil, status.Errorf(400, "Node not found")
	}

	// Update the status of the node
	nodeNewStatus := obj.LiveNodeStatusEnum(in.NodeStatus)

	node.NodeStatus = nodeNewStatus
	node.NodeGRPCAddress = in.NodeGRPCAddress

	logger.Info("Node status updated", "RESOURCE_ALLOC_SERVER", logrus.Fields{"node_id": in.NodeID, "node_status": nodeNewStatus})

	node.Sync()
	response := &pb.UpdateOk{
		Success: true,
	}

	return response, nil
}

func (s *ResourceAllocServer) NodeHeartbeat(ctx context.Context, in *pb.NodeHeartbeatRequest) (*pb.UpdateOk, error) {
	// Get the node from the live memory
	logger.Info("Request to update node heartbeat received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"node_id": in.NodeID})
	node, ok := obj.GetLiveNode(in.NodeID)
	if !ok {
		return nil, status.Errorf(400, "Node not found")
	}

	// Update the last heartbeat received
	node.LastHeartbeatReceived = time.Now().Unix()
	node.Sync()

	response := &pb.UpdateOk{
		Success: true,
	}

	return response, nil
}

func (s *ResourceAllocServer) LogEvent(ctx context.Context, in *pb.EventLog) (*pb.UpdateOk, error) {
	// Check if the eventLogging is enabled
	if !s.EventLoggingEnabled {
		response := &pb.UpdateOk{
			Success: false,
		}
		return response, nil
	}
	// Basic validation
	eventTime, err := time.Parse(time.RFC3339, in.EventTimeISO)
	if err != nil {
		logger.Error("Failed to parse event time", "RESOURCE_ALLOC_SERVER", err)
		return nil, status.Errorf(400, "Failed to parse event time")
	}
	if in.EventID == "" {
		logger.Error("Event ID is empty", "RESOURCE_ALLOC_SERVER", nil)
		return nil, status.Errorf(400, "Event ID is empty")
	}
	if in.EventTypeName == "" {
		logger.Error("Event type name is empty", "RESOURCE_ALLOC_SERVER", nil)
		return nil, status.Errorf(400, "Event type name is empty")
	}
	if in.EventGroupID == "" {
		logger.Error("Event group name is empty", "RESOURCE_ALLOC_SERVER", nil)
		return nil, status.Errorf(400, "Event group name is empty")
	}
	if in.EventJSON == "" {
		logger.Error("Event JSON is empty", "RESOURCE_ALLOC_SERVER", nil)
		return nil, status.Errorf(400, "Event JSON is empty")
	}

	// This function is used to log events
	logger.Debug("Request to log event received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"event": in.EventID})
	event := &obj.EventLogEntry{
		EventLogID:   in.EventID,
		EventLogType: in.EventTypeName,
		EventGroupID: in.EventGroupID,
		EventLogTime: eventTime.Unix(),
		EventLogJSON: in.EventJSON,
	}
	event.Sync()

	obj.CleanupEventLog(s.EventLoggingMaxSize)

	response := &pb.UpdateOk{
		Success: true,
	}

	return response, nil
}

func NewResourceAllocServer(exit chan bool, wg *sync.WaitGroup, port string) error {
	wg.Add(1)
	port = ":" + port
	s := grpc.NewServer()

	// Get the configs for event logging
	eventLoggingMaxEntries := obj.GetConfigWithDefault("harpy.resourceManager.eventLog.maxEntries", "1000")
	eventLoggingEnabled := obj.GetConfigWithDefault("harpy.resourceManager.eventLog.enabled", "true")
	eventLoggingMaxEntriesInt, err := strconv.ParseInt(eventLoggingMaxEntries, 10, 64)
	if err != nil {
		logger.Error("Failed to convert maxEntries to int", "RESOURCE_ALLOC_SERVER", err)
		eventLoggingMaxEntriesInt = 1000
	}

	recServer := &ResourceAllocServer{
		EventLoggingEnabled: eventLoggingEnabled == "true",
		EventLoggingMaxSize: int(eventLoggingMaxEntriesInt),
	}

	pb.RegisterNodeRequestingServiceServer(s, recServer)
	pb.RegisterNodeStatusUpdateServiceServer(s, recServer)
	pb.RegisterEventLogServiceServer(s, recServer)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		defer wg.Done()
		logger.Error("failed to listen", "RESOURCE_ALLOC_SERVER", err)
		return err
	} else {
		go func() {
			if err := s.Serve(lis); err != nil {
				logger.Error("failed_to_serve", "SERVER", err)
				defer wg.Done()
			}
		}()

		// Goroutine to listen for exit signal
		go func() {
			<-exit
			logger.Info("Stopping resource allocation server", "RESOURCE_ALLOC_SERVER", nil)
			s.GracefulStop()
			logger.Info("Resource allocation server stopped", "RESOURCE_ALLOC_SERVER", nil)
			defer wg.Done()
		}()
	}
	return nil
}
