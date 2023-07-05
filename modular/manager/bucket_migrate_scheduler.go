package manager

import (
	"context"
	"errors"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type BucketMigrateExecutePlan struct {
	Manager                     *ManageModular
	Scheduler                   *BucketMigrateScheduler
	VirtualGroupManager         vgmgr.VirtualGroupManager
	BucketID                    uint64
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
	var (
		migrateStatus MigrateStatus
	)
	gvgID := task.GetGvg().GetId()
	migrateExecuteUnit, ok := plan.PrimaryGVGIDMapMigrateUnits[gvgID]
	if ok {
		// update memory
		migrateExecuteUnit.lastMigrateObjectID = task.GetLastMigratedObjectID()
		if task.GetFinished() == true {
			migrateStatus = Migrated
		} else {
			migrateStatus = Migrating
		}
		migrateExecuteUnit.migrateStatus = migrateStatus
	} else {
		return errors.New("no such migrate gvg task")
	}

	// update db
	err := plan.Manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(&spdb.MigrateGVGUnitMeta{
		GlobalVirtualGroupID:   migrateExecuteUnit.gvg.GetId(),
		MigrateRedundancyIndex: migrateExecuteUnit.redundantIndex,
		BucketID:               plan.BucketID,
		IsSecondary:            migrateExecuteUnit.isSecondary,
		IsConflict:             migrateExecuteUnit.isConflict,
		LastMigrateObjectID:    task.GetLastMigratedObjectID(),
	}, int(migrateStatus))
	log.Debugw("update migrate gvg progress", "gvg_meta", migrateExecuteUnit, "error", err)
	return nil
}

func (plan *BucketMigrateExecutePlan) startSPSchedule() {
	// dispatch to task-dispatcher, TODO: if CompleteEvents terminate the scheduling
	for {
		select {
		case <-plan.Scheduler.stopSignal:
			return // Terminate the scheduling
		default:
			for _, migrateGVGUnit := range plan.PrimaryGVGIDMapMigrateUnits {
				migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{}
				migrateGVGTask.InitMigrateGVGTask(plan.Manager.baseApp.TaskPriority(migrateGVGTask),
					plan.BucketID, migrateGVGUnit.gvg, migrateGVGUnit.redundantIndex,
					migrateGVGUnit.srcSP, migrateGVGUnit.destSP)
				err := plan.Manager.migrateGVGQueue.Push(migrateGVGTask)
				if err != nil {
					log.Errorw("failed to push migrate gvg task to queue", "error", err)
					time.Sleep(5 * time.Second) // Sleep for 5 seconds before retrying
				}
			}
			// TODO: Update database

			time.Sleep(1 * time.Minute) // Sleep for 1 minute before next iteration
		}
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
	lastSubscribedBlockHeight uint64                               // load from db
	executePlanIDMap          map[uint64]*BucketMigrateExecutePlan // bucketID -> BucketMigrateExecutePlan

	isExiting  bool // load from db
	isExited   bool
	stopSignal chan struct{} // stop schedule
}

// NewBucketMigrateScheduler returns a bucket migrate scheduler instance.
func NewBucketMigrateScheduler(manager *ManageModular) (*BucketMigrateScheduler, error) {
	var err error
	bucketMigrateScheduler := &BucketMigrateScheduler{}
	bucketMigrateScheduler.stopSignal = make(chan struct{})
	if err = bucketMigrateScheduler.Init(manager); err != nil {
		return nil, err
	}
	if err = bucketMigrateScheduler.Start(); err != nil {
		return nil, err
	}
	return bucketMigrateScheduler, nil
}

func (s *BucketMigrateScheduler) Init(m *ManageModular) error {
	s.manager = m
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

	//plan load from db, TODO (multi execute plan)
	s.loadBucketMigrateExecutePlansFromDB()

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
			// 1. subscribe migrate bucket events
			migrationBucketEvents, err = s.manager.baseApp.GfSpClient().ListMigrateBucketEvents(context.Background(), s.lastSubscribedBlockHeight+1, s.selfSP.GetId())
			if err != nil {
				log.Errorw("failed to list migrate bucket events", "error", err)
				return
			}
			// 2. make plan, start plan
			// TODO
			if migrationBucketEvents.CompleteEvents != nil {
				s.isExited = true
			}
			for _, event := range migrationBucketEvents.Events {
				if s.isExiting || s.isExited {
					return
				}
				// TODO plan ?
				plan, _ := s.produceBucketMigrateExecutePlan(event)
				if err := plan.Start(); err != nil {
					log.Errorw("failed to start bucket migrate execute plan", "error", err)
					return
				}
				s.executePlanIDMap[plan.BucketID] = plan
				s.isExiting = true
			}

			// 3.update subscribe progress to db
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

	// query metadata service to get primary sp's gvg list.
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
	err := s.executePlanIDMap[task.GetBucketID()].UpdateProgress(task)
	return err
}

func (s *BucketMigrateScheduler) loadBucketMigrateExecutePlansFromDB() error {
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
		// load from db & construct plan
		migrateGVGUnitMeta, err = s.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByBucketID(bucketID, s.selfSP.GetId())
		if err != nil {
		}
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
