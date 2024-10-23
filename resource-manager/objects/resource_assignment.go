package objects

import pb "resource-manager/grpc_resource_alloc_procotol"

type ResourceAssignmentStatusEnum pb.ServingStatus

const (
	// ResourceAssignmentStatusEnum
	ResourceAssignmentStatusEnum_RESOURCE_UNKNOWN           ResourceAssignmentStatusEnum = ResourceAssignmentStatusEnum(pb.ServingStatus_SERVING_STATUS_UNKNOWN)
	ResourceAssignmentStatusEnum_RESOURCE_REQUESTED         ResourceAssignmentStatusEnum = ResourceAssignmentStatusEnum(pb.ServingStatus_SERVING_STATUS_REQUESTED)
	ResourceAssignmentStatusEnum_RESOURCE_ALLOCATING        ResourceAssignmentStatusEnum = ResourceAssignmentStatusEnum(pb.ServingStatus_SERVING_STATUS_ALLOCATING)
	ResourceAssignmentStatusEnum_RESOURCE_ALLOCATED         ResourceAssignmentStatusEnum = ResourceAssignmentStatusEnum(pb.ServingStatus_SERVING_STATUS_ALLOCATED)
	ResourceAssignmentStatusEnum_RESOURCE_RELEASE_REQUESTED ResourceAssignmentStatusEnum = ResourceAssignmentStatusEnum(pb.ServingStatus_SERVING_STATUS_RELEASE_REQUESTED)
	ResourceAssignmentStatusEnum_RESOURCE_RELEASED          ResourceAssignmentStatusEnum = ResourceAssignmentStatusEnum(pb.ServingStatus_SERVING_STATUS_RELEASED)
	ResourceAssignmentStatusEnum_RESOURCE_PANIC             ResourceAssignmentStatusEnum = ResourceAssignmentStatusEnum(pb.ServingStatus_SERVING_STATUS_ERROR)
	TypeMarker_ResourceAssignmentStatusEnum                 TypeMarkerIndicator          = 0
)

var (
	// ResourceAssignmentStatus Mapping to string
	ResourceAssignmentStatusMapping = map[string]ResourceAssignmentStatusEnum{
		"UNKNOWN":           ResourceAssignmentStatusEnum_RESOURCE_UNKNOWN,
		"REQUESTED":         ResourceAssignmentStatusEnum_RESOURCE_REQUESTED,
		"ALLOCATING":        ResourceAssignmentStatusEnum_RESOURCE_ALLOCATING,
		"ALLOCATED":         ResourceAssignmentStatusEnum_RESOURCE_ALLOCATED,
		"RELEASE_REQUESTED": ResourceAssignmentStatusEnum_RESOURCE_RELEASE_REQUESTED,
		"RELEASED":          ResourceAssignmentStatusEnum_RESOURCE_RELEASED,
		"PANIC":             ResourceAssignmentStatusEnum_RESOURCE_PANIC,
	}
)

type ResourceAssignment struct {
	// This struct is used to store the resource assignment to a session
	RequestID             string `gorm:"primary_key"`
	NodeType              string
	NodeCount             uint32
	ServingStatus         ResourceAssignmentStatusEnum `gorm:"type:int"`
	RequestDetails        string                       `gorm:"type:text"`
	LastHeartbeatReceived int64
	RequestCreatedAt      int64
}

func (ra *ResourceAssignment) Sync()   { SyncGenericStruct(ra) }
func (ra *ResourceAssignment) Delete() { DeleteGenericStruct(ra) }

// Getter functions for ResourceAssignment
func ResourceAssignments() []*ResourceAssignment {
	db := GetDBInstance().db
	var ras []*ResourceAssignment
	result := db.Find(&ras)
	if result.Error != nil {
		return nil
	}
	return ras
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

func GetResourceAssignment(requestID string) (*ResourceAssignment, bool) {
	result := ResourceAssignmentsWithWhereClause("request_id = ?", requestID)
	if len(result) == 0 {
		return nil, false
	}
	return result[0], true
}

func GetResourceAssignmentsByStatus(status ResourceAssignmentStatusEnum) []*ResourceAssignment {
	return ResourceAssignmentsWithWhereClause("serving_status = ?", status)
}
