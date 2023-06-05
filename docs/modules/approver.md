# Approver

Approver module is used to handle approval requests including `CreateBucketApproval` and `CreateObjectApproval`. The workflow of Approver users can refer [GetApproval](../workflow/workflow.md#get-approval). We currently abstract SP as the GfSp framework, which provides users with customizable capabilities to meet their specific requirements. Approver module provides an abstract interface, which is called `Approver`, as follows:

Approver is an abstract interface to handle ask approval requests.

```go
type Approver interface {
    Modular
    // PreCreateBucketApproval prepares to handle CreateBucketApproval, it can do some checks such as checking for duplicates, if limitation of SP has been reached, etc.
    PreCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask) error
    // HandleCreateBucketApprovalTask handles the CreateBucketApproval, it can set expired height, sign the MsgCreateBucket and so on.
    HandleCreateBucketApprovalTask(ctx context.Context, task task.ApprovalCreateBucketTask) (bool, error)
    // PostCreateBucketApproval is called after HandleCreateBucketApprovalTask, it can recycle resources, make statistics and do some other operations.
    PostCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask)

    // PreCreateObjectApproval prepares to handle CreateObjectApproval, it can do some checks such as check for duplicates, if limitation of SP has been reached, etc.
    PreCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask) error
    // HandleCreateObjectApprovalTask handles the CreateObjectApproval, it can set expired height, sign the MsgCreateObject and so on.
    HandleCreateObjectApprovalTask(ctx context.Context, task task.ApprovalCreateObjectTask) (bool, error)
    // PostCreateObjectApproval is called after HandleCreateObjectApprovalTask, it can recycle resources, make statistics and do some other operations.
    PostCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask)
    // QueryTasks queries tasks that running on approver by task sub-key.
    QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
}
```

Approver interface inherits [Modular interface](./lifecycle_modular.md#modular-interface), so Approver module can be managed by lifycycle and resource manager.

In terms of the functions provided by Approver module, it can be divided into two parts: CreateBucketApproval and CreateObjectApproval. Both CreateBucketApproval and CreateObjectApproval have three methods: PreXXX, HanldeXXX and PostXXX. Therefore, if you can rewrite these methods to meet your own requirements.

As we can see from the second parameter of the methods defined in `Approver` interface, bucketApproval is splitted into `ApprovalCreateBucketTask` and objectApproval is splitted into `ApprovalCreateObjectTask`. They are also defined as an interface.

We can query ApprovalCreateBucket and ApprovalCreateObject tasks that we care about by `QueryTasks` method through using subKey.

## ApprovalCreateBucketTask and ApprovalCreateObjectTask

ApprovalTask is used to record approval information for users creating buckets and objects. Primary SP approval is required before serving the bucket and object. If the SP approves the message, it will sign the approval message. The greenfield will verify the signature of the approval message to determine whether the SP accepts the bucket and object. ApprovalTask includes `ApprovalCreateBucketTask` and `ApprovalCreateBucketTask`.

```go
// ApprovalTask is an abstract interface to record the ask approval information.
// ApprovalTask uses block height to verify whether the approval is expired. If reached expired height, the approval invalid.
type ApprovalTask interface {
    Task
    // GetExpiredHeight returns the expired height of the approval.
    GetExpiredHeight() uint64
    // SetExpiredHeight sets the expired height of the approval, when SP
    // approved the approval, it should set the expired height to stands
    // the approval timeliness. This is one of the ways SP prevents being
    // attacked.
    SetExpiredHeight(uint64)
}

// ApprovalCreateBucketTask is an abstract interface to record the ask create bucket approval information. 
// The user account will create MsgCreateBucket, the SP should decide whether approved the request based on the MsgCreateBucket.
// If so, the sp will SetExpiredHeight and signs the MsgCreateBucket.
type ApprovalCreateBucketTask interface {
    ApprovalTask
    // InitApprovalCreateBucketTask inits the ApprovalCreateBucketTask by MsgCreateBucket and task priority. the SP only fill the MsgCreateBucket's
    // PrimarySpApproval field, can not change other fields.
    InitApprovalCreateBucketTask(*storagetypes.MsgCreateBucket, TPriority)
    // GetCreateBucketInfo returns the user's MsgCreateBucket.
    GetCreateBucketInfo() *storagetypes.MsgCreateBucket
    // SetCreateBucketInfo sets the MsgCreateBucket. Should try to avoid calling this method, it will change the approval information.
    SetCreateBucketInfo(*storagetypes.MsgCreateBucket)
}

// ApprovalCreateObjectTask is an abstract interface to record the ask create object approval information. 
// The user account will create MsgCreateObject, SP should decide whether approved the request based on the MsgCreateObject.
// If so, SP will SetExpiredHeight and signs the MsgCreateObject.
type ApprovalCreateObjectTask interface {
    ApprovalTask
    // InitApprovalCreateObjectTask inits the ApprovalCreateObjectTask by MsgCreateObject and task priority. SP only
    // fill the MsgCreateObject's PrimarySpApproval field, can not change other fields.
    InitApprovalCreateObjectTask(*storagetypes.MsgCreateObject, TPriority)
    // GetCreateObjectInfo returns the user's MsgCreateObject.
    GetCreateObjectInfo() *storagetypes.MsgCreateObject
    // SetCreateObjectInfo sets the MsgCreateObject. Should try to avoid calling this method, it will change the approval information.
    SetCreateObjectInfo(*storagetypes.MsgCreateObject)
}
```

ApprovalTask interface inherits [Task interface](./lifecycle_modular.md#task-interface), it describes what operations does a Task have. The corresponding protobuf definition is shown below:

```proto
message GfSpCreateBucketApprovalTask {
  GfSpTask task = 1;
  greenfield.storage.MsgCreateBucket create_bucket_info = 2;
}

message GfSpCreateObjectApprovalTask {
  GfSpTask task = 1;
  greenfield.storage.MsgCreateObject create_object_info = 2;
}

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

## GfSp Framework Approver Code

Approver module code implementation: [Approver](https://github.com/bnb-chain/greenfield-storage-provider/tree/master/modular/approver)
