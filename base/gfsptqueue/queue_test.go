package gfsptqueue

import (
	"testing"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func TestApprovalTaskRetireByExpiredHeight(t *testing.T) {
	task1 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_1",
			PrimarySpApproval: &storagetypes.Approval{ExpiredHeight: 99},
		},
	}

	task2 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_2",
			PrimarySpApproval: &storagetypes.Approval{ExpiredHeight: 100},
		},
	}

	task3 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_3",
			PrimarySpApproval: &storagetypes.Approval{ExpiredHeight: 101},
		},
	}

	task4 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_4",
			PrimarySpApproval: &storagetypes.Approval{ExpiredHeight: 102},
		},
	}

	task5 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_5",
			PrimarySpApproval: &storagetypes.Approval{ExpiredHeight: 103},
		},
	}

	task6 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_4",
			PrimarySpApproval: &storagetypes.Approval{ExpiredHeight: 103},
		},
	}

	testCases := []struct {
		name string
		task *gfsptask.GfSpCreateObjectApprovalTask
	}{
		{
			name: "push task expired height 99 current height 100",
			task: task1,
		},
		{
			name: "push task expired height 100 current height 100",
			task: task2,
		},
		{
			name: "push task expired height 101 current height 100",
			task: task3,
		},
		{
			name: "push task expired height 102 current height 100",
			task: task4,
		},
		{
			name: "queue exceed",
			task: task5,
		},
		{
			name: "repeated task",
			task: task6,
		},
	}

	currentBlockHeight := uint64(100)
	retireFunc := func(qTask task.Task) bool {
		task := qTask.(task.ApprovalTask)
		return task.GetExpiredHeight() < currentBlockHeight
	}
	queue := NewGfSpTQueue("test_expired_height_queue", 3)
	queue.SetRetireTaskStrategy(retireFunc)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := queue.Push(testCase.task)
			if testCase.task == task5 {
				require.Error(t, err)
			} else if testCase.task == task6 {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
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
	retireFunc := func(qTask task.Task) bool {
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
