package module

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type Modular interface {
	Name() string
	Description() string
	Endpoint() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error)
	ReleaseResource(ctx context.Context, scope rcmgr.ResourceScopeSpan)
}

type Approver interface {
	Modular
	PreCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask) error
	HandleCreateBucketApprovalTask(ctx context.Context, task task.ApprovalCreateBucketTask) (bool, error)
	PostCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask)

	PreCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask) error
	HandleCreateObjectApprovalTask(ctx context.Context, task task.ApprovalCreateObjectTask) (bool, error)
	PostCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask)
}

type Downloader interface {
	Modular
	PreDownloadObject(ctx context.Context, task task.DownloadObjectTask) error
	HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) ([]byte, error)
	PostDownloadObject(ctx context.Context, task task.DownloadObjectTask)

	PreChallengePiece(ctx context.Context, task task.ChallengePieceTask) error
	HandleChallengePiece(ctx context.Context, task task.ChallengePieceTask) ([]byte, [][]byte, []byte, error)
	PostChallengePiece(ctx context.Context, task task.ChallengePieceTask)
}

type Uploader interface {
	Modular
	PreUploadObject(ctx context.Context, task task.UploadObjectTask) error
	HandleUploadObjectTask(ctx context.Context, task task.UploadObjectTask, stream io.Reader) error
	PostUploadObject(ctx context.Context, task task.UploadObjectTask)
}

type Manager interface {
	Modular
	DispatchTask(ctx context.Context, limit rcmgr.Limit) (task.Task, error)
	QueryTask(ctx context.Context, key task.TKey) (task.Task, error)
	HandleCreateUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error
	HandleDoneUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error
	HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask) error
	HandleSealObjectTask(ctx context.Context, task task.SealObjectTask) error
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) error
	HandleGCObjectTask(ctx context.Context, task task.GCObjectTask) error
	HandleGCZombiePieceTask(ctx context.Context, task task.GCZombiePieceTask) error
	HandleGCMetaTask(ctx context.Context, task task.GCMetaTask) error
	HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) error
	HandleChallengePieceTask(ctx context.Context, task task.ChallengePieceTask) error
}

type TaskExecutor interface {
	Modular
	ReportTask(ctx context.Context, task task.Task) error
	HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask)
	HandleSealObjectTask(ctx context.Context, task task.SealObjectTask)
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask)
	HandleGCObjectTask(ctx context.Context, task task.GCObjectTask)
	HandleGCZombiePieceTask(ctx context.Context, task task.GCZombiePieceTask)
	HandleGCMetaTask(ctx context.Context, task task.GCMetaTask)
}

type Receiver interface {
	Modular
	HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask, data []byte) error
	HandleDoneReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, []byte, error)
}

type P2P interface {
	Modular
	HandleReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask,
		min, max int32, timeout int64) ([]task.ApprovalReplicatePieceTask, error)
	HandleQueryBootstrap(ctx context.Context) ([]string, error)
}

type Signer interface {
	SignCreateBucketApproval(ctx context.Context, bucket *storagetypes.MsgCreateBucket) ([]byte, error)
	SignCreateObjectApproval(ctx context.Context, task *storagetypes.MsgCreateObject) ([]byte, error)
	SignReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask) ([]byte, error)
	SignReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, error)
	SignIntegrityHash(ctx context.Context, objectID uint64, hash [][]byte) ([]byte, []byte, error)
	SignP2PPingMsg(ctx context.Context, ping *gfspp2p.GfSpPing) ([]byte, error)
	SignP2PPongMsg(ctx context.Context, pong *gfspp2p.GfSpPong) ([]byte, error)
	SealObject(ctx context.Context, object *storagetypes.MsgSealObject) error
}

type AuthOpType int32

const (
	AuthOpTypeUnKnown AuthOpType = iota
	AuthOpAskCreateBucketApproval
	AuthOpAskCreateObjectApproval
	AuthOpTypeChallengePiece
	AuthOpTypePutObject
	AuthOpTypeGetObject
	AuthOpTypeGetUploadingState
	AuthOpTypeGetBucketQuota
	AuthOpTypeListBucketReadRecord
)

type Authorizer interface {
	Modular
	VerifyAuthorize(ctx context.Context, auth AuthOpType, account, bucket, object string) (bool, error)
}
