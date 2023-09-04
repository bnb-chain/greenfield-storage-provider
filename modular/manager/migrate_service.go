package manager

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (m *ManageModular) getSPID() (uint32, error) {
	if m.spID != 0 {
		return m.spID, nil
	}
	spInfo, err := m.baseApp.Consensus().QuerySP(context.Background(), m.baseApp.OperatorAddress())
	if err != nil {
		return 0, err
	}
	m.spID = spInfo.GetId()
	return m.spID, nil
}

// NotifyMigrateSwapOut is used to receive migrate gvg task from src sp.
func (m *ManageModular) NotifyMigrateSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) error {
	var (
		err      error
		selfSPID uint32
	)
	if m.spExitScheduler == nil {
		log.CtxError(ctx, "sp exit scheduler has no init")
		return ErrNotifyMigrateSwapOut
	}
	selfSPID, err = m.getSPID()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get self sp id", "error", err)
		return err
	}
	if selfSPID != swapOut.SuccessorSpId {
		log.CtxErrorw(ctx, "successor sp id is mismatch", "self_sp_id", selfSPID, "swap_out_successor_sp_id", swapOut.SuccessorSpId)
		return fmt.Errorf("successor sp id is mismatch")
	}

	return m.spExitScheduler.AddSwapOutToTaskRunner(swapOut)
}

func (m *ManageModular) QuerySPByOperatorAddress(ctx context.Context, operatorAddress string) (*sptypes.StorageProvider, error) {
	return m.virtualGroupManager.QuerySPByAddress(operatorAddress)
}
