package objects

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
	db.AutoMigrate(&Config{})
	db.AutoMigrate(&Provider{})
	db.AutoMigrate(&ProviderProps{})
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

func DeleteGenericStruct(data interface{}) {
	db := GetDBInstance().db
	db.Delete(data)
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
	NodeCreatedAt         int64
}

func (ln *LiveNode) Sync()   { SyncGenericStruct(ln) }
func (ln *LiveNode) Delete() { DeleteGenericStruct(ln) }

type ResourceAssignment struct {
	// This struct is used to store the resource assignment to a session
	RequestID             string `gorm:"primary_key"`
	NodeType              string
	NodeCount             uint32
	ServingStatus         ResourceRequestStatusEnum `gorm:"type:int"`
	RequestDetails        string                    `gorm:"type:text"`
	LastHeartbeatReceived int64
	RequestCreatedAt      int64
}

func (ra *ResourceAssignment) Sync()   { SyncGenericStruct(ra) }
func (ra *ResourceAssignment) Delete() { DeleteGenericStruct(ra) }

type Config struct {
	// This struct is used to store the configuration of the resource manager
	// It is used to store the configuration of the resource manager
	ConfigKey   string `gorm:"primary_key"`
	ConfigValue string `gorm:"type:text"`
}

func (c *Config) Sync()   { SyncGenericStruct(c) }
func (c *Config) Delete() { DeleteGenericStruct(c) }

func GetConfig(configKey string) (*Config, bool) {
	db := GetDBInstance().db
	var c Config
	result := db.First(&c, "config_key = ?", configKey)
	if result.Error != nil {
		return nil, false
	}
	return &c, true
}

type Provider struct {
	ProviderName        string `gorm:"primary_key"`
	ProviderDescription string `gorm:"type:text"`
}

func (p *Provider) Sync()   { SyncGenericStruct(p) }
func (p *Provider) Delete() { DeleteGenericStruct(p) }

type ProviderProps struct {
	ProviderName string `gorm:"primary_key"`
	PropKey      string `gorm:"primary_key"`
	PropValue    string `gorm:"type:text"`
}

func (p *ProviderProps) Sync()   { SyncGenericStruct(p) }
func (p *ProviderProps) Delete() { DeleteGenericStruct(p) }

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

func GetLiveNodes() []*LiveNode {
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

func CatalogsWithWarmPoolNotMet() []*NodeCatalog {
	db := GetDBInstance().db
	var ncs []*NodeCatalog

	// Subquery to count live nodes per node type
	subQuery := db.Model(&LiveNode{}).
		Select("node_type, COUNT(*) as live_node_count").
		Group("node_type")

	// Join NodeCatalog with the subquery and filter based on NodeWarmpoolSize
	result := db.Table("node_catalogs").
		Select("node_catalogs.*").
		Joins("LEFT JOIN (?) AS live_nodes ON node_catalogs.node_type = live_nodes.node_type", subQuery).
		Where("node_catalogs.node_warmpool_size > IFNULL(live_nodes.live_node_count, 0)").
		Find(&ncs)

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

func GetProviders() []*Provider {
	db := GetDBInstance().db
	var providers []*Provider
	result := db.Find(&providers)
	if result.Error != nil {
		return nil
	}
	return providers
}
