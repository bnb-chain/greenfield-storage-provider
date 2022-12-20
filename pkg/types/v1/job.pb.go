// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: pkg/types/v1/job.proto

package v1

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

type JobType int32

const (
	JobType_JOB_TYPE_UNSPECIFIED              JobType = 0
	JobType_JOB_TYPE_CREATE_OBJECT            JobType = 1
	JobType_JOB_TYPE_ALLOC_SECONDARY          JobType = 2
	JobType_JOB_TYPE_SEAL_OBJECT              JobType = 3
	JobType_JOB_TYPE_UPLOAD_PRIMARY           JobType = 4
	JobType_JOB_TYPE_UPLOAD_SECONDARY_REPLICA JobType = 5
	JobType_JOB_TYPE_UPLOAD_SECONDARY_EC      JobType = 6
	JobType_JOB_TYPE_UPLOAD_SECONDARY_INLINE  JobType = 7
)

// Enum value maps for JobType.
var (
	JobType_name = map[int32]string{
		0: "JOB_TYPE_UNSPECIFIED",
		1: "JOB_TYPE_CREATE_OBJECT",
		2: "JOB_TYPE_ALLOC_SECONDARY",
		3: "JOB_TYPE_SEAL_OBJECT",
		4: "JOB_TYPE_UPLOAD_PRIMARY",
		5: "JOB_TYPE_UPLOAD_SECONDARY_REPLICA",
		6: "JOB_TYPE_UPLOAD_SECONDARY_EC",
		7: "JOB_TYPE_UPLOAD_SECONDARY_INLINE",
	}
	JobType_value = map[string]int32{
		"JOB_TYPE_UNSPECIFIED":              0,
		"JOB_TYPE_CREATE_OBJECT":            1,
		"JOB_TYPE_ALLOC_SECONDARY":          2,
		"JOB_TYPE_SEAL_OBJECT":              3,
		"JOB_TYPE_UPLOAD_PRIMARY":           4,
		"JOB_TYPE_UPLOAD_SECONDARY_REPLICA": 5,
		"JOB_TYPE_UPLOAD_SECONDARY_EC":      6,
		"JOB_TYPE_UPLOAD_SECONDARY_INLINE":  7,
	}
)

func (x JobType) Enum() *JobType {
	p := new(JobType)
	*p = x
	return p
}

func (x JobType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (JobType) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_types_v1_job_proto_enumTypes[0].Descriptor()
}

func (JobType) Type() protoreflect.EnumType {
	return &file_pkg_types_v1_job_proto_enumTypes[0]
}

func (x JobType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use JobType.Descriptor instead.
func (JobType) EnumDescriptor() ([]byte, []int) {
	return file_pkg_types_v1_job_proto_rawDescGZIP(), []int{0}
}

type JobState int32

const (
	JobState_JOB_STATE_INIT_UNSPECIFIED           JobState = 0
	JobState_JOB_STATE_DONE                       JobState = 1
	JobState_JOB_STATE_ERROR                      JobState = 2
	JobState_JOB_STATE_CREATE_OBJECT_INIT         JobState = 3
	JobState_JOB_STATE_CREATE_OBJECT_TX_DOING     JobState = 4
	JobState_JOB_STATE_CREATE_OBJECT_DONE         JobState = 5
	JobState_JOB_STATE_UPLOAD_PAYLOAD_INIT        JobState = 6
	JobState_JOB_STATE_UPLOAD_PAYLOAD_DOING       JobState = 7
	JobState_JOB_STATE_UPLOAD_PAYLOAD_DONE        JobState = 8
	JobState_JOB_STATE_ALLOC_SECONDARY_INIT       JobState = 9
	JobState_JOB_STATE_ALLOC_SECONDARY_DOING      JobState = 10
	JobState_JOB_STATE_ALLOC_SECONDARY_DONE       JobState = 11
	JobState_JOB_STATE_SEAL_OBJECT_INIT           JobState = 12
	JobState_JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE JobState = 13
	JobState_JOB_STATE_SEAL_OBJECT_TX_DOING       JobState = 14
	JobState_JOB_STATE_SEAL_OBJECT_DONE           JobState = 15
)

// Enum value maps for JobState.
var (
	JobState_name = map[int32]string{
		0:  "JOB_STATE_INIT_UNSPECIFIED",
		1:  "JOB_STATE_DONE",
		2:  "JOB_STATE_ERROR",
		3:  "JOB_STATE_CREATE_OBJECT_INIT",
		4:  "JOB_STATE_CREATE_OBJECT_TX_DOING",
		5:  "JOB_STATE_CREATE_OBJECT_DONE",
		6:  "JOB_STATE_UPLOAD_PAYLOAD_INIT",
		7:  "JOB_STATE_UPLOAD_PAYLOAD_DOING",
		8:  "JOB_STATE_UPLOAD_PAYLOAD_DONE",
		9:  "JOB_STATE_ALLOC_SECONDARY_INIT",
		10: "JOB_STATE_ALLOC_SECONDARY_DOING",
		11: "JOB_STATE_ALLOC_SECONDARY_DONE",
		12: "JOB_STATE_SEAL_OBJECT_INIT",
		13: "JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE",
		14: "JOB_STATE_SEAL_OBJECT_TX_DOING",
		15: "JOB_STATE_SEAL_OBJECT_DONE",
	}
	JobState_value = map[string]int32{
		"JOB_STATE_INIT_UNSPECIFIED":           0,
		"JOB_STATE_DONE":                       1,
		"JOB_STATE_ERROR":                      2,
		"JOB_STATE_CREATE_OBJECT_INIT":         3,
		"JOB_STATE_CREATE_OBJECT_TX_DOING":     4,
		"JOB_STATE_CREATE_OBJECT_DONE":         5,
		"JOB_STATE_UPLOAD_PAYLOAD_INIT":        6,
		"JOB_STATE_UPLOAD_PAYLOAD_DOING":       7,
		"JOB_STATE_UPLOAD_PAYLOAD_DONE":        8,
		"JOB_STATE_ALLOC_SECONDARY_INIT":       9,
		"JOB_STATE_ALLOC_SECONDARY_DOING":      10,
		"JOB_STATE_ALLOC_SECONDARY_DONE":       11,
		"JOB_STATE_SEAL_OBJECT_INIT":           12,
		"JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE": 13,
		"JOB_STATE_SEAL_OBJECT_TX_DOING":       14,
		"JOB_STATE_SEAL_OBJECT_DONE":           15,
	}
)

func (x JobState) Enum() *JobState {
	p := new(JobState)
	*p = x
	return p
}

func (x JobState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (JobState) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_types_v1_job_proto_enumTypes[1].Descriptor()
}

func (JobState) Type() protoreflect.EnumType {
	return &file_pkg_types_v1_job_proto_enumTypes[1]
}

func (x JobState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use JobState.Descriptor instead.
func (JobState) EnumDescriptor() ([]byte, []int) {
	return file_pkg_types_v1_job_proto_rawDescGZIP(), []int{1}
}

type JobContext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JobId      uint64      `protobuf:"varint,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"` // unique, identifies one job
	JobType    JobType     `protobuf:"varint,2,opt,name=job_type,json=jobType,proto3,enum=pkg.types.v1.JobType" json:"job_type,omitempty"`
	JobState   JobState    `protobuf:"varint,3,opt,name=job_state,json=jobState,proto3,enum=pkg.types.v1.JobState" json:"job_state,omitempty"`
	JobErr     string      `protobuf:"bytes,4,opt,name=job_err,json=jobErr,proto3" json:"job_err,omitempty"`              // default empty, if the job is interrupted, will log the error message
	CreateTime uint64      `protobuf:"varint,5,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"` // the job create time, used to jobs garbage collection
	ModifyTime uint64      `protobuf:"varint,6,opt,name=modify_time,json=modifyTime,proto3" json:"modify_time,omitempty"` // the job last modified time, used to judge timeout
	ObjectInfo *ObjectInfo `protobuf:"bytes,7,opt,name=object_info,json=objectInfo,proto3" json:"object_info,omitempty"`
}

func (x *JobContext) Reset() {
	*x = JobContext{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_types_v1_job_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JobContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JobContext) ProtoMessage() {}

func (x *JobContext) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_types_v1_job_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JobContext.ProtoReflect.Descriptor instead.
func (*JobContext) Descriptor() ([]byte, []int) {
	return file_pkg_types_v1_job_proto_rawDescGZIP(), []int{0}
}

func (x *JobContext) GetJobId() uint64 {
	if x != nil {
		return x.JobId
	}
	return 0
}

func (x *JobContext) GetJobType() JobType {
	if x != nil {
		return x.JobType
	}
	return JobType_JOB_TYPE_UNSPECIFIED
}

func (x *JobContext) GetJobState() JobState {
	if x != nil {
		return x.JobState
	}
	return JobState_JOB_STATE_INIT_UNSPECIFIED
}

func (x *JobContext) GetJobErr() string {
	if x != nil {
		return x.JobErr
	}
	return ""
}

func (x *JobContext) GetCreateTime() uint64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

func (x *JobContext) GetModifyTime() uint64 {
	if x != nil {
		return x.ModifyTime
	}
	return 0
}

func (x *JobContext) GetObjectInfo() *ObjectInfo {
	if x != nil {
		return x.ObjectInfo
	}
	return nil
}

type SegmentPieceJob struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BucketName  string `protobuf:"bytes,1,opt,name=bucket_name,json=bucketName,proto3" json:"bucket_name,omitempty"`
	ObjectName  string `protobuf:"bytes,2,opt,name=object_name,json=objectName,proto3" json:"object_name,omitempty"`
	Done        bool   `protobuf:"varint,3,opt,name=done,proto3" json:"done,omitempty"`
	SegmentIdx  uint32 `protobuf:"varint,4,opt,name=segment_idx,json=segmentIdx,proto3" json:"segment_idx,omitempty"` // the index of segment in payload data
	SecondarySp string `protobuf:"bytes,5,opt,name=secondary_sp,json=secondarySp,proto3" json:"secondary_sp,omitempty"`
	CheckSum    []byte `protobuf:"bytes,6,opt,name=check_sum,json=checkSum,proto3" json:"check_sum,omitempty"`
}

func (x *SegmentPieceJob) Reset() {
	*x = SegmentPieceJob{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_types_v1_job_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SegmentPieceJob) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SegmentPieceJob) ProtoMessage() {}

func (x *SegmentPieceJob) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_types_v1_job_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SegmentPieceJob.ProtoReflect.Descriptor instead.
func (*SegmentPieceJob) Descriptor() ([]byte, []int) {
	return file_pkg_types_v1_job_proto_rawDescGZIP(), []int{1}
}

func (x *SegmentPieceJob) GetBucketName() string {
	if x != nil {
		return x.BucketName
	}
	return ""
}

func (x *SegmentPieceJob) GetObjectName() string {
	if x != nil {
		return x.ObjectName
	}
	return ""
}

func (x *SegmentPieceJob) GetDone() bool {
	if x != nil {
		return x.Done
	}
	return false
}

func (x *SegmentPieceJob) GetSegmentIdx() uint32 {
	if x != nil {
		return x.SegmentIdx
	}
	return 0
}

func (x *SegmentPieceJob) GetSecondarySp() string {
	if x != nil {
		return x.SecondarySp
	}
	return ""
}

func (x *SegmentPieceJob) GetCheckSum() []byte {
	if x != nil {
		return x.CheckSum
	}
	return nil
}

type InlinePieceJob struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BucketName  string `protobuf:"bytes,1,opt,name=bucket_name,json=bucketName,proto3" json:"bucket_name,omitempty"`
	ObjectName  string `protobuf:"bytes,2,opt,name=object_name,json=objectName,proto3" json:"object_name,omitempty"`
	Done        bool   `protobuf:"varint,3,opt,name=done,proto3" json:"done,omitempty"`
	SecondarySp string `protobuf:"bytes,4,opt,name=secondary_sp,json=secondarySp,proto3" json:"secondary_sp,omitempty"`
	CheckSum    []byte `protobuf:"bytes,5,opt,name=check_sum,json=checkSum,proto3" json:"check_sum,omitempty"`
}

func (x *InlinePieceJob) Reset() {
	*x = InlinePieceJob{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_types_v1_job_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InlinePieceJob) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InlinePieceJob) ProtoMessage() {}

func (x *InlinePieceJob) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_types_v1_job_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InlinePieceJob.ProtoReflect.Descriptor instead.
func (*InlinePieceJob) Descriptor() ([]byte, []int) {
	return file_pkg_types_v1_job_proto_rawDescGZIP(), []int{2}
}

func (x *InlinePieceJob) GetBucketName() string {
	if x != nil {
		return x.BucketName
	}
	return ""
}

func (x *InlinePieceJob) GetObjectName() string {
	if x != nil {
		return x.ObjectName
	}
	return ""
}

func (x *InlinePieceJob) GetDone() bool {
	if x != nil {
		return x.Done
	}
	return false
}

func (x *InlinePieceJob) GetSecondarySp() string {
	if x != nil {
		return x.SecondarySp
	}
	return ""
}

func (x *InlinePieceJob) GetCheckSum() []byte {
	if x != nil {
		return x.CheckSum
	}
	return nil
}

type BatchECPieceJob struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BucketName string `protobuf:"bytes,1,opt,name=bucket_name,json=bucketName,proto3" json:"bucket_name,omitempty"`
	ObjectName string `protobuf:"bytes,2,opt,name=object_name,json=objectName,proto3" json:"object_name,omitempty"`
	Done       bool   `protobuf:"varint,3,opt,name=done,proto3" json:"done,omitempty"`
	SegmentIdx uint32 `protobuf:"varint,4,opt,name=segment_idx,json=segmentIdx,proto3" json:"segment_idx,omitempty"`
	EcM        uint32 `protobuf:"varint,5,opt,name=ec_m,json=ecM,proto3" json:"ec_m,omitempty"`
	EcK        uint32 `protobuf:"varint,6,opt,name=ec_k,json=ecK,proto3" json:"ec_k,omitempty"`
	CheckSum   []byte `protobuf:"bytes,7,opt,name=check_sum,json=checkSum,proto3" json:"check_sum,omitempty"`
}

func (x *BatchECPieceJob) Reset() {
	*x = BatchECPieceJob{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_types_v1_job_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchECPieceJob) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchECPieceJob) ProtoMessage() {}

func (x *BatchECPieceJob) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_types_v1_job_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchECPieceJob.ProtoReflect.Descriptor instead.
func (*BatchECPieceJob) Descriptor() ([]byte, []int) {
	return file_pkg_types_v1_job_proto_rawDescGZIP(), []int{3}
}

func (x *BatchECPieceJob) GetBucketName() string {
	if x != nil {
		return x.BucketName
	}
	return ""
}

func (x *BatchECPieceJob) GetObjectName() string {
	if x != nil {
		return x.ObjectName
	}
	return ""
}

func (x *BatchECPieceJob) GetDone() bool {
	if x != nil {
		return x.Done
	}
	return false
}

func (x *BatchECPieceJob) GetSegmentIdx() uint32 {
	if x != nil {
		return x.SegmentIdx
	}
	return 0
}

func (x *BatchECPieceJob) GetEcM() uint32 {
	if x != nil {
		return x.EcM
	}
	return 0
}

func (x *BatchECPieceJob) GetEcK() uint32 {
	if x != nil {
		return x.EcK
	}
	return 0
}

func (x *BatchECPieceJob) GetCheckSum() []byte {
	if x != nil {
		return x.CheckSum
	}
	return nil
}

type ECPieceJob struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BucketName   string `protobuf:"bytes,1,opt,name=bucket_name,json=bucketName,proto3" json:"bucket_name,omitempty"`
	ObjectName   string `protobuf:"bytes,2,opt,name=object_name,json=objectName,proto3" json:"object_name,omitempty"`
	Done         bool   `protobuf:"varint,3,opt,name=done,proto3" json:"done,omitempty"`
	SegmentIdx   uint32 `protobuf:"varint,4,opt,name=segment_idx,json=segmentIdx,proto3" json:"segment_idx,omitempty"`         // the index of segment in payload data
	SegmentEcIdx uint32 `protobuf:"varint,5,opt,name=segment_ec_idx,json=segmentEcIdx,proto3" json:"segment_ec_idx,omitempty"` // the index of ec-piece in segment data
	SecondarySp  string `protobuf:"bytes,6,opt,name=secondary_sp,json=secondarySp,proto3" json:"secondary_sp,omitempty"`
	EcM          uint32 `protobuf:"varint,7,opt,name=ec_m,json=ecM,proto3" json:"ec_m,omitempty"`
	EcK          uint32 `protobuf:"varint,8,opt,name=ec_k,json=ecK,proto3" json:"ec_k,omitempty"`
	CheckSum     []byte `protobuf:"bytes,9,opt,name=check_sum,json=checkSum,proto3" json:"check_sum,omitempty"`
}

func (x *ECPieceJob) Reset() {
	*x = ECPieceJob{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_types_v1_job_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ECPieceJob) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ECPieceJob) ProtoMessage() {}

func (x *ECPieceJob) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_types_v1_job_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ECPieceJob.ProtoReflect.Descriptor instead.
func (*ECPieceJob) Descriptor() ([]byte, []int) {
	return file_pkg_types_v1_job_proto_rawDescGZIP(), []int{4}
}

func (x *ECPieceJob) GetBucketName() string {
	if x != nil {
		return x.BucketName
	}
	return ""
}

func (x *ECPieceJob) GetObjectName() string {
	if x != nil {
		return x.ObjectName
	}
	return ""
}

func (x *ECPieceJob) GetDone() bool {
	if x != nil {
		return x.Done
	}
	return false
}

func (x *ECPieceJob) GetSegmentIdx() uint32 {
	if x != nil {
		return x.SegmentIdx
	}
	return 0
}

func (x *ECPieceJob) GetSegmentEcIdx() uint32 {
	if x != nil {
		return x.SegmentEcIdx
	}
	return 0
}

func (x *ECPieceJob) GetSecondarySp() string {
	if x != nil {
		return x.SecondarySp
	}
	return ""
}

func (x *ECPieceJob) GetEcM() uint32 {
	if x != nil {
		return x.EcM
	}
	return 0
}

func (x *ECPieceJob) GetEcK() uint32 {
	if x != nil {
		return x.EcK
	}
	return 0
}

func (x *ECPieceJob) GetCheckSum() []byte {
	if x != nil {
		return x.CheckSum
	}
	return nil
}

var File_pkg_types_v1_job_proto protoreflect.FileDescriptor

var file_pkg_types_v1_job_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x6a,
	0x6f, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x70, 0x6b, 0x67, 0x2e, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x19, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2f, 0x76, 0x31, 0x2f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xa0, 0x02, 0x0a, 0x0a, 0x4a, 0x6f, 0x62, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74,
	0x12, 0x15, 0x0a, 0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49, 0x64, 0x12, 0x30, 0x0a, 0x08, 0x6a, 0x6f, 0x62, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x70, 0x6b, 0x67, 0x2e,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4a, 0x6f, 0x62, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x07, 0x6a, 0x6f, 0x62, 0x54, 0x79, 0x70, 0x65, 0x12, 0x33, 0x0a, 0x09, 0x6a, 0x6f, 0x62,
	0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x16, 0x2e, 0x70,
	0x6b, 0x67, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4a, 0x6f, 0x62, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x52, 0x08, 0x6a, 0x6f, 0x62, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x17,
	0x0a, 0x07, 0x6a, 0x6f, 0x62, 0x5f, 0x65, 0x72, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x6a, 0x6f, 0x62, 0x45, 0x72, 0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x69,
	0x66, 0x79, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x6d,
	0x6f, 0x64, 0x69, 0x66, 0x79, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x39, 0x0a, 0x0b, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18,
	0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0a, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x49, 0x6e, 0x66, 0x6f, 0x22, 0xc8, 0x01, 0x0a, 0x0f, 0x53, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74,
	0x50, 0x69, 0x65, 0x63, 0x65, 0x4a, 0x6f, 0x62, 0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x75, 0x63, 0x6b,
	0x65, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x62,
	0x75, 0x63, 0x6b, 0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x6f,
	0x6e, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x12, 0x1f,
	0x0a, 0x0b, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x78, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x0a, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x78, 0x12,
	0x21, 0x0a, 0x0c, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x5f, 0x73, 0x70, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x61, 0x72, 0x79,
	0x53, 0x70, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x73, 0x75, 0x6d, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x53, 0x75, 0x6d, 0x22,
	0xa6, 0x01, 0x0a, 0x0e, 0x49, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x50, 0x69, 0x65, 0x63, 0x65, 0x4a,
	0x6f, 0x62, 0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x63, 0x6f,
	0x6e, 0x64, 0x61, 0x72, 0x79, 0x5f, 0x73, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x53, 0x70, 0x12, 0x1b, 0x0a, 0x09, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x5f, 0x73, 0x75, 0x6d, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x53, 0x75, 0x6d, 0x22, 0xcb, 0x01, 0x0a, 0x0f, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x45, 0x43, 0x50, 0x69, 0x65, 0x63, 0x65, 0x4a, 0x6f, 0x62, 0x12, 0x1f, 0x0a, 0x0b,
	0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a,
	0x0b, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x64, 0x6f,
	0x6e, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64,
	0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74,
	0x49, 0x64, 0x78, 0x12, 0x11, 0x0a, 0x04, 0x65, 0x63, 0x5f, 0x6d, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x03, 0x65, 0x63, 0x4d, 0x12, 0x11, 0x0a, 0x04, 0x65, 0x63, 0x5f, 0x6b, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x65, 0x63, 0x4b, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x5f, 0x73, 0x75, 0x6d, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x63, 0x68,
	0x65, 0x63, 0x6b, 0x53, 0x75, 0x6d, 0x22, 0x8f, 0x02, 0x0a, 0x0a, 0x45, 0x43, 0x50, 0x69, 0x65,
	0x63, 0x65, 0x4a, 0x6f, 0x62, 0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x62, 0x75, 0x63, 0x6b,
	0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73,
	0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x0a, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x78, 0x12, 0x24, 0x0a, 0x0e,
	0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x65, 0x63, 0x5f, 0x69, 0x64, 0x78, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x0c, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x45, 0x63, 0x49,
	0x64, 0x78, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x5f,
	0x73, 0x70, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64,
	0x61, 0x72, 0x79, 0x53, 0x70, 0x12, 0x11, 0x0a, 0x04, 0x65, 0x63, 0x5f, 0x6d, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x03, 0x65, 0x63, 0x4d, 0x12, 0x11, 0x0a, 0x04, 0x65, 0x63, 0x5f, 0x6b,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x65, 0x63, 0x4b, 0x12, 0x1b, 0x0a, 0x09, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x5f, 0x73, 0x75, 0x6d, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x53, 0x75, 0x6d, 0x2a, 0x83, 0x02, 0x0a, 0x07, 0x4a, 0x6f, 0x62,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x14, 0x4a, 0x4f, 0x42, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1a,
	0x0a, 0x16, 0x4a, 0x4f, 0x42, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54,
	0x45, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x10, 0x01, 0x12, 0x1c, 0x0a, 0x18, 0x4a, 0x4f,
	0x42, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x41, 0x4c, 0x4c, 0x4f, 0x43, 0x5f, 0x53, 0x45, 0x43,
	0x4f, 0x4e, 0x44, 0x41, 0x52, 0x59, 0x10, 0x02, 0x12, 0x18, 0x0a, 0x14, 0x4a, 0x4f, 0x42, 0x5f,
	0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x45, 0x41, 0x4c, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54,
	0x10, 0x03, 0x12, 0x1b, 0x0a, 0x17, 0x4a, 0x4f, 0x42, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55,
	0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x50, 0x52, 0x49, 0x4d, 0x41, 0x52, 0x59, 0x10, 0x04, 0x12,
	0x25, 0x0a, 0x21, 0x4a, 0x4f, 0x42, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x50, 0x4c, 0x4f,
	0x41, 0x44, 0x5f, 0x53, 0x45, 0x43, 0x4f, 0x4e, 0x44, 0x41, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x50,
	0x4c, 0x49, 0x43, 0x41, 0x10, 0x05, 0x12, 0x20, 0x0a, 0x1c, 0x4a, 0x4f, 0x42, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x53, 0x45, 0x43, 0x4f, 0x4e, 0x44,
	0x41, 0x52, 0x59, 0x5f, 0x45, 0x43, 0x10, 0x06, 0x12, 0x24, 0x0a, 0x20, 0x4a, 0x4f, 0x42, 0x5f,
	0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x53, 0x45, 0x43, 0x4f,
	0x4e, 0x44, 0x41, 0x52, 0x59, 0x5f, 0x49, 0x4e, 0x4c, 0x49, 0x4e, 0x45, 0x10, 0x07, 0x2a, 0xa2,
	0x04, 0x0a, 0x08, 0x4a, 0x6f, 0x62, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1e, 0x0a, 0x1a, 0x4a,
	0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x49, 0x4e, 0x49, 0x54, 0x5f, 0x55, 0x4e,
	0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x12, 0x0a, 0x0e, 0x4a,
	0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x01, 0x12,
	0x13, 0x0a, 0x0f, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x45, 0x52, 0x52,
	0x4f, 0x52, 0x10, 0x02, 0x12, 0x20, 0x0a, 0x1c, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54,
	0x45, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f,
	0x49, 0x4e, 0x49, 0x54, 0x10, 0x03, 0x12, 0x24, 0x0a, 0x20, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54,
	0x41, 0x54, 0x45, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43,
	0x54, 0x5f, 0x54, 0x58, 0x5f, 0x44, 0x4f, 0x49, 0x4e, 0x47, 0x10, 0x04, 0x12, 0x20, 0x0a, 0x1c,
	0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45,
	0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x05, 0x12, 0x21,
	0x0a, 0x1d, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x55, 0x50, 0x4c, 0x4f,
	0x41, 0x44, 0x5f, 0x50, 0x41, 0x59, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x49, 0x4e, 0x49, 0x54, 0x10,
	0x06, 0x12, 0x22, 0x0a, 0x1e, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x55,
	0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x50, 0x41, 0x59, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x44, 0x4f,
	0x49, 0x4e, 0x47, 0x10, 0x07, 0x12, 0x21, 0x0a, 0x1d, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41,
	0x54, 0x45, 0x5f, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x50, 0x41, 0x59, 0x4c, 0x4f, 0x41,
	0x44, 0x5f, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x08, 0x12, 0x22, 0x0a, 0x1e, 0x4a, 0x4f, 0x42, 0x5f,
	0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x41, 0x4c, 0x4c, 0x4f, 0x43, 0x5f, 0x53, 0x45, 0x43, 0x4f,
	0x4e, 0x44, 0x41, 0x52, 0x59, 0x5f, 0x49, 0x4e, 0x49, 0x54, 0x10, 0x09, 0x12, 0x23, 0x0a, 0x1f,
	0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x41, 0x4c, 0x4c, 0x4f, 0x43, 0x5f,
	0x53, 0x45, 0x43, 0x4f, 0x4e, 0x44, 0x41, 0x52, 0x59, 0x5f, 0x44, 0x4f, 0x49, 0x4e, 0x47, 0x10,
	0x0a, 0x12, 0x22, 0x0a, 0x1e, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x41,
	0x4c, 0x4c, 0x4f, 0x43, 0x5f, 0x53, 0x45, 0x43, 0x4f, 0x4e, 0x44, 0x41, 0x52, 0x59, 0x5f, 0x44,
	0x4f, 0x4e, 0x45, 0x10, 0x0b, 0x12, 0x1e, 0x0a, 0x1a, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41,
	0x54, 0x45, 0x5f, 0x53, 0x45, 0x41, 0x4c, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x49,
	0x4e, 0x49, 0x54, 0x10, 0x0c, 0x12, 0x28, 0x0a, 0x24, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41,
	0x54, 0x45, 0x5f, 0x53, 0x45, 0x41, 0x4c, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x53,
	0x49, 0x47, 0x4e, 0x41, 0x54, 0x55, 0x52, 0x45, 0x5f, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x0d, 0x12,
	0x22, 0x0a, 0x1e, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x53, 0x45, 0x41,
	0x4c, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x54, 0x58, 0x5f, 0x44, 0x4f, 0x49, 0x4e,
	0x47, 0x10, 0x0e, 0x12, 0x1e, 0x0a, 0x1a, 0x4a, 0x4f, 0x42, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45,
	0x5f, 0x53, 0x45, 0x41, 0x4c, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x44, 0x4f, 0x4e,
	0x45, 0x10, 0x0f, 0x42, 0x40, 0x5a, 0x3e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x62, 0x6e, 0x62, 0x2d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2f, 0x69, 0x6e, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2d, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2d,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_types_v1_job_proto_rawDescOnce sync.Once
	file_pkg_types_v1_job_proto_rawDescData = file_pkg_types_v1_job_proto_rawDesc
)

func file_pkg_types_v1_job_proto_rawDescGZIP() []byte {
	file_pkg_types_v1_job_proto_rawDescOnce.Do(func() {
		file_pkg_types_v1_job_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_types_v1_job_proto_rawDescData)
	})
	return file_pkg_types_v1_job_proto_rawDescData
}

var file_pkg_types_v1_job_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_pkg_types_v1_job_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pkg_types_v1_job_proto_goTypes = []interface{}{
	(JobType)(0),            // 0: pkg.types.v1.JobType
	(JobState)(0),           // 1: pkg.types.v1.JobState
	(*JobContext)(nil),      // 2: pkg.types.v1.JobContext
	(*SegmentPieceJob)(nil), // 3: pkg.types.v1.SegmentPieceJob
	(*InlinePieceJob)(nil),  // 4: pkg.types.v1.InlinePieceJob
	(*BatchECPieceJob)(nil), // 5: pkg.types.v1.BatchECPieceJob
	(*ECPieceJob)(nil),      // 6: pkg.types.v1.ECPieceJob
	(*ObjectInfo)(nil),      // 7: pkg.types.v1.ObjectInfo
}
var file_pkg_types_v1_job_proto_depIdxs = []int32{
	0, // 0: pkg.types.v1.JobContext.job_type:type_name -> pkg.types.v1.JobType
	1, // 1: pkg.types.v1.JobContext.job_state:type_name -> pkg.types.v1.JobState
	7, // 2: pkg.types.v1.JobContext.object_info:type_name -> pkg.types.v1.ObjectInfo
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_pkg_types_v1_job_proto_init() }
func file_pkg_types_v1_job_proto_init() {
	if File_pkg_types_v1_job_proto != nil {
		return
	}
	file_pkg_types_v1_object_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_pkg_types_v1_job_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JobContext); i {
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
		file_pkg_types_v1_job_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SegmentPieceJob); i {
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
		file_pkg_types_v1_job_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InlinePieceJob); i {
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
		file_pkg_types_v1_job_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchECPieceJob); i {
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
		file_pkg_types_v1_job_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ECPieceJob); i {
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
			RawDescriptor: file_pkg_types_v1_job_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_types_v1_job_proto_goTypes,
		DependencyIndexes: file_pkg_types_v1_job_proto_depIdxs,
		EnumInfos:         file_pkg_types_v1_job_proto_enumTypes,
		MessageInfos:      file_pkg_types_v1_job_proto_msgTypes,
	}.Build()
	File_pkg_types_v1_job_proto = out.File
	file_pkg_types_v1_job_proto_rawDesc = nil
	file_pkg_types_v1_job_proto_goTypes = nil
	file_pkg_types_v1_job_proto_depIdxs = nil
}
