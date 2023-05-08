package task

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// Task is the interface to the smallest unit of SP background service interaction.
//
// Task Type:
//
//	There are three main types of task, ApprovalTask, ObjectTask and GCTask.
//	The ApprovalTask is used to record the ask approval information, for user
//	creating bucket and object need ask primary SP approval if willing serve
//	the bucket and object, the SP will sign the approval msg if it approved
//	the msg, and the greenfield will verify the signature of the approval msg
//	to judge whether SP accepts the bucket and object, for primary replicating
//	pieces to the secondary SPs need broadcast the approval msg to other SPs,
//	if they approved the msg, the primary SP will pick up some of them that
//	approved the msg and replicate the pieces to the these, and they will verify
//	the signature of the approval msg before receive the pieces. so the
//	ApprovalTask includes ApprovalCreateBucketTask, ApprovalCreateBucketTask and
//	ApprovalReplicatePieceTask.
//	The ObjectTask associated with an object, and records the information of
//	different stages of the object, includes UploadObjectTask stands upload the
//	object payload data to the primary SP, ReplicatePieceTask stands replicate
//	the object pieces to the secondary SPs, ReceivePieceTask only belong to the
//	secondary SP, records the information of receiving piece and the secondary SP
//	use it to confirm the object if success to seal on the greenfield, this will
//	guarantee a return of the secondary SP. SealObjectTask stands seal object on
//	the greenfield, DownloadObjectTask stands the user download the part or all
//	object payload data, ChallengePieceTask stands the validator get the challenge
//	piece info, the validator to challenge the SP if store the user's payload data
//	correctly by this way.
//	The GCTask is the interface to record the information of garbage collection,
//	includes GCObjectTask stands the collection of piece store space by deleting
//	the payload data that has been deleted on the greenfield, GCZombiePieceTask
//	stands the collection of piece store space by deleting zombie pieces data that
//	dues to any exception, the piece data meta is not on the greenfield, GCMetaTask
//	stands the collection of the SP meta store space by deleting the expired data.
//
// Task Priority:
//
//	Each type of task has a priority, the range of priority is [0, 255], the higher
//	the priority, the higher the urgency to be executed, the greater the probability
//	of being executed by priority scheduling.
//
// Task Priority Level:
//
//	Task priority is divided into three levels, TLowPriorityLevel, TMediumPriorityLevel,
//	THighPriorityLevel. The TLowPriorityLevel default priority range is [0, 85), The
//	TMediumPriorityLevel default priority range is [85, 170), The THighPriorityLevel
//	default priority range is [170, 256). When allocating for task execution resources
//	from ResourceManager, the resources are allocated according to task priority level,
//	but not task priority, because task priority up to 256 levels, the task priority
//	level make resource management easier.
//	Example:
//		the resource limit configuration of task execution node :
//			[TasksHighPriority: 30, TasksMediumPriority: 20, TasksLowPriority: 2]
//		the executor of the task can run 30 high level tasks at the same time that the
//			task priority >= 170
//	 	the executor of the task can run 20 medium level tasks at the same time that the
//			task priority between [85, 170)
//		the executor of the task can run 2 medium level tasks at the same time that the
//			task priority < 85
//
// Task Init:
//
//	Each task needs to call its InitXXX method before use. This method requires passing
//	in the necessary parameters of each type of task. These parameters will not be changed
//	in most cases and are necessary, such as task priority, timeout, max retries, and
//	necessary information for resource estimation.
//	Any changes to initialization parameters during task execution may cause unpredictable
//	consequences. For example, changes in parameters that affect resource estimation may
//	lead to OOM, etc.
type Task interface {
	// Key returns the uniquely identify of the task. It is recommended that each task
	// has its own prefix. In addition, it should also include the information of the
	// task's own identity.
	// Example:
	//	the ApprovalTask maybe includes the bucket and object name,
	//	the ObjectTask maybe includes the object ID,
	// 	the GCTask maybe includes the timestamp.
	Key() TKey
	// Type returns the type of the task. A task has a unique type, such as
	// TypeTaskCreateBucketApproval, TypeTaskUpload etc. has the only one TType
	// definition.
	Type() TType
	// GetAddress returns the task runner address. there is only one runner at the
	// same time, which will assist in quickly locating the running node of the task.
	GetAddress() string
	// SetAddress sets the runner address to the task.
	SetAddress(string)
	// GetCreateTime returns the creation time of the task. The creation time used to
	// judge task execution time.
	GetCreateTime() int64
	// SetCreateTime sets the creation time of the tas.
	SetCreateTime(int64)
	// GetUpdateTime returns the last updated time of the task. The updated time used
	// to determine whether the task is expired with the timeout.
	GetUpdateTime() int64
	// SetUpdateTime sets last updated time of the task. Any changes in task information
	// requires to set the update time.
	SetUpdateTime(int64)
	// GetTimeout returns the timeout of the task, the timeout is a duration, if update
	// time adds timeout lesser now stands the task is expired.
	GetTimeout() int64
	// SetTimeout sets timeout duration of the task.
	SetTimeout(int64)
	// ExceedTimeout returns an indicator whether timeout, if update time adds timeout
	// lesser now returns true, otherwise returns false.
	ExceedTimeout() bool
	// GetMaxRetry returns the max retry times of the task. Each type of task has a
	// fixed max retry times.
	GetMaxRetry() int64
	// SetMaxRetry sets the max retry times of the task.
	SetMaxRetry(int64)
	// GetRetry returns the retry counter of the task.
	GetRetry() int64
	// SetRetry sets the retry counter of the task.
	SetRetry(int)
	// IncRetry increases the retry counter of the task. Each task has the max retry
	// times, if retry counter exceed the max retry, the task should be canceled.
	IncRetry()
	// ExceedRetry returns an indicator whether retry counter greater that max retry.
	ExceedRetry() bool
	// Expired returns an indicator whether ExceedTimeout and ExceedRetry.
	Expired() bool
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

// ApprovalTask is the interface to record the ask approval information, the
// approval task timeliness uses the block height, if reached expired height,
// the approval invalid.
type ApprovalTask interface {
	Task
	// GetExpiredHeight returns the expired height of the approval.
	GetExpiredHeight() uint64
	// SetExpiredHeight sets the expired height of the approval, when SP
	// approved the approval, it should set the expired height to stands
	// the approval timeliness. This is one of the ways SP prevents being
	// attacked.
	SetExpiredHeight(uint64)
}

// ApprovalCreateBucketTask is the interface to record the ask create bucket
// approval information. The user account will create MsgCreateBucket, the SP
// should decide whether approved the request based on the MsgCreateBucket.
// If so, the sp will SetExpiredHeight and signs the MsgCreateBucket.
type ApprovalCreateBucketTask interface {
	ApprovalTask
	// InitApprovalCreateBucketTask inits the ApprovalCreateBucketTask by
	// MsgCreateBucket and task priority. the SP only fill the MsgCreateBucket's
	// PrimarySpApproval field, can not change other fields.
	InitApprovalCreateBucketTask(*storagetypes.MsgCreateBucket, TPriority)
	// GetCreateBucketInfo returns the user's MsgCreateBucket.
	GetCreateBucketInfo() *storagetypes.MsgCreateBucket
	// SetCreateBucketInfo sets the MsgCreateBucket. Should try to avoid calling
	// this method, it will change the approval information.
	SetCreateBucketInfo(*storagetypes.MsgCreateBucket)
}

// ApprovalCreateObjectTask is the interface to record the ask create object
// approval information. The user account will create MsgCreateObject, the SP
// should decide whether approved the request based on the MsgCreateObject.
// If so, the sp will SetExpiredHeight and signs the MsgCreateObject.
type ApprovalCreateObjectTask interface {
	ApprovalTask
	// InitApprovalCreateObjectTask inits the ApprovalCreateObjectTask by
	// MsgCreateObject and task priority. the SP only fill the MsgCreateObject's
	// PrimarySpApproval field, can not change other fields.
	InitApprovalCreateObjectTask(*storagetypes.MsgCreateObject, TPriority)
	// GetCreateObjectInfo returns the user's MsgCreateObject.
	GetCreateObjectInfo() *storagetypes.MsgCreateObject
	// SetCreateObjectInfo sets the MsgCreateObject. Should try to avoid calling
	// this method, it will change the approval information.
	SetCreateObjectInfo(*storagetypes.MsgCreateObject)
}

// ApprovalReplicatePieceTask is the interface to record the ask replicate pieces
// to other SPs(as secondary SP for the object). It is initiated by the primary SP
// in the replicate pieces phase.  Before the primary SP sends it to other SPs, the
// primary SP will sign the task, other SPs will verify it is sent by a legitimate
// SP. If other SPs approved the approval, they will SetExpiredHeight and signs the
// ApprovalReplicatePieceTask.
type ApprovalReplicatePieceTask interface {
	ObjectTask
	ApprovalTask
	// InitApprovalReplicatePieceTask inits the ApprovalReplicatePieceTask by ObjectInfo,
	// storage params, task priority and primary operator address. the storage params
	// can affect the size of the data accepted by the secondary SP, so this is a necessary
	// and cannot be changed parameter.
	InitApprovalReplicatePieceTask(
		object *storagetypes.ObjectInfo,
		params *storagetypes.Params,
		priority TPriority,
		askOpAddress string)
	// GetAskSpOperatorAddress returns the SP's operator address that initiated the ask
	// replicate piece approval request.
	GetAskSpOperatorAddress() string
	// SetAskSpOperatorAddress sets the SP's operator address that initiated the ask
	// replicate piece approval request. Should try to avoid calling this method,
	// it will change the approval information.
	SetAskSpOperatorAddress(string)
	// GetAskSignature returns the initiated SP's signature by its operator private key.
	GetAskSignature() []byte
	// SetAskSignature sets the initiated SP's signature by its operator private key.
	SetAskSignature([]byte)
	// GetApprovedSpOperatorAddress returns the approved SP's operator address.
	GetApprovedSpOperatorAddress() string
	// SetApprovedSpOperatorAddress sets the approved SP's operator address.
	SetApprovedSpOperatorAddress(string)
	// GetApprovedSignature returns the approved SP's signature.
	GetApprovedSignature() []byte
	// SetApprovedSignature sets the approved SP's signature.
	SetApprovedSignature([]byte)
	// GetApprovedSpEndpoint returns the approved SP's endpoint. It is used to replicate
	// pieces to secondary SP.
	GetApprovedSpEndpoint() string
	// SetApprovedSpEndpoint sets the approved SP's endpoint.
	SetApprovedSpEndpoint(string)
	// GetApprovedSpApprovalAddress returns the approved SP's approval address. It is
	// used to seal object on greenfield.
	GetApprovedSpApprovalAddress() string
	// SetApprovedSpApprovalAddress sets the approved SP's approval address.
	SetApprovedSpApprovalAddress(string)
	// GetSignBytes returns the bytes from the task for initiated and approved SPs
	// to sign.
	GetSignBytes() []byte
}

// The ObjectTask associated with an object and storage params, and records the
// information of different stages of the object. Considering the change of storage
// params on the greenfield, the storage params of each object should be determined
// when it is created, and it should not be queried during the task flow, which is
// inefficient and error-prone.
type ObjectTask interface {
	Task
	// GetObjectInfo returns the associated object.
	GetObjectInfo() *storagetypes.ObjectInfo
	// SetObjectInfo set the  associated object.
	SetObjectInfo(*storagetypes.ObjectInfo)
	// GetStorageParams returns the storage params.
	GetStorageParams() *storagetypes.Params
	// SetStorageParams sets the storage params.Should try to avoid calling this
	// method, it will change the task base information.
	// Example:
	// 	it will change resource estimate for UploadObjectTask and so on.
	SetStorageParams(*storagetypes.Params)
}

// The UploadObjectTask is the interface to record the information for uploading object
// payload data to the primary SP.
type UploadObjectTask interface {
	ObjectTask
	// InitUploadObjectTask inits the UploadObjectTask by ObjectInfo and Params.
	InitUploadObjectTask(*storagetypes.ObjectInfo, *storagetypes.Params)
}

// The ReplicatePieceTask is the interface to record the information for replicating
// pieces of object pieces data to secondary SPs.
type ReplicatePieceTask interface {
	ObjectTask
	// InitReplicatePieceTask inits the ReplicatePieceTask by ObjectInfo, params,
	// task priority, timeout and max retry.
	InitReplicatePieceTask(
		object *storagetypes.ObjectInfo,
		params *storagetypes.Params,
		priority TPriority,
		timeout int64,
		retry int64)
	// GetSealed returns an indicator whether successful seal object on greenfield
	// after replicate pieces, it is an optimization method. ReplicatePieceTask and
	// SealObjectTask are combined. Otherwise, the two tasks will be completed in
	// two stages. If the combination is successful and the seal object is successful,
	// the number of SealObjectTask can be reduced, saving resource overhead.
	GetSealed() bool
	// SetSealed sets the state successful seal object after replicating piece.
	SetSealed(bool)
	// GetSecondarySignature returns the secondary SP's signatures. It is used to
	// generate MsgSealObject.
	GetSecondarySignature() [][]byte
	// SetSecondarySignature sets the the secondary SP's signatures.
	SetSecondarySignature([][]byte)
}

// The ReceivePieceTask is the interface to record the information for receiving pieces
// of object payload data from primary SP, it exists only in secondary SP.
type ReceivePieceTask interface {
	ObjectTask
	// InitReceivePieceTask init the ReceivePieceTask.
	InitReceivePieceTask(
		object *storagetypes.ObjectInfo,
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
	// GetPieceSize returns the received piece data size, it is used to resource estimate.
	GetPieceSize() int64
	// SetPieceSize sets the received piece data size.
	SetPieceSize(int64)
	// GetPieceChecksum returns the checksum of received piece data, it is used to check
	// the piece data is correct.
	GetPieceChecksum() []byte
	// SetPieceChecksum set the checksum of received piece data.
	SetPieceChecksum([]byte)
	// GetSignature returns the primary SP's signature, because the InitReceivePieceTask
	// will be transfer to secondary SP's, It is necessary to prove that the task was
	// sent by a legitimate SP.
	GetSignature() []byte
	// SetSignature sets the primary SP's signature.
	SetSignature([]byte)
	// GetSignBytes returns the bytes from the task for primary SP to sign.
	GetSignBytes() []byte
	// GetSealed returns an indicator whether the object of receiving piece data is
	// sealed on greenfield, the secondary SP has an incentive to confirm that otherwise
	// it wastes its storage resources
	GetSealed() bool
	// SetSealed sets the object of receiving piece data whether is successfully sealed.
	SetSealed(bool)
}

// The SealObjectTask is the interface to  record the information for sealing object to
// the greenfield.
type SealObjectTask interface {
	ObjectTask
	// InitSealObjectTask inits the SealObjectTask.
	InitSealObjectTask(
		object *storagetypes.ObjectInfo,
		params *storagetypes.Params,
		priority TPriority,
		signature [][]byte,
		timeout int64,
		retry int64)
	// GetSecondarySignature returns the secondary SP's signature, it is used to generate
	// MsgSealObject.
	GetSecondarySignature() [][]byte
}

// The DownloadObjectTask is the interface to record the information for downloading
// pieces of object payload data.
type DownloadObjectTask interface {
	ObjectTask
	// InitDownloadObjectTask inits DownloadObjectTask.
	InitDownloadObjectTask(object *storagetypes.ObjectInfo,
		params *storagetypes.Params,
		priority TPriority,
		userAddress string,
		low int64,
		high int64,
		timeout int64,
		retry int64)
	// GetBucketInfo returns the BucketInfo of the download object.
	// It is used to Query and calculate bucket read quota.
	GetBucketInfo() *storagetypes.BucketInfo
	// GetUserAddress returns the user account of downloading object.
	// It is used to records the read bucket information.
	GetUserAddress() string
	// SetUserAddress sets the user account of downloading object.
	SetUserAddress(string)
	// GetSize returns the download payload data size, high - low + 1.
	GetSize() int64
	// GetLow returns the start offset of download payload data.
	GetLow() int64
	// GetHigh returns the end offset of download payload data.
	GetHigh() int64
}

// ChallengePieceTask is the interface to record the information for get challenge
// piece info, the validator get challenge info to confirm whether the sp stores
// the user's data correctly.
type ChallengePieceTask interface {
	ObjectTask
	// InitChallengePieceTask inits InitChallengePieceTask.
	InitChallengePieceTask(
		object *storagetypes.ObjectInfo,
		bucket *storagetypes.BucketInfo,
		priority TPriority,
		userAddress string,
		replicateIdx int32,
		segmentIdx uint32,
		timeout int64,
		retry int64)
	// GetBucketInfo returns the BucketInfo of challenging piece
	GetBucketInfo() *storagetypes.BucketInfo
	// SetBucketInfo sets the the BucketInfo of challenging piece
	SetBucketInfo(*storagetypes.BucketInfo)
	// GetUserAddress returns the user account of challenging object.
	// It is used to records the read bucket information.
	GetUserAddress() string
	// SetUserAddress sets the user account of challenging object.
	SetUserAddress(string)
	// GetSegmentIdx returns the segment index of challenge piece.
	GetSegmentIdx() uint32
	// SetSegmentIdx sets the segment index of challenge piece.
	SetSegmentIdx(uint32)
	// GetRedundancyIdx returns the replicate index of challenge piece.
	GetRedundancyIdx() int32
	// SetRedundancyIdx sets the replicate index of challenge piece.
	SetRedundancyIdx(idx int32)
	// GetIntegrityHash returns the integrity hash of the object.
	GetIntegrityHash() []byte
	// SetIntegrityHash sets the integrity hash of the object.
	SetIntegrityHash([]byte)
	// GetPieceHash returns the hash of  challenge piece.
	GetPieceHash() [][]byte
	// SetPieceHash sets the hash of  challenge piece.
	SetPieceHash([][]byte)
	// GetPieceDataSize returns the data of challenge piece.
	GetPieceDataSize() int64
	// SetPieceDataSize sets the data of challenge piece.
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
	// InitGCObjectTask inits InitGCObjectTask.
	InitGCObjectTask(priority TPriority, start, end uint64, timeout int64)
	// SetStartBlockNumber sets start block number for collecting object.
	SetStartBlockNumber(uint64)
	// GetStartBlockNumber returns start block number for collecting object.
	GetStartBlockNumber() uint64
	// SetEndBlockNumber sets end block number for collecting object.
	SetEndBlockNumber(uint64)
	// GetEndBlockNumber returns end block number for collecting object.
	GetEndBlockNumber() uint64
	// SetCurrentBlockNumber sets the collecting block number.
	SetCurrentBlockNumber(uint64)
	// GetCurrentBlockNumber returns the collecting block number.
	GetCurrentBlockNumber() uint64
	// GetDeletingObjectId returns the lasted deleting ObjectID.
	GetDeletingObjectId() uint64
	// SetDeletingObjectId sets the lasted deleting ObjectID.
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
