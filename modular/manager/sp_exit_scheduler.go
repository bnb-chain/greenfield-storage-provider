package manager

import (
	"context"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
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
	redundantIndex int32 // if < 0, represents migrate primary.
	isConflict     bool  // only be used in sp exit.
	isSecondary    bool  // only be used in sp exit.
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
						isConflict:     true,
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
					isConflict:     false,
					isSecondary:    false,
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
	scheduler                *SPExitScheduler
	virtualGroupManager      vgmgr.VirtualGroupManager
	runningMigrateGVG        int                                     // load from db
	PrimaryVGFMigrateUnits   []*VirtualGroupFamilyMigrateExecuteUnit // sp exit, primary family, include gvg list
	SecondaryGVGMigrateUnits []*GlobalVirtualGroupMigrateExecuteUnit // sp exit, secondary gvg

	// for scheduling, the slice only can append to ensure iterator work fine.
	ToMigrateMutex sync.RWMutex
	ToMigrateGVG   []*GlobalVirtualGroupMigrateExecuteUnit
	MigratingMutex sync.RWMutex
	MigratingGVG   []*GlobalVirtualGroupMigrateExecuteUnit
}

func (plan *SPExitExecutePlan) makeGVGUnit(gvgMeta *spdb.MigrateGVGUnitMeta, isConflict bool) (*GlobalVirtualGroupMigrateExecuteUnit, error) {
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
		gvg:            gvg,
		redundantIndex: gvgMeta.MigrateRedundancyIndex,
		isConflict:     isConflict,
		srcSP:          srcSP,
		destSP:         destSP,
		migrateStatus:  MigrateStatus(gvgMeta.MigrateStatus),
	}
	if gUnit.migrateStatus == WaitForMigrate {
		plan.ToMigrateGVG = append(plan.ToMigrateGVG, gUnit)
	}
	if gUnit.migrateStatus == Migrating {
		plan.MigratingGVG = append(plan.MigratingGVG, gUnit)
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
		conflictGVGList, listConflictErr := plan.manager.baseApp.GfSpDB().ListConflictsMigrateGVGUnitsByFamilyID(f.GetId())
		if listConflictErr != nil {
			log.Errorw("failed to load from db due to list conflict gvg", "error", err)
			return listConflictErr
		}
		for _, conflictGVG := range conflictGVGList {
			gUnit, makeErr := plan.makeGVGUnit(conflictGVG, true)
			if makeErr != nil {
				log.Errorw("failed to load from db due to make gvg unit")
				return makeErr
			}
			vgfUnit.conflictedSecondaryGVGMigrateUnits = append(vgfUnit.conflictedSecondaryGVGMigrateUnits, gUnit)
		}
		familyGVGList, listFamilyGVGErr := plan.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByFamilyID(f.GetId(), plan.scheduler.selfSP.GetId())
		if listFamilyGVGErr != nil {
			log.Errorw("failed to load from db due to list family gvg", "error", err)
			return listFamilyGVGErr
		}
		for _, familyGVG := range familyGVGList {
			gUnit, makeErr := plan.makeGVGUnit(familyGVG, false)
			if makeErr != nil {
				log.Errorw("failed to load from db due to make gvg unit")
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
		redundantIndex, getSecondaryIndexErr := util.GetSecondarySPIndexFromGVG(gvg, plan.scheduler.selfSP.GetId())
		if getSecondaryIndexErr != nil {
			log.Errorw("failed to get secondary sp index from gvg", "error", err)
			return getSecondaryIndexErr
		}
		gvgMeta, queryError := plan.manager.baseApp.GfSpDB().QueryMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
			GlobalVirtualGroupID:   gvg.GetId(),
			VirtualGroupFamilyID:   0,
			MigrateRedundancyIndex: redundantIndex,
			BucketID:               0,
			IsSecondary:            true,
			IsConflict:             false,
		})
		if queryError != nil {
			log.Errorw("failed to load from db due to query gvg", "error", queryError)
			return queryError
		}
		gUnit, makeErr := plan.makeGVGUnit(gvgMeta, false)
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
		for _, conflictGVG := range vgfUnit.conflictedSecondaryGVGMigrateUnits {
			if err = plan.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
				GlobalVirtualGroupID:   conflictGVG.gvg.GetId(),
				VirtualGroupFamilyID:   conflictGVG.gvg.GetFamilyId(),
				MigrateRedundancyIndex: conflictGVG.redundantIndex,
				BucketID:               0,
				IsSecondary:            true,
				IsConflict:             true,
				SrcSPID:                conflictGVG.srcSP.GetId(),
				DestSPID:               conflictGVG.destSP.GetId(),
				LastMigrateObjectID:    0,
				MigrateStatus:          int(conflictGVG.migrateStatus),
			}); err != nil {
				log.Errorw("failed to store to db", "error", err)
				return err
			}
		}
		for _, familyGVG := range vgfUnit.primaryGVGMigrateUnits {
			if err = plan.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
				GlobalVirtualGroupID:   familyGVG.gvg.GetId(),
				VirtualGroupFamilyID:   familyGVG.gvg.GetFamilyId(),
				MigrateRedundancyIndex: -1,
				BucketID:               0,
				IsSecondary:            false,
				IsConflict:             false,
				SrcSPID:                familyGVG.srcSP.GetId(),
				DestSPID:               familyGVG.destSP.GetId(),
				LastMigrateObjectID:    0,
				MigrateStatus:          int(familyGVG.migrateStatus),
			}); err != nil {
				log.Errorw("failed to store to db", "error", err)
				return err
			}
		}
	}
	for _, secondaryGVG := range plan.SecondaryGVGMigrateUnits {
		if err = plan.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
			GlobalVirtualGroupID:   secondaryGVG.gvg.GetId(),
			VirtualGroupFamilyID:   0,
			MigrateRedundancyIndex: secondaryGVG.redundantIndex,
			BucketID:               0,
			IsSecondary:            true,
			IsConflict:             false,
			SrcSPID:                secondaryGVG.srcSP.GetId(),
			DestSPID:               secondaryGVG.destSP.GetId(),
			LastMigrateObjectID:    0,
			MigrateStatus:          int(secondaryGVG.migrateStatus),
		}); err != nil {
			log.Errorw("failed to store to db", "error", err)
			return err
		}
	}
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
		err              error
		notifyLoopNumber uint64
		notifyUnitNumber uint64
	)
	for {
		time.Sleep(10 * time.Second)
		notifyLoopNumber++
		notifyUnitNumber = 0
		iter := NewToMigrateExecuteUnitIterator(plan)
		for iter.SeekToFirst(); iter.Valid(); iter.Next() {
			notifyUnitNumber++
			toMigrateGVG := iter.Value()
			migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{}
			migrateGVGTask.InitMigrateGVGTask(plan.manager.baseApp.TaskPriority(migrateGVGTask),
				0, toMigrateGVG.gvg, toMigrateGVG.redundantIndex,
				toMigrateGVG.srcSP, toMigrateGVG.destSP)
			err = plan.manager.baseApp.GfSpClient().NotifyDestSPMigrateGVG(context.Background(), toMigrateGVG.destSP.GetEndpoint(), migrateGVGTask)
			log.Infow("notify dest sp migrate gvg", "migrate_gvg_task", migrateGVGTask, "error", err)
			// ignore this error , fail over is handled in check phase.
			toMigrateGVG.migrateStatus = Migrating

			plan.MigratingMutex.Lock()
			plan.MigratingGVG = append(plan.MigratingGVG, toMigrateGVG)
			plan.MigratingMutex.Unlock()

			err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(&spdb.MigrateGVGUnitMeta{
				GlobalVirtualGroupID:   toMigrateGVG.gvg.GetId(),
				VirtualGroupFamilyID:   toMigrateGVG.gvg.GetFamilyId(),
				MigrateRedundancyIndex: toMigrateGVG.redundantIndex,
				BucketID:               0,
				IsSecondary:            toMigrateGVG.isSecondary,
				IsConflict:             toMigrateGVG.isConflict,
			}, int(Migrating))
			log.Errorw("notify migrate gvg", "gvg_meta", toMigrateGVG, "error", err)
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
		gUnit.isSecondary = true
		gUnit.isConflict = false
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
	manager                         *ManageModular
	selfSP                          *sptypes.StorageProvider
	lastSubscribedSPExitBlockHeight uint64 // load from db
	isExiting                       bool
	isExited                        bool
	executePlan                     *SPExitExecutePlan
	// sp exit workflow dest sp.
	// migrate task runner
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
		ToMigrateGVG:             make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		MigratingGVG:             make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
	}
	if err = s.executePlan.Init(); err != nil {
		log.Errorw("failed to init sp exit scheduler due to plan init", "error", err)
		return err
	}
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
			spStartExitEvents, spEndExitExitEvents, subscribeError := s.manager.baseApp.GfSpClient().ListSpExitEvents(context.Background(), s.lastSubscribedSPExitBlockHeight+1, s.manager.baseApp.OperatorAddress())
			if subscribeError != nil {
				log.Errorw("failed to subscribe sp exit event", "error", subscribeError)
				return
			}
			if len(spStartExitEvents) > 0 {
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
				updateErr := s.manager.baseApp.GfSpDB().UpdateSPExitSubscribeProgress(s.lastSubscribedSPExitBlockHeight + 1)
				if updateErr != nil {
					log.Errorw("failed to update sp exit progress", "error", updateErr)
					return
				}
				s.executePlan = plan
				s.isExiting = true
				s.lastSubscribedSPExitBlockHeight++
			}
			if len(spEndExitExitEvents) > 0 {
				s.isExited = true
			}

		case <-subscribeSwapOutEventsTicker.C:
			if s.isExited {
				return
			}
			// TODO:
			// TODO: is changing.
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
		ToMigrateGVG:             make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		MigratingGVG:             make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
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
