package executor

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const (
	primarySPRedundancyIdx = -1
	queryLimit             = 100
)

// HandleMigrateGVGTask handles the migrate gvg task.
// There are two cases: sp exit and bucket migration
// srcSP is a sp who wants to exit or need to migrate bucket, destSP is used to accept data from srcSP
func (e *ExecuteModular) HandleMigrateGVGTask(ctx context.Context, task coretask.MigrateGVGTask) {
	var (
		oldGvgID             = task.GetSrcGvg().GetId()
		bucketID             = task.GetBucketID()
		lastMigratedObjectID = task.GetLastMigratedObjectID()
		objectList           []*metadatatypes.Object
		err                  error
	)

	for {
		if bucketID == 0 {
			// if bucketID is 0, it indicates that it is a sp exiting task
			objectList, err = e.baseApp.GfSpClient().ListObjectsInGVG(ctx, oldGvgID, lastMigratedObjectID+1, queryLimit)
			if err != nil {
				log.CtxErrorw(ctx, "failed to list objects in gvg", "old_gvg_id", oldGvgID,
					"current_migrated_object_id", lastMigratedObjectID+1, "error", err)
			}
			log.Infow("sp exiting", "objectList", objectList)
		} else {
			// if bucketID is not equal to 0, it indicates that it is a bucket migration task
			objectList, err = e.baseApp.GfSpClient().ListObjectsInGVGAndBucket(ctx, oldGvgID, bucketID, lastMigratedObjectID, queryLimit)
			if err != nil {
				log.CtxErrorw(ctx, "failed to list objects in gvg and bucket", "old_gvg_id", oldGvgID,
					"bucket_id", bucketID, "current_migrated_object_id", lastMigratedObjectID, "error", err)
			}
			log.CtxDebugw(ctx, "succeed to list objects in gvg and bucket", "old_gvg_id", oldGvgID,
				"bucket_id", bucketID, "current_migrated_object_id", lastMigratedObjectID, "error", err)
			log.Infow("migrate bucket", "objectList", objectList)
		}
		for index, object := range objectList {
			if object.GetRemoved() || object.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
				log.CtxInfow(ctx, "object has been removed or object status is not sealed", "removed",
					object.GetRemoved(), "object_status", object.GetObjectInfo().GetObjectStatus(), "object_id",
					object.GetObjectInfo().Id.String(), "object_name", object.GetObjectInfo().GetObjectName())
				continue
			}
			if err = e.doMigrationGVGTask(ctx, task, object, bucketID); err != nil {
				log.CtxErrorw(ctx, "failed to do migration gvg task", "error", err)
				return
			}
			if index%10 == 0 {
				// report task per 10 objects
				if err = e.ReportTask(ctx, task); err != nil {
					log.CtxErrorw(ctx, "failed to report task", "index", index, "error", err)
				}
			}
		}
		if len(objectList) < queryLimit {
			// When the total count of objectList is less than 100, it indicates that this gvg has finished
			task.SetFinished(true)
			return
		}
	}
}

func (e *ExecuteModular) doMigrationGVGTask(ctx context.Context, task coretask.MigrateGVGTask, object *metadatatypes.Object,
	bucketID uint64) error {
	params, err := e.baseApp.Consensus().QueryStorageParamsByTimestamp(ctx, object.GetObjectInfo().GetCreateAt())
	if err != nil {
		log.CtxErrorw(ctx, "failed to query storage params by timestamp", "object_id",
			object.GetObjectInfo().Id.String(), "object_name", object.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}

	bucketName := object.GetObjectInfo().GetBucketName()
	bucketInfo, err := e.baseApp.Consensus().QueryBucketInfo(ctx, bucketName)
	if err != nil {
		log.CtxErrorw(ctx, "failed to query bucket info by bucket name", "bucketName", bucketName, "error", err)
		return err
	}
	// bucket migration, check secondary whether is conflict, if true replicate own secondary SP data to another SP
	if bucketID != 0 {
		if err = e.checkGvgConflict(ctx, task.GetSrcGvg(), task.GetDestGvg(), object.GetObjectInfo(), params); err != nil {
			log.Errorw("failed to check gvg conflict", "error", err)
			return err
		}
	}
	if bucketID != bucketInfo.Id.Uint64() {
		bucketID = bucketInfo.Id.Uint64()
	}

	redundancyIdx, isPrimary, err := util.ValidateAndGetSPIndexWithinGVGSecondarySPs(ctx, e.baseApp.GfSpClient(),
		task.GetSrcSp().GetId(), bucketID, object.GetObjectInfo().GetLocalVirtualGroupId())
	if err != nil {
		log.CtxErrorw(ctx, "failed to validate and get sp index within gvg secondary sps", "error", err)
		return err
	}
	migratePieceTask := &gfsptask.GfSpMigratePieceTask{
		ObjectInfo:    object.GetObjectInfo(),
		StorageParams: params,
		SrcSpEndpoint: task.GetSrcSp().GetEndpoint(),
	}
	if redundancyIdx == primarySPRedundancyIdx && !isPrimary {
		migratePieceTask.RedundancyIdx = primarySPRedundancyIdx
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

func (e *ExecuteModular) checkGvgConflict(ctx context.Context, srcGvg, destGvg *virtualgrouptypes.GlobalVirtualGroup,
	objectInfo *storagetypes.ObjectInfo, params *storagetypes.Params) error {
	index := util.ContainOnlyOneDifferentElement(srcGvg.GetSecondarySpIds(), destGvg.GetSecondarySpIds())
	if index == -1 {
		return fmt.Errorf("invalid gvg secondary sp id list")
	}
	// srcSecondarySPID := srcGvg.GetSecondarySpIds()[index]
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
		}
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
		log.CtxErrorw(ctx, "failed to sign receive task", "segment_idx", segmentIdx,
			"redundancy_idx", redundancyIdx, "error", err)
		return err
	}
	receive.SetSignature(signature)
	if err := e.baseApp.GfSpClient().ReplicatePieceToSecondary(ctx, destSPEndpoint, receive, data); err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "segment_idx", segmentIdx,
			"redundancy_idx", redundancyIdx, "error", err)
	}
	return nil
}

func (e *ExecuteModular) doneBucketMigrationReplicatePiece(ctx context.Context, gvgID uint32, objectInfo *storagetypes.ObjectInfo,
	params *storagetypes.Params, destSPEndpoint string, segmentIdx uint32) ([]byte, error) {
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(gvgID, objectInfo, params, coretask.DefaultSmallerPriority, segmentIdx, -1, 0)
	taskSignature, err := e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign done receive task", "segment_idx", segmentIdx, "error", err)
		return nil, err
	}
	receive.SetSignature(taskSignature)
	signature, err := e.baseApp.GfSpClient().DoneReplicatePieceToSecondary(ctx, destSPEndpoint, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece", "dest_sp_endpoint", destSPEndpoint,
			"segment_idx", segmentIdx, "error", err)
		return nil, err
	}
	if int(segmentIdx+1) >= len(objectInfo.GetChecksums()) {
		log.CtxErrorw(ctx, "failed to done replicate piece, replicate idx out of bounds", "segment_idx", segmentIdx)
		return nil, ErrReplicateIdsOutOfBounds
	}
	// var blsPubKey bls.PublicKey
	// err = veritySignature(ctx, rTask.GetObjectInfo().Id.Uint64(), rTask.GetGlobalVirtualGroupId(), integrity,
	//	storagetypes.GenerateHash(rTask.GetObjectInfo().GetChecksums()[:]), signature, blsPubKey)
	log.CtxDebugw(ctx, "succeed to done replicate", "dest_sp_endpoint", destSPEndpoint, "segment_idx", segmentIdx)
	return signature, nil
}

// HandleMigratePieceTask handles the migrate piece task
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
		pieceDataList = make([][]byte, segmentCount)
		doneCh        = make(chan bool)
		pieceCount    = int32(0)
		redundancyIdx = task.GetRedundancyIdx()
	)

	if task == nil {
		return ErrDanglingPointer
	}
	defer func() {
		close(doneCh)
	}()

	for i := 0; i < int(segmentCount); i++ {
		go func(mpt *gfsptask.GfSpMigratePieceTask, segIdx int) {
			pieceDataList[segIdx] = nil
			// check migrating segment pieces or ec pieces
			pieceData, err := e.sendRequest(ctx, mpt)
			if err != nil {
				log.CtxErrorw(ctx, "failed to migrate piece data", "objectID", mpt.GetObjectInfo().Id.Uint64(),
					"object_name", mpt.GetObjectInfo().GetObjectName(), "segment index", segIdx, "SP endpoint",
					task.GetSrcSpEndpoint(), "error", err)
				return
			}
			pieceDataList[segIdx] = pieceData
			if atomic.AddInt32(&pieceCount, 1) == int32(segmentCount) {
				doneCh <- true
			}
		}(task, i)
	}

	ok := <-doneCh
	if !ok {
		log.CtxErrorw(ctx, "failed to get pieces from another SP")
	}

	if err := e.setMigratePiecesMetadata(task.GetObjectInfo(), pieceDataList, task.GetRedundancyIdx()); err != nil {
		log.CtxErrorw(ctx, "failed to set piece checksum and integrity hash into spdb", "objectID",
			task.GetObjectInfo().Id.Uint64(), "object_name", task.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}

	for index, pieceData := range pieceDataList {
		var pieceKey string
		if redundancyIdx == primarySPRedundancyIdx {
			pieceKey = e.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), uint32(index))
		} else {
			pieceKey = e.baseApp.PieceOp().ECPieceKey(task.GetObjectInfo().Id.Uint64(), uint32(index), uint32(redundancyIdx))
		}
		if err := e.baseApp.PieceStore().PutPiece(ctx, pieceKey, pieceData); err != nil {
			log.CtxErrorw(ctx, "failed to put piece data to primarySP", "pieceKey", pieceKey, "error", err)
			return err
		}
	}
	log.Infow("migrate all pieces successfully", "objectID", task.GetObjectInfo().Id.Uint64(),
		"object_name", task.GetObjectInfo().GetObjectName(), "SP endpoint", task.GetSrcSpEndpoint())
	return nil
}

func (e *ExecuteModular) sendRequest(ctx context.Context, task *gfsptask.GfSpMigratePieceTask) ([]byte, error) {
	var (
		// sig       []byte
		pieceData []byte
		err       error
	)
	// sig, err = e.baseApp.GfSpClient().SignMigratePiece(ctx, &task)
	// if err != nil {
	// 	log.CtxErrorw(ctx, "failed to sign migrate piece task", "objectID", migratePiece.GetObjectInfo().Id.Uint64(),
	// 		"object_name", migratePiece.GetObjectInfo().GetObjectName(), "error", err)
	// 	return nil, err
	// }
	// migratePiece.Signature = sig
	// migrate primarySP or secondarySP piece data
	pieceData, err = e.baseApp.GfSpClient().MigratePiece(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to migrate piece data", "objectID",
			task.GetObjectInfo().Id.Uint64(), "object_name", task.GetObjectInfo().GetObjectName(), "error", err)
		return nil, err
	}
	log.CtxInfow(ctx, "succeed to get piece from another sp", "objectID", task.GetObjectInfo().Id.Uint64(),
		"object_name", task.GetObjectInfo().GetObjectName(), "segIdx", task.GetSegmentIdx(), "redundancyIdx", task.GetRedundancyIdx())
	return pieceData, nil
}

// setMigratePiecesMetadata writes piece checksum and integrity hash into db
// 1. generate piece checksum list and integrity hash
// 2. compare generated integrity hash to chain integrity hash, if they are equal write to db
func (e *ExecuteModular) setMigratePiecesMetadata(objectInfo *storagetypes.ObjectInfo, pieceData [][]byte, redundancyIdx int32) error {
	pieceChecksumList := make([][]byte, len(pieceData))
	for i, v := range pieceData {
		pieceChecksumList[i] = nil
		checksum := hash.GenerateChecksum(v)
		pieceChecksumList[i] = checksum
	}
	migratedIntegrityHash := hash.GenerateIntegrityHash(pieceChecksumList)
	var chainIntegrityHash []byte
	if redundancyIdx < -1 || redundancyIdx > 5 {
		// TODO: define an error
		return fmt.Errorf("invalid redundancy index")
	}
	if redundancyIdx == -1 {
		// primarySP
		chainIntegrityHash = objectInfo.GetChecksums()[0]
	} else {
		// secondarySP
		chainIntegrityHash = objectInfo.GetChecksums()[redundancyIdx+1]
	}
	if !bytes.Equal(migratedIntegrityHash, chainIntegrityHash) {
		log.Errorw("migrated pieces integrity is different from integrity hash on chain", "object_name",
			objectInfo.GetObjectName(), "object_id", objectInfo.Id.String())
		return ErrMigratedPieceChecksum
	}

	if err := e.baseApp.GfSpDB().SetObjectIntegrity(&corespdb.IntegrityMeta{
		ObjectID:          objectInfo.Id.Uint64(),
		RedundancyIndex:   redundancyIdx,
		IntegrityChecksum: migratedIntegrityHash,
		PieceChecksumList: pieceChecksumList,
	}); err != nil {
		log.Errorw("failed to set object integrity into spdb", "object_id", objectInfo.Id.String(),
			"object_name", objectInfo.GetObjectName(), "error", err)
		// TODO: define an sql db error
		return err
	}
	log.Infow("succeed to compute and set object integrity", "object_id", objectInfo.Id.String(),
		"object_name", objectInfo.GetObjectName())
	return nil
}

// TODO: mock function, this will be deleted in the future
// func (e *ExecuteModular) mockHandleMigratePiece(ctx context.Context) {
// 	log.Info("enter mockHandleMigratePiece function")
// 	time.Sleep(15 * time.Second)
// 	log.Info("let's do!!!")
// 	objectOne, err := e.baseApp.Consensus().QueryObjectInfo(ctx, "migratepiece", "random_file")
// 	if err != nil {
// 		log.Errorw("failed to query objectOne", "error", err)
// 	}
// 	log.Infow("print object one detailed messages", "objectOne", objectOne)
// 	objectTwo, err := e.baseApp.Consensus().QueryObjectInfo(ctx, "mybu", "random")
// 	if err != nil {
// 		log.Errorw("failed to query objectTwo", "error", err)
// 	}
// 	log.Infow("print object two detailed messages", "objectTwo", objectTwo)
// 	params := &storagetypes.Params{
// 		VersionedParams: storagetypes.VersionedParams{
// 			MaxSegmentSize:          16777216,
// 			RedundantDataChunkNum:   4,
// 			RedundantParityChunkNum: 2,
// 			MinChargeSize:           1048576,
// 		},
// 		MaxPayloadSize:                   2147483648,
// 		MirrorBucketRelayerFee:           "250000000000000",
// 		MirrorBucketAckRelayerFee:        "250000000000000",
// 		MirrorObjectRelayerFee:           "250000000000000",
// 		MirrorObjectAckRelayerFee:        "250000000000000",
// 		MirrorGroupRelayerFee:            "250000000000000",
// 		MirrorGroupAckRelayerFee:         "250000000000000",
// 		MaxBucketsPerAccount:             100,
// 		DiscontinueCountingWindow:        10000,
// 		DiscontinueObjectMax:             18446744073709551615,
// 		DiscontinueBucketMax:             18446744073709551615,
// 		DiscontinueConfirmPeriod:         5,
// 		DiscontinueDeletionMax:           100,
// 		StalePolicyCleanupMax:            200,
// 		MinQuotaUpdateInterval:           2592000,
// 		MaxLocalVirtualGroupNumPerBucket: 10,
// 	}
//
// 	log.Infow("start executing migrate piece: PrimarySP", "objectInfoOne", objectOne.Id.String())
// 	ecPieceCount := params.VersionedParams.GetRedundantDataChunkNum() + params.VersionedParams.GetRedundantParityChunkNum()
// 	migratePieceOne := &gfsptask.GfSpMigratePieceTask{
// 		ObjectInfo:    objectOne,
// 		StorageParams: params,
// 		SrcSpEndpoint: "http://127.0.0.1:9033",
// 		RedundancyIdx: primarySPRedundancyIdx,
// 	}
// 	if err := e.HandleMigratePieceTask(ctx, migratePieceOne); err != nil {
// 		log.CtxErrorw(ctx, "failed to migrate segment pieces", "error", err)
// 	}
//
// 	log.Infow("start executing migrate piece: SecondarySP", "objectInfoTwo", objectTwo.Id.String())
// 	migratePieceTwo := &gfsptask.GfSpMigratePieceTask{
// 		ObjectInfo:    objectTwo,
// 		StorageParams: params,
// 		SrcSpEndpoint: "http://127.0.0.1:9033",
// 		RedundancyIdx: int32(ecPieceCount),
// 	}
// 	if err := e.HandleMigratePieceTask(ctx, migratePieceTwo); err != nil {
// 		log.CtxErrorw(ctx, "failed to migrate ec pieces", "error", err)
// 	}
// }
