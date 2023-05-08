package module

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ Modular = (*NullModular)(nil)
var _ Approver = (*NullModular)(nil)
var _ Uploader = (*NullModular)(nil)
var _ Manager = (*NullModular)(nil)
var _ Authorizer = (*NullModular)(nil)

type NullModular struct{}

func (*NullModular) Name() string                { return "" }
func (*NullModular) Start(context.Context) error { return nil }
func (*NullModular) Stop(context.Context) error  { return nil }
func (*NullModular) ReserveResource(context.Context, *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	return nil, nil
}
func (*NullModular) ReleaseResource(context.Context, rcmgr.ResourceScopeSpan) { return }
func (*NullModular) PreCreateBucketApproval(context.Context, task.ApprovalCreateBucketTask) error {
	return nil
}
func (*NullModular) HandleCreateBucketApprovalTask(context.Context, task.ApprovalCreateBucketTask) (bool, error) {
	return false, nil
}
func (*NullModular) PostCreateBucketApproval(context.Context, task.ApprovalCreateBucketTask) { return }
func (*NullModular) PreCreateObjectApproval(context.Context, task.ApprovalCreateObjectTask) error {
	return nil
}
func (*NullModular) HandleCreateObjectApprovalTask(context.Context, task.ApprovalCreateObjectTask) (bool, error) {
	return false, nil
}
func (*NullModular) PostCreateObjectApproval(context.Context, task.ApprovalCreateObjectTask) { return }
func (*NullModular) PreReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask) error {
	return nil
}
func (*NullModular) HandleReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask) (bool, error) {
	return false, nil
}
func (*NullModular) PostReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask) {
	return
}
func (*NullModular) PreUploadObject(ctx context.Context, task task.UploadObjectTask) error {
	return nil
}
func (*NullModular) HandleUploadObjectTask(ctx context.Context, task task.UploadObjectTask, stream io.Reader) error {
	return nil
}
func (*NullModular) PostUploadObject(ctx context.Context, task task.UploadObjectTask) { return }
func (*NullModular) DispatchTask(context.Context, rcmgr.Limit) (task.Task, error)     { return nil, nil }
func (*NullModular) QueryTask(context.Context, task.TKey) (task.Task, error)          { return nil, nil }
func (*NullModular) HandleCreateUploadObjectTask(context.Context, task.UploadObjectTask) error {
	return nil
}
func (*NullModular) HandleDoneUploadObjectTask(context.Context, task.UploadObjectTask) error {
	return nil
}

func (*NullModular) HandleReplicatePieceTask(context.Context, task.ReplicatePieceTask) error {
	return nil
}
func (*NullModular) HandleSealObjectTask(context.Context, task.SealObjectTask) error     { return nil }
func (*NullModular) HandleReceivePieceTask(context.Context, task.ReceivePieceTask) error { return nil }
func (*NullModular) HandleGCObjectTask(context.Context, task.GCObjectTask) error         { return nil }
func (*NullModular) HandleGCZombiePieceTask(context.Context, task.GCZombiePieceTask) error {
	return nil
}
func (*NullModular) HandleGCMetaTask(context.Context, task.GCMetaTask) error { return nil }
func (*NullModular) HandleDownloadObjectTask(context.Context, task.DownloadObjectTask) error {
	return nil
}
func (*NullModular) HandleChallengePieceTask(context.Context, task.ChallengePieceTask) error {
	return nil
}
func (*NullModular) VerifyAuthorize(context.Context, AuthOpType, string, string, string) (bool, error) {
	return false, nil
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
	return nil, nil
}
func (*NilModular) ReleaseResource(context.Context, rcmgr.ResourceScopeSpan)         { return }
func (*NilModular) PreDownloadObject(context.Context, task.DownloadObjectTask) error { return nil }
func (*NilModular) HandleDownloadObjectTask(context.Context, task.DownloadObjectTask) ([]byte, error) {
	return nil, nil
}
func (*NilModular) PostDownloadObject(context.Context, task.DownloadObjectTask)      {}
func (*NilModular) PreChallengePiece(context.Context, task.ChallengePieceTask) error { return nil }
func (*NilModular) HandleChallengePiece(context.Context, task.ChallengePieceTask) ([]byte, [][]byte, []byte, error) {
	return nil, nil, nil, nil
}
func (*NilModular) AskTask(context.Context, rcmgr.Limit)                              {}
func (*NilModular) PostChallengePiece(context.Context, task.ChallengePieceTask)       {}
func (*NilModular) ReportTask(context.Context, task.Task) error                       { return nil }
func (*NilModular) HandleReplicatePieceTask(context.Context, task.ReplicatePieceTask) {}
func (*NilModular) HandleSealObjectTask(context.Context, task.SealObjectTask)         {}
func (*NilModular) HandleReceivePieceTask(context.Context, task.ReceivePieceTask)     {}
func (*NilModular) HandleGCObjectTask(context.Context, task.GCObjectTask)             {}
func (*NilModular) HandleGCZombiePieceTask(context.Context, task.GCZombiePieceTask)   {}
func (*NilModular) HandleGCMetaTask(context.Context, task.GCMetaTask)                 {}
func (*NilModular) HandleReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask, int32, int32, int64) ([]task.ApprovalReplicatePieceTask, error) {
	return nil, nil
}
func (*NilModular) HandleQueryBootstrap(context.Context) ([]string, error) { return nil, nil }

func (*NilModular) SignCreateBucketApproval(context.Context, *storagetypes.MsgCreateBucket) ([]byte, error) {
	return nil, nil
}
func (*NilModular) SignCreateObjectApproval(context.Context, *storagetypes.MsgCreateObject) ([]byte, error) {
	return nil, nil
}
func (*NilModular) SignReplicatePieceApproval(context.Context, task.ApprovalReplicatePieceTask) ([]byte, error) {
	return nil, nil
}
func (*NilModular) SignReceivePieceTask(context.Context, task.ReceivePieceTask) ([]byte, error) {
	return nil, nil
}
func (*NilModular) SignIntegrityHash(ctx context.Context, objectID uint64, hash [][]byte) ([]byte, []byte, error) {
	return nil, nil, nil
}
func (*NilModular) SignP2PPingMsg(context.Context, *gfspp2p.GfSpPing) ([]byte, error) {
	return nil, nil
}
func (*NilModular) SignP2PPongMsg(context.Context, *gfspp2p.GfSpPong) ([]byte, error) {
	return nil, nil
}
func (*NilModular) SealObject(context.Context, *storagetypes.MsgSealObject) error { return nil }

var _ Receiver = (*NullReceiveModular)(nil)

type NullReceiveModular struct{}

func (*NullReceiveModular) Name() string                { return "" }
func (*NullReceiveModular) Start(context.Context) error { return nil }
func (*NullReceiveModular) Stop(context.Context) error  { return nil }
func (*NullReceiveModular) ReserveResource(context.Context, *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	return nil, nil
}
func (*NullReceiveModular) ReleaseResource(context.Context, rcmgr.ResourceScopeSpan) { return }
func (*NullReceiveModular) HandleReceivePieceTask(context.Context, task.ReceivePieceTask, []byte) error {
	return nil
}
func (*NullReceiveModular) HandleDoneReceivePieceTask(context.Context, task.ReceivePieceTask) ([]byte, []byte, error) {
	return nil, nil, nil
}
