package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const (
	// MaxSrcRunningMigrateGVG defines src sp max running migrate gvg units, and avoid src sp overload.
	MaxSrcRunningMigrateGVG = 100
)

var _ vgmgr.PickFilter = &PickDestSPFilter{}

// PickDestSPFilter is used to pick sp id which is not in excluded sp ids.
type PickDestSPFilter struct {
	excludedSPIDs []uint32
}

func NewPickDestSPFilterWithMap(m map[uint32]int) *PickDestSPFilter {
	spIDs := make([]uint32, 0)
	for spID := range m {
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

type VirtualGroupFamilyMigrateExecuteUnit struct {
	vgf                       *virtualgrouptypes.GlobalVirtualGroupFamily
	srcSP                     *sptypes.StorageProvider
	conflictedGVGMigrateUnits []*GlobalVirtualGroupMigrateExecuteUnit // need be resolved firstly
	destSP                    *sptypes.StorageProvider
	primaryGVGMigrateUnits    []*GlobalVirtualGroupMigrateExecuteUnit
}

func NewVirtualGroupFamilyMigrateExecuteUnit(vgf *virtualgrouptypes.GlobalVirtualGroupFamily, selfSP *sptypes.StorageProvider) *VirtualGroupFamilyMigrateExecuteUnit {
	return &VirtualGroupFamilyMigrateExecuteUnit{
		vgf:                       vgf,
		srcSP:                     selfSP,
		conflictedGVGMigrateUnits: make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		primaryGVGMigrateUnits:    make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
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
		familyGVGs             []*virtualgrouptypes.GlobalVirtualGroup
	)
	if familyGVGs, err = plan.manager.baseApp.Consensus().ListGlobalVirtualGroupsByFamilyID(context.Background(), plan.scheduler.selfSP.GetId(), vgfUnit.vgf.GetId()); err != nil {
		log.Errorw("failed to load from db due to list virtual groups by family id", "error", err)
		return err
	}

	for _, gvg := range familyGVGs {
		for _, secondarySPID := range gvg.GetSecondarySpIds() {
			familySecondarySPIDMap[secondarySPID] = familySecondarySPIDMap[secondarySPID] + 1
		}
		hasPrimaryGVG = true
	}
	if hasPrimaryGVG {
		// check conflicts.
		if destFamilySP, err = vgm.PickSPByFilter(NewPickDestSPFilterWithMap(familySecondarySPIDMap)); err != nil {
			// primary family migrate has conflicts, choose a sp with fewer conflicts.
			secondarySPIDBindingLeastGVGs := familyGVGs[0].GetSecondarySpIds()[0]
			for spID, count := range familySecondarySPIDMap {
				if count < familySecondarySPIDMap[secondarySPIDBindingLeastGVGs] {
					secondarySPIDBindingLeastGVGs = spID
				}
			}
			srcSP, queryErr := vgm.QuerySPByID(secondarySPIDBindingLeastGVGs)
			if queryErr != nil {
				log.Errorw("failed to query sp", "sp_id", secondarySPIDBindingLeastGVGs, "error", queryErr)
				return queryErr
			}
			// resolve conflict, swap out secondarySPIDBindingLeastGVGs.
			for _, gvg := range familyGVGs {
				if redundancyIndex, _ := util.GetSecondarySPIndexFromGVG(gvg, secondarySPIDBindingLeastGVGs); redundancyIndex > 0 {
					// gvg has conflicts.
					destSecondarySP, pickErr := vgm.PickSPByFilter(NewPickDestSPFilterWithSlice(gvg.GetSecondarySpIds()))
					if pickErr != nil {
						log.Errorw("failed to check conflict due to pick secondary sp", "gvg", gvg, "error", pickErr)
						return pickErr
					}
					gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
						gvg:             gvg,
						redundancyIndex: redundancyIndex,
						isSecondary:     true,
						isRemoted:       false,
						isConflicted:    true,
						srcSP:           srcSP,
						destSP:          destSecondarySP,
						migrateStatus:   WaitForNotifyDestSP}
					vgfUnit.conflictedGVGMigrateUnits = append(vgfUnit.conflictedGVGMigrateUnits, gUnit)
					plan.WaitForNotifyDestSPGVGs = append(plan.WaitForNotifyDestSPGVGs, gUnit)
				}
			}
		} else { // has no conflicts
			vgfUnit.destSP = destFamilySP
			for _, gvg := range familyGVGs {
				gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
					gvg:             gvg,
					redundancyIndex: -1,
					srcSP:           vgfUnit.srcSP,
					destSP:          destFamilySP,
					isConflicted:    false,
					isRemoted:       false,
					isSecondary:     false,
					migrateStatus:   WaitForNotifyDestSP}
				vgfUnit.primaryGVGMigrateUnits = append(vgfUnit.primaryGVGMigrateUnits, gUnit)
				plan.WaitForNotifyDestSPGVGs = append(plan.WaitForNotifyDestSPGVGs, gUnit)
			}
		}
	}
	return nil
}

// SPExitExecutePlan is used to record the execution of subtasks in src sp.
type SPExitExecutePlan struct {
	manager             *ManageModular
	scheduler           *SPExitScheduler
	virtualGroupManager vgmgr.VirtualGroupManager
	// runningMigrateGVG        int                                     // TODO: refine it.
	PrimaryVGFMigrateUnits   []*VirtualGroupFamilyMigrateExecuteUnit // sp exit, primary family, include gvg list, maybe has conflicted.
	SecondaryGVGMigrateUnits []*GlobalVirtualGroupMigrateExecuteUnit // sp exit, secondary gvg

	// for scheduling, the slice only can append to ensure iterator work fine.
	WaitForNotifyDestSPMutex sync.RWMutex
	WaitForNotifyDestSPGVGs  []*GlobalVirtualGroupMigrateExecuteUnit
	NotifiedDestSPMutex      sync.RWMutex
	NotifiedDestSPGVGs       []*GlobalVirtualGroupMigrateExecuteUnit
}

func (plan *SPExitExecutePlan) makeGVGUnit(gvgMeta *spdb.MigrateGVGUnitMeta, isConflicted bool, isSecondary bool) (*GlobalVirtualGroupMigrateExecuteUnit, error) {
	gvg, queryGVGErr := plan.manager.baseApp.Consensus().QueryGlobalVirtualGroup(context.Background(), gvgMeta.GlobalVirtualGroupID)
	if queryGVGErr != nil {
		log.Errorw("failed to make gvg unit due to query gvg", "error", queryGVGErr)
		return nil, queryGVGErr
	}
	srcSP, querySPErr := plan.virtualGroupManager.QuerySPByID(gvgMeta.SrcSPID)
	if querySPErr != nil {
		log.Errorw("failed to make gvg unit due to query sp", "error", querySPErr)
		return nil, querySPErr
	}
	destSP, querySPErr := plan.virtualGroupManager.QuerySPByID(gvgMeta.DestSPID)
	if querySPErr != nil {
		log.Errorw("failed to make gvg unit due to query sp", "error", querySPErr)
		return nil, querySPErr
	}
	gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
		gvg:             gvg,
		redundancyIndex: gvgMeta.RedundancyIndex,
		isConflicted:    isConflicted,
		isRemoted:       false,
		isSecondary:     isSecondary,
		srcSP:           srcSP,
		destSP:          destSP,
		migrateStatus:   MigrateStatus(gvgMeta.MigrateStatus),
	}
	if gUnit.migrateStatus == WaitForNotifyDestSP {
		plan.WaitForNotifyDestSPGVGs = append(plan.WaitForNotifyDestSPGVGs, gUnit)
	}
	if gUnit.migrateStatus == NotifiedDestSP {
		plan.NotifiedDestSPGVGs = append(plan.NotifiedDestSPGVGs, gUnit)
	}
	return gUnit, nil
}

// loadFromDB is used to rebuild the memory plan topology.
func (plan *SPExitExecutePlan) loadFromDB() error {
	var (
		err              error
		vgfList          []*virtualgrouptypes.GlobalVirtualGroupFamily
		secondaryGVGList []*virtualgrouptypes.GlobalVirtualGroup
	)

	if vgfList, err = plan.manager.baseApp.Consensus().ListVirtualGroupFamilies(context.Background(), plan.scheduler.selfSP.GetId()); err != nil {
		log.Errorw("failed to load from db due to list virtual group family", "error", err)
		return err
	}
	for _, f := range vgfList {
		vgfUnit := NewVirtualGroupFamilyMigrateExecuteUnit(f, plan.scheduler.selfSP)
		conflictedGVGList, listConflictedErr := plan.manager.baseApp.GfSpDB().ListConflictedMigrateGVGUnitsByFamilyID(f.GetId())
		if listConflictedErr != nil {
			log.Errorw("failed to load from db due to list conflict gvg", "error", listConflictedErr)
			return listConflictedErr
		}
		for _, conflictedGVG := range conflictedGVGList {
			gUnit, makeErr := plan.makeGVGUnit(conflictedGVG, true, true)
			if makeErr != nil {
				log.Errorw("failed to load from db due to make gvg unit", "error", makeErr)
				return makeErr
			}
			vgfUnit.conflictedGVGMigrateUnits = append(vgfUnit.conflictedGVGMigrateUnits, gUnit)
		}
		// TODO: refine it, need check whether complete conflicted gvg migrate.
		// gvg which has no conflict.
		familyGVGList, listFamilyGVGErr := plan.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByFamilyID(f.GetId(), plan.scheduler.selfSP.GetId())
		if listFamilyGVGErr != nil {
			log.Errorw("failed to load from db due to list family gvg", "error", listFamilyGVGErr)
			return listFamilyGVGErr
		}
		for _, familyGVG := range familyGVGList {
			gUnit, makeErr := plan.makeGVGUnit(familyGVG, false, false)
			if makeErr != nil {
				log.Errorw("failed to load from db due to make gvg unit", "error", makeErr)
				return makeErr
			}
			vgfUnit.primaryGVGMigrateUnits = append(vgfUnit.primaryGVGMigrateUnits, gUnit)
		}
		plan.PrimaryVGFMigrateUnits = append(plan.PrimaryVGFMigrateUnits, vgfUnit)
	}

	if secondaryGVGList, err = plan.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsBySecondarySP(context.Background(), plan.scheduler.selfSP.GetId()); err != nil {
		log.Errorw("failed to list secondary virtual group", "error", err)
		return err
	}
	for _, gvg := range secondaryGVGList {
		redundancyIndex, getSecondaryIndexErr := util.GetSecondarySPIndexFromGVG(gvg, plan.scheduler.selfSP.GetId())
		if getSecondaryIndexErr != nil {
			log.Errorw("failed to get secondary sp index from gvg", "error", getSecondaryIndexErr)
			return getSecondaryIndexErr
		}
		migrateKey := MakeSecondaryGVGMigrateKey(gvg.GetId(), gvg.GetFamilyId(), redundancyIndex)
		gvgMeta, queryError := plan.manager.baseApp.GfSpDB().QueryMigrateGVGUnit(migrateKey)
		if queryError != nil {
			log.Errorw("failed to load from db due to query gvg", "migrate_key", migrateKey, "error", queryError)
			return queryError
		}
		gUnit, makeErr := plan.makeGVGUnit(gvgMeta, false, true)
		if makeErr != nil {
			log.Errorw("failed to load from db due to make gvg unit")
			return makeErr
		}
		plan.SecondaryGVGMigrateUnits = append(plan.SecondaryGVGMigrateUnits, gUnit)
	}

	return nil
}

// it is called at start of the execute plan.
func (plan *SPExitExecutePlan) storeToDB() error {
	var err error
	for _, vgfUnit := range plan.PrimaryVGFMigrateUnits {
		for _, conflictGVG := range vgfUnit.conflictedGVGMigrateUnits {
			if err = plan.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
				MigrateGVGKey:        conflictGVG.Key(),
				GlobalVirtualGroupID: conflictGVG.gvg.GetId(),
				VirtualGroupFamilyID: conflictGVG.gvg.GetFamilyId(),
				RedundancyIndex:      conflictGVG.redundancyIndex,
				BucketID:             0,
				IsSecondary:          true,
				IsConflicted:         true,
				IsRemoted:            false,
				SrcSPID:              conflictGVG.srcSP.GetId(),
				DestSPID:             conflictGVG.destSP.GetId(),
				LastMigratedObjectID: 0,
				MigrateStatus:        int(conflictGVG.migrateStatus),
			}); err != nil {
				log.Errorw("failed to store to db", "error", err)
				return err
			}
		}
		for _, familyGVG := range vgfUnit.primaryGVGMigrateUnits {
			if err = plan.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
				MigrateGVGKey:        familyGVG.Key(),
				GlobalVirtualGroupID: familyGVG.gvg.GetId(),
				VirtualGroupFamilyID: familyGVG.gvg.GetFamilyId(),
				RedundancyIndex:      -1,
				BucketID:             0,
				IsSecondary:          false,
				IsConflicted:         false,
				IsRemoted:            false,
				SrcSPID:              familyGVG.srcSP.GetId(),
				DestSPID:             familyGVG.destSP.GetId(),
				LastMigratedObjectID: 0,
				MigrateStatus:        int(familyGVG.migrateStatus),
			}); err != nil {
				log.Errorw("failed to store to db", "error", err)
				return err
			}
		}
	}
	for _, secondaryGVG := range plan.SecondaryGVGMigrateUnits {
		if err = plan.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
			MigrateGVGKey:        secondaryGVG.Key(),
			GlobalVirtualGroupID: secondaryGVG.gvg.GetId(),
			VirtualGroupFamilyID: 0,
			RedundancyIndex:      secondaryGVG.redundancyIndex,
			BucketID:             0,
			IsSecondary:          true,
			IsConflicted:         false,
			IsRemoted:            false,
			SrcSPID:              secondaryGVG.srcSP.GetId(),
			DestSPID:             secondaryGVG.destSP.GetId(),
			LastMigratedObjectID: 0,
			MigrateStatus:        int(secondaryGVG.migrateStatus),
		}); err != nil {
			log.Errorw("failed to store to db", "error", err)
			return err
		}
	}
	return nil
}

//func (plan *SPExitExecutePlan) updateProgress() error {
//	// TODO: update memory and db.
//	return nil
//}

// WaitForNotifyDestSPIterator is used to notify migrate units to dest sp.
type WaitForNotifyDestSPIterator struct {
	plan               *SPExitExecutePlan
	waitForNotifyIndex int
	isValid            bool
}

func NewWaitForNotifyDestSPIterator(plan *SPExitExecutePlan) *WaitForNotifyDestSPIterator {
	return &WaitForNotifyDestSPIterator{
		plan:               plan,
		waitForNotifyIndex: 0,
		isValid:            true,
	}
}

func (ti *WaitForNotifyDestSPIterator) SeekToFirst() {
	ti.plan.WaitForNotifyDestSPMutex.RLock()
	defer ti.plan.WaitForNotifyDestSPMutex.RUnlock()
	for index, gvg := range ti.plan.WaitForNotifyDestSPGVGs {
		if gvg.migrateStatus == WaitForMigrate {
			ti.waitForNotifyIndex = index
			return
		}
	}
	ti.isValid = false
}

func (ti *WaitForNotifyDestSPIterator) Valid() bool {
	// TODO: refine it.
	// if ti.plan.runningMigrateGVG >= MaxSrcRunningMigrateGVG {
	// 	return false
	// }
	if !ti.isValid {
		return false
	}
	ti.plan.WaitForNotifyDestSPMutex.RLock()
	isValid := ti.waitForNotifyIndex < len(ti.plan.WaitForNotifyDestSPGVGs)
	ti.plan.WaitForNotifyDestSPMutex.RUnlock()
	return isValid
}

func (ti *WaitForNotifyDestSPIterator) Next() {
	ti.waitForNotifyIndex++
}

func (ti *WaitForNotifyDestSPIterator) Value() *GlobalVirtualGroupMigrateExecuteUnit {
	ti.plan.WaitForNotifyDestSPMutex.RLock()
	defer ti.plan.WaitForNotifyDestSPMutex.RUnlock()
	return ti.plan.WaitForNotifyDestSPGVGs[ti.waitForNotifyIndex]
}

// NotifiedDestSPIterator is used to check dest migrating units.
type NotifiedDestSPIterator struct {
	plan          *SPExitExecutePlan
	notifiedIndex int
	isValid       bool
}

func NewNotifiedDestSPIterator(plan *SPExitExecutePlan) *NotifiedDestSPIterator {
	return &NotifiedDestSPIterator{
		plan:          plan,
		notifiedIndex: 0,
		isValid:       true,
	}
}

func (mi *NotifiedDestSPIterator) SeekToFirst() {
	mi.plan.NotifiedDestSPMutex.RLock()
	defer mi.plan.NotifiedDestSPMutex.RUnlock()
	for index, gvg := range mi.plan.NotifiedDestSPGVGs {
		if gvg.migrateStatus == NotifiedDestSP {
			mi.notifiedIndex = index
			return
		}
	}
	mi.isValid = false
}

func (mi *NotifiedDestSPIterator) Valid() bool {
	if !mi.isValid {
		return false
	}
	mi.plan.NotifiedDestSPMutex.RLock()
	isValid := mi.notifiedIndex < len(mi.plan.NotifiedDestSPGVGs)
	mi.plan.NotifiedDestSPMutex.RUnlock()
	return isValid
}

func (mi *NotifiedDestSPIterator) Next() {
	mi.notifiedIndex++
}

func (mi *NotifiedDestSPIterator) Value() *GlobalVirtualGroupMigrateExecuteUnit {
	mi.plan.NotifiedDestSPMutex.RLock()
	defer mi.plan.NotifiedDestSPMutex.RUnlock()
	return mi.plan.NotifiedDestSPGVGs[mi.notifiedIndex]
}

func (plan *SPExitExecutePlan) startSrcSPSchedule() {
	// notify dest sp start migrate gvg and check them migrate status.
	go plan.notifyDestSPMigrateExecuteUnits()
	go plan.checkDestSPMigrateExecuteUnitsStatus()
}

func (plan *SPExitExecutePlan) notifyDestSPMigrateExecuteUnits() {
	// dispatch migrate unit to corresponding dest sp.
	// maybe need get dest sp migrate approval.

	var (
		err              error
		notifyLoopNumber uint64
		notifyUnitNumber uint64
	)
	for {
		time.Sleep(10 * time.Second)
		notifyLoopNumber++
		notifyUnitNumber = 0
		iter := NewWaitForNotifyDestSPIterator(plan)
		for iter.SeekToFirst(); iter.Valid(); iter.Next() {
			notifyUnitNumber++
			waitForMigrateGVG := iter.Value()
			migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{}
			migrateGVGTask.InitMigrateGVGTask(plan.manager.baseApp.TaskPriority(migrateGVGTask),
				0, waitForMigrateGVG.gvg, waitForMigrateGVG.redundancyIndex,
				waitForMigrateGVG.srcSP, waitForMigrateGVG.destSP)
			err = plan.manager.baseApp.GfSpClient().NotifyDestSPMigrateGVG(context.Background(), waitForMigrateGVG.destSP.GetEndpoint(), migrateGVGTask)
			log.Infow("notify dest sp migrate gvg", "migrate_gvg_task", migrateGVGTask, "error", err)
			// ignore this error , fail over is handled in check phase.
			waitForMigrateGVG.migrateStatus = NotifiedDestSP

			plan.NotifiedDestSPMutex.Lock()
			plan.NotifiedDestSPGVGs = append(plan.NotifiedDestSPGVGs, waitForMigrateGVG)
			plan.NotifiedDestSPMutex.Unlock()

			err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(waitForMigrateGVG.Key(), int(NotifiedDestSP))
			log.Errorw("notify migrate gvg", "gvg_meta", waitForMigrateGVG, "error", err)
		}
		log.Infow("notify migrate unit to dest sp", "loop_number", notifyLoopNumber, "notify_number", notifyUnitNumber)
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
		iter := NewNotifiedDestSPIterator(plan)
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
		var (
			redundancyIndex int32
			destSecondarySP *sptypes.StorageProvider
		)
		if redundancyIndex, err = util.GetSecondarySPIndexFromGVG(gUnit.gvg, gUnit.srcSP.GetId()); err != nil {
			log.Errorw("failed to start migrate execute plan due to get secondary sp index", "gvg_unit", gUnit, "error", err)
			return err
		}
		if destSecondarySP, err = plan.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(gUnit.gvg.GetSecondarySpIds())); err != nil {
			log.Errorw("failed to start migrate execute plan due to get secondary dest sp", "gvg_unit", gUnit, "error", err)
			return err
		}
		gUnit.redundancyIndex = redundancyIndex
		gUnit.destSP = destSecondarySP
		gUnit.migrateStatus = WaitForNotifyDestSP
		gUnit.isSecondary = true
		gUnit.isConflicted = false
		plan.WaitForNotifyDestSPGVGs = append(plan.WaitForNotifyDestSPGVGs, gUnit)
	}

	if err = plan.storeToDB(); err != nil {
		log.Errorw("failed to start migrate execute plan due to store db", "error", err)
		return err
	}
	go plan.startSrcSPSchedule()
	return nil
}

// MigrateTaskRunner is used to manage task migrate progress/status in dest sp.
type MigrateTaskRunner struct {
	manager             *ManageModular
	virtualGroupManager vgmgr.VirtualGroupManager
	mutex               sync.RWMutex
	keyIndexMap         map[string]int
	gvgUnits            []*GlobalVirtualGroupMigrateExecuteUnit
}

func (runner *MigrateTaskRunner) addGVGUnit(gvgMeta *spdb.MigrateGVGUnitMeta) error {
	gvg, queryGVGErr := runner.manager.baseApp.Consensus().QueryGlobalVirtualGroup(context.Background(), gvgMeta.GlobalVirtualGroupID)
	if queryGVGErr != nil {
		log.Errorw("failed to make gvg unit due to query gvg", "error", queryGVGErr)
		return queryGVGErr
	}
	srcSP, querySPErr := runner.virtualGroupManager.QuerySPByID(gvgMeta.SrcSPID)
	if querySPErr != nil {
		log.Errorw("failed to make gvg unit due to query sp", "error", querySPErr)
		return querySPErr
	}
	destSP, querySPErr := runner.virtualGroupManager.QuerySPByID(gvgMeta.DestSPID)
	if querySPErr != nil {
		log.Errorw("failed to make gvg unit due to query sp", "error", querySPErr)
		return querySPErr
	}
	gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
		gvg:             gvg,
		redundancyIndex: gvgMeta.RedundancyIndex,
		srcSP:           srcSP,
		destSP:          destSP,
		isRemoted:       true,
		migrateStatus:   MigrateStatus(gvgMeta.MigrateStatus),
	}

	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	runner.gvgUnits = append(runner.gvgUnits, gUnit)
	runner.keyIndexMap[gUnit.Key()] = len(runner.gvgUnits) - 1
	return nil
}

func (runner *MigrateTaskRunner) loadFromDB() error {
	remotedGVGList, listRemotedErr := runner.manager.baseApp.GfSpDB().ListRemotedMigrateGVGUnits()
	if listRemotedErr != nil {
		return listRemotedErr
	}
	for _, gvgMeta := range remotedGVGList {
		if addErr := runner.addGVGUnit(gvgMeta); addErr != nil {
			return addErr
		}
	}
	return nil
}

func (runner *MigrateTaskRunner) Init() error {
	return runner.loadFromDB()
}

func (runner *MigrateTaskRunner) Start() error {
	go runner.startDestSPSchedule()
	return nil
}

func (runner *MigrateTaskRunner) UpdateMigrateGVGLastMigratedObjectID(migrateKey string, lastMigratedObjectID uint64) error {
	runner.mutex.Lock()

	if _, found := runner.keyIndexMap[migrateKey]; !found {
		runner.mutex.Unlock()
		return fmt.Errorf("gvg unit is not found")
	}
	index := runner.keyIndexMap[migrateKey]
	if index >= len(runner.gvgUnits) {
		runner.mutex.Unlock()
		return fmt.Errorf("gvg unit index is invalid")
	}
	unit := runner.gvgUnits[index]
	unit.lastMigratedObjectID = lastMigratedObjectID
	runner.mutex.Unlock()

	return runner.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitLastMigrateObjectID(migrateKey, lastMigratedObjectID)
}

func (runner *MigrateTaskRunner) UpdateMigrateGVGStatus(migrateKey string, st MigrateStatus) error {
	runner.mutex.Lock()

	if _, found := runner.keyIndexMap[migrateKey]; !found {
		runner.mutex.Unlock()
		return fmt.Errorf("gvg unit is not found")
	}
	index := runner.keyIndexMap[migrateKey]
	if index >= len(runner.gvgUnits) {
		runner.mutex.Unlock()
		return fmt.Errorf("gvg unit index is invalid")
	}
	unit := runner.gvgUnits[index]
	unit.migrateStatus = st
	runner.mutex.Unlock()

	return runner.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateKey, int(st))
}

func (runner *MigrateTaskRunner) AddNewMigrateGVGUnit(remotedGVGUnit *GlobalVirtualGroupMigrateExecuteUnit) error {
	runner.mutex.Lock()
	if _, found := runner.keyIndexMap[remotedGVGUnit.Key()]; found {
		runner.mutex.Unlock()
		return nil
	}
	runner.gvgUnits = append(runner.gvgUnits, remotedGVGUnit)
	runner.keyIndexMap[remotedGVGUnit.Key()] = len(runner.gvgUnits) - 1
	runner.mutex.Unlock()

	// add to db
	if err := runner.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
		MigrateGVGKey:        remotedGVGUnit.Key(),
		GlobalVirtualGroupID: remotedGVGUnit.gvg.GetId(),
		VirtualGroupFamilyID: remotedGVGUnit.gvg.GetFamilyId(),
		RedundancyIndex:      remotedGVGUnit.redundancyIndex,
		BucketID:             0,
		IsSecondary:          remotedGVGUnit.isSecondary,
		IsConflicted:         remotedGVGUnit.isConflicted,
		IsRemoted:            remotedGVGUnit.isRemoted,
		SrcSPID:              remotedGVGUnit.srcSP.GetId(),
		DestSPID:             remotedGVGUnit.destSP.GetId(),
		LastMigratedObjectID: 0,
		MigrateStatus:        int(remotedGVGUnit.migrateStatus),
	}); err != nil {
		log.Errorw("failed to store to db", "error", err)
		return err
	}

	return nil
}

func (runner *MigrateTaskRunner) startDestSPSchedule() {
	// loop try push
	for {
		time.Sleep(1 * time.Second)
		runner.mutex.RLock()
		for _, unit := range runner.gvgUnits {
			if unit.migrateStatus == WaitForMigrate {
				var err error
				migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{}
				migrateGVGTask.InitMigrateGVGTask(runner.manager.baseApp.TaskPriority(migrateGVGTask),
					0, unit.gvg, unit.redundancyIndex,
					unit.srcSP, unit.destSP)
				if err = runner.manager.migrateGVGQueue.Push(migrateGVGTask); err != nil {
					log.Errorw("failed to push migrate gvg task to queue", "error", err)
					time.Sleep(5 * time.Second) // Sleep for 5 seconds before retrying
				}
				if err = runner.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(unit.Key(), int(Migrating)); err != nil {
					log.Errorw("failed to update task status", "error", err)
					time.Sleep(5 * time.Second) // Sleep for 5 seconds before retrying
				}
				unit.migrateStatus = Migrating
				break
			}
		}
		runner.mutex.RUnlock()
	}
}

// SPExitScheduler is used to manage and schedule sp exit process.
type SPExitScheduler struct {
	manager *ManageModular
	selfSP  *sptypes.StorageProvider

	// sp exit workflow src sp.
	// manage subscribe progress and execute plan.
	lastSubscribedSPExitBlockHeight uint64
	isExiting                       bool
	isExited                        bool
	executePlan                     *SPExitExecutePlan

	// sp exit workflow dest sp.
	// manage specific gvg execution tasks.
	migrateTaskRunner *MigrateTaskRunner
}

// NewSPExitScheduler returns a sp exit scheduler instance.
func NewSPExitScheduler(manager *ManageModular) (*SPExitScheduler, error) {
	var err error
	spExitScheduler := &SPExitScheduler{}
	if err = spExitScheduler.Init(manager); err != nil {
		return nil, err
	}
	if err = spExitScheduler.Start(); err != nil {
		return nil, err
	}
	return spExitScheduler, nil
}

// Init function is used to load db subscribe block progress and migrate gvg progress.
func (s *SPExitScheduler) Init(m *ManageModular) error {
	var (
		err error
		sp  *sptypes.StorageProvider
	)
	s.manager = m
	if sp, err = s.manager.baseApp.Consensus().QuerySP(context.Background(), s.manager.baseApp.OperatorAddress()); err != nil {
		log.Errorw("failed to init sp exit scheduler due to query sp error", "error", err)
		return err
	}
	s.selfSP = sp
	s.isExiting = sp.GetStatus() == sptypes.STATUS_GRACEFUL_EXITING
	if s.lastSubscribedSPExitBlockHeight, err = s.manager.baseApp.GfSpDB().QuerySPExitSubscribeProgress(); err != nil {
		log.Errorw("failed to init sp exit scheduler due to init subscribe sp exit progress", "error", err)
		return err
	}
	s.executePlan = &SPExitExecutePlan{
		manager:                  s.manager,
		scheduler:                s,
		virtualGroupManager:      s.manager.virtualGroupManager,
		PrimaryVGFMigrateUnits:   make([]*VirtualGroupFamilyMigrateExecuteUnit, 0),
		SecondaryGVGMigrateUnits: make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		WaitForNotifyDestSPGVGs:  make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		NotifiedDestSPGVGs:       make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
	if err = s.executePlan.Init(); err != nil {
		log.Errorw("failed to init sp exit scheduler due to plan init", "error", err)
		return err
	}
	s.migrateTaskRunner = &MigrateTaskRunner{
		manager:     s.manager,
		keyIndexMap: make(map[string]int),
		gvgUnits:    make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
	return nil
}

// Start function is used to subscribe sp exit event from metadata and produces a gvg migrate plan.
func (s *SPExitScheduler) Start() error {
	go s.subscribeEvents()
	return nil
}

func (s *SPExitScheduler) AddNewMigrateGVGUnit(remotedUnit *GlobalVirtualGroupMigrateExecuteUnit) error {
	return s.migrateTaskRunner.AddNewMigrateGVGUnit(remotedUnit)
}

func (s *SPExitScheduler) subscribeEvents() {
	subscribeSPExitEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSPExitEventInterval) * time.Second)
	subscribeSwapOutEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSwapOutEventInterval) * time.Second)
	for {
		select {
		case <-subscribeSPExitEventsTicker.C:
			spExitEvents, subscribeError := s.manager.baseApp.GfSpClient().ListSpExitEvents(context.Background(), s.lastSubscribedSPExitBlockHeight+1, s.manager.baseApp.OperatorAddress())
			if subscribeError != nil {
				log.Errorw("failed to subscribe sp exit event", "error", subscribeError)
				return
			}
			if spExitEvents.Event != nil {
				if s.isExiting || s.isExited {
					return
				}
				plan, err := s.produceSPExitExecutePlan()
				if err != nil {
					log.Errorw("failed to produce sp exit execute plan", "error", err)
					return
				}
				if startErr := plan.Start(); startErr != nil {
					log.Errorw("failed to start sp exit execute plan", "error", startErr)
					return
				}

				s.executePlan = plan
				s.isExiting = true
			}
			if spExitEvents.CompleteEvent != nil {
				s.isExited = true
			}
			updateErr := s.manager.baseApp.GfSpDB().UpdateSPExitSubscribeProgress(s.lastSubscribedSPExitBlockHeight + 1)
			if updateErr != nil {
				log.Errorw("failed to update sp exit progress", "error", updateErr)
				return
			}
			s.lastSubscribedSPExitBlockHeight++
			log.Infow("sp exit subscribe progress", "last_subscribed_block_height", s.lastSubscribedSPExitBlockHeight)

		case <-subscribeSwapOutEventsTicker.C:
			if s.isExited {
				return
			}
			// TODO: refine it, proto is changing.
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
		scheduler:                s,
		virtualGroupManager:      s.manager.virtualGroupManager,
		PrimaryVGFMigrateUnits:   make([]*VirtualGroupFamilyMigrateExecuteUnit, 0),
		SecondaryGVGMigrateUnits: make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		WaitForNotifyDestSPGVGs:  make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		NotifiedDestSPGVGs:       make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
	for _, f := range vgfList {
		plan.PrimaryVGFMigrateUnits = append(plan.PrimaryVGFMigrateUnits, NewVirtualGroupFamilyMigrateExecuteUnit(f, s.selfSP))
	}
	if secondaryGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsBySecondarySP(context.Background(), s.selfSP.GetId()); err != nil {
		log.Errorw("failed to list secondary virtual group", "error", err)
		return plan, err
	}
	for _, g := range secondaryGVGList {
		plan.SecondaryGVGMigrateUnits = append(plan.SecondaryGVGMigrateUnits, &GlobalVirtualGroupMigrateExecuteUnit{gvg: g, srcSP: s.selfSP})
	}
	return plan, err
}

func (s *SPExitScheduler) updateSPExitExecutePlan() {
	// TODO: check
}

func (s *SPExitScheduler) UpdateMigrateProgress(task task.MigrateGVGTask) error {
	var (
		err        error
		migrateKey string
	)
	migrateKey = MakeRemotedGVGMigrateKey(task.GetGvg().GetId(), task.GetGvg().GetFamilyId(), task.GetRedundancyIdx())
	if task.GetFinished() {
		err = s.migrateTaskRunner.UpdateMigrateGVGStatus(migrateKey, Migrated)
	} else {
		err = s.migrateTaskRunner.UpdateMigrateGVGLastMigratedObjectID(migrateKey, task.GetLastMigratedObjectID())
	}
	return err
}
