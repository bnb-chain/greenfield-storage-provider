package gfspvgmgr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

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
	vgfm := &virtualGroupFamilyManager{vgfIDToVgf: map[uint32]*corevgmgr.VirtualGroupFamilyMeta{
		1: {FamilyUsedStorageSize: 1, FamilyStakingStorageSize: 100}}}
	filter := corevgmgr.NewPickVGFFilter([]uint32{1, 2})
	result, err := vgfm.pickVirtualGroupFamily(filter, corevgmgr.NewExcludeIDFilter(corevgmgr.NewIDSetFromList([]uint32{2, 3})))
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func Test_pickVirtualGroupFamilyFailure(t *testing.T) {
	vgfm := &virtualGroupFamilyManager{vgfIDToVgf: map[uint32]*corevgmgr.VirtualGroupFamilyMeta{}}
	filter := &corevgmgr.PickVGFFilter{AvailableVgfIDSet: map[uint32]struct{}{}}
	result, err := vgfm.pickVirtualGroupFamily(filter, corevgmgr.NewExcludeIDFilter(corevgmgr.NewIDSetFromList([]uint32{1, 2})))
	assert.Equal(t, ErrFailedPickVGF, err)
	assert.Nil(t, result)
}
