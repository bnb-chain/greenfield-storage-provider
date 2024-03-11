package spdb

import (
	"time"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

// SPDB contains all sp db operations
//
//go:generate mockgen -source=./spdb.go -destination=./spdb_mock.go -package=spdb
type SPDB interface {
	UploadObjectProgressDB
	GCObjectProgressDB
	SignatureDB
	TrafficDB
	SPInfoDB
	OffChainAuthKeyDB
	OffChainAuthKeyV2DB
	MigrateDB
	ExitRecoverDB
}

// UploadObjectProgressDB interface which records upload object related progress(includes foreground and background) and state.
type UploadObjectProgressDB interface {
	// InsertUploadProgress inserts a new upload object progress.
	InsertUploadProgress(objectID uint64) error
	// DeleteUploadProgress deletes the upload object progress.
	DeleteUploadProgress(objectID uint64) error
	// UpdateUploadProgress updates the upload object progress state.
	UpdateUploadProgress(uploadMeta *UploadObjectMeta) error
	// GetUploadState queries the task state by object id.
	GetUploadState(objectID uint64) (storetypes.TaskState, string, error)
	// GetUploadMetasToReplicate queries the latest upload_done/replicate_doing object to continue replicate.
	// It is only used in startup.
	GetUploadMetasToReplicate(limit int, timeout int64) ([]*UploadObjectMeta, error)
	// GetUploadMetasToSeal queries the latest replicate_done/seal_doing object to continue seal.
	// It is only used in startup.
	GetUploadMetasToSeal(limit int, timeout int64) ([]*UploadObjectMeta, error)
	// GetUploadMetasToReplicateByStartTS queries the upload_done/replicate_doing object to continue replicate.
	// It is used in task retry scheduler.
	GetUploadMetasToReplicateByStartTS(limit int, startTimeStamp int64) ([]*UploadObjectMeta, error)
	// GetUploadMetasToSealByStartTS queries the replicate_done/seal_doing object to continue seal.
	// It is used in task retry scheduler.
	GetUploadMetasToSealByStartTS(limit int, startTimeStamp int64) ([]*UploadObjectMeta, error)
	// GetUploadMetasToRejectUnsealByRangeTS queries the upload_done/replicate_doing object to reject.
	// It is used in task retry scheduler.
	GetUploadMetasToRejectUnsealByRangeTS(limit int, startTimeStamp int64, endTimeStamp int64) ([]*UploadObjectMeta, error)
	// InsertPutEvent inserts a new upload event progress.
	InsertPutEvent(task coretask.Task) error
}

// GCObjectProgressDB interface which records gc object related progress.
type GCObjectProgressDB interface {
	// InsertGCObjectProgress inserts a new gc object progress.
	InsertGCObjectProgress(gcMeta *GCObjectMeta) error
	// DeleteGCObjectProgress deletes the gc object progress.
	DeleteGCObjectProgress(taskKey string) error
	// UpdateGCObjectProgress updates the gc object progress.
	UpdateGCObjectProgress(gcMeta *GCObjectMeta) error
	// GetGCMetasToGC queries the latest gc meta to continue gc.
	// It is only used in startup.
	GetGCMetasToGC(limit int) ([]*GCObjectMeta, error)
}

// SignatureDB abstract object integrity interface.
type SignatureDB interface {
	/*
		Object Signature is used to get challenge info.
	*/
	// GetObjectIntegrity gets integrity meta info by object id and redundancy index.
	GetObjectIntegrity(objectID uint64, redundancyIndex int32) (*IntegrityMeta, error)
	// SetObjectIntegrity sets(maybe overwrite) integrity hash info to db.
	SetObjectIntegrity(integrity *IntegrityMeta) error
	// DeleteObjectIntegrity deletes the integrity hash.
	DeleteObjectIntegrity(objectID uint64, redundancyIndex int32) error
	// UpdateIntegrityChecksum update IntegrityMetaTable's integrity checksum
	UpdateIntegrityChecksum(integrity *IntegrityMeta) error
	// UpdatePieceChecksum if the IntegrityMetaTable already exists, it will be appended to the existing PieceChecksumList.
	UpdatePieceChecksum(objectID uint64, redundancyIndex int32, checksum []byte) error
	// UpdateIntegrityMeta update both piece checksum and integrity
	UpdateIntegrityMeta(integrity *IntegrityMeta) error
	// ListIntegrityMetaByObjectIDRange list integrity meta in range
	ListIntegrityMetaByObjectIDRange(startObjectID int64, endObjectID int64, includePrivate bool) ([]*IntegrityMeta, error)

	/*
		Piece Signature is used to help replicate object's piece data to secondary sps, which is temporary.
	*/
	// SetReplicatePieceChecksum sets(maybe overwrite) the piece hash.
	SetReplicatePieceChecksum(objectID uint64, segmentIdx uint32, redundancyIdx int32, checksum []byte, version int64) error
	// GetReplicatePieceChecksum gets a piece hash.
	GetReplicatePieceChecksum(objectID uint64, segmentIdx uint32, redundancyIdx int32) ([]byte, error)
	// GetAllReplicatePieceChecksum gets all piece hashes.
	GetAllReplicatePieceChecksum(objectID uint64, redundancyIdx int32, pieceCount uint32) ([][]byte, error)
	// GetAllReplicatePieceChecksumOptimized gets all piece hashes.
	GetAllReplicatePieceChecksumOptimized(objectID uint64, redundancyIdx int32, pieceCount uint32) ([][]byte, error)
	// DeleteReplicatePieceChecksum deletes piece hashes.
	DeleteReplicatePieceChecksum(objectID uint64, segmentIdx uint32, redundancyIdx int32) (err error)
	// DeleteAllReplicatePieceChecksum deletes all piece hashes.
	DeleteAllReplicatePieceChecksum(objectID uint64, redundancyIdx int32, pieceCount uint32) error
	// DeleteAllReplicatePieceChecksumOptimized deletes all piece hashes.
	DeleteAllReplicatePieceChecksumOptimized(objectID uint64, redundancyIdx int32) error
	// ListReplicatePieceChecksumByObjectIDRange list replicate piece checksum in range
	ListReplicatePieceChecksumByObjectIDRange(startObjectID int64, endObjectID int64) ([]*GCPieceMeta, error)

	/*
	  Used for Object Update, the ShadowObjectIntegrity is used for temporally storing used updated object meta;
	  it will be clean when the object is re-sealed
	*/
	// GetShadowObjectIntegrity gets the shadow object integrity meta info by object id and redundancy index.
	GetShadowObjectIntegrity(objectID uint64, redundancyIndex int32) (*ShadowIntegrityMeta, error)
	// SetShadowObjectIntegrity sets(maybe overwrite) shadow integrity hash info to db.
	SetShadowObjectIntegrity(integrity *ShadowIntegrityMeta) error
	// DeleteShadowObjectIntegrity deletes the shadow integrity hash.
	DeleteShadowObjectIntegrity(objectID uint64, redundancyIndex int32) error
	// UpdateShadowIntegrityChecksum update ShadowIntegrityMetaTable's integrity checksum
	UpdateShadowIntegrityChecksum(integrity *ShadowIntegrityMeta) error
	// UpdateShadowPieceChecksum if the IntegrityMetaTable already exists, it will be appended to the existing PieceChecksumList.
	UpdateShadowPieceChecksum(objectID uint64, redundancyIndex int32, checksum []byte, version int64) error
	// ListShadowIntegrityMeta list Shadow IntegrityMeta records with default limit
	ListShadowIntegrityMeta() ([]*ShadowIntegrityMeta, error)
}

// TrafficDB defines a series of traffic interfaces.
type TrafficDB interface {
	// CheckQuotaAndAddReadRecord get the traffic info from db, update the quota meta and check
	// whether the added traffic record exceeds the quota, if it exceeds the quota,
	// it will return error, Otherwise, add a record and return nil.
	CheckQuotaAndAddReadRecord(record *ReadRecord, quota *BucketQuota) error
	// InitBucketTraffic init the traffic info
	InitBucketTraffic(record *ReadRecord, quota *BucketQuota) error
	// GetBucketTraffic return bucket traffic info,
	// notice maybe return (nil, nil) while there is no bucket traffic.
	GetBucketTraffic(bucketID uint64, yearMonth string) (*BucketTraffic, error)
	// UpdateExtraQuota update the read consumed quota and free consumed quota in traffic db with the extra quota
	UpdateExtraQuota(bucketID, extraQuota uint64, yearMonth string) error
	// GetLatestBucketTraffic return latest bucket traffic info
	GetLatestBucketTraffic(bucketID uint64) (*BucketTraffic, error)
	// UpdateBucketTraffic update bucket traffic info
	UpdateBucketTraffic(bucketID uint64, update *BucketTraffic) (err error)
	// GetReadRecord return record list by time range.
	GetReadRecord(timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetBucketReadRecord return bucket record list by time range.
	GetBucketReadRecord(bucketID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetObjectReadRecord return object record list by time range.
	GetObjectReadRecord(objectID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetUserReadRecord return user record list by time range.
	GetUserReadRecord(userAddress string, timeRange *TrafficTimeRange) ([]*ReadRecord, error)

	// DeleteExpiredReadRecord delete all read record before ts with limit
	DeleteExpiredReadRecord(ts, limit uint64) (err error)
	// DeleteExpiredBucketTraffic delete all bucket traffic before yearMonth, when large dataset
	DeleteExpiredBucketTraffic(yearMonth string) (err error)
}

// SPInfoDB defines a series of sp interfaces.
type SPInfoDB interface {
	// UpdateAllSp update all sp info, delete old sp info.
	UpdateAllSp(spList []*sptypes.StorageProvider) error
	// FetchAllSp if status is nil return all sp info; otherwise return sp info by status.
	FetchAllSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error)
	// FetchAllSpWithoutOwnSp if status is nil return all sp info without own sp;
	// otherwise return sp info by status without own sp.
	FetchAllSpWithoutOwnSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error)
	// GetSpByAddress return sp info by address and addressType.
	GetSpByAddress(address string, addressType SpAddressType) (*sptypes.StorageProvider, error)
	// GetSpByEndpoint return sp info by endpoint.
	GetSpByEndpoint(endpoint string) (*sptypes.StorageProvider, error)
	// GetSpByID return sp info by id.
	GetSpByID(id uint32) (*sptypes.StorageProvider, error)
	// GetOwnSpInfo return own sp info.
	GetOwnSpInfo() (*sptypes.StorageProvider, error)
	// SetOwnSpInfo set(maybe overwrite) own sp info.
	SetOwnSpInfo(sp *sptypes.StorageProvider) error
}

// OffChainAuthKeyDB interface.
type OffChainAuthKeyDB interface {
	GetAuthKey(userAddress string, domain string) (*OffChainAuthKey, error)
	UpdateAuthKey(userAddress string, domain string, oldNonce int32, newNonce int32, newPublicKey string, newExpiryDate time.Time) error
	InsertAuthKey(newRecord *OffChainAuthKey) error
}

// OffChainAuthKeyV2DB interface.
type OffChainAuthKeyV2DB interface {
	GetAuthKeyV2(userAddress string, domain string, publicKey string) (*OffChainAuthKeyV2, error)
	InsertAuthKeyV2(newRecord *OffChainAuthKeyV2) error
	ClearExpiredOffChainAuthKeys() error
	ListAuthKeysV2(userAddress string, domain string) ([]string, error)
	DeleteAuthKeysV2(userAddress string, domain string, publicKey []string) (bool, error)
}

// MigrateDB is used to support sp exit and bucket migrate.
type MigrateDB interface {
	// UpdateSPExitSubscribeProgress includes insert and update.
	UpdateSPExitSubscribeProgress(blockHeight uint64) error
	// QuerySPExitSubscribeProgress returns blockHeight which is called at startup.
	QuerySPExitSubscribeProgress() (uint64, error)
	// UpdateSwapOutSubscribeProgress includes insert and update.
	UpdateSwapOutSubscribeProgress(blockHeight uint64) error
	// QuerySwapOutSubscribeProgress returns blockHeight which is called at startup.
	QuerySwapOutSubscribeProgress() (uint64, error)
	// UpdateBucketMigrateSubscribeProgress includes insert and update.
	UpdateBucketMigrateSubscribeProgress(blockHeight uint64) error
	// UpdateBucketMigrateGCSubscribeProgress includes insert and update.
	UpdateBucketMigrateGCSubscribeProgress(blockHeight uint64) error
	// QueryBucketMigrateSubscribeProgress returns blockHeight which is called at startup.
	QueryBucketMigrateSubscribeProgress() (uint64, error)

	// InsertSwapOutUnit is used to insert a swap out unit.
	InsertSwapOutUnit(meta *SwapOutMeta) error
	// UpdateSwapOutUnitCompletedGVGList is used to record dest swap out progress, manager restart can load it.
	UpdateSwapOutUnitCompletedGVGList(swapOutKey string, completedGVGList []uint32) error
	// QuerySwapOutUnitInSrcSP is used to rebuild swap out plan at startup.
	QuerySwapOutUnitInSrcSP(swapOutKey string) (*SwapOutMeta, error)
	// ListDestSPSwapOutUnits is used to rebuild swap out plan at startup.
	ListDestSPSwapOutUnits() ([]*SwapOutMeta, error)

	// InsertMigrateGVGUnit inserts a new gvg migrate unit.
	InsertMigrateGVGUnit(meta *MigrateGVGUnitMeta) error
	// DeleteMigrateGVGUnit deletes the gvg migrate unit.
	DeleteMigrateGVGUnit(meta *MigrateGVGUnitMeta) error
	// UpdateMigrateGVGUnitStatus updates gvg unit status.
	UpdateMigrateGVGUnitStatus(migrateKey string, migrateStatus int) error
	// UpdateMigrateGVGUnitLastMigrateObjectID updates gvg unit LastMigrateObjectID.
	UpdateMigrateGVGUnitLastMigrateObjectID(migrateKey string, lastMigrateObjectID uint64) error
	// UpdateMigrateGVGRetryCount updates gvg unit retry time
	UpdateMigrateGVGRetryCount(migrateKey string, retryTime int) error
	// UpdateMigrateGVGMigratedBytesSize updates gvg unit retry time
	UpdateMigrateGVGMigratedBytesSize(migrateKey string, migratedBytes uint64) error
	// QueryMigrateGVGUnit returns the gvg migrate unit info.
	QueryMigrateGVGUnit(migrateKey string) (*MigrateGVGUnitMeta, error)
	// ListMigrateGVGUnitsByBucketID is used to load at dest sp startup(bucket migrate).
	ListMigrateGVGUnitsByBucketID(bucketID uint64) ([]*MigrateGVGUnitMeta, error)
	// DeleteMigrateGVGUnitsByBucketID is used to delete migrate gvg units at bucket migrate
	DeleteMigrateGVGUnitsByBucketID(bucketID uint64) error

	// UpdateBucketMigrationProgress update MigrateBucketProgress migrate state.
	UpdateBucketMigrationProgress(bucketID uint64, migrateState int) error
	// UpdateBucketMigrationPreDeductedQuota update pre-deducted quota and migration state when src sp receives a preMigrateBucket request.
	UpdateBucketMigrationPreDeductedQuota(bucketID uint64, deductedQuota uint64, state int) error
	// UpdateBucketMigrationRecoupQuota update RecoupQuota and the corresponding state in MigrateBucketProgress.
	UpdateBucketMigrationRecoupQuota(bucketID uint64, recoupQuota uint64, state int) error
	// UpdateBucketMigrationGCProgress update bucket migration gc progress
	UpdateBucketMigrationGCProgress(progressMeta MigrateBucketProgressMeta) error
	// UpdateBucketMigrationMigratingProgress update bucket migration migrating progress
	UpdateBucketMigrationMigratingProgress(bucketID uint64, gvgUnits uint32, gvgUnitsFinished uint32) error
	// QueryMigrateBucketState returns the migrate state.
	QueryMigrateBucketState(bucketID uint64) (int, error)
	// QueryMigrateBucketProgress returns the migration progress.
	QueryMigrateBucketProgress(bucketID uint64) (*MigrateBucketProgressMeta, error)
	// ListBucketMigrationToConfirm returns the migrate bucket id to be confirmed.
	ListBucketMigrationToConfirm(migrationStates []int) ([]*MigrateBucketProgressMeta, error)
	// DeleteMigrateBucket delete the bucket migrate status
	DeleteMigrateBucket(bucketID uint64) error
}

// ExitRecoverDB is used to support sp exit and recover resource.
type ExitRecoverDB interface {
	// GetRecoverGVGStats return recover gvg stats
	GetRecoverGVGStats(gvgID uint32) (*RecoverGVGStats, error)
	// BatchGetRecoverGVGStats return recover gvg stats list
	BatchGetRecoverGVGStats(gvgID []uint32) ([]*RecoverGVGStats, error)
	// SetRecoverGVGStats insert a recover gvg stats unit
	SetRecoverGVGStats(stats []*RecoverGVGStats) error
	// UpdateRecoverGVGStats update recover gvg stats
	UpdateRecoverGVGStats(stats *RecoverGVGStats) (err error)
	// DeleteRecoverGVGStats delete recover gvg stats
	DeleteRecoverGVGStats(gvgID uint32) (err error)

	// InsertRecoverFailedObject inserts a new failed object unit.
	InsertRecoverFailedObject(object *RecoverFailedObject) error
	// UpdateRecoverFailedObject update failed object unit
	UpdateRecoverFailedObject(object *RecoverFailedObject) (err error)
	// DeleteRecoverFailedObject delete failed object unit
	DeleteRecoverFailedObject(objectID uint64) (err error)
	// GetRecoverFailedObject return the failed object.
	GetRecoverFailedObject(objectID uint64) (*RecoverFailedObject, error)
	// GetRecoverFailedObjects return the failed object by retry time
	GetRecoverFailedObjects(retry, limit uint32) ([]*RecoverFailedObject, error)
	// GetRecoverFailedObjectsByRetryTime return the failed object by retry time
	GetRecoverFailedObjectsByRetryTime(retry uint32) ([]*RecoverFailedObject, error)
	// CountRecoverFailedObject return the failed object total count
	CountRecoverFailedObject() (int64, error)
}
