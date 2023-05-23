package taskqueue

import (
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
)

// NewTQueue defines the new func type of TQueue
type NewTQueue = func(name string, cap int) TQueue

// NewTQueueWithLimit defines the new func type of TQueueWithLimit
type NewTQueueWithLimit = func(name string, cap int) TQueueWithLimit

// NewTQueueOnStrategy defines the new func type of TQueueOnStrategy
type NewTQueueOnStrategy = func(name string, cap int) TQueueOnStrategy

// NewTQueueOnStrategyWithLimit the new func type of TQueueOnStrategyWithLimit
type NewTQueueOnStrategyWithLimit = func(name string, cap int) TQueueOnStrategyWithLimit

// TQueue is the interface to task queue. The task queue is mainly used to maintain tasks are running.
// In addition to supporting conventional FIFO operations, task queue also has some customized operations
// for task. For example, Has, PopByKey.
type TQueue interface {
	// Top returns the top task in the queue, if the queue empty, returns nil.
	Top() task.Task
	// Pop pops and returns the top task in queue, if the queue empty, returns nil.
	Pop() task.Task
	// PopByKey pops the task by the task key, if the task does not exist , returns nil.
	PopByKey(task.TKey) task.Task
	// Has returns an indicator whether the task in queue.
	Has(task.TKey) bool
	// Push pushes the task in queue tail, if the queue len greater the capacity, returns error.
	Push(task.Task) error
	// Len returns the length of queue.
	Len() int
	// Cap returns the capacity of queue.
	Cap() int
	// ScanTask scans all tasks, and call the func one by one task.
	ScanTask(func(task.Task))
}

// TQueueWithLimit is the interface task queue that takes resources into account. Only tasks with less
// than required resources can be popped out.
type TQueueWithLimit interface {
	// TopByLimit returns the top task that the LimitEstimate less than the param in the queue.
	TopByLimit(rcmgr.Limit) task.Task
	// PopByLimit pops and returns the top task that the LimitEstimate less than the param in the queue.
	PopByLimit(rcmgr.Limit) task.Task
	// PopByKey pops the task by the task key, if the task does not exist , returns nil.
	PopByKey(task.TKey) task.Task
	// Has returns an indicator whether the task in queue.
	Has(task.TKey) bool
	// Push pushes the task in queue tail, if the queue len greater the capacity, returns error.
	Push(task.Task) error
	// Len returns the length of queue.
	Len() int
	// Cap returns the capacity of queue.
	Cap() int
	// ScanTask scans all tasks, and call the func one by one task.
	ScanTask(func(task.Task))
}

// TQueueOnStrategy is the interface to task queue and the queue supports customize strategies to filter
// task for popping and retiring task.
type TQueueOnStrategy interface {
	TQueue
	TQueueStrategy
}

// TQueueOnStrategyWithLimit is the interface to task queue that takes resources into account, and the
// queue supports customize strategies to filter task for popping and retiring task.
type TQueueOnStrategyWithLimit interface {
	TQueueWithLimit
	TQueueStrategy
}

// TQueueStrategy is the interface to queue customize strategies, it supports filter task for popping and
// retiring task strategies.
type TQueueStrategy interface {
	// SetFilterTaskStrategy sets the callback func to filter task for popping or topping.
	SetFilterTaskStrategy(func(task.Task) bool)
	// SetRetireTaskStrategy sets the callback func to retire task, when the queue is full, it will be
	// called to retire tasks.
	SetRetireTaskStrategy(func(task.Task) bool)
}

func ScanTQueueBySubKey(queue TQueue, subKey task.TKey) ([]task.Task, error) {
	var tasks []task.Task
	scan := func(t task.Task) {
		if strings.Contains(string(t.Key()), string(subKey)) {
			tasks = append(tasks, t)
		}
	}
	queue.ScanTask(scan)
	return tasks, nil
}

func ScanTQueueWithLimitBySubKey(queue TQueueWithLimit, subKey task.TKey) ([]task.Task, error) {
	var tasks []task.Task
	scan := func(t task.Task) {
		if strings.Contains(string(t.Key()), string(subKey)) {
			tasks = append(tasks, t)
		}
	}
	queue.ScanTask(scan)
	return tasks, nil
}

var _ TQueue = (*NilQueue)(nil)
var _ TQueueWithLimit = (*NilQueue)(nil)
var _ TQueueOnStrategy = (*NilQueue)(nil)
var _ TQueueOnStrategyWithLimit = (*NilQueue)(nil)
var _ TQueueStrategy = (*NilQueue)(nil)

type NilQueue struct{}

func (*NilQueue) Top() task.Task                             { return nil }
func (*NilQueue) Pop() task.Task                             { return nil }
func (*NilQueue) PopByKey(task.TKey) task.Task               { return nil }
func (*NilQueue) Has(task.TKey) bool                         { return false }
func (*NilQueue) Push(task.Task) error                       { return nil }
func (*NilQueue) Len() int                                   { return 0 }
func (*NilQueue) Cap() int                                   { return 0 }
func (*NilQueue) ScanTask(func(task.Task))                   {}
func (*NilQueue) TopByLimit(rcmgr.Limit) task.Task           { return nil }
func (*NilQueue) PopByLimit(rcmgr.Limit) task.Task           { return nil }
func (*NilQueue) SetFilterTaskStrategy(func(task.Task) bool) {}
func (*NilQueue) SetRetireTaskStrategy(func(task.Task) bool) {}
