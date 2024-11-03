package main

import (
	"context"
	"net"
	pb "resource-manager/grpc_resource_alloc_procotol"
	"resource-manager/logger"
	obj "resource-manager/objects"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type ResourceAllocServer struct {
	pb.UnimplementedNodeRequestingServiceServer
	pb.UnimplementedNodeStatusUpdateServiceServer
}

func (s *ResourceAllocServer) RequestNodes(ctx context.Context, in *pb.NodeRequest) (*pb.NodeAllocationResponse, error) {
	// This function is used to handle the request for a node
	requestId := uuid.New().String()
	logger.Info("Request for node received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"request_id": requestId, "node_type": in.NodeType, "node_count": in.NodeCount})

	// Check if the node type is valid
	_, ok := obj.GetNodeCatalog(in.NodeType)
	if !ok {
		response := &pb.NodeAllocationResponse{
			Success:        false,
			ErrorMessage:   "Invalid node type",
			RequestHandler: nil,
		}
		return response, nil
	}

	// Create new assigment for the request
	now := time.Now()
	assignment := &obj.ResourceAssignment{
		RequestID:             requestId,
		NodeType:              in.NodeType,
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
			NodeType:        node.NodeType,
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
		response := &pb.UpdateOk{
			Success:      false,
			ErrorMessage: "Invalid request ID",
		}
		return response, nil
	}

	now := time.Now().Unix()
	// Update the last heartbeat received
	resourceAssignment.LastHeartbeatReceived = now
	resourceAssignment.Sync()

	response := &pb.UpdateOk{
		Success:      true,
		ErrorMessage: "",
	}

	return response, nil
}

func (s *ResourceAllocServer) UpdateNodeStatus(ctx context.Context, in *pb.LiveNode) (*pb.UpdateOk, error) {
	// Get the node from the live memory
	logger.Info("Request to update node status received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"node_id": in.NodeID})
	node, ok := obj.GetLiveNode(in.NodeID)
	if !ok {
		response := &pb.UpdateOk{
			Success:      false,
			ErrorMessage: "Node not found",
		}
		return response, nil
	}

	// Update the status of the node
	nodeNewStatus := obj.LiveNodeStatusEnum(in.NodeStatus)

	node.NodeStatus = nodeNewStatus
	node.NodeGRPCAddress = in.NodeGRPCAddress

	logger.Info("Node status updated", "RESOURCE_ALLOC_SERVER", logrus.Fields{"node_id": in.NodeID, "node_status": nodeNewStatus})

	node.Sync()
	response := &pb.UpdateOk{
		Success:      true,
		ErrorMessage: "",
	}

	return response, nil
}

func (s *ResourceAllocServer) NodeHeartbeat(ctx context.Context, in *pb.NodeHeartbeatRequest) (*pb.UpdateOk, error) {
	// Get the node from the live memory
	logger.Info("Request to update node heartbeat received", "RESOURCE_ALLOC_SERVER", logrus.Fields{"node_id": in.NodeID})
	node, ok := obj.GetLiveNode(in.NodeID)
	if !ok {
		response := &pb.UpdateOk{
			Success:      false,
			ErrorMessage: "Node not found",
		}
		return response, nil
	}

	// Update the last heartbeat received
	node.LastHeartbeatReceived = time.Now().Unix()
	node.Sync()

	response := &pb.UpdateOk{
		Success:      true,
		ErrorMessage: "",
	}

	return response, nil
}

func NewResourceAllocServer(exit chan bool, wg *sync.WaitGroup, port string) error {
	port = ":" + port
	s := grpc.NewServer()

	recServer := &ResourceAllocServer{}

	pb.RegisterNodeRequestingServiceServer(s, recServer)
	pb.RegisterNodeStatusUpdateServiceServer(s, recServer)

	lis, err := net.Listen("tcp", port)
	if err != nil {
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
