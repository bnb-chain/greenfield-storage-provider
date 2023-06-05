# Common Abstract Interface

Every service implements Lifecycle and Modular interface. Therefore, we can So we can do a unified lifecycle and resource management through GfSp framework.

## Lifecycle Interface

Lifecycle interface manages the lifecycle of a service and tracks its state changes. It also listens for signals from the process to ensure a graceful shutdown.

Service is an interface for Lifecycle to manage. The component that plans to use Lifecycle needs to implement the following interface:

```go
// Service provides abstract methods to control the lifecycle of a service
// Every service must implement Service interface.
type Service interface {
    // Name describe service name
    Name() string
    // Start a service, this method should be used in non-block form
    Start(ctx context.Context) error
    // Stop a service, this method should be used in non-block form
    Stop(ctx context.Context) error
}

// Lifecycle is an interface to describe how service is managed.
// The Lifecycle tracks the Service lifecycle, listens for signals from
// the process to ensure a graceful shutdown.
//
// All managed services must firstly call RegisterServices to register with Lifecycle.
type Lifecycle interface {
    // RegisterServices registers service to ServiceLifecycle for managing.
    RegisterServices(modular ...Service)
    // StartServices starts all registered services by calling Service.Start method.
    StartServices(ctx context.Context) Lifecycle
    // StopServices stops all registered services by calling Service.Stop method.
    StopServices(ctx context.Context)
    // Signals listens the system signals for gracefully stop the registered services.
    Signals(sigs ...os.Signal) Lifecycle
    // Wait waits the signal for stopping the ServiceLifecycle, before stopping
    // the ServiceLifecycle will call StopServices stops all registered services.
    Wait(ctx context.Context)
}
```

- [Lifecycle Code Snippet](https://github.com/bnb-chain/greenfield-storage-provider/blob/master/core/lifecycle/lifecycle.go)

## Modular Interface

```go
// Modular is a common interface for submodules that are scheduled by the GfSp framework.
// It inherits lifecycle.Service interface, which is used to manage lifecycle of services. Additionally, Modular is managed
// by ResourceManager, which allows the GfSp framework to reserve and release resources from the Modular resource pool.
type Modular interface {
    lifecycle.Service
    // ReserveResource reserves the resources from Modular resources pool.
    ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error)
    // ReleaseResource releases the resources to Modular resources pool.
    ReleaseResource(ctx context.Context, scope rcmgr.ResourceScopeSpan)
}
```

- [Modular Code Snippet](https://github.com/bnb-chain/greenfield-storage-provider/blob/master/core/module/modular.go)

## Task Interface

All data and requests in SP is splitted into task that is the smallest unit of SP service how to interact. There are three main types of task: ApprovalTask, ObjectTask and GCTask.

```go
type Task interface {
    // Key returns the uniquely identify of the task. It is recommended that each task
    // has its own prefix. In addition, it should also include the information of the
    // task's own identity.
    // For example:
    // 1. ApprovalTask maybe includes the bucket name and object name,
    // 2. ObjectTask maybe includes the object ID,
    // 3. GCTask maybe includes the timestamp.
    Key() TKey
    // Type returns the type of the task. A task has a unique type, such as
    // TypeTaskCreateBucketApproval, TypeTaskUpload etc. has the only one TType
    // definition.
    Type() TType
    // GetAddress returns the task runner address. there is only one runner at the
    // same time, which will assist in quickly locating the running node of the task.
    GetAddress() string
    // SetAddress sets the runner address to the task.
    SetAddress(string)
    // GetCreateTime returns the creation time of the task. The creation time used to
    // judge task execution time.
    GetCreateTime() int64
    // SetCreateTime sets the creation time of the tas.
    SetCreateTime(int64)
    // GetUpdateTime returns the last updated time of the task. The updated time used
    // to determine whether the task is expired with the timeout.
    GetUpdateTime() int64
    // SetUpdateTime sets last updated time of the task. Any changes in task information
    // requires to set the update time.
    SetUpdateTime(int64)
    // GetTimeout returns the timeout of the task, the timeout is a duration, if update
    // time adds timeout lesser now stands the task is expired.
    GetTimeout() int64
    // SetTimeout sets timeout duration of the task.
    SetTimeout(int64)
    // ExceedTimeout returns an indicator whether timeout, if update time adds timeout
    // lesser now returns true, otherwise returns false.
    ExceedTimeout() bool
    // GetMaxRetry returns the max retry times of the task. Each type of task has a
    // fixed max retry times.
    GetMaxRetry() int64
    // SetMaxRetry sets the max retry times of the task.
    SetMaxRetry(int64)
    // GetRetry returns the retry counter of the task.
    GetRetry() int64
    // SetRetry sets the retry counter of the task.
    SetRetry(int)
    // IncRetry increases the retry counter of the task. Each task has the max retry
    // times, if retry counter exceed the max retry, the task should be canceled.
    IncRetry()
    // ExceedRetry returns an indicator whether retry counter greater that max retry.
    ExceedRetry() bool
    // Expired returns an indicator whether ExceedTimeout and ExceedRetry.
    Expired() bool
    // GetPriority returns the priority of the task. Each type of task has a fixed
    // priority. The higher the priority, the higher the urgency of the task, and
    // it will be executed first.
    GetPriority() TPriority
    // SetPriority sets the priority of the task. In most cases, the priority of the
    // task does not need to be set, because the priority of the task corresponds to
    // the task type one by one. Once the task type is determined, the priority is
    // determined. But some scenarios need to dynamically adjust the priority of the
    // task type, then this interface is needed.
    SetPriority(TPriority)
    // EstimateLimit returns estimated resource will be consumed. It is used for
    // application resources to the rcmgr and decide whether it can be executed
    // immediately.
    EstimateLimit() rcmgr.Limit
    // Info returns the task detail info for log and debug.
    Info() string
    // Error returns the task error. if the task is normal, returns nil.
    Error() error
    // SetError sets the error to task. Any errors that occur during task execution
    // will be logged through the SetError method.
    SetError(error)
}
```

The corresponding protobuf definition is shown below:

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

message GfSpError {
  string code_space = 1;
  int32 http_status_code = 2;
  int32 inner_code = 3;
  string description = 4;
}
```
