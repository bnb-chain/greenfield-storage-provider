package manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

var _ vgmgr.GVGPickFilter = &PickDestGVGFilter{}

// PickDestGVGFilter is used to pick dest gvg for bucket migrate.
type PickDestGVGFilter struct {
	expectedFamilyID       uint32
	expectedSecondarySPIDs []uint32
	expectedMinFreeSize    uint64
}

func NewPickDestGVGFilter(familyID uint32, secondarySPIDs []uint32, minFreeSize uint64) *PickDestGVGFilter {
	return &PickDestGVGFilter{expectedFamilyID: familyID, expectedSecondarySPIDs: secondarySPIDs, expectedMinFreeSize: minFreeSize}
}

func (f *PickDestGVGFilter) CheckFamily(familyID uint32) bool {
	if f.expectedFamilyID == 0 {
		return true
	}
	return f.expectedFamilyID == familyID
}

func (f *PickDestGVGFilter) CheckGVG(gvgMeta *vgmgr.GlobalVirtualGroupMeta) bool {
	if len(f.expectedSecondarySPIDs) == len(gvgMeta.SecondarySPIDs) {
		if gvgMeta.UsedStorageSize+2*f.expectedMinFreeSize > gvgMeta.StakingStorageSize {
			return false
		}
		for index, expectedSPID := range f.expectedSecondarySPIDs {
			if expectedSPID != gvgMeta.SecondarySPIDs[index] {
				return false
			}
		}
		return true
	}
	return false
}

// BucketMigrateExecutePlan is used to manage bucket migrate process.
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

// UpdateMigrateGVGLastMigratedObjectID persistent user updates and periodic progress reporting by Executor
func (plan *BucketMigrateExecutePlan) UpdateMigrateGVGLastMigratedObjectID(migrateKey string, lastMigratedObjectID uint64) error {
	err := plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitLastMigrateObjectID(migrateKey, lastMigratedObjectID)
	if err != nil {
		log.Errorw("failed to update migrate gvg progress", "migrate_key", migrateKey, "error", err)
		return err
	}
	return nil
}

func (plan *BucketMigrateExecutePlan) updateMigrateGVGStatus(migrateKey string, migrateExecuteUnit *GlobalVirtualGroupMigrateExecuteUnitByBucket, migrateStatus MigrateStatus) error {
	var (
		err   error
		vgfID uint32
	)

	plan.finished++
	migrateExecuteUnit.migrateStatus = migrateStatus

	// update migrate gvg status
	err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateKey, int(migrateStatus))
	if err != nil {
		log.Errorw("update migrate gvg status", "migrate_key", migrateKey, "error", err)
		return err
	}

	// all migrate units success, send tx to chain
	if plan.finished == len(plan.gvgUnitMap) {
		var bucket *types.Bucket
		bucket, err = plan.manager.baseApp.GfSpClient().GetBucketByBucketID(context.Background(), int64(plan.bucketID), true)
		if err != nil {
			return err
		}
		var gvgMappings []*storage_types.GVGMapping
		for _, migrateGVGUnit := range plan.gvgUnitMap {
			aggBlsSig, getBlsError := plan.getBlsAggregateSigForBucketMigration(context.Background(), migrateExecuteUnit)
			if getBlsError != nil {
				log.Errorw("failed to get bls aggregate signature", "error", getBlsError)
				return err
			}
			vgfID = migrateGVGUnit.destGVG.GetFamilyId()
			gvgMappings = append(gvgMappings, &storage_types.GVGMapping{SrcGlobalVirtualGroupId: migrateGVGUnit.gvg.GetId(),
				DstGlobalVirtualGroupId: migrateGVGUnit.destGVGID, SecondarySpBlsSignature: aggBlsSig})
		}

		migrateBucket := &storage_types.MsgCompleteMigrateBucket{Operator: plan.manager.baseApp.OperatorAddress(),
			BucketName: bucket.BucketInfo.GetBucketName(), GvgMappings: gvgMappings, GlobalVirtualGroupFamilyId: vgfID}
		txHash, txErr := plan.manager.baseApp.GfSpClient().CompleteMigrateBucket(context.Background(), migrateBucket)
		log.Infow("send complete migrate bucket msg to chain", "msg", migrateBucket, "tx_hash", txHash, "error", txErr)
	}
	return nil
}

// getBlsAggregateSigForBucketMigration get bls sign from secondary sp which is used for bucket migration
func (plan *BucketMigrateExecutePlan) getBlsAggregateSigForBucketMigration(ctx context.Context, migrateExecuteUnit *GlobalVirtualGroupMigrateExecuteUnitByBucket) ([]byte, error) {
	signDoc := storage_types.NewSecondarySpMigrationBucketSignDoc(plan.manager.baseApp.ChainID(),
		sdkmath.NewUint(plan.bucketID), migrateExecuteUnit.destSP.GetId(), migrateExecuteUnit.gvg.GetId(), migrateExecuteUnit.destGVGID)
	secondarySigs := make([][]byte, 0)
	for _, spID := range migrateExecuteUnit.gvg.GetSecondarySpIds() {
		spInfo, err := plan.manager.virtualGroupManager.QuerySPByID(spID)
		if err != nil {
			log.CtxErrorw(ctx, "failed to query sp by id", "error", err)
			return nil, err
		}
		sig, err := plan.manager.baseApp.GfSpClient().GetSecondarySPMigrationBucketApproval(ctx, spInfo.GetEndpoint(), signDoc)
		if err != nil {
			log.Errorw("failed to get secondary sp migration bucket approval", "error", err)
			return nil, err
		}
		secondarySigs = append(secondarySigs, sig)
	}
	aggBlsSig, err := util.BlsAggregate(secondarySigs)
	if err != nil {
		log.Errorw("failed to aggregate secondary sp bls signatures", "error", err)
		return nil, err
	}
	return aggBlsSig, nil
}

func (plan *BucketMigrateExecutePlan) startSPSchedule() {
	// dispatch to task-dispatcher, TODO: if CompleteEvents terminate the scheduling
	for {
		select {
		case <-plan.stopSignal:
			return // Terminate the scheduling
		default:
			log.Debugw("BucketMigrateExecutePlan Start startSPSchedule", "gvgUnitMap", plan.gvgUnitMap)
			for _, migrateGVGUnit := range plan.gvgUnitMap {

				// Skipping units that have already been scheduled
				if migrateGVGUnit.migrateStatus != WaitForMigrate {
					continue
				}

				migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{}
				migrateGVGTask.InitMigrateGVGTask(plan.manager.baseApp.TaskPriority(migrateGVGTask),
					plan.bucketID, migrateGVGUnit.gvg, migrateGVGUnit.redundancyIndex,
					migrateGVGUnit.srcSP,
					// TODO if add add a new tasktimeout
					plan.manager.baseApp.TaskTimeout(migrateGVGTask, 0),
					plan.manager.baseApp.TaskMaxRetry(migrateGVGTask))
				migrateGVGTask.SetDestGvg(migrateGVGUnit.destGVG)
				err := plan.manager.migrateGVGQueue.Push(migrateGVGTask)
				if err != nil {
					log.Errorw("failed to push migrate gvg task to queue", "error", err)
					time.Sleep(5 * time.Second) // Sleep for 5 seconds before retrying
				}
				log.Debugw("BucketMigrateExecutePlan Start push queue success", "migrateGVGUnit", migrateGVGUnit, "migrateGVGTask", migrateGVGTask)

				// Update database: migrateStatus to migrating
				migrateGVGUnit.migrateStatus = Migrating

				// update migrateStatus
				err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateGVGUnit.Key(), int(migrateGVGUnit.migrateStatus))
				if err != nil {
					log.Errorw("update migrate gvg status", "gvg_unit", migrateGVGUnit, "error", err)
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
	log.Debugf("BucketMigrateExecutePlan Start success")
	go plan.startSPSchedule()
	return nil
}

// BucketMigrateScheduler subscribes bucket migrate events and produces a gvg migrate plan.
type BucketMigrateScheduler struct {
	manager                   *ManageModular
	selfSP                    *sptypes.StorageProvider
	lastSubscribedBlockHeight uint64                               // load from db
	executePlanIDMap          map[uint64]*BucketMigrateExecutePlan // bucketID -> BucketMigrateExecutePlan
	isExited                  bool
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
	s.isExited = false

	// plan load from db
	s.loadBucketMigrateExecutePlansFromDB()

	return nil
}

func (s *BucketMigrateScheduler) Start() error {
	go s.subscribeEvents()
	return nil
}

func (s *BucketMigrateScheduler) subscribeEvents() {
	// TODO replace to time.second
	subscribeBucketMigrateEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeBucketMigrateEventInterval) * 100 * time.Millisecond)
	for {
		<-subscribeBucketMigrateEventsTicker.C
		var (
			migrationBucketEvents []*types.ListMigrateBucketEvents
			err                   error
			executePlan           *BucketMigrateExecutePlan
		)

		// 1. subscribe migrate bucket events
		migrationBucketEvents, err = s.manager.baseApp.GfSpClient().ListMigrateBucketEvents(context.Background(), s.lastSubscribedBlockHeight+1, s.selfSP.GetId())
		if err != nil {
			log.Errorw("failed to list migrate bucket events", "error", err)
			continue
		}

		// 2. make plan, start plan
		for _, migrateBucketEvents := range migrationBucketEvents {
			// when receive chain CompleteMigrationBucket event
			if migrateBucketEvents.CompleteEvents != nil {
				executePlan, err = s.getExecutePlanByBucketID(migrateBucketEvents.CompleteEvents.BucketId.Uint64())
				// TODO check db should be migrated
				if err != nil {
					log.Errorw("bucket migrate schedule received EventCompleteMigrationBucket")
					continue
				}

				for _, unit := range executePlan.gvgUnitMap {
					if unit.migrateStatus != Migrated {
						log.Errorw("report task may error, unit should be migrated", "unit", unit)
					}
				}
				// TODO when receive CompleteMigrationBucket event, we should delete memory & db's status
				executePlan.stopSPSchedule()
				continue
			}
			if migrateBucketEvents.Events != nil {
				// TODO migrating, switch to db
				if s.executePlanIDMap[migrateBucketEvents.Events.BucketId.Uint64()] != nil {
					continue
				}
				// debug
				log.Debugw("BucketMigrateScheduler subscribeEvents  ", "migrationBucketEvents", migrateBucketEvents.Events, "lastSubscribedBlockHeight", s.lastSubscribedBlockHeight)

				if s.isExited {
					return
				}
				log.Debugf("parse migrateBucketEvents.Events, then produceBucketMigrateExecutePlan")
				executePlan, err = s.produceBucketMigrateExecutePlan(migrateBucketEvents.Events)
				if err != nil {
					log.Errorw("failed to produce bucket migrate execute plan", "error", err)
					continue
				}
				if err = executePlan.Start(); err != nil {
					log.Errorw("failed to start bucket migrate execute plan", "error", err)
					continue
				}
				s.executePlanIDMap[executePlan.bucketID] = executePlan
			}
		}

		// 3.update subscribe progress to db
		updateErr := s.manager.baseApp.GfSpDB().UpdateBucketMigrateSubscribeProgress(s.lastSubscribedBlockHeight + 1)
		if updateErr != nil {
			log.Errorw("failed to update sp exit progress", "error", updateErr)
			continue
		}

		s.lastSubscribedBlockHeight++
	}
}

// pickGlobalVirtualGroupForBucketMigrate is used to pick a suitable gvg for replicating object.
func (s *BucketMigrateScheduler) pickGlobalVirtualGroupForBucketMigrate(filter *PickDestGVGFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
	var (
		err error
		gvg *vgmgr.GlobalVirtualGroupMeta
	)

	if gvg, err = s.manager.virtualGroupManager.PickGlobalVirtualGroupForBucketMigrate(filter); err != nil {
		// create a new gvg, and retry pick.
		if err = s.createGlobalVirtualGroupForBucketMigrate(filter.expectedFamilyID, filter.expectedSecondarySPIDs, 3*filter.expectedMinFreeSize); err != nil {
			log.Errorw("failed to create global virtual group for bucket migrate", "vgf_id", filter.expectedFamilyID, "error", err)
			return gvg, err
		}
		s.manager.virtualGroupManager.ForceRefreshMeta()
		if gvg, err = s.manager.virtualGroupManager.PickGlobalVirtualGroupForBucketMigrate(filter); err != nil {
			log.Errorw("failed to pick gvg", "vgf_id", filter.expectedFamilyID, "error", err)
			return gvg, err
		}
		return gvg, nil
	}
	log.Debugw("succeed to pick gvg for bucket migrate", "gvg", gvg)
	return gvg, nil
}

func (s *BucketMigrateScheduler) createGlobalVirtualGroupForBucketMigrate(vgfID uint32, secondarySPIDs []uint32, stakingSize uint64) error {
	virtualGroupParams, err := s.manager.baseApp.Consensus().QueryVirtualGroupParams(context.Background())
	if err != nil {
		return err
	}
	return s.manager.baseApp.GfSpClient().CreateGlobalVirtualGroup(context.Background(), &gfspserver.GfSpCreateGlobalVirtualGroup{
		VirtualGroupFamilyId: vgfID,
		PrimarySpAddress:     s.manager.baseApp.OperatorAddress(), // it is useless
		SecondarySpIds:       secondarySPIDs,
		Deposit: &sdk.Coin{
			Denom: virtualGroupParams.GetDepositDenom(),
			// TODO this be zero, how to fix ?
			Amount: sdkmath.NewInt(1000000),
		},
	})
}

func (s *BucketMigrateScheduler) produceBucketMigrateExecutePlan(event *storage_types.EventMigrationBucket) (*BucketMigrateExecutePlan, error) {
	var (
		primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup
		plan             *BucketMigrateExecutePlan
		err              error
		destFamilyID     uint32
		destGVG          *vgmgr.GlobalVirtualGroupMeta
		bucketInfo       *storage_types.BucketInfo
	)

	plan = newBucketMigrateExecutePlan(s.manager, event.BucketId.Uint64())

	log.Debugf("produceBucketMigrateExecutePlan bucketID", plan.bucketID)
	// query metadata service to get primary sp's gvg list.
	primarySPGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), plan.bucketID)
	if err != nil {
		log.Errorw("failed to list gvg ", "error", err)
		return nil, errors.New("failed to list gvg")
	}

	bucketInfo, err = s.manager.baseApp.Consensus().QueryBucketInfo(context.Background(), event.BucketName)
	if err != nil {
		return nil, err
	}
	srcSP, err := s.manager.virtualGroupManager.QuerySPByID(bucketInfo.GetPrimarySpId())
	if err != nil {
		log.Errorw("failed to query sp", "error", err)
		return nil, err
	}
	destSP, err := s.manager.virtualGroupManager.QuerySPByID(event.DstPrimarySpId)
	if err != nil {
		log.Errorw("failed to query sp", "error", err)
		return nil, err
	}

	destFamilyID = 0
	log.Debugw("produceBucketMigrateExecutePlan list", "primarySPGVGList", primarySPGVGList, "len:", len(primarySPGVGList))
	for _, srcGVG := range primarySPGVGList {
		secondarySPIDs := srcGVG.GetSecondarySpIds()
		// check conflicts.
		conflictedIndex, errNotInSecondarySPs := util.GetSecondarySPIndexFromGVG(srcGVG, destSP.GetId())
		log.Debugw("produceBucketMigrateExecutePlan prepare to check conflicts", "srcGVG", srcGVG, "destSP", destSP, "conflictedIndex", conflictedIndex, "errNotInSecondarySPs", errNotInSecondarySPs)
		if errNotInSecondarySPs == nil {
			// gvg has conflicts.
			excludedSPIDs := srcGVG.GetSecondarySpIds()
			excludedSPIDs = append(excludedSPIDs, srcSP.GetId())
			replacedSP, pickErr := s.manager.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(excludedSPIDs))
			if pickErr != nil {
				log.Errorw("failed to pick new sp to replace conflict secondary sp", "srcGVG", srcGVG, "destSP", destSP, "excludedSPIDs", excludedSPIDs, "error", pickErr)
				return nil, pickErr
			}
			secondarySPIDs[conflictedIndex] = replacedSP.GetId()
			log.Debugw("produceBucketMigrateExecutePlan resolve conflict", "excludedSPIDs", excludedSPIDs)
		}
		log.Debugw("produceBucketMigrateExecutePlan prepare to pick new gvg", "secondarySPIDs", secondarySPIDs)
		destGVG, err = s.pickGlobalVirtualGroupForBucketMigrate(NewPickDestGVGFilter(destFamilyID, secondarySPIDs, srcGVG.StoredSize))
		if err != nil {
			log.Errorw("failed to pick gvg for migrate bucket", "error", err)
			return nil, err
		}
		destFamilyID = destGVG.FamilyID
		destGlobalVirtualGroup := &virtualgrouptypes.GlobalVirtualGroup{Id: destGVG.ID, FamilyId: destGVG.FamilyID, PrimarySpId: destGVG.PrimarySPID, SecondarySpIds: destGVG.SecondarySPIDs, StoredSize: destGVG.StakingStorageSize}

		bucketUnit := newGlobalVirtualGroupMigrateExecuteUnitByBucket(plan.bucketID, srcGVG, srcSP, destSP, WaitForMigrate, destGVG.ID, 0, destGlobalVirtualGroup)
		plan.gvgUnitMap[srcGVG.GetId()] = bucketUnit
		log.Debugw("generate a new ", "MigrateExecuteUnitByBucket", bucketUnit)
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
		return fmt.Errorf("bucket execute plan is not found")
	}
	gvgID := task.GetSrcGvg().GetId()

	migrateExecuteUnit, ok := executePlan.gvgUnitMap[gvgID]
	if !ok {
		return fmt.Errorf("gvg unit is not found")
	}
	migrateKey := MakeBucketMigrateKey(migrateExecuteUnit.bucketID, migrateExecuteUnit.gvg.GetId())

	if task.GetFinished() {
		migrateExecuteUnit.migrateStatus = Migrated
		err = executePlan.updateMigrateGVGStatus(migrateKey, migrateExecuteUnit, Migrated)
		if err != nil {
			log.Errorw("failed to update migrate gvg status", "migrate_key", migrateKey, "error", err)
			return err
		}
	} else {
		migrateExecuteUnit.lastMigratedObjectID = task.GetLastMigratedObjectID()
		err = executePlan.UpdateMigrateGVGLastMigratedObjectID(migrateKey, task.GetLastMigratedObjectID())
		if err != nil {
			log.Errorw("failed to update migrate gvg last migrate object id", "migrate_key", migrateKey, "error", err)
			return err
		}
	}
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
			if queryErr != nil {
				log.Errorw("failed to query sp", "error", queryErr)
				return queryErr
			}
			destSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(migrateGVG.DestSPID)
			if queryErr != nil {
				log.Errorw("failed to query sp", "error", queryErr)
				return queryErr
			}
			primarySPGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), uint64(s.selfSP.GetId()))
			if err != nil {
				log.Errorw("failed to list gvg", "error", err)
				return errors.New("failed to list gvg")
			}
			for _, gvg := range primarySPGVGList {
				// TODO destGVG
				bucketUnit := newGlobalVirtualGroupMigrateExecuteUnitByBucket(bucketID, gvg, srcSP, destSP, WaitForMigrate, migrateGVG.DestSPID, migrateGVG.LastMigratedObjectID, nil)
				executePlan.gvgUnitMap[gvg.Id] = bucketUnit
			}
		}

		s.executePlanIDMap[executePlan.bucketID] = executePlan
	}
	return err
}
