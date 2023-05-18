package task

// TKey defines the type of task key that is the uniquely identify.
type TKey string

// String transfer TKey to string type.
func (k TKey) String() string {
	return string(k)
}

// TType is enum type, it defines the type of task.
type TType int32

const (
	// TypeTaskUnknown defines the default task type.
	TypeTaskUnknown TType = iota
	// TypeTaskCreateBucketApproval defines the type of asking create bucket approval
	// to primary SP task
	TypeTaskCreateBucketApproval
	// TypeTaskCreateObjectApproval defines the type of asking create object approval
	// to primary SP task
	TypeTaskCreateObjectApproval
	// TypeTaskReplicatePieceApproval defines the type of asking create object approval
	// to secondary SP task
	TypeTaskReplicatePieceApproval
	// TypeTaskUpload defines the type of uploading object to primary SP task.
	TypeTaskUpload
	// TypeTaskReplicatePiece defines the type of replicating pieces to secondary SPs
	// task.
	TypeTaskReplicatePiece
	// TypeTaskSealObject defines the type of sealing object to the chain task.
	TypeTaskSealObject
	// TypeTaskReceivePiece defines the type of receiving pieces for secondary SP task.
	TypeTaskReceivePiece
	// TypeTaskDownloadObject defines the type of downloading object task.
	TypeTaskDownloadObject
	// TypeTaskChallengePiece defines the type of challenging piece task.
	TypeTaskChallengePiece
	// TypeTaskGCObject defines the type of collecting object payload data task.
	TypeTaskGCObject
	// TypeTaskGCZombiePiece defines the type of collecting zombie piece task.
	TypeTaskGCZombiePiece
	// TypeTaskGCMeta defines the type of collecting SP metadata task.
	TypeTaskGCMeta
)

var TypeTaskMap = map[TType]string{
	TypeTaskUnknown:                "UnknownTask",
	TypeTaskCreateBucketApproval:   "CreateBucketApprovalTask",
	TypeTaskCreateObjectApproval:   "CreateObjectApprovalTask",
	TypeTaskReplicatePieceApproval: "ReplicatePieceApprovalTask",
	TypeTaskUpload:                 "UploadObjectTask",
	TypeTaskReplicatePiece:         "ReplicatePieceTask",
	TypeTaskSealObject:             "SealObjectTask",
	TypeTaskReceivePiece:           "ReceivePieceTask",
	TypeTaskDownloadObject:         "DownloadObjectTask",
	TypeTaskChallengePiece:         "ChallengePieceTask",
	TypeTaskGCObject:               "GCObjectTask",
	TypeTaskGCZombiePiece:          "GCZombiePieceTask",
	TypeTaskGCMeta:                 "GCMetaTask",
}

func TaskTypeName(taskType TType) string {
	if _, ok := TypeTaskMap[taskType]; !ok {
		return TypeTaskMap[TypeTaskUnknown]
	}
	return TypeTaskMap[taskType]
}

// TPriority defines the type of task priority, the priority can be used as an important
// basis for task scheduling within the SP. The higher the priority, the faster it is
// expected to be executed, and the resources will be assigned priority for execution.
// The lower the priority, it can be executed later, and the resource requirements are
// not so urgent.
type TPriority uint8

const (
	// UnKnownTaskPriority defines the default task priority.
	UnKnownTaskPriority TPriority = 0
	// UnSchedulingPriority defines the task priority that should be never scheduled.
	UnSchedulingPriority TPriority = 0
	// MaxTaskPriority defines the max task priority.
	MaxTaskPriority TPriority = 255
	// DefaultLargerTaskPriority defines the larger task priority.
	DefaultLargerTaskPriority TPriority = 170
	// DefaultSmallerPriority defines the smaller task priority.
	DefaultSmallerPriority TPriority = 85
)

// TPriorityLevel defines the type of task priority level. The executor of the task
// will reserve the resources from the resource manager(rcmgr) before execution, and
// the rcmgr can limit the execution of concurrent tasks number according to the task
// priority level.
//
// Example:
//
//		the configuration the rcmgr:
//			[TasksHighPriority: 30, TasksMediumPriority: 20, TasksLowPriority: 2]
//		the executor of the task can run 30 high level tasks at the same time that the
//			task priority >= DefaultLargerTaskPriority
//	 	the executor of the task can run 20 medium level tasks at the same time that the
//			task priority between (DefaultLargerTaskPriority, DefaultSmallerPriority]
//		the executor of the task can run 2 medium level tasks at the same time that the
//			task priority < DefaultSmallerPriority
type TPriorityLevel int32

const (
	// TLowPriorityLevel defines the low task priority level.
	TLowPriorityLevel TPriorityLevel = iota
	// TMediumPriorityLevel defines the medium task priority level.
	TMediumPriorityLevel
	// THighPriorityLevel defines the high task priority level.
	THighPriorityLevel
)
