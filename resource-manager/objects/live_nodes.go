package objects

import pb "resource-manager/grpc_resource_alloc_procotol"

type LiveNodeStatusEnum pb.NodeStatus

const (
	// LiveNodeStatusEnum
	LiveNodeStatusEnum_NODE_STATUS_UNKNOWN       LiveNodeStatusEnum  = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_UNKNOWN)
	LiveNodeStatusEnum_NODE_STATUS_BOOTING       LiveNodeStatusEnum  = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_BOOTING)
	LiveNodeStatusEnum_NODE_STATUS_READY         LiveNodeStatusEnum  = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_READY)
	LiveNodeStatusEnum_NODE_STATUS_SHUTTING_DOWN LiveNodeStatusEnum  = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_SHUTTING_DOWN)
	LiveNodeStatusEnum_NODE_STATUS_SHUTDOWN      LiveNodeStatusEnum  = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_SHUTDOWN)
	LiveNodeStatusEnum_NODE_STATUS_PANIC         LiveNodeStatusEnum  = LiveNodeStatusEnum(pb.NodeStatus_NODE_STATUS_ERROR)
	TypeMarker_LiveNodeStatusEnum                TypeMarkerIndicator = 1
)

var (
	// NodeStatus Mapping to string
	NodeStatusMapping = map[string]LiveNodeStatusEnum{
		"UNKNOWN":       LiveNodeStatusEnum_NODE_STATUS_UNKNOWN,
		"BOOTING":       LiveNodeStatusEnum_NODE_STATUS_BOOTING,
		"READY":         LiveNodeStatusEnum_NODE_STATUS_READY,
		"SHUTTING_DOWN": LiveNodeStatusEnum_NODE_STATUS_SHUTTING_DOWN,
		"SHUTDOWN":      LiveNodeStatusEnum_NODE_STATUS_SHUTDOWN,
		"PANIC":         LiveNodeStatusEnum_NODE_STATUS_PANIC,
	}
)

type LiveNode struct {
	// The Live Node is a struct that holds all the live data of a node
	// It is used to store all the live data of a node
	NodeID                string             `gorm:"primary_key"`
	NodeType              string             `gorm:"type:text"`
	NodeGRPCAddress       string             `gorm:"type:text"`
	NodeStatus            LiveNodeStatusEnum `gorm:"type:int"`
	LastHeartbeatReceived int64
	IsServingRequest      bool `gorm:"type:boolean"`
	ServingRequestID      string
	NodeCreatedAt         int64
}

func (ln *LiveNode) Sync()   { SyncGenericStruct(ln) }
func (ln *LiveNode) Delete() { DeleteGenericStruct(ln) }

// Getter functions for LiveNode
func GetLiveNodes() []*LiveNode {
	db := GetDBInstance().db
	var lns []*LiveNode
	result := db.Find(&lns)
	if result.Error != nil {
		return nil
	}
	return lns
}

func LiveNodesWithWhereClause(whereClause string, args ...interface{}) []*LiveNode {
	db := GetDBInstance().db
	var lns []*LiveNode
	result := db.Where(whereClause, args...).Find(&lns)
	if result.Error != nil {
		return nil
	}
	return lns
}

func GetLiveNode(nodeID string) (*LiveNode, bool) {
	result := LiveNodesWithWhereClause("node_id = ?", nodeID)
	if len(result) == 0 {
		return nil, false
	}
	return result[0], true
}

func GetLiveNodesByStatus(status LiveNodeStatusEnum) []*LiveNode {
	return LiveNodesWithWhereClause("node_status = ?", status)
}

func GetLiveNodesNotServingRequest(NodeType string) []*LiveNode {
	return LiveNodesWithWhereClause("node_type = ? AND (is_serving_request = ? OR serving_request_id IS NULL)", NodeType, false)
}

func GetLiveNodesServingRequest(requestID string) []*LiveNode {
	return LiveNodesWithWhereClause("serving_request_id = ?", requestID)
}
