package manager

import (
	"context"

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
