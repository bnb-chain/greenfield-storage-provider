package executor

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const (
	queryLimit         = uint32(100)
	reportProgressPerN = 10

	migrateGVGCostLabel              = "migrate_gvg_cost"
	migrateGVGSucceedCounterLabel    = "migrate_gvg_succeed_counter"
	migrateGVGFailedCounterLabel     = "migrate_gvg_failed_counter"
	migrateObjectCostLabel           = "migrate_object_cost"
	migrateObjectSucceedCounterLabel = "migrate_object_succeed_counter"
	migrateObjectFailedCounterLabel  = "migrate_object_failed_counter"
)

// HandleMigrateGVGTask handles migrate gvg task, including two cases: sp exit and bucket migration.
// srcSP is a sp who wants to exit or need to migrate bucket, destSP is used to accept data from srcSP.
func (e *ExecuteModular) HandleMigrateGVGTask(ctx context.Context, task coretask.MigrateGVGTask) {
	var (
		srcGvgID                  = task.GetSrcGvg().GetId()
		bucketID                  = task.GetBucketID()
		lastMigratedObjectID      = task.GetLastMigratedObjectID()
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
		log.CtxInfow(ctx, "finished to migrate gvg task", "gvg_id", srcGvgID, "bucket_id", bucketID,
			"total_migrated_object_number", migratedObjectNumberInGVG, "last_migrated_object_id", lastMigratedObjectID, "error", err)
	}()

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
			if err = e.doObjectMigration(ctx, task, bucketID, object); err != nil && !e.enableSkipFailedToMigrateObject {
				log.CtxErrorw(ctx, "failed to do migration gvg task", "gvg_id", srcGvgID,
					"bucket_id", bucketID, "object_info", object,
					"enable_skip_failed_to_migrate_object", e.enableSkipFailedToMigrateObject, "error", err)
				return
			}
			if (index+1)%reportProgressPerN == 0 || index == len(objectList)-1 {
				log.Infow("migrate gvg report task", "gvg_id", srcGvgID, "bucket_id", bucketID,
					"current_migrated_object_number", migratedObjectNumberInGVG,
					"last_migrated_object_id", object.GetObject().GetObjectInfo().Id.Uint64())
				task.SetLastMigratedObjectID(object.GetObject().GetObjectInfo().Id.Uint64())
				if err = e.ReportTask(ctx, task); err != nil {
					log.CtxErrorw(ctx, "failed to report migrate gvg task progress", "task_info", task, "error", err)
					return
				}
			}
			lastMigratedObjectID = object.GetObject().GetObjectInfo().Id.Uint64()
			migratedObjectNumberInGVG++
		}
		if len(objectList) < int(queryLimit) { // it indicates that gvg all objects have been migrated.
			task.SetLastMigratedObjectID(lastMigratedObjectID)
			task.SetFinished(true)
			return
		}
	}
}

func (e *ExecuteModular) doObjectMigration(ctx context.Context, task coretask.MigrateGVGTask, bucketID uint64,
	objectDetails *metadatatypes.ObjectDetails) error {
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
		log.CtxDebugw(ctx, "finish to migrate object", "task_info", task, "bucket_id", bucketID, "object_info", objectDetails, "error", err)
	}()

	params, err := e.baseApp.Consensus().QueryStorageParamsByTimestamp(ctx, object.GetObjectInfo().GetCreateAt())
	if err != nil {
		log.CtxErrorw(ctx, "failed to query storage params by timestamp", "object_id",
			object.GetObjectInfo().Id.String(), "object_name", object.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}

	if bucketID != 0 {
		// bucket migration, check secondary whether is conflict, if true replicate own secondary SP data to another secondary SP
		if err = e.checkGVGConflict(ctx, task.GetSrcGvg(), task.GetDestGvg(), object.GetObjectInfo(), params); err != nil {
			log.Debugw("no gvg conflict", "error", err)
		}
		isBucketMigrate = true
	}

	selfSpID := task.GetSrcSp().GetId()
	redundancyIdx, isSecondary := util.ValidateSecondarySPs(selfSpID, objectDetails.GetGvg().GetSecondarySpIds())
	isPrimary := util.ValidatePrimarySP(selfSpID, objectDetails.GetGvg().GetPrimarySpId())
	if !isPrimary && !isSecondary {
		return fmt.Errorf("invalid sp id: %d", selfSpID)
	}
	migratePieceTask := &gfsptask.GfSpMigratePieceTask{
		ObjectInfo:      object.GetObjectInfo(),
		StorageParams:   params,
		SrcSpEndpoint:   task.GetSrcSp().GetEndpoint(),
		IsBucketMigrate: isBucketMigrate,
	}
	if !isSecondary && isPrimary {
		migratePieceTask.RedundancyIdx = piecestore.PrimarySPRedundancyIndex
	} else {
		migratePieceTask.RedundancyIdx = int32(redundancyIdx)
	}
	if err = e.HandleMigratePieceTask(ctx, migratePieceTask); err != nil {
		log.CtxErrorw(ctx, "failed to migrate object pieces", "object_id", object.GetObjectInfo().Id.String(),
			"object_name", object.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}
	return nil
}

func (e *ExecuteModular) checkGVGConflict(ctx context.Context, srcGvg, destGvg *virtualgrouptypes.GlobalVirtualGroup,
	objectInfo *storagetypes.ObjectInfo, params *storagetypes.Params) error {
	index := util.ContainOnlyOneDifferentElement(srcGvg.GetSecondarySpIds(), destGvg.GetSecondarySpIds())
	if index == -1 {
		return fmt.Errorf("invalid gvg secondary sp id list")
	}
	if e.spID != srcGvg.GetSecondarySpIds()[index] {
		return fmt.Errorf("invalid secondary sp id in src gvg")
	}
	destSecondarySPID := destGvg.GetSecondarySpIds()[index]
	spInfo, err := e.baseApp.Consensus().QuerySPByID(ctx, destSecondarySPID)
	if err != nil {
		log.Errorw("failed to query dest sp info", "dest_sp_id", destSecondarySPID, "error", err)
	}

	var (
		segmentCount = e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
			params.VersionedParams.GetMaxSegmentSize())
		pieceKey string
	)
	for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
		if objectInfo.GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
			pieceKey = e.baseApp.PieceOp().ECPieceKey(objectInfo.Id.Uint64(), segIdx, uint32(index))
		} else {
			pieceKey = e.baseApp.PieceOp().SegmentPieceKey(objectInfo.Id.Uint64(), segIdx)
		}
		pieceData, err := e.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get piece data from piece store", "error", err)
			return err
		}
		err = e.doBucketMigrationReplicatePiece(ctx, destGvg.GetId(), objectInfo, params, spInfo.GetEndpoint(), segIdx, uint32(index), pieceData)
		if err != nil {
			log.CtxErrorw(ctx, "failed to do bucket migration to replicate pieces", "error", err)
			return err
		}
	}

	err = e.doneBucketMigrationReplicatePiece(ctx, destGvg.GetId(), objectInfo, params, spInfo.GetEndpoint(), uint32(index))
	if err != nil {
		log.CtxErrorw(ctx, "failed to done bucket migration replicate piece", "error", err)
		return err
	}

	log.Debugw("bucket migration replicates piece", "dest_sp_endpoint", spInfo.GetEndpoint())
	return nil
}

func (e *ExecuteModular) doBucketMigrationReplicatePiece(ctx context.Context, gvgID uint32, objectInfo *storagetypes.ObjectInfo,
	params *storagetypes.Params, destSPEndpoint string, segmentIdx, redundancyIdx uint32, data []byte) error {
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(gvgID, objectInfo, params, coretask.DefaultSmallerPriority, segmentIdx,
		int32(redundancyIdx), int64(len(data)))
	receive.SetPieceChecksum(hash.GenerateChecksum(data))
	ctx = log.WithValue(ctx, log.CtxKeyTask, receive.Key().String())
	signature, err := e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign receive task", "segment_piece_index", segmentIdx,
			"redundancy_index", redundancyIdx, "error", err)
		return err
	}
	receive.SetSignature(signature)
	if err = e.baseApp.GfSpClient().ReplicatePieceToSecondary(ctx, destSPEndpoint, receive, data); err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "segment_piece_index", segmentIdx,
			"redundancy_index", redundancyIdx, "error", err)
	}
	return nil
}

func (e *ExecuteModular) doneBucketMigrationReplicatePiece(ctx context.Context, gvgID uint32, objectInfo *storagetypes.ObjectInfo,
	params *storagetypes.Params, destSPEndpoint string, segmentIdx uint32) error {
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(gvgID, objectInfo, params, coretask.DefaultSmallerPriority, segmentIdx, -1, 0)
	taskSignature, err := e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign done receive task", "segment_piece_index", segmentIdx, "error", err)
		return err
	}
	receive.SetSignature(taskSignature)
	receive.SetBucketMigration(true)
	_, err = e.baseApp.GfSpClient().DoneReplicatePieceToSecondary(ctx, destSPEndpoint, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece", "dest_sp_endpoint", destSPEndpoint,
			"segment_idx", segmentIdx, "error", err)
		return err
	}
	if int(segmentIdx+1) >= len(objectInfo.GetChecksums()) {
		log.CtxErrorw(ctx, "failed to done replicate piece, replicate idx out of bounds", "segment_piece_index", segmentIdx)
		return ErrReplicateIdsOutOfBounds
	}
	log.CtxDebugw(ctx, "succeed to done replicate", "dest_sp_endpoint", destSPEndpoint, "segment_piece_index", segmentIdx)
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
func (e *ExecuteModular) HandleMigratePieceTask(ctx context.Context, task *gfsptask.GfSpMigratePieceTask) error {
	var (
		segmentCount = e.baseApp.PieceOp().SegmentPieceCount(task.GetObjectInfo().GetPayloadSize(),
			task.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		redundancyIdx = task.GetRedundancyIdx()
		objectID      = task.GetObjectInfo().Id.Uint64()
	)

	if task == nil {
		return ErrDanglingPointer
	}

	// get a piece of data, calculate a piece checksum once, store it into db and piece store
	// finally verify the integrity hash, and incorrect objects are deleted by gc
	for i := 0; i < int(segmentCount); i++ {
		task.SegmentIdx = uint32(i)
		pieceData, err := e.sendRequest(ctx, task)
		if err != nil {
			log.CtxErrorw(ctx, "failed to migrate piece data", "object_id", objectID,
				"object_name", task.GetObjectInfo().GetObjectName(), "segment_piece_index", task.GetSegmentIdx(), "sp_endpoint",
				task.GetSrcSpEndpoint(), "error", err)
			return err
		}

		var pieceKey string
		if redundancyIdx == piecestore.PrimarySPRedundancyIndex {
			pieceKey = e.baseApp.PieceOp().SegmentPieceKey(objectID, uint32(i))
		} else {
			pieceKey = e.baseApp.PieceOp().ECPieceKey(objectID, uint32(i), uint32(redundancyIdx))
		}
		if err = e.baseApp.PieceStore().PutPiece(ctx, pieceKey, pieceData); err != nil {
			log.CtxErrorw(ctx, "failed to put piece data into primary sp", "piece_key", pieceKey, "error", err)
			return ErrPieceStore
		}

		pieceChecksum := hash.GenerateChecksum(pieceData)
		if err = e.baseApp.GfSpDB().SetReplicatePieceChecksum(objectID, uint32(i), redundancyIdx, pieceChecksum); err != nil {
			log.CtxErrorw(ctx, "failed to set replicate piece checksum", "object_id", task.GetObjectInfo().Id.Uint64(),
				"segment_index", i, "redundancy_index", redundancyIdx, "error", err)
			return ErrGfSpDB
		}
	}

	if err := e.setMigratePiecesMetadata(task.GetObjectInfo(), segmentCount, redundancyIdx); err != nil {
		log.Errorw("failed to set object integrity meta", "error", err)
		return err
	}
	return nil
}

func (e *ExecuteModular) sendRequest(ctx context.Context, task *gfsptask.GfSpMigratePieceTask) ([]byte, error) {
	var (
		pieceData []byte
		err       error
	)
	pieceData, err = e.baseApp.GfSpClient().MigratePiece(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to migrate piece data", "object_id",
			task.GetObjectInfo().Id.Uint64(), "object_name", task.GetObjectInfo().GetObjectName(), "task_info", task, "error", err)
		return nil, err
	}
	log.CtxInfow(ctx, "succeed to get piece from another sp", "object_id", task.GetObjectInfo().Id.Uint64(),
		"object_name", task.GetObjectInfo().GetObjectName(), "segment_piece_index", task.GetSegmentIdx(), "redundancy_index", task.GetRedundancyIdx())
	return pieceData, nil
}

// setMigratePiecesMetadata writes integrity hash into db
// 1. get piece checksum list from db and generate integrity hash
// 2. compare generated integrity hash to chain integrity hash, if they are equal write to db
func (e *ExecuteModular) setMigratePiecesMetadata(objectInfo *storagetypes.ObjectInfo, segmentCount uint32, redundancyIdx int32) error {
	var (
		objectID = objectInfo.Id.Uint64()
	)
	pieceChecksums, err := e.baseApp.GfSpDB().GetAllReplicatePieceChecksum(objectID, redundancyIdx, segmentCount)
	if err != nil {
		log.Errorw("failed to get checksum from db", "object_info", objectInfo, "error", err)
		return ErrGfSpDB
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
