package types

import (
	"strconv"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
)

var _ tqueue.GCObjectTask = &GCObjectTask{}

func NewGCObjectTask(low, high uint64) (*GCObjectTask, error) {
	task := &Task{
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
		RetryLimit:   DefaultGCObjectTaskRetryLimit,
		Timeout:      DefaultGCObjectTimeout,
		TaskPriority: int32(GetTaskPriorityMap().GetPriority(tqueue.TypeTaskGCObject)),
	}
	return &GCObjectTask{
		Task:             task,
		StartBlockNumber: low,
		EndBlockNumber:   high,
	}, nil
}

func (m *GCObjectTask) Key() tqueue.TKey {
	if m == nil {
		return ""
	}
	return tqueue.TKey("GCObject-" + strconv.FormatInt(m.GetCreateTime(), 10) +
		"-" + strconv.FormatUint(m.GetStartBlockNumber(), 10) +
		"-" + strconv.FormatUint(m.GetEndBlockNumber(), 10))
}

func (m *GCObjectTask) Type() tqueue.TType {
	return tqueue.TypeTaskGCObject
}

func (m *GCObjectTask) LimitEstimate() rcmgr.Limit {
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

func (m *GCObjectTask) GetCreateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetCreateTime()
}

func (m *GCObjectTask) SetCreateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetCreateTime(time)
}

func (m *GCObjectTask) GetUpdateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetUpdateTime()
}

func (m *GCObjectTask) SetUpdateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetUpdateTime(time)
}

func (m *GCObjectTask) GetTimeout() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetTimeout()
}

func (m *GCObjectTask) SetTimeout(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetTimeout(time)
}

func (m *GCObjectTask) GetPriority() tqueue.TPriority {
	if m == nil {
		return tqueue.TPriority(0)
	}
	if m.GetTask() == nil {
		return tqueue.TPriority(0)
	}
	return m.GetTask().GetPriority()
}

func (m *GCObjectTask) SetPriority(prio tqueue.TPriority) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetPriority(prio)
}

func (m *GCObjectTask) IncRetry() bool {
	if m == nil {
		return false
	}
	if m.GetTask() == nil {
		return false
	}
	return m.GetTask().GetRetry() <= m.GetTask().GetRetryLimit()
}

func (m *GCObjectTask) GetRetry() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetry()
}

func (m *GCObjectTask) Expired() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetUpdateTime()+m.GetTask().GetTimeout() < time.Now().Unix()
}

func (m *GCObjectTask) GetRetryLimit() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetryLimit()
}

func (m *GCObjectTask) SetRetryLimit(limit int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetRetryLimit(limit)
}

func (m *GCObjectTask) RetryExceed() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetRetry() > m.GetTask().GetRetryLimit()
}

func (m *GCObjectTask) Error() error {
	if m == nil {
		return nil
	}
	if m.GetTask() == nil {
		return nil
	}
	return m.GetTask().Error()
}

func (m *GCObjectTask) SetError(err error) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetError(err)
}

func (m *GCObjectTask) SetStartBlockNumber(num uint64) {
	if m == nil {
		return
	}
	m.StartBlockNumber = num
}

func (m *GCObjectTask) SetEndBlockNumber(num uint64) {
	if m == nil {
		return
	}
	m.EndBlockNumber = num
}

func (m *GCObjectTask) GetGCObjectProcess() (uint64, uint64) {
	if m == nil {
		return 0, 0
	}
	return m.GetCurrentBlockNumber(), m.GetObjectId()
}

func (m *GCObjectTask) SetGCObjectProcess(block, object uint64) {
	if m == nil {
		return
	}
	m.CurrentBlockNumber = block
	m.ObjectId = object
}
