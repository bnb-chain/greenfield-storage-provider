package util

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
)

// GetSPID query sp id
func GetSPID(selfSPID uint32, chainClient consensus.Consensus, operatorAddress string) (uint32, error) {
	if selfSPID != 0 {
		return selfSPID, nil
	}
	spInfo, err := chainClient.QuerySP(context.Background(), operatorAddress)
	if err != nil {
		return 0, err
	}
	return spInfo.GetId(), nil
}
