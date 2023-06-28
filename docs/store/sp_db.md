# SPDB

SP(Storage Provider Database) store needs to implement SPDB interface. SQL database is used by default.
The following mainly introduces the data schemas corresponding to several core interfaces.

```go
// SPDB contains all the methods required by sql database
type SPDB interface {
    UploadObjectProgressDB
    GCObjectProgressDB
    SignatureDB
    TrafficDB
    SPInfoDB
    OffChainAuthKeyDB
}
```

## UploadObjectProgressDB

UploadObjectProgressDB interface which records upload object related progress(includeing foreground and background) and state. You can overwrite all these methods to meet your requirements.

```go
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
}

// UploadObjectMeta defines the upload object state and related seal info, etc.
type UploadObjectMeta struct {
    ObjectID            uint64
    TaskState           storetypes.TaskState
    SecondaryAddresses  []string
    SecondarySignatures [][]byte
    ErrorDescription    string
}
```

TaskState is defined in protobuf enum:

```proto
enum TaskState {
  TASK_STATE_INIT_UNSPECIFIED = 0;

  TASK_STATE_UPLOAD_OBJECT_DOING = 1;
  TASK_STATE_UPLOAD_OBJECT_DONE = 2;
  TASK_STATE_UPLOAD_OBJECT_ERROR = 3;

  TASK_STATE_ALLOC_SECONDARY_DOING = 4;
  TASK_STATE_ALLOC_SECONDARY_DONE = 5;
  TASK_STATE_ALLOC_SECONDARY_ERROR = 6;

  TASK_STATE_REPLICATE_OBJECT_DOING = 7;
  TASK_STATE_REPLICATE_OBJECT_DONE = 8;
  TASK_STATE_REPLICATE_OBJECT_ERROR = 9;

  TASK_STATE_SIGN_OBJECT_DOING = 10;
  TASK_STATE_SIGN_OBJECT_DONE = 11;
  TASK_STATE_SIGN_OBJECT_ERROR = 12;

  TASK_STATE_SEAL_OBJECT_DOING = 13;
  TASK_STATE_SEAL_OBJECT_DONE = 14;
  TASK_STATE_SEAL_OBJECT_ERROR = 15;
}
```

## GCObjectProgressDB

GCObjectProgressDB interface which records gc object related progress. You can overwrite all these methods to meet your requirements.

```go
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

// GCObjectMeta defines the gc object range progress info.
type GCObjectMeta struct {
    TaskKey             string
    StartBlockHeight    uint64
    EndBlockHeight      uint64
    CurrentBlockHeight  uint64
    LastDeletedObjectID uint64
}
```

## SignatureDB

SignatureDB abstract object integrity interface. You can overwrite all these methods to meet your requirements.

```go
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

// IntegrityMeta defines the payload integrity hash and piece checksum with objectID.
type IntegrityMeta struct {
    ObjectID          uint64
    IntegrityChecksum []byte
    PieceChecksumList [][]byte
    Signature         []byte
}
```

## TrafficDB

TrafficDB defines a series of traffic interfaces. You can overwrite all these methods to meet your requirements.

```go
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

// ReadRecord defines a read request record, will decrease the bucket read quota.
type ReadRecord struct {
    BucketID        uint64
    ObjectID        uint64
    UserAddress     string
    BucketName      string
    ObjectName      string
    ReadSize        uint64
    ReadTimestampUs int64
}

// BucketQuota defines read quota of a bucket.
type BucketQuota struct {
    ReadQuotaSize uint64
}

// BucketTraffic is record traffic by year and month.
type BucketTraffic struct {
    BucketID         uint64
    YearMonth        string // YearMonth is traffic's YearMonth, format "2023-02".
    BucketName       string
    ReadConsumedSize uint64
    ReadQuotaSize    uint64
    ModifyTime       int64
}

// TrafficTimeRange is used by query, return records in [StartTimestampUs, EndTimestampUs).
type TrafficTimeRange struct {
    StartTimestampUs int64
    EndTimestampUs   int64
    LimitNum         int // is unlimited if LimitNum <= 0.
}
```

## SPInfoDB

SPInfoDB defines a series of sp interfaces. You can overwrite all these methods to meet your requirements.

```go
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

// SpAddressType identify address type of SP.
type SpAddressType int32

const (
    OperatorAddressType SpAddressType = iota + 1
    FundingAddressType
    SealAddressType
    ApprovalAddressType
)
```

protobuf definition is as follwos:

```proto
// StorageProvider defines the meta info of storage provider
message StorageProvider {
  // operator_address defines the account address of the storage provider's operator; It also is the unique index key of sp.
  string operator_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // funding_address defines one of the storage provider's accounts which is used to deposit and reward.
  string funding_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // seal_address defines one of the storage provider's accounts which is used to SealObject
  string seal_address = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // approval_address defines one of the storage provider's accounts which is used to approve use's createBucket/createObject request
  string approval_address = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // gc_address defines one of the storage provider's accounts which is used for gc purpose.
  string gc_address = 5 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // total_deposit defines the number of tokens deposited by this storage provider for staking.
  string total_deposit = 6 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
  // status defines the current service status of this storage provider
  Status status = 7;
  // endpoint define the storage provider's network service address
  string endpoint = 8;
  // description defines the description terms for the storage provider.
  Description description = 9 [(gogoproto.nullable) = false];
}
```

## OffChainAuthKeyDB

OffChainAuthKeyDB defines authentication operations in SpDB. You can overwrite all these methods to meet your requirements.

```go
type OffChainAuthKeyDB interface {
    GetAuthKey(userAddress string, domain string) (*OffChainAuthKey, error)
    UpdateAuthKey(userAddress string, domain string, oldNonce int32, newNonce int32, newPublicKey string, newExpiryDate time.Time) error
    InsertAuthKey(newRecord *OffChainAuthKey) error
}

// OffChainAuthKey contains some info about authentication
type OffChainAuthKey struct {
    UserAddress string
    Domain      string

    CurrentNonce     int32
    CurrentPublicKey string
    NextNonce        int32
    ExpiryDate       time.Time

    CreatedTime  time.Time
    ModifiedTime time.Time
}
```
