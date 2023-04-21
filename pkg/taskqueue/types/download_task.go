package types

import (
	"strconv"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ tqueue.DownloadObjectTask = &DownloadObjectTask{}

func NewDownloadObjectTask(object *storagetypes.ObjectInfo) (*DownloadObjectTask, error) {
	if object == nil {
		return nil, ErrObjectDangling
	}
	task := &Task{
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
		TaskPriority: int32(GetTaskPriorityMap().GetPriority(tqueue.TypeTaskDownloadObject)),
	}
	return &DownloadObjectTask{
		Object: object,
		Task:   task,
	}, nil
}

func (m *DownloadObjectTask) Key() tqueue.TKey {
	if m == nil {
		return ""
	}
	if m.GetObject() == nil {
		return ""
	}
	return tqueue.TKey("DownloadObject" + m.GetObject().Id.String() +
		"-" + strconv.FormatUint(m.GetLow(), 10) +
		"-" + strconv.FormatUint(m.GetHigh(), 10))
}

func (m *DownloadObjectTask) Type() tqueue.TType {
	return tqueue.TypeTaskDownloadObject
}

func (m *DownloadObjectTask) LimitEstimate() rcmgr.Limit {
	return rcmgr.InfiniteLimit()
}

func (m *DownloadObjectTask) GetCreateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetCreateTime()
}

func (m *DownloadObjectTask) SetCreateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetCreateTime(time)
}

func (m *DownloadObjectTask) GetUpdateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetUpdateTime()
}

func (m *DownloadObjectTask) SetUpdateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetUpdateTime(time)
}

func (m *DownloadObjectTask) GetTimeout() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetTimeout()
}

func (m *DownloadObjectTask) SetTimeout(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetTimeout(time)
}

func (m *DownloadObjectTask) GetPriority() tqueue.TPriority {
	if m == nil {
		return tqueue.TPriority(0)
	}
	if m.GetTask() == nil {
		return tqueue.TPriority(0)
	}
	return m.GetTask().GetPriority()
}

func (m *DownloadObjectTask) SetPriority(prio tqueue.TPriority) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetPriority(prio)
}

func (m *DownloadObjectTask) IncRetry() bool {
	if m == nil {
		return false
	}
	if m.GetTask() == nil {
		return false
	}
	return m.GetTask().IncRetry()
}

func (m *DownloadObjectTask) GetRetry() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetry()
}

func (m *DownloadObjectTask) Expired() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetUpdateTime()+m.GetTask().GetTimeout() > time.Now().Unix()
}

func (m *DownloadObjectTask) GetRetryLimit() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetryLimit()
}

func (m *DownloadObjectTask) SetRetryLimit(limit int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetRetryLimit(limit)
}

func (m *DownloadObjectTask) RetryExceed() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetRetry() > m.GetTask().GetRetryLimit()
}

func (m *DownloadObjectTask) Error() error {
	if m == nil {
		return nil
	}
	if m.GetTask() == nil {
		return nil
	}
	return m.GetTask().Error()
}

func (m *DownloadObjectTask) SetError(err error) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetError(err)
}

func (m *DownloadObjectTask) GetSize() uint64 {
	if m == nil {
		return 0
	}
	if m.GetHigh() < m.GetLow() {
		return 0
	}
	return m.GetHigh() - m.GetLow()
}

func (m *DownloadObjectTask) SetLow(low uint64) {
	if m == nil {
		return
	}
	m.Low = low
}

func (m *DownloadObjectTask) SetHigh(high uint64) {
	if m == nil {
		return
	}
	m.High = high
}
