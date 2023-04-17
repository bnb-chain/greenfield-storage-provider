package types

import (
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueue "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ tqueue.ReceivePieceTask = &ReceivePieceTask{}

func NewReceivePieceTask(object *storagetypes.ObjectInfo) (*ReceivePieceTask, error) {
	if object == nil {
		return nil, ErrObjectDangling
	}
	task := &Task{
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
		RetryLimit:   DefaultReplicatePieceTaskRetryLimit,
		Timeout:      DefaultUploadMaxTimeout,
		TaskPriority: int32(GetTaskPriorityMap().GetPriority(tqueue.TypeTaskReceivePiece)),
	}
	return &ReceivePieceTask{
		Object: object,
		Task:   task,
	}, nil
}

func (m *ReceivePieceTask) Key() tqueue.TKey {
	if m == nil {
		return ""
	}
	if m.GetObject() == nil {
		return ""
	}
	return tqueue.TKey(m.GetObject().Id.String())
}

func (m *ReceivePieceTask) Type() tqueue.TType {
	return tqueue.TypeTaskReceivePiece
}

func (m *ReceivePieceTask) LimitEstimate() rcmgr.Limit {
	return rcmgr.InfiniteLimit()
}

func (m *ReceivePieceTask) GetCreateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetCreateTime()
}

func (m *ReceivePieceTask) SetCreateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetCreateTime(time)
}

func (m *ReceivePieceTask) GetUpdateTime() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetUpdateTime()
}

func (m *ReceivePieceTask) SetUpdateTime(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetUpdateTime(time)
}

func (m *ReceivePieceTask) GetTimeout() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetTimeout()
}

func (m *ReceivePieceTask) SetTimeout(time int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetTimeout(time)
}

func (m *ReceivePieceTask) GetPriority() tqueue.TPriority {
	if m == nil {
		return tqueue.TPriority(0)
	}
	if m.GetTask() == nil {
		return tqueue.TPriority(0)
	}
	return m.GetTask().GetPriority()
}

func (m *ReceivePieceTask) SetPriority(prio tqueue.TPriority) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetPriority(prio)
}

func (m *ReceivePieceTask) IncRetry() bool {
	if m == nil {
		return false
	}
	if m.GetTask() == nil {
		return false
	}
	return m.GetTask().GetRetry() <= m.GetTask().GetRetryLimit()
}

func (m *ReceivePieceTask) GetRetry() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetry()
}

func (m *ReceivePieceTask) Expired() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetUpdateTime()+m.GetTask().GetTimeout() < time.Now().Unix()
}

func (m *ReceivePieceTask) GetRetryLimit() int64 {
	if m == nil {
		return 0
	}
	if m.GetTask() == nil {
		return 0
	}
	return m.GetTask().GetRetryLimit()
}

func (m *ReceivePieceTask) SetRetryLimit(limit int64) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetRetryLimit(limit)
}

func (m *ReceivePieceTask) RetryExceed() bool {
	if m == nil {
		return true
	}
	if m.GetTask() == nil {
		return true
	}
	return m.GetTask().GetRetry() > m.GetTask().GetRetryLimit()
}

func (m *ReceivePieceTask) Error() error {
	if m == nil {
		return nil
	}
	if m.GetTask() == nil {
		return nil
	}
	return m.GetTask().Error()
}

func (m *ReceivePieceTask) SetError(err error) {
	if m == nil {
		return
	}
	if m.GetTask() == nil {
		return
	}
	m.GetTask().SetError(err)
}

func (m *ReceivePieceTask) SetReplicateIdx(rIdx uint32) {
	m.ReplicateIdx = rIdx
}
