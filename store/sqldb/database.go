package sqldb

import (
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
)

// Job interface
type Job interface {
	// CreateUploadJob create upload job and return job context
	CreateUploadJob(objectInfo *storagetypes.ObjectInfo) (*servicetypes.JobContext, error)
	// UpdateJobState update the state of a job by object id
	UpdateJobState(objectID uint64, state servicetypes.JobState) error
	// GetJobByID get job context by job id and return job context
	GetJobByID(jobID uint64) (*servicetypes.JobContext, error)
	// GetJobByObjectID get job context by object id
	GetJobByObjectID(objectID uint64) (*servicetypes.JobContext, error)

	// Iterator
	// Batch
}

// Object interface
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
	// UpdateAllSP update all sp info, delete old sp info
	UpdateAllSP(spList []*sptypes.StorageProvider) error
	// FetchAllSP if status is nil return all sp info; otherwise return sp info by status
	FetchAllSP(status ...sptypes.Status) ([]*sptypes.StorageProvider, error)
	// FetchAllSPWithoutOwnSP if status is nil return all sp info without own sp;
	// otherwise return sp info by status without own sp
	FetchAllSPWithoutOwnSP(status ...sptypes.Status) ([]*sptypes.StorageProvider, error)
	// GetSPByAddress return sp info by address and addressType
	GetSPByAddress(address string, addressType SPAddressType) (*sptypes.StorageProvider, error)
	// GetSPByEndpoint return sp info by endpoint
	GetSPByEndpoint(endpoint string) (*sptypes.StorageProvider, error)

	// GetOwnSPInfo return own sp info
	GetOwnSPInfo() (*sptypes.StorageProvider, error)
	// SetOwnSPInfo set(maybe overwrite) own sp info
	SetOwnSPInfo(sp *sptypes.StorageProvider) error
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

// SPDB contains all the methods required by sql database
type SPDB interface {
	Job
	Object
	ObjectIntegrity
	Traffic
	SPInfo
	StorageParam
}
