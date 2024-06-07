package util

import (
	"context"
	"errors"
	"math"

	sdkmath "cosmossdk.io/math"
	"github.com/prysmaticlabs/prysm/crypto/bls"

	storagetypes "github.com/evmos/evmos/v12/x/storage/types"
	virtualgrouptypes "github.com/evmos/evmos/v12/x/virtualgroup/types"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspclient"
	"github.com/zkMeLabs/mechain-storage-provider/core/consensus"
)

// ErrNotInSecondarySPs define the specified sp does not exist error.
var ErrNotInSecondarySPs = errors.New("target sp is not in the gvg secondary sp list")

// GetSecondarySPIndexFromGVG returns secondary sp index in the secondary sp list.
func GetSecondarySPIndexFromGVG(gvg *virtualgrouptypes.GlobalVirtualGroup, spID uint32) (int32, error) {
	for index, secondarySPID := range gvg.GetSecondarySpIds() {
		if secondarySPID == spID {
			return int32(index), nil
		}
	}
	return -1, ErrNotInSecondarySPs
}

// TotalStakingStoreSizeOfGVG calculates the global virtual group total staking storage size
func TotalStakingStoreSizeOfGVG(gvg *virtualgrouptypes.GlobalVirtualGroup, stakingPerBytes sdkmath.Int) uint64 {
	total := gvg.TotalDeposit.Quo(stakingPerBytes)
	if !total.IsUint64() {
		return math.MaxUint64
	}
	return total.Uint64()
}

// ValidateAndGetSPIndexWithinGVGSecondarySPs return whether current sp is one of the object gvg's secondary sp and its index within GVG(if is)
func ValidateAndGetSPIndexWithinGVGSecondarySPs(ctx context.Context, client gfspclient.GfSpClientAPI, selfSpID uint32,
	bucketID uint64, lvgID uint32,
) (int, bool, error) {
	gvg, err := client.GetGlobalVirtualGroup(ctx, bucketID, lvgID)
	if err != nil {
		return -1, false, err
	}
	for i, sspID := range gvg.GetSecondarySpIds() {
		if selfSpID == sspID {
			return i, true, nil
		}
	}
	return -1, false, nil
}

// ValidateSecondarySPs returns whether current sp is one of the object gvg's secondary sp and its index within GVG(if is)
func ValidateSecondarySPs(selfSpID uint32, secondarySpIDs []uint32) (int, bool) {
	for i, sspID := range secondarySpIDs {
		if selfSpID == sspID {
			return i, true
		}
	}
	return -1, false
}

// ValidatePrimarySP returns whether selfSpID is primarySpID
func ValidatePrimarySP(selfSpID, primarySpID uint32) bool {
	return selfSpID == primarySpID
}

// BlsAggregate aggregate secondary sp bls signature
func BlsAggregate(secondarySigs [][]byte) ([]byte, error) {
	blsSigs, err := bls.MultipleSignaturesFromBytes(secondarySigs)
	if err != nil {
		return nil, err
	}
	aggBlsSig := bls.AggregateSignatures(blsSigs).Marshal()
	return aggBlsSig, nil
}

// GetBucketPrimarySPID return bucket sp id by vgf id
func GetBucketPrimarySPID(ctx context.Context, chainClient consensus.Consensus, bucketInfo *storagetypes.BucketInfo) (uint32, error) {
	resp, err := chainClient.QueryVirtualGroupFamily(ctx, bucketInfo.GetGlobalVirtualGroupFamilyId())
	if err != nil {
		return 0, err
	}
	return resp.GetPrimarySpId(), nil
}
