# Manager

Manager is responsible for task scheduling such as dispatching tasks to TaskExecutor module, GC objects, GC zombie piece tasks and so on. Therefore, Manager module is somewhat similar to daemon process to help do some works. The workflow of Manager users can refer [Manager](../workflow/workflow.md#uploader). We currently abstract SP as the GfSp framework, which provides users with customizable capabilities to meet their specific requirements. Manager module provides an abstract interface, which is called `Manager`, as follows:

Manager is an abstract interface to do some internal services management, it is responsible for task scheduling and other management of SP.

```go
type Manager interface {
    Modular
    // DispatchTask dispatches tasks to TaskExecutor module when it asks tasks.
    // It will consider remaining resources when dispatching task.
    DispatchTask(ctx context.Context, limit rcmgr.Limit) (task.Task, error)
    // QueryTasks queries tasks that hold on manager by task sub-key.
    QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error)
    // HandleCreateUploadObjectTask handles the CreateUploadObject request from Uploader, before Uploader handles
    // the users' UploadObject requests, it should send CreateUploadObject requests to Manager ask if it's ok.
    // Through this interface SP implements the global uploading object strategy.
    //
    // For example: control the concurrency of global uploads, avoid repeated uploads, rate control, etc.
    HandleCreateUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error
    // HandleDoneUploadObjectTask handles the result of uploading object payload data to primary, Manager should
    // generate ReplicatePieceTask for TaskExecutor to run.
    HandleDoneUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error
    // HandleReplicatePieceTask handles the result of replicating piece data to secondary SPs,
    // the request comes from TaskExecutor.
    HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask) error
    // HandleSealObjectTask handles the result of sealing object to the greenfield the request comes from TaskExecutor.
    HandleSealObjectTask(ctx context.Context, task task.SealObjectTask) error
    // HandleReceivePieceTask handles the result of receiving piece task, the request comes from Receiver that
    // reports have completed ReceivePieceTask to manager and TaskExecutor that the result of confirming whether
    // the object that is syncer to secondary SP has been sealed.
    HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) error
    // HandleGCObjectTask handles GCObjectTask, the request comes from TaskExecutor.
    HandleGCObjectTask(ctx context.Context, task task.GCObjectTask) error
    // HandleGCZombiePieceTask handles GCZombiePieceTask, the request comes from TaskExecutor.
    HandleGCZombiePieceTask(ctx context.Context, task task.GCZombiePieceTask) error
    // HandleGCMetaTask handles GCMetaTask, the request comes from TaskExecutor.
    HandleGCMetaTask(ctx context.Context, task task.GCMetaTask) error
    // HandleDownloadObjectTask handles DownloadObjectTask, the request comes from Downloader.
    HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) error
    // HandleChallengePieceTask handles ChallengePieceTask, the request comes from Downloader.
    HandleChallengePieceTask(ctx context.Context, task task.ChallengePieceTask) error
}
```

Manager interface inherits [Modular interface](./lifecycle_modular.md#modular-interface), so Uploader module can be managed by lifycycle and resource manager.

In terms of the functions provided by Manager module, there are multiple handling tasks that do a lot of chores. In general, tasks handled by Manager module can be divided into `UploadObjectTask`, `ReplicatePieceTask`, `SealObjectTask`, `ReceivePieceTask`, `GCObjectTask`, `GCZombieTask`, `DownloadObjectTask` and `ChallengePieceTask`. In addition, it also provides `DispatchTask` and `QueryTasks`. The tasks handled by TaskExecutor module is dispatched from Manager module. We can query tasks that we care about by `QueryTasks` method through using subKey.

Manager module only schedules ReplicatePiece and SealObject which belong to background tasks. For front tasks, Manager module just reponds and don't schedule them. As we can see from the second parameter of the methods defined in `Manager` interface, different is splitted into different tasks. They are also defined as an interface.

## DispatchTask

Manager module dispatches tasks to TaskExecutor. When dispatching tasks, Manager module should consider limiting resources to prevent resources from being exhasuted. So the second param of DispatchTask functions is rcmgr.Limit that is an interface to alloc system resources. Limit is an interface that that specifies basic resource limits.

- [Limit](./common/lifecycle_modular.md#limit)

## GfSp Framework Manager Code

Manager module code implementation: [Manager](https://github.com/bnb-chain/greenfield-storage-provider/tree/master/modular/manager)
