package types

import (
	"strconv"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
)

var _ tqueue.GCStoreTask = &GCStoreTask{}

func NewGCStoreTask() (*GCStoreTask, error) {
	task := &Task{
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
		RetryLimit:   DefaultGCStoreTaskRetryLimit,
		Timeout:      DefaultGCStoreTimeout,
		TaskPriority: int32(GetTaskPriorityMap().GetPriority(tqueue.TypeTaskGCStore)),
	}
	return &GCStoreTask{
		Task: task,
	}, nil
}

func (m *GCStoreTask) Key() tqueue.TKey {
	if m == nil {
		return ""
	}
	return tqueue.TKey("GCStore" + "-" + strconv.FormatInt(m.GetCreateTime(), 10))
}

func (m *GCStoreTask) Type() tqueue.TType {
	return tqueue.TypeTaskGCStore
}

func (m *GCStoreTask) LimitEstimate() rcmgr.Limit {
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

func (m *GCStoreTask) GetCreateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetCreateTime()
}

func (m *GCStoreTask) SetCreateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetCreateTime(time)
}

func (m *GCStoreTask) GetUpdateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetUpdateTime()
}

func (m *GCStoreTask) SetUpdateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetUpdateTime(time)
}

func (m *GCStoreTask) GetTimeout() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetTimeout()
}

func (m *GCStoreTask) SetTimeout(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetTimeout(time)
}

func (m *GCStoreTask) GetPriority() tqueue.TPriority {
	if m == nil {
		return tqueue.TPriority(0)
	}
	if m.GetTask() == nil {
		return tqueue.TPriority(0)
	}
	return m.GetTask().GetPriority()
}

func (m *GCStoreTask) SetPriority(prio tqueue.TPriority) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetPriority(prio)
}

func (m *GCStoreTask) IncRetry() bool {
	if m == nil {
		return false
	}
	if m.GetTask() == nil {
		return false
	}
	return m.GetTask().GetRetry() <= m.GetTask().GetRetryLimit()
}

func (m *GCStoreTask) GetRetry() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetry()
}

func (m *GCStoreTask) Expired() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetUpdateTime()+m.GetTask().GetTimeout() < time.Now().Unix()
}

func (m *GCStoreTask) GetRetryLimit() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetryLimit()
}

func (m *GCStoreTask) SetRetryLimit(limit int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetRetryLimit(limit)
}

func (m *GCStoreTask) RetryExceed() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetRetry() > m.GetTask().GetRetryLimit()
}

func (m *GCStoreTask) Error() error {
	if m == nil {
		return nil
	}
	if m.GetTask() == nil {
		return nil
	}
	return m.GetTask().Error()
}

func (m *GCStoreTask) SetError(err error) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetError(err)
}

func (m *GCStoreTask) GetGCStoreStatus() (uint64, uint64) {
	if m == nil {
		return 0, 0
	}
	return m.GetCurrentIdx(), m.GetDeleteCount()
}

func (m *GCStoreTask) SetGCStoreStatus(current uint64, delete uint64) {
	if m == nil {
		return
	}
	if m.GetCurrentIdx() > current {
		return
	}
	m.CurrentIdx = current
	if m.GetDeleteCount() < delete {
		return
	}
	m.DeleteCount = delete
}
