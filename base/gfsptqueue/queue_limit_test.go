package gfsptqueue

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestGfSpTQueueWithLimit_Len(t *testing.T) {
	queue := NewGfSpTQueueWithLimit("mock", 1)
	result := queue.Len()
	assert.Equal(t, 0, result)
}

func TestGfSpTQueueWithLimit_Cap(t *testing.T) {
	queue := NewGfSpTQueueWithLimit("mock", 1)
	result := queue.Cap()
	assert.Equal(t, 1, result)
}

func TestGfSpTQueueWithLimit_Has(t *testing.T) {
	queue := NewGfSpTQueueWithLimit("mock", 1)
	result := queue.Has("test")
	assert.Equal(t, false, result)
}

func TestGfSpTQueueWithLimit_TopByLimit1(t *testing.T) {
	queue := NewGfSpTQueueWithLimit("mock", 1)
	result := queue.TopByLimit(nil)
	assert.Nil(t, result)
}

func TestGfSpTQueueWithLimit_TopByLimit2(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().NotLess(gomock.Any()).Return(true).AnyTimes()
	queue := NewGfSpTQueueWithLimit("mock", 1)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{MaxRetry: 3, Retry: 5},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
	queue.SetRetireTaskStrategy(func(task coretask.Task) bool { return true })
	result := queue.TopByLimit(m)
	assert.Nil(t, result)
}

func TestGfSpTQueueWithLimit_TopByLimit3(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().NotLess(gomock.Any()).Return(true).AnyTimes()
	queue := NewGfSpTQueueWithLimit("mock", 1)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{MaxRetry: 3, Retry: 5},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
	queue.SetFilterTaskStrategy(func(task coretask.Task) bool { return false })
	result := queue.TopByLimit(m)
	assert.Nil(t, result)
}

func TestGfSpTQueueWithLimit_TopByLimit4(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().NotLess(gomock.Any()).Return(true).AnyTimes()
	queue := NewGfSpTQueueWithLimit("mock", 2)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 1},
	}
	task2 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_2"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 2},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
	err = queue.Push(task2)
	assert.Nil(t, err)
	result := queue.TopByLimit(m)
	assert.NotNil(t, result)
}

func TestGfSpTQueueWithLimit_TopByLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().NotLess(gomock.Any()).Return(true).AnyTimes()
	queue := NewGfSpTQueueWithLimit("mock", 1)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 1},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
	result := queue.PopByLimit(m)
	assert.NotNil(t, result)
}

func TestGfSpTQueueWithLimit_PopByKey1(t *testing.T) {
	queue := NewGfSpTQueueWithLimit("mock", 1)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1", Id: sdkmath.NewUint(1)},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 1},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
	result := queue.PopByKey("Uploading-bucket:-object:task_1-id:1")
	assert.NotNil(t, result)
}

func TestGfSpTQueueWithLimit_PopByKey2(t *testing.T) {
	queue := NewGfSpTQueueWithLimit("mock", 1)
	result := queue.PopByKey("Uploading-bucket:-object:task_1-id:1")
	assert.Nil(t, result)
}

func TestGfSpTQueueWithLimit_Push1(t *testing.T) {
	t.Log("Case description: push 1 task")
	queue := NewGfSpTQueueWithLimit("mock", 1)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 1},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
}

func TestGfSpTQueueWithLimit_Push2(t *testing.T) {
	t.Log("Case description: repeated task")
	queue := NewGfSpTQueueWithLimit("mock", 1)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 1},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
	err = queue.Push(task1)
	assert.Equal(t, ErrTaskRepeated, err)
}

func TestGfSpTQueueWithLimit_Push3(t *testing.T) {
	t.Log("Case description: task queue exceed")
	queue := NewGfSpTQueueWithLimit("mock", 1)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 1},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
	task2 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_2"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 2},
	}
	err = queue.Push(task2)
	assert.Equal(t, ErrTaskQueueExceed, err)
}

func TestReplicateTaskRetireByExpired(t *testing.T) {
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{MaxRetry: 3, Retry: 5},
	}
	task2 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_2"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{MaxRetry: 3, Retry: 4},
	}
	task3 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_3"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{MaxRetry: 3, Retry: 2},
	}
	task4 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_4"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{MaxRetry: 3, Retry: 2},
	}
	task5 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_5"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{MaxRetry: 3, Retry: 1},
	}
	task6 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_6"},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{MaxRetry: 3, Retry: 0},
	}

	testCases := []struct {
		name string
		task *gfsptask.GfSpReplicatePieceTask
	}{
		{
			name: "push task expired 1",
			task: task1,
		},
		{
			name: "push task expired 2",
			task: task2,
		},
		{
			name: "push task unexpired 3",
			task: task3,
		},
		{
			name: "push task unexpired 4",
			task: task4,
		},
		{
			name: "push task unexpired 1",
			task: task5,
		},
		{
			name: "push task unexpired 2",
			task: task6,
		},
	}
	retireFunc := func(qTask coretask.Task) bool {
		return qTask.ExceedRetry()
	}
	queue := NewGfSpTQueueWithLimit("test_expired_queue", 3)
	queue.SetRetireTaskStrategy(retireFunc)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := queue.Push(testCase.task)
			if testCase.task == task6 {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGfSpTQueueWithLimit_ScanTask(t *testing.T) {
	queue := NewGfSpTQueueWithLimit("mock", 1)
	task1 := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectName: "task_1", Id: sdkmath.NewUint(1)},
		StorageParams: &storagetypes.Params{},
		Task:          &gfsptask.GfSpTask{CreateTime: 1},
	}
	err := queue.Push(task1)
	assert.Nil(t, err)
	queue.ScanTask(func(task coretask.Task) { log.Info(task) })
}
