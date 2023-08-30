package gfsptqueue

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield/types/common"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestGfSpTQueue_Len(t *testing.T) {
	queue := NewGfSpTQueue("mock", 1)
	result := queue.Len()
	assert.Equal(t, 0, result)
}

func TestGfSpTQueue_Cap(t *testing.T) {
	queue := NewGfSpTQueue("mock", 1)
	result := queue.Cap()
	assert.Equal(t, 1, result)
}

func TestGfSpTQueue_Has(t *testing.T) {
	queue := NewGfSpTQueue("mock", 1)
	result := queue.Has("mock")
	assert.Equal(t, false, result)
}

func TestGfSpTQueue_Top1(t *testing.T) {
	t.Log("Case description: empty queue")
	queue := NewGfSpTQueue("mock", 1)
	result := queue.Top()
	assert.Nil(t, result)
}

func TestGfSpTQueue_Top2(t *testing.T) {
	t.Log("Case description: gcFunc returns false")
	queue := NewGfSpTQueue("mock", 1)
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}
	err := queue.Push(approvalTask)
	assert.Nil(t, err)
	queue.SetFilterTaskStrategy(func(task coretask.Task) bool { return false })
	result := queue.Top()
	assert.Nil(t, result)
}

func TestGfSpTQueue_Top3(t *testing.T) {
	t.Log("Case description: gcFunc returns true")
	queue := NewGfSpTQueue("mock", 1)
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}
	err := queue.Push(approvalTask)
	assert.Nil(t, err)
	queue.SetRetireTaskStrategy(func(task coretask.Task) bool { return true })
	result := queue.Top()
	assert.Nil(t, result)
}

func TestGfSpTQueue_Top4(t *testing.T) {
	queue := NewGfSpTQueue("mock", 2)
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{
			CreateTime: 1,
		},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}
	approvalTask1 := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{
			CreateTime: 2,
		},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName1",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 100},
		},
	}
	err := queue.Push(approvalTask)
	assert.Nil(t, err)
	err = queue.Push(approvalTask1)
	assert.Nil(t, err)
	result := queue.Top()
	assert.NotNil(t, result)
}

func TestGfSpTQueue_Pop(t *testing.T) {
	queue := NewGfSpTQueue("mock", 1)
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}
	err := queue.Push(approvalTask)
	assert.Nil(t, err)
	result := queue.Pop()
	assert.NotNil(t, result)
}

func TestGfSpTQueue_PopByKey1(t *testing.T) {
	t.Log("Case description: has no key")
	queue := NewGfSpTQueue("mock", 1)
	result := queue.PopByKey("test")
	assert.Nil(t, result)
}

func TestGfSpTQueue_PopByKey2(t *testing.T) {
	queue := NewGfSpTQueue("mock", 1)
	approvalTask1 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}
	err := queue.Push(approvalTask1)
	assert.Nil(t, err)
	result := queue.PopByKey("CreateObjectApproval-bucket:-object:mockObjectName-account:-fingerprint:")
	fmt.Println(result)
	assert.NotNil(t, result)
}

func TestGfSpTQueue_Push(t *testing.T) {
	t.Log("Case description: queue exceeds")
	queue := NewGfSpTQueue("mock", 1)
	approvalTask1 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName1",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}
	approvalTask2 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName2",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}
	err := queue.Push(approvalTask1)
	assert.Nil(t, err)
	err = queue.Push(approvalTask2)
	assert.Equal(t, ErrTaskQueueExceed, err)
}

func TestApprovalTaskRetireByExpiredHeight(t *testing.T) {
	task1 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_1",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}

	task2 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_2",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 100},
		},
	}

	task3 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_3",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 101},
		},
	}

	task4 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_4",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 102},
		},
	}

	task5 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_5",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 103},
		},
	}

	task6 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "test_task_4",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 103},
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
	retireFunc := func(qTask coretask.Task) bool {
		task := qTask.(coretask.ApprovalTask)
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

func TestGfSpTQueue_SetFilterTaskStrategy(t *testing.T) {
	queue := NewGfSpTQueue("mock", 1)
	queue.SetFilterTaskStrategy(func(task coretask.Task) bool { return true })
}

func TestGfSpTQueue_ScanTask(t *testing.T) {
	queue := NewGfSpTQueue("mock", 1)
	approvalTask1 := &gfsptask.GfSpCreateObjectApprovalTask{
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			ObjectName:        "mockObjectName",
			PrimarySpApproval: &common.Approval{ExpiredHeight: 99},
		},
	}
	err := queue.Push(approvalTask1)
	assert.Nil(t, err)
	queue.ScanTask(func(task coretask.Task) { log.Info(task) })
}
