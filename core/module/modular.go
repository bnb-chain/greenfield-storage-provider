package module

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/core/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// Modular is the interface to submodule that units scheduled by the GfSp framework.
// the Modular inherits lifecycle.Service interface, used to managed by lifecycle.
// and it also is managed by ResourceManager, the GfSp framework will reserve and
// release resources from Modular resources pool.
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
)

// Authenticator is the interface to authentication verification modular.
type Authenticator interface {
	Modular
	// VerifyAuthentication verifies the operator authentication.
	VerifyAuthentication(ctx context.Context, auth AuthOpType, account, bucket, object string) (bool, error)
	// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
	GetAuthNonce(ctx context.Context, account string, domain string) (*spdb.OffChainAuthKey, error)
	// UpdateUserPublicKey updates the user public key once the Dapp or client generates the EDDSA key pairs.
	UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32, userPublicKey string, expiryDate int64) (bool, error)
	// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
	VerifyOffChainSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign string) (bool, error)
}

// Approver is the interface to handle ask approval.
type Approver interface {
	Modular
	// PreCreateBucketApproval prepares to handle CreateBucketApproval, it can do some
	// checks Example: check for duplicates, if limit specified by SP is reached, etc.
	PreCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask) error
	// HandleCreateBucketApprovalTask handles the CreateBucketApproval, set expired
	// height and sign the MsgCreateBucket etc.
	HandleCreateBucketApprovalTask(ctx context.Context, task task.ApprovalCreateBucketTask) (bool, error)
	// PostCreateBucketApproval is called after HandleCreateBucketApprovalTask, it can
	// recycle resources, statistics and other operations.
	PostCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask)

	// PreCreateObjectApproval prepares to handle CreateObjectApproval, it can do some
	// checks Example: check for duplicates, if limit specified by SP is reached, etc.
	PreCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask) error
	// HandleCreateObjectApprovalTask handles the MsgCreateObject, set expired height
	// and sign the MsgCreateBucket etc.
	HandleCreateObjectApprovalTask(ctx context.Context, task task.ApprovalCreateObjectTask) (bool, error)
	// PostCreateObjectApproval is called after HandleCreateObjectApprovalTask, it can
	// recycle resources, statistics and other operations.
	PostCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask)
	// QueryTasks queries tasks that running on approver by task sub key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}

// Downloader is the interface to handle get object request from user account, and get
// challenge info request from other components in the system.
type Downloader interface {
	Modular
	// PreDownloadObject prepares to handle DownloadObject, it can do some checks
	// Example: check for duplicates, if limit specified by SP is reached, etc.
	PreDownloadObject(ctx context.Context, task task.DownloadObjectTask) error
	// HandleDownloadObjectTask handles the DownloadObject, get data from piece store.
	HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) ([]byte, error)
	// PostDownloadObject is called after HandleDownloadObjectTask, it can recycle
	// resources, statistics and other operations.
	PostDownloadObject(ctx context.Context, task task.DownloadObjectTask)

	// PreDownloadPiece prepares to handle DownloadPiece, it can do some checks
	// Example: check for duplicates, if limit specified by SP is reached, etc.
	PreDownloadPiece(ctx context.Context, task task.DownloadPieceTask) error
	// HandleDownloadPieceTask handles the DownloadPiece, get data from piece store.
	HandleDownloadPieceTask(ctx context.Context, task task.DownloadPieceTask) ([]byte, error)
	// PostDownloadPiece is called after HandleDownloadPieceTask, it can recycle
	// resources, statistics and other operations.
	PostDownloadPiece(ctx context.Context, task task.DownloadPieceTask)

	// PreChallengePiece prepares to handle ChallengePiece, it can do some checks
	// Example: check for duplicates, if limit specified by SP is reached, etc.
	PreChallengePiece(ctx context.Context, task task.ChallengePieceTask) error
	// HandleChallengePiece handles the ChallengePiece, get piece data from piece
	// store and get integrity hash from db.
	HandleChallengePiece(ctx context.Context, task task.ChallengePieceTask) ([]byte, [][]byte, []byte, error)
	// PostChallengePiece is called after HandleChallengePiece, it can recycle
	// resources, statistics and other operations.
	PostChallengePiece(ctx context.Context, task task.ChallengePieceTask)
	// QueryTasks queries download/challenge tasks that running on downloader by
	// task sub key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}

// TaskExecutor is the interface to handle background task, it will ask task from
// manager modular, handle the task and report the result or status to the manager
// modular includes: ReplicatePieceTask, SealObjectTask, ReceivePieceTask, GCObjectTask
// GCZombiePieceTask, GCMetaTask.
type TaskExecutor interface {
	Modular
	// AskTask asks the task by remaining limit from manager modular.
	AskTask(ctx context.Context) error
	// HandleReplicatePieceTask handles the ReplicatePieceTask that is asked from
	// manager modular.
	HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask)
	// HandleSealObjectTask handles the SealObjectTask that is asked from manager
	// modular.
	HandleSealObjectTask(ctx context.Context, task task.SealObjectTask)
	// HandleReceivePieceTask handles the ReceivePieceTask that is asked from manager
	// modular. It will confirm the object that as secondary SP whether has been sealed.
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask)
	// HandleGCObjectTask handles the GCObjectTask that is asked from manager modular.
	HandleGCObjectTask(ctx context.Context, task task.GCObjectTask)
	// HandleGCZombiePieceTask handles the GCZombiePieceTask that is asked from manager
	// modular.
	HandleGCZombiePieceTask(ctx context.Context, task task.GCZombiePieceTask)
	// HandleGCMetaTask handles the GCMetaTask that is asked from manager modular.
	HandleGCMetaTask(ctx context.Context, task task.GCMetaTask)
	// ReportTask reports the result or status of running task to manager modular.
	ReportTask(ctx context.Context, task task.Task) error
}

// Manager is the interface to SP's manage modular, it is Responsible for task
// scheduling and other management of SP.
type Manager interface {
	Modular
	// DispatchTask dispatches the task to TaskExecutor modular when it asks task.
	// It will consider task remaining resources when dispatches task.
	DispatchTask(ctx context.Context, limit rcmgr.Limit) (task.Task, error)
	// QueryTasks queries tasks that hold on manager by task sub key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
	// HandleCreateUploadObjectTask handles the CreateUploadObject request from
	// Uploader, before Uploader handles the user's UploadObject request, it should
	// send CreateUploadObject request to Manager ask if it's ok. Through this
	// interface that SP implements the global upload object strategy.
	//
	// Example: control the concurrency of global uploads, avoid repeated uploads,
	// rate control, etc.
	HandleCreateUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error
	// HandleDoneUploadObjectTask handles the result of uploading object payload
	// data to primary, Manager should generate ReplicatePieceTask for TaskExecutor
	// to run.
	HandleDoneUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error
	// HandleReplicatePieceTask handles the result of replicating pieces data to
	// secondary SPs, the request comes from TaskExecutor.
	HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask) error
	// HandleSealObjectTask handles the result of sealing object to the greenfield
	// the request comes from TaskExecutor.
	HandleSealObjectTask(ctx context.Context, task task.SealObjectTask) error
	// HandleReceivePieceTask handles the result of receiving piece task, the request
	// comes from Receiver that reports have completed the receive task to manager and
	// TaskExecutor that the result of confirming whether the object as secondary SP
	// has been sealed.
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) error
	// HandleGCObjectTask handles the result or status of GCObjectTask, the request
	// comes from TaskExecutor.
	HandleGCObjectTask(ctx context.Context, task task.GCObjectTask) error
	// HandleGCZombiePieceTask handles the result or status GCZombiePieceTask, the
	// request comes from TaskExecutor.
	HandleGCZombiePieceTask(ctx context.Context, task task.GCZombiePieceTask) error
	// HandleGCMetaTask handles the result or status GCMetaTask, the request comes
	// from TaskExecutor.
	HandleGCMetaTask(ctx context.Context, task task.GCMetaTask) error
	// HandleDownloadObjectTask handles the result DownloadObjectTask, the request comes
	// from Downloader.
	HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) error
	// HandleChallengePieceTask handles the result ChallengePieceTask, the request comes
	// from Downloader.
	HandleChallengePieceTask(ctx context.Context, task task.ChallengePieceTask) error
	// PickVirtualGroupFamily is used to pick vgf for the new bucket.
	PickVirtualGroupFamily(ctx context.Context, task task.ApprovalCreateBucketTask) (uint32, error)
}

// P2P is the interface to the interaction of control information between Sps.
type P2P interface {
	Modular
	// HandleReplicatePieceApproval handles the ask replicate piece approval, it will
	// broadcast the approval to other SPs, wait the responses, if up to min approved
	// number or max approved number before timeout, will return the approvals.
	HandleReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask,
		min, max int32, timeout int64) ([]task.ApprovalReplicatePieceTask, error)
	// HandleQueryBootstrap handles the query p2p node bootstrap node info.
	HandleQueryBootstrap(ctx context.Context) ([]string, error)
	// QueryTasks queries replicate piece approval tasks that running on p2p by task
	// sub key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}

// Receiver is the interface to receive the piece data from primary SP.
type Receiver interface {
	Modular
	// HandleReceivePieceTask stores the piece data from primary SP.
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask, data []byte) error
	// HandleDoneReceivePieceTask calculates the integrity hash of the object and sign
	// it, returns to the primary SP for seal object.
	HandleDoneReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, []byte, error)
	// QueryTasks queries replicate piece tasks that running on receiver by task sub
	// key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}

// Signer is the interface to handle the SP's sign and on greenfield chain operator.
// It holds SP all private key. Considering the sp account's sequence number, it must
// be a singleton.
type Signer interface {
	Modular
	// SignCreateBucketApproval signs the MsgCreateBucket for asking create bucket
	// approval.
	SignCreateBucketApproval(ctx context.Context, bucket *storagetypes.MsgCreateBucket) ([]byte, error)
	// SignCreateObjectApproval signs the MsgCreateObject for asking create object
	// approval.
	SignCreateObjectApproval(ctx context.Context, task *storagetypes.MsgCreateObject) ([]byte, error)
	// SignReplicatePieceApproval signs the ApprovalReplicatePieceTask for asking
	// replicate pieces to secondary SPs.
	SignReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask) ([]byte, error)
	// SignReceivePieceTask signs the ReceivePieceTask for replicating pieces data
	// between SPs.
	SignReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, error)
	// SignIntegrityHash signs the integrity hash of object for sealing object.
	SignIntegrityHash(ctx context.Context, objectID uint64, hash [][]byte) ([]byte, []byte, error)
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
}

// Uploader is the interface to handle put object request from user account, and
// store it in primary SP's piece store.
type Uploader interface {
	Modular
	// PreUploadObject prepares to handle UploadObject, it can do some checks
	// Example: check for duplicates, if limit specified by SP is reached, etc.
	PreUploadObject(ctx context.Context, task task.UploadObjectTask) error
	// HandleUploadObjectTask handles the UploadObject, store the payload data
	// to piece store by data stream.
	HandleUploadObjectTask(ctx context.Context, task task.UploadObjectTask, stream io.Reader) error
	// PostUploadObject is called after HandleUploadObjectTask, it can recycle
	// resources, statistics and other operations.
	PostUploadObject(ctx context.Context, task task.UploadObjectTask)
	// QueryTasks queries upload object tasks that running on uploading by task
	// sub key.
	QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}
