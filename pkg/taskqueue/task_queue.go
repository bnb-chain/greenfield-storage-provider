package task

import (
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// Task defines the smallest unit of SP background service interaction.
// there are two types of task, ObjectTask and GCTask.
//
// The ObjectTask include: UploadObjectTask, ReplicatePieceTask, SealObjectTask,
// DownloadObjectTask and ReceiveTask.These tasks associated with an object, and
// records the information of different stages of the object.
//
// The GCTask include: GCObjectTask, GCZombiePieceTask and GCStoreTask.
// The GCObjectTask collects the piece store space by deleting object payload data
// that the object has been deleted on the chain.
// The GCZombiePieceTask collects the piece store space by deleting zombie pieces
// data that dues to any exception, the piece data meta is not on chain.
// The GCStoreTask collects the SP meta store space by deleting the expired data.
type Task interface {
	// Key returns the uniquely identify of task.
	Key() TKey
	// Type returns the type of task, such as TypeTaskUploadObject etc.
	Type() TType
	// GetCreateTime returns the creation time of task.
	GetCreateTime() int64
	// SetCreateTime sets the creation time of task.
	SetCreateTime(int64)
	// GetUpdateTime returns the last updated time of task.
	GetUpdateTime() int64
	// SetUpdateTime sets last updated time of task.
	SetUpdateTime(int64)
	// GetTimeout returns the timeout of task.
	// notice: the timeout is a duration, not the end timestamp.
	GetTimeout() int64
	// SetTimeout sets timeout duration of task.
	SetTimeout(int64)
	// Expired returns an indicator whether timeout.
	Expired() bool
	// GetPriority returns the priority of task.
	GetPriority() TPriority
	// SetPriority sets the priority of task.
	SetPriority(TPriority)
	// IncRetry increases the retry counter of task.
	IncRetry() bool
	// RetryExceed returns an indicator whether retry max times.
	RetryExceed() bool
	// GetRetry returns the retry counter of task.
	GetRetry() int64
	// GetRetryLimit returns the limit of retry counter.
	GetRetryLimit() int64
	// SetRetryLimit sets the limit of retry counter.
	SetRetryLimit(int64)
	// LimitEstimate returns estimated resource limit.
	LimitEstimate() rcmgr.Limit
	// Error returns the task error.
	Error() error
	// SetError sets the error to task.
	SetError(err error)
}

// TKey defines the type of task key that is the uniquely identify.
type TKey string

// TType is enum type, it defines the type of task.
type TType int32

const (
	// TypeTaskUnknown defines the default task type.
	TypeTaskUnknown TType = iota
	// TypeTaskUploadObject defines the type of uploading object to primary SP task.
	TypeTaskUploadObject
	// TypeTaskReplicatePiece defines the type of replicating pieces to secondary SPs task.
	TypeTaskReplicatePiece
	// TypeTaskSealObject defines the type of sealing object to the chain task.
	TypeTaskSealObject
	// TypeTaskReceivePiece defines the type of receiving pieces for secondary SP task.
	TypeTaskReceivePiece
	// TypeTaskDownloadObject defines the type of downloading object task.
	TypeTaskDownloadObject
	// TypeTaskGCObject defines the type of collecting object payload data task.
	TypeTaskGCObject
	// TypeTaskGCZombiePiece defines the type of collecting zombie piece task.
	TypeTaskGCZombiePiece
	// TypeTaskGCStore defines the type of collecting SP metadata task.
	TypeTaskGCStore
)

// TPriority defines the type of task priority, the task priority can use for picking up
// task from priority queue, it can be used as an auxiliary data for the TQueueStrategy
// to pick up the task.
type TPriority uint8

const (
	// UnKnownTaskPriority defines the default task priority.
	UnKnownTaskPriority = 0
	// UnSchedulingPriority defines the task priority that should be never scheduled.
	UnSchedulingPriority = 0
	// MaxTaskPriority defines the max task priority.
	MaxTaskPriority = 255
	// DefaultLargerTaskPriority defines the larger task priority.
	DefaultLargerTaskPriority = 170
	// DefaultSmallerPriority defines the smaller task priority.
	DefaultSmallerPriority = 85
)

// TPriorityLevel defines the type of task priority level. The executor of the task
// will reserve the resources from the resource manager(rcmgr) before execution, and
// the rcmgr can limit the execution of concurrent tasks number according to the task
// priority level.
//
// Example:
//
//		the configuration the rcmgr:
//			[TasksHighPriority: 10, TasksMediumPriority: 20, TasksLowPriority: 2]
//		the executor of the task can run 10 high level task at the same time that the
//			task priority >= DefaultLargerTaskPriority
//	    the executor of the task can run 20 medium level task at the same time that the
//			task priority between (DefaultLargerTaskPriority, DefaultSmallerPriority]
//		the executor of the task can run 2 medium level task at the same time that the
//			task priority < DefaultSmallerPriority
type TPriorityLevel int32

const (
	// TLowPriorityLevel defines the low task priority level.
	TLowPriorityLevel TPriorityLevel = iota
	// THighPriorityLevel defines the high task priority level.
	THighPriorityLevel
)

// ObjectTask defines the task that is associated with an object, and
// records the information of different stages of the object.
type ObjectTask interface {
	Task
	// GetObjectInfo returns the associated object.
	GetObjectInfo() *storagetypes.ObjectInfo
}

// UploadObjectTask defines the task that uploads object payload data to primary SP.
type UploadObjectTask interface {
	ObjectTask
}

// ReplicatePieceTask defines the task that replicates pieces of object to secondary SPs.
type ReplicatePieceTask interface {
	ObjectTask
}

// SealObjectTask defines the task that seals object to the chain.
type SealObjectTask interface {
	ObjectTask
}

// ReceivePieceTask defines the task that receives pieces data, only belong to secondary SP.
type ReceivePieceTask interface {
	ObjectTask
	// GetReplicateIdx returns the replicate index.
	GetReplicateIdx() uint32
	// SetReplicateIdx sets the replicate index by the executor of the task.
	SetReplicateIdx(uint32)
}

// DownloadObjectTask defines the task that download object.
// TODO: refine interface.
type DownloadObjectTask interface {
	ObjectTask
	// GetNeedIntegrity returns an indicator whether needs integrity information.
	// when get challenge info, in addition to piece data, also need integrity.
	GetNeedIntegrity() bool
	// GetSize returns the download payload data size, high - low + 1.
	GetSize() uint64
	// SetLow sets the start offset of download payload data.
	SetLow(uint64)
	// GetLow returns the start offset of download payload data.
	GetLow() uint64
	// SetHigh sets the end offset of download payload data.
	SetHigh(uint64)
	// GetHigh returns the end offset of download payload data.
	GetHigh() uint64
}

// GCTask defines gc task that collects the resources that no longer use.
type GCTask interface {
	Task
}

// GCObjectTask defines the collection of object task.
type GCObjectTask interface {
	GCTask
	// SetStartBlockNumber sets start block number.
	SetStartBlockNumber(uint64)
	// GetStartBlockNumber returns start block number.
	GetStartBlockNumber() uint64
	// SetEndBlockNumber sets end block number.
	SetEndBlockNumber(uint64)
	// GetEndBlockNumber returns end block number.
	GetEndBlockNumber() uint64
	// GetGCObjectProgress returns the process of collection object.
	// returns the deleting block number and the last deleted object id.
	GetGCObjectProgress() (uint64, uint64)
	// SetGCObjectProgress sets the process of collection object.
	// params stand the deleting block number and the last deleted object id.
	SetGCObjectProgress(uint64, uint64)
}

// GCZombiePieceTask defines the collection of zombie piece data task.
type GCZombiePieceTask interface {
	GCTask
	// GetGCZombiePieceStatus returns the status of collecting zombie pieces.
	// returns the number that has been deleted.
	GetGCZombiePieceStatus() uint64
	// SetGCZombiePieceStatus sets the status of collecting zombie pieces.
	// param stands the has been deleted pieces number.
	SetGCZombiePieceStatus(uint64)
}

// GCStoreTask defines the collection of SP metadata task.
type GCStoreTask interface {
	GCTask
	// GetGCStoreStatus returns the status of collecting metadata.
	// returns the number that has been deleted.
	GetGCStoreStatus() (uint64, uint64)
	// SetGCStoreStatus sets the status of collecting metadata.
	// parma stands the number that has been deleted.
	SetGCStoreStatus(uint64, uint64)
}

// TTaskPriority defines the all supported tasks type and its priorities.
type TTaskPriority interface {
	// GetPriority returns the special task type's priority
	GetPriority(TType) TPriority
	// SetPriority sets the special task type's priority
	SetPriority(TType, TPriority)
	// GetAllPriorities returns all supported tasks type and its priorities.
	GetAllPriorities() map[TType]TPriority
	// SetAllPriorities sets all supported tasks type and its priorities.
	SetAllPriorities(map[TType]TPriority)
	// GetLowLevelPriority returns the low level priority's watermark.
	GetLowLevelPriority() TPriority
	// SetLowLevelPriority sets the low level priority's watermark.
	SetLowLevelPriority(TPriority) error
	// GetHighLevelPriority returns the high level priority's watermark.
	GetHighLevelPriority() TPriority
	// SetHighLevelPriority sets the high level priority's watermark.
	SetHighLevelPriority(TPriority) error
	// HighLevelPriority returns an indicator whether high level priority.
	HighLevelPriority(TPriority) bool
	// LowLevelPriority returns an indicator whether low level priority.
	LowLevelPriority(TPriority) bool
	// SupportTask returns an indicator whether support type of task.
	SupportTask(TType) bool
}

// TQueueStrategy defines the strategy operator on queue.
type TQueueStrategy interface {
	// SetPickUpCallback sets the callback func for picking up task.
	SetPickUpCallback(func([]Task) Task)
	// RunPickUpStrategy calls the pickup callback to pick up task on backup task.
	RunPickUpStrategy([]Task) Task
	// SetCollectionCallback sets the callback func for gc task.
	SetCollectionCallback(func(TQueueOnStrategy, []TKey))
	// RunCollectionStrategy calls collection callback func for gc task on queue.
	RunCollectionStrategy(TQueueOnStrategy, []TKey)
	// SetPickUpFilterCallback sets the callback func for picking up filter task.
	SetPickUpFilterCallback(func(Task) bool)
	// RunPickUpFilterStrategy calls the pickup filter callback to pick up filter task.
	RunPickUpFilterStrategy(Task) bool
}

// TQueue defines the task queue operator.
type TQueue interface {
	// Top returns the top task in queue.
	Top() Task
	// Pop pops the top task in queue.
	Pop() Task
	// PopByKey pops the task by task key.
	PopByKey(TKey) Task
	// Has returns an indicator whether the task in queue.
	Has(TKey) bool
	// Push pushes the task in queue.
	Push(Task) error
	// PopPush pops task by task key, and pushes task again
	PopPush(Task) error
	// Len returns the length of queue.
	Len() int
	// Cap returns the capacity of queue.
	Cap() int
}

// TQueueOnStrategy defines queue operator that supports TQueueStrategy.
type TQueueOnStrategy interface {
	TQueue
	// SetStrategy sets TQueueStrategy on the queue.
	SetStrategy(TQueueStrategy)
	// Expired returns an indicator whether the task is timeout.
	Expired(TKey) bool
	// IsActiveTask returns an indicator whether active task.
	IsActiveTask(TKey) bool
	// RunCollection calls TQueueStrategy's callback to collect tasks.
	RunCollection()
}

// TQueueWithLimit defines the operator of queue that supports resource limits.
type TQueueWithLimit interface {
	TQueueOnStrategy
	// PopByLimit returns the task that consumes less resources than the param.
	PopByLimit(rcmgr.Limit) Task
	// GetSupportPickUpByLimit returns an indicator whether supports resource limits.
	GetSupportPickUpByLimit() bool
	// SetSupportPickUpByLimit set whether supports resource limits.
	// if false, PopWithLimit equal Pop
	SetSupportPickUpByLimit(bool)
}

// TPriorityQueueWithLimit defines the priority queue, its sub queue supports TQueueWithLimit.
type TPriorityQueueWithLimit interface {
	TQueueWithLimit
	// SubQueueLen returns the length of sub queue.
	SubQueueLen(TPriority) int
	// GetPriorities returns all supports priorities.
	GetPriorities() []TPriority
	//SetPriorityQueue sets the sub queue with priority.
	SetPriorityQueue(TPriority, TQueueWithLimit) error
}
