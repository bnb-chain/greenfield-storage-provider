package executor

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-common/go/hash"
	storagetypes "github.com/evmos/evmos/v12/x/storage/types"
	virtualgrouptypes "github.com/evmos/evmos/v12/x/virtualgroup/types"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsptask"
	"github.com/zkMeLabs/mechain-storage-provider/core/piecestore"
	corespdb "github.com/zkMeLabs/mechain-storage-provider/core/spdb"
	coretask "github.com/zkMeLabs/mechain-storage-provider/core/task"
	metadatatypes "github.com/zkMeLabs/mechain-storage-provider/modular/metadata/types"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/metrics"
	"github.com/zkMeLabs/mechain-storage-provider/util"
)

const (
	queryLimit             = uint32(100)
	reportProgressPerN     = 10
	renewSigIntervalSecond = 60 * 60

	migrateGVGCostLabel              = "migrate_gvg_cost"
	migrateGVGSucceedCounterLabel    = "migrate_gvg_succeed_counter"
	migrateGVGFailedCounterLabel     = "migrate_gvg_failed_counter"
	migrateObjectCostLabel           = "migrate_object_cost"
	migrateObjectSucceedCounterLabel = "migrate_object_succeed_counter"
	migrateObjectFailedCounterLabel  = "migrate_object_failed_counter"
)

// HandleMigrateGVGTask handles migrate gvg task, including two cases: sp exit and bucket migration.
// srcSP is a sp who wants to exit or need to migrate bucket, destSP is used to accept data from srcSP.
func (e *ExecuteModular) HandleMigrateGVGTask(ctx context.Context, gvgTask coretask.MigrateGVGTask) {
	var (
		srcGvgID                  = gvgTask.GetSrcGvg().GetId()
		bucketID                  = gvgTask.GetBucketID()
		lastMigratedObjectID      = gvgTask.GetLastMigratedObjectID()
		objectList                []*metadatatypes.ObjectDetails
		err                       error
		migratedObjectNumberInGVG = 0
		startMigrateGVGTime       = time.Now()
	)

	defer func() {
		metrics.MigrateGVGTimeHistogram.WithLabelValues(migrateGVGCostLabel).Observe(time.Since(startMigrateGVGTime).Seconds())
		if err == nil {
			metrics.MigrateGVGCounter.WithLabelValues(migrateGVGSucceedCounterLabel).Inc()
		} else {
			metrics.MigrateGVGCounter.WithLabelValues(migrateGVGFailedCounterLabel).Inc()
		}
		gvgTask.SetError(err)
		log.CtxInfow(ctx, "finished to migrate gvg task", "gvg_id", srcGvgID, "bucket_id", bucketID,
			"total_migrated_object_number", migratedObjectNumberInGVG, "last_migrated_object_id", lastMigratedObjectID, "error", err)
	}()

	if gvgTask == nil {
		err = ErrDanglingPointer
		return
	}

	for {
		if bucketID == 0 { // sp exit task
			objectList, err = e.baseApp.GfSpClient().ListObjectsInGVG(ctx, srcGvgID, lastMigratedObjectID, queryLimit)
		} else { // bucket migrate task
			objectList, err = e.baseApp.GfSpClient().ListObjectsInGVGAndBucket(ctx, srcGvgID, bucketID, lastMigratedObjectID, queryLimit)
		}
		if err != nil {
			log.Errorw("migrate gvg task", "gvg_id", srcGvgID, "bucket_id", bucketID,
				"last_migrated_object_id", lastMigratedObjectID, "object_list", objectList, "error", err)
			return
		}

		for index, object := range objectList {
			if err = e.checkAndTryRenewSig(gvgTask.(*gfsptask.GfSpMigrateGVGTask)); err != nil {
				log.CtxErrorw(ctx, "failed to check and renew gvg task signature", "gvg_task", gvgTask, "error", err)
				return
			}

			if err = e.doObjectMigrationRetry(ctx, gvgTask, bucketID, object); err != nil {
				log.CtxErrorw(ctx, "failed to do object migration", "gvg_task", gvgTask, "object", object, "error", err)
				return
			}

			if (index+1)%reportProgressPerN == 0 || index == len(objectList)-1 {
				log.Infow("migrate gvg report task", "gvg_id", srcGvgID, "bucket_id", bucketID,
					"current_migrated_object_number", migratedObjectNumberInGVG,
					"last_migrated_object_id", object.GetObject().GetObjectInfo().Id.Uint64())
				gvgTask.SetLastMigratedObjectID(object.GetObject().GetObjectInfo().Id.Uint64())
				if err = e.ReportTask(ctx, gvgTask); err != nil { // report task is already automatically triggered.
					log.CtxErrorw(ctx, "failed to report migrate gvg task progress", "task_info", gvgTask, "error", err)
					return
				}
			}
			lastMigratedObjectID = object.GetObject().GetObjectInfo().Id.Uint64()
			migratedObjectNumberInGVG++
			gvgTask.SetMigratedBytesSize(gvgTask.GetMigratedBytesSize() + object.GetObject().GetObjectInfo().GetPayloadSize())
		}
		if len(objectList) < int(queryLimit) { // it indicates that gvg all objects have been migrated.
			gvgTask.SetLastMigratedObjectID(lastMigratedObjectID)
			gvgTask.SetFinished(true)
			return
		}
	}
}

func (e *ExecuteModular) checkAndTryRenewSig(gvgTask *gfsptask.GfSpMigrateGVGTask) error {
	var (
		signature []byte
		err       error
	)
	if time.Now().Unix()+renewSigIntervalSecond/10 > gvgTask.ExpireTime {
		originExpireTime := gvgTask.ExpireTime
		gvgTask.ExpireTime = time.Now().Unix() + renewSigIntervalSecond
		signature, err = e.baseApp.GfSpClient().SignMigrateGVG(context.Background(), gvgTask)
		if err != nil {
			gvgTask.ExpireTime = originExpireTime       // revert to origin expire time
			if time.Now().Unix() > gvgTask.ExpireTime { // the signature is indeed expired
				log.Errorw("failed to sign migrate gvg", "gvg_task", gvgTask, "error", err)
				return err
			}
		} else {
			gvgTask.SetSignature(signature)
		}
	}
	return nil
}

func (e *ExecuteModular) doObjectMigration(ctx context.Context, gvgTask coretask.MigrateGVGTask, bucketID uint64,
	objectDetails *metadatatypes.ObjectDetails,
) error {
	var (
		err                    error
		isBucketMigrate        bool
		startMigrateObjectTime = time.Now()
		object                 = objectDetails.GetObject()
	)

	defer func() {
		metrics.MigrateObjectTimeHistogram.WithLabelValues(migrateObjectCostLabel).Observe(time.Since(startMigrateObjectTime).Seconds())
		if err == nil {
			metrics.MigrateObjectCounter.WithLabelValues(migrateObjectSucceedCounterLabel).Inc()
		} else {
			metrics.MigrateObjectCounter.WithLabelValues(migrateObjectFailedCounterLabel).Inc()
		}
		log.CtxDebugw(ctx, "finish to migrate object", "task_info", gvgTask, "bucket_id", bucketID, "object_info", objectDetails, "error", err)
	}()

	params, err := e.baseApp.Consensus().QueryStorageParamsByTimestamp(ctx, object.GetObjectInfo().GetCreateAt())
	if err != nil {
		log.CtxErrorw(ctx, "failed to query storage params by timestamp", "object_id",
			object.GetObjectInfo().Id.String(), "object_name", object.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}

	if bucketID != 0 {
		// bucket migration, check secondary whether is conflict, if true replicate own secondary SP data to another secondary SP
		if err = e.checkGVGConflict(ctx, gvgTask.GetSrcGvg(), gvgTask.GetDestGvg(), object.GetObjectInfo(), params); err != nil {
			log.Debugw("failed to resolve gvg conflict", "error", err, "task", gvgTask, "object", object.GetObjectInfo())
			return err
		}
		isBucketMigrate = true
	}

	selfSpID := gvgTask.GetSrcSp().GetId()
	redundancyIdx, isSecondary := util.ValidateSecondarySPs(selfSpID, objectDetails.GetGvg().GetSecondarySpIds())
	isPrimary := util.ValidatePrimarySP(selfSpID, objectDetails.GetGvg().GetPrimarySpId())
	if !isPrimary && !isSecondary {
		return fmt.Errorf("invalid sp id: %d", selfSpID)
	}
	migratePieceTask := &gfsptask.GfSpMigratePieceTask{
		ObjectInfo:      object.GetObjectInfo(),
		StorageParams:   params,
		SrcSpEndpoint:   gvgTask.GetSrcSp().GetEndpoint(),
		IsBucketMigrate: isBucketMigrate,
	}
	if !isSecondary && isPrimary {
		migratePieceTask.RedundancyIdx = piecestore.PrimarySPRedundancyIndex
	} else {
		migratePieceTask.RedundancyIdx = int32(redundancyIdx)
	}
	if err = e.HandleMigratePieceTask(ctx, gvgTask.(*gfsptask.GfSpMigrateGVGTask), migratePieceTask); err != nil {
		log.CtxErrorw(ctx, "failed to migrate object pieces", "object_id", object.GetObjectInfo().Id.String(),
			"object_name", object.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}
	return err
}

func (e *ExecuteModular) doObjectMigrationRetry(ctx context.Context, gvgTask coretask.MigrateGVGTask, bucketID uint64, object *metadatatypes.ObjectDetails) error {
	var (
		srcGvgID = gvgTask.GetSrcGvg().GetId()
		err      error
	)
	for retry := 0; retry < e.maxObjectMigrationRetry; retry++ {
		// when cancel migrate bucket, the dest sp event may be slower than src sp, so we retry this migration
		if err = e.doObjectMigration(ctx, gvgTask, bucketID, object); err != nil {
			// 1) error happens, but will skip error
			if e.isSkipFailedToMigrateObject(ctx, object) {
				log.CtxErrorw(ctx, "failed to do migration gvg task and the error will skip", "gvg_id", srcGvgID,
					"bucket_id", bucketID, "object_info", object,
					"enable_skip_failed_to_migrate_object", e.enableSkipFailedToMigrateObject, "retry", retry, "error", err)
				return nil
			} else {
				// 2) error happens, sleep and will retry
				log.CtxErrorw(ctx, "failed to do migration gvg task", "gvg_id", srcGvgID,
					"bucket_id", bucketID, "object_info", object,
					"enable_skip_failed_to_migrate_object", e.enableSkipFailedToMigrateObject, "retry", retry, "error", err)
				time.Sleep(time.Duration(e.objectMigrationRetryTimeout) * time.Second)
				continue
			}
		} else {
			// 3) no error case
			return nil
		}
	}
	return err
}

func (e *ExecuteModular) checkGVGConflict(ctx context.Context, srcGvg, destGvg *virtualgrouptypes.GlobalVirtualGroup,
	objectInfo *storagetypes.ObjectInfo, params *storagetypes.Params,
) error {
	index := util.ContainOnlyOneDifferentElement(srcGvg.GetSecondarySpIds(), destGvg.GetSecondarySpIds())
	if index == piecestore.PrimarySPRedundancyIndex {
		// no conflict
		return nil
	}

	spID, err := e.getSPID()
	if err != nil {
		return err
	}
	if spID != srcGvg.GetSecondarySpIds()[index] {
		log.Debugw("invalid secondary sp id in src gvg", "index", index, "sp_id", e.spID, "src_gvg", srcGvg, "dest_gvg", destGvg)
		return fmt.Errorf("invalid secondary sp id in src gvg")
	}
	destSecondarySPID := destGvg.GetSecondarySpIds()[index]
	spInfo, err := e.baseApp.Consensus().QuerySPByID(ctx, destSecondarySPID)
	if err != nil {
		log.Errorw("failed to query dest sp info", "dest_sp_id", destSecondarySPID, "error", err)
		return err
	}

	var (
		segmentCount = e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
			params.VersionedParams.GetMaxSegmentSize())
		pieceKey string
	)
	for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
		if objectInfo.GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
			pieceKey = e.baseApp.PieceOp().ECPieceKey(objectInfo.Id.Uint64(), segIdx, uint32(index), objectInfo.Version)
		} else {
			pieceKey = e.baseApp.PieceOp().SegmentPieceKey(objectInfo.Id.Uint64(), segIdx, objectInfo.Version)
		}
		pieceData, err := e.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get piece data from piece store", "error", err, "piece_key", pieceKey)
			return err
		}
		err = e.doBucketMigrationReplicatePiece(ctx, destGvg.GetId(), objectInfo, params, spInfo.GetEndpoint(), segIdx, uint32(index), pieceData)
		if err != nil {
			log.CtxErrorw(ctx, "failed to do bucket migration to replicate pieces", "error", err, "piece_key", pieceKey)
			return err
		}
	}

	err = e.doneBucketMigrationReplicatePiece(ctx, destGvg.GetId(), objectInfo, params, spInfo.GetEndpoint(), uint32(index))
	if err != nil {
		log.CtxErrorw(ctx, "failed to done bucket migration replicate piece", "error", err, "piece_key", pieceKey)
		return err
	}

	log.Debugw("bucket migration replicates piece", "dest_sp_endpoint", spInfo.GetEndpoint())
	return nil
}

func (e *ExecuteModular) doBucketMigrationReplicatePiece(ctx context.Context, gvgID uint32, objectInfo *storagetypes.ObjectInfo,
	params *storagetypes.Params, destSPEndpoint string, segmentIdx, redundancyIdx uint32, data []byte,
) error {
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(gvgID, objectInfo, params, coretask.DefaultSmallerPriority, segmentIdx,
		int32(redundancyIdx), int64(len(data)), false)
	receive.SetPieceChecksum(hash.GenerateChecksum(data))
	receive.SetBucketMigration(true)
	ctx = log.WithValue(ctx, log.CtxKeyTask, receive.Key().String())
	signature, err := e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign receive task", "segment_piece_index", segmentIdx,
			"redundancy_index", redundancyIdx, "error", err)
		return err
	}
	if int(redundancyIdx+1) >= len(objectInfo.GetChecksums()) {
		log.CtxErrorw(ctx, "failed to done replicate piece, replicate idx out of bounds", "redundancy_index", redundancyIdx)
		return ErrReplicateIdsOutOfBounds
	}
	receive.SetSignature(signature)
	if err = e.baseApp.GfSpClient().ReplicatePieceToSecondary(ctx, destSPEndpoint, receive, data); err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "segment_piece_index", segmentIdx,
			"redundancy_index", redundancyIdx, "error", err)
		return err
	}
	return nil
}

func (e *ExecuteModular) doneBucketMigrationReplicatePiece(ctx context.Context, gvgID uint32, objectInfo *storagetypes.ObjectInfo,
	params *storagetypes.Params, destSPEndpoint string, redundancyIdx uint32,
) error {
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(gvgID, objectInfo, params, coretask.DefaultSmallerPriority, 0 /* useless */, int32(redundancyIdx), 0, false)
	taskSignature, err := e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign done receive task", "error", err, "redundancy_index", redundancyIdx)
		return err
	}
	receive.SetSignature(taskSignature)
	receive.SetBucketMigration(true)
	receive.SetFinished(true)
	_, err = e.baseApp.GfSpClient().DoneReplicatePieceToSecondary(ctx, destSPEndpoint, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece", "dest_sp_endpoint", destSPEndpoint, "redundancy_index", redundancyIdx, "error", err)
		return err
	}

	log.CtxDebugw(ctx, "succeed to done replicate", "dest_sp_endpoint", destSPEndpoint, "redundancy_index", redundancyIdx)
	return nil
}

// HandleMigratePieceTask handles migrate piece task
// It will send requests to the src SP(exiting SP or bucket migration) to get piece data. Using piece data to generate
// piece checksum and integrity hash, if integrity hash is similar to chain's, piece data would be written into PieceStore,
// generated piece checksum and integrity hash will be written into sql db.
//
// storagetypes.ObjectInfo struct contains LocalVirtualGroupId field which we can use it to get a GVG consisting of
// one PrimarySP and six ordered secondarySP(the order cannot be changed). Therefore, we can know what kinds of object
// we want to migrate: primary or secondary. Now we cannot use objectInfo operator address or secondaryAddress straightly.
// We should encapsulate a new method to get.
// objectInfo->lvg->gvg->(1 primarySP, 6 secondarySPs)
func (e *ExecuteModular) HandleMigratePieceTask(ctx context.Context, gvgTask *gfsptask.GfSpMigrateGVGTask, pieceTask *gfsptask.GfSpMigratePieceTask) error {
	var (
		segmentCount = e.baseApp.PieceOp().SegmentPieceCount(pieceTask.GetObjectInfo().GetPayloadSize(),
			pieceTask.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		redundancyIdx = pieceTask.GetRedundancyIdx()
		objectID      = pieceTask.GetObjectInfo().Id.Uint64()
		objectVersion = pieceTask.GetObjectInfo().GetVersion()
	)

	if pieceTask == nil {
		return ErrDanglingPointer
	}

	// get a piece of data, calculate a piece checksum once, store it into db and piece store
	// finally verify the integrity hash, and incorrect objects are deleted by gc
	for i := 0; i < int(segmentCount); i++ {
		pieceTask.SegmentIdx = uint32(i)
		pieceData, err := e.sendRequest(ctx, gvgTask, pieceTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to migrate piece data", "object_id", objectID,
				"object_name", pieceTask.GetObjectInfo().GetObjectName(), "segment_piece_index", pieceTask.GetSegmentIdx(), "sp_endpoint",
				pieceTask.GetSrcSpEndpoint(), "error", err)
			return err
		}

		var pieceKey string
		if redundancyIdx == piecestore.PrimarySPRedundancyIndex {
			pieceKey = e.baseApp.PieceOp().SegmentPieceKey(objectID, uint32(i), objectVersion)
		} else {
			pieceKey = e.baseApp.PieceOp().ECPieceKey(objectID, uint32(i), uint32(redundancyIdx), objectVersion)
		}
		if err = e.baseApp.PieceStore().PutPiece(ctx, pieceKey, pieceData); err != nil {
			log.CtxErrorw(ctx, "failed to put piece data into primary sp", "piece_key", pieceKey, "error", err)
			return ErrPieceStoreWithDetail("failed to put piece data into primary sp, piece_key: " + pieceKey + ",error: " + err.Error())
		}

		pieceChecksum := hash.GenerateChecksum(pieceData)
		if err = e.baseApp.GfSpDB().SetReplicatePieceChecksum(objectID, uint32(i), redundancyIdx, pieceChecksum, objectVersion); err != nil {
			log.CtxErrorw(ctx, "failed to set replicate piece checksum", "object_id", pieceTask.GetObjectInfo().Id.Uint64(),
				"segment_index", i, "redundancy_index", redundancyIdx, "error", err)
			detail := fmt.Sprintf("failed to set replicate piece checksum, object_id: %s, segment_index: %v, redundancy_index: %v, error: %s",
				pieceTask.GetObjectInfo().Id.String(), i, redundancyIdx, err.Error())
			return ErrGfSpDBWithDetail(detail)
		}
	}

	if err := e.setMigratePiecesMetadata(pieceTask.GetObjectInfo(), segmentCount, redundancyIdx); err != nil {
		log.Errorw("failed to set object integrity meta", "error", err)
		return err
	}
	return nil
}

func (e *ExecuteModular) sendRequest(ctx context.Context, gvgTask *gfsptask.GfSpMigrateGVGTask, pieceTask *gfsptask.GfSpMigratePieceTask) ([]byte, error) {
	var (
		pieceData []byte
		err       error
	)
	pieceData, err = e.baseApp.GfSpClient().MigratePiece(ctx, gvgTask, pieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to migrate piece data", "object_id",
			pieceTask.GetObjectInfo().Id.Uint64(), "object_name", pieceTask.GetObjectInfo().GetObjectName(), "task_info", pieceTask, "error", err)
		return nil, err
	}
	log.CtxInfow(ctx, "succeed to get piece from another sp", "object_id", pieceTask.GetObjectInfo().Id.Uint64(),
		"object_name", pieceTask.GetObjectInfo().GetObjectName(), "segment_piece_index", pieceTask.GetSegmentIdx(), "redundancy_index", pieceTask.GetRedundancyIdx())
	return pieceData, nil
}

// setMigratePiecesMetadata writes integrity hash into db
// 1. get piece checksum list from db and generate integrity hash
// 2. compare generated integrity hash to chain integrity hash, if they are equal write to db
func (e *ExecuteModular) setMigratePiecesMetadata(objectInfo *storagetypes.ObjectInfo, segmentCount uint32, redundancyIdx int32) error {
	objectID := objectInfo.Id.Uint64()
	pieceChecksums, err := e.baseApp.GfSpDB().GetAllReplicatePieceChecksumOptimized(objectID, redundancyIdx, segmentCount)
	if err != nil {
		log.Errorw("failed to get checksum from db", "object_info", objectInfo, "error", err)
		return ErrGfSpDBWithDetail("failed to get checksum from db, object_info: " + objectInfo.String() + ",error: " + err.Error())
	}
	if len(pieceChecksums) != int(segmentCount) {
		log.Errorw("returned piece checksum length does not match segment count",
			"expected_segment_number", segmentCount, "actual_segment_number", len(pieceChecksums))
		return ErrInvalidPieceChecksumLength
	}
	migratedIntegrityHash := hash.GenerateIntegrityHash(pieceChecksums)

	var chainIntegrityHash []byte
	if redundancyIdx == piecestore.PrimarySPRedundancyIndex {
		// primarySP
		chainIntegrityHash = objectInfo.GetChecksums()[0]
	} else {
		// secondarySP
		chainIntegrityHash = objectInfo.GetChecksums()[redundancyIdx+1]
	}
	if !bytes.Equal(migratedIntegrityHash, chainIntegrityHash) {
		log.Errorw("migrated pieces integrity is different from integrity hash on chain", "object_info",
			objectInfo, "expected_checksum", chainIntegrityHash, "actual_checksum", migratedIntegrityHash, "redundancy_index", redundancyIdx)
		return ErrMigratedPieceChecksum
	}

	if err = e.baseApp.GfSpDB().SetObjectIntegrity(&corespdb.IntegrityMeta{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIdx,
		IntegrityChecksum: migratedIntegrityHash,
		PieceChecksumList: pieceChecksums,
	}); err != nil {
		log.Errorw("failed to set object integrity into sp db", "object_id", objectID,
			"object_name", objectInfo.GetObjectName(), "error", err)
		return ErrSetObjectIntegrity
	}
	if err = e.baseApp.GfSpDB().DeleteAllReplicatePieceChecksum(objectID, redundancyIdx, segmentCount); err != nil {
		log.Errorw("failed to delete all migrated piece checksum", "error", err)
	}
	log.Infow("succeed to compute and set object integrity", "object_id", objectID,
		"object_name", objectInfo.GetObjectName())
	return nil
}

// isSkipFailedToMigrateObject incorrect migration can be an expected error, errors will be ignored.
// such as: 1) deletion of objects during migration, or 2) the enableSkipFailedToMigrateObject parameter being set.
func (e *ExecuteModular) isSkipFailedToMigrateObject(ctx context.Context, objectDetails *metadatatypes.ObjectDetails) bool {
	if e.enableSkipFailedToMigrateObject {
		return true
	}
	// if the object do not exist on chain, should ignore the error
	objectInfo, err := e.baseApp.Consensus().QueryObjectInfo(ctx, objectDetails.GetObject().GetObjectInfo().GetBucketName(), objectDetails.GetObject().GetObjectInfo().GetObjectName())
	if err != nil {
		if strings.Contains(err.Error(), "No such object") {
			log.CtxErrorw(ctx, "failed to get object info from consensus, the object may be deleted", "object", objectInfo, "error", err)
			return true
		}
		return false
	}

	return false
}
