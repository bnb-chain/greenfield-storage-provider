package gfsptask

import (
	"fmt"
	"time"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

var _ coretask.DiscontinueBucketTask = &GfSpDiscontinueBucketTask{}

func (m *GfSpDiscontinueBucketTask) InitDiscontinueBucketTask(createAt uint64,
	reason string,
	limit uint64,
	priority coretask.TPriority,
	timeout int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetCreateAt(createAt)
	m.SetReason(reason)
	m.SetLimit(limit)
	m.SetPriority(priority)
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetTimeout(timeout)
}

func (m *GfSpDiscontinueBucketTask) Key() coretask.TKey {
	return GfSpDiscontinueBucketTaskKey(m.CreateAt)
}

func (m *GfSpDiscontinueBucketTask) Type() coretask.TType {
	return coretask.TypeTaskGCObject
}

func (m *GfSpDiscontinueBucketTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], createAt[%d], reason[%s], limit[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.GetCreateAt(), m.GetReason(), m.GetLimit(), m.GetTask().Info())
}

func (m *GfSpDiscontinueBucketTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpDiscontinueBucketTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpDiscontinueBucketTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpDiscontinueBucketTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpDiscontinueBucketTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpDiscontinueBucketTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpDiscontinueBucketTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpDiscontinueBucketTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpDiscontinueBucketTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpDiscontinueBucketTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpDiscontinueBucketTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpDiscontinueBucketTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpDiscontinueBucketTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpDiscontinueBucketTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpDiscontinueBucketTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpDiscontinueBucketTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpDiscontinueBucketTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpDiscontinueBucketTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpDiscontinueBucketTask) EstimateLimit() corercmgr.Limit {
	return LimitEstimateByPriority(m.GetPriority())
}

func (m *GfSpDiscontinueBucketTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpDiscontinueBucketTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpDiscontinueBucketTask) SetCreateAt(createAt uint64) {
	m.CreateAt = createAt
}

func (m *GfSpDiscontinueBucketTask) SetReason(reason string) {
	m.Reason = reason
}

func (m *GfSpDiscontinueBucketTask) SetLimit(limit uint64) {
	m.Limit = limit
}
