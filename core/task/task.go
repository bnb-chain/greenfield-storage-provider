package task

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// Task is the interface to the smallest unit of SP background service interaction.
//
// Task Type:
//
//	There are two main types of task, ObjectTask and GCTask. The ObjectTask
//	associated with an object, and records the information of different stages
//	of the object, includes UploadObjectTask, ReplicatePieceTask, ReceivePieceTask,
//	SealObjectTask, DownloadObjectTask. The GCTask is the interface to record
//	the information of garbage collection, includes GCObjectTask, GCZombiePieceTask,
//	GCStoreTask.
//
//	The UploadObjectTask is the interface to record the information for uploading
//	object payload data to the primary SP.
//
//	The ReplicatePieceTask is the interface to record the information for
//	replicating	pieces of object payload data to secondary SPs.
//
//	The ReceivePieceTask is the interface to record the information for receiving
//	pieces of object payload data from primary SP, it exists only in secondary SP.
//
//	The SealObjectTask is the interface to  record the information for sealing
//	object to the Greenfield chain.
//
//	The DownloadObjectTask is the interface to record the information for
//	downloading pieces of object payload data.
//
//	The GCObjectTask is the interface to record the information for collecting
//	the piece store space by deleting object payload data that the object has
//	been deleted 	on the Greenfield chain.
//
//	The GCZombiePieceTask is the interface to record the information for collecting
//	the piece store space by deleting zombie pieces data that dues to any exception,
//	the piece data meta is not on chain but the pieces has been store in piece store.
//
//	The GCStoreTask is the interface to record the information for collecting the SP
//	meta store space by deleting the expired data.
//
// Task Priority:
//
//	Each type of task has a priority, the range of priority is [0, 255], the higher
//	the priority, the higher the urgency to be executed, the higher the priority will
//	be allocated resources to execute
//
// Task Priority Level:
//
//	Task priority is divided into three levels, TLowPriorityLevel, TMediumPriorityLevel,
//	THighPriorityLevel. The TLowPriorityLevel default priority range is [0, 85), The
//	TMediumPriorityLevel default priority range is [85, 170), The THighPriorityLevel
//	default priority range is [170, 256). When applying for task execution resources
//	from ResourceManager, the application is based on the task priority level.
//	Example:
//		the resource limit configuration of task execution node :
//			[TasksHighPriority: 30, TasksMediumPriority: 20, TasksLowPriority: 2]
//		the executor of the task can run 30 high level tasks at the same time that the
//			task priority >= 170
//	 	the executor of the task can run 20 medium level tasks at the same time that the
//			task priority between [85, 170)
//		the executor of the task can run 2 medium level tasks at the same time that the
//			task priority < 85
type Task interface {
	// Key returns the uniquely identify of the task. For the ObjectTask the object id
	// can use as the uniquely identify. For the GCTask maybe a task needs to operate
	// multiple objects, the prefix may be needed to distinguish.
	Key() TKey
	// Type returns the type of the task. A task has a unique type, such as
	// UploadObjectTask, ReplicatePieceTask etc. has the only one TType definition.
	Type() TType
	// GetAddress returns the task runner address. there is only one runner at the same
	// time, which will assist in quickly locating the running node of the task.
	GetAddress() string
	SetAddress(string)
	// GetCreateTime returns the creation time of the task. The creation time used to
	// judge task execution time.
	GetCreateTime() int64
	SetCreateTime(int64)
	// GetUpdateTime returns the last updated time of the task. The updated time used
	// to determine whether the task is expired with the timeout.
	GetUpdateTime() int64
	// SetUpdateTime sets last updated time of the task. Any changes in task
	// information requires to set the update time.
	SetUpdateTime(int64)
	// GetTimeout returns the timeout of the task, the timeout is a duration, if
	// update time adds timeout lesser now stands the task is expired.
	GetTimeout() int64
	// SetTimeout sets timeout duration of the task.
	SetTimeout(int64)
	// Expired returns an indicator whether timeout, if update time adds timeout
	// lesser now returns true, otherwise returns false.
	Expired() bool
	ExceedTimeout() bool
	// GetPriority returns the priority of the task. Each type of task has a fixed
	// priority. The higher the priority, the higher the urgency of the task, and
	// it will be executed first.
	GetPriority() TPriority
	// SetPriority sets the priority of the task. In most cases, the priority of the
	// task does not need to be set, because the priority of the task corresponds to
	// the task type one by one. Once the task type is determined, the priority is
	// determined. But some scenarios need to dynamically adjust the priority of the
	// task type, then this interface is needed.
	SetPriority(TPriority)
	SetRetry(int)
	// IncRetry increases the retry counter of the task. Each task has the retry
	// limits, if retry counter exceed the retry limits, the task should be canceled.
	IncRetry()
	// ExceedRetry returns an indicator whether retry max times. If retry counter is
	// greater than the retry limits returns true, otherwise returns false.
	ExceedRetry() bool
	// GetRetry returns the retry counter of the task.
	GetRetry() int64
	// GetMaxRetry returns the limit of retry counter. Each type of task has a fixed
	// retry limit.
	GetMaxRetry() int64
	// SetMaxRetry sets the limit of retry counter.
	SetMaxRetry(int64)
	// EstimateLimit returns estimated resource will be consumed. It is used for
	// application resources to the rcmgr and decide whether it can be executed
	// immediately.
	EstimateLimit() rcmgr.Limit
	// Error returns the task error. if the task is normal, returns nil.
	Error() error
	// SetError sets the error to task. Any errors that occur during task execution
	// will be logged through the SetError method.
	SetError(error)
}

type ApprovalTask interface {
	Task
	GetExpiredHeight() uint64
	SetExpiredHeight(uint64)
}

type ApprovalCreateBucketTask interface {
	ApprovalTask
	InitApprovalCreateBucketTask(*storagetypes.MsgCreateBucket, TPriority)
	GetCreateBucketInfo() *storagetypes.MsgCreateBucket
	SetCreateBucketInfo(*storagetypes.MsgCreateBucket)
}

type ApprovalCreateObjectTask interface {
	ApprovalTask
	InitApprovalCreateObjectTask(*storagetypes.MsgCreateObject, TPriority)
	GetCreateObjectInfo() *storagetypes.MsgCreateObject
	SetCreateObjectInfo(*storagetypes.MsgCreateObject)
}

type ApprovalReplicatePieceTask interface {
	ObjectTask
	ApprovalTask
	InitApprovalReplicatePieceTask(object *storagetypes.ObjectInfo, params *storagetypes.Params, priority TPriority, askOpAddress string)
	GetAskSpOperatorAddress() string
	SetAskSpOperatorAddress(string)
	GetAskSignature() []byte
	SetAskSignature([]byte)
	GetApprovedSpOperatorAddress() string
	SetApprovedSpOperatorAddress(string)
	GetApprovedSignature() []byte
	SetApprovedSignature([]byte)
	GetApprovedSpEndpoint() string
	SetApprovedSpEndpoint(string)
	GetApprovedSpApprovalAddress() string
	SetApprovedSpApprovalAddress(string)
	GetSignBytes() []byte
}

// The ObjectTask associated with an object, and records the information of different
// stages of the object.
type ObjectTask interface {
	Task
	// GetObjectInfo returns the associated object.
	GetObjectInfo() *storagetypes.ObjectInfo
	SetObjectInfo(*storagetypes.ObjectInfo)
	GetStorageParams() *storagetypes.Params
	SetStorageParams(*storagetypes.Params)
}

// The UploadObjectTask is the interface to record the information for uploading object
// payload data to the primary SP.
type UploadObjectTask interface {
	ObjectTask
	InitUploadObjectTask(*storagetypes.ObjectInfo, *storagetypes.Params)
}

// The ReplicatePieceTask is the interface to record the information for replicating
// pieces of object payload data to secondary SPs.
type ReplicatePieceTask interface {
	ObjectTask
	InitReplicatePieceTask(object *storagetypes.ObjectInfo,
		params *storagetypes.Params,
		priority TPriority,
		timeout int64,
		retry int64)
	GetSealed() bool
	SetSealed(bool)
	GetSecondarySignature() [][]byte
	SetSecondarySignature([][]byte)
}

// The ReceivePieceTask is the interface to record the information for receiving pieces
// of object payload data from primary SP, it exists only in secondary SP.
type ReceivePieceTask interface {
	ObjectTask
	InitReceivePieceTask(object *storagetypes.ObjectInfo,
		params *storagetypes.Params,
		priority TPriority,
		replicateIdx uint32,
		pieceIdx int32,
		pieceSize int64)
	// GetReplicateIdx returns the replicate index. The replicate index identifies the
	// serial number of the secondary SP for object piece copy.
	GetReplicateIdx() uint32
	// SetReplicateIdx sets the replicate index.
	SetReplicateIdx(uint32)
	// GetPieceIdx returns the piece index. The piece index identifies the serial number
	// of segment of object payload data for object piece copy.
	GetPieceIdx() int32
	// SetPieceIdx sets the piece index.
	SetPieceIdx(int32)
	GetPieceSize() int64
	SetPieceSize(int64)
	SetPieceChecksum([]byte)
	GetPieceChecksum() []byte
	GetSignature() []byte
	SetSignature([]byte)
	GetSealed() bool
	SetSealed(bool)
	GetSignBytes() []byte
}

// The SealObjectTask is the interface to  record the information for sealing object to
// the Greenfield chain.
type SealObjectTask interface {
	ObjectTask
	InitSealObjectTask(object *storagetypes.ObjectInfo,
		params *storagetypes.Params,
		priority TPriority,
		signature [][]byte,
		timeout int64,
		retry int64)
	GetSecondarySignature() [][]byte
}

// The DownloadObjectTask is the interface to record the information for downloading
// pieces of object payload data.
type DownloadObjectTask interface {
	ObjectTask
	InitDownloadObjectTask(object *storagetypes.ObjectInfo,
		params *storagetypes.Params,
		priority TPriority,
		low int64,
		high int64,
		timeout int64,
		retry int64)
	GetBucketInfo() *storagetypes.BucketInfo
	GetUserAddress() string
	// GetSize returns the download payload data size, high - low + 1.
	GetSize() int64
	// GetLow returns the start offset of download payload data.
	GetLow() int64
	// GetHigh returns the end offset of download payload data.
	GetHigh() int64
}

type ChallengePieceTask interface {
	ObjectTask
	InitChallengePieceTask(object *storagetypes.ObjectInfo,
		bucket *storagetypes.BucketInfo,
		priority TPriority,
		replicateIdx int32,
		segmentIdx uint32,
		timeout int64,
		retry int64)
	GetBucketInfo() *storagetypes.BucketInfo
	SetBucketInfo(*storagetypes.BucketInfo)
	GetUserAddress() string
	SetUserAddress(string)
	GetSegmentIdx() uint32
	SetSegmentIdx(uint32)
	GetRedundancyIdx() int32
	SetRedundancyIdx(idx int32)
	GetIntegrityHash() []byte
	SetIntegrityHash([]byte)
	GetPieceHash() [][]byte
	SetPieceHash([][]byte)
	GetPieceDataSize() int64
	SetPieceDataSize(int64)
}

// The GCTask is the interface to record the information of garbage collection.
type GCTask interface {
	Task
}

// The GCObjectTask is the interface to record the information for collecting the
// piece store space by deleting object payload data that the object has been deleted
// on the Greenfield chain.
type GCObjectTask interface {
	GCTask
	InitGCObjectTask(priority TPriority, start, end uint64, timeout int64)
	// SetStartBlockNumber sets start block number for collecting object.
	SetStartBlockNumber(uint64)
	// GetStartBlockNumber returns start block number for collecting object.
	GetStartBlockNumber() uint64
	// SetEndBlockNumber sets end block number for collecting object.
	SetEndBlockNumber(uint64)
	// GetEndBlockNumber returns end block number for collecting object.
	GetEndBlockNumber() uint64
	SetCurrentBlockNumber(uint64)
	GetCurrentBlockNumber() uint64
	GetDeletingObjectId() uint64
	SetDeletingObjectId(uint64)
	// GetGCObjectProcess returns the process of collecting object, returns the
	// deleting block number and the last deleted object id.
	GetGCObjectProcess() (uint64, uint64)
	// SetGCObjectProcess sets the process of collecting object, params stand
	// the deleting block number and the last deleted object id.
	SetGCObjectProcess(uint64, uint64)
}

// The GCZombiePieceTask is the interface to record the information for collecting
// the piece store space by deleting zombie pieces data that dues to any exception,
// the piece data meta is not on chain but the pieces has been store in piece store.
type GCZombiePieceTask interface {
	GCTask
	// GetGCZombiePieceStatus returns the status of collecting zombie pieces, returns
	// the last deleted object id and the number that has been deleted.
	GetGCZombiePieceStatus() (uint64, uint64)
	// SetGCZombiePieceStatus sets the status of collecting zombie pieces, param
	// stands the last deleted object id and the has been deleted pieces number.
	SetGCZombiePieceStatus(uint64, uint64)
}

// The GCMetaTask is the interface to record the information for collecting the SP
// meta store space by deleting the expired data.
type GCMetaTask interface {
	GCTask
	// GetGCMetaStatus returns the status of collecting metadata, returns the last
	// deleted object id and the number that has been deleted.
	GetGCMetaStatus() (uint64, uint64)
	// SetGCMetaStatus sets the status of collecting metadata, parma stands the last
	// deleted object id and the number that has been deleted.
	SetGCMetaStatus(uint64, uint64)
}
