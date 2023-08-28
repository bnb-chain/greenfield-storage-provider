package gfspapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func TestGfSpBaseApp_GfSpBeginTaskSuccess1(t *testing.T) {
	t.Log("Success case description: upload object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleCreateUploadObjectTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpBeginTaskRequest{Request: &gfspserver.GfSpBeginTaskRequest_UploadObjectTask{
		UploadObjectTask: &gfsptask.GfSpUploadObjectTask{
			Task: &gfsptask.GfSpTask{
				Address: "mockAddress",
			},
			VirtualGroupFamilyId: 1,
			ObjectInfo:           mockObjectInfo,
		}}}
	result, err := g.GfSpBeginTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, gfsperrors.MakeGfSpError(nil), result.GetErr())
}

func TestGfSpBaseApp_GfSpBeginTaskSuccess2(t *testing.T) {
	t.Log("Success case description: resumable upload object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleCreateResumableUploadObjectTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpBeginTaskRequest{Request: &gfspserver.GfSpBeginTaskRequest_ResumableUploadObjectTask{
		ResumableUploadObjectTask: &gfsptask.GfSpResumableUploadObjectTask{
			Task: &gfsptask.GfSpTask{
				Address: "mockAddress",
			},
			VirtualGroupFamilyId: 1,
			ObjectInfo:           mockObjectInfo,
		}}}
	result, err := g.GfSpBeginTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, gfsperrors.MakeGfSpError(nil), result.GetErr())
}

func TestGfSpBaseApp_GfSpBeginTaskFailure1(t *testing.T) {
	t.Log("Success case description: pointer dangling")
	g := setup(t)
	result, err := g.GfSpBeginTask(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrUploadTaskDangling, result.GetErr())
}

func TestGfSpBaseApp_OnBeginUploadObjectTaskFailure1(t *testing.T) {
	t.Log("Failure case description: dangling pointer")
	g := setup(t)
	err := g.OnBeginUploadObjectTask(context.TODO(), nil)
	assert.Equal(t, ErrUploadTaskDangling, err)
}

func TestGfSpBaseApp_OnBeginUploadObjectTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to begin upload object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleCreateUploadObjectTask(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
	req := &gfsptask.GfSpUploadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		VirtualGroupFamilyId: 1,
		ObjectInfo:           mockObjectInfo,
	}
	err := g.OnBeginUploadObjectTask(context.TODO(), req)
	assert.Equal(t, mockErr, err)
}

func TestGfSpBaseApp_OnBeginResumableUploadObjectTaskFailure1(t *testing.T) {
	t.Log("Failure case description: dangling pointer")
	g := setup(t)
	err := g.OnBeginResumableUploadObjectTask(context.TODO(), nil)
	assert.Equal(t, ErrUploadTaskDangling, err)
}

func TestGfSpBaseApp_OnBeginResumableUploadObjectTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to begin resumable upload object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleCreateResumableUploadObjectTask(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
	req := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		VirtualGroupFamilyId: 1,
		ObjectInfo:           mockObjectInfo,
	}
	err := g.OnBeginResumableUploadObjectTask(context.TODO(), req)
	assert.Equal(t, mockErr, err)
}

func TestGfSpBaseApp_GfSpAskTaskSuccess1(t *testing.T) {
	t.Log("Success case description: replicate piece task with retry")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	replicatePieceTask := &gfsptask.GfSpReplicatePieceTask{
		Task:          &gfsptask.GfSpTask{Retry: 1},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(replicatePieceTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, replicatePieceTask, result.GetReplicatePieceTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess2(t *testing.T) {
	t.Log("Success case description: replicate piece task without retry")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	replicatePieceTask := &gfsptask.GfSpReplicatePieceTask{
		Task:          &gfsptask.GfSpTask{Retry: 0},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(replicatePieceTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, replicatePieceTask, result.GetReplicatePieceTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess3(t *testing.T) {
	t.Log("Success case description: seal object task with retry")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	sealObjectTask := &gfsptask.GfSpSealObjectTask{
		Task:          &gfsptask.GfSpTask{Retry: 1},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(sealObjectTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, sealObjectTask, result.GetSealObjectTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess4(t *testing.T) {
	t.Log("Success case description: seal object task without retry")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	sealObjectTask := &gfsptask.GfSpSealObjectTask{
		Task:          &gfsptask.GfSpTask{Retry: 0},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(sealObjectTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, sealObjectTask, result.GetSealObjectTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess5(t *testing.T) {
	t.Log("Success case description: receive piece task with retry")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	receivePieceTask := &gfsptask.GfSpReceivePieceTask{
		Task:          &gfsptask.GfSpTask{Retry: 1},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(receivePieceTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, receivePieceTask, result.GetReceivePieceTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess6(t *testing.T) {
	t.Log("Success case description: receive piece task without retry")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	receivePieceTask := &gfsptask.GfSpReceivePieceTask{
		Task:          &gfsptask.GfSpTask{Retry: 0},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(receivePieceTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, receivePieceTask, result.GetReceivePieceTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess7(t *testing.T) {
	t.Log("Success case description: gc object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	gcObjectTask := &gfsptask.GfSpGCObjectTask{
		Task: &gfsptask.GfSpTask{Retry: 0},
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(gcObjectTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, gcObjectTask, result.GetGcObjectTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess8(t *testing.T) {
	t.Log("Success case description: gc zombie piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	gcZombiePieceTask := &gfsptask.GfSpGCZombiePieceTask{
		Task: &gfsptask.GfSpTask{Retry: 0},
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(gcZombiePieceTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, gcZombiePieceTask, result.GetGcZombiePieceTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess9(t *testing.T) {
	t.Log("Success case description: gc meta task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	gcMetaTask := &gfsptask.GfSpGCMetaTask{
		Task: &gfsptask.GfSpTask{Retry: 0},
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(gcMetaTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, gcMetaTask, result.GetGcMetaTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess10(t *testing.T) {
	t.Log("Success case description: recover piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	recoverPieceTask := &gfsptask.GfSpRecoverPieceTask{
		Task:          &gfsptask.GfSpTask{Retry: 0},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(recoverPieceTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, recoverPieceTask, result.GetRecoverPieceTask())
}

func TestGfSpBaseApp_GfSpAskTaskSuccess11(t *testing.T) {
	t.Log("Success case description: migrate gvg task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{
		Task: &gfsptask.GfSpTask{Retry: 0},
	}
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(migrateGVGTask, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, migrateGVGTask, result.GetMigrateGvgTask())
}

func TestGfSpBaseApp_GfSpAskTaskFailure1(t *testing.T) {
	t.Log("Failure case description: failed to dispatch task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpAskTaskFailure2(t *testing.T) {
	t.Log("Failure case description: no task match limit")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().DispatchTask(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	req := &gfspserver.GfSpAskTaskRequest{NodeLimit: &gfsplimit.GfSpLimit{Memory: 1}}
	result, err := g.GfSpAskTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, ErrNoTaskMatchLimit, result.GetErr())
}

func TestGfSpBaseApp_GfSpReportTaskSuccess1(t *testing.T) {
	t.Log("Success case description: upload object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m1 := corespdb.NewMockSPDB(ctrl)
	g.gfSpDB = m1
	m1.EXPECT().InsertPutEvent(gomock.Any()).Return(nil)
	m.EXPECT().HandleDoneUploadObjectTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_UploadObjectTask{
		UploadObjectTask: &gfsptask.GfSpUploadObjectTask{
			Task:                 &gfsptask.GfSpTask{Retry: 0},
			VirtualGroupFamilyId: 1,
			ObjectInfo:           mockObjectInfo,
			StorageParams:        mockStorageParams,
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess2(t *testing.T) {
	t.Log("Success case description: resumable upload object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleDoneResumableUploadObjectTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_ResumableUploadObjectTask{
		ResumableUploadObjectTask: &gfsptask.GfSpResumableUploadObjectTask{
			Task:                 &gfsptask.GfSpTask{Retry: 0},
			VirtualGroupFamilyId: 1,
			ObjectInfo:           mockObjectInfo,
			StorageParams:        mockStorageParams,
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess3(t *testing.T) {
	t.Log("Success case description: replicate piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleReplicatePieceTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_ReplicatePieceTask{
		ReplicatePieceTask: &gfsptask.GfSpReplicatePieceTask{
			Task:          &gfsptask.GfSpTask{Retry: 0},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess4(t *testing.T) {
	t.Log("Success case description: seal object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleSealObjectTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_SealObjectTask{
		SealObjectTask: &gfsptask.GfSpSealObjectTask{
			Task:          &gfsptask.GfSpTask{Retry: 0},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess5(t *testing.T) {
	t.Log("Success case description: receive piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleReceivePieceTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_ReceivePieceTask{
		ReceivePieceTask: &gfsptask.GfSpReceivePieceTask{
			Task:          &gfsptask.GfSpTask{Retry: 0},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess6(t *testing.T) {
	t.Log("Success case description: gc object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleGCObjectTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_GcObjectTask{
		GcObjectTask: &gfsptask.GfSpGCObjectTask{
			Task: &gfsptask.GfSpTask{Retry: 0},
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess7(t *testing.T) {
	t.Log("Success case description: gc zombie piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleGCZombiePieceTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_GcZombiePieceTask{
		GcZombiePieceTask: &gfsptask.GfSpGCZombiePieceTask{
			Task: &gfsptask.GfSpTask{Retry: 0},
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess8(t *testing.T) {
	t.Log("Success case description: gc zombie piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleGCZombiePieceTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_GcZombiePieceTask{
		GcZombiePieceTask: &gfsptask.GfSpGCZombiePieceTask{
			Task: &gfsptask.GfSpTask{Retry: 0},
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess9(t *testing.T) {
	t.Log("Success case description: gc meta task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleGCMetaTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_GcMetaTask{
		GcMetaTask: &gfsptask.GfSpGCMetaTask{
			Task: &gfsptask.GfSpTask{Retry: 0},
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess10(t *testing.T) {
	t.Log("Success case description: download object task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleDownloadObjectTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_DownloadObjectTask{
		DownloadObjectTask: &gfsptask.GfSpDownloadObjectTask{
			Task:          &gfsptask.GfSpTask{Retry: 0},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess11(t *testing.T) {
	t.Log("Success case description: challenge piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleChallengePieceTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_ChallengePieceTask{
		ChallengePieceTask: &gfsptask.GfSpChallengePieceTask{
			Task:          &gfsptask.GfSpTask{Retry: 0},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess12(t *testing.T) {
	t.Log("Success case description: recover piece task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleRecoverPieceTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_RecoverPieceTask{
		RecoverPieceTask: &gfsptask.GfSpRecoverPieceTask{
			Task:          &gfsptask.GfSpTask{Retry: 0},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskSuccess13(t *testing.T) {
	t.Log("Success case description: migrate gvg task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleMigrateGVGTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_MigrateGvgTask{
		MigrateGvgTask: &gfsptask.GfSpMigrateGVGTask{
			Task: &gfsptask.GfSpTask{Retry: 0},
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReportTaskResponse{}, result)
}

func TestGfSpBaseApp_GfSpReportTaskFailure1(t *testing.T) {
	t.Log("Failure case description: failed to report migrate gvg task")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().HandleMigrateGVGTask(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
	req := &gfspserver.GfSpReportTaskRequest{Request: &gfspserver.GfSpReportTaskRequest_MigrateGvgTask{
		MigrateGvgTask: &gfsptask.GfSpMigrateGVGTask{
			Task: &gfsptask.GfSpTask{Retry: 0},
		}}}
	result, err := g.GfSpReportTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpPickVirtualGroupFamilySuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().PickVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(uint32(1), nil).Times(1)
	req := &gfspserver.GfSpPickVirtualGroupFamilyRequest{CreateBucketApprovalTask: &gfsptask.GfSpCreateBucketApprovalTask{
		Task:             &gfsptask.GfSpTask{Retry: 0},
		CreateBucketInfo: mockCreateBucketInfo,
	}}
	result, err := g.GfSpPickVirtualGroupFamily(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), result.GetVgfId())
}

func TestGfSpBaseApp_GfSpPickVirtualGroupFamilyFailure(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().PickVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(uint32(0), mockErr).Times(1)
	req := &gfspserver.GfSpPickVirtualGroupFamilyRequest{CreateBucketApprovalTask: &gfsptask.GfSpCreateBucketApprovalTask{
		Task:             &gfsptask.GfSpTask{Retry: 0},
		CreateBucketInfo: mockCreateBucketInfo,
	}}
	result, err := g.GfSpPickVirtualGroupFamily(context.TODO(), req)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_GfSpNotifyMigrateSwapOutSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().NotifyMigrateSwapOut(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpNotifyMigrateSwapOutRequest{SwapOut: &virtual_types.MsgSwapOut{
		StorageProvider: "mockStorageProvider",
	}}
	result, err := g.GfSpNotifyMigrateSwapOut(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpNotifyMigrateSwapOutResponse{}, result)
}

func TestGfSpBaseApp_GfSpNotifyMigrateSwapOutFailure(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockManager(ctrl)
	g.manager = m
	m.EXPECT().NotifyMigrateSwapOut(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
	req := &gfspserver.GfSpNotifyMigrateSwapOutRequest{SwapOut: &virtual_types.MsgSwapOut{
		StorageProvider: "mockStorageProvider",
	}}
	result, err := g.GfSpNotifyMigrateSwapOut(context.TODO(), req)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}
