package gfsptqueue

import (
	"sort"
	"sync"
	"time"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/util/maps"
)

var _ taskqueue.TQueueWithLimit = &GfSpTQueueWithLimit{}
var _ taskqueue.TQueueOnStrategyWithLimit = &GfSpTQueueWithLimit{}

type GfSpTQueueWithLimit struct {
	name    string
	current int64
	tasks   map[coretask.TKey]coretask.Task
	cap     int
	mux     sync.RWMutex

	gcFunc     func(task2 coretask.Task) bool
	filterFunc func(task2 coretask.Task) bool
}

func NewGfSpTQueueWithLimit(name string, cap int) taskqueue.TQueueOnStrategyWithLimit {
	return &GfSpTQueueWithLimit{
		name:  name,
		cap:   cap,
		tasks: make(map[coretask.TKey]coretask.Task),
	}
}

// Len returns the length of queue.
func (t *GfSpTQueueWithLimit) Len() int {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return len(t.tasks)
}

// Cap returns the capacity of queue.
func (t *GfSpTQueueWithLimit) Cap() int {
	return t.cap
}

// Has returns an indicator whether the task in queue.
func (t *GfSpTQueueWithLimit) Has(key coretask.TKey) bool {
	// maybe gc task, need RWLock, not RLock
	t.mux.Lock()
	startTime := time.Now()
	defer func() {
		t.mux.Unlock()
		metrics.QueueTime.WithLabelValues(t.name + "-has").Observe(time.Since(startTime).Seconds())
	}()
	return t.has(key)
}

func (t *GfSpTQueueWithLimit) TopByLimit(limit corercmgr.Limit) coretask.Task {
	// maybe trigger gc task, need RWLock not RLock
	t.mux.Lock()
	startTime := time.Now()
	defer func() {
		t.mux.Unlock()
		metrics.QueueTime.WithLabelValues(t.name + "-top_by_limit").Observe(time.Since(startTime).Seconds())
	}()
	return t.topByLimit(limit)
}

// PopByLimit pops and returns the top task that the LimitEstimate less than the param in the queue.
func (t *GfSpTQueueWithLimit) PopByLimit(limit corercmgr.Limit) coretask.Task {
	// maybe trigger gc task, need RWLock not RLock
	t.mux.Lock()
	startTime := time.Now()
	defer func() {
		t.mux.Unlock()
		metrics.QueueTime.WithLabelValues(t.name + "-pop_by_limit").Observe(time.Since(startTime).Seconds())
	}()
	task := t.topByLimit(limit)
	if task != nil {
		t.delete(task)
	}
	return task
}

// PopByKey pops the task by the task key, if the task does not exist , returns nil.
func (t *GfSpTQueueWithLimit) PopByKey(key coretask.TKey) coretask.Task {
	t.mux.Lock()
	startTime := time.Now()
	defer func() {
		t.mux.Unlock()
		metrics.QueueTime.WithLabelValues(t.name + "-pop_by_key").Observe(time.Since(startTime).Seconds())
	}()
	if !t.has(key) {
		return nil
	}
	task, ok := t.tasks[key]
	if !ok {
		return nil
	}
	t.delete(task)
	return task
}

// Push pushes the task in queue tail, if the queue len greater the capacity, returns error.
func (t *GfSpTQueueWithLimit) Push(task coretask.Task) error {
	t.mux.Lock()
	startTime := time.Now()
	defer func() {
		t.mux.Unlock()
		metrics.QueueTime.WithLabelValues(t.name + "-push").Observe(time.Since(startTime).Seconds())
	}()
	if t.has(task.Key()) {
		return ErrTaskRepeated
	}
	if t.exceed() {
		if t.gcFunc == nil {
			log.Warnw("queue exceed", "queue", t.name, "cap", t.cap, "len", len(t.tasks))
			return ErrTaskQueueExceed
		}
		clear := false
		keys := maps.SortKeys(t.tasks)
		for _, key := range keys {
			if t.gcFunc(t.tasks[key]) {
				t.delete(t.tasks[key])
				clear = true
				// only retire one task
				break
			}
		}
		if !clear {
			log.Warnw("queue exceed", "queue", t.name, "cap", t.cap, "len", len(t.tasks))
			return ErrTaskQueueExceed
		}
	}
	t.add(task)
	return nil
}

func (t *GfSpTQueueWithLimit) exceed() bool {
	return len(t.tasks) >= t.cap
}

func (t *GfSpTQueueWithLimit) add(task coretask.Task) {
	defer func() {
		metrics.QueueSizeGauge.WithLabelValues(t.name).Set(float64(len(t.tasks)))
		metrics.QueueCapGauge.WithLabelValues(t.name).Set(float64(t.cap))
	}()
	if task == nil || t.has(task.Key()) {
		return
	}
	t.tasks[task.Key()] = task
}

func (t *GfSpTQueueWithLimit) delete(task coretask.Task) {
	if task == nil || !t.has(task.Key()) {
		return
	}
	defer func() {
		metrics.QueueSizeGauge.WithLabelValues(t.name).Set(float64(len(t.tasks)))
		metrics.QueueCapGauge.WithLabelValues(t.name).Set(float64(t.cap))
		metrics.TaskInQueueTime.WithLabelValues(t.name).Observe(
			time.Now().Sub(time.UnixMilli(task.GetCreateTime())).Seconds())
	}()
	delete(t.tasks, task.Key())
}

func (t *GfSpTQueueWithLimit) has(key coretask.TKey) bool {
	task, ok := t.tasks[key]
	if ok && t.gcFunc != nil {
		if t.gcFunc(task) {
			delete(t.tasks, task.Key())
			return false
		}
	}
	return ok
}

func (t *GfSpTQueueWithLimit) topByLimit(limit corercmgr.Limit) coretask.Task {
	if len(t.tasks) == 0 {
		return nil
	}
	var backupTasks []coretask.Task
	var gcTasks []coretask.Task
	defer func() {
		for _, task := range gcTasks {
			delete(t.tasks, task.Key())
		}
	}()

	for _, task := range t.tasks {
		if t.gcFunc != nil {
			if t.gcFunc(task) {
				gcTasks = append(gcTasks, task)
				continue
			}
		}
		if limit.NotLess(task.EstimateLimit()) {
			if t.filterFunc != nil && !t.filterFunc(task) {
				continue
			}
			backupTasks = append(backupTasks, task)
		}
	}
	if len(backupTasks) == 0 {
		return nil
	}
	sort.Slice(backupTasks, func(i, j int) bool {
		return backupTasks[i].GetCreateTime() < backupTasks[j].GetCreateTime()
	})
	index := sort.Search(len(backupTasks), func(i int) bool { return backupTasks[i].GetCreateTime() > t.current })
	if index == len(backupTasks) {
		index = 0
	}
	if backupTasks[index] != nil {
		t.current = backupTasks[index].GetCreateTime()
	}
	return backupTasks[index]
}

// SetFilterTaskStrategy sets the callback func to filter task for popping or topping.
func (t *GfSpTQueueWithLimit) SetFilterTaskStrategy(filter func(coretask.Task) bool) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.filterFunc = filter
}

// SetRetireTaskStrategy sets the callback func to retire task, when the queue is full, it will be
// called to retire tasks.
func (t *GfSpTQueueWithLimit) SetRetireTaskStrategy(retire func(coretask.Task) bool) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.gcFunc = retire
}

// ScanTask scans all tasks, and call the func one by one task.
func (t *GfSpTQueueWithLimit) ScanTask(scan func(coretask.Task)) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	for _, task := range t.tasks {
		scan(task)
	}
}
