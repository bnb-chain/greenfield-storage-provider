package sqldb

import (
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/greenfield-storage-provider/model/metadata"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
)

// Job interface which contains job related to object id interface
type Job interface {
	// CreateUploadJob create upload job and return job context
	CreateUploadJob(objectInfo *storagetypes.ObjectInfo) (*servicetypes.JobContext, error)
	// UpdateJobState update the state of a job by object id
	UpdateJobState(objectID uint64, state servicetypes.JobState) error
	// GetJobByID get job context by job id and return job context
	GetJobByID(jobID uint64) (*servicetypes.JobContext, error)
	// GetJobByObjectID get job context by object id
	GetJobByObjectID(objectID uint64) (*servicetypes.JobContext, error)

	// TODO:: supports Iterator and Batch interface for gc the job and retry the job
	// Iterator
	// Batch
}

// Object interface which contains get and set object info interface
type Object interface {
	// GetObjectInfo get object info by object id
	GetObjectInfo(objectID uint64) (*storagetypes.ObjectInfo, error)
	// SetObjectInfo set(maybe overwrite) object info by object id
	SetObjectInfo(objectID uint64, objectInfo *storagetypes.ObjectInfo) error
}

// ObjectIntegrity abstract object integrity interface
type ObjectIntegrity interface {
	// GetIntegrityMeta get integrity meta info by object id
	GetObjectIntegrity(objectID uint64) (*IntegrityMeta, error)
	// SetIntegrityMeta set(maybe overwrite) integrity hash info to db
	SetObjectIntegrity(integrity *IntegrityMeta) error
}

// SPInfo interface
type SPInfo interface {
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

// StorageParam interface
type StorageParam interface {
	// GetStorageParams return storage params
	GetStorageParams() (*storagetypes.Params, error)
	// SetStorageParams set(maybe overwrite) storage params
	SetStorageParams(params *storagetypes.Params) error
}

// Traffic define a series of traffic interfaces
type Traffic interface {
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

// ServiceConfig defines a series of reading and setting service config interfaces
type ServiceConfig interface {
	GetAllServiceConfigs() (string, string, error)
	SetAllServiceConfigs(version, config string) error
}

// Metadata contains all the methods required by bs db database
type Metadata interface {
	// GetUserBuckets get buckets info by a user address
	GetUserBuckets(accountID common.Address) ([]*metadata.Bucket, error)
	// GetUserBucketsCount get buckets count by a user address
	GetUserBucketsCount(accountID common.Address) (int64, error)
	// GetBucketByName get buckets info by a bucket name
	GetBucketByName(bucketName string, isFullList bool) (*metadata.Bucket, error)
	// GetBucketByID get buckets info by by a bucket id
	GetBucketByID(bucketID int64, isFullList bool) (*metadata.Bucket, error)
	// GetLatestBlockNumber get current latest block number
	GetLatestBlockNumber() (int64, error)
	// ListObjectsByBucketName list objects info by a bucket name
	ListObjectsByBucketName(bucketName string) ([]*metadata.Object, error)
	// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
	ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, isFullList bool) ([]*metadata.Object, error)
}

// SPDB contains all the methods required by sql database
type SPDB interface {
	Job
	Object
	ObjectIntegrity
	Traffic
	SPInfo
	StorageParam
	Metadata
}
