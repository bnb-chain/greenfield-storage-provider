package manager

import (
	"context"

	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// NotifyMigrateSwapOut is used to receive migrate gvg task from src sp.
func (m *ManageModular) NotifyMigrateSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) error {
	if m.spExitScheduler == nil {
		return ErrNotifyMigrateSwapOut
	}

	// TODO: check SuccessorSpId
	return m.spExitScheduler.AddSwapOutToTaskRunner(swapOut)
}
