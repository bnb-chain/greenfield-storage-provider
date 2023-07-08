package manager

import (
	"context"

	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// NotifyMigrateSwapOut is used to receive migrate gvg task from src sp.
func (m *ManageModular) NotifyMigrateSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) error {
	if m.spExitScheduler == nil {
		return ErrNotifyMigrateGVG
	}

	//return m.spExitScheduler.AddNewMigrateGVGUnit(&GlobalVirtualGroupMigrateExecuteUnit{
	//	gvg:             task.GetGvg(),
	//	redundancyIndex: task.GetRedundancyIdx(),
	//	isRemoted:       true, // from src sp
	//	isConflicted:    false,
	//	isSecondary:     false,
	//	srcSP:           task.GetSrcSp(),
	//	destSP:          task.GetDestSp(),
	//	migrateStatus:   WaitForMigrate,
	//})
	return nil
}
