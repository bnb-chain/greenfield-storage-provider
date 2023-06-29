package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type GlobalVirtualGroupByBucketMigrateExecuteUnit struct {
	bucketID uint64
	GlobalVirtualGroupMigrateExecuteUnit
}

type GlobalVirtualGroupMigrateExecuteUnit struct {
	gvg          *virtualgrouptypes.GlobalVirtualGroup
	migrateIndex int32 // -1 represents migrate primary
	destSPID     uint32
}

func (gvgUnit *GlobalVirtualGroupMigrateExecuteUnit) pickMigrateDestSP() error {
	return nil
}

type VirtualGroupFamilyMigrateExecuteUnit struct {
	vgf                            *virtualgrouptypes.GlobalVirtualGroupFamily
	gvgList                        []*virtualgrouptypes.GlobalVirtualGroup
	resolveConflictGVGMigrateUnits []*GlobalVirtualGroupMigrateExecuteUnit
	destSPID                       uint32
	primaryGVGMigrateUnits         []*GlobalVirtualGroupMigrateExecuteUnit
}

type PickDestSPFilter struct {
	currentSecondarySPMap map[uint32]bool
}

func (f *PickDestSPFilter) Check(spID uint32) bool {
	if _, found := f.currentSecondarySPMap[spID]; found {
		return false
	}
	return true
}

/*
Conflict Description
1.Current virtual group and sp status

	sp_list=[sp1,sp2,sp3,sp4,sp5,sp6,sp7,sp8]
	family1 = {primary=sp1, gvg1=(sp1,sp2,sp3,sp4,sp5,sp6,sp7), gvg2=(sp1,sp2,sp3,sp4,sp5,sp6,sp8))}

2.Resolve conflict

	sp1 exit, cannot pick a sp to replace sp1, there is a conflict.
	resolveConflictGVGMigrateUnits = gvg1=(sp1,sp2,sp3,sp4,sp5,sp6,sp7)->gvg1=(sp1,sp2,sp3,sp4,sp5,sp6,sp8)

3.After resolve conflict

	family1 = {primary=sp1, gvg1=(sp1,sp2,sp3,sp4,sp5,sp6,sp8), gvg2=(sp1,sp2,sp3,sp4,sp5,sp6,sp8))}

4.Primary migrate

	pick dst_primary_sp=sp7, and migrate family:
	family1 = {primary=sp7, gvg1=(sp1,sp2,sp3,sp4,sp5,sp6,sp8), gvg2=(sp7,sp2,sp3,sp4,sp5,sp6,sp8))}

5.Complete sp exit

	sp1 complete sp exit.
*/
func (vgfUnit *VirtualGroupFamilyMigrateExecuteUnit) checkConflict(vgm vgmgr.VirtualGroupManager) error {
	var (
		err              error
		secondarySPIDMap = make(map[uint32]bool)
		destSP           *sptypes.StorageProvider
	)
	for _, gvg := range vgfUnit.gvgList {
		for _, secondarySPID := range gvg.GetSecondarySpIds() {
			secondarySPIDMap[secondarySPID] = true
		}
	}
	filter := &PickDestSPFilter{
		secondarySPIDMap,
	}
	if destSP, err = vgm.PickSPByFilter(filter); err != nil {
		// TODO: has conflict
	}
	vgfUnit.destSPID = destSP.GetId()
	// TODO: impl
	return err
}

type MigrateExecutePlan struct {
	virtualGroupManager     vgmgr.VirtualGroupManager
	VGFMigrateUnits         []*VirtualGroupFamilyMigrateExecuteUnit         // sp exit
	GVGMigrateUnits         []*GlobalVirtualGroupMigrateExecuteUnit         // sp exit
	GVGByBucketMigrateUnits []*GlobalVirtualGroupByBucketMigrateExecuteUnit // bucket migrate
}

func (plan *MigrateExecutePlan) loadFromDB() error {
	// subscribe progress
	// plan progress
	// task progress
	return nil
}
func (plan *MigrateExecutePlan) storeToDB() error {
	// TODO:
	return nil
}

func (plan *MigrateExecutePlan) updateProgress() error {
	// TODO: update memory and db.
	return nil
}

// Init load from db.
func (plan *MigrateExecutePlan) Init() error {
	return plan.loadFromDB()
}

// Start persist plan and task to db and task dispatcher
func (plan *MigrateExecutePlan) Start() error {
	var err error
	for _, fUnit := range plan.VGFMigrateUnits {
		if err = fUnit.checkConflict(plan.virtualGroupManager); err != nil {
			log.Errorw("failed to start migrate execute plan due to check conflict", "family_unit", fUnit, "error", err)
			return err
		}
	}
	for _, gUnit := range plan.GVGMigrateUnits {
		if err = gUnit.pickMigrateDestSP(); err != nil {
			log.Errorw("failed to start migrate execute plan due to pick migrate dest sp", "gvg_unit", gUnit, "error", err)
			return err
		}
	}

	return nil
}

// MigrationScheduler subscribes sp exit events and produces a gvg migrate plan.
type MigrationScheduler struct {
	manager                     *ManageModular
	spID                        uint32
	currentSubscribeBlockHeight int  // load from db
	isExiting                   bool // load from db
	executePlan                 *MigrateExecutePlan
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
			plan, _ := s.produceSPExitExecutePlan()
			if err := plan.Start(); err != nil {
				log.Errorw("failed to start sp exit execute plan", "error", err)
				return
			}
			s.isExiting = true
			// TODO: update subscribe progress to db
		case <-subscribeBucketMigrateEventsTicker.C:
			// TODO: subscribe sp exit events from metadata service.
			// spExitEvent, err = s.manager.baseApp.GfSpClient().ListBucketMigrateEvents(s.currentSubscribeBlockHeight, s.manager.baseApp.OperatorAddress())
			// TODO: start migrate, produce plan and start
			plan, _ := s.produceBucketMigrateExecutePlan()
			plan.Start()
			// TODO: end migrate, update bucket migrate status
			s.updateBucketMigrateExecutePlan()
		case <-subscribeSwapOutEventsTicker.C:
			// TODO:
			s.updateSPExitExecutePlan()
		}
	}
}

func (s *MigrationScheduler) produceSPExitExecutePlan() (*MigrateExecutePlan, error) {
	var (
		err              error
		vgfList          []*virtualgrouptypes.GlobalVirtualGroupFamily
		secondaryGVGList []*virtualgrouptypes.GlobalVirtualGroup
		plan             *MigrateExecutePlan
	)

	if vgfList, err = s.manager.baseApp.Consensus().ListVirtualGroupFamilies(context.Background(), s.spID); err != nil {
		log.Errorw("failed to list virtual group family", "error", err)
		return plan, err
	}
	plan = &MigrateExecutePlan{
		virtualGroupManager: s.manager.virtualGroupManager,
		VGFMigrateUnits:     make([]*VirtualGroupFamilyMigrateExecuteUnit, 0),
		GVGMigrateUnits:     make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
	for _, f := range vgfList {
		plan.VGFMigrateUnits = append(plan.VGFMigrateUnits, &VirtualGroupFamilyMigrateExecuteUnit{vgf: f})
	}
	// TODO: query metadata service to get secondary sp's gvg list.
	for _, g := range secondaryGVGList {
		plan.GVGMigrateUnits = append(plan.GVGMigrateUnits, &GlobalVirtualGroupMigrateExecuteUnit{gvg: g})
	}
	return plan, err
}

func (s *MigrationScheduler) updateSPExitExecutePlan() {
	// TODO: check
}

func (s *MigrationScheduler) produceBucketMigrateExecutePlan() (*MigrateExecutePlan, error) {
	var (
		err  error
		plan *MigrateExecutePlan
	)
	// TODO:
	return plan, err
}

func (s *MigrationScheduler) updateBucketMigrateExecutePlan() {
	// TODO:
}
