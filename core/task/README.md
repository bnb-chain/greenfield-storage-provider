# Task

Task is the interface to the smallest unit of SP background service interaction.

# Concept

## Task Type

There are three main types of task, ApprovalTask, ObjectTask and GCTask.

The ApprovalTask is used to record the ask approval information, for user
creating bucket and object need ask primary SP approval if willing serve
the bucket and object, the SP will sign the approval msg if it approved
the msg, and the greenfield will verify the signature of the approval msg
to judge whether SP accepts the bucket and object, for primary replicating
pieces to the secondary SPs need broadcast the approval msg to other SPs,
if they approved the msg, the primary SP will pick up some of them that
approved the msg and replicate the pieces to the these, and they will verify
the signature of the approval msg before receive the pieces. so the
ApprovalTask includes ApprovalCreateBucketTask, ApprovalCreateBucketTask and
ApprovalReplicatePieceTask.

The ObjectTask associated with an object, and records the information of
different stages of the object, includes UploadObjectTask stands upload the
object payload data to the primary SP, ReplicatePieceTask stands replicate
the object pieces to the secondary SPs, ReceivePieceTask only belong to the
secondary SP, records the information of receiving piece and the secondary SP
use it to confirm the object if success to seal on the greenfield, this will
guarantee a return of the secondary SP. SealObjectTask stands seal object on
the greenfield, DownloadObjectTask stands the user download the part or all
object payload data, ChallengePieceTask stands the validator get the challenge
piece info, the validator to challenge the SP if store the user's payload data
correctly by this way.

The GCTask is the interface to record the information of garbage collection,
includes GCObjectTask stands the collection of piece store space by deleting
the payload data that has been deleted on the greenfield, GCZombiePieceTask
stands the collection of piece store space by deleting zombie pieces data that
dues to any exception, the piece data meta is not on the greenfield, GCMetaTask
stands the collection of the SP meta store space by deleting the expired data.

### Approval Task

ApprovalTask is the interface to record the ask approval information, the
approval task timeliness uses the block height, if reached expired height,
the approval invalid.

#### ApprovalCreateBucketTask 
ApprovalCreateBucketTask is the interface to record the ask create bucket
approval information. The user account will create MsgCreateBucket, the SP
should decide whether approved the request based on the MsgCreateBucket.
If so, the sp will SetExpiredHeight and signs the MsgCreateBucket.

#### ApprovalCreateObjectTask
ApprovalCreateObjectTask is the interface to record the ask create object
approval information. The user account will create MsgCreateObject, the SP
should decide whether approved the request based on the MsgCreateObject.
If so, the sp will SetExpiredHeight and signs the MsgCreateObject.

#### ApprovalReplicatePieceTask
ApprovalReplicatePieceTask is the interface to record the ask replicate pieces
to other SPs(as secondary SP for the object). It is initiated by the primary SP
in the replicate pieces phase.  Before the primary SP sends it to other SPs, the
primary SP will sign the task, other SPs will verify it is sent by a legitimate
SP. If other SPs approved the approval, they will SetExpiredHeight and signs the
ApprovalReplicatePieceTask.

### Object Task

The ObjectTask associated with an object and storage params, and records the
information of different stages of the object. Considering the change of storage
params on the greenfield, the storage params of each object should be determined
when it is created, and it should not be queried during the task flow, which is
inefficient and error-prone.

#### UploadObjectTask
The UploadObjectTask is the interface to record the information for uploading 
object payload data to the primary SP.

#### ReplicatePieceTask
The ReplicatePieceTask is the interface to record the information for replicating 
pieces of object payload data to secondary SPs.

#### ReceivePieceTask
The ReceivePieceTask is the interface to record the information for receiving 
pieces of object payload data from primary SP, it exists only in secondary SP.

#### SealObjectTask
The SealObjectTask is the interface to  record the information for sealing object 
to the Greenfield chain.

#### DownloadObjectTask
The DownloadObjectTask is the interface to record the information for downloading 
pieces of object payload data.

#### ChallengePieceTask
ChallengePieceTask is the interface to record the information for get challenge
piece info, the validator get challenge info to confirm whether the sp stores
the user's data correctly.


### GC Task

#### GCObjectTask
The GCObjectTask is the interface to record the information for collecting the 
piece store space by deleting object payload data that the object has been deleted 
on the Greenfield chain.

#### GCZombiePieceTask
The GCZombiePieceTask is the interface to record the information for collecting 
the piece store space by deleting zombie pieces data that dues to any exception, 
the piece data meta is not on chain but the pieces has been store in piece store.

#### GCMetaTask
The GCMetaTask is the interface to record the information for collecting the SP 
meta store space by deleting the expired data.


## Task Priority

Each type of task has a priority, the range of priority is [0, 255], the higher
the priority, the higher the urgency to be executed, the greater the probability
of being executed by priority scheduling.


## Task Priority Level

Task priority is divided into three levels, TLowPriorityLevel, TMediumPriorityLevel,
THighPriorityLevel. The TLowPriorityLevel default priority range is [0, 85), The
TMediumPriorityLevel default priority range is [85, 170), The THighPriorityLevel
default priority range is [170, 256). When allocating for task execution resources
from ResourceManager, the resources are allocated according to task priority level,
but not task priority, because task priority up to 256 levels, the task priority
level make resource management easier.
```go
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

Each task needs to call its InitXXXTask method before use. This method requires passing
in the necessary parameters of each type of task. These parameters will not be changed
in most cases and are necessary, such as task priority, timeout, max retries, and
necessary information for resource estimation.

Any changes to initialization parameters during task execution may cause unpredictable
consequences. For example, changes in parameters that affect resource estimation may
lead to OOM, etc.
