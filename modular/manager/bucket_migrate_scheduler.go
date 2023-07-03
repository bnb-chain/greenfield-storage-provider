package manager

import (
	"context"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
)

type GlobalVirtualGroupByBucketMigrateExecuteUnit struct {
	bucketID       uint64
	gvgMigrateUnit *GlobalVirtualGroupMigrateExecuteUnit
}

type BucketMigrateExecutePlan struct {
	virtualGroupManager            vgmgr.VirtualGroupManager
	PrimaryGVGByBucketMigrateUnits []*GlobalVirtualGroupByBucketMigrateExecuteUnit // bucket migrate，primary gvg
}

func (plan *BucketMigrateExecutePlan) loadFromDB() error {
	// TODO: MigrateDB interface impl.

	// subscribe progress
	// plan progress
	// task progress
	return nil
}
func (plan *BucketMigrateExecutePlan) storeToDB() error {
	// TODO: MigrateDB interface impl.

	// TODO:
	return nil
}

func (plan *BucketMigrateExecutePlan) updateProgress() error {
	// TODO: MigrateDB interface impl.

	// TODO: update memory and db.
	return nil
}

func (plan *BucketMigrateExecutePlan) startSrcSPSchedule() {
	// notify dest sp start migrate gvg and check them migrate status.
	go plan.notifyDestSPMigrateExecuteUnits()
	go plan.checkDestSPMigrateExecuteUnitsStatus()
}

func (plan *BucketMigrateExecutePlan) startDestSPSchedule() {
	// TODO:
	// dispatch to task-dispatcher
}

func (plan *BucketMigrateExecutePlan) notifyDestSPMigrateExecuteUnits() {
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
		log.Infow("dispatch migrate unit to dest sp", "loop_number", dispatchLoopNumber, "dispatch_number", dispatchUnitNumber)
	}
}

func (plan *BucketMigrateExecutePlan) checkDestSPMigrateExecuteUnitsStatus() {
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
		log.Infow("check migrating unit status", "loop_number", checkLoopNumber, "check_number", checkUnitNumber)
	}
}

func (plan *BucketMigrateExecutePlan) Init() error {
	plan.loadFromDB()
	return nil
}

func (plan *BucketMigrateExecutePlan) Start() error {
	var err error
	if err = plan.storeToDB(); err != nil {
		log.Errorw("failed to start migrate execute plan due to store db", "error", err)
		return err
	}
	//go plan.startSrcSPSchedule()
	return nil
}

// BucketMigrateScheduler subscribes bucket migrate events and produces a gvg migrate plan.
// TODO: support multiple buckets migrate
type BucketMigrateScheduler struct {
	manager                   *ManageModular
	selfSP                    *sptypes.StorageProvider
	spID                      uint32
	lastSubscribedBlockHeight uint64 // load from db
	isExiting                 bool   // load from db
	executePlan               *BucketMigrateExecutePlan
}

func (s *BucketMigrateScheduler) Init() error {
	if s.manager == nil {
		return fmt.Errorf("manger is nil")
	}
	spInfo, err := s.manager.baseApp.Consensus().QuerySP(context.Background(), s.manager.baseApp.OperatorAddress())
	if err != nil {
		return err
	}
	s.selfSP = spInfo
	s.isExiting = spInfo.GetStatus() == sptypes.STATUS_GRACEFUL_EXITING
	if s.lastSubscribedBlockHeight, err = s.manager.baseApp.GfSpDB().QueryBucketMigrateSubscribeProgress(); err != nil {
		log.Errorw("failed to init bucket migrate scheduler due to init subscribe migrate bucket progress", "error", err)
		return err
	}
	return nil
}

func (s *BucketMigrateScheduler) Start() error {
	go s.subscribeEvents()
	return nil
}

func (s *BucketMigrateScheduler) subscribeEvents() {
	subscribeBucketMigrateEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeBucketMigrateEventInterval) * time.Second)
	for {
		select {
		case <-subscribeBucketMigrateEventsTicker.C:
			var (
				migrationBucketEvents         []*model.EventMigrationBucket
				migrationBucketCompleteEvents []*model.EventCompleteMigrationBucket
				err                           error
			)
			// 1. subscribe metadata event
			migrationBucketEvents, migrationBucketCompleteEvents, err = s.manager.baseApp.GfBsDB().ListMigrateBucketEvents(s.lastSubscribedBlockHeight, s.spID)
			if err != nil {
				log.Errorw("failed to list migrate bucket events", "error", err)
				return
			}
			// 2. make plan, start plan
			for _, event := range migrationBucketEvents {
				// TODO plan ?
				plan, _ := s.produceBucketMigrateExecutePlan(event)
				if err := plan.Start(); err != nil {
					log.Errorw("failed to start bucket migrate execute plan", "error", err)
					return
				}
			}

			for _, _ = range migrationBucketCompleteEvents {
				// TODO Complete event ?
			}
			// TODO: pdate subscribe progress to db
			// 3.update plan
		}
	}
}

func (s *BucketMigrateScheduler) produceBucketMigrateExecutePlan(event *model.EventMigrationBucket) (*BucketMigrateExecutePlan, error) {
	var (
		primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup
		plan             *BucketMigrateExecutePlan
	)

	plan = &BucketMigrateExecutePlan{
		virtualGroupManager:            s.manager.virtualGroupManager,
		PrimaryGVGByBucketMigrateUnits: make([]*GlobalVirtualGroupByBucketMigrateExecuteUnit, 0),
	}

	// TODO: query metadata service to get primary sp's gvg list.
	srcSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(s.spID)
	if queryErr != nil {
		log.Errorw("failed to query sp", "error", queryErr)
		return nil, queryErr
	}
	destSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(event.DstPrimarySpId)
	if queryErr != nil {
		log.Errorw("failed to query sp", "error", queryErr)
		return nil, queryErr
	}
	for _, g := range primarySPGVGList {
		gvgUnit := &GlobalVirtualGroupMigrateExecuteUnit{gvg: g, srcSP: srcSP, destSP: destSP}
		plan.PrimaryGVGByBucketMigrateUnits = append(plan.PrimaryGVGByBucketMigrateUnits, &GlobalVirtualGroupByBucketMigrateExecuteUnit{bucketID: event.ID, gvgMigrateUnit: gvgUnit})
	}

	return plan, nil
}
