package spdb

import (
	"time"

	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

const (
	GatewayBeginReceiveUpload           = "gateway_begin_receive_upload"
	GatewayEndReceiveUpload             = "gateway_end_receive_upload"
	UploaderBeginReceiveData            = "uploader_begin_receive_data"
	UploaderEndReceiveData              = "uploader_end_receive_data"
	ManagerReceiveAndWaitSchedulingTask = "manager_receive_and_wait_scheduling_task"
	ManagerSchedulingTask               = "manager_scheduling_task"
	ExecutorBeginTask                   = "executor_begin_task"
	ExecutorEndTask                     = "executor_end_task"
	ExecutorBeginP2P                    = "executor_begin_p2p"
	ExecutorEndP2P                      = "executor_end_p2p"
	ExecutorBeginReplicateOnePiece      = "executor_begin_replicate_one_piece"
	ExecutorEndReplicateOnePiece        = "executor_end_replicate_one_piece"
	ExecutorBeginReplicateAllPiece      = "executor_begin_replicate_all_piece"
	ExecutorEndReplicateAllPiece        = "executor_end_replicate_all_piece"
	ExecutorBeginDoneReplicatePiece     = "executor_begin_done_replicate_piece"
	ExecutorEndDoneReplicatePiece       = "executor_end_done_replicate_piece"
	ExecutorBeginSealTx                 = "executor_begin_seal_tx"
	ExecutorEndSealTx                   = "executor_end_seal_tx"
	ExecutorBeginConfirmSeal            = "executor_begin_confirm_seal"
	ExecutorEndConfirmSeal              = "executor_end_confirm_seal"
)

// UploadObjectProgressDB interface which records upload object related progress(includes foreground and background) and state.
type UploadObjectProgressDB interface {
	// InsertUploadProgress inserts a new upload object progress.
	InsertUploadProgress(objectID uint64) error
	// DeleteUploadProgress deletes the upload object progress.
	DeleteUploadProgress(objectID uint64) error
	// UpdateUploadProgress updates the upload object progress state.
	UpdateUploadProgress(uploadMeta *UploadObjectMeta) error
	// GetUploadState queries the task state by object id.
	GetUploadState(objectID uint64) (storetypes.TaskState, error)
	// GetUploadMetasToReplicate queries the latest upload_done/replicate_doing object to continue replicate.
	// It is only used in startup.
	GetUploadMetasToReplicate(limit int) ([]*UploadObjectMeta, error)
	// GetUploadMetasToSeal queries the latest replicate_done/seal_doing object to continue seal.
	// It is only used in startup.
	GetUploadMetasToSeal(limit int) ([]*UploadObjectMeta, error)
	// InsertUploadEvent inserts a new upload event progress.
	InsertUploadEvent(objectID uint64, state string, description string) error
}

// GCObjectProgressDB interface which records gc object related progress.
type GCObjectProgressDB interface {
	// InsertGCObjectProgress inserts a new gc object progress.
	InsertGCObjectProgress(taskKey string, gcMeta *GCObjectMeta) error
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
	// GetObjectIntegrity gets integrity meta info by object id.
	GetObjectIntegrity(objectID uint64) (*IntegrityMeta, error)
	// SetObjectIntegrity sets(maybe overwrite) integrity hash info to db.
	SetObjectIntegrity(integrity *IntegrityMeta) error
	// DeleteObjectIntegrity deletes the integrity hash.
	DeleteObjectIntegrity(objectID uint64) error
	/*
		Piece Signature is used to help replicate object's piece data to secondary sps, which is temporary.
	*/
	// SetReplicatePieceChecksum sets(maybe overwrite) the piece hash.
	SetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32, checksum []byte) error
	// GetAllReplicatePieceChecksum gets all piece hashes.
	GetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) ([][]byte, error)
	// DeleteAllReplicatePieceChecksum deletes all piece hashes.
	DeleteAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) error
}

// TrafficDB defines a series of traffic interfaces.
type TrafficDB interface {
	// CheckQuotaAndAddReadRecord create bucket traffic firstly if bucket is not existed,
	// and check whether the added traffic record exceeds the quota, if it exceeds the quota,
	// it will return error, Otherwise, add a record and return nil.
	CheckQuotaAndAddReadRecord(record *ReadRecord, quota *BucketQuota) error
	// GetBucketTraffic return bucket traffic info,
	// notice maybe return (nil, nil) while there is no bucket traffic.
	GetBucketTraffic(bucketID uint64, yearMonth string) (*BucketTraffic, error)
	// GetReadRecord return record list by time range.
	GetReadRecord(timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetBucketReadRecord return bucket record list by time range.
	GetBucketReadRecord(bucketID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetObjectReadRecord return object record list by time range.
	GetObjectReadRecord(objectID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	// GetUserReadRecord return user record list by time range.
	GetUserReadRecord(userAddress string, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
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

type SPDB interface {
	UploadObjectProgressDB
	GCObjectProgressDB
	SignatureDB
	TrafficDB
	SPInfoDB
	OffChainAuthKeyDB
}
