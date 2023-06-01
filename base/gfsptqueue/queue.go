package gfsptqueue

import (
	"errors"
	"sync"
	"time"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var _ taskqueue.TQueue = &GfSpTQueue{}
var _ taskqueue.TQueueOnStrategy = &GfSpTQueue{}

type GfSpTQueue struct {
	name    string
	tasks   []coretask.Task
	indexer map[coretask.TKey]int
	cap     int
	mux     sync.RWMutex

	gcFunc     func(task2 coretask.Task) bool
	filterFunc func(task2 coretask.Task) bool
}

func NewGfSpTQueue(name string, cap int) taskqueue.TQueueOnStrategy {
	return &GfSpTQueue{
		name:    name,
		cap:     cap,
		tasks:   make([]coretask.Task, 0),
		indexer: make(map[coretask.TKey]int),
	}
}

// Len returns the length of queue.
func (t *GfSpTQueue) Len() int {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return len(t.tasks)
}

// Cap returns the capacity of queue.
func (t *GfSpTQueue) Cap() int {
	return t.cap
}

// Has returns an indicator whether the task in queue.
func (t *GfSpTQueue) Has(key coretask.TKey) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()
	idx, ok := t.indexer[key]
	if ok {
		if idx >= len(t.tasks) {
			log.Errorw("[BUG] index out of bounds", "queue", t.name,
				"len", len(t.tasks), "index", idx, "key", key)
			return false
		}
		task := t.tasks[idx]
		if t.gcFunc != nil {
			if t.gcFunc(task) {
				t.delete(task)
				return false
			}
		}
	}
	return ok
}

// Top returns the top task in the queue, if the queue empty, returns nil.
func (t *GfSpTQueue) Top() coretask.Task {
	var gcTasks []coretask.Task
	t.mux.RLock()
	defer func() {
		defer t.mux.RUnlock()
		for _, task := range gcTasks {
			t.delete(task)
		}
	}()

	if len(t.tasks) == 0 {
		return nil
	}
	for i := len(t.tasks) - 1; i >= 0; i-- {
		if t.gcFunc != nil {
			if t.gcFunc(t.tasks[i]) {
				gcTasks = append(gcTasks, t.tasks[i])
			}
		}
		if t.filterFunc != nil {
			if t.filterFunc(t.tasks[i]) {
				return t.tasks[i]
			}
		} else {
			return t.tasks[i]
		}
	}
	return nil
}

// Pop pops and returns the top task in queue, if the queue empty, returns nil.
func (t *GfSpTQueue) Pop() coretask.Task {
	var gcTasks []coretask.Task
	var popTask coretask.Task
	t.mux.RLock()
	defer func() {
		defer t.mux.RUnlock()
		for _, task := range gcTasks {
			t.delete(task)
		}
		if popTask != nil {
			t.delete(popTask)
		}
	}()

	if len(t.tasks) == 0 {
		return nil
	}
	for i := len(t.tasks) - 1; i >= 0; i-- {
		if t.gcFunc != nil {
			if t.gcFunc(t.tasks[i]) {
				gcTasks = append(gcTasks, t.tasks[i])
			}
		}
		if t.filterFunc != nil {
			if t.filterFunc(t.tasks[i]) {
				popTask = t.tasks[i]
				return popTask
			}
		} else {
			popTask = t.tasks[i]
			return popTask
		}
	}
	return nil
}

// PopByKey pops the task by the task key, if the task does not exist , returns nil.
func (t *GfSpTQueue) PopByKey(key coretask.TKey) coretask.Task {
	t.mux.Lock()
	defer t.mux.Unlock()
	idx, ok := t.indexer[key]
	if !ok {
		return nil
	}
	if idx >= len(t.tasks) {
		log.Errorw("[BUG] index out of bounds", "queue", t.name,
			"len", len(t.tasks), "index", idx)
		return nil
	}
	task := t.tasks[idx]
	t.delete(task)
	return task
}

// Push pushes the task in queue tail, if the queue len greater the capacity, returns error.
func (t *GfSpTQueue) Push(task coretask.Task) error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if _, ok := t.indexer[task.Key()]; ok {
		log.Warnw("push repeat task", "queue", t.name, "task", task.Key())
		return errors.New("repeated task")
	}
	if t.exceed() {
		var gcTasks []coretask.Task
		clear := false
		if t.gcFunc != nil {
			for i := len(t.tasks) - 1; i >= 0; i-- {
				if t.gcFunc(t.tasks[i]) {
					gcTasks = append(gcTasks, t.tasks[i])
					clear = true
				}
			}
		}
		if !clear {
			log.Warnw("queue exceed", "queue", t.name, "cap", t.cap, "len", len(t.tasks))
			return errors.New("queue exceed")
		} else {
			for _, gcTask := range gcTasks {
				t.delete(gcTask)
			}
		}
	}
	t.add(task)
	return nil
}

func (t *GfSpTQueue) exceed() bool {
	return len(t.tasks) >= t.cap
}

func (t *GfSpTQueue) add(task coretask.Task) {
	t.tasks = append(t.tasks, task)
	t.indexer[task.Key()] = len(t.tasks) - 1
	metrics.QueueSizeGauge.WithLabelValues(t.name).Inc()
	metrics.QueueCapGauge.WithLabelValues(t.name).Set(float64(t.cap))
}

func (t *GfSpTQueue) delete(task coretask.Task) {
	idx, ok := t.indexer[task.Key()]
	if !ok {
		return
	}
	if idx >= len(t.tasks) {
		log.Errorw("[BUG] index out of bounds", "queue", t.name,
			"len", len(t.tasks), "index", idx)
		return
	}
	t.tasks = append(t.tasks[0:idx], t.tasks[idx+1:]...)
	delete(t.indexer, task.Key())
	metrics.QueueSizeGauge.WithLabelValues(t.name).Dec()
	metrics.QueueCapGauge.WithLabelValues(t.name).Set(float64(t.cap))
	metrics.TaskInQueueTimeHistogram.WithLabelValues(t.name).Observe(
		time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
}

// SetFilterTaskStrategy sets the callback func to filter task for popping or topping.
func (t *GfSpTQueue) SetFilterTaskStrategy(filter func(coretask.Task) bool) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.filterFunc = filter
}

// SetRetireTaskStrategy sets the callback func to retire task, when the queue is full, it will be
// called to retire tasks.
func (t *GfSpTQueue) SetRetireTaskStrategy(retire func(coretask.Task) bool) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.gcFunc = retire
}

// ScanTask scans all tasks, and call the func one by one task.
func (t *GfSpTQueue) ScanTask(scan func(coretask.Task)) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	for _, task := range t.tasks {
		scan(task)
	}
}
