package queue

import (
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
)

var _ tqueue.TQueue = &TaskQueue{}
var _ tqueue.TQueueWithLimit = &TaskQueue{}

type TaskQueue struct {
	name         string
	tasks        []tqueue.Task
	indexer      map[tqueue.TKey]int
	cap          int
	strategy     tqueue.TQueueStrategy
	supportLimit bool
	mux          sync.RWMutex
}

func NewTaskQueue(name string, cap int, strategy tqueue.TQueueStrategy, limit bool) tqueue.TQueueWithLimit {
	queue := &TaskQueue{
		name:         name,
		tasks:        make([]tqueue.Task, 0),
		indexer:      make(map[tqueue.TKey]int),
		cap:          cap,
		strategy:     strategy,
		supportLimit: limit,
	}
	return queue
}

func (q *TaskQueue) Len() int {
	q.mux.RLock()
	defer q.mux.RUnlock()
	return len(q.tasks)
}

func (q *TaskQueue) Cap() int {
	q.mux.RLock()
	defer q.mux.RUnlock()
	return q.cap
}

func (q *TaskQueue) Top() tqueue.Task {
	q.mux.RLock()
	defer q.mux.RUnlock()
	if len(q.tasks) == 0 {
		return nil
	}
	return q.tasks[len(q.tasks)-1]
}

func (q *TaskQueue) Pop() tqueue.Task {
	q.mux.Lock()
	defer q.mux.Unlock()
	if len(q.tasks) == 0 {
		return nil
	}
	task := q.tasks[len(q.tasks)-1]
	q.tasks = q.tasks[0 : len(q.tasks)-1]
	delete(q.indexer, task.Key())
	return task
}

func (q *TaskQueue) Push(task tqueue.Task) error {
	q.mux.Lock()
	defer q.mux.Unlock()
	if _, ok := q.indexer[task.Key()]; ok {
		log.Warnf("[%s] task has repeated", task.Key())
		return ErrTaskRepeated
	}
	if len(q.tasks)+1 > q.cap {
		log.Warnf("[%s] queue has exceeded, cap [%d]", q.name, q.cap)
		return ErrQueueExceeded
	}
	q.tasks = append(q.tasks, task)
	q.indexer[task.Key()] = len(q.tasks) - 1
	return nil
}

func (q *TaskQueue) PopPush(task tqueue.Task) error {
	q.mux.Lock()
	defer q.mux.Unlock()
	idx, ok := q.indexer[task.Key()]
	if !ok {
		q.tasks = append(q.tasks, task)
		q.indexer[task.Key()] = len(q.tasks) - 1
		return nil
	}
	if idx >= len(q.tasks) {
		log.Errorf("BUG: [%s] indexer out of bounds, idx [%d], len [%d]", q.name, idx, len(q.tasks))
		return nil
	}
	q.tasks[idx] = task
	return nil
}

func (q *TaskQueue) Has(key tqueue.TKey) bool {
	q.mux.RLock()
	defer q.mux.RUnlock()
	_, ok := q.indexer[key]
	return ok
}

func (q *TaskQueue) PopByKey(key tqueue.TKey) tqueue.Task {
	q.mux.Lock()
	defer q.mux.Unlock()
	if _, ok := q.indexer[key]; !ok {
		return nil
	}
	idx := q.indexer[key]
	if idx >= len(q.tasks) {
		log.Errorf("BUG: [%s] indexer out of bounds, idx [%d], len [%d]", q.name, idx, len(q.tasks))
		return nil
	}
	task := q.tasks[idx]
	q.tasks = append(q.tasks[0:idx], q.tasks[idx+1:]...)
	delete(q.indexer, key)
	return task
}

func (q *TaskQueue) PopByLimit(limit rcmgr.Limit) tqueue.Task {
	q.mux.Lock()
	defer q.mux.Unlock()
	index := -1
	for idx := len(q.tasks) - 1; idx >= 0; idx-- {
		task := q.tasks[idx]
		if !q.strategy.RunPickUpFilterStrategy(task) {
			continue
		}
		if task.LimitEstimate().Greater(limit) {
			index = idx
			break
		}
	}
	if index == -1 {
		log.Debugw("not found match task", "queue", q.name, "limit", limit.String())
		return nil
	}
	task := q.tasks[index]
	q.tasks = append(q.tasks[0:index], q.tasks[index+1:]...)
	delete(q.indexer, task.Key())
	log.Debugw("found match task", "queue", q.name, "limit", limit.String(),
		"task_limit", task.LimitEstimate().String())
	return task
}

func (q *TaskQueue) IsActiveTask(key tqueue.TKey) bool {
	q.mux.RLock()
	defer q.mux.RUnlock()
	if _, ok := q.indexer[key]; !ok {
		return false
	}
	idx := q.indexer[key]
	if idx >= len(q.tasks) {
		log.Errorf("BUG: [%s] indexer out of bounds, idx [%d], len [%d]", q.name, idx, len(q.tasks))
		return false
	}
	task := q.tasks[idx]
	return task.RetryExceed() && task.Expired()
}

func (q *TaskQueue) Expired(key tqueue.TKey) bool {
	q.mux.RLock()
	defer q.mux.RUnlock()
	if _, ok := q.indexer[key]; !ok {
		return false
	}
	idx := q.indexer[key]
	if idx >= len(q.tasks) {
		log.Errorf("BUG: [%s] indexer out of bounds, idx [%d], len [%d]", q.name, idx, len(q.tasks))
		return false
	}
	task := q.tasks[idx]
	return task.Expired()
}

func (q *TaskQueue) GetSupportPickUpByLimit() bool {
	q.mux.RLock()
	defer q.mux.RUnlock()
	return q.supportLimit
}

func (q *TaskQueue) SetSupportPickUpByLimit(support bool) {
	q.mux.Lock()
	defer q.mux.Unlock()
	q.supportLimit = support
}

func (q *TaskQueue) RunCollection() {
	q.mux.RLock()
	var keys []tqueue.TKey
	for key, _ := range q.indexer {
		keys = append(keys, key)
	}
	q.mux.RUnlock()
	go q.strategy.RunCollectionStrategy(q, keys)
}

func (q *TaskQueue) SetStrategy(strategy tqueue.TQueueStrategy) {
	q.mux.Lock()
	defer q.mux.Unlock()
	q.strategy = strategy
}
