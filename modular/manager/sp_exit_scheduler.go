package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
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
		if s.swapOutPlan, err = s.produceSwapOutPlan(true); err != nil {
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
	migrateKey = MakeGVGMigrateKey(task.GetGvg().GetId(), task.GetGvg().GetFamilyId(), task.GetRedundancyIdx())
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
			log.Errorw("failed to add swap out to task runner due to list virtual groups by family id",
				"src_sp_id", srcSP.GetId(), "family_id", swapOutFamilyID, "error", err)
			return err
		}
	} else {
		var gvg *virtualgrouptypes.GlobalVirtualGroup
		for _, gvgID := range swapOut.GetGlobalVirtualGroupIds() {
			if gvg, err = s.manager.baseApp.Consensus().QueryGlobalVirtualGroup(context.Background(), gvgID); err != nil {
				log.Errorw("failed to add swap out to task runner due to query gvg", "error", err)
				return err
			}
			gvgList = append(gvgList, gvg)
		}
	}

	for _, gvg := range gvgList {
		redundancyIndex := int32(-1)
		if gvg.GetFamilyId() == 0 {
			if redundancyIndex, err = util.GetSecondarySPIndexFromGVG(gvg, srcSP.GetId()); err != nil {
				log.Errorw("failed to add swap out to task runner due to get redundancy index", "error", err)
				return err
			}
		}
		gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
			gvg:             gvg,
			redundancyIndex: redundancyIndex,
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
	return s.taskRunner.AddNewSwapOut(swapOut)
}

func (s *SPExitScheduler) subscribeEvents() {
	go func() {
		// subscribeSPExitEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSPExitEventInterval) * time.Second)
		subscribeSPExitEventsTicker := time.NewTicker(100 * time.Millisecond)
		defer subscribeSPExitEventsTicker.Stop()
		for range subscribeSPExitEventsTicker.C {
			spExitEvents, subscribeError := s.manager.baseApp.GfSpClient().ListSpExitEvents(context.Background(), s.lastSubscribedSPExitBlockHeight+1, s.manager.baseApp.OperatorAddress())
			if subscribeError != nil {
				log.Errorw("failed to subscribe sp exit event", "error", subscribeError)
				continue
			}
			log.Infow("loop subscribe sp exit event", "sp_exit_events", spExitEvents, "block_id", s.lastSubscribedSPExitBlockHeight+1, "sp_address", s.manager.baseApp.OperatorAddress())
			if spExitEvents.Event != nil {
				if s.isExiting || s.isExited {
					continue
				}
				plan, err := s.produceSwapOutPlan(false)
				if err != nil {
					log.Errorw("failed to produce sp exit execute plan", "error", err)
					continue
				}
				if startErr := plan.Start(); startErr != nil {
					log.Errorw("failed to start sp exit execute plan", "error", startErr)
					continue
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
				continue
			}
			s.lastSubscribedSPExitBlockHeight++
			log.Infow("sp exit subscribe progress", "last_subscribed_block_height", s.lastSubscribedSPExitBlockHeight)
		}
	}()

	go func() {
		// subscribeSwapOutEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeSwapOutEventInterval) * time.Second)
		subscribeSwapOutEventsTicker := time.NewTicker(100 * time.Millisecond)
		defer subscribeSwapOutEventsTicker.Stop()
		for range subscribeSwapOutEventsTicker.C {
			if s.lastSubscribedSwapOutBlockHeight >= s.lastSubscribedSPExitBlockHeight {
				continue
			}

			swapOutEvents, subscribeError := s.manager.baseApp.GfSpClient().ListSwapOutEvents(context.Background(), s.lastSubscribedSwapOutBlockHeight+1, s.selfSP.GetId())
			if subscribeError != nil {
				log.Errorw("failed to subscribe swap out event", "error", subscribeError)
				continue
			}
			log.Infow("loop subscribe swap out event", "swap_out_events", swapOutEvents, "block_id", s.lastSubscribedSwapOutBlockHeight+1, "sp_id", s.selfSP.GetId())
			for _, swapOutEvent := range swapOutEvents {
				if swapOutEvent.GetCompleteEvents() != nil {
					s.updateSPExitExecutePlan(swapOutEvent.GetCompleteEvents())
				}
				// TODO: support cancel event.
			}
			updateErr := s.manager.baseApp.GfSpDB().UpdateSwapOutSubscribeProgress(s.lastSubscribedSwapOutBlockHeight + 1)
			if updateErr != nil {
				log.Errorw("failed to update swap out progress", "error", updateErr)
				continue
			}
			s.lastSubscribedSwapOutBlockHeight++
			log.Infow("swap out subscribe progress", "last_subscribed_block_height", s.lastSubscribedSwapOutBlockHeight)
		}
	}()
}

func (s *SPExitScheduler) updateSPExitExecutePlan(event *virtualgrouptypes.EventCompleteSwapOut) error {
	return s.swapOutPlan.CheckAndSendCompleteSPExitTx(event)
}

func (s *SPExitScheduler) produceSwapOutPlan(buildMetaByDB bool) (*SrcSPSwapOutPlan, error) {
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
		log.Infow("list vgf", "family", f)
		conflictChecker := NewFamilyConflictChecker(f, plan, s.selfSP)
		swapOutUnits, getFamilySwapOutErr := conflictChecker.GenerateSwapOutUnits(buildMetaByDB)
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
		needSendTX := true
		if buildMetaByDB {
			// check db meta, avoid repeated send tx
			swapOutDBMeta, _ := s.manager.baseApp.GfSpDB().QuerySwapOutUnitInSrcSP(GetSwapOutKey(swapOut))
			if swapOutDBMeta != nil {
				if swapOutDBMeta.SwapOutMsg.SuccessorSpId == swapOut.SuccessorSpId {
					needSendTX = false
				}
			}
		}

		if needSendTX {
			swapOut, err = GetApprovalAndSendTx(plan.manager.baseApp.GfSpClient(), destSecondarySP, swapOut)
			if err != nil {
				return nil, err
			}
		}

		sUnit := &SwapOutUnit{
			isFamily:     false,
			isConflicted: false,
			isSecondary:  true,
			swapOut:      swapOut,
		}
		plan.swapOutUnitMap[GetSwapOutKey(sUnit.swapOut)] = sUnit
	}

	if len(plan.swapOutUnitMap) == 0 {
		// the sp is empty, directly complete sp exit.
		msg := &virtualgrouptypes.MsgCompleteStorageProviderExit{
			StorageProvider: plan.manager.baseApp.OperatorAddress(),
		}
		txHash, sendTxError := plan.manager.baseApp.GfSpClient().CompleteSPExit(context.Background(), msg)
		log.Infow("sp is empty, send complete sp exit tx directly", "tx_hash", txHash, "error", sendTxError)
	}

	log.Infow("succeed to produce swap out plan")
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
		needCompleted = append(needCompleted, s.swapOut.GetGlobalVirtualGroupIds()...)
	}

	for _, gvgID := range needCompleted {
		if _, found := s.completedGVG[gvgID]; !found { // not completed
			return nil
		}
	}

	// all gvg are completed
	msg := &virtualgrouptypes.MsgCompleteSwapOut{
		StorageProvider:            runner.manager.baseApp.OperatorAddress(),
		GlobalVirtualGroupFamilyId: s.swapOut.GetGlobalVirtualGroupFamilyId(),
		GlobalVirtualGroupIds:      s.swapOut.GetGlobalVirtualGroupIds(),
	}
	txHash, err := runner.manager.baseApp.GfSpClient().CompleteSwapOut(context.Background(), msg)
	log.Infow("send complete swap out tx", "swap_out", msg, "tx_hash", txHash, "error", err)
	return err
}

func GetSwapOutKey(swapOut *virtualgrouptypes.MsgSwapOut) string {
	if swapOut.GetGlobalVirtualGroupFamilyId() != 0 {
		return "familyID" + util.Uint32ToString(swapOut.GetGlobalVirtualGroupFamilyId())
	} else {
		return "gvgIDList" + util.Uint32SliceToString(swapOut.GetGlobalVirtualGroupIds())
	}
}

func GetEventSwapOutKey(swapOut *virtualgrouptypes.EventCompleteSwapOut) string {
	if swapOut.GetGlobalVirtualGroupFamilyId() != 0 {
		return "familyID" + util.Uint32ToString(swapOut.GetGlobalVirtualGroupFamilyId())
	} else {
		return "gvgIDList" + util.Uint32SliceToString(swapOut.GetGlobalVirtualGroupIds())
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
	var (
		err                    error
		familyGVGs             []*virtualgrouptypes.GlobalVirtualGroup
		familySecondarySPIDMap = make(map[uint32]int)
		destFamilySP           *sptypes.StorageProvider
	)
	if familyGVGs, err = plan.manager.baseApp.Consensus().ListGlobalVirtualGroupsByFamilyID(context.Background(),
		plan.scheduler.selfSP.GetId(), s.conflictedFamilyID); err != nil {
		return err
	}
	for _, gvg := range familyGVGs {
		for _, secondarySPID := range gvg.GetSecondarySpIds() {
			familySecondarySPIDMap[secondarySPID] = familySecondarySPIDMap[secondarySPID] + 1
		}
	}
	if destFamilySP, err = plan.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithMap(familySecondarySPIDMap)); err != nil {
		// still has conflict
		return nil
	}

	// conflict has resolved, produce family swap out.
	swapOut := &virtualgrouptypes.MsgSwapOut{
		StorageProvider:            plan.scheduler.selfSP.GetOperatorAddress(),
		GlobalVirtualGroupFamilyId: s.conflictedFamilyID,
		SuccessorSpId:              destFamilySP.GetId(),
	}

	swapOut, err = GetApprovalAndSendTx(plan.manager.baseApp.GfSpClient(), destFamilySP, swapOut)
	if err != nil {
		return err
	}
	sUnit := &SwapOutUnit{
		isFamily:     true,
		isConflicted: false,
		isSecondary:  false,
		swapOut:      swapOut,
	}

	plan.swapOutUnitMap[GetSwapOutKey(sUnit.swapOut)] = sUnit

	if err = plan.manager.baseApp.GfSpDB().InsertSwapOutUnit(&spdb.SwapOutMeta{
		SwapOutKey: GetSwapOutKey(sUnit.swapOut),
		IsDestSP:   false,
		SwapOutMsg: sUnit.swapOut,
	}); err != nil {
		log.Infow("failed to store swap out plan to db", "swap_unit", sUnit, "error", err)
		return err
	}
	return nil
}

func (plan *SrcSPSwapOutPlan) checkAllCompletedAndSendCompleteSPExitTx() error {
	// check completed
	for key, runningSwapOut := range plan.swapOutUnitMap {
		if _, found := plan.completedSwapOut[key]; !found { // not completed
			return nil
		}
		if runningSwapOut.isConflicted {
			if _, found := plan.completedSwapOut[util.Uint32ToString(runningSwapOut.conflictedFamilyID)]; !found { // not completed
				return nil
			}
		}
	}

	// all is completed
	msg := &virtualgrouptypes.MsgCompleteStorageProviderExit{
		StorageProvider: plan.manager.baseApp.OperatorAddress(),
	}
	txHash, err := plan.manager.baseApp.GfSpClient().CompleteSPExit(context.Background(), msg)
	log.Infow("send complete sp exit tx", "sp_exit", msg, "tx_hash", txHash, "error", err)
	return err
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

// it is called at start of the execute plan.
func (plan *SrcSPSwapOutPlan) storeToDB() error {
	var err error
	for key, swapOutUnit := range plan.swapOutUnitMap {
		if err = plan.manager.baseApp.GfSpDB().InsertSwapOutUnit(&spdb.SwapOutMeta{
			SwapOutKey: key,
			IsDestSP:   false,
			SwapOutMsg: swapOutUnit.swapOut,
		}); err != nil {
			log.Infow("failed to store swap out plan to db", "error", err)
			return err
		}
	}
	log.Infow("succeed to store swap out plan to db")
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
			log.Infow("notify dest sp swap out", "dest_sp_endpoint", sp.GetEndpoint(), "swap_out", swapOut, "error", err)

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

func (runner *DestSPTaskRunner) LoadFromDB() error {
	var (
		err          error
		swapOutList  []*spdb.SwapOutMeta
		completedMap = make(map[uint32]*GlobalVirtualGroupMigrateExecuteUnit)
	)
	if swapOutList, err = runner.manager.baseApp.GfSpDB().ListDestSPSwapOutUnits(); err != nil {
		return err
	}

	for _, swapOut := range swapOutList {
		for _, completedID := range swapOut.CompletedGVGs {
			completedMap[completedID] = nil
		}
		runner.swapOutUnitMap[swapOut.SwapOutKey] = &SwapOutUnit{
			swapOut:      swapOut.SwapOutMsg,
			completedGVG: completedMap,
		}
		srcSP, querySPError := runner.manager.baseApp.Consensus().QuerySP(context.Background(), swapOut.SwapOutMsg.GetStorageProvider())
		if srcSP != nil {
			log.Errorw("failed to add swap out to task runner", "error", querySPError)
			return querySPError
		}

		if swapOut.SwapOutMsg.GetGlobalVirtualGroupFamilyId() != 0 {
			allGVGList, queryGVGError := runner.manager.baseApp.Consensus().ListGlobalVirtualGroupsByFamilyID(context.Background(),
				srcSP.GetId(), swapOut.SwapOutMsg.GetGlobalVirtualGroupFamilyId())
			if queryGVGError != nil {
				log.Errorw("failed to add swap out to task runner due to list virtual groups by family id", "error", queryGVGError)
				return queryGVGError
			}
			for _, gvg := range allGVGList {
				if _, found := completedMap[gvg.GetId()]; !found {
					migrateKey := MakeGVGMigrateKey(gvg.GetId(), gvg.GetFamilyId(), -1)
					gvgMeta, queryErr := runner.manager.baseApp.GfSpDB().QueryMigrateGVGUnit(migrateKey)
					if queryErr != nil {
						return queryErr
					}
					gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
						gvg:                  gvg,
						redundancyIndex:      gvgMeta.RedundancyIndex,
						swapOutKey:           gvgMeta.SwapOutKey,
						lastMigratedObjectID: gvgMeta.LastMigratedObjectID,
					}
					runner.gvgUnits = append(runner.gvgUnits, gUnit)
					runner.keyIndexMap[gUnit.Key()] = len(runner.gvgUnits) - 1
				}
			}
		} else {
			notFinishedGVGList := make([]*virtualgrouptypes.GlobalVirtualGroup, 0)
			for _, gvgID := range swapOut.SwapOutMsg.GetGlobalVirtualGroupIds() {
				if _, found := completedMap[gvgID]; !found {
					gvg, queryGVGError := runner.manager.baseApp.Consensus().QueryGlobalVirtualGroup(context.Background(), gvgID)
					if queryGVGError != nil {
						log.Errorw("failed to add swap out to task runner due to query gvg", "error", queryGVGError)
						return queryGVGError
					}
					notFinishedGVGList = append(notFinishedGVGList, gvg)
				}
			}
			for _, gvg := range notFinishedGVGList {
				redundancyIndex, getIndexErr := util.GetSecondarySPIndexFromGVG(gvg, srcSP.GetId())
				if getIndexErr != nil {
					log.Errorw("failed to add swap out to task runner due to get redundancy index", "error", getIndexErr)
					return getIndexErr
				}
				migrateKey := MakeGVGMigrateKey(gvg.GetId(), gvg.GetFamilyId(), redundancyIndex)
				gvgMeta, queryErr := runner.manager.baseApp.GfSpDB().QueryMigrateGVGUnit(migrateKey)
				if queryErr != nil {
					return queryErr
				}
				gUnit := &GlobalVirtualGroupMigrateExecuteUnit{
					gvg:                  gvg,
					redundancyIndex:      gvgMeta.RedundancyIndex,
					swapOutKey:           gvgMeta.SwapOutKey,
					lastMigratedObjectID: gvgMeta.LastMigratedObjectID,
				}
				runner.gvgUnits = append(runner.gvgUnits, gUnit)
				runner.keyIndexMap[gUnit.Key()] = len(runner.gvgUnits) - 1
			}
		}
	}

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
		SwapOutKey:           remotedGVGUnit.swapOutKey,
		GlobalVirtualGroupID: remotedGVGUnit.gvg.GetId(),
		VirtualGroupFamilyID: remotedGVGUnit.gvg.GetFamilyId(),
		RedundancyIndex:      remotedGVGUnit.redundancyIndex,
		BucketID:             0,
		SrcSPID:              remotedGVGUnit.srcSP.GetId(),
		DestSPID:             remotedGVGUnit.destSP.GetId(),
		LastMigratedObjectID: 0,
		MigrateStatus:        int(remotedGVGUnit.migrateStatus),
		IsRemoted:            true,
	}); err != nil {
		log.Errorw("failed to store to db", "error", err)
		return err
	}

	return nil
}

func (runner *DestSPTaskRunner) AddNewSwapOut(swapOut *virtualgrouptypes.MsgSwapOut) error {
	var err error
	runner.mutex.Lock()
	if _, found := runner.swapOutUnitMap[GetSwapOutKey(swapOut)]; found {
		runner.mutex.Unlock()
		return nil
	}
	runner.swapOutUnitMap[GetSwapOutKey(swapOut)] = &SwapOutUnit{
		swapOut:      swapOut,
		completedGVG: make(map[uint32]*GlobalVirtualGroupMigrateExecuteUnit),
	}
	runner.mutex.Unlock()

	// add to db
	if err = runner.manager.baseApp.GfSpDB().InsertSwapOutUnit(&spdb.SwapOutMeta{
		SwapOutKey: GetSwapOutKey(swapOut),
		IsDestSP:   true,
		SwapOutMsg: swapOut,
	}); err != nil {
		log.Infow("failed to add new swap out", "swap_unit", swapOut, "error", err)
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
					unit.srcSP, unit.destSP,
					// TODO if add add a new tasktimeout
					runner.manager.baseApp.TaskTimeout(migrateGVGTask, 0),
					runner.manager.baseApp.TaskMaxRetry(migrateGVGTask))
				if err = runner.manager.migrateGVGQueue.Push(migrateGVGTask); err != nil {
					log.Errorw("failed to push migrate gvg task to queue", "error", err)
					time.Sleep(5 * time.Second) // Sleep for 5 seconds before retrying
				}
				if err = runner.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(unit.Key(), int(Migrating)); err != nil {
					log.Errorw("failed to update task status", "error", err)
					time.Sleep(5 * time.Second) // Sleep for 5 seconds before retrying
				}
				unit.migrateStatus = Migrating
				log.Infow("succeed to push migrate gvg task into task dispatcher", "migrate_gvg_task", migrateGVGTask)
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
func (checker *FamilyConflictChecker) GenerateSwapOutUnits(buildMetaByDB bool) ([]*SwapOutUnit, error) {
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

					needSendTX := true
					if buildMetaByDB {
						// check db meta, avoid repeated send tx
						swapOutDBMeta, _ := checker.plan.manager.baseApp.GfSpDB().QuerySwapOutUnitInSrcSP(GetSwapOutKey(swapOut))
						if swapOutDBMeta != nil {
							if swapOutDBMeta.SwapOutMsg.SuccessorSpId == swapOut.SuccessorSpId {
								needSendTX = false
							}
						}
					}

					if needSendTX {
						swapOut, err = GetApprovalAndSendTx(checker.plan.manager.baseApp.GfSpClient(), destSecondarySP, swapOut)
						if err != nil {
							return nil, err
						}
					}

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
			needSendTX := true
			if buildMetaByDB {
				// check db meta, avoid repeated send tx
				swapOutDBMeta, _ := checker.plan.manager.baseApp.GfSpDB().QuerySwapOutUnitInSrcSP(GetSwapOutKey(swapOut))
				if swapOutDBMeta != nil {
					if swapOutDBMeta.SwapOutMsg.SuccessorSpId == swapOut.SuccessorSpId {
						needSendTX = false
					}
				}
			}

			if needSendTX {
				swapOut, err = GetApprovalAndSendTx(checker.plan.manager.baseApp.GfSpClient(), destFamilySP, swapOut)
				if err != nil {
					return nil, err
				}
			}

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

func GetApprovalAndSendTx(client *gfspclient.GfSpClient, destSP *sptypes.StorageProvider, originMsg *virtualgrouptypes.MsgSwapOut) (*virtualgrouptypes.MsgSwapOut, error) {
	ctx := context.Background()
	approvalSwapOut, err := client.GetSwapOutApproval(ctx, destSP.GetEndpoint(), originMsg)
	if err != nil {
		log.Errorw("failed to get swap out approval from dest sp", "dest_sp", destSP.GetEndpoint(), "swap_out_msg", approvalSwapOut, "error", err)
		return nil, err
	}
	if _, err = client.SwapOut(ctx, approvalSwapOut); err != nil {
		log.Errorw("failed to send swap out tx to chain", "swap_out_msg", approvalSwapOut, "error", err)
		return nil, err
	}
	log.Infow("succeed to get approval and send swap out tx", "dest_sp", destSP.GetEndpoint(), "swap_out_msg", approvalSwapOut)
	return approvalSwapOut, nil
}
