# Downloader

Downloader is responsible for downloading object data (including range download) and challenge piece. The workflow of Uploader users can refer [Downloader](../workflow/workflow.md#downloader). We currently abstract SP as the GfSp framework, which provides users with customizable capabilities to meet their specific requirements. Downloader module provides an abstract interface, which is called `Downloader`, as follows:

Downloader is an abstract interface to handle getting object requests from users' account, and getting challenge info requests from other components in the system.

```go
type Downloader interface {
    Modular
    // PreDownloadObject prepares to handle DownloadObject, it can do some checks
    // such as checking for duplicates, if limitation of SP has been reached, etc.
    PreDownloadObject(ctx context.Context, task task.DownloadObjectTask) error
    // HandleDownloadObjectTask handles the DownloadObject and get data from piece store.
    HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) ([]byte, error)
    // PostDownloadObject is called after HandleDownloadObjectTask, it can recycle
    // resources, make statistics and do some other operations..
    PostDownloadObject(ctx context.Context, task task.DownloadObjectTask)

    // PreDownloadPiece prepares to handle DownloadPiece, it can do some checks such as check for duplicates,
    // if limitation of SP has been reached, etc.
    PreDownloadPiece(ctx context.Context, task task.DownloadPieceTask) error
    // HandleDownloadPieceTask handles the DownloadPiece and get data from piece store.
    HandleDownloadPieceTask(ctx context.Context, task task.DownloadPieceTask) ([]byte, error)
    // PostDownloadPiece is called after HandleDownloadPieceTask, it can recycle
    // resources, make statistics and do some other operations.
    PostDownloadPiece(ctx context.Context, task task.DownloadPieceTask)

    // PreChallengePiece prepares to handle ChallengePiece, it can do some checks
    // such as checking for duplicates, if limitation of SP has been reached, etc.
    PreChallengePiece(ctx context.Context, task task.ChallengePieceTask) error
    // HandleChallengePiece handles ChallengePiece, get piece data from piece store and get integrity hash from db.
    HandleChallengePiece(ctx context.Context, task task.ChallengePieceTask) ([]byte, [][]byte, []byte, error)
    // PostChallengePiece is called after HandleChallengePiece, it can recycle resources, make statistics
    // and do some other operations.
    PostChallengePiece(ctx context.Context, task task.ChallengePieceTask)
    // QueryTasks queries download/challenge tasks that running on downloader by task sub-key.
    QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}
```

Downloader interface inherits [Modular interface](./lifecycle_modular.md#modular-interface), so Downloader module can be managed by lifycycle and resource manager.

In terms of the functions provided by Downloader module, it can be divided into three parts: DownloadObject, DownloadPiece and ChallengePiece. They all have three methods: PreXXX, HanldeXXX and PostXXX. Therefore, if you can rewrite these methods to meet your own requirements.

As we can see from the second parameter of the methods defined in `Downloader` interface, DownloadObject is splitted into `DownloadObjectTask`, DownloadPiece is splitted into `DownloadPieceTask` and ChallengePiece is splitted into `ChallengePieceTask`. They are also defined as an interface.

We can query DownloadObject, DownloadPiece and ChallengePiece tasks that we care about by `QueryTasks` method through using subKey.

## ObjectTask

DownloadObjectTask, DownloadPieceTask and ChallengePieceTask all inherits `ObjectTask` interface. ObjectTask associated with an object and storage params, and records the information of different stages of the object. Considering the change of storage params on the greenfield, the storage params of each object should be determined when it is created, and it should not be queried during the task flow, which is inefficient and error-prone.

```go
type ObjectTask interface {
    Task
    // GetObjectInfo returns the associated object.
    GetObjectInfo() *storagetypes.ObjectInfo
    // SetObjectInfo set the  associated object.
    SetObjectInfo(*storagetypes.ObjectInfo)
    // GetStorageParams returns the storage params.
    GetStorageParams() *storagetypes.Params
    // SetStorageParams sets the storage params.Should try to avoid calling this method, it will change the task base information.
    // For example: it will change resource estimate for UploadObjectTask and so on.
    SetStorageParams(*storagetypes.Params)
}
```

The corresponding protobuf definition is shown below:

```proto
message ObjectInfo {
  string owner = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is the name of the bucket
  string bucket_name = 2;
  // object_name is the name of object
  string object_name = 3;
  // id is the unique identifier of object
  string id = 4 [
    (cosmos_proto.scalar) = "cosmos.Uint",
    (gogoproto.customtype) = "Uint",
    (gogoproto.nullable) = false
  ];
  // payloadSize is the total size of the object payload
  uint64 payload_size = 5;
  // visibility defines the highest permissions for object. When an object is public, everyone can access it.
  VisibilityType visibility = 6;
  // content_type define the format of the object which should be a standard MIME type.
  string content_type = 7;
  // create_at define the block timestamp when the object is created
  int64 create_at = 8;
  // object_status define the upload status of the object.
  ObjectStatus object_status = 9;
  // redundancy_type define the type of the redundancy which can be multi-replication or EC.
  RedundancyType redundancy_type = 10;
  // source_type define the source of the object.
  SourceType source_type = 11;
  // checksums define the root hash of the pieces which stored in a SP.
  // add omit tag to omit the field when converting to NFT metadata
  repeated bytes checksums = 12 [(gogoproto.moretags) = "traits:\"omit\""];
  // secondary_sp_addresses define the addresses of secondary_sps
  repeated string secondary_sp_addresses = 13 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}

// Params defines the parameters for the module.
message Params {
  option (gogoproto.goproto_stringer) = false;
  VersionedParams versioned_params = 1 [(gogoproto.nullable) = false];

  // max_payload_size is the maximum size of the payload, default: 2G
  uint64 max_payload_size = 2;
  // relayer fee for the mirror bucket tx
  string mirror_bucket_relayer_fee = 3;
  // relayer fee for the ACK or FAIL_ACK package of the mirror bucket tx
  string mirror_bucket_ack_relayer_fee = 4;
  // relayer fee for the mirror object tx
  string mirror_object_relayer_fee = 5;
  // Relayer fee for the ACK or FAIL_ACK package of the mirror object tx
  string mirror_object_ack_relayer_fee = 6;
  // relayer fee for the mirror object tx
  string mirror_group_relayer_fee = 7;
  // Relayer fee for the ACK or FAIL_ACK package of the mirror object tx
  string mirror_group_ack_relayer_fee = 8;
  // The maximum number of buckets that can be created per account
  uint32 max_buckets_per_account = 9;
  // The window to count the discontinued objects or buckets
  uint64 discontinue_counting_window = 10;
  // The max objects can be requested in a window
  uint64 discontinue_object_max = 11;
  // The max buckets can be requested in a window
  uint64 discontinue_bucket_max = 12;
  // The object will be deleted after the confirm period in seconds
  int64 discontinue_confirm_period = 13;
  // The max delete objects in each end block
  uint64 discontinue_deletion_max = 14;
  // The max number for deleting policy in each end block
  uint64 stale_policy_cleanup_max = 15;
}
```

## DownloadObjectTask

DownloadObjectTask is an abstract interface to record the information for downloading pieces of object payload data. DownloadObjectTask inherits ObjectTask interface. DownloadObjectTask also defines seven methods to help query info or set data. You can overwrite all these methods in your own.

```go
type DownloadObjectTask interface {
    ObjectTask
    // InitDownloadObjectTask inits DownloadObjectTask.
    InitDownloadObjectTask(object *storagetypes.ObjectInfo, bucket *storagetypes.BucketInfo, params *storagetypes.Params,
        priority TPriority, userAddress string, low int64, high int64, timeout int64, retry int64)
    // GetBucketInfo returns the BucketInfo of the download object.
    // It is used to Query and calculate bucket read quota.
    GetBucketInfo() *storagetypes.BucketInfo
    // SetBucketInfo sets the BucketInfo of the download object.
    SetBucketInfo(*storagetypes.BucketInfo)
    // GetUserAddress returns the user account of downloading object.
    // It is used to record the read bucket information.
    GetUserAddress() string
    // SetUserAddress sets the user account of downloading object.
    SetUserAddress(string)
    // GetSize returns the download payload data size, high - low + 1.
    GetSize() int64
    // GetLow returns the start offset of download payload data.
    GetLow() int64
    // GetHigh returns the end offset of download payload data.
    GetHigh() int64
}
```

The corresponding protobuf definition is shown below:

```proto
message GfSpDownloadObjectTask {
  GfSpTask task = 1;
  greenfield.storage.ObjectInfo object_info = 2;
  greenfield.storage.BucketInfo bucket_info = 3;
  greenfield.storage.Params storage_params = 4;
  string user_address = 5;
  int64 low = 6;
  int64 high = 7;
}

message BucketInfo {
  // owner is the account address of bucket creator, it is also the bucket owner.
  string owner = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is a globally unique name of bucket
  string bucket_name = 2;
  // visibility defines the highest permissions for bucket. When a bucket is public, everyone can get storage objects in it.
  VisibilityType visibility = 3;
  // id is the unique identification for bucket.
  string id = 4 [
    (cosmos_proto.scalar) = "cosmos.Uint",
    (gogoproto.customtype) = "Uint",
    (gogoproto.nullable) = false
  ];
  // source_type defines which chain the user should send the bucket management transactions to
  SourceType source_type = 5;
  // create_at define the block timestamp when the bucket created.
  int64 create_at = 6;
  // payment_address is the address of the payment account
  string payment_address = 7 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // primary_sp_address is the address of the primary sp. Objects belongs to this bucket will never
  // leave this SP, unless you explicitly shift them to another SP.
  string primary_sp_address = 8 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // charged_read_quota defines the traffic quota for read in bytes per month.
  // The available read data for each user is the sum of the free read data provided by SP and
  // the ChargeReadQuota specified here.
  uint64 charged_read_quota = 9;
  // billing info of the bucket
  BillingInfo billing_info = 10 [(gogoproto.nullable) = false];
  // bucket_status define the status of the bucket.
  BucketStatus bucket_status = 11;
}
```

## DownloadPieceTask

DownloadPieceTask is an abstract interface to record the information for downloading piece data. DownloadPieceTask inherits ObjectTask interface. DownloadPieceTask also defines ten methods to help query info or set data. You can overwrite all these methods in your own.

```go
type DownloadPieceTask interface {
    ObjectTask
    // InitDownloadPieceTask inits DownloadPieceTask.
    InitDownloadPieceTask(object *storagetypes.ObjectInfo, bucket *storagetypes.BucketInfo, params *storagetypes.Params,
      priority TPriority, enableCheck bool, userAddress string, totalSize uint64, pieceKey string, pieceOffset uint64,
      pieceLength uint64, timeout int64, maxRetry int64)
    // GetBucketInfo returns the BucketInfo of the download object.
    // It is used to Query and calculate bucket read quota.
    GetBucketInfo() *storagetypes.BucketInfo
    // SetBucketInfo sets the BucketInfo of the download object.
    SetBucketInfo(*storagetypes.BucketInfo)
    // GetUserAddress returns the user account of downloading object.
    // It is used to record the read bucket information.
    GetUserAddress() string
    // SetUserAddress sets the user account of downloading object.
    SetUserAddress(string)
    // GetSize returns the download payload data size.
    GetSize() int64
    // GetEnableCheck returns enable_check flag.
    GetEnableCheck() bool
    // GetTotalSize returns total size.
    GetTotalSize() uint64
    // GetPieceKey returns piece key.
    GetPieceKey() string
    // GetPieceOffset returns piece offset.
    GetPieceOffset() uint64
    // GetPieceLength returns piece length.
    GetPieceLength() uint64
}
```

The corresponding protobuf definition is shown below:

```proto
message GfSpDownloadPieceTask {
  GfSpTask task = 1;
  greenfield.storage.ObjectInfo object_info = 2;
  greenfield.storage.BucketInfo bucket_info = 3;
  greenfield.storage.Params storage_params = 4;
  bool enable_check = 5; // check read quota, only in first piece
  string user_address = 6;
  uint64 total_size = 7;
  string piece_key = 8;
  uint64 piece_offset = 9;
  uint64 piece_length = 10;
}
```

## ChallengePieceTask

It is always the first priority of any decentralized storage network to guarantee data integrity and availability. We use data challenge instead of storage proof to get better HA. There will be some data challenges to random pieces on greenfield chain continuously. And the SP, which stores the challenged piece, uses the challenge workflow to response. Each SP splits the object payload data to segments, and store segment data to piece store and store segment checksum to SP DB.

ChallengePieceTask is an abstract interface to record the information for get challenge piece info, the validator get challenge info to confirm whether the sp stores the user's data correctly.

```go
type ChallengePieceTask interface {
    ObjectTask
    // InitChallengePieceTask inits InitChallengePieceTask.
    InitChallengePieceTask(object *storagetypes.ObjectInfo, bucket *storagetypes.BucketInfo, params *storagetypes.Params,
      priority TPriority, userAddress string, replicateIdx int32, segmentIdx uint32, timeout int64, retry int64)
    // GetBucketInfo returns the BucketInfo of challenging piece
    GetBucketInfo() *storagetypes.BucketInfo
    // SetBucketInfo sets the BucketInfo of challenging piece
    SetBucketInfo(*storagetypes.BucketInfo)
    // GetUserAddress returns the user account of challenging object.
    // It is used to record the read bucket information.
    GetUserAddress() string
    // SetUserAddress sets the user account of challenging object.
    SetUserAddress(string)
    // GetSegmentIdx returns the segment index of challenge piece.
    GetSegmentIdx() uint32
    // SetSegmentIdx sets the segment index of challenge piece.
    SetSegmentIdx(uint32)
    // GetRedundancyIdx returns the replicate index of challenge piece.
    GetRedundancyIdx() int32
    // SetRedundancyIdx sets the replicate index of challenge piece.
    SetRedundancyIdx(idx int32)
    // GetIntegrityHash returns the integrity hash of the object.
    GetIntegrityHash() []byte
    // SetIntegrityHash sets the integrity hash of the object.
    SetIntegrityHash([]byte)
    // GetPieceHash returns the hash of  challenge piece.
    GetPieceHash() [][]byte
    // SetPieceHash sets the hash of  challenge piece.
    SetPieceHash([][]byte)
    // GetPieceDataSize returns the data of challenge piece.
    GetPieceDataSize() int64
    // SetPieceDataSize sets the data of challenge piece.
    SetPieceDataSize(int64)
}
```

ChallengePieceTask inherits ObjectTask interface. ChallengePieceTask also defines 15 methods to help query info or set data. You can overwrite all these methods in your own.

## GfSp Framework Downloader Code

Downloader module code implementation: [Downloader](https://github.com/bnb-chain/greenfield-storage-provider/tree/master/modular/downloader)
