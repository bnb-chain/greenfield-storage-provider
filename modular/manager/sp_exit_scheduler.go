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

var _ vgmgr.SPPickFilter = &PickDestSPFilter{}

// SPExitScheduler is used to manage and schedule sp exit process.
type SPExitScheduler struct {
	manager *ManageModular
	selfSP  *sptypes.StorageProvider

	// sp exit workflow src sp.
	// manage subscribe progress and swap out plan.
	lastSubscribedSPExitBlockHeight  uint64
	lastSubscribedSwapOutBlockHeight uint64
	isExiting                        bool
	isExited                         bool
	swapOutPlan                      *SrcSPSwapOutPlan // swap out unit

	// sp exit workflow dest sp.
	// manage specific gvg execution tasks and swap out status.
	taskRunner *DestSPTaskRunner // gvg migrate unit
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

// Init is used to load db subscribe progress and migrate progress.
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
	if s.lastSubscribedSPExitBlockHeight, err = s.manager.baseApp.GfSpDB().QuerySPExitSubscribeProgress(); err != nil {
		log.Errorw("failed to init sp exit scheduler due to init subscribe sp exit progress", "error", err)
		return err
	}
	spExitEvents, subscribeError := s.manager.baseApp.GfSpClient().ListSpExitEvents(context.Background(), s.lastSubscribedSPExitBlockHeight, s.manager.baseApp.OperatorAddress())
	if subscribeError != nil {
		log.Errorw("failed to init due to subscribe sp exit", "error", subscribeError)
		return subscribeError
	}
	s.isExiting = spExitEvents.GetEvent() != nil
	s.isExited = spExitEvents.GetCompleteEvent() != nil

	if s.lastSubscribedSwapOutBlockHeight, err = s.manager.baseApp.GfSpDB().QuerySwapOutSubscribeProgress(); err != nil {
		log.Errorw("failed to init sp exit scheduler due to init subscribe swap out progress", "error", err)
		return err
	}
	if s.isExiting {
		s.swapOutPlan = NewSrcSPSwapOutPlan(s.manager, s, s.manager.virtualGroupManager)
		if err = s.swapOutPlan.LoadFromDB(); err != nil {
			log.Errorw("failed to init sp exit scheduler due to plan init", "error", err)
			return err
		}
	}

	s.taskRunner = NewDestSPTaskRunner(s.manager, s.manager.virtualGroupManager)
	return s.taskRunner.LoadFromDB()
}

// Start function is used to subscribe sp exit event from metadata and produces a gvg migrate plan.
func (s *SPExitScheduler) Start() error {
	var err error
	if s.swapOutPlan != nil {
		if err = s.swapOutPlan.Start(); err != nil {
			return err
		}
	}
	if err = s.taskRunner.Start(); err != nil {
		return err
	}
	go s.subscribeEvents()
	return nil
}

// UpdateMigrateProgress is used to update migrate status from task executor.
func (s *SPExitScheduler) UpdateMigrateProgress(task task.MigrateGVGTask) error {
	var (
		err        error
		migrateKey string
	)
	migrateKey = MakeRemotedGVGMigrateKey(task.GetGvg().GetId(), task.GetGvg().GetFamilyId(), task.GetRedundancyIdx())
	if task.GetFinished() {
		err = s.taskRunner.UpdateMigrateGVGStatus(migrateKey, Migrated)
	} else {
		err = s.taskRunner.UpdateMigrateGVGLastMigratedObjectID(migrateKey, task.GetLastMigratedObjectID())
	}
	return err
}

// AddSwapOutToTaskRunner is used to swap out to task runner from src sp.
func (s *SPExitScheduler) AddSwapOutToTaskRunner(swapOut *virtualgrouptypes.MsgSwapOut) error {
	var (
		err             error
		srcSP           *sptypes.StorageProvider
		swapOutFamilyID uint32
		gvgList         []*virtualgrouptypes.GlobalVirtualGroup
	)
	if srcSP, err = s.manager.baseApp.Consensus().QuerySP(context.Background(), swapOut.GetStorageProvider()); err != nil {
		log.Errorw("failed to add swap out to task runner", "error", err)
		return err
	}
	swapOutFamilyID = swapOut.GetGlobalVirtualGroupFamilyId()
	gvgList = make([]*virtualgrouptypes.GlobalVirtualGroup, 0)

	if swapOutFamilyID != 0 {
		if gvgList, err = s.manager.baseApp.Consensus().ListGlobalVirtualGroupsByFamilyID(context.Background(), srcSP.GetId(), swapOutFamilyID); err != nil {
			log.Errorw("failed to add swap out to task runner due to list virtual groups by family id", "error", err)
			return err
		}
	} else {
		var gvg *virtualgrouptypes.GlobalVirtualGroup
		for _, gvgID := range swapOut.GetGlobalVirtualGroupIds() {
			if gvg, err = s.manager.baseApp.Consensus().QueryGlobalVirtualGroup(context.Background(), gvgID); err != nil {
				log.Errorw("failed to add swap out to task runner due to query gvg", "error", err)
				return err
			}
			gvg.FamilyId = 0
			gvgList = append(gvgList, gvg)
		}
	}

	for _, gvg := range gvgList {
		redundancyIndex := int32(0)
		if gvg.GetFamilyId() == 0 {
			if redundancyIndex, err = util.GetSecondarySPIndexFromGVG(gvg, srcSP.GetId()); err != nil {
				log.Errorw("failed to add swap out to task runner due to get redundancy index", "error", err)
				return err
			}
		}
		gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
			gvg:             gvg,
			redundancyIndex: redundancyIndex,
			isRemoted:       true, // from src sp
			isConflicted:    false,
			isSecondary:     false,
			swapOutKey:      GetSwapOutKey(swapOut),
			srcSP:           srcSP,
			destSP:          s.selfSP,
			migrateStatus:   WaitForMigrate,
		}
		if err = s.taskRunner.AddNewMigrateGVGUnit(gUnit); err != nil {
			log.Errorw("failed to add swap out to task runner", "error", err)
			return err
		}
	}
	return s.taskRunner.addSwapOut(swapOut)
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
				plan, err := s.produceSwapOutPlan()
				if err != nil {
					log.Errorw("failed to produce sp exit execute plan", "error", err)
					return
				}
				if startErr := plan.Start(); startErr != nil {
					log.Errorw("failed to start sp exit execute plan", "error", startErr)
					return
				}

				s.swapOutPlan = plan
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
			if !s.isExiting {
				return
			}
			if s.lastSubscribedSwapOutBlockHeight >= s.lastSubscribedSPExitBlockHeight {
				return
			}

			swapOutEvents, subscribeError := s.manager.baseApp.GfSpClient().ListSwapOutEvents(context.Background(), s.lastSubscribedSwapOutBlockHeight+1, s.selfSP.GetId())
			if subscribeError != nil {
				log.Errorw("failed to subscribe swap out event", "error", subscribeError)
				return
			}
			for _, swapOutEvent := range swapOutEvents {
				if swapOutEvent.GetCompleteEvents() != nil {
					s.updateSPExitExecutePlan(swapOutEvent.GetCompleteEvents())
				}
				// TODO: support cancel event.
			}
			updateErr := s.manager.baseApp.GfSpDB().UpdateSwapOutSubscribeProgress(s.lastSubscribedSwapOutBlockHeight + 1)
			if updateErr != nil {
				log.Errorw("failed to update swap out progress", "error", updateErr)
				return
			}
			s.lastSubscribedSwapOutBlockHeight++
			log.Infow("swap out subscribe progress", "last_subscribed_block_height", s.lastSubscribedSwapOutBlockHeight)

		}
	}
}

func (s *SPExitScheduler) updateSPExitExecutePlan(event *virtualgrouptypes.EventCompleteSwapOut) error {
	return s.swapOutPlan.CheckAndSendCompleteSPExitTx(event)
}

// TODO: swap out approval + swap out approval
func (s *SPExitScheduler) produceSwapOutPlan() (*SrcSPSwapOutPlan, error) {
	var (
		err              error
		vgfList          []*virtualgrouptypes.GlobalVirtualGroupFamily
		secondaryGVGList []*virtualgrouptypes.GlobalVirtualGroup
		plan             *SrcSPSwapOutPlan
	)

	if vgfList, err = s.manager.baseApp.Consensus().ListVirtualGroupFamilies(context.Background(), s.selfSP.GetId()); err != nil {
		log.Errorw("failed to list virtual group family", "error", err)
		return plan, err
	}
	plan = NewSrcSPSwapOutPlan(s.manager, s, s.manager.virtualGroupManager)
	for _, f := range vgfList {
		conflictChecker := NewFamilyConflictChecker(f, plan, s.selfSP)
		swapOutUnits, getFamilySwapOutErr := conflictChecker.GenerateSwapOutUnits()
		if getFamilySwapOutErr != nil {
			log.Errorw("failed to produce swap out plan", "error", getFamilySwapOutErr)
			return nil, getFamilySwapOutErr
		}
		for _, sUnit := range swapOutUnits {
			plan.swapOutUnitMap[GetSwapOutKey(sUnit.swapOut)] = sUnit
		}
	}
	if secondaryGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsBySecondarySP(context.Background(), s.selfSP.GetId()); err != nil {
		log.Errorw("failed to list secondary virtual group", "error", err)
		return plan, err
	}
	for _, g := range secondaryGVGList {
		var destSecondarySP *sptypes.StorageProvider
		if destSecondarySP, err = plan.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(g.GetSecondarySpIds())); err != nil {
			log.Errorw("failed to start migrate execute plan due to get secondary dest sp", "gvg_unit", g, "error", err)
			return plan, err
		}

		swapOut := &virtualgrouptypes.MsgSwapOut{
			StorageProvider:            s.selfSP.GetOperatorAddress(),
			GlobalVirtualGroupFamilyId: 0,
			GlobalVirtualGroupIds:      []uint32{g.GetId()},
			SuccessorSpId:              destSecondarySP.GetId(),
		}
		// TODO: get secondary swap out approval
		// TODO: send secondary swap out tx

		sUnit := &SwapOutUnit{
			isFamily:     false,
			isConflicted: false,
			isSecondary:  true,
			swapOut:      swapOut,
		}
		plan.swapOutUnitMap[GetSwapOutKey(sUnit.swapOut)] = sUnit
	}

	err = plan.storeToDB()
	return plan, err
}

// PickDestSPFilter is used to pick sp id which is not in excluded sp ids.
// which is used by src sp to pick dest sp.
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

// SwapOutUnit is used by swap out plan and task runner.
type SwapOutUnit struct {
	isFamily           bool                          // is used by src sp.
	isConflicted       bool                          // is used by src sp.
	conflictedFamilyID uint32                        // is meaningful when swap out is conflicted
	isSecondary        bool                          // is used by src sp.
	swapOut            *virtualgrouptypes.MsgSwapOut // is used by srd/dest sp.

	// check completed swap out's gvg, and send tx
	completedGVGMutex sync.RWMutex
	completedGVG      map[uint32]*GlobalVirtualGroupMigrateExecuteUnit // is used by dest sp.
}

// CheckAndSendCompleteSwapOutTx check whether complete swap out's gvg, if all finish, send tx to chain.
func (s *SwapOutUnit) CheckAndSendCompleteSwapOutTx(gUnit *GlobalVirtualGroupMigrateExecuteUnit, runner *DestSPTaskRunner) error {
	s.completedGVGMutex.Lock()
	defer s.completedGVGMutex.Unlock()
	s.completedGVG[gUnit.gvg.GetId()] = gUnit

	needCompleted := make([]uint32, 0)
	if s.isFamily {
		srcSP, err := runner.manager.baseApp.Consensus().QuerySP(context.Background(), s.swapOut.GetStorageProvider())
		if err != nil {
			return err
		}
		familyGVGs, err := runner.manager.baseApp.Consensus().ListGlobalVirtualGroupsByFamilyID(context.Background(), srcSP.GetId(), s.swapOut.GetGlobalVirtualGroupFamilyId())
		if err != nil {
			return err
		}
		for _, g := range familyGVGs {
			needCompleted = append(needCompleted, g.GetId())
		}
	} else {
		for _, gvgID := range s.swapOut.GetGlobalVirtualGroupIds() {
			needCompleted = append(needCompleted, gvgID)
		}
	}

	for _, gvgID := range needCompleted {
		if _, found := s.completedGVG[gvgID]; !found {
			return nil
		}
	}

	// TODO: send complete swap out tx.
	return nil
}

func GetSwapOutKey(swapOut *virtualgrouptypes.MsgSwapOut) string {
	if swapOut.GetGlobalVirtualGroupFamilyId() == 0 {
		return util.Uint32ToString(swapOut.GetGlobalVirtualGroupFamilyId())
	} else {
		return util.Uint32SliceToString(swapOut.GetGlobalVirtualGroupIds())
	}
}

func GetEventSwapOutKey(swapOut *virtualgrouptypes.EventCompleteSwapOut) string {
	if swapOut.GetGlobalVirtualGroupFamilyId() == 0 {
		return util.Uint32ToString(swapOut.GetGlobalVirtualGroupFamilyId())
	} else {
		return util.Uint32SliceToString(swapOut.GetGlobalVirtualGroupIds())
	}
}

// SrcSPSwapOutPlan is used to record the execution of swap out.
type SrcSPSwapOutPlan struct {
	manager             *ManageModular
	scheduler           *SPExitScheduler
	virtualGroupManager vgmgr.VirtualGroupManager

	// src sp swap unit plan.
	swapOutUnitMapMutex sync.RWMutex
	swapOutUnitMap      map[string]*SwapOutUnit
	completedSwapOut    map[string]*SwapOutUnit
}

func NewSrcSPSwapOutPlan(m *ManageModular, s *SPExitScheduler, v vgmgr.VirtualGroupManager) *SrcSPSwapOutPlan {
	return &SrcSPSwapOutPlan{
		manager:             m,
		scheduler:           s,
		virtualGroupManager: v,
		swapOutUnitMap:      make(map[string]*SwapOutUnit),
		completedSwapOut:    make(map[string]*SwapOutUnit),
	}
}

// add family swap out if all conflicted is resolved.
func (plan *SrcSPSwapOutPlan) recheckConflictAndAddFamilySwapOut(s *SwapOutUnit) error {
	familyGVGs, err := plan.manager.baseApp.Consensus().ListGlobalVirtualGroupsByFamilyID(context.Background(), plan.scheduler.selfSP.GetId(), s.conflictedFamilyID)
	if err != nil {
		return err
	}
	familySecondarySPIDMap := make(map[uint32]int)
	for _, gvg := range familyGVGs {
		for _, secondarySPID := range gvg.GetSecondarySpIds() {
			familySecondarySPIDMap[secondarySPID] = familySecondarySPIDMap[secondarySPID] + 1
		}
	}
	destFamilySP, err := plan.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithMap(familySecondarySPIDMap))
	if err != nil {
		// still has conflict
		return nil
	}

	swapOut := &virtualgrouptypes.MsgSwapOut{
		StorageProvider:            plan.scheduler.selfSP.GetOperatorAddress(),
		GlobalVirtualGroupFamilyId: s.conflictedFamilyID,
		SuccessorSpId:              destFamilySP.GetId(),
	}

	// TODO: get family swap out approval
	// TODO: send start swap out tx
	sUnit := &SwapOutUnit{
		isFamily:     true,
		isConflicted: false,
		isSecondary:  false,
		swapOut:      swapOut,
	}

	plan.swapOutUnitMap[GetSwapOutKey(sUnit.swapOut)] = sUnit
	// TODO: store to db.

	return nil
}

func (plan *SrcSPSwapOutPlan) checkAllCompletedAndSendCompleteSPExitTx() error {
	// check completed
	for key, runningSwapOut := range plan.swapOutUnitMap {
		if _, found := plan.completedSwapOut[key]; !found {
			return nil
		}
		if runningSwapOut.isConflicted {
			if _, found := plan.completedSwapOut[util.Uint32ToString(runningSwapOut.conflictedFamilyID)]; !found {
				return nil
			}
		}
	}

	// TODO: send complete sp exit tx.

	return nil
}

func (plan *SrcSPSwapOutPlan) CheckAndSendCompleteSPExitTx(event *virtualgrouptypes.EventCompleteSwapOut) error {
	var err error
	plan.swapOutUnitMapMutex.Lock()
	defer plan.swapOutUnitMapMutex.Unlock()

	if _, found := plan.swapOutUnitMap[GetEventSwapOutKey(event)]; !found {
		return fmt.Errorf("not found swap out key")
	}

	unit := plan.swapOutUnitMap[GetEventSwapOutKey(event)]
	if unit.isConflicted {
		err = plan.recheckConflictAndAddFamilySwapOut(unit)
		if err != nil {
			return err
		}
	}
	return plan.checkAllCompletedAndSendCompleteSPExitTx()
}

// LoadFromDB is used to rebuild the memory plan topology.
func (plan *SrcSPSwapOutPlan) LoadFromDB() error {
	// TODO:
	return nil
}

// it is called at start of the execute plan.
func (plan *SrcSPSwapOutPlan) storeToDB() error {
	// TODO: swapOutUnitMap
	return nil
}

// NotifyDestSPIterator is used to notify/check migrate units to dest sp.
type NotifyDestSPIterator struct {
	plan        *SrcSPSwapOutPlan
	notifyIndex int
	swapOuts    []*virtualgrouptypes.MsgSwapOut
}

func NewNotifyDestSPIterator(plan *SrcSPSwapOutPlan) *NotifyDestSPIterator {
	plan.swapOutUnitMapMutex.Lock()
	defer plan.swapOutUnitMapMutex.Unlock()

	iter := &NotifyDestSPIterator{
		plan:        plan,
		notifyIndex: 0,
		swapOuts:    make([]*virtualgrouptypes.MsgSwapOut, 0),
	}

	for _, s := range plan.swapOutUnitMap {
		iter.swapOuts = append(iter.swapOuts, s.swapOut)
	}

	return iter
}

func (iter *NotifyDestSPIterator) Valid() bool {
	return iter.notifyIndex < len(iter.swapOuts)
}

func (iter *NotifyDestSPIterator) Next() {
	iter.notifyIndex++
}

func (iter *NotifyDestSPIterator) Value() *virtualgrouptypes.MsgSwapOut {
	return iter.swapOuts[iter.notifyIndex]
}

func (plan *SrcSPSwapOutPlan) startSrcSPSchedule() {
	// notify dest sp start migrate swap out and check them migrate status.
	go plan.notifyDestSPSwapOut()
}

// dispatch swap out to corresponding dest sp.
func (plan *SrcSPSwapOutPlan) notifyDestSPSwapOut() {
	var (
		err              error
		notifyLoopNumber uint64
		notifyUnitNumber uint64
	)
	for {
		time.Sleep(10 * time.Second)
		notifyLoopNumber++
		notifyUnitNumber = 0
		iter := NewNotifyDestSPIterator(plan)
		for ; iter.Valid(); iter.Next() {
			notifyUnitNumber++
			swapOut := iter.Value()

			sp, querySPError := iter.plan.virtualGroupManager.QuerySPByID(swapOut.GetSuccessorSpId())
			if querySPError != nil {
				log.Infow("failed to notify swap out due to query successor sp id", "error", querySPError)
				continue
			}

			err = plan.manager.baseApp.GfSpClient().NotifyDestSPMigrateSwapOut(context.Background(), sp.GetEndpoint(), swapOut)
			log.Infow("notify dest sp swap out", "swap_out", swapOut, "error", err)

		}
		log.Infow("notify swap out to dest sp", "loop_number", notifyLoopNumber, "notify_number", notifyUnitNumber)
	}
}

// Start persist plan and task to db and task dispatcher
func (plan *SrcSPSwapOutPlan) Start() error {
	go plan.startSrcSPSchedule()
	return nil
}

// DestSPTaskRunner is used to manage task migrate progress/status in dest sp.
type DestSPTaskRunner struct {
	manager             *ManageModular
	virtualGroupManager vgmgr.VirtualGroupManager
	mutex               sync.RWMutex
	keyIndexMap         map[string]int
	gvgUnits            []*GlobalVirtualGroupMigrateExecuteUnit
	swapOutUnitMap      map[string]*SwapOutUnit
}

func NewDestSPTaskRunner(m *ManageModular, v vgmgr.VirtualGroupManager) *DestSPTaskRunner {
	return &DestSPTaskRunner{
		manager:             m,
		virtualGroupManager: v,
		keyIndexMap:         make(map[string]int),
		gvgUnits:            make([]*GlobalVirtualGroupMigrateExecuteUnit, 0),
		swapOutUnitMap:      make(map[string]*SwapOutUnit),
	}
}

func (runner *DestSPTaskRunner) addGVGUnit(gvgMeta *spdb.MigrateGVGUnitMeta) error {
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
	if _, found := runner.keyIndexMap[gUnit.Key()]; found {
		return nil
	}
	runner.gvgUnits = append(runner.gvgUnits, gUnit)
	runner.keyIndexMap[gUnit.Key()] = len(runner.gvgUnits) - 1
	return nil
}

func (runner *DestSPTaskRunner) addSwapOut(swapOut *virtualgrouptypes.MsgSwapOut) error {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	if _, found := runner.swapOutUnitMap[GetSwapOutKey(swapOut)]; found {
		return nil
	}
	runner.swapOutUnitMap[GetSwapOutKey(swapOut)] = &SwapOutUnit{
		swapOut:      swapOut,
		completedGVG: make(map[uint32]*GlobalVirtualGroupMigrateExecuteUnit),
	}
	return nil
}

func (runner *DestSPTaskRunner) LoadFromDB() error {
	// TODO: swap out and gvg
	return nil
}

func (runner *DestSPTaskRunner) Start() error {
	go runner.startDestSPSchedule()
	return nil
}

func (runner *DestSPTaskRunner) UpdateMigrateGVGLastMigratedObjectID(migrateKey string, lastMigratedObjectID uint64) error {
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

func (runner *DestSPTaskRunner) UpdateMigrateGVGStatus(migrateKey string, st MigrateStatus) error {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()

	if _, found := runner.keyIndexMap[migrateKey]; !found {
		return fmt.Errorf("gvg unit is not found")
	}
	index := runner.keyIndexMap[migrateKey]
	if index >= len(runner.gvgUnits) {
		return fmt.Errorf("gvg unit index is invalid")
	}
	unit := runner.gvgUnits[index]
	unit.migrateStatus = st

	if _, found := runner.swapOutUnitMap[unit.swapOutKey]; !found {
		return nil
	}
	if err := runner.swapOutUnitMap[unit.swapOutKey].CheckAndSendCompleteSwapOutTx(unit, runner); err != nil {
		log.Errorw("failed to check and send complete swap out", "error", err)
		return err
	}

	return runner.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateKey, int(st))
}

func (runner *DestSPTaskRunner) AddNewMigrateGVGUnit(remotedGVGUnit *GlobalVirtualGroupMigrateExecuteUnit) error {
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

func (runner *DestSPTaskRunner) startDestSPSchedule() {
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

/*
FamilyConflictChecker
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
type FamilyConflictChecker struct {
	vgf    *virtualgrouptypes.GlobalVirtualGroupFamily
	plan   *SrcSPSwapOutPlan
	selfSP *sptypes.StorageProvider
}

func NewFamilyConflictChecker(f *virtualgrouptypes.GlobalVirtualGroupFamily, p *SrcSPSwapOutPlan, s *sptypes.StorageProvider) *FamilyConflictChecker {
	return &FamilyConflictChecker{
		vgf:    f,
		plan:   p,
		selfSP: s,
	}
}

// GenerateSwapOutUnits generate the family swap out units.
// TODO: swap out approval + swap out approval
func (checker *FamilyConflictChecker) GenerateSwapOutUnits() ([]*SwapOutUnit, error) {
	var (
		err                    error
		familyGVGs             []*virtualgrouptypes.GlobalVirtualGroup
		hasPrimaryGVG          bool
		familySecondarySPIDMap = make(map[uint32]int)
		destFamilySP           *sptypes.StorageProvider
		swapOutUnits           = make([]*SwapOutUnit, 0)
	)
	if familyGVGs, err = checker.plan.manager.baseApp.Consensus().ListGlobalVirtualGroupsByFamilyID(context.Background(), checker.selfSP.GetId(), checker.vgf.GetId()); err != nil {
		log.Errorw("failed to generate swap out units due to list virtual groups by family id", "error", err)
		return nil, err
	}
	for _, gvg := range familyGVGs {
		for _, secondarySPID := range gvg.GetSecondarySpIds() {
			familySecondarySPIDMap[secondarySPID] = familySecondarySPIDMap[secondarySPID] + 1
		}
		hasPrimaryGVG = true
	}
	if hasPrimaryGVG {
		// check conflicts.
		if destFamilySP, err = checker.plan.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithMap(familySecondarySPIDMap)); err != nil {
			// primary family migrate has conflicts, choose a sp with fewer conflicts.
			secondarySPIDBindingLeastGVGs := familyGVGs[0].GetSecondarySpIds()[0]
			for spID, count := range familySecondarySPIDMap {
				if count < familySecondarySPIDMap[secondarySPIDBindingLeastGVGs] {
					secondarySPIDBindingLeastGVGs = spID
				}
			}
			// resolve conflict, swap out secondarySPIDBindingLeastGVGs.
			for _, gvg := range familyGVGs {
				if redundancyIndex, _ := util.GetSecondarySPIndexFromGVG(gvg, secondarySPIDBindingLeastGVGs); redundancyIndex > 0 {
					// gvg has conflicts.
					destSecondarySP, pickErr := checker.plan.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(gvg.GetSecondarySpIds()))
					if pickErr != nil {
						log.Errorw("failed to check conflict due to pick secondary sp", "gvg", gvg, "error", pickErr)
						return nil, pickErr
					}

					swapOut := &virtualgrouptypes.MsgSwapOut{
						StorageProvider:            checker.selfSP.GetOperatorAddress(),
						GlobalVirtualGroupFamilyId: 0,
						GlobalVirtualGroupIds:      []uint32{gvg.GetId()},
						SuccessorSpId:              destSecondarySP.GetId(),
					}
					// TODO: get secondary swap out approval
					// TODO: send start secondary swap out tx

					swapOutUnits = append(swapOutUnits, &SwapOutUnit{
						isFamily:           false,
						isConflicted:       true,
						conflictedFamilyID: checker.vgf.GetId(),
						isSecondary:        true,
						swapOut:            swapOut,
					})
				}
			}
		} else { // has no conflicts
			swapOut := &virtualgrouptypes.MsgSwapOut{
				StorageProvider:            checker.selfSP.GetOperatorAddress(),
				GlobalVirtualGroupFamilyId: checker.vgf.GetId(),
				SuccessorSpId:              destFamilySP.GetId(),
			}
			// TODO: get family swap out approval
			// TODO: swap start family swap out tx

			swapOutUnits = append(swapOutUnits, &SwapOutUnit{
				isFamily:     true,
				isConflicted: false,
				isSecondary:  false,
				swapOut:      swapOut,
			})
		}
	}
	return swapOutUnits, nil
}
