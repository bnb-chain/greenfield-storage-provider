package types

import (
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ tqueue.SealObjectTask = &SealObjectTask{}

func NewSealObjectTask(object *storagetypes.ObjectInfo) (*SealObjectTask, error) {
	if object == nil {
		return nil, ErrObjectDangling
	}
	task := &Task{
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
		RetryLimit:   DefaultSealObjectTaskRetryLimit,
		Timeout:      DefaultSealObjectTimeout,
		TaskPriority: int32(GetTaskPriorityMap().GetPriority(tqueue.TypeTaskSealObject)),
	}
	return &SealObjectTask{
		Object: object,
		Task:   task,
	}, nil
}

func (m *SealObjectTask) Key() tqueue.TKey {
	if m == nil {
		return ""
	}
	if m.GetObject() == nil {
		return ""
	}
	return tqueue.TKey(m.GetObject().Id.String())
}

func (m *SealObjectTask) Type() tqueue.TType {
	return tqueue.TypeTaskSealObject
}

func (m *SealObjectTask) LimitEstimate() rcmgr.Limit {
	limit := &rcmgr.BaseLimit{}
	prio := m.GetPriority()
	if GetTaskPriorityMap().HighLevelPriority(prio) {
		limit.TasksHighPriority = 1
	} else if GetTaskPriorityMap().LowLevelPriority(prio) {
		limit.TasksLowPriority = 1
	} else {
		limit.TasksMediumPriority = 1
	}
	return limit
}

func (m *SealObjectTask) GetCreateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetCreateTime()
}

func (m *SealObjectTask) SetCreateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetCreateTime(time)
}

func (m *SealObjectTask) GetUpdateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetUpdateTime()
}

func (m *SealObjectTask) SetUpdateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetUpdateTime(time)
}

func (m *SealObjectTask) GetTimeout() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetTimeout()
}

func (m *SealObjectTask) SetTimeout(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetTimeout(time)
}

func (m *SealObjectTask) GetPriority() tqueue.TPriority {
	if m == nil {
		return tqueue.TPriority(0)
	}
	if m.GetTask() == nil {
		return tqueue.TPriority(0)
	}
	return m.GetTask().GetPriority()
}

func (m *SealObjectTask) SetPriority(prio tqueue.TPriority) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetPriority(prio)
}

func (m *SealObjectTask) IncRetry() bool {
	if m == nil {
		return false
	}
	if m.GetTask() == nil {
		return false
	}
	return m.GetTask().IncRetry()
}

func (m *SealObjectTask) GetRetry() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetry()
}

func (m *SealObjectTask) Expired() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetUpdateTime()+m.GetTask().GetTimeout() > time.Now().Unix()
}

func (m *SealObjectTask) GetRetryLimit() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetryLimit()
}

func (m *SealObjectTask) SetRetryLimit(limit int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetRetryLimit(limit)
}

func (m *SealObjectTask) RetryExceed() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetRetry() > m.GetTask().GetRetryLimit()
}

func (m *SealObjectTask) Error() error {
	if m == nil {
		return nil
	}
	if m.GetTask() == nil {
		return nil
	}
	return m.GetTask().Error()
}

func (m *SealObjectTask) SetError(err error) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetError(err)
}
