package manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

//type GlobalVirtualGroupByBucketMigrateExecuteUnit struct {
//
//	gvgMigrateUnit *GlobalVirtualGroupMigrateExecuteUnit
//}

type BucketMigrateExecutePlan struct {
	Manager             *ManageModular
	Scheduler           *BucketMigrateScheduler
	VirtualGroupManager vgmgr.VirtualGroupManager
	BucketID            uint64
	//PrimaryGVGByBucketMigrateUnits []*GlobalVirtualGroupByBucketMigrateExecuteUnit          // bucket migrate，primary gvg
	PrimaryGVGIDMapMigrateUnits map[uint32]*GlobalVirtualGroupMigrateExecuteUnit // gvgID -> GlobalVirtualGroupByBucketMigrateExecuteUnit
}

func (plan *BucketMigrateExecutePlan) loadFromDB() error {
	// TODO: MigrateDB interface impl.

	// subscribe progress
	// plan progress
	// task progress
	// multi bucket

	return nil
}
func (plan *BucketMigrateExecutePlan) storeToDB() error {
	// TODO: MigrateDB interface impl.
	var err error
	for _, migrateGVGUnit := range plan.PrimaryGVGIDMapMigrateUnits {
		if err = plan.Manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
			GlobalVirtualGroupID:   migrateGVGUnit.gvg.GetId(),
			VirtualGroupFamilyID:   0,
			MigrateRedundancyIndex: -1,
			BucketID:               plan.BucketID,
			IsSecondary:            false,
			IsConflict:             false,
			SrcSPID:                migrateGVGUnit.srcSP.GetId(),
			DestSPID:               migrateGVGUnit.destSP.GetId(),
			LastMigrateObjectID:    0,
			MigrateStatus:          int(migrateGVGUnit.migrateStatus),
		}); err != nil {
			log.Errorw("failed to store to db", "error", err)
			return err
		}
	}
	return nil
}

func (plan *BucketMigrateExecutePlan) UpdateProgress(task task.MigrateGVGTask) error {
	// TODO: MigrateDB interface impl.
	gvgID := task.GetGvg().GetId()
	migrateExecuteUnit, ok := plan.PrimaryGVGIDMapMigrateUnits[gvgID]
	if ok {
		// TODO: Task if finished, deleted from PrimaryGVGIDMapMigrateUnits
		migrateExecuteUnit.lastMigrateObjectID = task.GetLastMigratedObjectId()
	} else {
		return errors.New("no such migrate gvg task")
	}

	// TODO: update memory and db.
	// TODO : if finished,
	migrateStatus := Migrating
	err := plan.Manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(&spdb.MigrateGVGUnitMeta{
		GlobalVirtualGroupID:   migrateExecuteUnit.gvg.GetId(),
		MigrateRedundancyIndex: migrateExecuteUnit.redundantIndex,
		BucketID:               plan.BucketID,
		IsSecondary:            migrateExecuteUnit.isSecondary,
		IsConflict:             migrateExecuteUnit.isConflict,
		LastMigrateObjectID:    task.GetLastMigratedObjectId(),
	}, int(migrateStatus))
	log.Debugw("update migrate gvg progress", "gvg_meta", migrateExecuteUnit, "error", err)
	return nil
}

func (plan *BucketMigrateExecutePlan) startSPSchedule() {
	// TODO: dispatch to task-dispatcher
	for _, migrateGVGUnit := range plan.PrimaryGVGIDMapMigrateUnits {
		migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{}
		migrateGVGTask.InitMigrateGVGTask(plan.Manager.baseApp.TaskPriority(migrateGVGTask),
			plan.BucketID, migrateGVGUnit.gvg, migrateGVGUnit.redundantIndex,
			migrateGVGUnit.srcSP, migrateGVGUnit.destSP)
		// check error
		err := plan.Manager.migrateGVGQueue.Push(migrateGVGTask)
		log.Errorw("failed to push migrate gvg task to queue", "error", err)
	}
}

func (plan *BucketMigrateExecutePlan) makeGVGUnit(gvgMeta []*spdb.MigrateGVGUnitMeta) error {
	var (
		primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup
		err              error
	)

	for _, migrateGVG := range gvgMeta {
		srcSP, queryErr := plan.Manager.virtualGroupManager.QuerySPByID(migrateGVG.SrcSPID)
		destSP, queryErr := plan.Manager.virtualGroupManager.QuerySPByID(migrateGVG.DestSPID)
		if queryErr != nil {
			log.Errorw("failed to query sp", "error", queryErr)
			return queryErr
		}
		primarySPGVGList, err = plan.Manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), uint64(plan.Scheduler.selfSP.GetId()))
		if err != nil {
			log.Errorw("failed to list gvg ", "error", err)
			return errors.New("failed to list gvg")
		}
		for _, gvg := range primarySPGVGList {
			gvgUnit := &GlobalVirtualGroupMigrateExecuteUnit{gvg: gvg, srcSP: srcSP, destSP: destSP}
			plan.PrimaryGVGIDMapMigrateUnits[gvg.Id] = gvgUnit
		}
	}
	return nil
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
	go plan.startSPSchedule()
	return nil
}

// BucketMigrateScheduler subscribes bucket migrate events and produces a gvg migrate plan.
// TODO: support multiple buckets migrate
type BucketMigrateScheduler struct {
	manager                   *ManageModular
	selfSP                    *sptypes.StorageProvider
	lastSubscribedBlockHeight uint64 // load from db
	isExiting                 bool   // load from db
	//executePlan               []*BucketMigrateExecutePlan // bucketid -> BucketMigrateExecutePlan
	executePlanIDMap map[uint64]*BucketMigrateExecutePlan
}

// NewBucketMigrateScheduler returns a virtual group manager interface.
func NewBucketMigrateScheduler(manager *ManageModular) (*BucketMigrateScheduler, error) {
	bucketMigrateScheduler := &BucketMigrateScheduler{
		manager: manager,
	}

	bucketMigrateScheduler.Init()
	bucketMigrateScheduler.Start()

	return bucketMigrateScheduler, nil
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
		log.Errorw("failed to init bucket migrate Scheduler due to init subscribe migrate bucket progress", "error", err)
		return err
	}
	s.executePlanIDMap = make(map[uint64]*BucketMigrateExecutePlan)

	// plan load from db, TODO (multi execute plan)
	executePlan := &BucketMigrateExecutePlan{
		Manager:                     s.manager,
		Scheduler:                   s,
		VirtualGroupManager:         s.manager.virtualGroupManager,
		PrimaryGVGIDMapMigrateUnits: make(map[uint32]*GlobalVirtualGroupMigrateExecuteUnit),
	}
	if err = executePlan.Init(); err != nil {
		log.Errorw("failed to init sp exit Scheduler due to plan init", "error", err)
		return err
	}
	s.executePlanIDMap[executePlan.BucketID] = executePlan

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
				migrationBucketEvents *types.ListMigrateBucketEvents
				err                   error
			)
			// 1. subscribe metadata event error,
			// TODO: will replace GfBsDB() to GfClient()
			migrationBucketEvents, err = s.manager.baseApp.GfSpClient().ListMigrateBucketEvents(context.Background(), s.lastSubscribedBlockHeight+1, s.selfSP.GetId())
			if err != nil {
				log.Errorw("failed to list migrate bucket events", "error", err)
				return
			}
			// 2. make plan, start plan
			for _, event := range migrationBucketEvents.Events {
				// TODO plan ?
				plan, _ := s.produceBucketMigrateExecutePlan(event)
				if err := plan.Start(); err != nil {
					log.Errorw("failed to start bucket migrate execute plan", "error", err)
					return
				}
				s.executePlanIDMap[plan.BucketID] = plan
			}

			// TODO: update subscribe progress to db
			// 3.update plan
			updateErr := s.manager.baseApp.GfSpDB().UpdateBucketMigrateSubscribeProgress(s.lastSubscribedBlockHeight + 1)
			if updateErr != nil {
				log.Errorw("failed to update sp exit progress", "error", updateErr)
				return
			}

			s.isExiting = true
			s.lastSubscribedBlockHeight++
		}
	}
}

func (s *BucketMigrateScheduler) produceBucketMigrateExecutePlan(event *storage_types.EventMigrationBucket) (*BucketMigrateExecutePlan, error) {
	var (
		primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup
		plan             *BucketMigrateExecutePlan
		err              error
	)

	plan = &BucketMigrateExecutePlan{
		VirtualGroupManager:         s.manager.virtualGroupManager,
		BucketID:                    event.BucketId.Uint64(),
		PrimaryGVGIDMapMigrateUnits: make(map[uint32]*GlobalVirtualGroupMigrateExecuteUnit),
	}

	// TODO: query metadata service to get primary sp's gvg list.
	primarySPGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), uint64(s.selfSP.GetId()))
	if err != nil {
		log.Errorw("failed to list gvg ", "error", err)
		return nil, errors.New("failed to list gvg")
	}

	srcSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(s.selfSP.GetId())
	destSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(event.DstPrimarySpId)
	if queryErr != nil {
		log.Errorw("failed to query sp", "error", queryErr)
		return nil, queryErr
	}
	for _, gvg := range primarySPGVGList {
		gvgUnit := &GlobalVirtualGroupMigrateExecuteUnit{gvg: gvg, srcSP: srcSP, destSP: destSP}
		plan.PrimaryGVGIDMapMigrateUnits[gvg.Id] = gvgUnit
	}

	return plan, nil
}

func (s *BucketMigrateScheduler) HandleMigrateGVGTask(task task.MigrateGVGTask) error {
	err := s.executePlanIDMap[task.GetBucketId()].UpdateProgress(task)
	return err
}

func (s *BucketMigrateScheduler) LoadBucketMigrateExecutePlansFromDB() error {
	var (
		bucketIDs             []uint64
		migrationBucketEvents *types.ListMigrateBucketEvents
		migrateGVGUnitMeta    []*spdb.MigrateGVGUnitMeta
		err                   error
	)

	// get bucket id, TODO: if you have any good idea
	migrationBucketEvents, err = s.manager.baseApp.GfSpClient().ListMigrateBucketEvents(context.Background(), s.lastSubscribedBlockHeight+1, s.selfSP.GetId())
	if err != nil {
		log.Errorw("failed to list migrate bucket events", "error", err)
		return errors.New("failed to list migrate bucket events")
	}
	for _, event := range migrationBucketEvents.Events {
		bucketIDs = append(bucketIDs, event.BucketId.Uint64())
	}

	for _, bucketID := range bucketIDs {
		migrateGVGUnitMeta, err = s.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByBucketID(bucketID, s.selfSP.GetId())
		if err != nil {
		}
		// construct plan
		// plan load from db, TODO (multi execute plan)
		executePlan := &BucketMigrateExecutePlan{
			Manager:                     s.manager,
			Scheduler:                   s,
			BucketID:                    bucketID,
			VirtualGroupManager:         s.manager.virtualGroupManager,
			PrimaryGVGIDMapMigrateUnits: make(map[uint32]*GlobalVirtualGroupMigrateExecuteUnit),
		}

		if err = executePlan.makeGVGUnit(migrateGVGUnitMeta); err != nil {
			log.Errorw("failed to init sp exit Scheduler due to plan init", "error", err)
			return err
		}
		s.executePlanIDMap[executePlan.BucketID] = executePlan
	}
	return err
}
