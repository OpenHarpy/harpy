// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.12.4
// source: grpc_resource_alloc_procotol/resourceallocprotocol.proto

package grpc_resource_alloc_procotol

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type NodeStatus int32

const (
	NodeStatus_NODE_STATUS_UNKNOWN       NodeStatus = 0
	NodeStatus_NODE_STATUS_BOOTING       NodeStatus = 1
	NodeStatus_NODE_STATUS_READY         NodeStatus = 2
	NodeStatus_NODE_STATUS_SHUTTING_DOWN NodeStatus = 3
	NodeStatus_NODE_STATUS_SHUTDOWN      NodeStatus = 4
	NodeStatus_NODE_STATUS_ERROR         NodeStatus = 5
)

// Enum value maps for NodeStatus.
var (
	NodeStatus_name = map[int32]string{
		0: "NODE_STATUS_UNKNOWN",
		1: "NODE_STATUS_BOOTING",
		2: "NODE_STATUS_READY",
		3: "NODE_STATUS_SHUTTING_DOWN",
		4: "NODE_STATUS_SHUTDOWN",
		5: "NODE_STATUS_ERROR",
	}
	NodeStatus_value = map[string]int32{
		"NODE_STATUS_UNKNOWN":       0,
		"NODE_STATUS_BOOTING":       1,
		"NODE_STATUS_READY":         2,
		"NODE_STATUS_SHUTTING_DOWN": 3,
		"NODE_STATUS_SHUTDOWN":      4,
		"NODE_STATUS_ERROR":         5,
	}
)

func (x NodeStatus) Enum() *NodeStatus {
	p := new(NodeStatus)
	*p = x
	return p
}

func (x NodeStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NodeStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_enumTypes[0].Descriptor()
}

func (NodeStatus) Type() protoreflect.EnumType {
	return &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_enumTypes[0]
}

func (x NodeStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NodeStatus.Descriptor instead.
func (NodeStatus) EnumDescriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{0}
}

type ServingStatus int32

const (
	ServingStatus_SERVING_STATUS_REQUESTED         ServingStatus = 0
	ServingStatus_SERVING_STATUS_ALLOCATING        ServingStatus = 1
	ServingStatus_SERVING_STATUS_ALLOCATED         ServingStatus = 2
	ServingStatus_SERVING_STATUS_RELEASE_REQUESTED ServingStatus = 3
	ServingStatus_SERVING_STATUS_RELEASED          ServingStatus = 4
	ServingStatus_SERVING_STATUS_UNKNOWN           ServingStatus = 5
	ServingStatus_SERVING_STATUS_ERROR             ServingStatus = 6
)

// Enum value maps for ServingStatus.
var (
	ServingStatus_name = map[int32]string{
		0: "SERVING_STATUS_REQUESTED",
		1: "SERVING_STATUS_ALLOCATING",
		2: "SERVING_STATUS_ALLOCATED",
		3: "SERVING_STATUS_RELEASE_REQUESTED",
		4: "SERVING_STATUS_RELEASED",
		5: "SERVING_STATUS_UNKNOWN",
		6: "SERVING_STATUS_ERROR",
	}
	ServingStatus_value = map[string]int32{
		"SERVING_STATUS_REQUESTED":         0,
		"SERVING_STATUS_ALLOCATING":        1,
		"SERVING_STATUS_ALLOCATED":         2,
		"SERVING_STATUS_RELEASE_REQUESTED": 3,
		"SERVING_STATUS_RELEASED":          4,
		"SERVING_STATUS_UNKNOWN":           5,
		"SERVING_STATUS_ERROR":             6,
	}
)

func (x ServingStatus) Enum() *ServingStatus {
	p := new(ServingStatus)
	*p = x
	return p
}

func (x ServingStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ServingStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_enumTypes[1].Descriptor()
}

func (ServingStatus) Type() protoreflect.EnumType {
	return &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_enumTypes[1]
}

func (x ServingStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ServingStatus.Descriptor instead.
func (ServingStatus) EnumDescriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{1}
}

// Node Requests
type RequestHandler struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestID string `protobuf:"bytes,1,opt,name=RequestID,proto3" json:"RequestID,omitempty"`
}

func (x *RequestHandler) Reset() {
	*x = RequestHandler{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestHandler) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestHandler) ProtoMessage() {}

func (x *RequestHandler) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestHandler.ProtoReflect.Descriptor instead.
func (*RequestHandler) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{0}
}

func (x *RequestHandler) GetRequestID() string {
	if x != nil {
		return x.RequestID
	}
	return ""
}

type NodeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeType  string `protobuf:"bytes,1,opt,name=NodeType,proto3" json:"NodeType,omitempty"`
	NodeCount uint32 `protobuf:"varint,2,opt,name=NodeCount,proto3" json:"NodeCount,omitempty"`
}

func (x *NodeRequest) Reset() {
	*x = NodeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeRequest) ProtoMessage() {}

func (x *NodeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeRequest.ProtoReflect.Descriptor instead.
func (*NodeRequest) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{1}
}

func (x *NodeRequest) GetNodeType() string {
	if x != nil {
		return x.NodeType
	}
	return ""
}

func (x *NodeRequest) GetNodeCount() uint32 {
	if x != nil {
		return x.NodeCount
	}
	return 0
}

type NodeAllocationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success        bool            `protobuf:"varint,1,opt,name=Success,proto3" json:"Success,omitempty"`
	ErrorMessage   string          `protobuf:"bytes,2,opt,name=ErrorMessage,proto3" json:"ErrorMessage,omitempty"`
	RequestHandler *RequestHandler `protobuf:"bytes,3,opt,name=RequestHandler,proto3" json:"RequestHandler,omitempty"`
}

func (x *NodeAllocationResponse) Reset() {
	*x = NodeAllocationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeAllocationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeAllocationResponse) ProtoMessage() {}

func (x *NodeAllocationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeAllocationResponse.ProtoReflect.Descriptor instead.
func (*NodeAllocationResponse) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{2}
}

func (x *NodeAllocationResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *NodeAllocationResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *NodeAllocationResponse) GetRequestHandler() *RequestHandler {
	if x != nil {
		return x.RequestHandler
	}
	return nil
}

type NodeReleaseRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestHandler *RequestHandler `protobuf:"bytes,1,opt,name=RequestHandler,proto3" json:"RequestHandler,omitempty"`
	ReleaseMessage string          `protobuf:"bytes,2,opt,name=ReleaseMessage,proto3" json:"ReleaseMessage,omitempty"`
}

func (x *NodeReleaseRequest) Reset() {
	*x = NodeReleaseRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeReleaseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeReleaseRequest) ProtoMessage() {}

func (x *NodeReleaseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeReleaseRequest.ProtoReflect.Descriptor instead.
func (*NodeReleaseRequest) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{3}
}

func (x *NodeReleaseRequest) GetRequestHandler() *RequestHandler {
	if x != nil {
		return x.RequestHandler
	}
	return nil
}

func (x *NodeReleaseRequest) GetReleaseMessage() string {
	if x != nil {
		return x.ReleaseMessage
	}
	return ""
}

type NodeReleaseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success      bool   `protobuf:"varint,1,opt,name=Success,proto3" json:"Success,omitempty"`
	ErrorMessage string `protobuf:"bytes,2,opt,name=ErrorMessage,proto3" json:"ErrorMessage,omitempty"`
}

func (x *NodeReleaseResponse) Reset() {
	*x = NodeReleaseResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeReleaseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeReleaseResponse) ProtoMessage() {}

func (x *NodeReleaseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeReleaseResponse.ProtoReflect.Descriptor instead.
func (*NodeReleaseResponse) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{4}
}

func (x *NodeReleaseResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *NodeReleaseResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

type LiveNode struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeID          string     `protobuf:"bytes,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
	NodeType        string     `protobuf:"bytes,2,opt,name=NodeType,proto3" json:"NodeType,omitempty"`
	NodeGRPCAddress string     `protobuf:"bytes,3,opt,name=NodeGRPCAddress,proto3" json:"NodeGRPCAddress,omitempty"`
	NodeStatus      NodeStatus `protobuf:"varint,4,opt,name=NodeStatus,proto3,enum=proto.NodeStatus" json:"NodeStatus,omitempty"`
}

func (x *LiveNode) Reset() {
	*x = LiveNode{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LiveNode) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LiveNode) ProtoMessage() {}

func (x *LiveNode) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LiveNode.ProtoReflect.Descriptor instead.
func (*LiveNode) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{5}
}

func (x *LiveNode) GetNodeID() string {
	if x != nil {
		return x.NodeID
	}
	return ""
}

func (x *LiveNode) GetNodeType() string {
	if x != nil {
		return x.NodeType
	}
	return ""
}

func (x *LiveNode) GetNodeGRPCAddress() string {
	if x != nil {
		return x.NodeGRPCAddress
	}
	return ""
}

func (x *LiveNode) GetNodeStatus() NodeStatus {
	if x != nil {
		return x.NodeStatus
	}
	return NodeStatus_NODE_STATUS_UNKNOWN
}

type NodeHeartbeatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeID string `protobuf:"bytes,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
}

func (x *NodeHeartbeatRequest) Reset() {
	*x = NodeHeartbeatRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeHeartbeatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeHeartbeatRequest) ProtoMessage() {}

func (x *NodeHeartbeatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeHeartbeatRequest.ProtoReflect.Descriptor instead.
func (*NodeHeartbeatRequest) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{6}
}

func (x *NodeHeartbeatRequest) GetNodeID() string {
	if x != nil {
		return x.NodeID
	}
	return ""
}

type NodeRequestStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success       bool          `protobuf:"varint,1,opt,name=Success,proto3" json:"Success,omitempty"`
	ErrorMessage  string        `protobuf:"bytes,2,opt,name=ErrorMessage,proto3" json:"ErrorMessage,omitempty"`
	ServingStatus ServingStatus `protobuf:"varint,4,opt,name=ServingStatus,proto3,enum=proto.ServingStatus" json:"ServingStatus,omitempty"`
	Nodes         []*LiveNode   `protobuf:"bytes,5,rep,name=Nodes,proto3" json:"Nodes,omitempty"`
}

func (x *NodeRequestStatusResponse) Reset() {
	*x = NodeRequestStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeRequestStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeRequestStatusResponse) ProtoMessage() {}

func (x *NodeRequestStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeRequestStatusResponse.ProtoReflect.Descriptor instead.
func (*NodeRequestStatusResponse) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{7}
}

func (x *NodeRequestStatusResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *NodeRequestStatusResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *NodeRequestStatusResponse) GetServingStatus() ServingStatus {
	if x != nil {
		return x.ServingStatus
	}
	return ServingStatus_SERVING_STATUS_REQUESTED
}

func (x *NodeRequestStatusResponse) GetNodes() []*LiveNode {
	if x != nil {
		return x.Nodes
	}
	return nil
}

// Update OK message
type UpdateOk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success      bool   `protobuf:"varint,1,opt,name=Success,proto3" json:"Success,omitempty"`
	ErrorMessage string `protobuf:"bytes,2,opt,name=ErrorMessage,proto3" json:"ErrorMessage,omitempty"`
}

func (x *UpdateOk) Reset() {
	*x = UpdateOk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateOk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateOk) ProtoMessage() {}

func (x *UpdateOk) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateOk.ProtoReflect.Descriptor instead.
func (*UpdateOk) Descriptor() ([]byte, []int) {
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP(), []int{8}
}

func (x *UpdateOk) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *UpdateOk) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

var File_grpc_resource_alloc_procotol_resourceallocprotocol_proto protoreflect.FileDescriptor

var file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDesc = []byte{
	0x0a, 0x38, 0x67, 0x72, 0x70, 0x63, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f,
	0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x5f, 0x70, 0x72, 0x6f, 0x63, 0x6f, 0x74, 0x6f, 0x6c, 0x2f, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x2e, 0x0a, 0x0e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61, 0x6e, 0x64,
	0x6c, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49,
	0x44, 0x22, 0x47, 0x0a, 0x0b, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x4e, 0x6f, 0x64, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x4e, 0x6f, 0x64, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x09,
	0x4e, 0x6f, 0x64, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x09, 0x4e, 0x6f, 0x64, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x95, 0x01, 0x0a, 0x16, 0x4e,
	0x6f, 0x64, 0x65, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12,
	0x22, 0x0a, 0x0c, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x3d, 0x0a, 0x0e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61,
	0x6e, 0x64, 0x6c, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61, 0x6e, 0x64, 0x6c,
	0x65, 0x72, 0x52, 0x0e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61, 0x6e, 0x64, 0x6c,
	0x65, 0x72, 0x22, 0x7b, 0x0a, 0x12, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x0e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x52, 0x0e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x12, 0x26, 0x0a, 0x0e, 0x52, 0x65, 0x6c, 0x65, 0x61,
	0x73, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0e, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22,
	0x53, 0x0a, 0x13, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x12, 0x22, 0x0a, 0x0c, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x22, 0x9b, 0x01, 0x0a, 0x08, 0x4c, 0x69, 0x76, 0x65, 0x4e, 0x6f, 0x64,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x08, 0x4e, 0x6f, 0x64,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x4e, 0x6f, 0x64,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x28, 0x0a, 0x0f, 0x4e, 0x6f, 0x64, 0x65, 0x47, 0x52, 0x50,
	0x43, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f,
	0x4e, 0x6f, 0x64, 0x65, 0x47, 0x52, 0x50, 0x43, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12,
	0x31, 0x0a, 0x0a, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x11, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x6f, 0x64, 0x65,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0a, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x22, 0x2e, 0x0a, 0x14, 0x4e, 0x6f, 0x64, 0x65, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62,
	0x65, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x4e, 0x6f,
	0x64, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x4e, 0x6f, 0x64, 0x65,
	0x49, 0x44, 0x22, 0xbc, 0x01, 0x0a, 0x19, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x3a,
	0x0a, 0x0d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0d, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x25, 0x0a, 0x05, 0x4e, 0x6f,
	0x64, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x4c, 0x69, 0x76, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x05, 0x4e, 0x6f, 0x64, 0x65,
	0x73, 0x22, 0x48, 0x0a, 0x08, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4f, 0x6b, 0x12, 0x18, 0x0a,
	0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07,
	0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2a, 0xa5, 0x01, 0x0a, 0x0a,
	0x4e, 0x6f, 0x64, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x17, 0x0a, 0x13, 0x4e, 0x4f,
	0x44, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57,
	0x4e, 0x10, 0x00, 0x12, 0x17, 0x0a, 0x13, 0x4e, 0x4f, 0x44, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54,
	0x55, 0x53, 0x5f, 0x42, 0x4f, 0x4f, 0x54, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x15, 0x0a, 0x11,
	0x4e, 0x4f, 0x44, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x41, 0x44,
	0x59, 0x10, 0x02, 0x12, 0x1d, 0x0a, 0x19, 0x4e, 0x4f, 0x44, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54,
	0x55, 0x53, 0x5f, 0x53, 0x48, 0x55, 0x54, 0x54, 0x49, 0x4e, 0x47, 0x5f, 0x44, 0x4f, 0x57, 0x4e,
	0x10, 0x03, 0x12, 0x18, 0x0a, 0x14, 0x4e, 0x4f, 0x44, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55,
	0x53, 0x5f, 0x53, 0x48, 0x55, 0x54, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x04, 0x12, 0x15, 0x0a, 0x11,
	0x4e, 0x4f, 0x44, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x45, 0x52, 0x52, 0x4f,
	0x52, 0x10, 0x05, 0x2a, 0xe3, 0x01, 0x0a, 0x0d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x6e, 0x67, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1c, 0x0a, 0x18, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e, 0x47,
	0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x45,
	0x44, 0x10, 0x00, 0x12, 0x1d, 0x0a, 0x19, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e, 0x47, 0x5f, 0x53,
	0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x41, 0x4c, 0x4c, 0x4f, 0x43, 0x41, 0x54, 0x49, 0x4e, 0x47,
	0x10, 0x01, 0x12, 0x1c, 0x0a, 0x18, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e, 0x47, 0x5f, 0x53, 0x54,
	0x41, 0x54, 0x55, 0x53, 0x5f, 0x41, 0x4c, 0x4c, 0x4f, 0x43, 0x41, 0x54, 0x45, 0x44, 0x10, 0x02,
	0x12, 0x24, 0x0a, 0x20, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e, 0x47, 0x5f, 0x53, 0x54, 0x41, 0x54,
	0x55, 0x53, 0x5f, 0x52, 0x45, 0x4c, 0x45, 0x41, 0x53, 0x45, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45,
	0x53, 0x54, 0x45, 0x44, 0x10, 0x03, 0x12, 0x1b, 0x0a, 0x17, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e,
	0x47, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x4c, 0x45, 0x41, 0x53, 0x45,
	0x44, 0x10, 0x04, 0x12, 0x1a, 0x0a, 0x16, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e, 0x47, 0x5f, 0x53,
	0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x05, 0x12,
	0x18, 0x0a, 0x14, 0x53, 0x45, 0x52, 0x56, 0x49, 0x4e, 0x47, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55,
	0x53, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x06, 0x32, 0xb3, 0x02, 0x0a, 0x15, 0x4e, 0x6f,
	0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x43, 0x0a, 0x0c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4e, 0x6f,
	0x64, 0x65, 0x73, 0x12, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x6f, 0x64, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x4e, 0x6f, 0x64, 0x65, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x0c, 0x52, 0x65, 0x6c, 0x65,
	0x61, 0x73, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x1a,
	0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x6c, 0x65,
	0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4e, 0x0a,
	0x11, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x1a, 0x20, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x40, 0x0a,
	0x14, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x65, 0x61, 0x72,
	0x74, 0x62, 0x65, 0x61, 0x74, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x1a, 0x0f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4f, 0x6b, 0x22, 0x00, 0x32,
	0x92, 0x01, 0x0a, 0x17, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x36, 0x0a, 0x10, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4c, 0x69, 0x76, 0x65, 0x4e, 0x6f, 0x64, 0x65,
	0x1a, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4f,
	0x6b, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x0d, 0x4e, 0x6f, 0x64, 0x65, 0x48, 0x65, 0x61, 0x72, 0x74,
	0x62, 0x65, 0x61, 0x74, 0x12, 0x1b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x6f, 0x64,
	0x65, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x4f, 0x6b, 0x22, 0x00, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescOnce sync.Once
	file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescData = file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDesc
)

func file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescGZIP() []byte {
	file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescOnce.Do(func() {
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescData = protoimpl.X.CompressGZIP(file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescData)
	})
	return file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDescData
}

var file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_goTypes = []any{
	(NodeStatus)(0),                   // 0: proto.NodeStatus
	(ServingStatus)(0),                // 1: proto.ServingStatus
	(*RequestHandler)(nil),            // 2: proto.RequestHandler
	(*NodeRequest)(nil),               // 3: proto.NodeRequest
	(*NodeAllocationResponse)(nil),    // 4: proto.NodeAllocationResponse
	(*NodeReleaseRequest)(nil),        // 5: proto.NodeReleaseRequest
	(*NodeReleaseResponse)(nil),       // 6: proto.NodeReleaseResponse
	(*LiveNode)(nil),                  // 7: proto.LiveNode
	(*NodeHeartbeatRequest)(nil),      // 8: proto.NodeHeartbeatRequest
	(*NodeRequestStatusResponse)(nil), // 9: proto.NodeRequestStatusResponse
	(*UpdateOk)(nil),                  // 10: proto.UpdateOk
}
var file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_depIdxs = []int32{
	2,  // 0: proto.NodeAllocationResponse.RequestHandler:type_name -> proto.RequestHandler
	2,  // 1: proto.NodeReleaseRequest.RequestHandler:type_name -> proto.RequestHandler
	0,  // 2: proto.LiveNode.NodeStatus:type_name -> proto.NodeStatus
	1,  // 3: proto.NodeRequestStatusResponse.ServingStatus:type_name -> proto.ServingStatus
	7,  // 4: proto.NodeRequestStatusResponse.Nodes:type_name -> proto.LiveNode
	3,  // 5: proto.NodeRequestingService.RequestNodes:input_type -> proto.NodeRequest
	2,  // 6: proto.NodeRequestingService.ReleaseNodes:input_type -> proto.RequestHandler
	2,  // 7: proto.NodeRequestingService.NodeRequestStatus:input_type -> proto.RequestHandler
	2,  // 8: proto.NodeRequestingService.SendRequestHeartbeat:input_type -> proto.RequestHandler
	7,  // 9: proto.NodeStatusUpdateService.UpdateNodeStatus:input_type -> proto.LiveNode
	8,  // 10: proto.NodeStatusUpdateService.NodeHeartbeat:input_type -> proto.NodeHeartbeatRequest
	4,  // 11: proto.NodeRequestingService.RequestNodes:output_type -> proto.NodeAllocationResponse
	6,  // 12: proto.NodeRequestingService.ReleaseNodes:output_type -> proto.NodeReleaseResponse
	9,  // 13: proto.NodeRequestingService.NodeRequestStatus:output_type -> proto.NodeRequestStatusResponse
	10, // 14: proto.NodeRequestingService.SendRequestHeartbeat:output_type -> proto.UpdateOk
	10, // 15: proto.NodeStatusUpdateService.UpdateNodeStatus:output_type -> proto.UpdateOk
	10, // 16: proto.NodeStatusUpdateService.NodeHeartbeat:output_type -> proto.UpdateOk
	11, // [11:17] is the sub-list for method output_type
	5,  // [5:11] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_init() }
func file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_init() {
	if File_grpc_resource_alloc_procotol_resourceallocprotocol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*RequestHandler); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*NodeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*NodeAllocationResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*NodeReleaseRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*NodeReleaseResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*LiveNode); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*NodeHeartbeatRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*NodeRequestStatusResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*UpdateOk); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_goTypes,
		DependencyIndexes: file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_depIdxs,
		EnumInfos:         file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_enumTypes,
		MessageInfos:      file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_msgTypes,
	}.Build()
	File_grpc_resource_alloc_procotol_resourceallocprotocol_proto = out.File
	file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_rawDesc = nil
	file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_goTypes = nil
	file_grpc_resource_alloc_procotol_resourceallocprotocol_proto_depIdxs = nil
}
