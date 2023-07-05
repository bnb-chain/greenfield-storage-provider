package manager

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/core/task"
)

// NotifyMigrateGVG is used to receive migrate gvg task from src sp.
func (m *ManageModular) NotifyMigrateGVG(ctx context.Context, task task.MigrateGVGTask) error {
	if m.spExitScheduler == nil {
		return ErrNotifyMigrateGVG
	}

	return m.spExitScheduler.AddNewMigrateGVGUnit(&GlobalVirtualGroupMigrateExecuteUnit{
		gvg:             task.GetGvg(),
		redundancyIndex: task.GetRedundancyIdx(),
		isRemoted:       true, // from src sp
		srcSP:           task.GetSrcSp(),
		destSP:          task.GetDestSp(),
		migrateStatus:   WaitForMigrate,
	})
}
