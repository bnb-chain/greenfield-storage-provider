package module

import (
	"context"
	"errors"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
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
func (*NullModular) ReserveResource(context.Context, *corercmgr.ScopeStat) (corercmgr.ResourceScopeSpan, error) {
	return &corercmgr.NullScope{}, nil
}
func (*NullModular) ReleaseResource(context.Context, corercmgr.ResourceScopeSpan) {}
func (*NullModular) QueryTasks(ctx context.Context, keyPrefix coretask.TKey) ([]coretask.Task, error) {
	return nil, ErrNilModular
}
func (m *NullModular) QueryBucketMigrate(ctx context.Context) (*gfspserver.GfSpQueryBucketMigrateResponse, error) {
	return nil, ErrNilModular
}
func (m *NullModular) QuerySpExit(ctx context.Context) (*gfspserver.GfSpQuerySpExitResponse, error) {
	return nil, ErrNilModular
}
func (*NullModular) PreCreateBucketApproval(context.Context, coretask.ApprovalCreateBucketTask) error {
	return ErrNilModular
}
func (*NullModular) HandleCreateBucketApprovalTask(context.Context, coretask.ApprovalCreateBucketTask) (bool, error) {
	return false, ErrNilModular
}
func (*NullModular) PostCreateBucketApproval(context.Context, coretask.ApprovalCreateBucketTask) {}
func (*NullModular) PreMigrateBucketApproval(context.Context, coretask.ApprovalMigrateBucketTask) error {
	return ErrNilModular
}
func (*NullModular) HandleMigrateBucketApprovalTask(context.Context, coretask.ApprovalMigrateBucketTask) (bool, error) {
	return false, ErrNilModular
}
func (*NullModular) PostMigrateBucketApproval(context.Context, coretask.ApprovalMigrateBucketTask) {}
func (*NullModular) PickVirtualGroupFamily(context.Context, coretask.ApprovalCreateBucketTask) (uint32, error) {
	return 0, ErrNilModular
}
func (*NullModular) NotifyMigrateSwapOut(context.Context, *virtualgrouptypes.MsgSwapOut) error {
	return ErrNilModular
}
func (*NullModular) PreCreateObjectApproval(context.Context, coretask.ApprovalCreateObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleCreateObjectApprovalTask(context.Context, coretask.ApprovalCreateObjectTask) (bool, error) {
	return false, ErrNilModular
}
func (*NullModular) PostCreateObjectApproval(context.Context, coretask.ApprovalCreateObjectTask) {}
func (*NullModular) PreReplicatePieceApproval(context.Context, coretask.ApprovalReplicatePieceTask) error {
	return ErrNilModular
}
func (*NullModular) HandleReplicatePieceApproval(context.Context, coretask.ApprovalReplicatePieceTask) (bool, error) {
	return false, ErrNilModular
}
func (*NullModular) HandleRecoverPieceTask(ctx context.Context, task coretask.RecoveryPieceTask) error {
	return ErrNilModular
}
func (*NullModular) PostReplicatePieceApproval(context.Context, coretask.ApprovalReplicatePieceTask) {
}
func (*NullModular) PreUploadObject(ctx context.Context, task coretask.UploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) PreResumableUploadObject(ctx context.Context, task coretask.ResumableUploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleResumableUploadObjectTask(ctx context.Context, task coretask.ResumableUploadObjectTask, stream io.Reader) error {
	return ErrNilModular
}
func (*NullModular) PostResumableUploadObject(ctx context.Context, task coretask.ResumableUploadObjectTask) {
}
func (*NullModular) HandleUploadObjectTask(ctx context.Context, task coretask.UploadObjectTask, stream io.Reader) error {
	return nil
}
func (*NullModular) PostUploadObject(ctx context.Context, task coretask.UploadObjectTask) {}
func (*NullModular) DispatchTask(context.Context, corercmgr.Limit) (coretask.Task, error) {
	return nil, ErrNilModular
}
func (*NullModular) QueryTask(context.Context, coretask.TKey) (coretask.Task, error) {
	return nil, ErrNilModular
}
func (*NullModular) HandleCreateUploadObjectTask(context.Context, coretask.UploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleDoneUploadObjectTask(context.Context, coretask.UploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleCreateResumableUploadObjectTask(context.Context, coretask.ResumableUploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleDoneResumableUploadObjectTask(context.Context, coretask.ResumableUploadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleReplicatePieceTask(context.Context, coretask.ReplicatePieceTask) error {
	return ErrNilModular
}
func (*NullModular) HandleSealObjectTask(context.Context, coretask.SealObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleReceivePieceTask(context.Context, coretask.ReceivePieceTask) error {
	return ErrNilModular
}
func (*NullModular) HandleGCObjectTask(context.Context, coretask.GCObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleGCZombiePieceTask(context.Context, coretask.GCZombiePieceTask) error {
	return ErrNilModular
}
func (*NullModular) HandleGCMetaTask(context.Context, coretask.GCMetaTask) error {
	return ErrNilModular
}
func (*NullModular) HandleMigrateGVGTask(ctx context.Context, gvgTask coretask.MigrateGVGTask) error {
	return ErrNilModular
}
func (*NullModular) HandleDownloadObjectTask(context.Context, coretask.DownloadObjectTask) error {
	return ErrNilModular
}
func (*NullModular) HandleChallengePieceTask(context.Context, coretask.ChallengePieceTask) error {
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
func (*NullModular) VerifyGNFD1EddsaSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign []byte) (bool, error) {
	return false, ErrNilModular
}
func (*NullModular) NotifyBucketMigrationDone(ctx context.Context, bucketID uint64) error {
	return nil
}
func (*NullModular) HandleGCBucketMigrationTask(ctx context.Context, task coretask.GCBucketMigrationTask) error {
	return nil
}

var _ TaskExecutor = (*NilModular)(nil)
var _ P2P = (*NilModular)(nil)
var _ Signer = (*NilModular)(nil)
var _ Downloader = (*NilModular)(nil)

type NilModular struct{}

func (*NilModular) Name() string                { return "" }
func (*NilModular) Start(context.Context) error { return nil }
func (*NilModular) Stop(context.Context) error  { return nil }
func (*NilModular) ReserveResource(context.Context, *corercmgr.ScopeStat) (corercmgr.ResourceScopeSpan, error) {
	return &corercmgr.NullScope{}, nil
}
func (*NilModular) ReleaseResource(context.Context, corercmgr.ResourceScopeSpan) {}
func (*NilModular) QueryTasks(ctx context.Context, keyPrefix coretask.TKey) ([]coretask.Task, error) {
	return nil, ErrNilModular
}
func (*NilModular) PreDownloadObject(context.Context, coretask.DownloadObjectTask) error {
	return ErrNilModular
}
func (*NilModular) HandleDownloadObjectTask(context.Context, coretask.DownloadObjectTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) PostDownloadObject(context.Context, coretask.DownloadObjectTask) {}

func (*NilModular) PreDownloadPiece(context.Context, coretask.DownloadPieceTask) error {
	return ErrNilModular
}
func (*NilModular) HandleDownloadPieceTask(context.Context, coretask.DownloadPieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) PostDownloadPiece(context.Context, coretask.DownloadPieceTask) {}

func (*NilModular) PreChallengePiece(context.Context, coretask.ChallengePieceTask) error {
	return ErrNilModular
}
func (*NilModular) HandleChallengePiece(context.Context, coretask.ChallengePieceTask) ([]byte, [][]byte, []byte, error) {
	return nil, nil, nil, ErrNilModular
}
func (*NilModular) AskTask(context.Context) error                                         { return nil }
func (*NilModular) PostChallengePiece(context.Context, coretask.ChallengePieceTask)       {}
func (*NilModular) ReportTask(context.Context, coretask.Task) error                       { return ErrNilModular }
func (*NilModular) HandleReplicatePieceTask(context.Context, coretask.ReplicatePieceTask) {}
func (*NilModular) HandleSealObjectTask(context.Context, coretask.SealObjectTask)         {}
func (*NilModular) HandleReceivePieceTask(context.Context, coretask.ReceivePieceTask)     {}
func (*NilModular) HandleGCObjectTask(context.Context, coretask.GCObjectTask)             {}
func (*NilModular) HandleGCZombiePieceTask(context.Context, coretask.GCZombiePieceTask)   {}
func (*NilModular) HandleGCMetaTask(context.Context, coretask.GCMetaTask)                 {}
func (*NilModular) HandleReplicatePieceApproval(context.Context, coretask.ApprovalReplicatePieceTask, int32, int32, int64) ([]coretask.ApprovalReplicatePieceTask, error) {
	return nil, ErrNilModular
}
func (*NilModular) HandleMigrateGVGTask(ctx context.Context, gvgTask coretask.MigrateGVGTask) {}
func (*NilModular) HandleQueryBootstrap(context.Context) ([]string, error)                    { return nil, ErrNilModular }
func (*NilModular) SignCreateBucketApproval(context.Context, *storagetypes.MsgCreateBucket) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignMigrateBucketApproval(context.Context, *storagetypes.MsgMigrateBucket) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignCreateObjectApproval(context.Context, *storagetypes.MsgCreateObject) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignReplicatePieceApproval(context.Context, coretask.ApprovalReplicatePieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignReceivePieceTask(context.Context, coretask.ReceivePieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignSecondarySealBls(context.Context, uint64, uint32, [][]byte) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignRecoveryPieceTask(context.Context, coretask.RecoveryPieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignP2PPingMsg(context.Context, *gfspp2p.GfSpPing) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SignP2PPongMsg(context.Context, *gfspp2p.GfSpPong) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SealObject(context.Context, *storagetypes.MsgSealObject) (string, error) {
	return "", ErrNilModular
}
func (*NilModular) RejectUnSealObject(context.Context, *storagetypes.MsgRejectSealObject) (string, error) {
	return "", ErrNilModular
}
func (*NilModular) DiscontinueBucket(context.Context, *storagetypes.MsgDiscontinueBucket) (string, error) {
	return "", nil
}
func (*NilModular) CreateGlobalVirtualGroup(context.Context, *virtualgrouptypes.MsgCreateGlobalVirtualGroup) (string, error) {
	return "", ErrNilModular
}
func (*NilModular) SignMigratePiece(ctx context.Context, task *gfsptask.GfSpMigratePieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) CompleteMigrateBucket(ctx context.Context, migrateBucket *storagetypes.MsgCompleteMigrateBucket) (string, error) {
	return "", ErrNilModular
}
func (*NilModular) SignSecondarySPMigrationBucket(ctx context.Context, signDoc *storagetypes.SecondarySpMigrationBucketSignDoc) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) SwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) (string, error) {
	return "", ErrNilModular
}
func (*NilModular) SignSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) ([]byte, error) {
	return nil, ErrNilModular
}
func (*NilModular) CompleteSwapOut(ctx context.Context, completeSwapOut *virtualgrouptypes.MsgCompleteSwapOut) (string, error) {
	return "", ErrNilModular
}
func (*NilModular) SPExit(ctx context.Context, spExit *virtualgrouptypes.MsgStorageProviderExit) (string, error) {
	return "", ErrNilModular
}
func (*NilModular) CompleteSPExit(ctx context.Context, completeSPExit *virtualgrouptypes.MsgCompleteStorageProviderExit) (string, error) {
	return "", ErrNilModular
}

func (*NilModular) UpdateSPPrice(ctx context.Context, price *sptypes.MsgUpdateSpStoragePrice) (string, error) {
	return "", ErrNilModular
}

var _ Receiver = (*NullReceiveModular)(nil)

type NullReceiveModular struct{}

func (*NullReceiveModular) Name() string                { return "" }
func (*NullReceiveModular) Start(context.Context) error { return nil }
func (*NullReceiveModular) Stop(context.Context) error  { return nil }
func (*NullReceiveModular) ReserveResource(context.Context, *corercmgr.ScopeStat) (corercmgr.ResourceScopeSpan, error) {
	return &corercmgr.NullScope{}, nil
}
func (*NullReceiveModular) ReleaseResource(context.Context, corercmgr.ResourceScopeSpan) {}
func (*NullReceiveModular) QueryTasks(ctx context.Context, keyPrefix coretask.TKey) ([]coretask.Task, error) {
	return nil, ErrNilModular
}
func (*NullReceiveModular) HandleReceivePieceTask(context.Context, coretask.ReceivePieceTask, []byte) error {
	return ErrNilModular
}
func (*NullReceiveModular) HandleDoneReceivePieceTask(context.Context, coretask.ReceivePieceTask) ([]byte, error) {
	return nil, ErrNilModular
}
