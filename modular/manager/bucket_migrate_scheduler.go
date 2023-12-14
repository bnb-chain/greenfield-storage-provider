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

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
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
)

const (
	bucketCacheSize   = int(100)
	bucketCacheExpire = 30 * time.Minute

	SigExpireTimeSecond    = 60 * 60
	migrateGVGTaskMaxRetry = int(5)
	blockInterval          = 4 * time.Second
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

// CheckGVGMetaConsistent verifies whether expectedSPGVGList completely matches with the migrateGVGUnitMeta loaded from the database.
func CheckGVGMetaConsistent(chainMetaList []*virtualgrouptypes.GlobalVirtualGroup, dbMetaList []*spdb.MigrateGVGUnitMeta) bool {
	if len(chainMetaList) == len(dbMetaList) {
		chainGVGMap := make(map[uint32]*virtualgrouptypes.GlobalVirtualGroup)
		dbGVGMap := make(map[uint32]*spdb.MigrateGVGUnitMeta)
		for _, gvg := range chainMetaList {
			chainGVGMap[gvg.GetId()] = gvg
		}
		for _, dbGVG := range dbMetaList {
			dbGVGMap[dbGVG.GlobalVirtualGroupID] = dbGVG
		}

		for _, chainGVG := range chainMetaList {
			_, ok := dbGVGMap[chainGVG.GetId()]
			if !ok {
				return false
			}
		}

		for _, dbGVG := range dbMetaList {
			_, ok := chainGVGMap[dbGVG.GlobalVirtualGroupID]
			if !ok {
				return false
			}
		}

		return true
	}
	return false
}

// BucketMigrateExecutePlan is used to manage bucket migrate process.
type BucketMigrateExecutePlan struct {
	manager          *ManageModular
	scheduler        *BucketMigrateScheduler
	bucketID         uint64
	gvgUnitMap       map[uint32]*BucketMigrateGVGExecuteUnit // gvgID -> BucketMigrateGVGExecuteUnit
	stopSignal       chan struct{}                           // stop schedule
	finishedGvgUnits map[uint32]struct{}                     // used to count the number of successful migrate units
	srcSP            *sptypes.StorageProvider
}

func newBucketMigrateExecutePlan(manager *ManageModular, bucketID uint64, scheduler *BucketMigrateScheduler, srcSp *sptypes.StorageProvider) *BucketMigrateExecutePlan {
	executePlan := &BucketMigrateExecutePlan{
		manager:          manager,
		scheduler:        scheduler,
		bucketID:         bucketID,
		gvgUnitMap:       make(map[uint32]*BucketMigrateGVGExecuteUnit),
		stopSignal:       make(chan struct{}),
		finishedGvgUnits: make(map[uint32]struct{}),
		srcSP:            srcSp,
	}

	return executePlan
}

// storeToDB persist the BucketMigrateExecutePlan to the database
func (plan *BucketMigrateExecutePlan) storeToDB() error {
	var err error
	for _, migrateGVGUnit := range plan.gvgUnitMap {
		if err = plan.manager.baseApp.GfSpDB().InsertMigrateGVGUnit(&spdb.MigrateGVGUnitMeta{
			MigrateGVGKey:            migrateGVGUnit.Key(),
			GlobalVirtualGroupID:     migrateGVGUnit.SrcGVG.GetId(),
			DestGlobalVirtualGroupID: migrateGVGUnit.DestGVGID,
			VirtualGroupFamilyID:     migrateGVGUnit.SrcGVG.GetFamilyId(),
			RedundancyIndex:          piecestore.PrimarySPRedundancyIndex,
			BucketID:                 migrateGVGUnit.BucketID,
			SrcSPID:                  migrateGVGUnit.SrcSP.GetId(),
			DestSPID:                 migrateGVGUnit.DestSP.GetId(),
			LastMigratedObjectID:     migrateGVGUnit.LastMigratedObjectID,
			MigrateStatus:            int(migrateGVGUnit.MigrateStatus),
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

func (plan *BucketMigrateExecutePlan) UpdateMigrateGVGRetryCount(migrateKey string, retryTime int) error {
	err := plan.manager.baseApp.GfSpDB().UpdateMigrateGVGRetryCount(migrateKey, retryTime)
	if err != nil {
		log.Errorw("failed to update migrate gvg retry time", "migrate_key", migrateKey, "error", err)
		return err
	}
	return nil
}

// QueryMigrateGVG Query migrate GVG unit
func (plan *BucketMigrateExecutePlan) QueryMigrateGVG(migrateKey string) (*spdb.MigrateGVGUnitMeta, error) {
	gvgMeta, err := plan.manager.baseApp.GfSpDB().QueryMigrateGVGUnit(migrateKey)
	if err != nil {
		log.Errorw("failed to query migrate gvg", "migrate_key", migrateKey, "error", err)
		return nil, err
	}
	return gvgMeta, nil
}

// send CompleteMigrateBucket to chain 1) empty bucket: gvgUnitMap is nil; 2) normal bucket
func (plan *BucketMigrateExecutePlan) sendCompleteMigrateBucketTx(migrateExecuteUnit *BucketMigrateGVGExecuteUnit) error {
	var (
		vgfID uint32
		err   error
	)
	if err = UpdateBucketMigrationProgress(plan.manager.baseApp, plan.bucketID, MigratingGvgDone); err != nil {
		return err
	}
	// empty bucket, need to pick a vgf
	if migrateExecuteUnit == nil {
		if vgfID, err = plan.manager.PickVirtualGroupFamily(context.Background(), &task.NullTask{}); err != nil {
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
		vgfID = migrateGVGUnit.DestGVG.GetFamilyId()
		gvgMappings = append(gvgMappings, &storagetypes.GVGMapping{SrcGlobalVirtualGroupId: migrateGVGUnit.SrcGVG.GetId(),
			DstGlobalVirtualGroupId: migrateGVGUnit.DestGVGID, SecondarySpBlsSignature: aggBlsSig})
	}

	migrateBucket := &storagetypes.MsgCompleteMigrateBucket{Operator: plan.manager.baseApp.OperatorAddress(),
		BucketName: bucket.BucketInfo.GetBucketName(), GvgMappings: gvgMappings, GlobalVirtualGroupFamilyId: vgfID}
	txHash, txErr := plan.manager.baseApp.GfSpClient().CompleteMigrateBucket(context.Background(), migrateBucket)
	if txErr != nil {
		log.Errorw("failed to send complete migrate bucket msg to chain", "msg", migrateBucket, "tx_hash", txHash, "err", txErr)
		return txErr
	}
	if err = UpdateBucketMigrationProgress(plan.manager.baseApp, plan.bucketID, SendCompleteTxDone); err != nil {
		return err
	}
	log.Infow("sent complete migrate bucket msg to chain", "msg", migrateBucket, "tx_hash", txHash)
	return nil
}

func (plan *BucketMigrateExecutePlan) rejectBucketMigration() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	bucket, err := plan.manager.baseApp.GfSpClient().GetBucketByBucketID(ctx, int64(plan.bucketID), true)
	if err != nil {
		return err
	}
	rejectMigrateBucket := &storagetypes.MsgRejectMigrateBucket{Operator: plan.manager.baseApp.OperatorAddress(),
		BucketName: bucket.BucketInfo.GetBucketName()}
	txHash, txErr := plan.manager.baseApp.GfSpClient().RejectMigrateBucket(ctx, rejectMigrateBucket)
	if txErr != nil {
		log.Errorw("failed to send reject migrate bucket msg to chain", "msg", rejectMigrateBucket, "tx_hash", txHash, "err", txErr)
		return txErr
	}

	if err = UpdateBucketMigrationProgress(plan.manager.baseApp, bucket.BucketInfo.Id.Uint64(), SendRejectTxDone); err != nil {
		return err
	}

	log.Infow("sent reject migrate bucket msg to chain", "msg", rejectMigrateBucket, "tx_hash", txHash)
	return nil
}

func (plan *BucketMigrateExecutePlan) syncBucketQuotaFromSrcSP(migrateExecuteUnit *BucketMigrateGVGExecuteUnit) error {
	var (
		signature []byte
		err       error
		quota     gfsptask.GfSpBucketQuotaInfo
	)
	bucketID := migrateExecuteUnit.BucketID
	srcSPInfo := migrateExecuteUnit.SrcSP
	log.Infow("start to query quota from src SP", "src_sp", srcSPInfo, "bucket_id", bucketID)

	// query src sp, bucket migrate quota
	queryMsg := &gfsptask.GfSpBucketMigrationInfo{BucketId: bucketID}
	queryMsg.ExpireTime = time.Now().Unix() + SigExpireTimeSecond
	if signature, err = plan.manager.baseApp.GfSpClient().SignBucketMigrationInfo(context.Background(), queryMsg); err != nil {
		log.Errorw("failed to sign migrate bucket", "bucket_migration_info", queryMsg, "error", err)
		return err
	}
	queryMsg.SetSignature(signature)
	if quota, err = plan.manager.baseApp.GfSpClient().QueryLatestBucketQuota(context.Background(), srcSPInfo.GetEndpoint(), queryMsg); err != nil {
		log.Debugw("failed to query bucket quota from src sp", "src_sp", srcSPInfo, "error", err)
		return err
	}

	update := &spdb.BucketTraffic{
		BucketID:              quota.GetBucketId(),
		YearMonth:             quota.GetMonth(),
		BucketName:            quota.GetBucketName(),
		ReadConsumedSize:      quota.GetReadConsumedSize(),
		FreeQuotaConsumedSize: quota.GetFreeQuotaConsumedSize(),
		FreeQuotaSize:         quota.GetFreeQuotaSize(),
		ChargedQuotaSize:      quota.GetChargedQuotaSize(),
	}

	// set dest sp bucket quota info
	if err = plan.manager.baseApp.GfSpDB().UpdateBucketTraffic(bucketID, update); err != nil {
		log.Errorw("failed to update bucket traffic for bucket migrate", "bucket_id", bucketID, "error", err)
		return err
	}

	if err = UpdateBucketMigrationProgress(plan.manager.baseApp, bucketID, MigratingQuotaInfoDone); err != nil {
		return err
	}

	log.Infow("succeed to query quota from src SP", "src_sp", srcSPInfo, "quota", quota)
	return nil
}

func (plan *BucketMigrateExecutePlan) updateMigrateGVGStatus(migrateKey string, task task.MigrateGVGTask, migrateExecuteUnit *BucketMigrateGVGExecuteUnit, migrateStatus MigrateStatus) error {
	var err error
	bucketID := migrateExecuteUnit.BucketID
	gvgUnitsTotal := uint32(len(plan.gvgUnitMap))
	// update migrate gvg status
	if err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateKey, int(migrateStatus)); err != nil {
		log.Errorw("update migrate gvg status", "migrate_key", migrateKey, "error", err)
		return err
	}
	if err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGMigratedBytesSize(migrateKey, task.GetMigratedBytesSize()); err != nil {
		log.Errorw("update migrate gvg migrated bytes size", "migrate_key", migrateKey, "migrated_bytes", task.GetMigratedBytesSize(), "error", err)
		return err
	}

	migrateExecuteUnit.MigrateStatus = migrateStatus
	plan.finishedGvgUnits[migrateExecuteUnit.SrcGVG.GetId()] = struct{}{}

	gvgUnitsFinished := uint32(len(plan.finishedGvgUnits))
	if err = plan.manager.baseApp.GfSpDB().UpdateBucketMigrationMigratingProgress(bucketID, gvgUnitsTotal, gvgUnitsFinished); err != nil {
		log.Errorw("failed to update bucket migration migrating progress", "migrate_key", migrateKey, "total", gvgUnitsTotal, "finished", gvgUnitsFinished, "error", err)
		return err
	}

	// all migrate units success, send tx to chain
	if len(plan.finishedGvgUnits) == len(plan.gvgUnitMap) {
		// set bucket quota
		if err = plan.syncBucketQuotaFromSrcSP(migrateExecuteUnit); err != nil {
			log.Errorw("failed to update bucket quota", "error", err, "migrateExecuteUnit", migrateExecuteUnit)
			return err
		}

		if err = plan.sendCompleteMigrateBucketTx(migrateExecuteUnit); err != nil {
			log.Errorw("failed to send complete migrate bucket msg to chain", "error", err, "migrateExecuteUnit", migrateExecuteUnit)
			return err
		}

	}
	return nil
}

// getBlsAggregateSigForBucketMigration get bls sign from secondary sp which is used for bucket migration
func (plan *BucketMigrateExecutePlan) getBlsAggregateSigForBucketMigration(ctx context.Context, migrateExecuteUnit *BucketMigrateGVGExecuteUnit) ([]byte, error) {
	signDoc := storagetypes.NewSecondarySpMigrationBucketSignDoc(plan.manager.baseApp.ChainID(),
		sdkmath.NewUint(plan.bucketID), migrateExecuteUnit.DestSP.GetId(), migrateExecuteUnit.SrcGVG.GetId(), migrateExecuteUnit.DestGVGID)
	secondarySigs := make([][]byte, 0)
	for _, spID := range migrateExecuteUnit.DestGVG.GetSecondarySpIds() {
		spInfo, err := plan.manager.virtualGroupManager.QuerySPByID(spID)
		if err != nil {
			log.CtxErrorw(ctx, "failed to query sp by id", "error", err, "sp_id", spID)
			return nil, err
		}
		sig, err := plan.manager.baseApp.GfSpClient().GetSecondarySPMigrationBucketApproval(ctx, spInfo.GetEndpoint(), signDoc)
		if err != nil {
			log.Errorw("failed to get secondary sp migration bucket approval", "error", err, "sp_info", spInfo)
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

func (plan *BucketMigrateExecutePlan) startMigrateSchedule() {
	// dispatch to task-dispatcher
	for {
		select {
		case <-plan.stopSignal:
			return // Terminate the scheduling
		default:
			log.Debugw("start to push migrate gvg task to queue", "plan", plan, "gvg_unit_map", plan.gvgUnitMap)

			for _, migrateGVGUnit := range plan.gvgUnitMap {
				// Skipping units that have already been scheduled
				if migrateGVGUnit.MigrateStatus != WaitForMigrate {
					continue
				}

				migrateGVGTask := &gfsptask.GfSpMigrateGVGTask{}
				migrateGVGTask.InitMigrateGVGTask(plan.manager.baseApp.TaskPriority(migrateGVGTask),
					plan.bucketID, migrateGVGUnit.SrcGVG, piecestore.PrimarySPRedundancyIndex,
					migrateGVGUnit.SrcSP,
					plan.manager.baseApp.TaskTimeout(migrateGVGTask, 0),
					plan.manager.baseApp.TaskMaxRetry(migrateGVGTask))
				migrateGVGTask.SetDestGvg(migrateGVGUnit.DestGVG)
				err := plan.manager.migrateGVGQueuePush(migrateGVGTask)
				if err != nil {
					log.Errorw("failed to push migrate gvg task to queue", "error", err)
					time.Sleep(5 * time.Second) // Sleep for 5 seconds before retrying
					continue
				}
				log.Debugw("success to push migrate gvg task to queue", "migrateGVGUnit", migrateGVGUnit, "migrateGVGTask", migrateGVGTask)

				// Update database: migrateStatus to migrating
				if err = plan.manager.baseApp.GfSpDB().UpdateMigrateGVGUnitStatus(migrateGVGUnit.Key(), int(Migrating)); err != nil {
					log.Errorw("failed to update migrate gvg status", "gvg_unit", migrateGVGUnit, "error", err)
					return
				}
				if err = UpdateBucketMigrationProgress(plan.manager.baseApp, plan.bucketID, MigratingGvgDoing); err != nil {
					return
				}
				migrateGVGUnit.MigrateStatus = Migrating
			}

			time.Sleep(1 * time.Minute) // Sleep for 1 minute before next iteration
		}
	}
}

func (plan *BucketMigrateExecutePlan) stopSPSchedule() {
	plan.stopSignal <- struct{}{}
}

func (plan *BucketMigrateExecutePlan) Start() error {
	log.Debugf("succeed to start bucket migrate plan", "plan", *plan)
	go plan.startMigrateSchedule()
	return nil
}

// BucketMigrateScheduler subscribes bucket migrate events and produces a gvg migrate plan.
type BucketMigrateScheduler struct {
	manager                     *ManageModular
	selfSP                      *sptypes.StorageProvider
	lastSubscribedBlockHeight   uint64                               // load from db
	lastSubscribedBlockHeightGC uint64                               // src sp subscribe gc progress
	executePlanIDMap            map[uint64]*BucketMigrateExecutePlan // bucketID -> BucketMigrateExecutePlan
	bucketCache                 *BucketCache
	mutex                       sync.RWMutex // Protects the executePlanIDMap fields
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

	s.bucketCache = NewBucketCache(bucketCacheSize, bucketCacheExpire)
	if err != nil {
		return err
	}

	// plan load from db
	s.loadBucketMigrateExecutePlansFromDB()

	log.Infow("succeed to init bucket migrate scheduler", "self_sp", s.selfSP,
		"last_subscribed_block_height", s.lastSubscribedBlockHeight,
		"execute_plans", s.executePlanIDMap)

	return nil
}

func (s *BucketMigrateScheduler) Start() error {
	go s.subscribeEvents()
	go s.confirmEvents()
	return nil
}

// Before processing MigrateBucketEvents, first check if the status of the bucket on the chain meets the expectations. If it meets the expectations, proceed with the execution; otherwise, skip this MigrateBucketEvent event.
func (s *BucketMigrateScheduler) checkBucketFromChain(bucketID uint64, expectedStatus storagetypes.BucketStatus) (expected bool, err error) {
	// check the chain's bucket is migrating
	key := bucketCacheKey(bucketID)
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
		value, ok := elem.(*storagetypes.BucketInfo)
		if !ok {
			log.Debugw("failed to get bucket info from bucket cache", "key", key)
			s.bucketCache.Delete(key)
			err = QueryBucketInfoFromChainFunc()
		} else {
			bucketInfo = value
		}
	} else {
		err = QueryBucketInfoFromChainFunc()
	}

	if err != nil || bucketInfo == nil {
		return false, err
	}

	if bucketInfo.BucketStatus != expectedStatus {
		log.Debugw("the bucket status is not same, the event will skip", "bucketInfo", bucketInfo, "expectedStatus", expectedStatus)
		return false, nil
	}
	return true, nil
}

func (s *BucketMigrateScheduler) getMigratedBytesSize(bucketID uint64) (uint64, error) {
	var (
		err                error
		migrateGVGUnitMeta []*spdb.MigrateGVGUnitMeta
		migratedBytes      uint64
	)
	if migrateGVGUnitMeta, err = s.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByBucketID(bucketID); err != nil {
		return 0, err
	}

	for _, migrateUnit := range migrateGVGUnitMeta {
		migratedBytes += migrateUnit.MigratedBytesSize
	}

	return migratedBytes, nil
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

	// notify src sp to gc
	postMsg := &gfsptask.GfSpBucketMigrationInfo{BucketId: bucketID, Finished: true}
	if err = executePlan.manager.bucketMigrateScheduler.PostMigrateBucket(postMsg, executePlan.srcSP); err != nil {
		log.Errorw("failed to post migrate bucket", "msg", postMsg, "error", err)
	}

	for _, unit := range executePlan.gvgUnitMap {
		if unit.MigrateStatus != Migrated {
			log.Errorw("report task may error, unit should be migrated", "unit", unit)
		}
	}
	s.deleteExecutePlanByBucketID(bucketID)
	executePlan.stopSPSchedule()
	err = s.manager.baseApp.GfSpDB().DeleteMigrateGVGUnitsByBucketID(bucketID)

	return err
}

func (s *BucketMigrateScheduler) cancelMigrateBucket(bucketID uint64, reject bool) error {
	var (
		executePlan *BucketMigrateExecutePlan
		err         error
		state       BucketMigrateState
	)
	ctx := context.Background()
	if executePlan, err = s.getExecutePlanByBucketID(bucketID); err != nil {
		log.Errorw("bucket migrate schedule received EventCancelMigrationBucket", "bucket_id", bucketID, "error", err)
		return err
	}
	if reject {
		state = WaitRejectTxEventDone
	} else {
		state = WaitCancelTxEventDone
	}
	if err = UpdateBucketMigrationProgress(executePlan.manager.baseApp, bucketID, state); err != nil {
		return err
	}

	for _, migrateGVGUnit := range executePlan.gvgUnitMap {
		key := gfsptask.GfSpMigrateGVGTaskKey(migrateGVGUnit.SrcGVG.GetId(), migrateGVGUnit.BucketID, piecestore.PrimarySPRedundancyIndex)
		s.manager.migrateGVGQueuePopByKey(key)
	}
	s.deleteExecutePlanByBucketID(bucketID)

	executePlan.stopSPSchedule()
	if err = s.manager.baseApp.GfSpDB().DeleteMigrateGVGUnitsByBucketID(bucketID); err != nil {
		return err
	}

	// if bucket migration failed, gc for dest sp
	// generate a gc bucket migration task(list objects and delete)
	if err = UpdateBucketMigrationProgress(executePlan.manager.baseApp, bucketID, DestSPGCDoing); err != nil {
		return err
	}
	go s.manager.GenerateGCBucketMigrationTask(ctx, bucketID)

	log.CtxInfow(ctx, "succeed to cancel migration event from memory, the bucket migration will generate a gc task", "bucket_id", bucketID)

	return err
}

func (s *BucketMigrateScheduler) processEvents(migrateBucketEvents *types.ListMigrateBucketEvents) error {
	var (
		err         error
		executePlan *BucketMigrateExecutePlan
	)
	// 1. process EventCancelMigrationBucket
	if migrateBucketEvents.CancelEvent != nil {
		// no need to process cancel migration event, maybe already canceled
		if _, err = s.getExecutePlanByBucketID(migrateBucketEvents.CancelEvent.BucketId.Uint64()); err != nil {
			log.Infow("bucket migrate schedule received EventCancelMigrationBucket", "bucket_id", migrateBucketEvents.CancelEvent.BucketId.Uint64(), "error", err)
			return nil
		}
		log.Infow("begin to process cancel events", "cancel_event", migrateBucketEvents.CancelEvent)
		if err = s.cancelMigrateBucket(migrateBucketEvents.CancelEvent.BucketId.Uint64(), false); err != nil {
			log.Errorw("failed to process cancel events", "cancel_event", migrateBucketEvents.CancelEvent, "error", err)
		}
		return nil
	}

	// 2. process RejectEvents
	if migrateBucketEvents.RejectEvent != nil {
		log.Infow("begin to process reject events", "reject_event", migrateBucketEvents.RejectEvent)

		if err = s.cancelMigrateBucket(migrateBucketEvents.RejectEvent.BucketId.Uint64(), true); err != nil {
			log.Errorw("failed to process cancel events", "cancel_event", migrateBucketEvents.CancelEvent, "error", err)
		}
		return nil
	}

	// 3. process EventCompleteMigrationBucket
	if migrateBucketEvents.CompleteEvent != nil {
		return nil
	}
	// 4. process EventMigrationBucket
	if migrateBucketEvents.Event != nil {
		if executePlan, err = s.produceBucketMigrateExecutePlan(migrateBucketEvents.Event, false); err != nil || executePlan == nil {
			log.Errorw("failed to produce bucket migrate execute plan", "Events", migrateBucketEvents.Event, "error", err)
			return err
		}
		if err = executePlan.Start(); err != nil {
			log.Errorw("failed to start bucket migrate execute plan", "Events", migrateBucketEvents.Event, "executePlan", executePlan, "error", err)
			return err
		}
		s.executePlanIDMap[executePlan.bucketID] = executePlan
	}
	return nil
}

func (s *BucketMigrateScheduler) confirmCompleteTxEvents(ctx context.Context, event *spdb.MigrateBucketProgressMeta) {
	var (
		bucket *types.Bucket
		err    error
	)
	bucketID := event.BucketID

	if event.MigrationState != int(SendCompleteTxDone) {
		return
	}
	// confirm
	if bucket, err = s.manager.baseApp.GfSpClient().GetBucketByBucketID(ctx, int64(bucketID), true); err != nil {
		log.Errorw("failed to get bucket by bucket id", "bucket_id", bucketID, "error", err)
		return
	}

	if bucket.BucketInfo.GetBucketStatus() == storagetypes.BUCKET_STATUS_CREATED {
		if err = UpdateBucketMigrationProgress(s.manager.baseApp, bucketID, WaitCompleteTxEventDone); err != nil {
			return
		}
		if err = s.doneMigrateBucket(bucketID); err != nil {
			log.Errorw("failed to done migrate bucket", "error", err, "EventMigrationBucket", event)
			return
		}
	}
}

// After the dest sp completes the migration, it needs to send the CompleteMigrateBucketTx to the chain.
// The dest sp waits for the successful execution of the transaction and then proceeds to 1) notify the source SP and 2) clean up its own state.
func (s *BucketMigrateScheduler) confirmEvents() {
	subscribeBucketMigrateEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeBucketMigrateEventInterval) * time.Millisecond)
	defer subscribeBucketMigrateEventsTicker.Stop()
	ctx := context.Background()
	logNumber := uint64(0)

	for range subscribeBucketMigrateEventsTicker.C {
		confirmEvents, listError := s.manager.baseApp.GfSpDB().ListBucketMigrationToConfirm()
		if listError != nil {
			logNumber++
			if (logNumber % printLogPerN) == 0 {
				log.Errorw("failed to list migrate bucket events to confirm", "block_id", s.lastSubscribedBlockHeight+1,
					"error", listError)
			}
			continue
		}
		for _, event := range confirmEvents {
			s.confirmCompleteTxEvents(ctx, event)
		}
	}
}

func (s *BucketMigrateScheduler) subscribeEvents() {
	go func() {
		logNumber := uint64(0)

		UpdateBucketMigrateSubscribeProgressFunc := func(num uint64) {
			updateErr := s.manager.baseApp.GfSpDB().UpdateBucketMigrateSubscribeProgress(s.lastSubscribedBlockHeight + 1)
			if updateErr != nil {
				log.Errorw("failed to update bucket migrate progress", "error", updateErr)
			}
			s.lastSubscribedBlockHeight++
			if (num % printLogPerN) == 0 {
				log.Infow("bucket migrate subscribe progress", "last_subscribed_block_height", s.lastSubscribedBlockHeight)
			}
		}
		subscribeBucketMigrateEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeBucketMigrateEventInterval) * time.Millisecond)
		defer subscribeBucketMigrateEventsTicker.Stop()

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

			// 2. make plan, start plan
			for _, migrateBucketEvents := range migrationBucketEvents {
				log.Infow("loop subscribe bucket migrate event", "migration_bucket_events", migrationBucketEvents, "block_id", s.lastSubscribedBlockHeight+1, "sp_address", s.manager.baseApp.OperatorAddress())
				if err := s.processEvents(migrateBucketEvents); err != nil {
					log.Errorw("bucket migrate process error", "migration_bucket_events", migrateBucketEvents, "error", err)
				}
			}

			// 3.update subscribe progress to db
			UpdateBucketMigrateSubscribeProgressFunc(logNumber)
		}
	}()
	// src sp subscribe complete migrationBucketEvents for gc
	go func() {
		logNumber := uint64(0)

		UpdateBucketMigrateGCSubscribeProgressFunc := func(num uint64) {
			if updateErr := s.manager.baseApp.GfSpDB().UpdateBucketMigrateGCSubscribeProgress(s.lastSubscribedBlockHeightGC + 1); updateErr != nil {
				log.Errorw("failed to update bucket migrate src sp gc progress", "error", updateErr)
			}
			s.lastSubscribedBlockHeightGC++
			if (num % printLogPerN) == 0 {
				log.Infow("src sp gc bucket migrate subscribe progress", "last_subscribed_block_height", s.lastSubscribedBlockHeightGC)
			}
		}
		subscribeBucketMigrateEventsTicker := time.NewTicker(time.Duration(s.manager.subscribeBucketMigrateEventInterval) * time.Millisecond)
		defer subscribeBucketMigrateEventsTicker.Stop()
		var (
			completeMigrationBucketEvents []*storagetypes.EventCompleteMigrationBucket
			subscribeError                error
		)

		for range subscribeBucketMigrateEventsTicker.C {
			// 1. subscribe migrate bucket events
			if completeMigrationBucketEvents, subscribeError = s.manager.baseApp.GfSpClient().ListCompleteMigrationBucketEvents(
				context.Background(), s.lastSubscribedBlockHeightGC+1, s.selfSP.GetId()); subscribeError != nil {
				logNumber++
				if (logNumber % printLogPerN) == 0 {
					log.Errorw("failed to list migrate completed bucket events", "block_id", s.lastSubscribedBlockHeightGC+1,
						"error", subscribeError)
				}
				continue
			}

			// 2. src confirm complete migration bucket event and update progress
			for _, migrateBucketEvents := range completeMigrationBucketEvents {
				log.Infow("loop subscribe completed bucket migrate event", "complete_migration_bucket_events", completeMigrationBucketEvents, "block_id", s.lastSubscribedBlockHeight+1, "sp_address", s.manager.baseApp.OperatorAddress())
				bucketID := migrateBucketEvents.BucketId.Uint64()
				ctx := context.Background()

				if err := UpdateBucketMigrationProgress(s.manager.baseApp, bucketID, SrcSPGCDoing); err != nil {
					return
				}

				go s.manager.GenerateGCBucketMigrationTask(ctx, bucketID)
			}

			// 3.update subscribe progress to db
			UpdateBucketMigrateGCSubscribeProgressFunc(logNumber)
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

// PostMigrateBucket is used to pick a suitable gvg for replicating object.
func (s *BucketMigrateScheduler) PostMigrateBucket(postMsg *gfsptask.GfSpBucketMigrationInfo, srcSPInfo *sptypes.StorageProvider) error {
	var (
		signature []byte
		err       error
	)

	if srcSPInfo == nil {
		srcSPInfo = s.getSPInfoByBucketID(postMsg.GetBucketId())
	}

	postMsg.ExpireTime = time.Now().Unix() + SigExpireTimeSecond
	if signature, err = s.manager.baseApp.GfSpClient().SignBucketMigrationInfo(context.Background(), postMsg); err != nil {
		log.Errorw("failed to sign migrate bucket", "bucket_migration", postMsg, "error", err)
		return err
	}
	postMsg.SetSignature(signature)

	if _, err = s.manager.baseApp.GfSpClient().PostMigrateBucket(context.Background(), srcSPInfo.GetEndpoint(), postMsg); err != nil {
		log.Debugw("failed to query bucket quota from src sp", "src_sp", srcSPInfo, "error", err)
		return err
	}
	log.Debugw("succeed to post migrate bucket quota", "src_sp", srcSPInfo, "postMsg", postMsg, "error", err)

	return nil
}

// PreMigrateBucket is used to Dest SP notifies Src SP and pre-deducts quota.
func (s *BucketMigrateScheduler) PreMigrateBucket(bucketID uint64, srcSPInfo *sptypes.StorageProvider) error {
	var (
		signature []byte
		err       error
	)

	log.Debugw("start to pre migrate bucket", "bucket_id", bucketID)
	preMsg := &gfsptask.GfSpBucketMigrationInfo{BucketId: bucketID, Finished: false}
	preMsg.ExpireTime = time.Now().Unix() + SigExpireTimeSecond
	if signature, err = s.manager.baseApp.GfSpClient().SignBucketMigrationInfo(context.Background(), preMsg); err != nil {
		log.Errorw("failed to sign migrate bucket", "bucket_migration", preMsg, "error", err)
		return err
	}
	preMsg.SetSignature(signature)

	// query src sp, bucket migrate quota
	if _, err = s.manager.baseApp.GfSpClient().PreMigrateBucket(context.Background(), srcSPInfo.GetEndpoint(), preMsg); err != nil {
		log.Debugw("failed to query bucket quota from src sp", "src_sp", srcSPInfo, "error", err)
		return err
	}
	return nil
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

	gvgMeta, err := manager.virtualGroupManager.GenerateGlobalVirtualGroupMeta(NewGenerateGVGSecondarySPsPolicyByPrefer(params, manager.gvgPreferSPList), vgmgr.NewExcludeIDFilter(gfspvgmgr.NewIDSetFromList(manager.spBlackList)))
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

func (s *BucketMigrateScheduler) getSPInfoByBucketID(bucketID uint64) *sptypes.StorageProvider {
	bucketInfo, err := s.manager.baseApp.Consensus().QueryBucketInfoById(context.Background(), bucketID)
	if err != nil {
		return nil
	}
	bucketSPID, err := util.GetBucketPrimarySPID(context.Background(), s.manager.baseApp.Consensus(), bucketInfo)
	if err != nil {
		return nil
	}

	spInfo, err := s.manager.virtualGroupManager.QuerySPByID(bucketSPID)
	if err != nil {
		log.Errorw("failed to query sp", "error", err, "sp_id", bucketSPID)
		return nil
	}
	return spInfo
}

func (s *BucketMigrateScheduler) getSrcSPAndDestSPFromMigrateEvent(event *storagetypes.EventMigrationBucket) (srcSP, destSP *sptypes.StorageProvider, err error) {
	srcSP = s.getSPInfoByBucketID(event.BucketId.Uint64())
	if srcSP == nil {
		return nil, nil, err
	}

	destSP, err = s.manager.virtualGroupManager.QuerySPByID(event.DstPrimarySpId)
	if err != nil {
		log.Errorw("failed to query sp", "error", err, "migration_bucket_events", event)
		return nil, nil, err
	}
	return srcSP, destSP, nil
}

func (s *BucketMigrateScheduler) produceBucketMigrateExecutePlan(event *storagetypes.EventMigrationBucket, buildMetaByDB bool) (*BucketMigrateExecutePlan, error) {
	var (
		primarySPGVGList   []*virtualgrouptypes.GlobalVirtualGroup
		plan               *BucketMigrateExecutePlan
		err                error
		migrateBucketUnits []*BucketMigrateGVGExecuteUnit
		migrateGVGUnitMeta []*spdb.MigrateGVGUnitMeta
	)

	// 1) check chain's bucket status
	expected, err := s.checkBucketFromChain(event.BucketId.Uint64(), storagetypes.BUCKET_STATUS_MIGRATING)
	if err != nil {
		return nil, err
	}
	if !expected {
		return nil, nil
	}
	if s.executePlanIDMap[event.BucketId.Uint64()] != nil {
		log.Debugw("the bucket is already in migrating", "migration_bucket_events", event)
		return nil, errors.New("bucket already in migrating")
	}

	// 2) new bucket migrate execute plan
	bucketID := event.BucketId.Uint64()

	log.Debugw("produce bucket migrate execute plan", "bucketID", bucketID, "migration_bucket_events", event)
	// query metadata service to get primary sp's gvg list.
	primarySPGVGList, err = s.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), bucketID)
	if err != nil {
		log.Errorw("failed to list gvg ", "error", err, "migration_bucket_events", event)
		return nil, errors.New("failed to list gvg")
	}

	srcSP, destSP, err := s.getSrcSPAndDestSPFromMigrateEvent(event)
	if err != nil {
		return nil, err
	}

	plan = newBucketMigrateExecutePlan(s.manager, event.BucketId.Uint64(), s, srcSP)
	conflictChecker := NewSPConflictChecker(plan, srcSP, destSP, bucketID)

	if buildMetaByDB {
		migrateGVGUnitMeta, err = s.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByBucketID(bucketID)
		if err != nil {
			return nil, err
		}

		// 2) not match, generate migrate gvg again
		if !CheckGVGMetaConsistent(primarySPGVGList, migrateGVGUnitMeta) {
			// delete db & gerenate again
			err = s.manager.baseApp.GfSpDB().DeleteMigrateGVGUnitsByBucketID(bucketID)
			if err != nil {
				return nil, err
			}
			migrateBucketUnits, err = conflictChecker.GenerateMigrateBucketUnits(false)
		} else {
			migrateBucketUnits, err = conflictChecker.GenerateMigrateBucketUnits(true)
		}
	} else {
		migrateBucketUnits, err = conflictChecker.GenerateMigrateBucketUnits(false)
	}

	if err != nil {
		log.Errorw("failed to produce bucket migrate plan", "error", err, "bucket_id", bucketID)
		return nil, err
	}

	for _, migrateBucketUnit := range migrateBucketUnits {
		plan.gvgUnitMap[migrateBucketUnit.SrcGVG.GetId()] = migrateBucketUnit
	}

	log.Debugw("produce bucket migrate execute plan list", "primarySPGVGList", primarySPGVGList, "bucket_gvg_list_len", len(primarySPGVGList), "EventMigrationBucket", event)
	if len(plan.gvgUnitMap) == 0 {
		// an empty bucket ends here, and it will not return a plan. The execution will not continue.
		err = plan.sendCompleteMigrateBucketTx(nil)
		if err != nil {
			log.Errorw("failed to send complete migrate bucket msg to chain", "error", err, "EventMigrationBucket", event)
			return nil, err
		}

		log.Infow("bucket is empty, send complete migrate bucket tx directly", "bucket_id", bucketID)
		return nil, nil
	}

	// lock quota
	if !buildMetaByDB {
		if err = plan.manager.bucketMigrateScheduler.PreMigrateBucket(bucketID, plan.srcSP); err != nil {
			log.Errorw("failed to pre migrate bucket(lock src sp quota)", "bucket_id", bucketID, "error", err)
			return nil, err
		}
		if err = UpdateBucketMigrationProgress(plan.manager.baseApp, bucketID, DestSPPreDeductQuotaDone); err != nil {
			return nil, err
		}
	}

	if err = plan.storeToDB(); err != nil {
		log.Errorw("failed to generate migrate execute plan due to store db", "error", err)
	}

	return plan, err
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

func (s *BucketMigrateScheduler) deleteExecutePlanByBucketID(bucketID uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.executePlanIDMap[bucketID]
	if ok {
		delete(s.executePlanIDMap, bucketID)
	} else {
		log.Debugw("failed to the delete execute plan from the map due to delete an nonexistent bucket", "bucket_id", bucketID)
	}
}

func (s *BucketMigrateScheduler) listExecutePlan() (*gfspserver.GfSpQueryBucketMigrateResponse, error) {
	var res gfspserver.GfSpQueryBucketMigrateResponse
	var plans []*gfspserver.GfSpBucketMigrate
	for _, executePlan := range s.executePlanIDMap {
		var (
			plan              gfspserver.GfSpBucketMigrate
			migratedBytesSize uint64
			err               error
		)

		plan.BucketId = executePlan.bucketID
		plan.Finished = uint32(len(executePlan.finishedGvgUnits))
		if migratedBytesSize, err = s.getMigratedBytesSize(executePlan.bucketID); err != nil {
			// if something error, set migratedBytesSize to zero
			log.Errorw("failed to get migrated bytes size", "execute_plan", executePlan, "error", err)
			migratedBytesSize = 0
		}
		plan.MigratedBytesSize = migratedBytesSize
		for _, unit := range executePlan.gvgUnitMap {
			plan.GvgTask = append(plan.GvgTask, &gfspserver.GfSpMigrateGVG{
				SrcGvgId:             unit.SrcGVG.GetId(),
				DestGvgId:            unit.DestGVG.GetId(),
				LastMigratedObjectId: unit.LastMigratedObjectID,
				Status:               int32(unit.MigrateStatus),
			})
		}
		plans = append(plans, &plan)
	}
	res.BucketMigrate = plans
	res.SelfSpId = s.selfSP.GetId()
	log.Debugw("succeed to query bucket migrate", "response", res)
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
		// maybe bucket migrate canceled
		log.Debugw("failed to update migrate progress", "task", task)
		return fmt.Errorf("gvg unit is not found")
	}
	migrateKey := MakeBucketMigrateKey(migrateExecuteUnit.BucketID, migrateExecuteUnit.SrcGVG.GetId())

	if task.GetFinished() {
		if err = executePlan.updateMigrateGVGStatus(migrateKey, task, migrateExecuteUnit, Migrated); err != nil {
			log.Errorw("failed to update migrate gvg status", "migrate_key", migrateKey, "error", err)
			return err
		}
	} else {
		// The report task from executor keeps track of the err when acquiring data from the src SP, if a gvg's migration
		// continuously fails(The LastMigratedObjectID stays still); it will send a tx to reject the migration
		if task.Error() != nil {
			log.Errorw("report migrate progress task", "migrate_key", migrateKey, "error", task.Error())
			migrateGVGUnit, queryErr := executePlan.QueryMigrateGVG(migrateKey)
			if queryErr != nil {
				log.Errorw("failed to query migrate gvg unit", "error", queryErr)
				return queryErr
			}
			log.Debugw("migrateGVGUnit", "migrate_key", migrateKey, "migrateGVGUnit", migrateGVGUnit)
			if task.GetLastMigratedObjectID() == migrateGVGUnit.LastMigratedObjectID {
				if migrateGVGUnit.RetryTime+1 >= migrateGVGTaskMaxRetry {
					if err = executePlan.rejectBucketMigration(); err != nil {
						log.Errorw("failed to send reject bucket migration tx to chain", "error", err, "migrateExecuteUnit", migrateExecuteUnit)
						return err
					}
					return nil
				}
				if err = executePlan.UpdateMigrateGVGRetryCount(migrateKey, migrateGVGUnit.RetryTime+1); err != nil {
					log.Errorw("failed to update migrate gvg retry count", "migrate_key", migrateKey, "error", err)
					return err
				}
				return nil
			}
		}
		err = executePlan.UpdateMigrateGVGLastMigratedObjectID(migrateKey, task.GetLastMigratedObjectID())
		if err != nil {
			log.Errorw("failed to update migrate gvg last migrate object id", "migrate_key", migrateKey, "error", err)
			return err
		}
		migrateExecuteUnit.LastMigratedObjectID = task.GetLastMigratedObjectID()
	}
	return nil
}

func (s *BucketMigrateScheduler) UpdateBucketMigrationGCProgress(ctx context.Context, gcBucketMigrationTask task.GCBucketMigrationTask) error {
	// update gc progress
	var state BucketMigrateState
	if gcBucketMigrationTask.GetFinished() {
		state = MigrationFinished
	} else {
		state = SrcSPGCDoing
	}

	meta := spdb.MigrateBucketProgressMeta{
		BucketID:         gcBucketMigrationTask.GetBucketID(),
		MigrationState:   int(state),
		LastGcObjectID:   gcBucketMigrationTask.GetLastGCObjectID(),
		LastGcGvgID:      gcBucketMigrationTask.GetLastGCGvgID(),
		TotalGvgNum:      uint32(gcBucketMigrationTask.GetTotalGvgNum()),
		GcFinishedGvgNum: uint32(gcBucketMigrationTask.GetGCFinishedGvgNum()),
	}
	if err := s.manager.baseApp.GfSpDB().UpdateBucketMigrationGCProgress(meta); err != nil {
		log.CtxErrorw(ctx, "failed to update bucket migration gc progress", "task", gcBucketMigrationTask, "error", err)
		return err
	}
	return nil
}

// loadBucketMigrateExecutePlansFromDB 1) subscribe progress 2) plan progress 3) task progress
func (s *BucketMigrateScheduler) loadBucketMigrateExecutePlansFromDB() error {
	var (
		migrationBucketEvents []*types.ListMigrateBucketEvents
		err                   error
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
		// if it has CompleteEvents & CancelEvents, skip it
		if migrateBucketEvents.CompleteEvent != nil || migrateBucketEvents.CancelEvent != nil || migrateBucketEvents.RejectEvent != nil {
			continue
		}
		if migrateBucketEvents.Event != nil {
			migratingBucketIDs = append(migratingBucketIDs, migrateBucketEvents.Event.BucketId.Uint64())
			migrationBucketEventsMap[migrateBucketEvents.Event.BucketId.Uint64()] = migrateBucketEvents
		}
	}
	log.Infow("load bucket migrate execute plans from db", "bucket_ids", migratingBucketIDs)
	// load from db by BucketID & construct plan
	for _, bucketID := range migratingBucketIDs {
		bucketMigrateEvent := migrationBucketEventsMap[bucketID]

		plan, err := s.produceBucketMigrateExecutePlan(bucketMigrateEvent.Event, true)
		if err != nil {
			return err
		}
		if plan != nil {
			log.Debugw("bucket migrate scheduler load from db", "execute_plan", plan, "bucket_id", bucketID)
			if err = plan.Start(); err != nil {
				log.Errorw("failed to start bucket migrate execute plan", "events", bucketMigrateEvent.Event, "execute_plan", plan, "error", err)
				return err
			}
			s.executePlanIDMap[plan.bucketID] = plan
		}
	}

	log.Debugw("bucket migrate scheduler load from db success", "bucket_ids", migratingBucketIDs)
	return err
}

func bucketCacheKey(bucketId uint64) string {
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
				// delete
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

type SPConflictChecker struct {
	plan     *BucketMigrateExecutePlan
	srcSP    *sptypes.StorageProvider
	selfSP   *sptypes.StorageProvider
	bucketID uint64
}

// NewSPConflictChecker returns a SP conflicted checker instance.
func NewSPConflictChecker(p *BucketMigrateExecutePlan, src, dest *sptypes.StorageProvider, bucketID uint64) *SPConflictChecker {
	return &SPConflictChecker{
		plan:     p,
		srcSP:    src,
		selfSP:   dest,
		bucketID: bucketID,
	}
}

// replace SecondarySp which is in STATUS_GRACEFUL_EXITING
func (checker *SPConflictChecker) replaceExitingSP(secondarySPIDs []uint32) ([]uint32, error) {
	replacedSPIDs := secondarySPIDs
	excludedSPIDs := secondarySPIDs

	for idx, spID := range secondarySPIDs {
		sp, err := checker.plan.manager.virtualGroupManager.QuerySPByID(spID)
		if err != nil {
			log.Errorw("failed to query sp", "error", err)
			return nil, err
		}
		if sp.Status == sptypes.STATUS_GRACEFUL_EXITING {
			replacedSP, pickErr := checker.plan.manager.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(excludedSPIDs))
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

func (checker *SPConflictChecker) generateMigrateBucketUnitsFromDB(primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup) ([]*BucketMigrateGVGExecuteUnit, error) {
	bucketID := checker.bucketID
	var bucketMigrateUnits []*BucketMigrateGVGExecuteUnit
	migrateGVGUnitMeta, err := checker.plan.manager.baseApp.GfSpDB().ListMigrateGVGUnitsByBucketID(bucketID)
	if err != nil {
		return nil, err
	}

	chainGVGMap := make(map[uint32]*virtualgrouptypes.GlobalVirtualGroup)
	for _, gvg := range primarySPGVGList {
		chainGVGMap[gvg.GetId()] = gvg
	}

	// Using migrateGVGUnitMeta to construct PrimaryGVGIDMapMigrateUnits and execute them one by one.
	for _, migrateGVG := range migrateGVGUnitMeta {
		srcSP, queryErr := checker.plan.manager.virtualGroupManager.QuerySPByID(migrateGVG.SrcSPID)
		if queryErr != nil {
			log.Errorw("failed to query sp", "error", queryErr, "sp_info", srcSP)
			return nil, queryErr
		}
		destSP, queryErr := checker.plan.manager.virtualGroupManager.QuerySPByID(migrateGVG.DestSPID)
		if queryErr != nil {
			log.Errorw("failed to query sp", "error", queryErr, "sp_info", destSP)
			return nil, queryErr
		}

		srcGvg, ok := chainGVGMap[migrateGVG.GlobalVirtualGroupID]
		if ok {
			destGvg, err := checker.plan.manager.baseApp.Consensus().QueryGlobalVirtualGroup(context.Background(), migrateGVG.DestGlobalVirtualGroupID)
			if err != nil {
				return nil, err
			}
			bucketUnit := newBucketMigrateGVGExecuteUnit(bucketID, srcGvg, srcSP, destSP, MigrateStatus(migrateGVG.MigrateStatus), migrateGVG.DestSPID, migrateGVG.LastMigratedObjectID, destGvg)
			bucketMigrateUnits = append(bucketMigrateUnits, bucketUnit)
			log.Debugw("succeeded to generate a new bucket migrate uint", "migrate_execute_unit", bucketUnit, "bucket_id", bucketID)
		} else {
			// gvgChecker has verified and this scenario should not occur.
			log.Debugw("failed to get src gvg", "gvg_id", migrateGVG.GlobalVirtualGroupID)
		}
	}

	return bucketMigrateUnits, nil
}

func (checker *SPConflictChecker) generateMigrateBucketUnitsFromMemory(primarySPGVGList []*virtualgrouptypes.GlobalVirtualGroup) ([]*BucketMigrateGVGExecuteUnit, error) {
	var (
		destGVG            *vgmgr.GlobalVirtualGroupMeta
		destFamilyID       uint32
		bucketMigrateUnits []*BucketMigrateGVGExecuteUnit
	)
	for _, srcGVG := range primarySPGVGList {
		srcSecondarySPIDs := make([]uint32, len(srcGVG.GetSecondarySpIds()))
		copy(srcSecondarySPIDs, srcGVG.GetSecondarySpIds())
		// check sp exiting
		secondarySPIDs, err := checker.replaceExitingSP(srcSecondarySPIDs)
		if err != nil {
			log.Errorw("failed to pick sp to replace exiting sp", "srcSecondarySPIDs", srcSecondarySPIDs, "secondarySPIDs", secondarySPIDs, "bucket_id", checker.bucketID)
			return nil, err
		}

		// check conflicts.
		conflictedIndex, errNotInSecondarySPs := util.GetSecondarySPIndexFromGVG(srcGVG, checker.selfSP.GetId())
		log.Debugw("prepare to check conflicts", "srcGVG", srcGVG, "destSP", checker.selfSP, "conflictedIndex", conflictedIndex, "bucket_id", checker.bucketID)
		if errNotInSecondarySPs == nil {
			// gvg has conflicts.
			excludedSPIDs := srcGVG.GetSecondarySpIds()
			replacedSP, pickErr := checker.plan.manager.virtualGroupManager.PickSPByFilter(NewPickDestSPFilterWithSlice(excludedSPIDs))
			if pickErr != nil {
				log.Errorw("failed to pick new sp to replace conflict secondary sp", "srcGVG", srcGVG, "destSP", checker.selfSP, "excludedSPIDs", excludedSPIDs, "error", pickErr, "bucket_id", checker.bucketID)
				return nil, pickErr
			}
			secondarySPIDs[conflictedIndex] = replacedSP.GetId()
			log.Debugw("succeeded to resolve conflict", "excludedSPIDs", excludedSPIDs, "bucket_id", checker.bucketID)
		}
		log.Debugw("prepare to pick new gvg", "secondarySPIDs", secondarySPIDs, "bucket_id", checker.bucketID)
		destGVG, err = checker.plan.scheduler.pickGlobalVirtualGroupForBucketMigrate(NewPickDestGVGFilter(destFamilyID, secondarySPIDs, srcGVG.StoredSize))
		if err != nil {
			log.Errorw("failed to pick gvg for migrate bucket", "error", err, "bucket_id", checker.bucketID)
			return nil, err
		}
		destFamilyID = destGVG.FamilyID
		destGlobalVirtualGroup := &virtualgrouptypes.GlobalVirtualGroup{Id: destGVG.ID, FamilyId: destGVG.FamilyID, PrimarySpId: destGVG.PrimarySPID,
			SecondarySpIds: destGVG.SecondarySPIDs, StoredSize: destGVG.StakingStorageSize}

		bucketUnit := newBucketMigrateGVGExecuteUnit(checker.bucketID, srcGVG, checker.srcSP, checker.selfSP, WaitForMigrate, destGVG.ID, 0, destGlobalVirtualGroup)
		bucketMigrateUnits = append(bucketMigrateUnits, bucketUnit)
		log.Infow("succeeded to generate a new bucket migrate unit", "migrate_execute_unit", bucketUnit, "bucket_id", checker.bucketID)
	}

	return bucketMigrateUnits, nil
}

// GenerateMigrateBucketUnits generate the migrate bucket units.
func (checker *SPConflictChecker) GenerateMigrateBucketUnits(buildMetaByDB bool) ([]*BucketMigrateGVGExecuteUnit, error) {
	var (
		err                error
		primarySPGVGList   []*virtualgrouptypes.GlobalVirtualGroup
		bucketMigrateUnits []*BucketMigrateGVGExecuteUnit
	)
	primarySPGVGList, err = checker.plan.manager.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(context.Background(), checker.bucketID)
	if err != nil {
		log.Errorw("failed to list gvg", "error", err, "bucket_id", checker.bucketID)
		return nil, errors.New("failed to list gvg")
	}

	if buildMetaByDB {
		bucketMigrateUnits, err = checker.generateMigrateBucketUnitsFromDB(primarySPGVGList)
	} else {
		bucketMigrateUnits, err = checker.generateMigrateBucketUnitsFromMemory(primarySPGVGList)
	}

	log.Infow("succeeded to generate migrate bucket units", "bucket_id", checker.bucketID, "bucket_migrate_units", bucketMigrateUnits, "error", err)
	return bucketMigrateUnits, err
}

func SendAndConfirmCompleteMigrateBucketTx(baseApp *gfspapp.GfSpBaseApp, msg *storagetypes.MsgCompleteMigrateBucket) error {
	return SendAndConfirmTx(baseApp.Consensus(),
		func() (string, error) {
			var (
				txHash string
				txErr  error
			)
			if txHash, txErr = baseApp.GfSpClient().CompleteMigrateBucket(context.Background(), msg); txErr != nil && !isAlreadyExists(txErr) {
				log.Errorw("failed to send complete migrate bucket msg to chain", "complete_migrate_bucket_msg", msg, "error", txErr)
				return "", txErr
			}
			return txHash, nil
		})
}

func UpdateBucketMigrationProgress(baseApp *gfspapp.GfSpBaseApp, bucketID uint64, migrateState BucketMigrateState) error {
	if err := baseApp.GfSpDB().UpdateBucketMigrationProgress(bucketID, int(migrateState)); err != nil {
		log.Errorw("failed to update bucket migration progress", "bucket_id", bucketID, "state", migrateState, "error", err)
		return err
	}
	return nil
}
