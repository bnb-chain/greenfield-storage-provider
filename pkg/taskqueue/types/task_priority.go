package types

import (
	"errors"
	"math"
	"sync"

	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
)

var (
	ErrTaskPriorityLvlOutOfBound = errors.New("priority task level out of bound")
	ErrLowLvlExceedHigh          = errors.New("low  priority task level more than high level")
	ErrHighLvlExceedLow          = errors.New("high priority task level less than low level")
)

var _ tqueue.TTaskPriority = &TaskPriority{}

var (
	gTaskPriority *TaskPriority
	once          sync.Once
)

func GetTaskPriorityMap() *TaskPriority {
	once.Do(func() {
		gTaskPriority = getTaskPriority()
	})
	return gTaskPriority
}

type TaskPriority struct {
	typeToPriority map[tqueue.TType]tqueue.TPriority
	priorityToType map[tqueue.TPriority][]tqueue.TType
	priorityLevel  map[tqueue.TPriorityLevel]tqueue.TPriority
	mux            sync.RWMutex
}

func getTaskPriority() *TaskPriority {
	t2p := map[tqueue.TType]tqueue.TPriority{
		tqueue.TypeTaskUnknown:        tqueue.UnSchedulingPriority,
		tqueue.TypeTaskReceivePiece:   tqueue.UnSchedulingPriority,
		tqueue.TypeTaskDownloadObject: tqueue.UnSchedulingPriority,
		tqueue.TypeTaskUpload:         tqueue.UnSchedulingPriority,
		tqueue.TypeTaskGCStore:        tqueue.DefaultSmallerPriority / 4,
		tqueue.TypeTaskGCZombiePiece:  tqueue.DefaultSmallerPriority / 2,
		tqueue.TypeTaskGCObject:       tqueue.DefaultSmallerPriority,
		tqueue.TypeTaskReplicatePiece: tqueue.DefaultLargerTaskPriority,
		tqueue.TypeTaskSealObject:     tqueue.MaxTaskPriority,
	}
	p2t := make(map[tqueue.TPriority][]tqueue.TType)
	for tType, priority := range t2p {
		p2t[priority] = append(p2t[priority], tType)
	}
	p2l := map[tqueue.TPriorityLevel]tqueue.TPriority{
		tqueue.TLowPriorityLevel:  tqueue.DefaultSmallerPriority,
		tqueue.THighPriorityLevel: tqueue.DefaultLargerTaskPriority,
	}
	return &TaskPriority{
		typeToPriority: t2p,
		priorityToType: p2t,
		priorityLevel:  p2l,
	}
}

func (t *TaskPriority) GetPriority(tType tqueue.TType) tqueue.TPriority {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if _, ok := t.typeToPriority[tType]; !ok {
		return tqueue.UnKnownTaskPriority
	}
	return t.typeToPriority[tType]
}

func (t *TaskPriority) SetPriority(tType tqueue.TType, proi tqueue.TPriority) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.deleteItem(tType, proi)
	t.typeToPriority[tType] = proi
	t.priorityToType[proi] = append(t.priorityToType[proi], tType)
}

func (t *TaskPriority) deleteItem(tType tqueue.TType, proi tqueue.TPriority) {
	index := -1
	for i, taskTpye := range t.priorityToType[proi] {
		if taskTpye == tType {
			index = i
			break
		}
	}
	if index != -1 {
		t.priorityToType[proi] = append(t.priorityToType[proi][0:index], t.priorityToType[proi][index+1:]...)
	}
	delete(t.typeToPriority, tType)
}

func (t *TaskPriority) GetAllPriorities() map[tqueue.TType]tqueue.TPriority {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return t.typeToPriority
}

func (t *TaskPriority) SetAllPriorities(t2p map[tqueue.TType]tqueue.TPriority) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.typeToPriority = t2p
	p2t := make(map[tqueue.TPriority][]tqueue.TType)
	for tType, priority := range t.typeToPriority {
		p2t[priority] = append(p2t[priority], tType)
	}
	t.priorityToType = p2t
}

func (t *TaskPriority) GetLowLevelPriority() tqueue.TPriority {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return t.priorityLevel[tqueue.TLowPriorityLevel]
}

func (t *TaskPriority) SetLowLevelPriority(lvl tqueue.TPriority) error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if lvl < 0 || lvl > math.MaxUint8 {
		return ErrTaskPriorityLvlOutOfBound
	}
	if uint8(lvl) > uint8(t.priorityLevel[tqueue.THighPriorityLevel]) {
		return ErrLowLvlExceedHigh
	}
	t.priorityLevel[tqueue.TLowPriorityLevel] = lvl
	return nil
}

func (t *TaskPriority) GetHighLevelPriority() tqueue.TPriority {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return t.priorityLevel[tqueue.THighPriorityLevel]
}

func (t *TaskPriority) SetHighLevelPriority(lvl tqueue.TPriority) error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if lvl < 0 || lvl > math.MaxUint8 {
		return ErrTaskPriorityLvlOutOfBound
	}
	if uint8(lvl) < uint8(t.priorityLevel[tqueue.TLowPriorityLevel]) {
		return ErrHighLvlExceedLow
	}
	t.priorityLevel[tqueue.THighPriorityLevel] = lvl
	return nil
}

func (t *TaskPriority) HighLevelPriority(lvl tqueue.TPriority) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()
	hPrio := t.priorityLevel[tqueue.THighPriorityLevel]
	return lvl >= hPrio
}

func (t *TaskPriority) LowLevelPriority(lvl tqueue.TPriority) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()
	lPrio := t.priorityLevel[tqueue.TLowPriorityLevel]
	return lvl >= lPrio
}

func (t *TaskPriority) SupportTask(tType tqueue.TType) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()
	_, ok := t.typeToPriority[tType]
	return ok
}
