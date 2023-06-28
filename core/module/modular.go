package module

import (
	"context"
	"io"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/core/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// Modular is a common interface for submodules that are scheduled by the GfSp framework.
// It inherits lifecycle.Service interface, which is used to manage lifecycle of services. Additionally, Modular is managed
// by ResourceManager, which allows the GfSp framework to reserve and release resources from the Modular resource pool.
type Modular interface {
	lifecycle.Service
	// ReserveResource reserves the resources from Modular resources pool.
	ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error)
	// ReleaseResource releases the resources to Modular resources pool.
	ReleaseResource(ctx context.Context, scope rcmgr.ResourceScopeSpan)
}

// AuthOpType defines the operator type used to authentication verification.
type AuthOpType int32

const (
	// AuthOpTypeUnKnown defines the default value of AuthOpType
	AuthOpTypeUnKnown AuthOpType = iota
	// AuthOpAskCreateBucketApproval defines the AskCreateBucketApproval operator
	AuthOpAskCreateBucketApproval
	// AuthOpAskMigrateBucketApproval defines the AskMigrateBucketApproval operator
	AuthOpAskMigrateBucketApproval
	// AuthOpAskCreateObjectApproval defines the AskCreateObjectApproval operator
	AuthOpAskCreateObjectApproval
	// AuthOpTypeGetChallengePieceInfo defines the GetChallengePieceInfo operator
	AuthOpTypeGetChallengePieceInfo
	// AuthOpTypePutObject defines the PutObject operator
	AuthOpTypePutObject
	// AuthOpTypeGetObject defines the GetObject operator
	AuthOpTypeGetObject
	// AuthOpTypeGetUploadingState defines the GetUploadingState operator
	AuthOpTypeGetUploadingState
	// AuthOpTypeGetBucketQuota defines the GetBucketQuota operator
	AuthOpTypeGetBucketQuota
	// AuthOpTypeListBucketReadRecord defines the ListBucketReadRecord operator
	AuthOpTypeListBucketReadRecord
	// AuthOpTypeGetRecoveryPiece defines the GetRecoveryPiece operator
	AuthOpTypeGetRecoveryPiece
)

// Authenticator is an abstract interface to verify users authentication.
type Authenticator interface {
	Modular
	// VerifyAuthentication verifies the operator authentication.
	VerifyAuthentication(ctx context.Context, auth AuthOpType, account, bucket, object string) (bool, error)
	// GetAuthNonce get the auth nonce for which the dApp or client can generate EDDSA key pairs.
	GetAuthNonce(ctx context.Context, account string, domain string) (*spdb.OffChainAuthKey, error)
	// UpdateUserPublicKey updates the user public key once the dApp or client generates the EDDSA key pairs.
	UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32,
		userPublicKey string, expiryDate int64) (bool, error)
	// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
	VerifyOffChainSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign string) (bool, error)
}

// Approver is an abstract interface to handle asking approval requests.
type Approver interface {
	Modular
	// PreCreateBucketApproval prepares to handle CreateBucketApproval, it can do some
	// checks such as checking for duplicates, if limitation of SP has been reached, etc.
	PreCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask) error
	// HandleCreateBucketApprovalTask handles the CreateBucketApproval, it can set expired height, sign the MsgCreateBucket and so on.
	HandleCreateBucketApprovalTask(ctx context.Context, task task.ApprovalCreateBucketTask) (bool, error)
	// PostCreateBucketApproval is called after HandleCreateBucketApprovalTask, it can recycle resources, make statistics
	// and do some other operations.
	PostCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask)

	// PreMigrateBucketApproval prepares to handle MigrateBucketApproval, it can do some
	// checks such as checking for duplicates, if limitation of SP has been reached, etc.
	PreMigrateBucketApproval(ctx context.Context, task task.ApprovalMigrateBucketTask) error
	// HandleMigrateBucketApprovalTask handles the MigrateBucketApproval, it can set expired height, sign the MsgMigrateBucket and so on.
	HandleMigrateBucketApprovalTask(ctx context.Context, task task.ApprovalMigrateBucketTask) (bool, error)
	// PostMigrateBucketApproval is called after HandleMigrateBucketApprovalTask, it can recycle resources, make statistics
	// and do some other operations.
	PostMigrateBucketApproval(ctx context.Context, task task.ApprovalMigrateBucketTask)

	// PreCreateObjectApproval prepares to handle CreateObjectApproval, it can do some
	// checks such as check for duplicates, if limitation of SP has been reached, etc.
	PreCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask) error
	// HandleCreateObjectApprovalTask handles the CreateObjectApproval, it can set expired height, sign the MsgCreateObject and so on.
	HandleCreateObjectApprovalTask(ctx context.Context, task task.ApprovalCreateObjectTask) (bool, error)
	// PostCreateObjectApproval is called after HandleCreateObjectApprovalTask, it can
	// recycle resources, make statistics and do some other operations.
	PostCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask)
	// QueryTasks queries tasks that running on approver by task sub-key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}

// Downloader is an abstract interface to handle getting object requests from users' account, and getting
// challenge info requests from other components in the system.
type Downloader interface {
	Modular
	// PreDownloadObject prepares to handle DownloadObject, it can do some checks
	// such as checking for duplicates, if limitation of SP has been reached, etc.
	PreDownloadObject(ctx context.Context, task task.DownloadObjectTask) error
	// HandleDownloadObjectTask handles the DownloadObject and get data from piece store.
	HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) ([]byte, error)
	// PostDownloadObject is called after HandleDownloadObjectTask, it can recycle
	// resources, make statistics and do some other operations..
	PostDownloadObject(ctx context.Context, task task.DownloadObjectTask)

	// PreDownloadPiece prepares to handle DownloadPiece, it can do some checks such as check for duplicates,
	// if limitation of SP has been reached, etc.
	PreDownloadPiece(ctx context.Context, task task.DownloadPieceTask) error
	// HandleDownloadPieceTask handles the DownloadPiece and get data from piece store.
	HandleDownloadPieceTask(ctx context.Context, task task.DownloadPieceTask) ([]byte, error)
	// PostDownloadPiece is called after HandleDownloadPieceTask, it can recycle
	// resources, make statistics and do some other operations.
	PostDownloadPiece(ctx context.Context, task task.DownloadPieceTask)

	// PreChallengePiece prepares to handle ChallengePiece, it can do some checks
	// such as checking for duplicates, if limitation of SP has been reached, etc.
	PreChallengePiece(ctx context.Context, task task.ChallengePieceTask) error
	// HandleChallengePiece handles ChallengePiece, get piece data from piece store and get integrity hash from db.
	HandleChallengePiece(ctx context.Context, task task.ChallengePieceTask) ([]byte, [][]byte, []byte, error)
	// PostChallengePiece is called after HandleChallengePiece, it can recycle resources, make statistics
	// and do some other operations.
	PostChallengePiece(ctx context.Context, task task.ChallengePieceTask)
	// QueryTasks queries download/challenge tasks that running on downloader by task sub-key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}

// TaskExecutor is an abstract interface to handle background tasks.
// It will ask tasks from manager modular, handle tasks and report the results or status to the manager modular
// It can handle these tasks: ReplicatePieceTask, SealObjectTask, ReceivePieceTask, GCObjectTask, GCZombiePieceTask, GCMetaTask.
type TaskExecutor interface {
	Modular
	// AskTask asks the task by remaining limitation from manager module.
	AskTask(ctx context.Context) error
	// HandleReplicatePieceTask handles ReplicatePieceTask that is asked from manager module.
	HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask)
	// HandleSealObjectTask handles SealObjectTask that is asked from manager module.
	HandleSealObjectTask(ctx context.Context, task task.SealObjectTask)
	// HandleReceivePieceTask handles the ReceivePieceTask that is asked from manager module.
	// It will confirm the piece data that is synced to secondary SP whether has been sealed.
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask)
	// HandleGCObjectTask handles the GCObjectTask that is asked from manager module.
	HandleGCObjectTask(ctx context.Context, task task.GCObjectTask)
	// HandleGCZombiePieceTask handles the GCZombiePieceTask that is asked from manager module.
	HandleGCZombiePieceTask(ctx context.Context, task task.GCZombiePieceTask)
	// HandleGCMetaTask handles the GCMetaTask that is asked from manager module.
	HandleGCMetaTask(ctx context.Context, task task.GCMetaTask)
	// ReportTask reports the results or status of running task to manager module.
	ReportTask(ctx context.Context, task task.Task) error
}

// Manager is an abstract interface to do some internal services management, it is responsible for task
// scheduling and other management of SP.
type Manager interface {
	Modular
	// DispatchTask dispatches the task to TaskExecutor module when it asks tasks.
	// It will consider task remaining resources when dispatching task.
	DispatchTask(ctx context.Context, limit rcmgr.Limit) (task.Task, error)
	// QueryTasks queries tasks that hold on manager by task sub-key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
	// HandleCreateUploadObjectTask handles the CreateUploadObject request from Uploader, before Uploader handles
	// the users' UploadObject requests, it should send CreateUploadObject requests to Manager ask if it's ok.
	// Through this interface SP implements the global uploading object strategy.
	//
	// For example: control the concurrency of global uploads, avoid repeated uploads, rate control, etc.
	HandleCreateUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error
	// HandleDoneUploadObjectTask handles the result of uploading object payload data to primary, Manager should
	// generate ReplicatePieceTask for TaskExecutor to run.
	HandleDoneUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error
	// HandleCreateResumableUploadObjectTask handles the CreateUploadObject request from
	// Uploader, before Uploader handles the user's UploadObject request, it should
	// send CreateUploadObject request to Manager ask if it's ok. Through this
	// interface that SP implements the global upload object strategy.
	//
	HandleCreateResumableUploadObjectTask(ctx context.Context, task task.ResumableUploadObjectTask) error

	// HandleDoneResumableUploadObjectTask handles the result of resumable uploading object payload data to primary,
	// Manager should generate ReplicatePieceTask for TaskExecutor to run.
	HandleDoneResumableUploadObjectTask(ctx context.Context, task task.ResumableUploadObjectTask) error
	// HandleReplicatePieceTask handles the result of replicating piece data to secondary SPs,
	// the request comes from TaskExecutor.
	HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask) error
	// HandleSealObjectTask handles the result of sealing object to the greenfield the request comes from TaskExecutor.
	HandleSealObjectTask(ctx context.Context, task task.SealObjectTask) error
	// HandleReceivePieceTask handles the result of receiving piece task, the request comes from Receiver that
	// reports have completed ReceivePieceTask to manager and TaskExecutor that the result of confirming whether
	// the object that is synced to secondary SP has been sealed.
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) error
	// HandleGCObjectTask handles GCObjectTask, the request comes from TaskExecutor.
	HandleGCObjectTask(ctx context.Context, task task.GCObjectTask) error
	// HandleGCZombiePieceTask handles GCZombiePieceTask, the request comes from TaskExecutor.
	HandleGCZombiePieceTask(ctx context.Context, task task.GCZombiePieceTask) error
	// HandleGCMetaTask handles GCMetaTask, the request comes from TaskExecutor.
	HandleGCMetaTask(ctx context.Context, task task.GCMetaTask) error
	// HandleDownloadObjectTask handles DownloadObjectTask, the request comes from Downloader.
	HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) error
	// HandleChallengePieceTask handles ChallengePieceTask, the request comes from Downloader.
	HandleChallengePieceTask(ctx context.Context, task task.ChallengePieceTask) error
	// PickVirtualGroupFamily is used to pick vgf for the new bucket.
	PickVirtualGroupFamily(ctx context.Context, task task.ApprovalCreateBucketTask) (uint32, error)
	// HandleRecoverPieceTask handles the result of recovering piece task, the request comes from TaskExecutor.
	HandleRecoverPieceTask(ctx context.Context, task task.RecoveryPieceTask) error
}

// P2P is an abstract interface to the to do replicate piece approvals between SPs.
type P2P interface {
	Modular
	// HandleReplicatePieceApproval handles the asking replicate piece approval, it will
	// broadcast the approval to other SPs, waiting the responses. If up to min approved
	// number or max approved number before timeout, it will return the approvals.
	HandleReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask, min, max int32,
		timeout int64) ([]task.ApprovalReplicatePieceTask, error)
	// HandleQueryBootstrap handles the query p2p node bootstrap node info.
	HandleQueryBootstrap(ctx context.Context) ([]string, error)
	// QueryTasks queries replicate piece approval tasks that running on p2p by task sub-key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}

// Receiver is an abstract interface to receive the piece data from primary SP.
type Receiver interface {
	Modular
	// HandleReceivePieceTask stores piece data into secondary SP.
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask, data []byte) error
	// HandleDoneReceivePieceTask calculates the secondary bls of the object and sign it, returns to the primary
	// SP for sealed object.
	HandleDoneReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, error)
	// QueryTasks queries replicate piece tasks that running on receiver by task sub-key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}

// Signer is an abstract interface to handle the signature of SP and on greenfield chain operator.
// It holds all private keys of one SP. Considering the SP account's sequence number, it must be a singleton.
type Signer interface {
	Modular
	// SignCreateBucketApproval signs the MsgCreateBucket for asking create bucket approval.
	SignCreateBucketApproval(ctx context.Context, bucket *storagetypes.MsgCreateBucket) ([]byte, error)
	// SignCreateObjectApproval signs the MsgCreateObject for asking create object approval.
	SignCreateObjectApproval(ctx context.Context, task *storagetypes.MsgCreateObject) ([]byte, error)
	// SignReplicatePieceApproval signs the ApprovalReplicatePieceTask for asking replicate pieces to secondary SPs.
	SignReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask) ([]byte, error)
	// SignReceivePieceTask signs the ReceivePieceTask for replicating pieces data between SPs.
	SignReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, error)
	// SignSecondaryBls signs the secondary bls for sealing object.
	SignSecondaryBls(ctx context.Context, objectID uint64, gvgId uint32, hash [][]byte) ([]byte, error)
	//SignRecoveryPieceTask signs the RecoveryPieceTask for recovering piece data
	SignRecoveryPieceTask(ctx context.Context, task task.RecoveryPieceTask) ([]byte, error)
	// SignP2PPingMsg signs the ping msg for p2p node probing.
	SignP2PPingMsg(ctx context.Context, ping *gfspp2p.GfSpPing) ([]byte, error)
	// SignP2PPongMsg signs the pong msg for p2p to response ping msg.
	SignP2PPongMsg(ctx context.Context, pong *gfspp2p.GfSpPong) ([]byte, error)
	// SealObject signs the MsgSealObject and broadcast the tx to greenfield.
	SealObject(ctx context.Context, object *storagetypes.MsgSealObject) error
	// RejectUnSealObject signs the MsgRejectSealObject and broadcast the tx to greenfield.
	RejectUnSealObject(ctx context.Context, object *storagetypes.MsgRejectSealObject) error
	// DiscontinueBucket signs the MsgDiscontinueBucket and broadcast the tx to greenfield.
	DiscontinueBucket(ctx context.Context, bucket *storagetypes.MsgDiscontinueBucket) error
	// CreateGlobalVirtualGroup signs the MsgCreateGlobalVirtualGroup and broadcast the tx to greenfield.
	CreateGlobalVirtualGroup(ctx context.Context, gvg *virtualgrouptypes.MsgCreateGlobalVirtualGroup) error
}

// Uploader is an abstract interface to handle putting object requests from users' account and store
// their payload data into primary SP piece store.
type Uploader interface {
	Modular
	// PreUploadObject prepares to handle UploadObject, it can do some checks
	// such as checking for duplicates, if limitation of SP has been reached, etc.
	PreUploadObject(ctx context.Context, task task.UploadObjectTask) error
	// HandleUploadObjectTask handles the UploadObject, store payload data into piece store by data stream.
	HandleUploadObjectTask(ctx context.Context, task task.UploadObjectTask, stream io.Reader) error
	// PostUploadObject is called after HandleUploadObjectTask, it can recycle
	// resources, make statistics and do some other operations.
	PostUploadObject(ctx context.Context, task task.UploadObjectTask)

	// PreResumableUploadObject prepares to handle ResumableUploadObject, it can do some checks
	// such as checking for duplicates, if limitation of SP has been reached, etc.
	PreResumableUploadObject(ctx context.Context, task task.ResumableUploadObjectTask) error
	// HandleResumableUploadObjectTask handles the ResumableUploadObject, store payload data into piece store by data stream.
	HandleResumableUploadObjectTask(ctx context.Context, task task.ResumableUploadObjectTask, stream io.Reader) error
	// PostResumableUploadObject is called after HandleResumableUploadObjectTask, it can recycle
	// resources, statistics and other operations.
	PostResumableUploadObject(ctx context.Context, task task.ResumableUploadObjectTask)

	// QueryTasks queries upload object tasks that running on uploading by task sub-key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}
