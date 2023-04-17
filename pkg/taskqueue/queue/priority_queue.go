package queue

import (
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
)

var _ tqueue.TPriorityQueueWithLimit = &TaskPriorityQueue{}

type TaskPriorityQueue struct {
	name         string
	pqueue       map[tqueue.TPriority]tqueue.TQueueWithLimit
	strategy     tqueue.TQueueStrategy
	supportLimit bool
	mux          sync.RWMutex
}

func NewTaskPriorityQueue(
	name string,
	pqueue map[tqueue.TPriority]tqueue.TQueueWithLimit,
	strategy tqueue.TQueueStrategy,
	limit bool) tqueue.TPriorityQueueWithLimit {
	return &TaskPriorityQueue{
		name:         name,
		pqueue:       pqueue,
		strategy:     strategy,
		supportLimit: limit,
	}
}

func (p *TaskPriorityQueue) SetStrategy(strategy tqueue.TQueueStrategy) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.strategy = strategy
}

func (p *TaskPriorityQueue) Len() int {
	p.mux.RLock()
	defer p.mux.RUnlock()
	var length int
	for _, queue := range p.pqueue {
		length += queue.Len()
	}
	return length
}

func (p *TaskPriorityQueue) Cap() int {
	p.mux.RLock()
	defer p.mux.RUnlock()
	var capacity int
	for _, queue := range p.pqueue {
		capacity += queue.Cap()
	}
	return capacity
}

func (p *TaskPriorityQueue) Top() tqueue.Task {
	p.mux.RLock()
	defer p.mux.RUnlock()
	if len(p.pqueue) == 0 {
		return nil
	}
	queue, ok := p.pqueue[p.getMaxPriority()]
	if !ok {
		log.Errorf("BUG: [%s] max priority out of bound", p.name)
		return nil
	}
	return queue.Top()
}

func (p *TaskPriorityQueue) Pop() tqueue.Task {
	p.mux.Lock()
	defer p.mux.Unlock()
	if len(p.pqueue) == 0 {
		return nil
	}
	queue, ok := p.pqueue[p.getMaxPriority()]
	if !ok {
		log.Errorf("BUG: [%s] max priority out of bound", p.name)
		return nil
	}
	return queue.Pop()
}

func (p *TaskPriorityQueue) Push(task tqueue.Task) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	prio := task.GetPriority()
	if _, ok := p.pqueue[prio]; !ok {
		log.Errorw("task priority unsupported", "queue", p.name, "priority", prio)
		return ErrUnsupportedTask
	}
	queue := p.pqueue[prio]
	return queue.Push(task)
}

func (p *TaskPriorityQueue) PopPush(task tqueue.Task) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	prio := task.GetPriority()
	if _, ok := p.pqueue[prio]; !ok {
		log.Errorw("task priority unsupported", "queue", p.name, "priority", prio)
		return ErrUnsupportedTask
	}
	queue := p.pqueue[prio]
	return queue.PopPush(task)
}

func (p *TaskPriorityQueue) Has(key tqueue.TKey) bool {
	p.mux.Lock()
	defer p.mux.Unlock()
	for _, queue := range p.pqueue {
		if queue.Has(key) {
			return true
		}
	}
	return false
}

func (p *TaskPriorityQueue) PopByKey(key tqueue.TKey) tqueue.Task {
	p.mux.Lock()
	defer p.mux.Unlock()
	for _, queue := range p.pqueue {
		if task := queue.PopByKey(key); task != nil {
			return task
		}
	}
	return nil
}

func (p *TaskPriorityQueue) GetPriorities() []tqueue.TPriority {
	p.mux.RLock()
	defer p.mux.RUnlock()
	var prios []tqueue.TPriority
	for prio, _ := range p.pqueue {
		prios = append(prios, prio)
	}
	return prios
}

func (p *TaskPriorityQueue) SetPriorityQueue(prio tqueue.TPriority, queue tqueue.TQueueWithLimit) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	if _, ok := p.pqueue[prio]; ok {
		return nil
	}
	p.pqueue[prio] = queue
	return nil
}

func (p *TaskPriorityQueue) PopByLimit(limit rcmgr.Limit) tqueue.Task {
	p.mux.Lock()
	defer p.mux.Unlock()
	var tasks []tqueue.Task
	for _, queue := range p.pqueue {
		if !queue.GetSupportPickUpByLimit() {
			continue
		}
		task := queue.PopByLimit(limit)
		if task == nil {
			continue
		}
		tasks = append(tasks, task)
	}
	task := p.strategy.RunPickUpStrategy(tasks)
	for _, t := range tasks {
		if task.GetPriority() == t.GetPriority() {
			continue
		}
		queue := p.pqueue[t.GetPriority()]
		queue.Push(t)
	}
	return task
}

func (p *TaskPriorityQueue) Expired(key tqueue.TKey) bool {
	p.mux.RLock()
	defer p.mux.RUnlock()
	for _, queue := range p.pqueue {
		if queue.Expired(key) {
			return true
		}
	}
	return false
}

func (q *TaskPriorityQueue) IsActiveTask(key tqueue.TKey) bool {
	q.mux.RLock()
	defer q.mux.RUnlock()
	for _, queue := range q.pqueue {
		if queue.IsActiveTask(key) {
			return true
		}
	}
	return false
}

func (q *TaskPriorityQueue) GetSupportPickUpByLimit() bool {
	q.mux.RLock()
	defer q.mux.RUnlock()
	return q.supportLimit
}

func (q *TaskPriorityQueue) SetSupportPickUpByLimit(support bool) {
	q.mux.Lock()
	defer q.mux.Unlock()
	q.supportLimit = support
}

func (p *TaskPriorityQueue) SubQueueLen(prio tqueue.TPriority) int {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.pqueue[prio].Len()
}

func (p *TaskPriorityQueue) RunCollection() {
	p.mux.RLock()
	defer p.mux.RUnlock()
	for _, queue := range p.pqueue {
		// performance should be better, otherwise it will block the write operation
		queue.RunCollection()
	}
}

func (p *TaskPriorityQueue) getMaxPriority() tqueue.TPriority {
	var maxPrio tqueue.TPriority
	for prio, queue := range p.pqueue {
		if queue.Len() == 0 {
			continue
		}
		if prio > maxPrio {
			maxPrio = prio
		}
	}
	log.Debugf("[%s] max priority [%d]", p.name, maxPrio)
	return maxPrio
}
