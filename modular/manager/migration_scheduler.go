package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type GlobalVirtualGroupByBucketMigrateExecuteUnit struct {
	bucketID uint64
	GlobalVirtualGroupMigrateExecuteUnit
}

type GlobalVirtualGroupMigrateExecuteUnit struct {
	gvgID        uint32
	migrateIndex int32 // -1 represents migrate primary
	destSPID     uint32
}

type VirtualGroupFamilyMigrateExecuteUnit struct {
	familyID                uint32
	gvgList                 []*virtualgrouptypes.GlobalVirtualGroup
	destSPID                uint32
	conflictGVGMigrateUnits []*GlobalVirtualGroupMigrateExecuteUnit
	primaryGVGMigrateUnits  []*GlobalVirtualGroupMigrateExecuteUnit
}

// sp1,sp2,sp3,sp4,sp5,sp6,sp7,sp8
// family1 primary=sp1 (gvg1(sp8,sp2,sp3,sp4,sp5,sp6,sp7), gvg2 (sp8,sp2,sp3,sp4,sp5,sp6,sp7))

// family1 dest_primary=sp8

func (vgfUnit *VirtualGroupFamilyMigrateExecuteUnit) checkConflict() {

}

type MigrateExecutePlan struct {
	virtualGroupManager     vgmgr.VirtualGroupManager
	VGFMigrateUnits         []*VirtualGroupFamilyMigrateExecuteUnit         // sp exit
	GVGMigrateUnits         []*GlobalVirtualGroupMigrateExecuteUnit         // sp exit
	GVGByBucketMigrateUnits []*GlobalVirtualGroupByBucketMigrateExecuteUnit // bucket migrate
}

// TODO: load from db.
func (plan *MigrateExecutePlan) load() {
	// subscribe progress
	// plan progress
	// task progress
}

func (plan *MigrateExecutePlan) run() {

}

// MigrationScheduler subscribes sp exit events and produces a gvg migrate plan.
type MigrationScheduler struct {
	manager                     *ManageModular
	spID                        uint32
	currentSubscribeBlockHeight int  // load from db
	isExiting                   bool // load from db
	// spStatus                    string  // active/exiting/exited
}

// Init function is used to load db subscribe block progress and migrate gvg progress.
func (s *MigrationScheduler) Init() error {
	if s.manager == nil {
		return fmt.Errorf("manger is nil")
	}
	spInfo, err := s.manager.baseApp.Consensus().QuerySP(context.Background(), s.manager.baseApp.OperatorAddress())
	if err != nil {
		return err
	}
	s.spID = spInfo.GetId()
	return nil
}

// Start function is used to subscribe sp exit event from metadata and produces a gvg migrate plan.
func (s *MigrationScheduler) Start() error {
	go s.subscribeEvents()
	return nil
}

func (s *MigrationScheduler) subscribeEvents() {
	subscribeSPExitEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSPExitEventInterval) * time.Second)
	subscribeBucketMigrateEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeBucketMigrateEventInterval) * time.Second)
	subscribeSwapOutEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSwapOutEventInterval) * time.Second)
	for {
		select {
		case <-subscribeSPExitEventsTicker.C:
			// TODO: subscribe sp exit events from metadata service.
			// spExitEvent, err = s.manager.baseApp.GfSpClient().ListSPExitEvents(s.currentSubscribeBlockHeight, s.manager.baseApp.OperatorAddress())

			if s.isExiting {
				return
			}
			// TODO exit
			plan, _ := s.produceSPExitExecutePlan()
			plan.run()
			s.isExiting = true
		case <-subscribeBucketMigrateEventsTicker.C:
			// TODO:
			s.produceBucketMigrateExecutePlan()
			// TODO:
			s.updateBucketMigrateExecutePlan()
		case <-subscribeSwapOutEventsTicker.C:
			// TODO:
			s.updateSPExitExecutePlan()

		}
	}
}

func (s *MigrationScheduler) produceSPExitExecutePlan() (*MigrateExecutePlan, error) {
	var (
		err     error
		vgfList []*virtualgrouptypes.GlobalVirtualGroupFamily
		plan    *MigrateExecutePlan
	)

	if vgfList, err = s.manager.baseApp.Consensus().ListVirtualGroupFamilies(context.Background(), s.spID); err != nil {
		log.Errorw("failed to list virtual group family", "error", err)
		return plan, err
	}
	// TODO:
	_ = vgfList
	return plan, err
}

func (s *MigrationScheduler) updateSPExitExecutePlan() {
}

func (s *MigrationScheduler) produceBucketMigrateExecutePlan() (*MigrateExecutePlan, error) {
	var (
		err  error
		plan *MigrateExecutePlan
	)

	return plan, err
}

func (s *MigrationScheduler) updateBucketMigrateExecutePlan() {

}
