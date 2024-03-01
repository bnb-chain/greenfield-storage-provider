package manager

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	types0 "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
)

var mockErr = errors.New("mock error")

func setup(t *testing.T) *ManageModular {
	t.Helper()
	manager := &ManageModular{
		baseApp:                          &gfspapp.GfSpBaseApp{},
		uploadQueue:                      &gfsptqueue.GfSpTQueue{},
		resumableUploadQueue:             &gfsptqueue.GfSpTQueue{},
		replicateQueue:                   &gfsptqueue.GfSpTQueueWithLimit{},
		sealQueue:                        &gfsptqueue.GfSpTQueueWithLimit{},
		receiveQueue:                     &gfsptqueue.GfSpTQueueWithLimit{},
		gcObjectQueue:                    &gfsptqueue.GfSpTQueueWithLimit{},
		gcZombieQueue:                    &gfsptqueue.GfSpTQueueWithLimit{},
		gcMetaQueue:                      &gfsptqueue.GfSpTQueueWithLimit{},
		gcBucketMigrationQueue:           &gfsptqueue.GfSpTQueueWithLimit{},
		gcStaleVersionObjectQueue:        &gfsptqueue.GfSpTQueueWithLimit{},
		downloadQueue:                    &gfsptqueue.GfSpTQueue{},
		challengeQueue:                   &gfsptqueue.GfSpTQueue{},
		recoveryQueue:                    &gfsptqueue.GfSpTQueueWithLimit{},
		migrateGVGQueue:                  &gfsptqueue.GfSpTQueueWithLimit{},
		gcObjectTimeInterval:             2,
		gcZombiePieceTimeInterval:        3,
		gcMetaTimeInterval:               4,
		gcStaleVersionObjectTimeInterval: 2,
		syncConsensusInfoInterval:        5,
		syncAvailableVGFInterval:         5,
		statisticsOutputInterval:         6,
		discontinueBucketTimeInterval:    7,
		gcSafeBlockDistance:              1,
		backupTaskMux:                    sync.Mutex{},
	}

	return manager
}

func TestManagerModular_Name(t *testing.T) {
	e := setup(t)
	result := e.Name()
	assert.Equal(t, coremodule.ManageModularName, result)
}

func TestExecuteModular_StartSuccess(t *testing.T) {
	manage := setup(t)
	manage.statisticsOutputInterval = 1
	manage.gcObjectTimeInterval = 1
	manage.gcZombiePieceTimeInterval = 1
	manage.gcMetaTimeInterval = 1
	manage.syncConsensusInfoInterval = 1
	manage.discontinueBucketTimeInterval = 1
	manage.bucketMigrateScheduler = &BucketMigrateScheduler{}
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceManager(ctrl)
	manage.baseApp.SetResourceManager(m)
	m1 := corercmgr.NewMockResourceScope(ctrl)
	m.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (corercmgr.ResourceScope, error) {
		return m1, nil
	}).Times(1)
	m2 := consensus.NewMockConsensus(ctrl)
	manage.baseApp.SetConsensus(m2)

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	m3.EXPECT().ListGlobalVirtualGroupsBySecondarySP(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func() ([]*virtualgrouptypes.GlobalVirtualGroup, error) {
		return []*virtualgrouptypes.GlobalVirtualGroup{}, nil
	}).AnyTimes()

	m2.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{
		{Id: 1, Endpoint: "endpoint"}}, nil).AnyTimes()
	m2.EXPECT().CurrentHeight(gomock.Any()).Return(uint64(100), nil).AnyTimes()

	m4 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m4)
	m4.EXPECT().UpdateAllSp(gomock.Any()).Return(nil).AnyTimes()
	m4.EXPECT().SetOwnSpInfo(gomock.Any()).Return(nil).AnyTimes()

	err := manage.Start(context.TODO())
	assert.Nil(t, err)
	ctrl.Finish()
}

func TestManageModular_StartFailure(t *testing.T) {
	manage := setup(t)
	manage.statisticsOutputInterval = 1
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceManager(ctrl)
	manage.baseApp.SetResourceManager(m)
	m.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (corercmgr.ResourceScope, error) {
		return nil, mockErr
	}).AnyTimes()
	err := manage.Start(context.TODO())
	assert.Equal(t, mockErr, err)
	ctrl.Finish()
}

func TestManageModular_EventLoop(t *testing.T) {
	manage := setup(t)
	ctrl := gomock.NewController(t)

	m1 := consensus.NewMockConsensus(ctrl)
	manage.baseApp.SetConsensus(m1)
	m1.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{
		{Id: 1, Endpoint: "endpoint"}}, nil).AnyTimes()
	m1.EXPECT().CurrentHeight(gomock.Any()).Return(uint64(0), nil).AnyTimes()
	m2 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m2)
	m2.EXPECT().UpdateAllSp(gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().SetOwnSpInfo(gomock.Any()).Return(nil).AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	m3.EXPECT().GetLatestObjectID(gomock.Any()).Return(uint64(0), nil).AnyTimes()
	m3.EXPECT().ListExpiredBucketsBySp(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*types.Bucket{
			{
				BucketInfo: &types0.BucketInfo{},
			},
		}, nil).AnyTimes()
	m3.EXPECT().DiscontinueBucket(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	ctx, cancel := context.WithTimeout(context.TODO(), 10)
	manage.eventLoop(ctx)
	time.Sleep(10 * time.Second)
	cancel()
}

func TestManageModular_EventLoop1(t *testing.T) {
	manage := setup(t)
	manage.gcZombiePieceEnabled = true
	manage.gcMetaEnabled = true
	ctrl := gomock.NewController(t)

	m1 := consensus.NewMockConsensus(ctrl)
	manage.baseApp.SetConsensus(m1)
	m1.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{
		{Id: 1, Endpoint: "endpoint"}}, nil).AnyTimes()
	m1.EXPECT().CurrentHeight(gomock.Any()).Return(uint64(0), nil).AnyTimes()
	m2 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m2)
	m2.EXPECT().UpdateAllSp(gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().SetOwnSpInfo(gomock.Any()).Return(nil).AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	m3.EXPECT().GetLatestObjectID(gomock.Any()).Return(uint64(0), nil).AnyTimes()
	m3.EXPECT().ListExpiredBucketsBySp(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*types.Bucket{
			{
				BucketInfo: &types0.BucketInfo{},
			},
		}, nil).AnyTimes()
	m3.EXPECT().DiscontinueBucket(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	ctx, cancel := context.WithTimeout(context.TODO(), 10)
	manage.eventLoop(ctx)
	time.Sleep(10 * time.Second)
	cancel()
}

func TestManageModular_EventLoopErr1(t *testing.T) {
	manage := setup(t)
	ctrl := gomock.NewController(t)
	manage.gcObjectTimeInterval = 1
	m1 := consensus.NewMockConsensus(ctrl)
	manage.baseApp.SetConsensus(m1)
	m1.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{
		{Id: 1, Endpoint: "endpoint"}}, nil).AnyTimes()
	m1.EXPECT().CurrentHeight(gomock.Any()).Return(uint64(0), mockErr).AnyTimes()
	m2 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m2)
	m2.EXPECT().UpdateAllSp(gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().SetOwnSpInfo(gomock.Any()).Return(nil).AnyTimes()

	ctx, cancel := context.WithTimeout(context.TODO(), 10)
	manage.eventLoop(ctx)
	time.Sleep(10 * time.Second)
	cancel()
}

func TestManageModular_LoadTaskFromDB(t *testing.T) {
	manage := setup(t)
	ctrl := gomock.NewController(t)
	manage.enableLoadTask = true

	m1 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m1)
	m1.EXPECT().GetUploadMetasToReplicate(gomock.Any(), gomock.Any()).Return([]*spdb.UploadObjectMeta{
		{
			ObjectID:             1,
			GlobalVirtualGroupID: 1,
			SecondaryEndpoints:   []string{"endpoint"},
		},
		{
			ObjectID:           2,
			SecondaryEndpoints: []string{"endpoint"},
		},
	}, nil).AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	manage.baseApp.SetGfSpClient(m3)
	m3.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&types.Bucket{BucketInfo: &types0.BucketInfo{
			GlobalVirtualGroupFamilyId: 1,
		}}, nil)

	vgm := vgmgr.NewMockVirtualGroupManager(ctrl)
	manage.virtualGroupManager = vgm
	vgm.EXPECT().PickGlobalVirtualGroup(gomock.Any(), gomock.Any()).Return(&vgmgr.GlobalVirtualGroupMeta{
		ID:                   1,
		SecondarySPEndpoints: []string{"endpoint"},
	}, nil).AnyTimes()

	m2 := consensus.NewMockConsensus(ctrl)
	manage.baseApp.SetConsensus(m2)
	m2.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(&types0.ObjectInfo{
		ObjectStatus: types0.OBJECT_STATUS_CREATED,
	}, nil).AnyTimes()
	m2.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&types0.Params{}, nil).AnyTimes()

	m1.EXPECT().GetUploadMetasToSeal(gomock.Any(), gomock.Any()).Return([]*spdb.UploadObjectMeta{
		{
			ObjectID:             1,
			GlobalVirtualGroupID: 1,
			SecondaryEndpoints:   []string{"endpoint"},
		},
	}, nil).AnyTimes()

	m1.EXPECT().GetGCMetasToGC(gomock.Any()).Return([]*spdb.GCObjectMeta{
		{
			StartBlockHeight:    1,
			EndBlockHeight:      2,
			CurrentBlockHeight:  1,
			LastDeletedObjectID: 0,
		},
	}, nil).AnyTimes()
	manage.gcBlockHeight = 3
	err := manage.LoadTaskFromDB()
	assert.Equal(t, nil, err)
}

func TestManageModular_QueryRecoverProcessVGF(t *testing.T) {
	manage := setup(t)
	ctrl := gomock.NewController(t)

	m1 := consensus.NewMockConsensus(ctrl)
	manage.baseApp.SetConsensus(m1)
	m1.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		GlobalVirtualGroupIds: []uint32{1, 2, 3},
	}, nil).AnyTimes()

	m2 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m2)
	m2.EXPECT().BatchGetRecoverGVGStats(gomock.Any()).Return([]*spdb.RecoverGVGStats{
		{
			VirtualGroupID:       1,
			VirtualGroupFamilyID: 1,
			RedundancyIndex:      0,
			StartAfter:           0,
			NextStartAfter:       10,
			Limit:                10,
			Status:               spdb.Processing,
		},
		{
			VirtualGroupID:       2,
			VirtualGroupFamilyID: 1,
			RedundancyIndex:      1,
			StartAfter:           0,
			NextStartAfter:       10,
			Limit:                10,
			Status:               spdb.Processing,
		},
		{
			VirtualGroupID:       3,
			VirtualGroupFamilyID: 1,
			RedundancyIndex:      2,
			StartAfter:           0,
			NextStartAfter:       10,
			Limit:                10,
			Status:               spdb.Processing,
		},
	}, nil).AnyTimes()
	m2.EXPECT().GetRecoverFailedObjectsByRetryTime(gomock.Any()).Return([]*spdb.RecoverFailedObject{}, nil).AnyTimes()

	_, _, err := manage.QueryRecoverProcess(context.TODO(), 1, 0)
	assert.Equal(t, nil, err)
}

func TestManageModular_QueryRecoverProcessGVG(t *testing.T) {
	manage := setup(t)
	ctrl := gomock.NewController(t)

	m2 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m2)
	m2.EXPECT().BatchGetRecoverGVGStats(gomock.Any()).Return([]*spdb.RecoverGVGStats{
		{
			VirtualGroupID:       1,
			VirtualGroupFamilyID: 1,
			RedundancyIndex:      0,
			StartAfter:           0,
			NextStartAfter:       10,
			Limit:                10,
			Status:               spdb.Processing,
		},
	}, nil).AnyTimes()
	m2.EXPECT().GetRecoverFailedObjectsByRetryTime(gomock.Any()).Return([]*spdb.RecoverFailedObject{}, nil).AnyTimes()

	_, _, err := manage.QueryRecoverProcess(context.TODO(), 0, 1)
	assert.Equal(t, nil, err)
}

func TestManageModular_DelayStartMigrateScheduler(t *testing.T) {
	manage := setup(t)
	ctrl := gomock.NewController(t)
	manage.bucketMigrateScheduler = nil
	manage.subscribeBucketMigrateEventInterval = 300
	m1 := consensus.NewMockConsensus(ctrl)
	manage.baseApp.SetConsensus(m1)
	m1.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).AnyTimes()

	m2 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m2)
	m2.EXPECT().QueryBucketMigrateSubscribeProgress().Return(uint64(1), nil).AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	manage.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ListMigrateBucketEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*types.ListMigrateBucketEvents{
		{
			Event: &types0.EventMigrationBucket{
				BucketId:   sdkmath.NewUint(1),
				BucketName: "test",
			},
		},
	}, nil).AnyTimes()

	m1.EXPECT().QueryBucketInfoById(gomock.Any(), gomock.Any()).Return(&types0.BucketInfo{
		BucketStatus: types0.BUCKET_STATUS_CREATED,
	}, nil).AnyTimes()

	m2.EXPECT().ListBucketMigrationToConfirm(gomock.Any()).Return([]*spdb.MigrateBucketProgressMeta{
		{
			BucketID: 1,
		},
	}, nil).AnyTimes()

	m3.EXPECT().GetBucketSize(gomock.Any(), gomock.Any()).Return("10", nil).AnyTimes()
	m3.EXPECT().ListCompleteMigrationBucketEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*types0.EventCompleteMigrationBucket{}, nil).AnyTimes()
	m2.EXPECT().UpdateBucketMigrateGCSubscribeProgress(gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().UpdateBucketMigrateSubscribeProgress(gomock.Any()).Return(nil).AnyTimes()

	manage.delayStartMigrateScheduler()
}

func TestManageModular_QueryBucketMigrationProgress(t *testing.T) {
	manage := setup(t)
	ctrl := gomock.NewController(t)
	manage.bucketMigrateScheduler = &BucketMigrateScheduler{
		manager: manage,
	}

	m1 := spdb.NewMockSPDB(ctrl)
	manage.baseApp.SetGfSpDB(m1)
	m1.EXPECT().QueryMigrateBucketProgress(gomock.Any()).Return(&spdb.MigrateBucketProgressMeta{
		BucketID: 1,
	}, nil).AnyTimes()
	m1.EXPECT().ListMigrateGVGUnitsByBucketID(gomock.Any()).Return([]*spdb.MigrateGVGUnitMeta{
		{
			MigratedBytesSize: 100,
		},
	}, nil).AnyTimes()

	_, err := manage.QueryBucketMigrationProgress(context.TODO(), 1)
	assert.Equal(t, nil, err)
}

func TestManageModular_QueryTasksStats(t *testing.T) {
	manage := setup(t)
	//ctrl := gomock.NewController(t)

	uploadTasks, replicateCount, sealCount, resumableUploadCount, maxUploadCount, migrateGVGCount, recoveryProcessCount, _ := manage.QueryTasksStats(context.TODO())
	assert.Equal(t, 0, uploadTasks)
	assert.Equal(t, 0, replicateCount)
	assert.Equal(t, 0, sealCount)
	assert.Equal(t, 0, resumableUploadCount)
	assert.Equal(t, 0, maxUploadCount)
	assert.Equal(t, 0, migrateGVGCount)
	assert.Equal(t, 0, recoveryProcessCount)
}

func TestManageModular_RejectUnSealObject(t *testing.T) {
	manage := setup(t)
	ctrl := gomock.NewController(t)

	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	manage.baseApp.SetGfSpClient(m1)
	m1.EXPECT().RejectUnSealObject(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	m2 := consensus.NewMockConsensus(ctrl)
	manage.baseApp.SetConsensus(m2)
	m2.EXPECT().ListenRejectUnSealObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	err := manage.RejectUnSealObject(context.TODO(), &types0.ObjectInfo{
		BucketName: "test",
		ObjectName: "test",
		Id:         sdkmath.NewUint(1),
	})

	assert.Equal(t, nil, err)
}
