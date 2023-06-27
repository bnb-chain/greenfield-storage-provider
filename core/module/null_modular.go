package module

import (
	"context"
	"errors"
	"io"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	ErrNilModular = errors.New("call nil module, please check again")
)

var _ Modular = (*NullModular)(nil)
var _ Approver = (*NullModular)(nil)
var _ Uploader = (*NullModular)(nil)
var _ Manager = (*NullModular)(nil)
var _ Authenticator = (*NullModular)(nil)

type NullModular struct{}

func (*NullModular) Name() string                { return "" }
func (*NullModular) Start(context.Context) error { return nil }
func (*NullModular) Stop(context.Context) error  { return nil }
func (*NullModular) ReserveResource(context.Context, *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}
func (*NullModular) ReleaseResource(context.Context, rcmgr.ResourceScopeSpan) {}
func (*NullModular) QueryTasks(ctx context.Context, keyPrefix task.TKey) ([]task.Task, error) {
	return nil, ErrNilModular
}
func (*NullModular) PreCreateBucketApproval(context.Context, task.ApprovalCreateBucketTask) error {
	return ErrNilModular
}
func (*NullModular) HandleCreateBucketApprovalTask(context.Context, task.ApprovalCreateBucketTask) (bool, error) {
	return false, ErrNilModular
}

func (*NullModular) PickVirtualGroupFamily(context.Context, task.ApprovalCreateBucketTask) (uint32, error) {
	return 0, ErrNilModular
}

func (*NullModular) PostCreateBucketApproval(context.Context, task.ApprovalCreateBucketTask) {}
func (*NullModular) PreCreateObjectApproval(context.Context, task.ApprovalCreateObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleCreateObjectApprovalTask(context.Context, task.ApprovalCreateObjectTask) (bool, error) {
	return false, ErrNilModular
}
func (*NullModular) PostCreateObjectApproval(context.Context, task.ApprovalCreateObjectTask) {}
func (*NullModular) PreReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask) error {
	return ErrNilModular
}
func (*NullModular) HandleReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask) (bool, error) {
	return false, ErrNilModular
}
func (*NullModular) PostReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask) {}
func (*NullModular) PreUploadObject(ctx context.Context, task task.UploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleUploadObjectTask(ctx context.Context, task task.UploadObjectTask, stream io.Reader) error {
	return nil
}
func (*NullModular) PostUploadObject(ctx context.Context, task task.UploadObjectTask) {}
func (*NullModular) DispatchTask(context.Context, rcmgr.Limit) (task.Task, error) {
	return nil, ErrNilModular
}
func (*NullModular) QueryTask(context.Context, task.TKey) (task.Task, error) {
	return nil, ErrNilModular
}
func (*NullModular) HandleCreateUploadObjectTask(context.Context, task.UploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleDoneUploadObjectTask(context.Context, task.UploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleReplicatePieceTask(context.Context, task.ReplicatePieceTask) error {
	return ErrNilModular
}
func (*NullModular) HandleSealObjectTask(context.Context, task.SealObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleReceivePieceTask(context.Context, task.ReceivePieceTask) error {
	return ErrNilModular
}
func (*NullModular) HandleGCObjectTask(context.Context, task.GCObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleGCZombiePieceTask(context.Context, task.GCZombiePieceTask) error {
	return ErrNilModular
}
func (*NullModular) HandleGCMetaTask(context.Context, task.GCMetaTask) error { return ErrNilModular }
func (*NullModular) HandleDownloadObjectTask(context.Context, task.DownloadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleChallengePieceTask(context.Context, task.ChallengePieceTask) error {
	return ErrNilModular
}

func (*NullModular) VerifyAuthentication(context.Context, AuthOpType, string, string, string) (bool, error) {
	return false, ErrNilModular
}

func (*NullModular) GetAuthNonce(ctx context.Context, account string, domain string) (*corespdb.OffChainAuthKey, error) {
	return nil, ErrNilModular
}

func (*NullModular) UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32, userPublicKey string, expiryDate int64) (bool, error) {
	return false, ErrNilModular
}

func (*NullModular) VerifyOffChainSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign string) (bool, error) {
	return false, ErrNilModular
}

func (*NullModular) HandleMigratePieceTask(ctx context.Context, task task.MigratePieceTask) error {
	return ErrNilModular
}

var _ TaskExecutor = (*NilModular)(nil)
var _ P2P = (*NilModular)(nil)
var _ Signer = (*NilModular)(nil)
var _ Downloader = (*NilModular)(nil)

type NilModular struct{}

func (*NilModular) Name() string                { return "" }
func (*NilModular) Start(context.Context) error { return nil }
func (*NilModular) Stop(context.Context) error  { return nil }
func (*NilModular) ReserveResource(context.Context, *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}
func (*NilModular) ReleaseResource(context.Context, rcmgr.ResourceScopeSpan) {}
func (*NilModular) QueryTasks(ctx context.Context, keyPrefix task.TKey) ([]task.Task, error) {
	return nil, ErrNilModular
}
func (*NilModular) PreDownloadObject(context.Context, task.DownloadObjectTask) error {
	return ErrNilModular
}
func (*NilModular) HandleDownloadObjectTask(context.Context, task.DownloadObjectTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) PostDownloadObject(context.Context, task.DownloadObjectTask) {}

func (*NilModular) PreDownloadPiece(context.Context, task.DownloadPieceTask) error {
	return ErrNilModular
}
func (*NilModular) HandleDownloadPieceTask(context.Context, task.DownloadPieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) PostDownloadPiece(context.Context, task.DownloadPieceTask) {}

func (*NilModular) PreChallengePiece(context.Context, task.ChallengePieceTask) error {
	return ErrNilModular
}
func (*NilModular) HandleChallengePiece(context.Context, task.ChallengePieceTask) ([]byte, [][]byte, []byte, error) {
	return nil, nil, nil, ErrNilModular
}
func (*NilModular) AskTask(context.Context) error                                     { return nil }
func (*NilModular) PostChallengePiece(context.Context, task.ChallengePieceTask)       {}
func (*NilModular) ReportTask(context.Context, task.Task) error                       { return ErrNilModular }
func (*NilModular) HandleReplicatePieceTask(context.Context, task.ReplicatePieceTask) {}
func (*NilModular) HandleSealObjectTask(context.Context, task.SealObjectTask)         {}
func (*NilModular) HandleReceivePieceTask(context.Context, task.ReceivePieceTask)     {}
func (*NilModular) HandleGCObjectTask(context.Context, task.GCObjectTask)             {}
func (*NilModular) HandleGCZombiePieceTask(context.Context, task.GCZombiePieceTask)   {}
func (*NilModular) HandleGCMetaTask(context.Context, task.GCMetaTask)                 {}
func (*NilModular) HandleReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask, int32, int32, int64) ([]task.ApprovalReplicatePieceTask, error) {
	return nil, ErrNilModular
}
func (*NilModular) HandleQueryBootstrap(context.Context) ([]string, error) { return nil, ErrNilModular }

func (*NilModular) SignCreateBucketApproval(context.Context, *storagetypes.MsgCreateBucket) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignCreateObjectApproval(context.Context, *storagetypes.MsgCreateObject) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignReceivePieceTask(context.Context, task.ReceivePieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignSecondaryBls(context.Context, uint64, uint32, [][]byte) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignP2PPingMsg(context.Context, *gfspp2p.GfSpPing) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignP2PPongMsg(context.Context, *gfspp2p.GfSpPong) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SealObject(context.Context, *storagetypes.MsgSealObject) error {
	return ErrNilModular
}
func (*NilModular) RejectUnSealObject(context.Context, *storagetypes.MsgRejectSealObject) error {
	return ErrNilModular
}
func (*NilModular) DiscontinueBucket(context.Context, *storagetypes.MsgDiscontinueBucket) error {
	return nil
}

func (*NilModular) CreateGlobalVirtualGroup(context.Context, *virtualgrouptypes.MsgCreateGlobalVirtualGroup) error {
	return nil
}

func (NilModular) SignMigratePieceTask(ctx context.Context, task task.MigratePieceTask) ([]byte, error) {
	return nil, ErrNilModular
}

var _ Receiver = (*NullReceiveModular)(nil)

type NullReceiveModular struct{}

func (*NullReceiveModular) Name() string                { return "" }
func (*NullReceiveModular) Start(context.Context) error { return nil }
func (*NullReceiveModular) Stop(context.Context) error  { return nil }
func (*NullReceiveModular) ReserveResource(context.Context, *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}
func (*NullReceiveModular) ReleaseResource(context.Context, rcmgr.ResourceScopeSpan) {}
func (*NullReceiveModular) QueryTasks(ctx context.Context, keyPrefix task.TKey) ([]task.Task, error) {
	return nil, ErrNilModular
}
func (*NullReceiveModular) HandleReceivePieceTask(context.Context, task.ReceivePieceTask, []byte) error {
	return ErrNilModular
}
func (*NullReceiveModular) HandleDoneReceivePieceTask(context.Context, task.ReceivePieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
