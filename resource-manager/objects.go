package main

import (
	pb "resource-manager/grpc_resource_alloc_procotol"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// enum
type ResourceRequestStatusEnum pb.ServingStatus
type LiveNodeStatusEnum pb.NodeStatus
type TypeMarkerIndicator int

const (
	// ResourceRequestStatusEnum
	ResourceRequestStatusEnum_RESOURCE_UNKNOWN           ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_UNKNOWN)
	ResourceRequestStatusEnum_RESOURCE_REQUESTED         ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_REQUESTED)
	ResourceRequestStatusEnum_RESOURCE_ALLOCATING        ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_ALLOCATING)
	ResourceRequestStatusEnum_RESOURCE_ALLOCATED         ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_ALLOCATED)
	ResourceRequestStatusEnum_RESOURCE_RELEASE_REQUESTED ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_RELEASE_REQUESTED)
	ResourceRequestStatusEnum_RESOURCE_RELEASED          ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_RELEASED)
	ResourceRequestStatusEnum_RESOURCE_PANIC             ResourceRequestStatusEnum = ResourceRequestStatusEnum(pb.ServingStatus_SERVING_STATUS_ERROR)
	TypeMarker_ResourceRequestStatusEnum                 TypeMarkerIndicator       = 0
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
	// ResourceRequestStatus Mapping to string
	ResourceRequestStatusMapping = map[string]ResourceRequestStatusEnum{
		"UNKNOWN":           ResourceRequestStatusEnum_RESOURCE_UNKNOWN,
		"REQUESTED":         ResourceRequestStatusEnum_RESOURCE_REQUESTED,
		"ALLOCATING":        ResourceRequestStatusEnum_RESOURCE_ALLOCATING,
		"ALLOCATED":         ResourceRequestStatusEnum_RESOURCE_ALLOCATED,
		"RELEASE_REQUESTED": ResourceRequestStatusEnum_RESOURCE_RELEASE_REQUESTED,
		"RELEASED":          ResourceRequestStatusEnum_RESOURCE_RELEASED,
		"PANIC":             ResourceRequestStatusEnum_RESOURCE_PANIC,
	}
)

// DB Struct (Singleton)
type DBInstance struct {
	db *gorm.DB
}

var dbInstance *DBInstance
var dbOnce sync.Once

func databaseMain() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("resource-manager.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&NodeCatalog{})
	db.AutoMigrate(&LiveNode{})
	db.AutoMigrate(&ResourceAssignment{})
	return db
}

func GetDBInstance() *DBInstance {
	dbOnce.Do(func() {
		db := databaseMain()
		dbInstance = &DBInstance{
			db: db,
		}
	})
	return dbInstance
}

func SyncGenericStruct(data interface{}) {
	db := GetDBInstance().db
	db.Save(data)
}

type NodeCatalog struct {
	// The Node Catalog is a struct that holds all the information about the nodes
	// It is used to store all the information about the nodes in the system
	NodeType         string `gorm:"primary_key"`
	NumCores         int
	AmountOfMemory   int
	AmountOfStorage  int
	NodeMaxCapacity  int
	NodeWarmpoolSize int
	NodeIdleTimeout  int
}

func (nc *NodeCatalog) Sync() { SyncGenericStruct(nc) }

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
}

func (ln *LiveNode) Sync() { SyncGenericStruct(ln) }

type ResourceAssignment struct {
	// This struct is used to store the resource assignment to a session
	RequestID             string `gorm:"primary_key"`
	NodeType              string
	NodeCount             uint32
	ServingStatus         ResourceRequestStatusEnum `gorm:"type:int"`
	LastHeartbeatReceived int64
}

func (ra *ResourceAssignment) Sync() { SyncGenericStruct(ra) }

func GetNodeCatalog(nodeType string) (*NodeCatalog, bool) {
	db := GetDBInstance().db
	var nc NodeCatalog
	result := db.First(&nc, "node_type = ?", nodeType)
	if result.Error != nil {
		return nil, false
	}
	return &nc, true
}

func GetLiveNode(nodeID string) (*LiveNode, bool) {
	db := GetDBInstance().db
	var ln LiveNode
	result := db.First(&ln, "node_id = ?", nodeID)
	if result.Error != nil {
		return nil, false
	}
	return &ln, true
}

func GetResourceAssignment(requestID string) (*ResourceAssignment, bool) {
	db := GetDBInstance().db
	var ra ResourceAssignment
	result := db.First(&ra, "request_id = ?", requestID)
	if result.Error != nil {
		return nil, false
	}
	return &ra, true
}

func GetLiveNodesByStatus(status LiveNodeStatusEnum) []*LiveNode {
	db := GetDBInstance().db
	var lns []*LiveNode
	result := db.Find(&lns, "node_status = ?", status)
	if result.Error != nil {
		return nil
	}
	return lns
}

func GetResourceAssignmentsByStatus(status ResourceRequestStatusEnum) []*ResourceAssignment {
	db := GetDBInstance().db
	var ras []*ResourceAssignment
	result := db.Find(&ras, "serving_status = ?", status)
	if result.Error != nil {
		return nil
	}
	return ras
}

func GetLiveNodesNotServingRequest(NodeType string) []*LiveNode {
	db := GetDBInstance().db
	var lns []*LiveNode
	// False or is null
	resultCondition := db.Where("node_type = ? AND (is_serving_request = ? OR serving_request_id IS NULL)", NodeType, false)
	result := resultCondition.Find(&lns)
	if result.Error != nil {
		return nil
	}
	return lns
}

func GetLiveNodesServingRequest(requestID string) []*LiveNode {
	db := GetDBInstance().db
	var lns []*LiveNode
	result := db.Find(&lns, "serving_request_id = ?", requestID)
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

func LiveNodes() []*LiveNode {
	db := GetDBInstance().db
	var lns []*LiveNode
	result := db.Find(&lns)
	if result.Error != nil {
		return nil
	}
	return lns
}

func ResourceAssignmentsWithWhereClause(whereClause string, args ...interface{}) []*ResourceAssignment {
	db := GetDBInstance().db
	var ras []*ResourceAssignment
	result := db.Where(whereClause, args...).Find(&ras)
	if result.Error != nil {
		return nil
	}
	return ras
}
func ResourceAssignments() []*ResourceAssignment {
	db := GetDBInstance().db
	var ras []*ResourceAssignment
	result := db.Find(&ras)
	if result.Error != nil {
		return nil
	}
	return ras
}

func CatalogsWithWhereClause(whereClause string, args ...interface{}) []*NodeCatalog {
	db := GetDBInstance().db
	var ncs []*NodeCatalog
	result := db.Where(whereClause, args...).Find(&ncs)
	if result.Error != nil {
		return nil
	}
	return ncs
}

func NodeCatalogs() []*NodeCatalog {
	db := GetDBInstance().db
	var ncs []*NodeCatalog
	result := db.Find(&ncs)
	if result.Error != nil {
		return nil
	}
	return ncs
}

func EnumToString(enum interface{}) string {
	// If the interface is a LiveNodeStatusEnum
	if _, ok := enum.(LiveNodeStatusEnum); ok {
		reverseMapping := make(map[LiveNodeStatusEnum]string)
		for k, v := range NodeStatusMapping {
			reverseMapping[v] = k
		}
		return reverseMapping[enum.(LiveNodeStatusEnum)]
	} else if _, ok := enum.(ResourceRequestStatusEnum); ok {
		reverseMapping := make(map[ResourceRequestStatusEnum]string)
		for k, v := range ResourceRequestStatusMapping {
			reverseMapping[v] = k
		}
		return reverseMapping[enum.(ResourceRequestStatusEnum)]
	}
	return ""
}

func StringToEnum(enumType TypeMarkerIndicator, str string) interface{} {
	if str == "" {
		return nil
	}
	if enumType == TypeMarker_ResourceRequestStatusEnum {
		return ResourceRequestStatusMapping[str]
	} else if enumType == TypeMarker_LiveNodeStatusEnum {
		return NodeStatusMapping[str]
	}
	return nil
}
