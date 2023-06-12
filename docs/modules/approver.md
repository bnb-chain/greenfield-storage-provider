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

Approver interface inherits [Modular interface](./common/lifecycle_modular.md#modular-interface), so Approver module can be managed by lifycycle and resource manager.

In terms of the functions provided by Approver module, it can be divided into two parts: CreateBucketApproval and CreateObjectApproval. Both CreateBucketApproval and CreateObjectApproval have three methods: PreXXX, HanldeXXX and PostXXX. Therefore, if you can rewrite these methods to meet your own requirements.

As we can see from the second parameter of the methods defined in `Approver` interface, bucketApproval is splitted into `ApprovalCreateBucketTask` and objectApproval is splitted into `ApprovalCreateObjectTask`. They are also defined as an interface.

We can query ApprovalCreateBucket and ApprovalCreateObject tasks that we care about by `QueryTasks` method through using subKey.

## ApprovalCreateBucketTask and ApprovalCreateObjectTask

ApprovalTask is used to record approval information for users creating buckets and objects. Primary SP approval is required before serving the bucket and object. If the SP approves the message, it will sign the approval message. The greenfield will verify the signature of the approval message to determine whether the SP accepts the bucket and object. ApprovalTask includes `ApprovalCreateBucketTask` and `ApprovalCreateBucketTask`.

The corresponding interfaces definition is shown below:

- [ApprovalTask](./common/task.md#approvaltask)
- [ApprovalCreateBucketTask](./common/task.md#approvalcreatebuckettask)
- [ApprovalCreateObjectTask](./common/task.md#approvalcreateobjecttask)

ApprovalTask interface inherits [Task interface](./common/task.md#task), it describes what operations does a Task have. You can overwrite all these methods in your own.

The corresponding `protobuf` definition is shown below:

- [GfSpCreateBucketApprovalTask](./common/proto.md#gfspcreatebucketapprovaltask-proto)
- [GfSpCreateObjectApprovalTask](./common/proto.md#gfspcreateobjectapprovaltask-proto)
- [MsgCreateBucket](./common/proto.md#msgcreatebucket-proto)
- [MsgCreateObject](./common/proto.md#msgcreateobject-proto)

## GfSp Framework Approver Code

Approver module code implementation: [Approver](https://github.com/bnb-chain/greenfield-storage-provider/tree/master/modular/approver)
