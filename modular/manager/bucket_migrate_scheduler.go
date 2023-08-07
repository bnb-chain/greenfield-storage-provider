package manager

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspvgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	bucketCacheSize = int(100)
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

// GVGMetaCheck verifies whether expectedSPGVGList completely matches with the migrateGVGUnitMeta loaded from the database.
func GVGMetaCheck(expectedSPGVGList []*virtualgrouptypes.GlobalVirtualGroup, migrateGVGUnitMeta []*spdb.MigrateGVGUnitMeta) bool {
	if len(expectedSPGVGList) == len(migrateGVGUnitMeta) {
		chainGvgMaps := make(map[uint32]*virtualgrouptypes.GlobalVirtualGroup)
		existMaps := make(map[uint32]bool)
		for _, gvg := range expectedSPGVGList {
			chainGvgMaps[gvg.GetId()] = gvg
		}

		for _, migrateGVG := range migrateGVGUnitMeta {
			srcGvg, ok := chainGvgMaps[migrateGVG.GlobalVirtualGroupID]
			if !ok {
				return false
			}
			existMaps[srcGvg.GetId()] = true
		}

		for _, exist := range existMaps {
			if !exist {
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
	scheduler  *BucketMigrateScheduler
	bucketID   uint64
	gvgUnitMap map[uint32]*BucketMigrateGVGExecuteUnit // gvgID -> BucketMigrateGVGExecuteUnit
	stopSignal chan struct{}                           // stop schedule
	finished   int                                     // used for count the number of successful migrate units
}

func newBucketMigrateExecutePlan(manager *ManageModular, bucketID uint64) *BucketMigrateExecutePlan {
	executePlan := &BucketMigrateExecutePlan{
		manager:    manager,
		bucketID:   bucketID,
		gvgUnitMap: make(map[uint32]*BucketMigrateGVGExecuteUnit),
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
			MigrateGVGKey:            migrateGVGUnit.Key(),
			GlobalVirtualGroupID:     migrateGVGUnit.srcGVG.GetId(),
			DestGlobalVirtualGroupID: migrateGVGUnit.destGVGID,
			VirtualGroupFamilyID:     migrateGVGUnit.srcGVG.GetFamilyId(),
			RedundancyIndex:          piecestore.PrimarySPRedundancyIndex,
			BucketID:                 migrateGVGUnit.bucketID,
			SrcSPID:                  migrateGVGUnit.srcSP.GetId(),
			DestSPID:                 migrateGVGUnit.destSP.GetId(),
			LastMigratedObjectID:     migrateGVGUnit.lastMigratedObjectID,
			MigrateStatus:            int(migrateGVGUnit.migrateStatus),
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

// send CompleteMigrateBucket to chain 1) empty bucket: gvgUnitMap is nil; 2) normal bucket
func (plan *BucketMigrateExecutePlan) sendCompleteMigrateBucketTx(migrateExecuteUnit *BucketMigrateGVGExecuteUnit) error {
	var (
		vgfID uint32
		err   error
	)
	// empty bucket, need to pick a vgf
	if migrateExecuteUnit == nil {
		vgfID, err = plan.manager.PickVirtualGroupFamily(context.Background(), &task.NullTask{})
		if err != nil {
			log.Errorw("failed to pick vgf for migrate bucket", "error", err)
			return err
		}
	}

	bucket, err := plan.manager.baseApp.GfSpClient().GetBucketByBucketID(context.Background(), int64(plan.bucketID), true)
	if err != nil {
		return err
	}
	var gvgMappings []*storagetypes.GVGMapping
	for _, migrateGVGUnit := range plan.gvgUnitMap {
		aggBlsSig, getBlsError := plan.getBlsAggregateSigForBucketMigration(context.Background(), migrateExecuteUnit)
		if getBlsError != nil {
			log.Errorw("failed to get bls aggregate signature", "error", getBlsError)
			return err
		}
		vgfID = migrateGVGUnit.destGVG.GetFamilyId()
		gvgMappings = append(gvgMappings, &storagetypes.GVGMapping{SrcGlobalVirtualGroupId: migrateGVGUnit.srcGVG.GetId(),
			DstGlobalVirtualGroupId: migrateGVGUnit.destGVGID, SecondarySpBlsSignature: aggBlsSig})
	}

	migrateBucket := &storagetypes.MsgCompleteMigrateBucket{Operator: plan.manager.baseApp.OperatorAddress(),
		BucketName: bucket.BucketInfo.GetBucketName(), GvgMappings: gvgMappings, GlobalVirtualGroupFamilyId: vgfID}
	txHash, txErr := plan.manager.baseApp.GfSpClient().CompleteMigrateBucket(context.Background(), migrateBucket)
	log.Infow("send complete migrate bucket msg to chain", "msg", migrateBucket, "tx_hash", txHash, "error", txErr)
	return nil
}

func (plan *BucketMigrateExecutePlan) updateMigrateGVGStatus(migrateKey string, migrateExecuteUnit *BucketMigrateGVGExecuteUnit, migrateStatus MigrateStatus) error {

	plan.finished++
	migrateExecuteUnit.migrateStatus = migrateStatus

	// update migrate gvg status
	err := plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateKey, int(migrateStatus))
	if err != nil {
		log.Errorw("update migrate gvg status", "migrate_key", migrateKey, "error", err)
		return err
	}

	// all migrate units success, send tx to chain
	if plan.finished == len(plan.gvgUnitMap) {
		err := plan.sendCompleteMigrateBucketTx(migrateExecuteUnit)
		if err != nil {
			log.Errorw("failed to send complete migrate bucket msg to chain", "error", err, "migrateExecuteUnit", migrateExecuteUnit)
			return err
		}
		err = plan.scheduler.doneMigrateBucket(migrateExecuteUnit.bucketID)
		if err != nil {
			log.Errorw("failed to done migrate bucket", "error", err, "bucket_id", migrateExecuteUnit.bucketID)
			return err
		}
	}
	return nil
}

// getBlsAggregateSigForBucketMigration get bls sign from secondary sp which is used for bucket migration
func (plan *BucketMigrateExecutePlan) getBlsAggregateSigForBucketMigration(ctx context.Context, migrateExecuteUnit *BucketMigrateGVGExecuteUnit) ([]byte, error) {
	signDoc := storagetypes.NewSecondarySpMigrationBucketSignDoc(plan.manager.baseApp.ChainID(),
		sdkmath.NewUint(plan.bucketID), migrateExecuteUnit.destSP.GetId(), migrateExecuteUnit.srcGVG.GetId(), migrateExecuteUnit.destGVGID)
	secondarySigs := make([][]byte, 0)
	for _, spID := range migrateExecuteUnit.srcGVG.GetSecondarySpIds() {
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
	// dispatch to task-dispatcher
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
					plan.bucketID, migrateGVGUnit.srcGVG, -1,
					migrateGVGUnit.srcSP,
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
	bucketCache               *BucketCache
	// src sp swap unit plan.
	mutex sync.RWMutex
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

	s.bucketCache = NewBucketCache(bucketCacheSize, 30*time.Minute)
	if err != nil {
		return err
	}

	// plan load from db
	s.loadBucketMigrateExecutePlansFromDB()

	return nil
}

func (s *BucketMigrateScheduler) Start() error {
	go s.subscribeEvents()
	return nil
}

// Before processing MigrateBucketEvents, first check if the status of the bucket on the chain meets the expectations. If it meets the expectations, proceed with the execution; otherwise, skip this MigrateBucketEvent event.
func (s *BucketMigrateScheduler) checkBucketFromChain(bucketID uint64, expectedStatus storagetypes.BucketStatus) (expected bool, err error) {
	// check the chain's bucket is migrating
	key := cacheKey(bucketID)
	var bucketInfo *storagetypes.BucketInfo
	QueryBucketInfoFromChainFunc := func() error {
		bucketInfo, err = s.manager.baseApp.Consensus().QueryBucketInfoById(context.Background(), bucketID)
		if err != nil {
			return err
		}
		s.bucketCache.Set(key, bucketInfo)
		return nil
	}

	elem, has := s.bucketCache.Get(key)
	if has {
		value, ok := elem.(storagetypes.BucketInfo)
		if !ok {
			log.Debugw("failed to get bucketinfo from bucket cache", "key", key)
			s.bucketCache.Delete(key)
			err = QueryBucketInfoFromChainFunc()
			if err != nil {
				return false, err
			}
		} else {
			bucketInfo = &value
		}
	} else {
		err = QueryBucketInfoFromChainFunc()
		if err != nil {
			return false, err
		}
	}

	if bucketInfo.BucketStatus != expectedStatus {
		log.Debugw("the bucket status is not same, the event will skip", "bucketInfo", bucketInfo, "expectedStatus", expectedStatus)
		return false, nil
	}
	return true, nil
}

func (s *BucketMigrateScheduler) doneMigrateBucket(bucketID uint64) error {
	expected, err := s.checkBucketFromChain(bucketID, storagetypes.BUCKET_STATUS_CREATED)
	if err != nil {
		return err
	}
	if !expected {
		return nil
	}
	executePlan, err := s.getExecutePlanByBucketID(bucketID)
	// 1) Received the CompleteEvents event for the first time.
	// 2) Subsequently received the CompleteEvents event.
	if err != nil {
		log.Errorw("bucket migrate schedule received EventCompleteMigrationBucket, the event may already finished", "bucket_id", bucketID)
		return err
	}

	for _, unit := range executePlan.gvgUnitMap {
		if unit.migrateStatus != Migrated {
			log.Errorw("report task may error, unit should be migrated", "unit", unit)
		}
	}
	s.deleteExecutePlanByBucketID(bucketID)
	executePlan.stopSPSchedule()
	return nil
}

func (s *BucketMigrateScheduler) processEvents(migrateBucketEvents *types.ListMigrateBucketEvents) error {
	// 1. process CancelEvents
	if migrateBucketEvents.CancelEvents != nil {
		expected, err := s.checkBucketFromChain(migrateBucketEvents.CancelEvents.BucketId.Uint64(), storagetypes.BUCKET_STATUS_CREATED)
		if err != nil {
			return err
		}
		if !expected {
			return nil
		}
		executePlan, err := s.getExecutePlanByBucketID(migrateBucketEvents.CancelEvents.BucketId.Uint64())
		if err != nil {
			log.Errorw("bucket migrate schedule received EventCompleteMigrationBucket", "CancelEvents", migrateBucketEvents.CancelEvents)
			return err
		}
		s.deleteExecutePlanByBucketID(migrateBucketEvents.CancelEvents.BucketId.Uint64())
		executePlan.stopSPSchedule()
	}
	// 2. process CompleteEvents
	if migrateBucketEvents.CompleteEvents != nil {
		return nil
	}
	// 3. process Events
	if migrateBucketEvents.Events != nil {
		expected, err := s.checkBucketFromChain(migrateBucketEvents.Events.BucketId.Uint64(), storagetypes.BUCKET_STATUS_MIGRATING)
		if err != nil {
			return err
		}
		if !expected {
			return nil
		}
		if s.executePlanIDMap[migrateBucketEvents.Events.BucketId.Uint64()] != nil {
			log.Debugw("bucket migrate scheduler the bucket is already in migrating", "Events", migrateBucketEvents.Events)
			return errors.New("bucket already in migrating")
		}
		log.Debugw("Bucket Migrate Scheduler process Events", "migrationBucketEvents", migrateBucketEvents.Events, "lastSubscribedBlockHeight", s.lastSubscribedBlockHeight)

		executePlan, err := s.produceBucketMigrateExecutePlan(migrateBucketEvents.Events)
		if err != nil {
			log.Errorw("failed to produce bucket migrate execute plan", "Events", migrateBucketEvents.Events, "error", err)
			return err
		}
		if err = executePlan.Start(); err != nil {
			log.Errorw("failed to start bucket migrate execute plan", "Events", migrateBucketEvents.Events, "executePlan", executePlan, "error", err)
			return err
		}
		s.executePlanIDMap[executePlan.bucketID] = executePlan
	}
	return nil
}

func (s *BucketMigrateScheduler) subscribeEvents() {
	go func() {
		UpdateBucketMigrateSubscribeProgressFunc := func() {
			updateErr := s.manager.baseApp.GfSpDB().UpdateBucketMigrateSubscribeProgress(s.lastSubscribedBlockHeight + 1)
			if updateErr != nil {
				log.Errorw("failed to update bucket migrate progress", "error", updateErr)
			}
			s.lastSubscribedBlockHeight++
			log.Infow("bucket migrate subscribe progress", "last_subscribed_block_height", s.lastSubscribedBlockHeight)
		}
		subscribeBucketMigrateEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeBucketMigrateEventInterval) * time.Millisecond)
		defer subscribeBucketMigrateEventsTicker.Stop()

		logNumber := 0
		for range subscribeBucketMigrateEventsTicker.C {
			// 1. subscribe migrate bucket events
			migrationBucketEvents, subscribeError := s.manager.baseApp.GfSpClient().ListMigrateBucketEvents(context.Background(), s.lastSubscribedBlockHeight+1, s.selfSP.GetId())
			if subscribeError != nil {
				logNumber++
				if (logNumber % printLogPerN) == 0 {
					log.Errorw("failed to list migrate bucket events", "block_id", s.lastSubscribedBlockHeight+1,
						"error", subscribeError)
				}
				continue
			}
			log.Infow("loop subscribe bucket migrate event", "migrationBucketEvents", migrationBucketEvents, "block_id", s.lastSubscribedBlockHeight+1, "sp_address", s.manager.baseApp.OperatorAddress())

			// 2. make plan, start plan
			for _, migrateBucketEvents := range migrationBucketEvents {
				err := s.processEvents(migrateBucketEvents)
				if err != nil {
					log.Errorw("bucket migrate process error", "migrateBucketEvents", migrateBucketEvents, "error", err)
				}
			}

			// 3.update subscribe progress to db
			UpdateBucketMigrateSubscribeProgressFunc()
		}
	}()
}

// pickGlobalVirtualGroupForBucketMigrate is used to pick a suitable gvg for replicating object.
func (s *BucketMigrateScheduler) pickGlobalVirtualGroupForBucketMigrate(filter *PickDestGVGFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
	var (
		err error
		gvg *vgmgr.GlobalVirtualGroupMeta
	)

	// TODO: The logic of GVGPickFilter is modified to ignore the StakingStorageSize when checking for a valid GVG.
	// If a GVG is considered suitable but its StakingStorageSize is insufficient, it will directly send a request to the blockchain
	// to add additional staking funds.
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

// Calculate the staking size strategy for the target GVG
func calculateStakingSizeStrategy(manager *ManageModular) (denom string, amount sdkmath.Int, err error) {
	var (
		params             *storagetypes.Params
		stakingStorageSize uint64
	)

	if params, err = manager.baseApp.Consensus().QueryStorageParamsByTimestamp(context.Background(), time.Now().Unix()); err != nil {
		return "", sdkmath.ZeroInt(), err
	}

	gvgMeta, err := manager.virtualGroupManager.GenerateGlobalVirtualGroupMeta(NewGenerateGVGSecondarySPsPolicyByPrefer(params, manager.gvgPreferSPList))
	if err != nil {
		return "", sdkmath.ZeroInt(), err
	}

	virtualGroupParams, err := manager.baseApp.Consensus().QueryVirtualGroupParams(context.Background())
	if err != nil {
		return "", sdkmath.ZeroInt(), err
	}
	// double check
	if gvgMeta.StakingStorageSize == 0 {
		stakingStorageSize = gfspvgmgr.DefaultInitialGVGStakingStorageSize
	} else {
		stakingStorageSize = gvgMeta.StakingStorageSize
	}
	amount = virtualGroupParams.GvgStakingPerBytes.Mul(math.NewIntFromUint64(stakingStorageSize))
	log.Infow("begin to create a gvg for bucket migrate", "gvg_meta", gvgMeta, "amount", amount)

	return virtualGroupParams.DepositDenom, amount, nil
}

func (s *BucketMigrateScheduler) createGlobalVirtualGroupForBucketMigrate(vgfID uint32, secondarySPIDs []uint32, stakingSize uint64) error {
	denom, amount, err := calculateStakingSizeStrategy(s.manager)
	if err != nil {
		return err
	}
	return s.manager.baseApp.GfSpClient().CreateGlobalVirtualGroup(context.Background(), &gfspserver.GfSpCreateGlobalVirtualGroup{
		VirtualGroupFamilyId: vgfID,
		PrimarySpAddress:     s.manager.baseApp.OperatorAddress(), // it is useless
		SecondarySpIds:       secondarySPIDs,
		Deposit: &sdk.Coin{
			Denom:  denom,
			Amount: amount,
		},
	})
}

// replace SecondarySp which is in STATUS_GRACEFUL_EXITING
func (s *BucketMigrateScheduler) replaceExitingSP(secondarySPIDs []uint32) ([]uint32, error) {
	replacedSPIDs := secondarySPIDs
	excludedSPIDs := secondarySPIDs

	for idx, spID := range secondarySPIDs {
		sp, err := s.manager.virtualGroupManager.QuerySPByID(spID)
		if err != nil {
			log.Errorw("failed to query sp", "error", err)
			return nil, err
		}
		if sp.Status == sptypes.STATUS_GRACEFUL_EXITING {
			replacedSP, pickErr := s.manager.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(excludedSPIDs))
			if pickErr != nil {
				log.Errorw("failed to pick new sp to replace exiting secondary sp", "excludedSPIDs", excludedSPIDs, "error", pickErr)
				return nil, pickErr
			}
			replacedSPIDs[idx] = replacedSP.GetId()
			excludedSPIDs = append(excludedSPIDs, replacedSP.GetId())
		}
	}

	return replacedSPIDs, nil
}

func (s *BucketMigrateScheduler) generateBucketMigrateGVGExecuteUnitFromDB(primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup, bucketID uint64, executePlan *BucketMigrateExecutePlan, event *storagetypes.EventMigrationBucket) (*BucketMigrateExecutePlan, error) {
	// check every migrateGVGUnitMeta must match primarySPGVGList

	// 1) check chain's bucket status
	expected, err := s.checkBucketFromChain(bucketID, storagetypes.BUCKET_STATUS_MIGRATING)
	if err != nil {
		return nil, err
	}
	if !expected {
		return nil, nil
	}

	chainGvgMaps := make(map[uint32]*virtualgrouptypes.GlobalVirtualGroup)
	for _, gvg := range primarySPGVGList {
		chainGvgMaps[gvg.GetId()] = gvg
	}

	migrateGVGUnitMeta, err := s.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByBucketID(bucketID)
	if err != nil {
		return nil, err
	}

	// 2) not match, generate migrate gvg again
	if !GVGMetaCheck(primarySPGVGList, migrateGVGUnitMeta) {
		srcSP, destSP, err := s.getSrcSPAndDestSPFromMigrateEvent(event)
		if err != nil {
			return nil, err
		}
		executePlan, err = s.generateBucketMigrateGVGExecuteUnit(primarySPGVGList, event, executePlan, srcSP, destSP)
		if err != nil {
			return nil, err
		}
	} else {
		// Using migrateGVGUnitMeta to construct PrimaryGVGIDMapMigrateUnits and execute them one by one.
		for _, migrateGVG := range migrateGVGUnitMeta {
			srcSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(migrateGVG.SrcSPID)
			if queryErr != nil {
				log.Errorw("failed to query sp", "error", queryErr, "sp_info", srcSP)
				return nil, queryErr
			}
			destSP, queryErr := s.manager.virtualGroupManager.QuerySPByID(migrateGVG.DestSPID)
			if queryErr != nil {
				log.Errorw("failed to query sp", "error", queryErr, "sp_info", destSP)
				return nil, queryErr
			}

			srcGvg, ok := chainGvgMaps[migrateGVG.GlobalVirtualGroupID]
			if ok {
				destGvg, err := s.manager.baseApp.Consensus().QueryGlobalVirtualGroup(context.Background(), migrateGVG.DestGlobalVirtualGroupID)
				if err != nil {
					return nil, err
				}
				bucketUnit := newBucketMigrateGVGExecuteUnit(bucketID, srcGvg, srcSP, destSP, MigrateStatus(migrateGVG.MigrateStatus), migrateGVG.DestSPID, migrateGVG.LastMigratedObjectID, destGvg)
				executePlan.gvgUnitMap[srcGvg.GetId()] = bucketUnit
			} else {
				// gvgChecker has verified and this scenario should not occur.
				log.Debugw("failed to get src gvg", "gvg_id", migrateGVG.GlobalVirtualGroupID)
			}
		}
	}

	return executePlan, nil
}

func (s *BucketMigrateScheduler) generateBucketMigrateGVGExecuteUnit(primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup, event *storagetypes.EventMigrationBucket, plan *BucketMigrateExecutePlan, srcSP, destSP *sptypes.StorageProvider) (*BucketMigrateExecutePlan, error) {
	var (
		destGVG      *vgmgr.GlobalVirtualGroupMeta
		destFamilyID uint32
	)
	for _, srcGVG := range primarySPGVGList {
		srcSecondarySPIDs := srcGVG.GetSecondarySpIds()
		// check sp exiting
		secondarySPIDs, err := s.replaceExitingSP(srcSecondarySPIDs)
		if err != nil {
			log.Errorw("pick sp to replace exiting sp error", "srcSecondarySPIDs", srcSecondarySPIDs, "secondarySPIDs", secondarySPIDs, "EventMigrationBucket", event)
			return nil, err
		}

		// check conflicts.
		conflictedIndex, errNotInSecondarySPs := util.GetSecondarySPIndexFromGVG(srcGVG, destSP.GetId())
		log.Debugw("produceBucketMigrateExecutePlan prepare to check conflicts", "srcGVG", srcGVG, "destSP", destSP, "conflictedIndex", conflictedIndex, "errNotInSecondarySPs", errNotInSecondarySPs, "EventMigrationBucket", event)
		if errNotInSecondarySPs == nil {
			// gvg has conflicts.
			excludedSPIDs := srcGVG.GetSecondarySpIds()
			replacedSP, pickErr := s.manager.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(excludedSPIDs))
			if pickErr != nil {
				log.Errorw("failed to pick new sp to replace conflict secondary sp", "srcGVG", srcGVG, "destSP", destSP, "excludedSPIDs", excludedSPIDs, "error", pickErr, "EventMigrationBucket", event)
				return nil, pickErr
			}
			secondarySPIDs[conflictedIndex] = replacedSP.GetId()
			log.Debugw("produceBucketMigrateExecutePlan resolve conflict", "excludedSPIDs", excludedSPIDs, "EventMigrationBucket", event)
		}
		log.Debugw("produceBucketMigrateExecutePlan prepare to pick new gvg", "secondarySPIDs", secondarySPIDs, "EventMigrationBucket", event)
		destGVG, err = s.pickGlobalVirtualGroupForBucketMigrate(NewPickDestGVGFilter(destFamilyID, secondarySPIDs, srcGVG.StoredSize))
		if err != nil {
			log.Errorw("failed to pick gvg for migrate bucket", "error", err, "EventMigrationBucket", event)
			return nil, err
		}
		destFamilyID = destGVG.FamilyID
		destGlobalVirtualGroup := &virtualgrouptypes.GlobalVirtualGroup{Id: destGVG.ID, FamilyId: destGVG.FamilyID, PrimarySpId: destGVG.PrimarySPID,
			SecondarySpIds: destGVG.SecondarySPIDs, StoredSize: destGVG.StakingStorageSize}

		bucketUnit := newBucketMigrateGVGExecuteUnit(plan.bucketID, srcGVG, srcSP, destSP, WaitForMigrate, destGVG.ID, 0, destGlobalVirtualGroup)
		plan.gvgUnitMap[srcGVG.GetId()] = bucketUnit
		log.Infow("succeeded to generate a new bucket migrate uint", "migrate_execute_unit", bucketUnit, "EventMigrationBucket", event)
	}
	log.Infow("succeeded to generate a new bucket migrate execute plan", "migrate_bucket_plan", plan, "EventMigrationBucket", event)

	return plan, nil
}

func (s *BucketMigrateScheduler) getSrcSPAndDestSPFromMigrateEvent(event *storagetypes.EventMigrationBucket) (srcSP, destSP *sptypes.StorageProvider, err error) {
	bucketInfo, err := s.manager.baseApp.Consensus().QueryBucketInfoById(context.Background(), event.BucketId.Uint64())
	if err != nil {
		return nil, nil, err
	}
	bucketSPID, err := util.GetBucketPrimarySPID(context.Background(), s.manager.baseApp.Consensus(), bucketInfo)
	if err != nil {
		return nil, nil, err
	}

	srcSP, err = s.manager.virtualGroupManager.QuerySPByID(bucketSPID)
	if err != nil {
		log.Errorw("failed to query sp", "error", err, "EventMigrationBucket", event)
		return nil, nil, err
	}
	destSP, err = s.manager.virtualGroupManager.QuerySPByID(event.DstPrimarySpId)
	if err != nil {
		log.Errorw("failed to query sp", "error", err, "EventMigrationBucket", event)
		return nil, nil, err
	}
	return srcSP, destSP, nil
}

func (s *BucketMigrateScheduler) produceBucketMigrateExecutePlan(event *storagetypes.EventMigrationBucket) (*BucketMigrateExecutePlan, error) {
	var (
		primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup
		plan             *BucketMigrateExecutePlan
		err              error
	)

	plan = newBucketMigrateExecutePlan(s.manager, event.BucketId.Uint64())

	log.Debugw("produceBucketMigrateExecutePlan", "bucketID", plan.bucketID, "EventMigrationBucket", event)
	// query metadata service to get primary sp's gvg list.
	primarySPGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), plan.bucketID)
	if err != nil {
		log.Errorw("failed to list gvg ", "error", err, "EventMigrationBucket", event)
		return nil, errors.New("failed to list gvg")
	}

	srcSP, destSP, err := s.getSrcSPAndDestSPFromMigrateEvent(event)
	if err != nil {
		return nil, err
	}

	log.Debugw("produceBucketMigrateExecutePlan list", "primarySPGVGList", primarySPGVGList, "bucket_gvg_list_len", len(primarySPGVGList), "EventMigrationBucket", event)
	if len(primarySPGVGList) == 0 {
		// an empty bucket ends here, and it will not return a plan. The execution will not continue.
		err := plan.sendCompleteMigrateBucketTx(nil)
		if err != nil {
			log.Errorw("failed to send complete migrate bucket msg to chain", "error", err, "EventMigrationBucket", event)
			return nil, err
		}
		err = s.doneMigrateBucket(event.BucketId.Uint64())
		if err != nil {
			log.Errorw("failed to done migrate bucket", "error", err, "EventMigrationBucket", event)
			return nil, err
		}
		return nil, nil
	} else {
		plan, err = s.generateBucketMigrateGVGExecuteUnit(primarySPGVGList, event, plan, srcSP, destSP)
		return plan, err
	}

}

func (s *BucketMigrateScheduler) getExecutePlanByBucketID(bucketID uint64) (*BucketMigrateExecutePlan, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	executePlan, ok := s.executePlanIDMap[bucketID]
	if ok {
		return executePlan, nil
	} else {
		return nil, errors.New("no such execute plan")
	}
}

func (s *BucketMigrateScheduler) deleteExecutePlanByBucketID(bucketID uint64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.executePlanIDMap[bucketID]
	if ok {
		delete(s.executePlanIDMap, bucketID)
	} else {
		log.Debugw("deleteExecutePlanByBucketID may error, delete an unexisted bucket", "bucketID", bucketID)
	}
	return nil
}

func (s *BucketMigrateScheduler) listExecutePlan() (*gfspserver.GfSpQueryBucketMigrateResponse, error) {
	var res gfspserver.GfSpQueryBucketMigrateResponse
	var plans []*gfspserver.GfSpBucketMigrate
	for _, executePlan := range s.executePlanIDMap {
		var plan gfspserver.GfSpBucketMigrate
		plan.BucketId = executePlan.bucketID
		plan.Finished = uint32(executePlan.finished)
		for _, unit := range executePlan.gvgUnitMap {
			plan.GvgTask = append(plan.GvgTask, &gfspserver.GfSpMigrateGVG{
				SrcGvgId:             unit.srcGVG.GetId(),
				DestGvgId:            unit.destGVG.GetId(),
				LastMigratedObjectId: unit.lastMigratedObjectID,
				Status:               int32(unit.migrateStatus),
			})
		}
		plans = append(plans, &plan)
	}
	res.BucketMigrate = plans
	res.SelfSpId = s.selfSP.GetId()
	log.Debugw("BucketMigrateScheduler listExecutePlan", "plans res", res)
	return &res, nil
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
	migrateKey := MakeBucketMigrateKey(migrateExecuteUnit.bucketID, migrateExecuteUnit.srcGVG.GetId())

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
		migrationBucketEvents []*types.ListMigrateBucketEvents
		err                   error
		primarySPGVGList      []*virtualgrouptypes.GlobalVirtualGroup
		migratingBucketIDs    []uint64
	)
	migrationBucketEventsMap := make(map[uint64]*types.ListMigrateBucketEvents)

	// get bucket id from metadata (migrate bucket events)
	migrationBucketEvents, err = s.manager.baseApp.GfSpClient().ListMigrateBucketEvents(context.Background(), s.lastSubscribedBlockHeight+1, s.selfSP.GetId())
	if err != nil {
		log.Errorw("failed to list migrate bucket events", "error", err)
		return errors.New("failed to list migrate bucket events")
	}

	for _, migrateBucketEvents := range migrationBucketEvents {
		// if has CompleteEvents & CancelEvents, skip it
		if migrateBucketEvents.CompleteEvents != nil || migrateBucketEvents.CancelEvents != nil {
			continue
		}
		if migrateBucketEvents.Events != nil {
			migratingBucketIDs = append(migratingBucketIDs, migrateBucketEvents.Events.BucketId.Uint64())
			migrationBucketEventsMap[migrateBucketEvents.Events.BucketId.Uint64()] = migrateBucketEvents
		}
	}
	log.Infow("load bucket migrate execute plans from DB", "bucketIDs", migratingBucketIDs)
	// load from db by BucketID & construct plan
	for _, bucketID := range migratingBucketIDs {
		executePlan := newBucketMigrateExecutePlan(s.manager, bucketID)
		if _, found := migrationBucketEventsMap[bucketID]; !found {
			return fmt.Errorf("migrate bucket events is not found, bucketID:%d", bucketID)
		}
		bucketMigrateEvent := migrationBucketEventsMap[bucketID]

		primarySPGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), bucketID)
		if err != nil {
			log.Errorw("failed to list gvg", "error", err, "EventMigrationBucket", bucketMigrateEvent.Events)
			return errors.New("failed to list gvg")
		}

		if len(primarySPGVGList) == 0 {
			// an empty bucket ends here, and it will not return a plan. The execution will not continue.
			err = executePlan.sendCompleteMigrateBucketTx(nil)
			if err != nil {
				log.Errorw("failed to send complete migrate bucket msg to chain", "error", err, "EventMigrationBucket", bucketMigrateEvent.Events)
				return err
			}
			err = s.doneMigrateBucket(bucketID)
			if err != nil {
				log.Errorw("failed to done migrate bucket", "error", err, "EventMigrationBucket", bucketMigrateEvent)
				return err
			}

		} else {
			executePlan, err = s.generateBucketMigrateGVGExecuteUnitFromDB(primarySPGVGList, bucketID, executePlan, bucketMigrateEvent.Events)
			if err != nil {
				log.Errorw("failed to produce bucket migrate execute plan", "Events", bucketMigrateEvent.Events, "error", err)
				return err
			}
		}

		log.Debugw("bucket migrate scheduler load from db", "executePlan", executePlan, "bucketID", bucketID)
		if err = executePlan.Start(); err != nil {
			log.Errorw("failed to start bucket migrate execute plan", "Events", bucketMigrateEvent.Events, "executePlan", executePlan, "error", err)
			return err
		}

		s.executePlanIDMap[executePlan.bucketID] = executePlan
	}
	log.Debugw("bucket migrate scheduler load from db success", "bucketIDs", migratingBucketIDs)

	return err
}

func cacheKey(bucketId uint64) string {
	return fmt.Sprintf("bucketid:%d", bucketId)
}

// BucketCache is an LRU cache.
type BucketCache struct {
	sync.RWMutex
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	maxEntries int

	lruIndex   *list.List
	ttlIndex   []*list.Element
	cache      map[string]*list.Element
	expiration time.Duration
}

type bucketEntry struct {
	key       string
	value     interface{}
	timestamp time.Time
}

// NewBucketCache creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewBucketCache(maxEntries int, expire time.Duration) *BucketCache {
	c := &BucketCache{
		maxEntries: maxEntries,
		expiration: expire,
		lruIndex:   list.New(),
		cache:      make(map[string]*list.Element),
	}
	if c.expiration > 0 {
		c.ttlIndex = make([]*list.Element, 0)
		go c.cleanExpired()
	}
	return c
}

// cleans expired entries performing minimal checks
func (c *BucketCache) cleanExpired() {
	for {
		c.RLock()
		if len(c.ttlIndex) == 0 {
			c.RUnlock()
			time.Sleep(c.expiration)
			continue
		}
		e := c.ttlIndex[0]

		en := e.Value.(*bucketEntry)
		exp := en.timestamp.Add(c.expiration)
		c.RUnlock()
		if time.Now().After(exp) {
			c.Lock()
			c.removeElement(e)
			c.Unlock()
		} else {
			time.Sleep(time.Since(exp))
		}
	}
}

// Set adds a value to the cache
func (c *BucketCache) Set(key string, value interface{}) {
	c.Lock()
	if c.cache == nil {
		c.cache = make(map[string]*list.Element)
		c.lruIndex = list.New()
		if c.expiration > 0 {
			c.ttlIndex = make([]*list.Element, 0)
		}
	}

	if e, ok := c.cache[key]; ok {
		c.lruIndex.MoveToFront(e)

		en := e.Value.(*bucketEntry)
		en.value = value
		en.timestamp = time.Now()

		c.Unlock()
		return
	}
	e := c.lruIndex.PushFront(&bucketEntry{key: key, value: value, timestamp: time.Now()})
	if c.expiration > 0 {
		c.ttlIndex = append(c.ttlIndex, e)
	}
	c.cache[key] = e

	if c.maxEntries != 0 && c.lruIndex.Len() > c.maxEntries {
		c.removeOldest()
	}
	c.Unlock()
}

// Get looks up a key's value from the cache.
func (c *BucketCache) Get(key string) (value interface{}, ok bool) {
	c.Lock()
	defer c.Unlock()
	if c.cache == nil {
		return
	}
	if e, hit := c.cache[key]; hit {
		c.lruIndex.MoveToFront(e)
		return e.Value.(*bucketEntry).value, true
	}
	return
}

// Delete removes the provided key from the cache.
func (c *BucketCache) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	if c.cache == nil {
		return
	}
	if e, hit := c.cache[key]; hit {
		c.removeElement(e)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *BucketCache) removeOldest() {
	if c.cache == nil {
		return
	}
	e := c.lruIndex.Back()
	if e != nil {
		c.removeElement(e)
	}
}

func (c *BucketCache) removeElement(e *list.Element) {
	c.lruIndex.Remove(e)
	if c.expiration > 0 {
		for i, se := range c.ttlIndex {
			if se == e {
				//delete
				copy(c.ttlIndex[i:], c.ttlIndex[i+1:])
				c.ttlIndex[len(c.ttlIndex)-1] = nil
				c.ttlIndex = c.ttlIndex[:len(c.ttlIndex)-1]
				break
			}
		}
	}
	if e.Value != nil {
		kv := e.Value.(*bucketEntry)
		delete(c.cache, kv.key)
	}
}

// Len returns the number of items in the cache.
func (c *BucketCache) Len() int {
	c.RLock()
	defer c.RUnlock()
	if c.cache == nil {
		return 0
	}
	return c.lruIndex.Len()
}

// Flush empties the whole cache
func (c *BucketCache) Flush() {
	c.Lock()
	defer c.Unlock()
	c.lruIndex = list.New()
	if c.expiration > 0 {
		c.ttlIndex = make([]*list.Element, 0)
	}
	c.cache = make(map[string]*list.Element)
}
