package util

import (
	"context"
	"errors"
	"math"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func TestGetSecondarySPIndexFromGVG(t *testing.T) {
	cases := []struct {
		name         string
		gvg          *virtualgrouptypes.GlobalVirtualGroup
		spID         uint32
		wantedResult int32
		wantedErr    error
	}{
		{
			name: "In secondary SPs",
			gvg: &virtualgrouptypes.GlobalVirtualGroup{
				SecondarySpIds: []uint32{1, 2, 3},
			},
			spID:         1,
			wantedResult: 0,
			wantedErr:    nil,
		},
		{
			name: "Not in secondary SPs",
			gvg: &virtualgrouptypes.GlobalVirtualGroup{
				SecondarySpIds: []uint32{1, 2, 3},
			},
			spID:         4,
			wantedResult: -1,
			wantedErr:    ErrNotInSecondarySPs,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetSecondarySPIndexFromGVG(tt.gvg, tt.spID)
			assert.Equal(t, tt.wantedResult, result)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestValidateAndGetSPIndexWithinGVGSecondarySPsSuccess1(t *testing.T) {
	t.Log("Success case description: current sp is one of the object gvg's secondary sp")
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, bucketID uint64, lvgID uint32, opts ...grpc.DialOption) (
			*virtualgrouptypes.GlobalVirtualGroup, error) {
			return &virtualgrouptypes.GlobalVirtualGroup{SecondarySpIds: []uint32{1, 2, 3, 4, 5}}, nil
		}).AnyTimes()
	result1, result2, err := ValidateAndGetSPIndexWithinGVGSecondarySPs(context.Background(), m, 3, 1, 2)
	assert.Equal(t, 2, result1)
	assert.Equal(t, true, result2)
	assert.Nil(t, err)
}

func TestValidateAndGetSPIndexWithinGVGSecondarySPsSuccess2(t *testing.T) {
	t.Log("Success case description: current sp is not one of the object gvg's secondary sp")
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, bucketID uint64, lvgID uint32, opts ...grpc.DialOption) (
			*virtualgrouptypes.GlobalVirtualGroup, error) {
			return &virtualgrouptypes.GlobalVirtualGroup{SecondarySpIds: []uint32{1, 2, 3, 4, 5}}, nil
		}).AnyTimes()
	result1, result2, err := ValidateAndGetSPIndexWithinGVGSecondarySPs(context.Background(), m, 8, 1, 2)
	assert.Equal(t, -1, result1)
	assert.Equal(t, false, result2)
	assert.Nil(t, err)
}

func TestValidateAndGetSPIndexWithinGVGSecondarySPsFailure(t *testing.T) {
	t.Log("Failure case description: call GetGlobalVirtualGroup returns error")
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, bucketID uint64, lvgID uint32, opts ...grpc.DialOption) (
			*virtualgrouptypes.GlobalVirtualGroup, error) {
			return nil, errors.New("mock error")
		}).AnyTimes()
	result1, result2, err := ValidateAndGetSPIndexWithinGVGSecondarySPs(context.Background(), m, 8, 1, 2)
	assert.Equal(t, -1, result1)
	assert.Equal(t, false, result2)
	assert.Equal(t, errors.New("mock error"), err)
}

func TestTotalStakingStoreSizeOfGVG(t *testing.T) {
	cases := []struct {
		name            string
		gvg             *virtualgrouptypes.GlobalVirtualGroup
		stakingPerBytes sdkmath.Int
		wantedResult    uint64
	}{
		{
			name: "Return right result",
			gvg: &virtualgrouptypes.GlobalVirtualGroup{
				TotalDeposit: sdkmath.NewInt(100),
			},
			stakingPerBytes: sdkmath.NewInt(10),
			wantedResult:    10,
		},
		{
			name: "Return math.MaxUint64",
			gvg: &virtualgrouptypes.GlobalVirtualGroup{
				TotalDeposit: sdkmath.NewInt(-1),
			},
			stakingPerBytes: sdkmath.NewInt(1),
			wantedResult:    math.MaxUint64,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := TotalStakingStoreSizeOfGVG(tt.gvg, tt.stakingPerBytes)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestValidateSecondarySPs(t *testing.T) {
	cases := []struct {
		name           string
		selfSpID       uint32
		secondarySpIDs []uint32
		wantedResult1  int
		wantedResult2  bool
	}{
		{
			name:           "Is secondary sp",
			selfSpID:       1,
			secondarySpIDs: []uint32{1, 2, 3},
			wantedResult1:  0,
			wantedResult2:  true,
		},
		{
			name:           "Not secondary sp",
			selfSpID:       4,
			secondarySpIDs: []uint32{1, 2, 3},
			wantedResult1:  -1,
			wantedResult2:  false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2 := ValidateSecondarySPs(tt.selfSpID, tt.secondarySpIDs)
			assert.Equal(t, tt.wantedResult1, result1)
			assert.Equal(t, tt.wantedResult2, result2)
		})
	}
}

func TestValidatePrimarySP(t *testing.T) {
	result := ValidatePrimarySP(1, 1)
	assert.Equal(t, result, true)
}

func TestBlsAggregate(t *testing.T) {
	cases := []struct {
		name          string
		secondarySigs [][]byte
		wantedResult  []byte
		wantedIsErr   bool
	}{
		{
			name: "Aggregate bls signature correctly",
			secondarySigs: [][]byte{
				{0xab, 0xb0, 0x12, 0x4c, 0x75, 0x74, 0xf2, 0x81, 0xa2, 0x93, 0xf4, 0x18, 0x5c, 0xad, 0x3c, 0xb2, 0x26, 0x81, 0xd5, 0x20, 0x91, 0x7c, 0xe4, 0x66, 0x65, 0x24, 0x3e, 0xac, 0xb0, 0x51, 0x00, 0x0d, 0x8b, 0xac, 0xf7, 0x5e, 0x14, 0x51, 0x87, 0x0c, 0xa6, 0xb3, 0xb9, 0xe6, 0xc9, 0xd4, 0x1a, 0x7b, 0x02, 0xea, 0xd2, 0x68, 0x5a, 0x84, 0x18, 0x8a, 0x4f, 0xaf, 0xd3, 0x82, 0x5d, 0xaf, 0x6a, 0x98, 0x96, 0x25, 0xd7, 0x19, 0xcc, 0xd2, 0xd8, 0x3a, 0x40, 0x10, 0x1f, 0x4a, 0x45, 0x3f, 0xca, 0x62, 0x87, 0x8c, 0x89, 0x0e, 0xca, 0x62, 0x23, 0x63, 0xf9, 0xdd, 0xb8, 0xf3, 0x67, 0xa9, 0x1e, 0x84},
				{0xb7, 0x86, 0xe5, 0x7, 0x43, 0xe2, 0x53, 0x6c, 0x15, 0x51, 0x9c, 0x6, 0x2a, 0xa7, 0xe5, 0x12, 0xf9, 0xb7, 0x77, 0x93, 0x3f, 0x55, 0xb3, 0xaf, 0x38, 0xf7, 0x39, 0xe4, 0x84, 0x6d, 0x88, 0x44, 0x52, 0x77, 0x65, 0x42, 0x95, 0xd9, 0x79, 0x93, 0x7e, 0xc8, 0x12, 0x60, 0xe3, 0x24, 0xea, 0x8, 0x10, 0x52, 0xcd, 0xd2, 0x7f, 0x5d, 0x25, 0x3a, 0xa8, 0x9b, 0xb7, 0x65, 0xa9, 0x31, 0xea, 0x7c, 0x85, 0x13, 0x53, 0xc0, 0xa3, 0x88, 0xd1, 0xa5, 0x54, 0x85, 0x2, 0x2d, 0xf8, 0xa1, 0xd7, 0xc1, 0x60, 0x58, 0x93, 0xec, 0x7c, 0xf9, 0x33, 0x43, 0x4, 0x48, 0x40, 0x97, 0xef, 0x67, 0x2a, 0x27},
				{0xb2, 0x12, 0xd0, 0xec, 0x46, 0x76, 0x6b, 0x24, 0x71, 0x91, 0x2e, 0xa8, 0x53, 0x9a, 0x48, 0xa3, 0x78, 0x30, 0xc, 0xe8, 0xf0, 0x86, 0xa3, 0x68, 0xec, 0xe8, 0x96, 0x43, 0x34, 0xda, 0xf, 0xf4, 0x65, 0x48, 0xbb, 0xe0, 0x92, 0xa1, 0x8, 0x12, 0x18, 0x46, 0xe6, 0x4a, 0xd6, 0x92, 0x88, 0xe, 0x2, 0xf5, 0xf3, 0x2a, 0x96, 0xb1, 0x4, 0xf1, 0x11, 0xa9, 0x92, 0x79, 0x52, 0x0, 0x64, 0x34, 0xeb, 0x25, 0xe, 0xf4, 0x29, 0x6b, 0x39, 0x4e, 0x28, 0x78, 0xfe, 0x25, 0xa3, 0xc0, 0x88, 0x5a, 0x40, 0xfd, 0x71, 0x37, 0x63, 0x79, 0xcd, 0x6b, 0x56, 0xda, 0xee, 0x91, 0x26, 0x72, 0xfc, 0xbc},
				{0x8f, 0xc0, 0xb4, 0x9e, 0x2e, 0xac, 0x50, 0x86, 0xe2, 0xe2, 0xaa, 0xf, 0xdc, 0x54, 0x23, 0x51, 0x6, 0xd8, 0x29, 0xf5, 0xae, 0x3, 0x5d, 0xb8, 0x31, 0x4d, 0x26, 0x3, 0x48, 0x18, 0xb9, 0x1f, 0x6b, 0xd7, 0x86, 0xb4, 0xa2, 0x69, 0xc7, 0xe7, 0xf5, 0xc0, 0x93, 0x19, 0x6e, 0xfd, 0x33, 0xb8, 0x1, 0xe1, 0x1f, 0x4e, 0xb4, 0xb1, 0xa0, 0x1, 0x30, 0x48, 0x8a, 0x6c, 0x97, 0x29, 0xd6, 0xcb, 0x1c, 0x45, 0xef, 0x87, 0xba, 0x4f, 0xce, 0x22, 0x84, 0x48, 0xad, 0x16, 0xf7, 0x5c, 0xb2, 0xa8, 0x34, 0xb9, 0xee, 0xb8, 0xbf, 0xe5, 0x58, 0x2c, 0x44, 0x7b, 0x1f, 0x9c, 0x22, 0x26, 0x3a, 0x22},
			},
			wantedIsErr: false,
		},
		{
			name:          "Cannot aggregate bls signature",
			secondarySigs: [][]byte{[]byte{1}},
			wantedIsErr:   true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BlsAggregate(tt.secondarySigs)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetBucketPrimarySPIDSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := consensus.NewMockConsensus(ctrl)
	m.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, vgfID uint32) (*virtualgrouptypes.GlobalVirtualGroupFamily, error) {
			return &virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil
		}).AnyTimes()
	result, err := GetBucketPrimarySPID(context.Background(), m, &storagetypes.BucketInfo{
		Owner:      "mockUser",
		BucketName: "mockBucket",
	})
	assert.Equal(t, uint32(1), result)
	assert.Nil(t, err)
}

func TestGetBucketPrimarySPIDFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := consensus.NewMockConsensus(ctrl)
	m.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, vgfID uint32) (*virtualgrouptypes.GlobalVirtualGroupFamily, error) {
			return nil, errors.New("failed to call rpc")
		}).AnyTimes()
	result, err := GetBucketPrimarySPID(context.Background(), m, &storagetypes.BucketInfo{
		Owner:      "mockUser",
		BucketName: "mockBucket",
	})
	assert.Equal(t, uint32(0), result)
	assert.NotNil(t, err)
}
