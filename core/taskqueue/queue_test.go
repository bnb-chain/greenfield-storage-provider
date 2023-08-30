package taskqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func TestScanTQueueBySubKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockTQueue(ctrl)
	m1 := coretask.NewMockTask(ctrl)
	m1.EXPECT().Key().Return(coretask.TKey("test")).Times(1)
	m.EXPECT().ScanTask(gomock.Any()).DoAndReturn(func(f func(task coretask.Task)) { f(m1) }).Times(1)
	result, err := ScanTQueueBySubKey(m, "test")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestScanTQueueWithLimitBySubKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockTQueueWithLimit(ctrl)
	m1 := coretask.NewMockTask(ctrl)
	m1.EXPECT().Key().Return(coretask.TKey("test")).Times(1)
	m.EXPECT().ScanTask(gomock.Any()).DoAndReturn(func(f func(task coretask.Task)) { f(m1) }).Times(1)
	result, err := ScanTQueueWithLimitBySubKey(m, "test")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestNilQueue(t *testing.T) {
	nq := &NilQueue{}
	nq.Top()
	nq.Pop()
	nq.PopByKey("test")
	nq.Has("test")
	err := nq.Push(nil)
	assert.Nil(t, err)
	nq.Len()
	nq.Cap()
	nq.ScanTask(func(task coretask.Task) {})
	nq.TopByLimit(nil)
	nq.PopByLimit(nil)
	nq.SetFilterTaskStrategy(func(task coretask.Task) bool { return false })
	nq.SetRetireTaskStrategy(func(task coretask.Task) bool { return false })
}
