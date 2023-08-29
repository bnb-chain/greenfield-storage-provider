package gfspvgmgr

import (
	"errors"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/stretchr/testify/assert"
)

func Test_addVirtualGroupFamily(t *testing.T) {
	cases := []struct {
		name string
		vgf  *vgmgr.VirtualGroupFamilyMeta
	}{
		{
			name: "1",
			vgf:  &vgmgr.VirtualGroupFamilyMeta{FamilyStakingStorageSize: 0},
		},
		{
			name: "2",
			vgf: &vgmgr.VirtualGroupFamilyMeta{
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
		vgf  *vgmgr.GlobalVirtualGroupMeta
	}{
		{
			name: "1",
			vgf:  &vgmgr.GlobalVirtualGroupMeta{},
		},
		{
			name: "2",
			vgf: &vgmgr.GlobalVirtualGroupMeta{
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

func Test_pickVirtualGroupFamily1(t *testing.T) {

}
