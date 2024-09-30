package main

import (
	pb "resource-manager/grpc_resource_alloc_procotol"
)

// enum
type ResourceRequestStatusEnum pb.ServingStatus
type LiveNodeStatusEnum pb.NodeStatus

const (
	ResourceRequestStatusEnum_RESOURCE_UNKNOWN           ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_UNKNOWN)
	ResourceRequestStatusEnum_RESOURCE_REQUESTED         ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_REQUESTED)
	ResourceRequestStatusEnum_RESOURCE_ALLOCATING        ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_ALLOCATING)
	ResourceRequestStatusEnum_RESOURCE_ALLOCATED         ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_ALLOCATED)
	ResourceRequestStatusEnum_RESOURCE_RELEASE_REQUESTED ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_RELEASE_REQUESTED)
	ResourceRequestStatusEnum_RESOURCE_RELEASED          ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_RELEASED)
	ResourceRequestStatusEnum_RESOURCE_PANIC             ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_ERROR)
)
const (
	LiveNodeStatusEnum_NODE_STATUS_UNKNOWN       LiveNodeStatusEnum = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_UNKNOWN)
	LiveNodeStatusEnum_NODE_STATUS_BOOTING       LiveNodeStatusEnum = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_BOOTING)
	LiveNodeStatusEnum_NODE_STATUS_READY         LiveNodeStatusEnum = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_READY)
	LiveNodeStatusEnum_NODE_STATUS_SHUTTING_DOWN LiveNodeStatusEnum = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_SHUTTING_DOWN)
	LiveNodeStatusEnum_NODE_STATUS_SHUTDOWN      LiveNodeStatusEnum = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_SHUTDOWN)
	LiveNodeStatusEnum_NODE_STATUS_PANIC         LiveNodeStatusEnum = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_ERROR)
)

type NodeCatalog struct {
	// The Node Catalog is a struct that holds all the information about the nodes
	// It is used to store all the information about the nodes in the system
	NodeType         string
	NumCores         int
	AmountOfMemory   int
	AmountOfStorage  int
	NodeMaxCapacity  int
	NodeWarmpoolSize int
	NodeIdleTimeout  int
}

type LiveNode struct {
	// The Live Node is a struct that holds all the live data of a node
	// It is used to store all the live data of a node
	NodeID          string
	NodeType        string
	NodeGRPCAddress string
	NodeStatus      LiveNodeStatusEnum
}

type ResourceAssignment struct {
	// This struct is used to store the resource assignment to a session
	RequestID     string
	NodeType      string
	NodeCount     uint32
	LiveNodes     []string
	ServingStatus ResourceRequestStatusEnum
}
