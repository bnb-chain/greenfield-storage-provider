package gfspclient

import (
	"context"
	"io"
	"net/http"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield/types/resource"
	payment_types "github.com/bnb-chain/greenfield/x/payment/types"
	permission_types "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// GfSpClientAPI for mock use
//
//go:generate mockgen -source=./interface.go -destination=./interface_mock.go -package=gfspclient
type GfSpClientAPI interface {
	ApproverAPI
	AuthenticatorAPI
	DownloaderAPI
	GaterAPI
	ManagerAPI
	MetadataAPI
	P2PAPI
	QueryAPI
	ReceiverAPI
	SignerAPI
	UploaderAPI
	GfSpConnAPI
}

// ApproverAPI for mock use
type ApproverAPI interface {
	AskCreateBucketApproval(ctx context.Context, t coretask.ApprovalCreateBucketTask) (bool, coretask.ApprovalCreateBucketTask, error)
	AskMigrateBucketApproval(ctx context.Context, t coretask.ApprovalMigrateBucketTask) (bool, coretask.ApprovalMigrateBucketTask, error)
	AskCreateObjectApproval(ctx context.Context, t coretask.ApprovalCreateObjectTask) (bool, coretask.ApprovalCreateObjectTask, error)
}

// AuthenticatorAPI for mock use
type AuthenticatorAPI interface {
	VerifyAuthentication(ctx context.Context, auth coremodule.AuthOpType, account, bucket, object string, opts ...grpc.DialOption) (bool, error)
	GetAuthNonce(ctx context.Context, account string, domain string, opts ...grpc.DialOption) (currentNonce int32, nextNonce int32, currentPublicKey string, expiryDate int64, err error)
	UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32, userPublicKey string, expiryDate int64, opts ...grpc.DialOption) (bool, error)
	VerifyGNFD1EddsaSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign []byte, opts ...grpc.DialOption) (bool, error)
}

// DownloaderAPI for mock use
type DownloaderAPI interface {
	GetObject(ctx context.Context, downloadObjectTask coretask.DownloadObjectTask, opts ...grpc.DialOption) ([]byte, error)
	GetPiece(ctx context.Context, downloadPieceTask coretask.DownloadPieceTask, opts ...grpc.DialOption) ([]byte, error)
	GetChallengeInfo(ctx context.Context, challengePieceTask coretask.ChallengePieceTask, opts ...grpc.DialOption) ([]byte, [][]byte, []byte, error)
	RecoupQuota(ctx context.Context, bucketID, extraQuota uint64, yearMonth string, opts ...grpc.DialOption) error
}

// GaterAPI for mock use
type GaterAPI interface {
	ReplicatePieceToSecondary(ctx context.Context, endpoint string, receive coretask.ReceivePieceTask, data []byte) error
	GetPieceFromECChunks(ctx context.Context, endpoint string, task coretask.RecoveryPieceTask) (io.ReadCloser, error)
	DoneReplicatePieceToSecondary(ctx context.Context, endpoint string, receive coretask.ReceivePieceTask) ([]byte, error)
	MigratePiece(ctx context.Context, task *gfsptask.GfSpMigratePieceTask) ([]byte, error)
	NotifyDestSPMigrateSwapOut(ctx context.Context, destEndpoint string, swapOut *virtualgrouptypes.MsgSwapOut) error
	GetSecondarySPMigrationBucketApproval(ctx context.Context, secondarySPEndpoint string, signDoc *storagetypes.SecondarySpMigrationBucketSignDoc) ([]byte, error)
	GetSwapOutApproval(ctx context.Context, destSPEndpoint string, swapOutApproval *virtualgrouptypes.MsgSwapOut) (*virtualgrouptypes.MsgSwapOut, error)
	NotifySrcSPBucketMigrationDone(ctx context.Context, srcEndpoint string, bucketID uint64) error
}

// ManagerAPI for mock use
type ManagerAPI interface {
	CreateUploadObject(ctx context.Context, task coretask.UploadObjectTask) error
	CreateResumableUploadObject(ctx context.Context, task coretask.ResumableUploadObjectTask) error
	AskTask(ctx context.Context, limit corercmgr.Limit) (coretask.Task, error)
	ReportTask(ctx context.Context, report coretask.Task) error
	PickVirtualGroupFamilyID(ctx context.Context, task coretask.ApprovalCreateBucketTask) (uint32, error)
	NotifyMigrateSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) error
	NotifyBucketMigrationDone(ctx context.Context, bucketID uint64) error
}

// MetadataAPI for mock use
type MetadataAPI interface {
	GetUserBucketsCount(ctx context.Context, account string, includeRemoved bool, opts ...grpc.DialOption) (int64, error)
	ListDeletedObjectsByBlockNumberRange(ctx context.Context, spOperatorAddress string, startBlockNumber uint64, endBlockNumber uint64, includePrivate bool, opts ...grpc.DialOption) ([]*types.Object, uint64, error)
	GetUserBuckets(ctx context.Context, account string, includeRemoved bool, opts ...grpc.DialOption) ([]*types.VGFInfoBucket, error)
	ListObjectsByBucketName(ctx context.Context, bucketName string, accountID string, maxKeys uint64, startAfter string, continuationToken string, delimiter string, prefix string, includeRemoved bool,
		opts ...grpc.DialOption) (objects []*types.Object, keyCount, maxKeysRe uint64, isTruncated bool, nextContinuationToken, name, prefixRe, delimiterRe string, commonPrefixes []string, continuationTokenRe string, err error)
	GetBucketByBucketName(ctx context.Context, bucketName string, includePrivate bool, opts ...grpc.DialOption) (*types.Bucket, error)
	GetBucketByBucketID(ctx context.Context, bucketID int64, includePrivate bool, opts ...grpc.DialOption) (*types.Bucket, error)
	ListExpiredBucketsBySp(ctx context.Context, createAt int64, primarySpID uint32, limit int64, opts ...grpc.DialOption) ([]*types.Bucket, error)
	GetObjectMeta(ctx context.Context, objectName string, bucketName string, includePrivate bool, opts ...grpc.DialOption) (*types.Object, error)
	GetPaymentByBucketName(ctx context.Context, bucketName string, includePrivate bool, opts ...grpc.DialOption) (*payment_types.StreamRecord, error)
	GetPaymentByBucketID(ctx context.Context, bucketID int64, includePrivate bool, opts ...grpc.DialOption) (*payment_types.StreamRecord, error)
	VerifyPermission(ctx context.Context, Operator string, bucketName string, objectName string, actionType permission_types.ActionType, opts ...grpc.DialOption) (*permission_types.Effect, error)
	GetBucketMeta(ctx context.Context, bucketName string, includePrivate bool, opts ...grpc.DialOption) (*types.Bucket, *payment_types.StreamRecord, error)
	GetEndpointBySpID(ctx context.Context, spID uint32, opts ...grpc.DialOption) (string, error)
	GetBucketReadQuota(ctx context.Context, bucket *storagetypes.BucketInfo, yearMonth string, opts ...grpc.DialOption) (uint64, uint64, uint64, uint64, error)
	ListBucketReadRecord(ctx context.Context, bucket *storagetypes.BucketInfo, startTimestampUs, endTimestampUs, maxRecordNum int64, opts ...grpc.DialOption) ([]*types.ReadRecord, int64, error)
	GetUploadObjectState(ctx context.Context, objectID uint64, opts ...grpc.DialOption) (int32, string, error)
	GetUploadObjectSegment(ctx context.Context, objectID uint64, opts ...grpc.DialOption) (uint32, error)
	GetGroupList(ctx context.Context, name string, prefix string, sourceType string, limit int64, offset int64, includeRemoved bool, opts ...grpc.DialOption) ([]*types.Group, int64, error)
	ListBucketsByIDs(ctx context.Context, bucketIDs []uint64, includeRemoved bool, opts ...grpc.DialOption) (map[uint64]*types.Bucket, error)
	ListObjectsByIDs(ctx context.Context, objectIDs []uint64, includeRemoved bool, opts ...grpc.DialOption) (map[uint64]*types.Object, error)
	ListVirtualGroupFamiliesSpID(ctx context.Context, spID uint32, opts ...grpc.DialOption) ([]*virtualgrouptypes.GlobalVirtualGroupFamily, error)
	GetGlobalVirtualGroupByGvgID(ctx context.Context, gvgID uint32, opts ...grpc.DialOption) (*virtualgrouptypes.GlobalVirtualGroup, error)
	ListObjectsInGVGAndBucket(ctx context.Context, gvgID uint32, bucketID uint64, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.ObjectDetails, error)
	ListObjectsInGVG(ctx context.Context, gvgID uint32, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.ObjectDetails, error)
	ListObjectsByGVGAndBucketForGC(ctx context.Context, gvgID uint32, bucketID uint64, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.ObjectDetails, error)
	GetVirtualGroupFamily(ctx context.Context, vgfID uint32, opts ...grpc.DialOption) (*virtualgrouptypes.GlobalVirtualGroupFamily, error)
	GetGlobalVirtualGroup(ctx context.Context, bucketID uint64, lvgID uint32, opts ...grpc.DialOption) (*virtualgrouptypes.GlobalVirtualGroup, error)
	ListGlobalVirtualGroupsByBucket(ctx context.Context, bucketID uint64, opts ...grpc.DialOption) ([]*virtualgrouptypes.GlobalVirtualGroup, error)
	ListGlobalVirtualGroupsBySecondarySP(ctx context.Context, spID uint32, opts ...grpc.DialOption) ([]*virtualgrouptypes.GlobalVirtualGroup, error)
	ListMigrateBucketEvents(ctx context.Context, blockID uint64, spID uint32, opts ...grpc.DialOption) ([]*types.ListMigrateBucketEvents, error)
	ListSwapOutEvents(ctx context.Context, blockID uint64, spID uint32, opts ...grpc.DialOption) ([]*types.ListSwapOutEvents, error)
	ListSpExitEvents(ctx context.Context, blockID uint64, spID uint32, opts ...grpc.DialOption) (*types.ListSpExitEvents, error)
	GetObjectByID(ctx context.Context, objectID uint64, opts ...grpc.DialOption) (*storagetypes.ObjectInfo, error)
	VerifyPermissionByID(ctx context.Context, Operator string, resourceType resource.ResourceType, resourceID uint64, actionType permission_types.ActionType, opts ...grpc.DialOption) (*permission_types.Effect, error)
	GetSPInfo(ctx context.Context, operatorAddress string, opts ...grpc.DialOption) (*sptypes.StorageProvider, error)
	GetStatus(ctx context.Context, opts ...grpc.DialOption) (*types.Status, error)
	GetUserGroups(ctx context.Context, accountID string, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.GroupMember, error)
	GetGroupMembers(ctx context.Context, groupID uint64, startAfter string, limit uint32, opts ...grpc.DialOption) ([]*types.GroupMember, error)
	GetUserOwnedGroups(ctx context.Context, accountID string, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.GroupMember, error)
	ListObjectPolicies(ctx context.Context, objectName, bucketName string, startAfter uint64, actionType int32, limit uint32, opts ...grpc.DialOption) ([]*types.Policy, error)
	ListPaymentAccountStreams(ctx context.Context, paymentAccount string, opts ...grpc.DialOption) ([]*types.Bucket, error)
	ListUserPaymentAccounts(ctx context.Context, accountID string, opts ...grpc.DialOption) ([]*types.StreamRecordMeta, error)
	ListGroupsByIDs(ctx context.Context, groupIDs []uint64, opts ...grpc.DialOption) (map[uint64]*types.Group, error)
	GetSPMigratingBucketNumber(ctx context.Context, spID uint32, opts ...grpc.DialOption) (uint64, error)
}

// P2PAPI for mock use
type P2PAPI interface {
	AskSecondaryReplicatePieceApproval(ctx context.Context, task coretask.ApprovalReplicatePieceTask, low, high int, timeout int64) ([]*gfsptask.GfSpReplicatePieceApprovalTask, error)
	QueryP2PBootstrap(ctx context.Context) ([]string, error)
}

// QueryAPI for mock use
type QueryAPI interface {
	QueryTasks(ctx context.Context, endpoint string, subKey string, opts ...grpc.DialOption) ([]string, error)
	QueryBucketMigrate(ctx context.Context, endpoint string, opts ...grpc.DialOption) (string, error)
	QuerySPExit(ctx context.Context, endpoint string, opts ...grpc.DialOption) (string, error)
}

// ReceiverAPI for mock use
type ReceiverAPI interface {
	ReplicatePiece(ctx context.Context, task coretask.ReceivePieceTask, data []byte, opts ...grpc.DialOption) error
	DoneReplicatePiece(ctx context.Context, task coretask.ReceivePieceTask, opts ...grpc.DialOption) ([]byte, []byte, error)
}

// SignerAPI for mock use
type SignerAPI interface {
	SignCreateBucketApproval(ctx context.Context, bucket *storagetypes.MsgCreateBucket) ([]byte, error)
	SignMigrateBucketApproval(ctx context.Context, bucket *storagetypes.MsgMigrateBucket) ([]byte, error)
	SignCreateObjectApproval(ctx context.Context, object *storagetypes.MsgCreateObject) ([]byte, error)
	SealObject(ctx context.Context, object *storagetypes.MsgSealObject) (string, error)
	UpdateSPPrice(ctx context.Context, price *sptypes.MsgUpdateSpStoragePrice) (string, error)
	CreateGlobalVirtualGroup(ctx context.Context, group *gfspserver.GfSpCreateGlobalVirtualGroup) error
	RejectUnSealObject(ctx context.Context, object *storagetypes.MsgRejectSealObject) (string, error)
	DiscontinueBucket(ctx context.Context, bucket *storagetypes.MsgDiscontinueBucket) (string, error)
	SignReplicatePieceApproval(ctx context.Context, task coretask.ApprovalReplicatePieceTask) ([]byte, error)
	SignSecondarySealBls(ctx context.Context, objectID uint64, gvgId uint32, checksums [][]byte) ([]byte, error)
	SignReceiveTask(ctx context.Context, receiveTask coretask.ReceivePieceTask) ([]byte, error)
	SignRecoveryTask(ctx context.Context, recoveryTask coretask.RecoveryPieceTask) ([]byte, error)
	SignP2PPingMsg(ctx context.Context, ping *gfspp2p.GfSpPing) ([]byte, error)
	SignP2PPongMsg(ctx context.Context, pong *gfspp2p.GfSpPong) ([]byte, error)
	SignMigratePiece(ctx context.Context, task *gfsptask.GfSpMigratePieceTask) ([]byte, error)
	CompleteMigrateBucket(ctx context.Context, migrateBucket *storagetypes.MsgCompleteMigrateBucket) (string, error)
	SignSecondarySPMigrationBucket(ctx context.Context, signDoc *storagetypes.SecondarySpMigrationBucketSignDoc) ([]byte, error)
	SignSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) ([]byte, error)
	SwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) (string, error)
	CompleteSwapOut(ctx context.Context, completeSwapOut *virtualgrouptypes.MsgCompleteSwapOut) (string, error)
	SPExit(ctx context.Context, spExit *virtualgrouptypes.MsgStorageProviderExit) (string, error)
	CompleteSPExit(ctx context.Context, completeSPExit *virtualgrouptypes.MsgCompleteStorageProviderExit) (string, error)
}

// UploaderAPI for mock use
type UploaderAPI interface {
	UploadObject(ctx context.Context, task coretask.UploadObjectTask, stream io.Reader, opts ...grpc.DialOption) error
	ResumableUploadObject(ctx context.Context, task coretask.ResumableUploadObjectTask, stream io.Reader, opts ...grpc.DialOption) error
}

// GfSpConnAPI for mock use
type GfSpConnAPI interface {
	Connection(ctx context.Context, address string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	ManagerConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	ApproverConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	P2PConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	SignerConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	HTTPClient(ctx context.Context) *http.Client
	Close() error
}

// stdLib for mock use
// Note: stdLib interface is forbidden to be used in non-UT code
// nolint:unused
type stdLib interface {
	io.Reader
	io.ReadCloser
}
