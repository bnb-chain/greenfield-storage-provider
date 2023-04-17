package queue

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
)

var _ tqueue.TQueueStrategy = &Strategy{}

type Strategy struct {
	pick       func([]tqueue.Task) tqueue.Task
	gc         func(tqueue.TQueueOnStrategy, []tqueue.TKey)
	pickFilter func(tqueue.Task) bool
	mux        sync.RWMutex
}

func NewStrategy() tqueue.TQueueStrategy {
	return &Strategy{
		pick: nil,
		gc:   nil,
	}
}

func (s *Strategy) SetPickUpCallback(pick func([]tqueue.Task) tqueue.Task) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.pick = pick
}

func (s *Strategy) SetCollectionCallback(gc func(queue tqueue.TQueueOnStrategy, keys []tqueue.TKey)) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.gc = gc
}

func (s *Strategy) RunPickUpStrategy(tasks []tqueue.Task) tqueue.Task {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if s.pick == nil {
		return nil
	}
	return s.pick(tasks)
}

func (s *Strategy) RunCollectionStrategy(queue tqueue.TQueueOnStrategy, keys []tqueue.TKey) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if s.gc == nil {
		return
	}
	s.gc(queue, keys)
}

// SetPickUpFilterCallback sets the callback func for picking up filter task.
func (s *Strategy) SetPickUpFilterCallback(filter func(tqueue.Task) bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	s.pickFilter = filter
}

// RunPickUpFilterStrategy calls the pick up filter callback to pick up filter task.
func (s *Strategy) RunPickUpFilterStrategy(task tqueue.Task) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if s.gc == nil {
		return true
	}
	return s.pickFilter(task)
}

func DefaultPickUpFilterTaskByRetry(task tqueue.Task) bool {
	if task.RetryExceed() {
		return false
	}
	if task.Expired() {
		return true
	}
	return false
}

func DefaultPickUpFilterTaskByTimeout(task tqueue.Task) bool {
	return task.Expired()
}

func DefaultGCTasksByRetry(queue tqueue.TQueueOnStrategy, keys []tqueue.TKey) {
	for _, key := range keys {
		if queue.IsActiveTask(key) {
			queue.PopByKey(key)
		}
	}
}

func DefaultGCTasksByTimeout(queue tqueue.TQueueOnStrategy, keys []tqueue.TKey) {
	for _, key := range keys {
		if queue.Expired(key) {
			queue.PopByKey(key)
		}
	}
}

func DefaultPickUpTaskByPriority(tasks []tqueue.Task) tqueue.Task {
	if len(tasks) == 0 {
		return nil
	}
	if len(tasks) == 1 {
		return tasks[0]
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].GetPriority() < tasks[j].GetPriority()
	})
	var totalPrio int
	for _, task := range tasks {
		totalPrio += int(task.GetPriority())
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var task tqueue.Task
	randPrio := r.Intn(totalPrio)
	for _, t := range tasks {
		totalPrio += int(t.GetPriority())
		if totalPrio >= randPrio {
			task = t
			break
		}
	}
	return task
}
