package manager

import (
	"context"
	"errors"
	"testing"

	sdkmath "cosmossdk.io/math"
	types2 "github.com/bnb-chain/greenfield/x/sp/types"
	types0 "github.com/bnb-chain/greenfield/x/storage/types"
	types1 "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
)

func TestManageModular_HandleCreateUploadObjectTask(t *testing.T) {
	m := setup(t)
	m.maxUploadObjectNumber = 1
	m.uploadQueue = gfsptqueue.NewGfSpTQueue("test_upload", 2)
	ctrl := gomock.NewController(t)

	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().InsertUploadProgress(gomock.Any()).Return(nil).AnyTimes()

	uot := &gfsptask.GfSpUploadObjectTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
		},
	}
	err := m.HandleCreateUploadObjectTask(context.TODO(), uot)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleCreateUploadObjectTask1(t *testing.T) {
	m := setup(t)
	err := m.HandleCreateUploadObjectTask(context.TODO(), nil)
	assert.Equal(t, ErrDanglingTask, err)
}

func TestManageModular_HandleCreateUploadObjectTask2(t *testing.T) {
	m := setup(t)
	uot := &gfsptask.GfSpUploadObjectTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
		},
	}
	err := m.HandleCreateUploadObjectTask(context.TODO(), uot)
	assert.Equal(t, ErrExceedTask, err)
}

func TestManageModular_HandleCreateUploadObjectTask3(t *testing.T) {
	m := setup(t)
	m.maxUploadObjectNumber = 1
	m.uploadQueue = gfsptqueue.NewGfSpTQueue("test_upload", 2)
	ctrl := gomock.NewController(t)

	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().InsertUploadProgress(gomock.Any()).Return(errors.New("db Duplicate entry")).AnyTimes()

	uot := &gfsptask.GfSpUploadObjectTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
		},
	}
	err := m.HandleCreateUploadObjectTask(context.TODO(), uot)
	assert.Equal(t, nil, err)
}

func TestManageModular_PickGVGAndReplicate(t *testing.T) {
	m := setup(t)
	m.replicateQueue = gfsptqueue.NewGfSpTQueueWithLimit("test_upload", 2)
	ctrl := gomock.NewController(t)
	vgm := vgmgr.NewMockVirtualGroupManager(ctrl)
	m.virtualGroupManager = vgm
	vgm.EXPECT().PickGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return(&vgmgr.GlobalVirtualGroupMeta{}, nil).AnyTimes()
	con := consensus.NewMockConsensus(ctrl)
	m.baseApp.SetConsensus(con)
	con.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&types0.Params{}, nil).AnyTimes()
	vgm.EXPECT().GenerateGlobalVirtualGroupMeta(gomock.Any(), gomock.Any()).Return(&vgmgr.GlobalVirtualGroupMeta{}, nil).AnyTimes()
	con.EXPECT().QueryVirtualGroupParams(gomock.Any()).Return(&types1.Params{
		GvgStakingPerBytes: sdkmath.NewInt(100),
		DepositDenom:       "test",
	}, nil).AnyTimes()
	vgm.EXPECT().ForceRefreshMeta().Return(nil).AnyTimes()
	spClient := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.baseApp.SetGfSpClient(spClient)
	spClient.EXPECT().CreateGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	vgm.EXPECT().PickVirtualGroupFamily(gomock.Any()).Return(&vgmgr.VirtualGroupFamilyMeta{
		ID: uint32(1),
	}, nil).AnyTimes()
	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().UpdateUploadProgress(gomock.Any()).Return(nil).AnyTimes()
	uot := &gfsptask.GfSpUploadObjectTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
		},
		StorageParams: &types0.Params{},
	}
	err := m.pickGVGAndReplicate(context.TODO(), 1, uot)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleCreateResumableUploadObjectTask(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)
	m.maxUploadObjectNumber = 2
	m.resumableUploadQueue = gfsptqueue.NewGfSpTQueue("test", 2)

	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().InsertUploadProgress(gomock.Any()).Return(nil).AnyTimes()

	uploadTask := &gfsptask.GfSpResumableUploadObjectTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
		},
		StorageParams: &types0.Params{},
	}
	err := m.HandleCreateResumableUploadObjectTask(context.TODO(), uploadTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleDoneResumableUploadObjectTask(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)
	m.replicateQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)

	vgm := vgmgr.NewMockVirtualGroupManager(ctrl)
	m.virtualGroupManager = vgm
	vgm.EXPECT().PickGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return(&vgmgr.GlobalVirtualGroupMeta{}, nil).AnyTimes()
	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().UpdateUploadProgress(gomock.Any()).Return(nil).AnyTimes()

	uploadTask := &gfsptask.GfSpResumableUploadObjectTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
		},
		StorageParams: &types0.Params{},
		Completed:     true,
	}
	err := m.HandleDoneResumableUploadObjectTask(context.TODO(), uploadTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleReplicatePieceTask(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)
	m.replicateQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)
	m.sealQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)
	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().InsertPutEvent(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().UpdateUploadProgress(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().DeleteUploadProgress(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().UpdateUploadProgress(gomock.Any()).Return(nil).AnyTimes()

	replicatePieceTask := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
		},
		StorageParams: &types0.Params{},
		Sealed:        true,
	}
	err := m.HandleReplicatePieceTask(context.TODO(), replicatePieceTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleFailedReplicatePieceTask(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)
	replicatePieceTask := &gfsptask.GfSpReplicatePieceTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
			Retry:        2,
			MaxRetry:     3,
		},
		StorageParams:        &types0.Params{},
		NotAvailableSpIdx:    0,
		GlobalVirtualGroupId: 1,
	}
	replicateQueue := gfsptqueue.NewGfSpTQueueWithLimit("test", 2)
	m.replicateQueue = replicateQueue
	_ = m.replicateQueue.Push(replicatePieceTask)

	con := consensus.NewMockConsensus(ctrl)
	m.baseApp.SetConsensus(con)
	con.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(&types0.ObjectInfo{
		ObjectStatus: types0.OBJECT_STATUS_CREATED,
	}, nil).AnyTimes()
	con.EXPECT().QueryGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return(&types1.GlobalVirtualGroup{
		SecondarySpIds: []uint32{1},
	}, nil).AnyTimes()

	spClient := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.baseApp.SetGfSpClient(spClient)
	spClient.EXPECT().ListGlobalVirtualGroupsBySecondarySP(gomock.Any(), gomock.Any()).Return([]*types1.GlobalVirtualGroup{
		{
			PrimarySpId: 1,
		},
	}, nil).AnyTimes()
	con.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&types2.StorageProvider{
		Id: 1,
	}, nil).AnyTimes()
	vgm := vgmgr.NewMockVirtualGroupManager(ctrl)
	m.virtualGroupManager = vgm
	vgm.EXPECT().FreezeSPAndGVGs(gomock.Any(), gomock.Any()).AnyTimes()
	vgm.EXPECT().PickGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return(&vgmgr.GlobalVirtualGroupMeta{}, nil).AnyTimes()
	vgm.EXPECT().ForceRefreshMeta().Return(nil).AnyTimes()

	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().UpdateUploadProgress(gomock.Any()).Return(nil).AnyTimes()

	err := m.handleFailedReplicatePieceTask(context.TODO(), replicatePieceTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleSealObjectTask(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)
	m.sealQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)

	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().InsertPutEvent(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().UpdateUploadProgress(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().DeleteUploadProgress(gomock.Any()).Return(nil).AnyTimes()

	sealTask := &gfsptask.GfSpSealObjectTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
			Retry:        2,
			MaxRetry:     3,
		},
		StorageParams:        &types0.Params{},
		GlobalVirtualGroupId: 1,
	}

	err := m.HandleSealObjectTask(context.TODO(), sealTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleFailedSealObjectTask(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)
	sealTask := &gfsptask.GfSpSealObjectTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
			Retry:        2,
			MaxRetry:     3,
			Err:          ErrRepeatedTask,
		},
		StorageParams:        &types0.Params{},
		GlobalVirtualGroupId: 1,
	}
	m.sealQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)
	_ = m.sealQueue.Push(sealTask)

	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().InsertPutEvent(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().UpdateUploadProgress(gomock.Any()).Return(nil).AnyTimes()

	err := m.handleFailedSealObjectTask(context.TODO(), sealTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleReceivePieceTask(t *testing.T) {
	m := setup(t)

	receiveTask := &gfsptask.GfSpReceivePieceTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
			Retry:        2,
			MaxRetry:     3,
		},
		StorageParams:        &types0.Params{},
		GlobalVirtualGroupId: 1,
	}

	m.receiveQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)

	err := m.HandleReceivePieceTask(context.TODO(), receiveTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleReceivePieceTask1(t *testing.T) {
	m := setup(t)

	receiveTask := &gfsptask.GfSpReceivePieceTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
			Retry:        2,
			MaxRetry:     3,
		},
		StorageParams:        &types0.Params{},
		GlobalVirtualGroupId: 1,
		Sealed:               true,
	}

	m.receiveQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)

	err := m.HandleReceivePieceTask(context.TODO(), receiveTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleReceivePieceTask2(t *testing.T) {
	m := setup(t)

	receiveTask := &gfsptask.GfSpReceivePieceTask{
		ObjectInfo: &types0.ObjectInfo{
			Id:         sdkmath.NewUint(1),
			BucketName: "test",
			ObjectName: "test",
		},
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
			Retry:        2,
			MaxRetry:     3,
			Err:          ErrRepeatedTask,
		},
		StorageParams:        &types0.Params{},
		GlobalVirtualGroupId: 1,
	}

	m.receiveQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)
	m.receiveQueue.Push(receiveTask)
	err := m.HandleReceivePieceTask(context.TODO(), receiveTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_HandleGCObjectTask(t *testing.T) {
	m := setup(t)
	m.gcObjectQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)
	ctrl := gomock.NewController(t)
	gcTask := &gfsptask.GfSpGCObjectTask{
		Task: &gfsptask.GfSpTask{
			TaskPriority: 1,
			Retry:        2,
			MaxRetry:     3,
			Err:          ErrRepeatedTask,
		},
		CurrentBlockNumber: 1,
		EndBlockNumber:     10,
	}
	m.gcObjectQueue.Push(gcTask)
	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().UpdateGCObjectProgress(gomock.Any()).Return(nil).AnyTimes()

	err := m.HandleGCObjectTask(context.TODO(), gcTask)
	assert.Equal(t, nil, err)
}

func TestManageModular_PickVirtualGroupFamily(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)
	vgm := vgmgr.NewMockVirtualGroupManager(ctrl)
	m.virtualGroupManager = vgm
	vgm.EXPECT().PickVirtualGroupFamily(gomock.Any()).Return(nil, mockErr).Times(1)
	con := consensus.NewMockConsensus(ctrl)
	m.baseApp.SetConsensus(con)
	con.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&types0.Params{}, nil).AnyTimes()
	vgm.EXPECT().GenerateGlobalVirtualGroupMeta(gomock.Any(), gomock.Any()).Return(&vgmgr.GlobalVirtualGroupMeta{}, nil).AnyTimes()
	con.EXPECT().QueryVirtualGroupParams(gomock.Any()).Return(&types1.Params{
		GvgStakingPerBytes: sdkmath.NewInt(100),
		DepositDenom:       "test",
	}, nil).AnyTimes()
	spClient := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.baseApp.SetGfSpClient(spClient)
	spClient.EXPECT().CreateGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	vgm.EXPECT().ForceRefreshMeta().Return(nil).AnyTimes()
	vgm.EXPECT().PickVirtualGroupFamily(gomock.Any()).Return(&vgmgr.VirtualGroupFamilyMeta{
		ID: uint32(1),
	}, nil).AnyTimes()

	id, err := m.PickVirtualGroupFamily(context.TODO(), &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, uint32(1), id)
}
