package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var _ vgmgr.PickFilter = &PickDestSPFilter{}

const (
	// MaxSrcRunningMigrateGVG defines src sp max running migrate gvg units, and avoid src sp overload.
	MaxSrcRunningMigrateGVG = 100
	// MaxDestRunningMigrateGVG defines dest sp max running migrate gvg units, and avoid dest sp overload.
	MaxDestRunningMigrateGVG = 10
)

type MigrateStatus int32

var (
	WaitForMigrate MigrateStatus = 0
	Migrating      MigrateStatus = 1
	Migrated       MigrateStatus = 2
)

type GlobalVirtualGroupMigrateExecuteUnit struct {
	gvg            *virtualgrouptypes.GlobalVirtualGroup
	redundantIndex int32 // if < 0, represents migrate primary
	srcSP          *sptypes.StorageProvider
	destSP         *sptypes.StorageProvider
	migrateStatus  MigrateStatus
	checkTimestamp uint64
	checkStatus    string // update to proto enum
}

type VirtualGroupFamilyMigrateExecuteUnit struct {
	vgf                                *virtualgrouptypes.GlobalVirtualGroupFamily
	srcSP                              *sptypes.StorageProvider
	gvgList                            []*virtualgrouptypes.GlobalVirtualGroup
	conflictedSecondaryGVGMigrateUnits []*GlobalVirtualGroupMigrateExecuteUnit // need be resolved firstly
	destSP                             *sptypes.StorageProvider
	primaryGVGMigrateUnits             []*GlobalVirtualGroupMigrateExecuteUnit
}

func NewVirtualGroupFamilyMigrateExecuteUnit(vgf *virtualgrouptypes.GlobalVirtualGroupFamily, selfSP *sptypes.StorageProvider) *VirtualGroupFamilyMigrateExecuteUnit {
	return &VirtualGroupFamilyMigrateExecuteUnit{
		vgf:                                vgf,
		srcSP:                              selfSP,
		gvgList:                            make([]*virtualgrouptypes.GlobalVirtualGroup, 0),
		conflictedSecondaryGVGMigrateUnits: make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
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
func (vgfUnit *VirtualGroupFamilyMigrateExecuteUnit) expandExecuteSubUnits(vgm vgmgr.VirtualGroupManager, plan *SPExitExecutePlan) error {
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
			srcSP, queryErr := vgm.QuerySPByID(secondarySPIDBindingLeastGVGs)
			if queryErr != nil {
				log.Errorw("failed to query sp", "error", queryErr)
				return queryErr
			}
			// resolve conflict, swap out secondarySPIDBindingLeastGVGs.
			for _, gvg := range vgfUnit.gvgList {
				if secondaryIndex, _ := util.GetSecondarySPIndexFromGVG(gvg, secondarySPIDBindingLeastGVGs); secondaryIndex > 0 {
					// gvg has conflicts.
					destSecondarySP, pickErr := vgm.PickSPByFilter(NewPickDestSPFilterWithSlice(gvg.GetSecondarySpIds()))
					if pickErr != nil {
						log.Errorw("failed to check conflict due to pick secondary sp", "error", pickErr)
						return pickErr
					}
					gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
						gvg:            gvg,
						redundantIndex: secondaryIndex,
						srcSP:          srcSP,
						destSP:         destSecondarySP,
						migrateStatus:  WaitForMigrate}
					vgfUnit.conflictedSecondaryGVGMigrateUnits = append(vgfUnit.conflictedSecondaryGVGMigrateUnits, gUnit)
					plan.ToMigrateGVG = append(plan.ToMigrateGVG, gUnit)
				}
			}
		} else { // has no conflicts
			vgfUnit.destSP = destFamilySP
			for _, gvg := range vgfUnit.gvgList {
				gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
					gvg:            gvg,
					redundantIndex: -1,
					srcSP:          vgfUnit.srcSP,
					destSP:         destFamilySP,
					migrateStatus:  WaitForMigrate}
				vgfUnit.primaryGVGMigrateUnits = append(vgfUnit.primaryGVGMigrateUnits, gUnit)
				plan.ToMigrateGVG = append(plan.ToMigrateGVG, gUnit)
			}
		}
	}
	return nil
}

type SPExitExecutePlan struct {
	manager                  *ManageModular
	virtualGroupManager      vgmgr.VirtualGroupManager
	runningMigrateGVG        int                                     // load from db
	PrimaryVGFMigrateUnits   []*VirtualGroupFamilyMigrateExecuteUnit // sp exit, primary family, include gvg list
	SecondaryGVGMigrateUnits []*GlobalVirtualGroupMigrateExecuteUnit // sp exit, secondary gvg

	// for scheduling, the slice only can append.
	ToMigrateMutex sync.RWMutex
	ToMigrateGVG   []*GlobalVirtualGroupMigrateExecuteUnit
	MigratingMutex sync.RWMutex
	MigratingGVG   []*GlobalVirtualGroupMigrateExecuteUnit
}

func (plan *SPExitExecutePlan) loadFromDB() error {
	// subscribe progress
	// plan progress
	// task progress
	return nil
}
func (plan *SPExitExecutePlan) storeToDB() error {
	// TODO:
	return nil
}

func (plan *SPExitExecutePlan) updateProgress() error {
	// TODO: update memory and db.
	return nil
}

// ToMigrateExecuteUnitIterator is used to dispatch migrate units.
type ToMigrateExecuteUnitIterator struct {
	plan           *SPExitExecutePlan
	toMigrateIndex int
	isValid        bool
}

func NewToMigrateExecuteUnitIterator(plan *SPExitExecutePlan) *ToMigrateExecuteUnitIterator {
	return &ToMigrateExecuteUnitIterator{
		plan:           plan,
		toMigrateIndex: 0,
		isValid:        true,
	}
}

func (ti *ToMigrateExecuteUnitIterator) SeekToFirst() {
	ti.plan.ToMigrateMutex.RLock()
	defer ti.plan.ToMigrateMutex.RUnlock()
	for index, gvg := range ti.plan.ToMigrateGVG {
		if gvg.migrateStatus == WaitForMigrate {
			ti.toMigrateIndex = index
			return
		}
	}
	ti.isValid = false
}

func (ti *ToMigrateExecuteUnitIterator) Valid() bool {
	if ti.plan.runningMigrateGVG >= MaxSrcRunningMigrateGVG {
		return false
	}
	if !ti.isValid {
		return false
	}
	ti.plan.ToMigrateMutex.RLock()
	isValid := ti.toMigrateIndex < len(ti.plan.ToMigrateGVG)
	ti.plan.ToMigrateMutex.RUnlock()
	if !isValid {
		return false
	}
	return true
}

func (ti *ToMigrateExecuteUnitIterator) Next() {
	ti.toMigrateIndex++
}

func (ti *ToMigrateExecuteUnitIterator) Value() *GlobalVirtualGroupMigrateExecuteUnit {
	ti.plan.ToMigrateMutex.RLock()
	defer ti.plan.ToMigrateMutex.RUnlock()
	return ti.plan.ToMigrateGVG[ti.toMigrateIndex]
}

// MigratingExecuteUnitIterator is used to check migrating units.
type MigratingExecuteUnitIterator struct {
	plan           *SPExitExecutePlan
	MigratingIndex int
	isValid        bool
}

func NewMigratingExecuteUnitIterator(plan *SPExitExecutePlan) *MigratingExecuteUnitIterator {
	return &MigratingExecuteUnitIterator{
		plan:           plan,
		MigratingIndex: 0,
		isValid:        true,
	}
}

func (mi *MigratingExecuteUnitIterator) SeekToFirst() {
	mi.plan.MigratingMutex.RLock()
	defer mi.plan.MigratingMutex.RUnlock()
	for index, gvg := range mi.plan.MigratingGVG {
		if gvg.migrateStatus == Migrating {
			mi.MigratingIndex = index
			return
		}
	}
	mi.isValid = false
}

func (mi *MigratingExecuteUnitIterator) Valid() bool {
	if !mi.isValid {
		return false
	}
	mi.plan.MigratingMutex.RLock()
	isValid := mi.MigratingIndex < len(mi.plan.MigratingGVG)
	mi.plan.MigratingMutex.RUnlock()
	if !isValid {
		return false
	}
	return true
}

func (mi *MigratingExecuteUnitIterator) Next() {
	mi.MigratingIndex++
}

func (mi *MigratingExecuteUnitIterator) Value() *GlobalVirtualGroupMigrateExecuteUnit {
	mi.plan.ToMigrateMutex.RLock()
	defer mi.plan.ToMigrateMutex.RUnlock()
	return mi.plan.MigratingGVG[mi.MigratingIndex]
}

func (plan *SPExitExecutePlan) startSrcSPSchedule() {
	// notify dest sp start migrate gvg and check them migrate status.
	go plan.notifyDestSPMigrateExecuteUnits()
	go plan.checkDestSPMigrateExecuteUnitsStatus()
}

func (plan *SPExitExecutePlan) startDestSPSchedule() {
	// TODO:
	// MaxDestRunningMigrateGVG

}

func (plan *SPExitExecutePlan) notifyDestSPMigrateExecuteUnits() {
	// dispatch migrate unit to corresponding dest sp.
	// maybe need get dest sp migrate approval.

	var (
		dispatchLoopNumber uint64
		dispatchUnitNumber uint64
	)
	for {
		time.Sleep(10 * time.Second)
		dispatchLoopNumber++
		dispatchUnitNumber = 0
		iter := NewToMigrateExecuteUnitIterator(plan)
		for iter.SeekToFirst(); iter.Valid(); iter.Next() {
			dispatchUnitNumber++
			toMigrateGVG := iter.Value()
			_ = toMigrateGVG
			// TODO:
			// send migrate gvg to dest sp, trigger dest sp pull the gvg object from src sp.
			// init migrate task
			// plan.manager.baseApp.GfSpClient().NotifyDestSPMigrateGVG(context.Background(), toMigrateGVG.destSP.GetEndpoint(), toMigrateGVG)

		}
		log.Infow("dispatch migrate unit to dest sp", "loop_number", dispatchLoopNumber, "dispatch_number", dispatchUnitNumber)
	}
}

func (plan *SPExitExecutePlan) checkDestSPMigrateExecuteUnitsStatus() {
	// Periodically check whether the execution unit is executing normally in dest sp.
	// maybe need retry the failed unit.
	var (
		checkLoopNumber uint64
		checkUnitNumber uint64
	)
	for {
		time.Sleep(10 * time.Second)
		checkLoopNumber++
		checkUnitNumber = 0
		iter := NewMigratingExecuteUnitIterator(plan)
		for iter.SeekToFirst(); iter.Valid(); iter.Next() {
			checkUnitNumber++
			MigratingGVG := iter.Value()
			_ = MigratingGVG
			// TODO:
			// send check unit status request, and record status.
		}
		log.Infow("check migrating unit status", "loop_number", checkLoopNumber, "check_number", checkUnitNumber)
	}
}

// Init load from db.
func (plan *SPExitExecutePlan) Init() error {
	return plan.loadFromDB()
}

// Start persist plan and task to db and task dispatcher
func (plan *SPExitExecutePlan) Start() error {
	// TODO: request other sp migrate approval, get dest sp set.
	// expand migrate units.
	var err error
	for _, fUnit := range plan.PrimaryVGFMigrateUnits {
		if err = fUnit.expandExecuteSubUnits(plan.virtualGroupManager, plan); err != nil {
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
		if redundantIndex, err = util.GetSecondarySPIndexFromGVG(gUnit.gvg, gUnit.srcSP.GetId()); err != nil {
			log.Errorw("failed to start migrate execute plan due to get secondary sp index", "gvg_unit", gUnit, "error", err)
			return err
		}
		if destSecondarySP, err = plan.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(gUnit.gvg.GetSecondarySpIds())); err != nil {
			log.Errorw("failed to start migrate execute plan due to get secondary dest sp", "gvg_unit", gUnit, "error", err)
			return err
		}
		gUnit.redundantIndex = redundantIndex
		gUnit.destSP = destSecondarySP
		gUnit.migrateStatus = WaitForMigrate
		plan.ToMigrateGVG = append(plan.ToMigrateGVG, gUnit)
	}

	if err = plan.storeToDB(); err != nil {
		log.Errorw("failed to start migrate execute plan due to store db", "error", err)
		return err
	}
	go plan.startSrcSPSchedule()
	return nil
}

// SPExitScheduler subscribes sp exit events and produces a gvg migrate plan.
type SPExitScheduler struct {
	// sp exit workflow src sp.
	manager                     *ManageModular
	selfSP                      *sptypes.StorageProvider
	currentSubscribeBlockHeight uint64 // load from db
	isExiting                   bool   // load from db
	executePlan                 *SPExitExecutePlan
	// sp exit workflow dest sp.
	// migrate task runner
}

// Init function is used to load db subscribe block progress and migrate gvg progress.
func (s *SPExitScheduler) Init() error {
	if s.manager == nil {
		return fmt.Errorf("manger is nil")
	}
	sp, err := s.manager.baseApp.Consensus().QuerySP(context.Background(), s.manager.baseApp.OperatorAddress())
	if err != nil {
		return err
	}
	s.selfSP = sp
	return nil
}

// Start function is used to subscribe sp exit event from metadata and produces a gvg migrate plan.
func (s *SPExitScheduler) Start() error {
	go s.subscribeEvents()
	return nil
}

func (s *SPExitScheduler) subscribeEvents() {
	subscribeSPExitEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSPExitEventInterval) * time.Second)
	subscribeSwapOutEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSwapOutEventInterval) * time.Second)
	for {
		select {
		case <-subscribeSPExitEventsTicker.C:
			// TODO: subscribe sp exit events from metadata service.
			// spExitEvent, err = s.manager.baseApp.GfSpClient().GfSpListSpExitEvents(context.Background(), s.currentSubscribeBlockHeight, s.manager.baseApp.OperatorAddress())
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
		case <-subscribeSwapOutEventsTicker.C:
			// TODO:
			s.updateSPExitExecutePlan()
		}
	}
}

func (s *SPExitScheduler) produceSPExitExecutePlan() (*SPExitExecutePlan, error) {
	var (
		err              error
		vgfList          []*virtualgrouptypes.GlobalVirtualGroupFamily
		secondaryGVGList []*virtualgrouptypes.GlobalVirtualGroup
		plan             *SPExitExecutePlan
	)

	if vgfList, err = s.manager.baseApp.Consensus().ListVirtualGroupFamilies(context.Background(), s.selfSP.GetId()); err != nil {
		log.Errorw("failed to list virtual group family", "error", err)
		return plan, err
	}
	plan = &SPExitExecutePlan{
		manager:                  s.manager,
		virtualGroupManager:      s.manager.virtualGroupManager,
		PrimaryVGFMigrateUnits:   make([]*VirtualGroupFamilyMigrateExecuteUnit, 0),
		SecondaryGVGMigrateUnits: make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		ToMigrateGVG:             make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		MigratingGVG:             make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
	for _, f := range vgfList {
		plan.PrimaryVGFMigrateUnits = append(plan.PrimaryVGFMigrateUnits, NewVirtualGroupFamilyMigrateExecuteUnit(f, s.selfSP))
	}
	// TODO: query metadata service to get secondary sp's gvg list.
	for _, g := range secondaryGVGList {
		plan.SecondaryGVGMigrateUnits = append(plan.SecondaryGVGMigrateUnits, &GlobalVirtualGroupMigrateExecuteUnit{gvg: g, srcSP: s.selfSP})
	}
	return plan, err
}

func (s *SPExitScheduler) updateSPExitExecutePlan() {
	// TODO: check
}
