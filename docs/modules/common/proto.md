# Proto Definition

GfSp framework uses protobuf to define structured data which is language-neutral, platform-neutral and extensible mechanism for serializing data. This section will display used protobuf definiton in GfSp code.

## GfSpTask Proto

Tasks in GfSp uses proto to describe themselves.

```proto
message GfSpTask {
  string address = 1;
  int64 create_time = 2;
  int64 update_time = 3;
  int64 timeout = 4;
  int32 task_priority = 5;
  int64 retry = 6;
  int64 max_retry = 7;
  base.types.gfsperrors.GfSpError err = 8;
}
```

### GfSpCreateBucketApprovalTask Proto

```proto
message GfSpCreateBucketApprovalTask {
  GfSpTask task = 1;
  greenfield.storage.MsgCreateBucket create_bucket_info = 2;
}
```

### GfSpCreateObjectApprovalTask Proto

```proto
message GfSpCreateObjectApprovalTask {
  GfSpTask task = 1;
  greenfield.storage.MsgCreateObject create_object_info = 2;
}
```

### GfSpReplicatePieceApprovalTask Proto

```proto
message GfSpReplicatePieceApprovalTask {
  GfSpTask task = 1;
  greenfield.storage.ObjectInfo object_info = 2;
  greenfield.storage.Params storage_params = 3;
  string ask_sp_operator_address = 4;
  bytes ask_signature = 5;
  string approved_sp_endpoint = 6;
  string approved_sp_operator_address = 7;
  bytes approved_signature = 8;
  string approved_sp_approval_address = 9;
  uint64 expired_height = 10;
}
```

### GfSpUploadObjectTask Proto

```proto
message GfSpUploadObjectTask {
  GfSpTask task = 1;
  greenfield.storage.ObjectInfo object_info = 2;
  greenfield.storage.Params storage_params = 3;
}
```

### GfSpReplicatePieceTask Proto

```proto
message GfSpReplicatePieceTask {
  GfSpTask task = 1;
  greenfield.storage.ObjectInfo object_info = 2;
  greenfield.storage.Params storage_params = 3;
  repeated bytes secondary_signature = 4;
  bool sealed = 5;
}
```

### GfSpReceivePieceTask Proto

```proto
message GfSpReceivePieceTask {
  GfSpTask task = 1;
  greenfield.storage.ObjectInfo object_info = 2;
  greenfield.storage.Params storage_params = 3;
  uint32 replicate_idx = 4;
  int32 piece_idx = 5;
  int64 piece_size = 6;
  bytes piece_checksum = 7;
  bytes signature = 8;
  bool sealed = 9;
}
```

### GfSpSealObjectTask Proto

```proto
message GfSpSealObjectTask {
  GfSpTask task = 1;
  greenfield.storage.ObjectInfo object_info = 2;
  greenfield.storage.Params storage_params = 3;
  repeated bytes secondary_signature = 4;
}
```

### GfSpDownloadObjectTask Proto

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
```

### GfSpDownloadPieceTask Proto

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

### GfSpChallengePieceTask Proto

```proto
message GfSpChallengePieceTask {
  GfSpTask task = 1;
  greenfield.storage.ObjectInfo object_info = 2;
  greenfield.storage.BucketInfo bucket_info = 3;
  greenfield.storage.Params storage_params = 4;
  string user_address = 5;
  uint32 segment_idx = 6;
  int32 redundancy_idx = 7;
  bytes integrity_hash = 8;
  repeated bytes piece_hash = 9;
  int64 piece_data_size = 10;
}
```

### GfSpGCObjectTask Proto

```proto
message GfSpGCObjectTask {
  GfSpTask task = 1;
  uint64 start_block_number = 2;
  uint64 end_block_number = 3;
  uint64 current_block_number = 4;
  uint64 last_deleted_object_id = 5;
  bool running = 6;
}
```

### GfSpGCZombiePieceTask Proto

```proto
message GfSpGCZombiePieceTask {
  GfSpTask task = 1;
  uint64 object_id = 2;
  uint64 delete_count = 3;
  bool running = 4;
}
```

### GfSpGCMetaTask Proto

```proto
message GfSpGCMetaTask {
  GfSpTask task = 1;
  uint64 current_idx = 2;
  uint64 delete_count = 3;
  bool running = 4;
}
```

## Greenfield Proto

Some structured data used in GfSp is deinfed in Greenfield chain repo, we display them as follows.

### MsgCreateBucket Proto

```proto
message MsgCreateBucket {
  option (cosmos.msg.v1.signer) = "creator";

  // creator defines the account address of bucket creator, it is also the bucket owner.
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name defines a globally unique name of bucket
  string bucket_name = 2;
  // visibility means the bucket is private or public. if private, only bucket owner or grantee can read it,
  // otherwise every greenfield user can read it.
  VisibilityType visibility = 3;
  // payment_address defines an account address specified by bucket owner to pay the read fee. Default: creator
  string payment_address = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // primary_sp_address defines the address of primary sp.
  string primary_sp_address = 6 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // primary_sp_approval defines the approval info of the primary SP which indicates that primary sp confirm the user's request.
  Approval primary_sp_approval = 7;
  // charged_read_quota defines the read data that users are charged for, measured in bytes.
  // The available read data for each user is the sum of the free read data provided by SP and
  // the ChargeReadQuota specified here.
  uint64 charged_read_quota = 8;
}
```

### MsgCreateObject Proto

```proto
message MsgCreateObject {
  option (cosmos.msg.v1.signer) = "creator";

  // creator defines the account address of object uploader
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name defines the name of the bucket where the object is stored.
  string bucket_name = 2;
  // object_name defines the name of object
  string object_name = 3;
  // payload_size defines size of the object's payload
  uint64 payload_size = 4;
  // visibility means the object is private or public. if private, only object owner or grantee can access it,
  // otherwise every greenfield user can access it.
  VisibilityType visibility = 5;
  // content_type defines a standard MIME type describing the format of the object.
  string content_type = 6;
  // primary_sp_approval defines the approval info of the primary SP which indicates that primary sp confirm the user's request.
  Approval primary_sp_approval = 7;
  // expect_checksums defines a list of hashes which was generate by redundancy algorithm.
  repeated bytes expect_checksums = 8;
  // redundancy_type can be ec or replica
  RedundancyType redundancy_type = 9;
  // expect_secondarySPs defines a list of StorageProvider address, which is optional
  repeated string expect_secondary_sp_addresses = 10 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### BucketInfo Proto

```proto
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

### ObjectInfo Proto

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
```

### Params Proto

Params defines the parameters for the module.

```proto
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

### GfSpPing Proto

Ping defines the heartbeat request between p2p nodes.

```proto
message GfSpPing {
  // sp_operator_address define sp operator public key
  string sp_operator_address = 1;
  // signature define the signature of sp sign the msg
  bytes signature = 2;
}
```

### GfSpPong Proto

Pong defines the heartbeat response between p2p nodes.

```proto
message GfSpPong {
  // nodes define the
  repeated GfSpNode nodes = 1;
  // sp_operator_address define sp operator public key
  string sp_operator_address = 2;
  // signature define the signature of sp sign the msg
  bytes signature = 3;
}

// Node defines the p2p node info
message GfSpNode {
  // node_id defines the node id
  string node_id = 1;
  // multi_addr define the node multi addr
  repeated string multi_addr = 2;
}
```

### MsgSealObject

```proto
message MsgSealObject {
  option (cosmos.msg.v1.signer) = "operator";

  // operator defines the account address of primary SP
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name defines the name of the bucket where the object is stored.
  string bucket_name = 2;
  // object_name defines the name of object to be sealed.
  string object_name = 3;
  // secondary_sp_addresses defines a list of storage provider which store the redundant data.
  repeated string secondary_sp_addresses = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // secondary_sp_signatures defines the signature of the secondary sp that can
  // acknowledge that the payload data has received and stored.
  repeated bytes secondary_sp_signatures = 5;
}
```

### MsgRejectSealObject Proto

```proto
message MsgRejectSealObject {
  option (cosmos.msg.v1.signer) = "operator";

  // operator defines the account address of the object owner
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name defines the name of the bucket where the object is stored.
  string bucket_name = 2;
  // object_name defines the name of unsealed object to be reject.
  string object_name = 3;
}
```

### MsgDiscontinueBucket

```proto
message MsgDiscontinueBucket {
  option (cosmos.msg.v1.signer) = "operator";

  // operator is the sp who wants to stop serving the bucket.
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name defines the name of the bucket where the object which to be discontinued is stored.
  string bucket_name = 2;
  // the reason for the request.
  string reason = 3;
}
```
