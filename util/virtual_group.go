package util

import (
	"context"
	"errors"
	"math"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
)

var (
	// ErrNotInSecondarySPs define the specified sp does not exist error.
	ErrNotInSecondarySPs = errors.New("target sp is not in the gvg secondary sp list")
)

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
func ValidateAndGetSPIndexWithinGVGSecondarySPs(ctx context.Context, client *gfspclient.GfSpClient, selfSpID uint32,
	bucketID uint64, lvgID uint32) (int, bool, error) {
	log.Infow("print info", "bucketID", bucketID, "selfSpID", selfSpID, "lvgID", lvgID)
	gvg, err := client.GetGlobalVirtualGroup(ctx, bucketID, lvgID)
	log.Infow("print gvg", "gvg", gvg)
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

// BlsAggregate aggregate secondary sp bls signature
func BlsAggregate(secondarySigs [][]byte) ([]byte, error) {
	blsSigs, err := bls.MultipleSignaturesFromBytes(secondarySigs)
	if err != nil {
		return nil, err
	}
	aggBlsSig := bls.AggregateSignatures(blsSigs).Marshal()
	return aggBlsSig, nil
}
