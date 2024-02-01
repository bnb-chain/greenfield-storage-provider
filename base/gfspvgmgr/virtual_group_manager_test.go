package gfspvgmgr

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	corevgmgr "github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
)

func Test_addVirtualGroupFamily(t *testing.T) {
	cases := []struct {
		name string
		vgf  *corevgmgr.VirtualGroupFamilyMeta
	}{
		{
			name: "1",
			vgf:  &corevgmgr.VirtualGroupFamilyMeta{FamilyStakingStorageSize: 0},
		},
		{
			name: "2",
			vgf: &corevgmgr.VirtualGroupFamilyMeta{
				FamilyUsedStorageSize:    1,
				FamilyStakingStorageSize: 100,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f := FreeStorageSizeWeightPicker{freeStorageSizeWeightMap: map[uint32]float64{1: 0.5}}
			f.addVirtualGroupFamily(tt.vgf)
		})
	}
}

func Test_addGlobalVirtualGroup(t *testing.T) {
	cases := []struct {
		name string
		vgf  *corevgmgr.GlobalVirtualGroupMeta
	}{
		{
			name: "1",
			vgf:  &corevgmgr.GlobalVirtualGroupMeta{},
		},
		{
			name: "2",
			vgf: &corevgmgr.GlobalVirtualGroupMeta{
				UsedStorageSize:    1,
				StakingStorageSize: 100,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f := FreeStorageSizeWeightPicker{freeStorageSizeWeightMap: map[uint32]float64{1: 0.5}}
			f.addGlobalVirtualGroup(tt.vgf)
		})
	}
}

func Test_pickIndex(t *testing.T) {
	cases := []struct {
		name         string
		picker       *FreeStorageSizeWeightPicker
		wantedResult uint32
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "1",
			picker:       &FreeStorageSizeWeightPicker{},
			wantedResult: 0,
			wantedIsErr:  true,
			wantedErr:    errors.New("failed to pick weighted random index"),
		},
		{
			name:         "2",
			picker:       &FreeStorageSizeWeightPicker{freeStorageSizeWeightMap: map[uint32]float64{1: 0.5}},
			wantedResult: 1,
			wantedIsErr:  false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.picker.pickIndex()
			assert.Equal(t, result, tt.wantedResult)
			if tt.wantedIsErr {
				assert.Equal(t, tt.wantedErr, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_pickVirtualGroupFamilySuccess(t *testing.T) {
	gvg := make(map[uint32]*corevgmgr.GlobalVirtualGroupMeta)
	gvg[1] = &corevgmgr.GlobalVirtualGroupMeta{
		ID:       1,
		FamilyID: 1,
	}
	vgfm := &virtualGroupFamilyManager{vgfIDToVgf: map[uint32]*corevgmgr.VirtualGroupFamilyMeta{
		1: {FamilyUsedStorageSize: 1, FamilyStakingStorageSize: 100, GVGMap: gvg}}}
	filter := corevgmgr.NewPickVGFFilter([]uint32{1, 2})
	result, err := vgfm.pickVirtualGroupFamily(filter, corevgmgr.NewPickVGFByGVGFilter([]uint32{2, 3}), nil)
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func Test_pickVirtualGroupFamilyFailure(t *testing.T) {
	vgfm := &virtualGroupFamilyManager{vgfIDToVgf: map[uint32]*corevgmgr.VirtualGroupFamilyMeta{}}
	filter := &corevgmgr.PickVGFFilter{AvailableVgfIDSet: map[uint32]struct{}{}}
	result, err := vgfm.pickVirtualGroupFamily(filter, corevgmgr.NewPickVGFByGVGFilter([]uint32{1, 2}), nil)
	assert.Equal(t, ErrFailedPickVGF, err)
	assert.Nil(t, result)
}

func TestHealthChecker_CheckAllSPHealth(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	ctrl := gomock.NewController(t)
	con := consensus.NewMockConsensus(ctrl)
	hc := NewHealthChecker(con)
	hc.addAllSP([]*sptypes.StorageProvider{
		{
			Id:       1,
			Status:   sptypes.STATUS_IN_SERVICE,
			Endpoint: testServer.URL,
		},
		{
			Id:       2,
			Status:   sptypes.STATUS_GRACEFUL_EXITING,
			Endpoint: testServer.URL,
		},
	})
	con.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
		VersionedParams: storagetypes.VersionedParams{
			RedundantDataChunkNum:   0,
			RedundantParityChunkNum: 0,
		},
	}, nil).AnyTimes()
	hc.checkAllSPHealth()
}

func TestHealthChecker_CheckAllSPHealth1(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	ctrl := gomock.NewController(t)
	con := consensus.NewMockConsensus(ctrl)
	hc := NewHealthChecker(con)
	hc.checkAllSPHealth()
}

func TestHealthChecker_IsVGFHealthy(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	ctrl := gomock.NewController(t)
	con := consensus.NewMockConsensus(ctrl)
	hc := NewHealthChecker(con)
	hc.addAllSP([]*sptypes.StorageProvider{
		{
			Id:       1,
			Status:   sptypes.STATUS_IN_SERVICE,
			Endpoint: testServer.URL,
		},
		{
			Id:       2,
			Status:   sptypes.STATUS_GRACEFUL_EXITING,
			Endpoint: testServer.URL,
		},
	})

	res := hc.isVGFHealthy(&vgmgr.VirtualGroupFamilyMeta{
		GVGMap: map[uint32]*corevgmgr.GlobalVirtualGroupMeta{
			1: {
				SecondarySPIDs: []uint32{1, 2},
			},
		},
	})
	assert.Equal(t, res, true)
}

func TestHealthChecker_IsVGFHealthy1(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	ctrl := gomock.NewController(t)
	con := consensus.NewMockConsensus(ctrl)
	hc := NewHealthChecker(con)
	hc.addAllSP([]*sptypes.StorageProvider{
		{
			Id:       1,
			Status:   sptypes.STATUS_IN_SERVICE,
			Endpoint: testServer.URL,
		},
		{
			Id:       2,
			Status:   sptypes.STATUS_GRACEFUL_EXITING,
			Endpoint: testServer.URL,
		},
	})
	hc.unhealthySPs[1] = &sptypes.StorageProvider{
		Id:       1,
		Status:   sptypes.STATUS_IN_SERVICE,
		Endpoint: testServer.URL,
	}

	hc.unhealthySPs[2] = &sptypes.StorageProvider{
		Id:       2,
		Status:   sptypes.STATUS_IN_SERVICE,
		Endpoint: testServer.URL,
	}

	res := hc.isVGFHealthy(&vgmgr.VirtualGroupFamilyMeta{
		GVGMap: map[uint32]*corevgmgr.GlobalVirtualGroupMeta{
			1: {
				SecondarySPIDs: []uint32{1, 2},
			},
		},
	})
	assert.Equal(t, res, false)
}
