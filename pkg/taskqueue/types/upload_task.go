package types

import (
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ tqueue.UploadObjectTask = &UploadObjectTask{}

func NewUploadObjectTask(objectInfo *storagetypes.ObjectInfo) (*UploadObjectTask, error) {
	if objectInfo == nil {
		return nil, ErrObjectDangling
	}
	task := &Task{
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
		Timeout:      ComputeTransferDataTime(objectInfo.GetPayloadSize()),
		TaskPriority: int32(GetTaskPriorityMap().GetPriority(tqueue.TypeTaskUploadObject)),
	}
	return &UploadObjectTask{
		ObjectInfo: objectInfo,
		Task:       task,
	}, nil
}

func (m *UploadObjectTask) Key() tqueue.TKey {
	if m == nil {
		return ""
	}
	if m.GetObjectInfo() == nil {
		return ""
	}
	return tqueue.TKey(m.GetObjectInfo().Id.String())
}

func (m *UploadObjectTask) Type() tqueue.TType {
	return tqueue.TypeTaskUploadObject
}

func (m *UploadObjectTask) LimitEstimate() rcmgr.Limit {
	return rcmgr.InfiniteLimit()
}

func (m *UploadObjectTask) GetCreateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetCreateTime()
}

func (m *UploadObjectTask) SetCreateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetCreateTime(time)
}

func (m *UploadObjectTask) GetUpdateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetUpdateTime()
}

func (m *UploadObjectTask) SetUpdateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetUpdateTime(time)
}

func (m *UploadObjectTask) GetTimeout() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetTimeout()
}

func (m *UploadObjectTask) SetTimeout(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetTimeout(time)
}

func (m *UploadObjectTask) GetPriority() tqueue.TPriority {
	if m == nil {
		return tqueue.TPriority(0)
	}
	if m.GetTask() == nil {
		return tqueue.TPriority(0)
	}
	return m.GetTask().GetPriority()
}

func (m *UploadObjectTask) SetPriority(prio tqueue.TPriority) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetPriority(prio)
}

func (m *UploadObjectTask) IncRetry() bool {
	if m == nil {
		return false
	}
	if m.GetTask() == nil {
		return false
	}
	return m.GetTask().IncRetry()
}

func (m *UploadObjectTask) GetRetry() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetry()
}

func (m *UploadObjectTask) Expired() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetUpdateTime()+m.GetTask().GetTimeout() > time.Now().Unix()
}

func (m *UploadObjectTask) GetRetryLimit() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetryLimit()
}

func (m *UploadObjectTask) SetRetryLimit(limit int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetRetryLimit(limit)
}

func (m *UploadObjectTask) RetryExceed() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetRetry() > m.GetTask().GetRetryLimit()
}

func (m *UploadObjectTask) Error() error {
	if m == nil {
		return nil
	}
	if m.GetTask() == nil {
		return nil
	}
	return m.GetTask().Error()
}

func (m *UploadObjectTask) SetError(err error) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetError(err)
}
