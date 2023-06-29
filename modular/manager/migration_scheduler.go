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

// if spID not in gvg secondary sps. return -1.
func getSecondarySPIndex(gvg *virtualgrouptypes.GlobalVirtualGroup, spID uint32) int32 {
	for index, secondarySPID := range gvg.GetSecondarySpIds() {
		if secondarySPID == spID {
			return int32(index)
		}
	}
	return -1
}

type GlobalVirtualGroupByBucketMigrateExecuteUnit struct {
	bucketID uint64
	GlobalVirtualGroupMigrateExecuteUnit
}

type GlobalVirtualGroupMigrateExecuteUnit struct {
	gvg            *virtualgrouptypes.GlobalVirtualGroup
	redundantIndex int32 // if < 0, represents migrate primary
	srcSPID        uint32
	destSPID       uint32
}

type VirtualGroupFamilyMigrateExecuteUnit struct {
	vgf                                *virtualgrouptypes.GlobalVirtualGroupFamily
	srcSPID                            uint32 // self sp id.
	gvgList                            []*virtualgrouptypes.GlobalVirtualGroup
	conflictedSecondaryGVGMigrateUnits []*GlobalVirtualGroupMigrateExecuteUnit // need be resolved firstly
	destSPID                           uint32
	primaryGVGMigrateUnits             []*GlobalVirtualGroupMigrateExecuteUnit
}

func NewVirtualGroupFamilyMigrateExecuteUnit(vgf *virtualgrouptypes.GlobalVirtualGroupFamily, selfSPID uint32) *VirtualGroupFamilyMigrateExecuteUnit {
	return &VirtualGroupFamilyMigrateExecuteUnit{
		vgf:                                vgf,
		srcSPID:                            selfSPID,
		gvgList:                            make([]*virtualgrouptypes.GlobalVirtualGroup, 0),
		conflictedSecondaryGVGMigrateUnits: make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		destSPID:                           0,
		primaryGVGMigrateUnits:             make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
}

// PickDestSPFilter is used to pick sp id which is not in excluded sp ids.
type PickDestSPFilter struct {
	excludedSPIDs []uint32
}

func NewPickDestSPFilterWithMap(m map[uint32]int) *PickDestSPFilter {
	spIDs := make([]uint32, 0)
	for spID, _ := range m {
		spIDs = append(spIDs, spID)
	}
	return &PickDestSPFilter{excludedSPIDs: spIDs}
}

func NewPickDestSPFilterWithSlice(s []uint32) *PickDestSPFilter {
	return &PickDestSPFilter{excludedSPIDs: s}
}

func (f *PickDestSPFilter) Check(spID uint32) bool {
	for _, v := range f.excludedSPIDs {
		if v == spID {
			return false
		}
	}
	return true
}

/*
Check Conflict Description
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
func (vgfUnit *VirtualGroupFamilyMigrateExecuteUnit) expandExecuteSubUnits(vgm vgmgr.VirtualGroupManager) error {
	var (
		err                    error
		hasPrimaryGVG          bool
		familySecondarySPIDMap = make(map[uint32]int)
		destFamilySP           *sptypes.StorageProvider
	)
	for _, gvg := range vgfUnit.gvgList {
		for _, secondarySPID := range gvg.GetSecondarySpIds() {
			familySecondarySPIDMap[secondarySPID] = familySecondarySPIDMap[secondarySPID] + 1
		}
		hasPrimaryGVG = true
	}
	if hasPrimaryGVG {
		// check conflicts.
		if destFamilySP, err = vgm.PickSPByFilter(NewPickDestSPFilterWithMap(familySecondarySPIDMap)); err != nil {
			// primary family migrate has conflicts, choose a sp with fewer conflicts.
			secondarySPIDBindingLeastGVGs := vgfUnit.gvgList[0].GetSecondarySpIds()[0]
			for spID, count := range familySecondarySPIDMap {
				if count < familySecondarySPIDMap[secondarySPIDBindingLeastGVGs] {
					secondarySPIDBindingLeastGVGs = spID
				}
			}
			// resolve conflict, swap out secondarySPIDBindingLeastGVGs.
			for _, gvg := range vgfUnit.gvgList {
				if secondaryIndex := getSecondarySPIndex(gvg, secondarySPIDBindingLeastGVGs); secondaryIndex > 0 {
					// gvg has conflicts.
					var (
						srcSecondarySPID uint32
						destSecondarySP  *sptypes.StorageProvider
					)
					srcSecondarySPID = gvg.GetSecondarySpIds()[secondaryIndex]
					if destSecondarySP, err = vgm.PickSPByFilter(NewPickDestSPFilterWithSlice(gvg.GetSecondarySpIds())); err != nil {
						log.Errorw("failed to check conflict due to pick secondary sp", "error", err)
						return err
					}
					vgfUnit.conflictedSecondaryGVGMigrateUnits = append(vgfUnit.conflictedSecondaryGVGMigrateUnits,
						&GlobalVirtualGroupMigrateExecuteUnit{
							gvg, int32(secondaryIndex),
							srcSecondarySPID, destSecondarySP.GetId()})
				}
			}
		} else {
			vgfUnit.destSPID = destFamilySP.GetId()
			for _, gvg := range vgfUnit.gvgList {
				vgfUnit.primaryGVGMigrateUnits = append(vgfUnit.primaryGVGMigrateUnits,
					&GlobalVirtualGroupMigrateExecuteUnit{
						gvg, -1,
						vgfUnit.srcSPID, destFamilySP.GetId()})
			}
		}
	}
	return nil
}

type MigrateExecutePlan struct {
	virtualGroupManager            vgmgr.VirtualGroupManager
	PrimaryVGFMigrateUnits         []*VirtualGroupFamilyMigrateExecuteUnit         // sp exit, primary family, include gvg list
	SecondaryGVGMigrateUnits       []*GlobalVirtualGroupMigrateExecuteUnit         // sp exit, secondary gvg
	PrimaryGVGByBucketMigrateUnits []*GlobalVirtualGroupByBucketMigrateExecuteUnit // bucket migrate，primary gvg
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

func (plan *MigrateExecutePlan) startSchedule() {
	// TODO:
	// send control msg to dest sp endpoint and trigger migrate.
}

// Init load from db.
func (plan *MigrateExecutePlan) Init() error {
	return plan.loadFromDB()
}

// Start persist plan and task to db and task dispatcher
func (plan *MigrateExecutePlan) Start() error {
	var err error
	for _, fUnit := range plan.PrimaryVGFMigrateUnits {
		if err = fUnit.expandExecuteSubUnits(plan.virtualGroupManager); err != nil {
			log.Errorw("failed to start migrate execute plan due to expand execute sub units", "family_unit", fUnit, "error", err)
			return err
		}
	}
	for _, gUnit := range plan.SecondaryGVGMigrateUnits {
		// expand execute unit
		var (
			redundantIndex  int32
			destSecondarySP *sptypes.StorageProvider
		)
		if redundantIndex = getSecondarySPIndex(gUnit.gvg, gUnit.srcSPID); redundantIndex < 0 {
			log.Errorw("failed to start migrate execute plan due to get secondary sp index", "gvg_unit", gUnit, "error", err)
			return err
		}
		if destSecondarySP, err = plan.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(gUnit.gvg.GetSecondarySpIds())); err != nil {
			log.Errorw("failed to start migrate execute plan due to get secondary dest sp", "gvg_unit", gUnit, "error", err)
			return err
		}
		gUnit.redundantIndex = redundantIndex
		gUnit.destSPID = destSecondarySP.GetId()
	}
	if err = plan.storeToDB(); err != nil {
		log.Errorw("failed to start migrate execute plan due to store db", "error", err)
		return err
	}
	go plan.startSchedule()
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
		virtualGroupManager:      s.manager.virtualGroupManager,
		PrimaryVGFMigrateUnits:   make([]*VirtualGroupFamilyMigrateExecuteUnit, 0),
		SecondaryGVGMigrateUnits: make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
	for _, f := range vgfList {
		plan.PrimaryVGFMigrateUnits = append(plan.PrimaryVGFMigrateUnits, NewVirtualGroupFamilyMigrateExecuteUnit(f, s.spID))
	}
	// TODO: query metadata service to get secondary sp's gvg list.
	for _, g := range secondaryGVGList {
		plan.SecondaryGVGMigrateUnits = append(plan.SecondaryGVGMigrateUnits, &GlobalVirtualGroupMigrateExecuteUnit{gvg: g, srcSPID: s.spID})
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
