# Task

Task is the interface to the smallest unit of SP background service interaction.

# Concept

## Task Type

There are two main types of task, namely ObjectTask and GCTask. The ObjectTask 
associated with an object, and records the information of different stages of 
the object, includes UploadObjectTask, ReplicatePieceTask, ReceivePieceTask, 
SealObjectTask, DownloadObjectTask. The GCTask is the interface to record the 
information of garbage collection, includes GCObjectTask, GCZombiePieceTask, 
GCStoreTask.

### Object Task

The UploadObjectTask is the interface to record the information for uploading 
object payload data to the primary SP.

The ReplicatePieceTask is the interface to record the information for replicating 
pieces of object payload data to secondary SPs.

The ReceivePieceTask is the interface to record the information for receiving 
pieces of object payload data from primary SP, it exists only in secondary SP.

The SealObjectTask is the interface to  record the information for sealing object 
to the Greenfield chain.

The DownloadObjectTask is the interface to record the information for downloading 
pieces of object payload data.

### GC Task

The GCObjectTask is the interface to record the information for collecting the 
piece store space by deleting object payload data that the object has been deleted 
on the Greenfield chain.

The GCZombiePieceTask is the interface to record the information for collecting 
the piece store space by deleting zombie pieces data that dues to any exception, 
the piece data meta is not on chain but the pieces has been store in piece store.

The GCStoreTask is the interface to record the information for collecting the SP 
meta store space by deleting the expired data.

## Task Priority

Each type of task has a priority, the range of priority is [0, 255], the higher the 
priority, the higher the urgency to be executed, the higher the priority will be 
allocated resources to execute

## Task Priority Level

Task priority is divided into three levels, TLowPriorityLevel, TMediumPriorityLevel, 
THighPriorityLevel. The TLowPriorityLevel default priority range is [0, 85), The 
TMediumPriorityLevel default priority range is [85, 170), The THighPriorityLevel 
default priority range is [170, 255]. When applying for task execution resources from 
ResourceManager, the application is based on the task priority level.
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

