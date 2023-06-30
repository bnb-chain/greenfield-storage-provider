package util

import (
	"errors"

	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	// ErrNotInSecondarySPs define the specified sp does not exist error.
	ErrNotInSecondarySPs = errors.New("integer overflow")
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
