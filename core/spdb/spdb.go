package spdb

import (
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

// UploadObjectProgressDB interface which records upload object related progress(includes foreground and background) and state.
type UploadObjectProgressDB interface {
	// InsertUploadProgress inserts a new upload object progress.
	InsertUploadProgress(objectID uint64) error
	// DeleteUploadProgress deletes the upload object progress.
	DeleteUploadProgress(objectID uint64) error
	// UpdateUploadProgress updates upload object progress state.
	UpdateUploadProgress(uploadMeta *UploadObjectMeta) error
	// GetUploadState queries the task state by object id.
	GetUploadState(objectID uint64) (storetypes.TaskState, error)
	// TODO: support load
}

// GCObjectProgressDB interface which records gc object related progress.
// TODO: refine interface
type GCObjectProgressDB interface {
	// InsertGCObjectProgress inserts a new gc object progress.
	InsertGCObjectProgress(taskKey string) error
	// DeleteGCObjectProgress deletes the gc object progress.
	DeleteGCObjectProgress(taskKey string) error
	// UpdateGCObjectProgress updates gc object progress state.
	UpdateGCObjectProgress(gcMeta *GCObjectMeta) error
	// TODO: refine it
	GetAllGCObjectTask(taskKey string) []task.GCObjectTask
}

// ObjectIntegrityDB abstract object integrity interface.
type ObjectIntegrityDB interface {
	// GetObjectIntegrity gets integrity meta info by object id.
	GetObjectIntegrity(objectID uint64) (*IntegrityMeta, error)
	// SetObjectIntegrity sets(maybe overwrite) integrity hash info to db.
	SetObjectIntegrity(integrity *IntegrityMeta) error
	// DeleteObjectIntegrity deletes the integrity hash.
	DeleteObjectIntegrity(objectID uint64) error

	GetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32) ([]byte, error)
	SetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32, checksum []byte) error
	DeleteReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32) error
	GetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) ([][]byte, error)
	SetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32, checksum [][]byte) error
	DeleteAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) error
}

// TrafficDB define a series of traffic interfaces
type TrafficDB interface {
	// CheckQuotaAndAddReadRecord create bucket traffic firstly if bucket is not existed,
	// and check whether the added traffic record exceeds the quota, if it exceeds the quota,
	// it will return error, Otherwise, add a record and return nil.
	CheckQuotaAndAddReadRecord(record *ReadRecord, quota *BucketQuota) error
	// GetBucketTraffic return bucket traffic info,
	// notice maybe return (nil, nil) while there is no bucket traffic
	GetBucketTraffic(bucketID uint64, yearMonth string) (*BucketTraffic, error)
	// GetReadRecord return record list by time range
	GetReadRecord(timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetBucketReadRecord return bucket record list by time range
	GetBucketReadRecord(bucketID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetObjectReadRecord return object record list by time range
	GetObjectReadRecord(objectID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetUserReadRecord return user record list by time range
	GetUserReadRecord(userAddress string, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
}

type SPInfoDB interface {
	// UpdateAllSp update all sp info, delete old sp info
	UpdateAllSp(spList []*sptypes.StorageProvider) error
	// FetchAllSp if status is nil return all sp info; otherwise return sp info by status
	FetchAllSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error)
	// FetchAllSpWithoutOwnSp if status is nil return all sp info without own sp;
	// otherwise return sp info by status without own sp
	FetchAllSpWithoutOwnSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error)
	// GetSpByAddress return sp info by address and addressType
	GetSpByAddress(address string, addressType SpAddressType) (*sptypes.StorageProvider, error)
	// GetSpByEndpoint return sp info by endpoint
	GetSpByEndpoint(endpoint string) (*sptypes.StorageProvider, error)
	// GetOwnSpInfo return own sp info
	GetOwnSpInfo() (*sptypes.StorageProvider, error)
	// SetOwnSpInfo set(maybe overwrite) own sp info
	SetOwnSpInfo(sp *sptypes.StorageProvider) error
}

// OffChainAuthKeyDB interface
type OffChainAuthKeyDB interface {
	GetAuthKey(userAddress string, domain string) (*OffChainAuthKey, error)
	UpdateAuthKey(userAddress string, domain string, oldNonce int32, newNonce int32, newPublicKey string, newExpiryDate time.Time) error
	InsertAuthKey(newRecord *OffChainAuthKey) error
}

type SPDB interface {
	UploadObjectProgressDB
	GCObjectProgressDB
	ObjectIntegrityDB
	TrafficDB
	SPInfoDB
	OffChainAuthKeyDB
}
