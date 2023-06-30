package util

import (
	"errors"
	"math"

	sdkmath "cosmossdk.io/math"

	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
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
