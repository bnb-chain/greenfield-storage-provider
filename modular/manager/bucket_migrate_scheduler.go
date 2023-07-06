package manager

import (
	"context"
	"errors"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type BucketMigrateExecutePlan struct {
	manager    *ManageModular
	bucketID   uint64
	gvgUnitMap map[uint32]*GlobalVirtualGroupMigrateExecuteUnitByBucket // gvgID -> GlobalVirtualGroupByBucketMigrateExecuteUnit
	stopSignal chan struct{}                                            // stop schedule
	finished   int                                                      // used for count the number of successful migrate units
}

func newBucketMigrateExecutePlan(manager *ManageModular, bucketID uint64) *BucketMigrateExecutePlan {
	executePlan := &BucketMigrateExecutePlan{
		manager:    manager,
		bucketID:   bucketID,
		gvgUnitMap: make(map[uint32]*GlobalVirtualGroupMigrateExecuteUnitByBucket),
		stopSignal: make(chan struct{}),
		finished:   0,
	}

	return executePlan
}

// storeToDB persist the BucketMigrateExecutePlan to the database
func (plan *BucketMigrateExecutePlan) storeToDB() error {
	var err error
	for _, migrateGVGUnit := range plan.gvgUnitMap {
		if err = plan.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
			MigrateGVGKey:        migrateGVGUnit.Key(),
			GlobalVirtualGroupID: migrateGVGUnit.gvg.GetId(),
			VirtualGroupFamilyID: migrateGVGUnit.gvg.GetFamilyId(),
			RedundancyIndex:      -1,
			BucketID:             migrateGVGUnit.bucketID,
			IsSecondary:          false,
			IsConflicted:         false,
			IsRemoted:            false,
			SrcSPID:              migrateGVGUnit.srcSP.GetId(),
			DestSPID:             migrateGVGUnit.destSP.GetId(),
			LastMigratedObjectID: migrateGVGUnit.lastMigratedObjectID,
			MigrateStatus:        int(migrateGVGUnit.migrateStatus),
		}); err != nil {
			log.Errorw("failed to store to db", "error", err)
			return err
		}
	}
	return nil
}

// UpdateProgress persistent user updates and periodic progress reporting by Executor
func (plan *BucketMigrateExecutePlan) UpdateProgress(task task.MigrateGVGTask) error {
	var (
		err error
	)
	gvgID := task.GetGvg().GetId()
	migrateExecuteUnit, ok := plan.gvgUnitMap[gvgID]
	if ok {
		// update memory
		migrateExecuteUnit.lastMigratedObjectID = task.GetLastMigratedObjectID()

	} else {
		return errors.New("no such migrate gvg task")
	}

	migrateKey := MakeBucketMigrateKey(migrateExecuteUnit.bucketID, migrateExecuteUnit.gvg.GetId())
	err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitLastMigrateObjectID(migrateKey, task.GetLastMigratedObjectID())
	if err != nil {
		log.Debugw("update migrate gvg migrateGVGUnit lastMigrateObjectID", "migrate_key", migrateKey, "error", err)
		return err
	}

	// update migrate status
	if task.GetFinished() == true {
		err = plan.updateMigrateStatus(migrateKey, migrateExecuteUnit, Migrated)
		if err != nil {
			return err
		}
	}

	return nil
}

func (plan *BucketMigrateExecutePlan) updateMigrateStatus(migrateKey string, migrateExecuteUnit *GlobalVirtualGroupMigrateExecuteUnitByBucket, migrateStatus MigrateStatus) error {
	var (
		err error
	)

	plan.finished++
	migrateExecuteUnit.migrateStatus = migrateStatus

	// update migrateStatus
	err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateKey, int(migrateStatus))
	if err != nil {
		log.Debugw("update migrate gvg migrateGVGUnit migrateStatus", "migrate_key", migrateKey, "error", err)
		return err
	}

	// all migrate units success, send CompleteMigrationBucket to chain
	if plan.finished == len(plan.gvgUnitMap) {
		// BucketName GlobalVirtualGroupFamilyId,GvgMappings
		var bucket *types.Bucket
		bucket, err = plan.manager.baseApp.GfSpClient().GetBucketByBucketID(context.Background(), int64(plan.bucketID), true)
		if err != nil {
			return err
		}
		bucketName := bucket.BucketInfo.BucketName
		var gvgMappings []*storage_types.GVGMapping
		for _, migrateGVGUnit := range plan.gvgUnitMap {
			gvgMappings = append(gvgMappings, &storage_types.GVGMapping{SrcGlobalVirtualGroupId: migrateGVGUnit.gvg.GetId(),
				DstGlobalVirtualGroupId: migrateGVGUnit.destGVGID})
		}

		migrateBucket := &storage_types.MsgCompleteMigrateBucket{Operator: plan.manager.baseApp.OperatorAddress(),
			BucketName: bucketName, GvgMappings: gvgMappings}
		plan.manager.baseApp.GfSpClient().CompleteMigrateBucket(context.Background(), migrateBucket)
	}
	return nil
}

func (plan *BucketMigrateExecutePlan) startSPSchedule() {
	// dispatch to task-dispatcher, TODO: if CompleteEvents terminate the scheduling
	for {
		select {
		case <-plan.stopSignal:
			return // Terminate the scheduling
		default:
			for _, migrateGVGUnit := range plan.gvgUnitMap {
				migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{}
				migrateGVGTask.InitMigrateGVGTask(plan.manager.baseApp.TaskPriority(migrateGVGTask),
					plan.bucketID, migrateGVGUnit.gvg, migrateGVGUnit.redundancyIndex,
					migrateGVGUnit.srcSP, migrateGVGUnit.destSP)
				err := plan.manager.migrateGVGQueue.Push(migrateGVGTask)
				if err != nil {
					log.Errorw("failed to push migrate gvg task to queue", "error", err)
					time.Sleep(5 * time.Second) // Sleep for 5 seconds before retrying
				}
				// Update database: migrateStatus to migrating
				migrateGVGUnit.migrateStatus = Migrating

				// update migrateStatus
				err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateGVGUnit.Key(), int(migrateGVGUnit.migrateStatus))
				if err != nil {
					log.Errorw("update migrate gvg migrateGVGUnit migrateStatus", "gvg_unit", migrateGVGUnit, "error", err)
					return
				}
			}

			time.Sleep(1 * time.Minute) // Sleep for 1 minute before next iteration
		}
	}
}

func (plan *BucketMigrateExecutePlan) stopSPSchedule() {
	plan.stopSignal <- struct{}{}
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

	isExited bool
}

// NewBucketMigrateScheduler returns a bucket migrate scheduler instance.
func NewBucketMigrateScheduler(manager *ManageModular) (*BucketMigrateScheduler, error) {
	var err error
	bucketMigrateScheduler := &BucketMigrateScheduler{}
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
	if s.lastSubscribedBlockHeight, err = s.manager.baseApp.GfSpDB().QueryBucketMigrateSubscribeProgress(); err != nil {
		log.Errorw("failed to init bucket migrate Scheduler due to init subscribe migrate bucket progress", "error", err)
		return err
	}
	s.executePlanIDMap = make(map[uint64]*BucketMigrateExecutePlan)

	// plan load from db
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
				migrationBucketEvents []*types.ListMigrateBucketEvents
				err                   error
				executePlan           *BucketMigrateExecutePlan
			)
			// 1. subscribe migrate bucket events
			migrationBucketEvents, err = s.manager.baseApp.GfSpClient().ListMigrateBucketEvents(context.Background(), s.lastSubscribedBlockHeight+1, s.selfSP.GetId())
			if err != nil {
				log.Errorw("failed to list migrate bucket events", "error", err)
				return
			}
			// 2. make plan, start plan
			for _, migrateBucketEvents := range migrationBucketEvents {
				// when receive chain CompleteMigrationBucket event
				if migrateBucketEvents.CompleteEvents != nil {
					executePlan, err = s.getExecutePlanByBucketID(migrateBucketEvents.CompleteEvents.BucketId.Uint64())
					if err != nil {
						return
					}
					executePlan.stopSPSchedule()
					continue
				}
				if migrateBucketEvents.Events != nil {
					if s.isExited {
						return
					}
					plan, _ := s.produceBucketMigrateExecutePlan(migrateBucketEvents.Events)
					if err := plan.Start(); err != nil {
						log.Errorw("failed to start bucket migrate execute plan", "error", err)
						return
					}
					s.executePlanIDMap[plan.bucketID] = plan
				}
			}

			// 3.update subscribe progress to db
			updateErr := s.manager.baseApp.GfSpDB().UpdateBucketMigrateSubscribeProgress(s.lastSubscribedBlockHeight + 1)
			if updateErr != nil {
				log.Errorw("failed to update sp exit progress", "error", updateErr)
				return
			}

			s.lastSubscribedBlockHeight++
		}
	}
}

func (s *BucketMigrateScheduler) produceBucketMigrateExecutePlan(event *storage_types.EventMigrationBucket) (*BucketMigrateExecutePlan, error) {
	var (
		primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup
		plan             *BucketMigrateExecutePlan
		err              error
		vgfID            uint32
		params           *storage_types.Params
		destGVG          *vgmgr.GlobalVirtualGroupMeta
	)

	plan = newBucketMigrateExecutePlan(s.manager, event.BucketId.Uint64())

	// query metadata service to get primary sp's gvg list.
	primarySPGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), uint64(s.selfSP.GetId()))
	if err != nil {
		log.Errorw("failed to list gvg ", "error", err)
		return nil, errors.New("failed to list gvg")
	}

	srcSP, err := s.manager.virtualGroupManager.QuerySPByID(s.selfSP.GetId())
	if err != nil {
		log.Errorw("failed to query sp", "error", err)
		return nil, err
	}
	destSP, err := s.manager.virtualGroupManager.QuerySPByID(event.DstPrimarySpId)
	if err != nil {
		log.Errorw("failed to query sp", "error", err)
		return nil, err
	}

	vgfID = 0
	if params, err = s.manager.baseApp.Consensus().QueryStorageParamsByTimestamp(context.Background(), time.Now().Unix()); err != nil {
		return nil, err
	}
	for _, gvg := range primarySPGVGList {
		// TODO how to get dest gvg
		destGVG, err = s.manager.pickGlobalVirtualGroupForBucketMigrate(context.Background(), vgfID, params, gvg, destSP)
		vgfID = destGVG.FamilyID
		bucketUnit := newGlobalVirtualGroupMigrateExecuteUnitByBucket(plan.bucketID, gvg, srcSP, destSP, WaitForMigrate, destGVG.ID, 0, false, false, false)
		plan.gvgUnitMap[gvg.Id] = bucketUnit
	}

	return plan, nil
}

func (s *BucketMigrateScheduler) getExecutePlanByBucketID(bucketID uint64) (*BucketMigrateExecutePlan, error) {
	executePlan, ok := s.executePlanIDMap[bucketID]
	if ok {
		return executePlan, nil
	} else {
		// TODO
		return nil, errors.New("no such execute plan")
	}
}

func (s *BucketMigrateScheduler) UpdateMigrateProgress(task task.MigrateGVGTask) error {
	executePlan, err := s.getExecutePlanByBucketID(task.GetBucketID())
	if err != nil {
		return err
	}
	executePlan.UpdateProgress(task)
	return nil
}

// loadBucketMigrateExecutePlansFromDB 1) subscribe progress 2) plan progress 3) task progress
func (s *BucketMigrateScheduler) loadBucketMigrateExecutePlansFromDB() error {
	var (
		bucketIDs             []uint64
		migrationBucketEvents []*types.ListMigrateBucketEvents
		migrateGVGUnitMeta    []*spdb.MigrateGVGUnitMeta
		err                   error
		primarySPGVGList      []*virtualgrouptypes.GlobalVirtualGroup
	)

	// get bucket id from metadata, TODO: if you have any good idea
	migrationBucketEvents, err = s.manager.baseApp.GfSpClient().ListMigrateBucketEvents(context.Background(), s.lastSubscribedBlockHeight+1, s.selfSP.GetId())
	if err != nil {
		log.Errorw("failed to list migrate bucket events", "error", err)
		return errors.New("failed to list migrate bucket events")
	}

	for _, migrateBucketEvents := range migrationBucketEvents {
		// if has CompleteEvents, skip it
		if migrateBucketEvents.CompleteEvents != nil {
			continue
		}
		if migrateBucketEvents.Events != nil {
			bucketIDs = append(bucketIDs, migrateBucketEvents.Events.BucketId.Uint64())
		}
	}
	// load from db by BucketID & construct plan
	for _, bucketID := range bucketIDs {
		migrateGVGUnitMeta, err = s.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByBucketID(bucketID)
		if err != nil {
			return err
		}

		executePlan := newBucketMigrateExecutePlan(s.manager, bucketID)
		// Using migrateGVGUnitMeta to construct PrimaryGVGIDMapMigrateUnits and execute them one by one.
		for _, migrateGVG := range migrateGVGUnitMeta {
			// TODO may reuse
			srcSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(migrateGVG.SrcSPID)
			destSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(migrateGVG.DestSPID)
			if queryErr != nil {
				log.Errorw("failed to query sp", "error", queryErr)
				return queryErr
			}
			primarySPGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), uint64(s.selfSP.GetId()))
			if err != nil {
				log.Errorw("failed to list gvg ", "error", err)
				return errors.New("failed to list gvg")
			}
			for _, gvg := range primarySPGVGList {
				bucketUnit := newGlobalVirtualGroupMigrateExecuteUnitByBucket(bucketID, gvg, srcSP, destSP, WaitForMigrate, migrateGVG.DestSPID, migrateGVG.LastMigratedObjectID,
					migrateGVG.IsSecondary, migrateGVG.IsConflicted, migrateGVG.IsRemoted)
				executePlan.gvgUnitMap[gvg.Id] = bucketUnit
			}
		}

		s.executePlanIDMap[executePlan.bucketID] = executePlan
	}
	return err
}
