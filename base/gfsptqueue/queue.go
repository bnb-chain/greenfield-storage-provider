package gfsptqueue

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

const (
	TaskQueue = "TaskQueue"
)

var (
	ErrTaskRepeated    = gfsperrors.Register(TaskQueue, http.StatusBadRequest, 970001, "request repeated")
	ErrTaskQueueExceed = gfsperrors.Register(TaskQueue, http.StatusBadRequest, 970002, "request exceed limit")
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
	metrics.QueueCapGauge.WithLabelValues(name).Set(float64(cap))
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
	// maybe gc task, need RWLock, not RLock
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.has(key)
}

// Top returns the top task in the queue, if the queue empty, returns nil.
func (t *GfSpTQueue) Top() coretask.Task {
	var gcTasks []coretask.Task
	// maybe gc task, need RWLock, not RLock
	t.mux.Lock()
	defer func() {
		defer t.mux.Unlock()
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
				continue
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
	// maybe trigger retire task, need RWLock not RLock
	t.mux.Lock()
	defer func() {
		defer t.mux.Unlock()
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
				continue
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
	if !t.has(key) {
		return nil
	}
	idx, ok := t.indexer[key]
	if !ok {
		log.Errorw("[BUG] no task in queue after has check", "queue", t.name,
			"task_key", key)
		return nil
	}
	if idx >= len(t.tasks) {
		log.Errorw("[BUG] index out of bounds", "queue", t.name,
			"len", len(t.tasks), "index", idx)
		t.reset()
		idx, ok = t.indexer[key]
		if !ok {
			return nil
		}
	}
	task := t.tasks[idx]
	if strings.EqualFold(task.Key().String(), key.String()) {
		log.Errorw("[BUG] index mismatch task", "queue", t.name,
			"index_key", key.String(), "task_key", task.Key().String())
		t.reset()
		idx, ok = t.indexer[key]
		if !ok {
			return nil
		}
		task = t.tasks[idx]
	}
	t.delete(task)
	return task
}

// Push pushes the task in queue tail, if the queue len greater the capacity, returns error.
func (t *GfSpTQueue) Push(task coretask.Task) error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.has(task.Key()) {
		return ErrTaskRepeated
	}
	if t.exceed() {
		var gcTasks []coretask.Task
		clear := false
		if t.gcFunc != nil {
			for i := len(t.tasks) - 1; i >= 0; i-- {
				if t.gcFunc(t.tasks[i]) {
					gcTasks = append(gcTasks, t.tasks[i])
					clear = true
					// only retire one task
					break
				}
			}
		}
		if !clear {
			log.Warnw("queue exceed", "queue", t.name, "cap", t.cap, "len", len(t.tasks))
			return ErrTaskQueueExceed
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
	if t.has(task.Key()) {
		return
	}
	t.tasks = append(t.tasks, task)
	t.indexer[task.Key()] = len(t.tasks) - 1
	metrics.QueueSizeGauge.WithLabelValues(t.name).Set(float64(len(t.tasks)))
}

func (t *GfSpTQueue) delete(task coretask.Task) {
	if !t.has(task.Key()) {
		return
	}
	idx, ok := t.indexer[task.Key()]
	if !ok {
		log.Errorw("[BUG] no task in queue after has check", "queue", t.name,
			"task_key", task.Key().String())
		return
	}
	defer func() {
		delete(t.indexer, task.Key())
		metrics.QueueSizeGauge.WithLabelValues(t.name).Set(float64(len(t.tasks)))
	}()
	if idx >= len(t.tasks) {
		log.Errorw("[BUG] index out of bounds", "queue", t.name,
			"len", len(t.tasks), "index", idx)
		t.reset()
		idx, ok = t.indexer[task.Key()]
		if !ok {
			return
		}
	}
	t.tasks = append(t.tasks[0:idx], t.tasks[idx+1:]...)
	metrics.TaskInQueueTimeHistogram.WithLabelValues(t.name).Observe(
		time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
}

func (t *GfSpTQueue) has(key coretask.TKey) bool {
	if len(t.tasks) != len(t.indexer) {
		log.Errorw("[BUG] index length mismatch task length", "queue", t.name,
			"index_length", len(t.indexer), "task_length", len(t.tasks))
		t.reset()
	}
	idx, ok := t.indexer[key]
	if ok {
		if idx >= len(t.tasks) {
			log.Errorw("[BUG] index out of bounds", "queue", t.name,
				"len", len(t.tasks), "index", idx, "key", key)
			t.reset()
		}
		idx, ok = t.indexer[key]
		if !ok {
			return false
		}
		task := t.tasks[idx]
		if !strings.EqualFold(task.Key().String(), key.String()) {
			log.Errorw("[BUG] index mismatch task", "queue", t.name,
				"index_key", key.String(), "task_key", task.Key().String())
			t.reset()
			idx, ok = t.indexer[key]
			if !ok {
				return false
			}
			task = t.tasks[idx]
		}
		if t.gcFunc != nil {
			if t.gcFunc(task) {
				delete(t.indexer, task.Key())
				t.tasks = append(t.tasks[0:idx], t.tasks[idx+1:]...)
				return false
			}
		}
		return true
	}
	return false
}

func (t *GfSpTQueue) reset() {
	t.indexer = make(map[coretask.TKey]int)
	for i, task := range t.tasks {
		t.indexer[task.Key()] = i
	}
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
