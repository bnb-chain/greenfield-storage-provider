package manager

import (
	"context"
	"sync"
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
	PriorityQueueManage    = "priority-queue-manager"
	SubQueueUpload         = "sub-queue-upload"
	SubQueueReplicatePiece = "sub-queue-replicate-piece"
	SubQueueSealObject     = "sub-queue-seal-object"
	SubQueueGCObject       = "sub-queue-gc-object"

	GCBlockDistance uint64 = 1800
	GCBlockInternal uint64 = 1000
)

type MPQueue struct {
	pqueue        tqueue.TPriorityQueueWithLimit
	chain         *greenfield.Greenfield
	gcBlockNumber uint64
	gcRunning     atomic.Int32
	mux           sync.RWMutex
}

func NewMPQueue(chain *greenfield.Greenfield, uploadQueueSize, replicateQueueSize, sealQueueSize, gcObjectQueueSize int) *MPQueue {
	mpq := &MPQueue{
		chain: chain,
	}
	uploadStrategy := queue.NewStrategy()
	uploadStrategy.SetCollectionCallback(queue.DefaultGCTasksByTimeout)
	uploadQueue := queue.NewTaskQueue(SubQueueUpload, uploadQueueSize, uploadStrategy, false)

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
		tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskUpload):         uploadQueue,
		tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskReplicatePiece): replicateQueue,
		tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskSealObject):     sealQueue,
		tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskGCObject):       gcObjectQueue,
	}

	mstrategy := queue.NewStrategy()
	mstrategy.SetPickUpCallback(queue.DefaultPickUpTaskByPriority)
	pqueue := queue.NewTaskPriorityQueue(PriorityQueueManage, subQueues, mstrategy, true)
	mpq.pqueue = pqueue
	return mpq
}

func (q *MPQueue) HasTask(key tqueue.TKey) bool {
	return q.pqueue.Has(key)
}

func (q *MPQueue) PopTask(key tqueue.TKey) tqueue.Task {
	return q.pqueue.PopByKey(key)
}

func (q *MPQueue) PopTaskByLimit(limit rcmgr.Limit) tqueue.Task {
	return q.pqueue.PopByLimit(limit)
}

func (q *MPQueue) PushTask(task tqueue.Task) error {
	return q.pqueue.Push(task)
}

func (q *MPQueue) PopPushTask(task tqueue.Task) error {
	return q.pqueue.PopPush(task)
}

func (q *MPQueue) GCMQueueTask() {
	if q.gcRunning.Swap(1) == 1 {
		log.Debugw("manager priority queue gc is running")
		return
	}
	defer q.gcRunning.Swap(0)
	q.pqueue.RunCollection()
}

func (q *MPQueue) GetUploadingTasksCount() int {
	return q.GetUploadPrimaryTasksCount() + q.GetReplicatePieceTasksCount() + q.GetSealObjectTasksCount()
}

func (q *MPQueue) GetUploadPrimaryTasksCount() int {
	prio := tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskUpload)
	return q.pqueue.SubQueueLen(prio)
}

func (q *MPQueue) GetReplicatePieceTasksCount() int {
	prio := tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskReplicatePiece)
	return q.pqueue.SubQueueLen(prio)
}

func (q *MPQueue) GetSealObjectTasksCount() int {
	prio := tqueuetypes.GetTaskPriorityMap().GetPriority(tqueue.TypeTaskSealObject)
	return q.pqueue.SubQueueLen(prio)
}

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
		for atomic.LoadUint64(&q.gcBlockNumber)+GCBlockInternal <= height-GCBlockDistance {
			task, err := tqueuetypes.NewGCObjectTask(atomic.LoadUint64(&q.gcBlockNumber)+1,
				atomic.LoadUint64(&q.gcBlockNumber)+GCBlockInternal)
			if err != nil {
				break
			}
			err = queue.Push(task)
			if err != nil {
				break
			}
			atomic.AddUint64(&q.gcBlockNumber, GCBlockInternal)
		}
	}
}
