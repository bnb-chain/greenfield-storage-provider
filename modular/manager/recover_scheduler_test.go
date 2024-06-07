package manager

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	types0 "github.com/evmos/evmos/v12/x/storage/types"
	types1 "github.com/evmos/evmos/v12/x/virtualgroup/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/zkMeLabs/mechain-storage-provider/base/gfspclient"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfsptqueue"
	"github.com/zkMeLabs/mechain-storage-provider/core/consensus"
	"github.com/zkMeLabs/mechain-storage-provider/core/spdb"
	"github.com/zkMeLabs/mechain-storage-provider/modular/metadata/types"
)

func TestManageModular_RecoverVGFScheduler(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)

	con := consensus.NewMockConsensus(ctrl)
	m.baseApp.SetConsensus(con)
	con.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&types1.GlobalVirtualGroupFamily{
		Id:                    1,
		PrimarySpId:           1,
		GlobalVirtualGroupIds: []uint32{2, 3, 4, 5, 6, 7},
	}, nil).AnyTimes()

	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().SetRecoverGVGStats(gomock.Any()).Return(nil).AnyTimes()

	recoverVGF, err := NewRecoverVGFScheduler(m, 1)
	assert.Equal(t, nil, err)

	m.recoveryQueue = gfsptqueue.NewGfSpTQueueWithLimit("test", 2)
	con.EXPECT().QueryStorageParams(gomock.Any()).Return(&types0.Params{
		VersionedParams: types0.VersionedParams{
			MaxSegmentSize: 10,
		},
	}, nil).AnyTimes()
	db.EXPECT().GetRecoverGVGStats(gomock.Any()).Return(&spdb.RecoverGVGStats{
		Status:          spdb.Processing,
		RedundancyIndex: 1,
		Limit:           10,
	}, nil).AnyTimes()
	spClient := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.baseApp.SetGfSpClient(spClient)
	spClient.EXPECT().ListObjectsInGVG(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*types.ObjectDetails{}, nil).AnyTimes()
	db.EXPECT().UpdateRecoverGVGStats(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().InsertRecoverFailedObject(gomock.Any()).Return(nil).AnyTimes()
	m.recoverObjectStats = NewObjectsSegmentsStats()
	db.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(&spdb.IntegrityMeta{}, nil).AnyTimes()
	db.EXPECT().GetRecoverFailedObject(gomock.Any()).Return(&spdb.RecoverFailedObject{
		RetryTime: 5,
	}, nil).AnyTimes()
	con.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	spClient.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&types0.ObjectInfo{Id: sdkmath.NewUint(1)}, nil).AnyTimes()

	recoverVGF.Start()
}

func TestManageModular_RecoverVGFScheduler1(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)

	con := consensus.NewMockConsensus(ctrl)
	m.baseApp.SetConsensus(con)
	con.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&types1.GlobalVirtualGroupFamily{
		Id:                    1,
		PrimarySpId:           1,
		GlobalVirtualGroupIds: []uint32{},
	}, nil).AnyTimes()

	spClient := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.baseApp.SetGfSpClient(spClient)
	spClient.EXPECT().CompleteSwapIn(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	_, err := NewRecoverVGFScheduler(m, 1)
	assert.Equal(t, nil, err)
}

func TestManageModular_ObjectsSegmentsStats(t *testing.T) {
	stats := NewObjectsSegmentsStats()
	stats.put(1, 1)
	has := stats.has(1)
	assert.Equal(t, true, has)
	stats.addSegmentRecord(1, true, 1)
	stats.addSegmentRecord(1, false, 1)
	l := stats.isObjectProcessed(1)
	assert.Equal(t, true, l)
	isFailed := stats.isRecoverFailed(1)
	assert.Equal(t, false, isFailed)
	stats.remove(1)
}

func TestVerifyGVGScheduler_Start(t *testing.T) {
	m := setup(t)
	ctrl := gomock.NewController(t)
	db := spdb.NewMockSPDB(ctrl)
	m.baseApp.SetGfSpDB(db)
	db.EXPECT().GetRecoverGVGStats(gomock.Any()).Return(&spdb.RecoverGVGStats{
		Status: spdb.Processed,
	}, nil).AnyTimes()
	spClient := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.baseApp.SetGfSpClient(spClient)
	spClient.EXPECT().ListObjectsInGVG(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	v := &VerifyGVGScheduler{
		manager:              m,
		gvgID:                1,
		redundancyIndex:      1,
		curStartAfter:        0,
		verifyFailedObjects:  make(map[uint64]struct{}, 0),
		verifySuccessObjects: make(map[uint64]struct{}, 0),
	}
	v.verifyFailedObjects[1] = struct{}{}
	db.EXPECT().GetRecoverFailedObject(gomock.Any()).Return(&spdb.RecoverFailedObject{
		ObjectID:  1,
		RetryTime: maxRecoveryRetry,
	}, nil).AnyTimes()
	con := consensus.NewMockConsensus(ctrl)
	m.baseApp.SetConsensus(con)
	con.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	spClient.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&types0.ObjectInfo{
		Id: sdkmath.NewUint(1),
	}, nil).AnyTimes()
	db.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(&spdb.IntegrityMeta{}, nil).AnyTimes()
	db.EXPECT().DeleteRecoverFailedObject(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().UpdateRecoverGVGStats(gomock.Any()).Return(nil).AnyTimes()
	v.Start()
}
