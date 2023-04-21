package manager

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue/queue"
	tqueuetypes "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue/types"
)

const (
	// PriorityQueueManager defines the manager priority queue name
	PriorityQueueManager = "priority-queue-manager"
	// SubQueueUploadObject defines the upload object sub queue name
	SubQueueUploadObject = "sub-queue-upload-object"
	// SubQueueReplicatePiece defines the replicate piece sub queue name
	SubQueueReplicatePiece = "sub-queue-replicate-piece"
	// SubQueueSealObject defines the seal object sub queue name
	SubQueueSealObject = "sub-queue-seal-object"
	// SubQueueGCObject defines the gc object sub queue name
	SubQueueGCObject = "sub-queue-gc-object"
	// GCBlockDistance defines the distance of current height for gc
	GCBlockDistance uint64 = 1800
	// GCBlockInterval defines the internal of blocks for gc object
	GCBlockInterval uint64 = 1000
)

// MPQueue defines the manager priority queue, include upload, replicate, seal and gc task
// sub queue, replicate, seal and gc supports scheduling to dispatch for task node execution.
type MPQueue struct {
	pqueue        tqueue.TPriorityQueueWithLimit
	chain         *greenfield.Greenfield
	gcBlockNumber uint64
	gcRunning     atomic.Value
}

// NewMPQueue returns an instance of MPQueue
func NewMPQueue(chain *greenfield.Greenfield, uploadQueueSize, replicateQueueSize, sealQueueSize, gcObjectQueueSize int) *MPQueue {
	mpq := &MPQueue{
		chain: chain,
	}
	uploadStrategy := queue.NewStrategy()
	uploadStrategy.SetCollectionCallback(queue.DefaultGCTasksByTimeout)
	uploadQueue := queue.NewTaskQueue(SubQueueUploadObject, uploadQueueSize, uploadStrategy, false)

	replicateStrategy := queue.NewStrategy()
	replicateStrategy.SetCollectionCallback(queue.DefaultGCTasksByRetry)
	replicateStrategy.SetPickUpFilterCallback(queue.DefaultPickUpFilterTaskByRetry)
	replicateQueue := queue.NewTaskQueue(SubQueueReplicatePiece, replicateQueueSize, replicateStrategy, true)

	sealStrategy := queue.NewStrategy()
	sealStrategy.SetCollectionCallback(queue.DefaultGCTasksByRetry)
	replicateStrategy.SetPickUpFilterCallback(queue.DefaultPickUpFilterTaskByRetry)
	sealQueue := queue.NewTaskQueue(SubQueueSealObject, sealQueueSize, sealStrategy, true)

	gcObjectStrategy := queue.NewStrategy()
	gcObjectStrategy.SetCollectionCallback(mpq.GCObjectQueueCallBack)
	gcObjectQueue := queue.NewTaskQueue(SubQueueGCObject, gcObjectQueueSize, gcObjectStrategy, true)

	subQueues := map[tqueue.TPriority]tqueue.TQueueWithLimit{
		tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskUploadObject):   uploadQueue,
		tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskReplicatePiece): replicateQueue,
		tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskSealObject):     sealQueue,
		tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskGCObject):       gcObjectQueue,
	}

	mstrategy := queue.NewStrategy()
	mstrategy.SetPickUpCallback(queue.DefaultPickUpTaskByPriority)
	mpq.pqueue = queue.NewTaskPriorityQueue(PriorityQueueManager, subQueues, mstrategy, true)
	return mpq
}

// HasTask returns an indicator whether exists the task in priority queue.
func (q *MPQueue) HasTask(key tqueue.TKey) bool {
	return q.pqueue.Has(key)
}

// PopTask pops the task from priority queue by task key.
func (q *MPQueue) PopTask(key tqueue.TKey) tqueue.Task {
	return q.pqueue.PopByKey(key)
}

// PopTaskByLimit pops the task from priority queue by resource limits.
func (q *MPQueue) PopTaskByLimit(limit rcmgr.Limit) tqueue.Task {
	return q.pqueue.PopByLimit(limit)
}

// PushTask pushes the task to priority queue.
func (q *MPQueue) PushTask(task tqueue.Task) error {
	return q.pqueue.Push(task)
}

// PopPushTask pushes the task to priority queue, overrides the exists one.
func (q *MPQueue) PopPushTask(task tqueue.Task) error {
	return q.pqueue.PopPush(task)
}

// GCMPQueueTask runs priority queue gc.
func (q *MPQueue) GCMPQueueTask() {
	if q.gcRunning.Swap(true) == true {
		log.Debugw("manager priority queue gc is running")
		return
	}
	defer q.gcRunning.Swap(false)
	q.pqueue.RunCollection()
}

// GetUploadingTasksCount returns the number task of uploading object.
func (q *MPQueue) GetUploadingTasksCount() int {
	return q.GetUploadPrimaryTasksCount() + q.GetReplicatePieceTasksCount() + q.GetSealObjectTasksCount()
}

// GetUploadPrimaryTasksCount returns the number task of uploading object to primary SP.
func (q *MPQueue) GetUploadPrimaryTasksCount() int {
	prio := tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskUploadObject)
	return q.pqueue.SubQueueLen(prio)
}

// GetReplicatePieceTasksCount returns the number task of replicating pieces to secondary SP.
func (q *MPQueue) GetReplicatePieceTasksCount() int {
	prio := tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskReplicatePiece)
	return q.pqueue.SubQueueLen(prio)
}

// GetSealObjectTasksCount returns the number task of sealing object.
func (q *MPQueue) GetSealObjectTasksCount() int {
	prio := tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskSealObject)
	return q.pqueue.SubQueueLen(prio)
}

// GCObjectQueueCallBack implements the gc object sub queue.
func (q *MPQueue) GCObjectQueueCallBack(queue tqueue.TQueueOnStrategy, keys []tqueue.TKey) {
	for _, key := range keys {
		if queue.Expired(key) {
			task := queue.PopByKey(key)
			if task == nil {
				continue
			}
			task.SetCreateTime(time.Now().Unix())
			queue.Push(task)
		}
	}
	if queue.Len() < queue.Cap() {
		height, err := q.chain.GetCurrentHeight(context.Background())
		if err != nil {
			return
		}
		for atomic.LoadUint64(&q.gcBlockNumber)+GCBlockInterval <= height-GCBlockDistance {
			task, err := tqueuetypes.NewGCObjectTask(atomic.LoadUint64(&q.gcBlockNumber)+1,
				atomic.LoadUint64(&q.gcBlockNumber)+GCBlockInterval)
			if err != nil {
				break
			}
			err = queue.Push(task)
			if err != nil {
				break
			}
			atomic.AddUint64(&q.gcBlockNumber, GCBlockInterval)
		}
	}
}
