package executor

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-common/go/hash"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
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
		gvgID                = task.GetGvg().GetId()
		bucketID             = task.GetBucketID()
		lastMigratedObjectID = task.GetLastMigratedObjectID()
		objectList           []*metadatatypes.Object
		err                  error
	)

	for {
		if bucketID == 0 {
			// if bucketID is 0, it indicates that it is a sp exiting task
			objectList, err = e.baseApp.GfSpClient().ListObjectsInGVG(ctx, gvgID, lastMigratedObjectID+1, queryLimit)
			if err != nil {
				log.CtxErrorw(ctx, "failed to list objects in gvg", "gvg_id", gvgID,
					"current_migrated_object_id", lastMigratedObjectID+1, "error", err)
			}
			log.Infow("sp exiting", "objectList", objectList)
		} else {
			// if bucketID is not equal to 0, it indicates that it is a bucket migration task
			objectList, err = e.baseApp.GfSpClient().ListObjectsInGVGAndBucket(ctx, gvgID, bucketID, lastMigratedObjectID, queryLimit)
			if err != nil {
				log.CtxErrorw(ctx, "failed to list objects in gvg and bucket", "gvg_id", gvgID, "bucket_id", bucketID,
					"current_migrated_object_id", lastMigratedObjectID, "error", err)
			}
			log.CtxDebugw(ctx, "success to list objects in gvg and bucket", "gvg_id", gvgID, "bucket_id", bucketID,
				"current_migrated_object_id", lastMigratedObjectID, "error", err)
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
	spInfo, err := e.baseApp.Consensus().QuerySP(ctx, e.baseApp.OperatorAddress())
	if err != nil {
		log.Errorw("failed to query sp info on chain", "error", err)
		return err
	}

	bucketName := object.GetObjectInfo().GetBucketName()
	bucketInfo, err := e.baseApp.Consensus().QueryBucketInfo(ctx, bucketName)
	if err != nil {
		log.CtxErrorw(ctx, "failed to query bucket info by bucket name", "bucketName", bucketName, "error", err)
		return err
	}
	if bucketID != bucketInfo.Id.Uint64() {
		bucketID = bucketInfo.Id.Uint64()
	}

	redundancyIdx, isPrimary, err := util.ValidateAndGetSPIndexWithinGVGSecondarySPs(ctx, e.baseApp.GfSpClient(),
		spInfo.GetId(), bucketID, object.GetObjectInfo().GetLocalVirtualGroupId())
	if err != nil {
		log.CtxErrorw(ctx, "failed to validate and get sp index within gvg secondary sps", "error", err)
		return err
	}
	migratePieceTask := &gfsptask.GfSpMigratePieceTask{
		ObjectInfo:    object.GetObjectInfo(),
		StorageParams: params,
		SrcSpEndpoint: task.GetSrcSp().GetEndpoint(),
		RedundancyIdx: task.GetRedundancyIdx(),
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

// HandleMigratePieceTask handles the migrate piece task, it will send requests to the exiting SP to get piece data
// integrity hash and piece checksum to migrate the receiving SP. PieceData would be written to PieceStore, integrity hash
// and piece checksum will be written sql db.
// currently get and handle data one by one; in the future, use concurrency
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
		quitCh        = make(chan bool)
		pieceCount    = int32(0)
		redundancyIdx = task.GetRedundancyIdx()
	)

	if task == nil {
		return ErrDanglingPointer
	}
	defer func() {
		close(quitCh)
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
				quitCh <- true
			}
		}(task, i)
	}

	ok := <-quitCh
	if !ok {
		log.CtxErrorw(ctx, "failed to get pieces from another SP")
	}

	if err := e.setMigratePiecesMetadata(task.GetObjectInfo(), pieceDataList, int(task.GetRedundancyIdx())); err != nil {
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
func (e *ExecuteModular) setMigratePiecesMetadata(objectInfo *storagetypes.ObjectInfo, pieceData [][]byte, redundancyIdx int) error {
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
