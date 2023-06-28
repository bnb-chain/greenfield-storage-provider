# Task

Task is an abstract interface to describe the smallest unit of SP background service how to interact.

## Task Type

There are three main types of task: ApprovalTask, ObjectTask and GCTask.

ApprovalTask is used to record approval information for users creating buckets and objects. Primary SP approval is
required before serving the bucket and object. If SP approves the message, it will sign the approval message. The
greenfield will verify the signature of the approval message to determine whether the SP accepts the bucket and object.
When primary replicating pieces to secondary SPs, the approval message is broadcast to other SPs. If they approve the
message, the primary SP will select some of them to replicate the pieces to. Before receiving the pieces, the selected SPs
will verify the signature of the approval message. ApprovalTask includes ApprovalCreateBucketTask, ApprovalCreateBucketTask
and ApprovalReplicatePieceTask.

ObjectTask is associated with an object and records information about its different stages. This includes
UploadObjectTask, which uploads the object payload data to the primary SP, ReplicatePieceTask, which replicates the 
object pieces to the secondary SPs, and the ReceivePieceTask, which is exclusive to the secondary SP and records 
information about receiving the piece. The secondary SP uses this information to confirm whether the object was 
successfully sealed on the greenfield, ensuring a return of the secondary SP. SealObjectTask seals the object on Greenfield,
while the DownloadObjectTask allows the user to download part or all of the object payload data. ChallengePieceTask 
provides the validator with challenge piece information, which they can use to challenge the SP if they suspect that 
the user's payload data was not stored correctly.

GCTask is an abstract interface that records information about garbage collection. This includes GCObjectTask, 
which collects piece store space by deleting payload data that has been deleted on the greenfield, GCZombiePieceTask,
which collects piece store space by deleting zombie piece data that resulted from any exception where the piece data 
meta is not on Greenfield chain, and GCMetaTask, which collects the SP meta store space by deleting expired data.

### Approval Task

ApprovalTask is an abstract interface to record the ask approval information, the approval task timeliness uses the block height,
if reached expired height, the approval invalid.

#### ApprovalCreateBucketTask

ApprovalCreateBucketTask is an abstract interface to record the ask create bucket approval information. The user account will
create MsgCreateBucket, the SP should decide whether approved the request based on the MsgCreateBucket. If so, the sp
will SetExpiredHeight and signs the MsgCreateBucket.

#### ApprovalCreateObjectTask

ApprovalCreateObjectTask is an abstract interface to record the ask create object approval information. The user account will
create MsgCreateObject, the SP should decide whether approved the request based on the MsgCreateObject. If so, the sp 
will SetExpiredHeight and signs the MsgCreateObject.

#### ApprovalReplicatePieceTask

ApprovalReplicatePieceTask is an abstract interface to record the ask replicate pieces to other SPs(as secondary SP for the object).
It is initiated by the primary SP in the replicate pieces phase.  Before the primary SP sends it to other SPs, the primary
SP will sign the task, other SPs will verify it is sent by a legitimate SP. If other SPs approved the approval, they will
SetExpiredHeight and signs the ApprovalReplicatePieceTask.

### Object Task

The ObjectTask associated with an object and storage params, and records the information of different stages of the object.
Considering the change of storage params on the greenfield, the storage params of each object should be determined when
it is created, and it should not be queried during the task flow, which is inefficient and error-prone.

#### UploadObjectTask

The UploadObjectTask is an abstract interface to record the information for uploading object payload data to the primary SP.

#### ReplicatePieceTask

The ReplicatePieceTask is an abstract interface to record the information for replicating pieces of object payload data to secondary SPs.

#### ReceivePieceTask

The ReceivePieceTask is an abstract interface to record the information for receiving pieces of object payload data from primary
SP, it exists only in secondary SP.

#### SealObjectTask

The SealObjectTask is an abstract interface to  record the information for sealing object to the Greenfield chain.

#### DownloadObjectTask

The DownloadObjectTask is an abstract interface to record the information for downloading pieces of object payload data.

#### ChallengePieceTask

ChallengePieceTask is an abstract interface to record the information for get challenge piece info, the validator get challenge
info to confirm whether the sp stores the user's data correctly.

### GC Task

#### GCObjectTask

The GCObjectTask is an abstract interface to record the information for collecting the piece store space by deleting object payload
data that the object has been deleted on Greenfield chain.

#### GCZombiePieceTask

The GCZombiePieceTask is an abstract interface to record the information for collecting the piece store space by deleting zombie
pieces data that dues to any exception, the piece data meta is not on chain but the pieces has been store in piece store.

#### GCMetaTask

The GCMetaTask is an abstract interface to record the information for collecting the SP meta store space by deleting the expired data.

## Task Priority

Each type of task has a priority, the range of priority is [0, 255], the higher the priority, the higher the urgency to
be executed, the greater the probability of being executed by priority scheduling.

## Task Priority Level

Task priority is divided into three levels, TLowPriorityLevel, TMediumPriorityLevel, THighPriorityLevel. The TLowPriorityLevel
default priority range is [0, 85), The TMediumPriorityLevel default priority range is [85, 170), The THighPriorityLevel
default priority range is [170, 256). When allocating for task execution resources from ResourceManager, the resources
are allocated according to task priority level, but not task priority, because task priority up to 256 levels, the task priority
level make resource management easier.

```text
Example:
    the resource limit configuration of task execution node :
        [TasksHighPriority: 30, TasksMediumPriority: 20, TasksLowPriority: 2]
    the executor of the task can run 30 high level tasks at the same time that the
        task priority between [170, 255]
    the executor of the task can run 20 medium level tasks at the same time that the
        task priority between [85, 170)
    the executor of the task can run 2 medium level tasks at the same time that the
        task priority < 85
```

## Task Init

Each task needs to call its InitXXXTask method before use. This method requires passing in the necessary parameters of
each type of task. These parameters will not be changed in most cases and are necessary, such as task priority, timeout,
max retries, and necessary information for resource estimation.

Any changes to initialization parameters during task execution may cause unpredictable consequences. For example, changes
in parameters that affect resource estimation may cause OOM, etc.
