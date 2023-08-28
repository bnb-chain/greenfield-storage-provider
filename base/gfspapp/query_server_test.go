package gfspapp

import (
	"context"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/core/module"
)

func TestGfSpBaseApp_GfSpQueryTasksSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)

	m1 := module.NewMockApprover(ctrl)
	g.approver = m1
	task1 := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			BucketName: "bucket1",
			ObjectName: "object1",
		},
	}
	taskList1 := make([]coretask.Task, 0)
	taskList1 = append(taskList1, task1)
	m1.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(taskList1, nil).Times(1)

	m2 := module.NewMockDownloader(ctrl)
	g.downloader = m2
	task2 := &gfsptask.GfSpDownloadObjectTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	taskList2 := make([]coretask.Task, 0)
	taskList2 = append(taskList2, task2)
	m2.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(taskList2, nil).Times(1)

	m3 := module.NewMockManager(ctrl)
	g.manager = m3
	task3 := &gfsptask.GfSpGCMetaTask{
		Task:        &gfsptask.GfSpTask{},
		CurrentIdx:  1,
		DeleteCount: 2,
	}
	taskList3 := make([]coretask.Task, 0)
	taskList3 = append(taskList3, task3)
	m3.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(taskList3, nil).Times(1)

	m4 := module.NewMockP2P(ctrl)
	g.p2p = m4
	task4 := &gfsptask.GfSpReplicatePieceApprovalTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	taskList4 := make([]coretask.Task, 0)
	taskList4 = append(taskList4, task4)
	m4.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(taskList4, nil).Times(1)

	m5 := module.NewMockReceiver(ctrl)
	g.receiver = m5
	task5 := &gfsptask.GfSpReceivePieceTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	taskList5 := make([]coretask.Task, 0)
	taskList5 = append(taskList5, task5)
	m5.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(taskList5, nil).Times(1)

	m6 := module.NewMockUploader(ctrl)
	g.uploader = m6
	task6 := &gfsptask.GfSpUploadObjectTask{
		Task:                 &gfsptask.GfSpTask{},
		VirtualGroupFamilyId: 1,
		ObjectInfo:           mockObjectInfo,
		StorageParams:        mockStorageParams,
	}
	taskList6 := make([]coretask.Task, 0)
	taskList6 = append(taskList6, task6)
	m6.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(taskList6, nil).Times(1)

	req := &gfspserver.GfSpQueryTasksRequest{TaskSubKey: "mockTaskSubKey"}
	result, err := g.GfSpQueryTasks(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(result.GetTaskInfo()))
}

func TestGfSpBaseApp_GfSpQueryTasksFailure1(t *testing.T) {
	t.Log("Failure case description: invalid query key")
	g := setup(t)
	req := &gfspserver.GfSpQueryTasksRequest{}
	result, err := g.GfSpQueryTasks(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, "invalid query key", result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpQueryTasksFailure2(t *testing.T) {
	t.Log("Failure case description: no match tasks")
	g := setup(t)
	ctrl := gomock.NewController(t)

	m1 := module.NewMockApprover(ctrl)
	g.approver = m1
	m1.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)

	m2 := module.NewMockDownloader(ctrl)
	g.downloader = m2
	m2.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)

	m3 := module.NewMockManager(ctrl)
	g.manager = m3
	m3.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)

	m4 := module.NewMockP2P(ctrl)
	g.p2p = m4
	m4.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)

	m5 := module.NewMockReceiver(ctrl)
	g.receiver = m5
	m5.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)

	m6 := module.NewMockUploader(ctrl)
	g.uploader = m6
	m6.EXPECT().QueryTasks(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)

	req := &gfspserver.GfSpQueryTasksRequest{TaskSubKey: "mockTaskSubKey"}
	result, err := g.GfSpQueryTasks(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, "no match tasks", result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpQueryBucketMigrate(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().QueryBucketMigrate(gomock.Any()).Return(&gfspserver.GfSpQueryBucketMigrateResponse{SelfSpId: 4},
		nil).Times(1)
	result, err := g.GfSpQueryBucketMigrate(context.TODO(), &gfspserver.GfSpQueryBucketMigrateRequest{})
	assert.Nil(t, err)
	assert.Equal(t, uint32(4), result.GetSelfSpId())
}

func TestGfSpBaseApp_GfSpQuerySpExit(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().QuerySpExit(gomock.Any()).Return(&gfspserver.GfSpQuerySpExitResponse{SelfSpId: 4},
		nil).Times(1)
	result, err := g.GfSpQuerySpExit(context.TODO(), &gfspserver.GfSpQuerySpExitRequest{})
	assert.Nil(t, err)
	assert.Equal(t, uint32(4), result.GetSelfSpId())
}
