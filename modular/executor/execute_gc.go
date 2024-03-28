package executor

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/modular/manager"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type GCWorker struct {
	e *ExecuteModular
}

// NewGCWorker returns gc worker instance.
func NewGCWorker(e *ExecuteModular) *GCWorker {
	return &GCWorker{
		e: e,
	}
}

// deleteObjectPiecesAndIntegrityMeta used by gcZombiePiece
func (gc *GCWorker) deleteObjectPiecesAndIntegrityMeta(ctx context.Context, integrityMeta *corespdb.IntegrityMeta, objectVersion int64) error {
	objID := integrityMeta.ObjectID
	redundancyIdx := integrityMeta.RedundancyIndex
	maxSegment := len(integrityMeta.PieceChecksumList)

	// delete object pieces
	for segmentIdx := uint32(0); segmentIdx <= uint32(maxSegment); segmentIdx++ {
		gc.deletePiece(ctx, objID, objectVersion, segmentIdx, redundancyIdx)
	}

	// delete integrity meta
	err := gc.e.baseApp.GfSpDB().DeleteObjectIntegrity(objID, redundancyIdx)
	log.CtxDebugw(ctx, "succeed to delete all object segment and integrity meta", "object_id", objID, "integrity_meta", integrityMeta, "error", err)

	return nil
}

func (gc *GCWorker) deleteObjectSegmentsAndIntegrity(ctx context.Context, objectInfo *storagetypes.ObjectInfo) error {
	var (
		storageParams *storagetypes.Params
		err           error
	)
	if storageParams, err = gc.e.baseApp.Consensus().QueryStorageParamsByTimestamp(
		context.Background(), objectInfo.GetLatestUpdatedTime()); err != nil {
		log.Errorw("failed to query storage params", "error", err)
		return errors.New("failed to query storage params")
	}
	segmentCount := gc.e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
		storageParams.VersionedParams.GetMaxSegmentSize())
	for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
		pieceKey := gc.e.baseApp.PieceOp().SegmentPieceKey(objectInfo.Id.Uint64(), segIdx, objectInfo.Version)
		deleteErr := gc.e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
		log.CtxDebugw(ctx, "succeed to delete the primary sp segment", "object_info", objectInfo,
			"piece_key", pieceKey, "error", deleteErr)
	}
	deleteErr := gc.e.baseApp.GfSpDB().DeleteObjectIntegrity(objectInfo.Id.Uint64(), piecestore.PrimarySPRedundancyIndex)
	log.CtxDebugw(ctx, "delete the object and integrity meta", "object_info", objectInfo, "error", deleteErr)
	return nil
}

// deletePiece delete single piece if meta data or chain has object info
func (gc *GCWorker) deletePiece(ctx context.Context, objID uint64, objectVersion int64, segmentIdx uint32, redundancyIdx int32) {
	var pieceKey string
	if redundancyIdx != piecestore.PrimarySPRedundancyIndex {
		pieceKey = gc.e.baseApp.PieceOp().ECPieceKey(objID, segmentIdx, uint32(redundancyIdx), objectVersion)
	} else {
		pieceKey = gc.e.baseApp.PieceOp().SegmentPieceKey(objID, segmentIdx, objectVersion)
	}
	deleteErr := gc.e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
	log.CtxDebugw(ctx, "succeed to delete the sp piece", "object_id", objID,
		"piece_key", pieceKey, "error", deleteErr)
}

// deletePieceAndPieceChecksum delete single piece and it's corresponding piece checksum
func (gc *GCWorker) deletePieceAndPieceChecksum(ctx context.Context, piece *spdb.GCPieceMeta) error {
	objID := piece.ObjectID
	segmentIdx := piece.SegmentIndex
	redundancyIdx := piece.RedundancyIndex
	version := piece.Version
	log.CtxInfow(ctx, "start to delete piece and piece checksum", "object_id", objID, "segmentIdx", segmentIdx, "redundancyIdx", redundancyIdx)

	gc.deletePiece(ctx, piece.ObjectID, version, piece.SegmentIndex, piece.RedundancyIndex)
	err := gc.e.baseApp.GfSpDB().DeleteReplicatePieceChecksum(objID, segmentIdx, redundancyIdx)
	if err != nil {
		log.Debugf("failed to delete replicate piece checksum", "object_id", objID)
		return err
	}
	log.CtxDebugw(ctx, "succeed to delete and piece checksum", "object_id", objID, "piece_meta", piece, "error", err)
	return nil
}

// isAllowGCCheck
func (gc *GCWorker) isAllowGCCheck(objectInfo *storagetypes.ObjectInfo, bucketInfo *metadatatypes.Bucket) bool {
	// the object is not in a sealed status
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.Infow("the object isn't sealed, do not need to gc check")
		return false
	}
	// the bucket is in the process of migration
	if bucketInfo.GetBucketInfo().GetBucketStatus() == storagetypes.BUCKET_STATUS_MIGRATING {
		log.Infow("bucket is migrating, do not need to gc check")
		return false
	}
	log.Debugw("the object is sealed and the bucket is not migrating, the object can gc", "object", objectInfo, "bucket", bucketInfo)
	return true
}

func (gc *GCWorker) getGvgAndSpId(ctx context.Context, objectInfo *storagetypes.ObjectInfo) (*metadatatypes.Bucket, *virtualgrouptypes.GlobalVirtualGroup, uint32, error) {
	bucketInfo, err := gc.e.baseApp.GfSpClient().GetBucketInfoByBucketName(ctx, objectInfo.BucketName)
	if err != nil || bucketInfo == nil {
		log.Errorw("failed to get bucket by bucket name", "bucket", bucketInfo, "bucket_name", objectInfo.BucketName, "error", err)
		return nil, nil, 0, err
	}

	// piece
	gvg, err := gc.e.baseApp.GfSpClient().GetGlobalVirtualGroup(ctx, bucketInfo.BucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
	if err != nil {
		log.Errorw("failed to get global virtual group", "bucket", bucketInfo, "object", objectInfo, "error", err)
		return bucketInfo, nil, 0, err
	}

	// gvg
	spID, err := gc.e.getSPID()
	if err != nil {
		log.Errorw("failed to get sp id", "error", err)
		return bucketInfo, gvg, 0, err
	}

	return bucketInfo, gvg, spID, nil
}

// checkGVGMatchSP only return ErrInvalidRedundancyIndex means the piece was dislocation
func (gc *GCWorker) checkGVGMatchSP(ctx context.Context, objectInfo *storagetypes.ObjectInfo, redundancyIndex int32) error {
	bucketInfo, gvg, spID, err := gc.getGvgAndSpId(ctx, objectInfo)
	if err != nil || bucketInfo == nil || gvg == nil {
		log.CtxErrorw(ctx, "failed to get gvg and sp id", "object", objectInfo, "bucket", bucketInfo, "error", err)
		return err
	}

	if !gc.isAllowGCCheck(objectInfo, bucketInfo) {
		return nil
	}

	if redundancyIndex == piecestore.PrimarySPRedundancyIndex {
		if gvg.GetPrimarySpId() != spID {
			swapInInfo, err := gc.e.baseApp.Consensus().QuerySwapInInfo(ctx, gvg.FamilyId, virtualgrouptypes.NoSpecifiedGVGId)
			if err != nil {
				if strings.Contains(err.Error(), "swap in info not exist") {
					return ErrInvalidRedundancyIndex
				}
				return nil
			}
			if swapInInfo.SuccessorSpId == spID && swapInInfo.TargetSpId == gvg.PrimarySpId {
				return nil
			}
			log.CtxInfow(ctx, "the piece isn't in correct location, will be delete",
				"object_info", objectInfo, "redundancy_index", redundancyIndex, "gvg", gvg, "sp_id", spID)
			return ErrInvalidRedundancyIndex
		}
	} else {
		if gvg.GetSecondarySpIds()[redundancyIndex] != spID {
			swapInInfo, err := gc.e.baseApp.Consensus().QuerySwapInInfo(ctx, 0, gvg.Id)
			if err != nil {
				if strings.Contains(err.Error(), "swap in info not exist") {
					return ErrInvalidRedundancyIndex
				}
				return nil
			}
			if swapInInfo.SuccessorSpId == spID && swapInInfo.TargetSpId == gvg.GetSecondarySpIds()[redundancyIndex] {
				return nil
			}
			log.CtxInfow(ctx, "the piece isn't in correct location, will be delete",
				"object_info", objectInfo, "redundancy_index", redundancyIndex, "gvg", gvg, "sp_id", spID)
			return ErrInvalidRedundancyIndex
		}
	}
	return nil
}

func (e *ExecuteModular) HandleGCObjectTask(ctx context.Context, task coretask.GCObjectTask) {
	var (
		err                error
		waitingGCObjects   []*metadatatypes.Object
		currentGCBlockID   uint64
		currentGCObjectID  uint64
		responseEndBlockID uint64
		storageParams      *storagetypes.Params
		gcObjectNumber     int
		tryAgainLater      bool
		taskIsCanceled     bool
		hasNoObject        bool
		isSucceed          bool
	)

	reportProgress := func() bool {
		reportErr := e.ReportTask(ctx, task)
		log.CtxDebugw(ctx, "gc object task report progress", "task_info", task.Info(), "error", reportErr)
		return errors.Is(reportErr, manager.ErrCanceledTask)
	}

	defer func() {
		if err == nil && (isSucceed || hasNoObject) { // succeed
			task.SetCurrentBlockNumber(task.GetEndBlockNumber() + 1)
			reportProgress()
		} else { // failed
			task.SetError(err)
			reportProgress()
		}
		log.CtxDebugw(ctx, "gc object task", "task_info", task.Info(), "is_succeed", isSucceed,
			"response_end_block_id", responseEndBlockID, "waiting_gc_object_number", len(waitingGCObjects),
			"has_gc_object_number", gcObjectNumber, "try_again_later", tryAgainLater,
			"task_is_canceled", taskIsCanceled, "has_no_object", hasNoObject, "error", err)
	}()

	if waitingGCObjects, responseEndBlockID, err = e.baseApp.GfSpClient().ListDeletedObjectsByBlockNumberRange(ctx,
		e.baseApp.OperatorAddress(), task.GetStartBlockNumber(), task.GetEndBlockNumber(), true); err != nil {
		log.CtxErrorw(ctx, "failed to query deleted object list", "task_info", task.Info(), "error", err)
		return
	}
	if responseEndBlockID < task.GetStartBlockNumber() || responseEndBlockID < task.GetEndBlockNumber() {
		tryAgainLater = true
		log.CtxInfow(ctx, "metadata is not latest, try again later", "response_end_block_id",
			responseEndBlockID, "task_info", task.Info())
		return
	}
	if len(waitingGCObjects) == 0 {
		log.Info("no waiting gc objects")
		hasNoObject = true
		return
	}

	// TODO get sp id from config
	spId, err := e.getSPID()
	if err != nil {
		log.Errorw("failed to get sp id", "error", err)
		return
	}

	for _, object := range waitingGCObjects {
		if storageParams, err = e.baseApp.Consensus().QueryStorageParamsByTimestamp(
			context.Background(), object.GetObjectInfo().GetCreateAt()); err != nil {
			log.Errorw("failed to query storage params", "task_info", task.Info(), "error", err)
			return
		}

		currentGCBlockID = uint64(object.GetDeleteAt())
		objectInfo := object.GetObjectInfo()
		objectVersion := objectInfo.Version
		currentGCObjectID = objectInfo.Id.Uint64()
		if currentGCBlockID < task.GetCurrentBlockNumber() {
			log.Errorw("skip gc object", "object_info", objectInfo,
				"task_current_gc_block_id", task.GetCurrentBlockNumber())
			continue
		}
		segmentCount := e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
			storageParams.VersionedParams.GetMaxSegmentSize())
		for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
			pieceKey := e.baseApp.PieceOp().SegmentPieceKey(currentGCObjectID, segIdx, objectVersion)
			// ignore this delete api error, TODO: refine gc workflow by enrich metadata index.
			deleteErr := e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
			log.CtxDebugw(ctx, "delete the primary sp pieces", "object_info", objectInfo,
				"piece_key", pieceKey, "error", deleteErr)
		}
		bucketInfo, err := e.baseApp.GfSpClient().GetBucketInfoByBucketName(ctx, objectInfo.BucketName)
		if err != nil || bucketInfo == nil {
			log.Errorw("failed to get bucket by bucket name", "bucket_name", objectInfo.BucketName, "error", err)
			return
		}
		gvg, err := e.baseApp.GfSpClient().GetGlobalVirtualGroup(ctx, bucketInfo.BucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
		if err != nil {
			log.Errorw("failed to get global virtual group", "error", err)
			return
		}

		var redundancyIndex int32 = -1
		for rIdx, sspId := range gvg.GetSecondarySpIds() {
			if spId == sspId {
				redundancyIndex = int32(rIdx)
				for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
					pieceKey := e.baseApp.PieceOp().ECPieceKey(currentGCObjectID, segIdx, uint32(rIdx), objectVersion)
					if objectInfo.GetRedundancyType() == storagetypes.REDUNDANCY_REPLICA_TYPE {
						pieceKey = e.baseApp.PieceOp().SegmentPieceKey(objectInfo.Id.Uint64(), segIdx, objectVersion)
					}
					// ignore this delete api error, TODO: refine gc workflow by enrich metadata index.
					deleteErr := e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
					log.CtxDebugw(ctx, "delete the secondary sp pieces",
						"object_info", objectInfo, "piece_key", pieceKey, "error", deleteErr)
				}
			}
		}
		// ignore this delete api error, TODO: refine gc workflow by enrich metadata index
		deleteErr := e.baseApp.GfSpDB().DeleteObjectIntegrity(objectInfo.Id.Uint64(), redundancyIndex)
		log.CtxDebugw(ctx, "delete the object integrity meta", "object_info", objectInfo, "error", deleteErr)
		task.SetCurrentBlockNumber(currentGCBlockID)
		task.SetLastDeletedObjectId(currentGCObjectID)
		metrics.GCObjectCounter.WithLabelValues(e.Name()).Inc()
		if taskIsCanceled = reportProgress(); taskIsCanceled {
			log.CtxErrorw(ctx, "gc object task has been canceled", "current_gc_object_info", objectInfo, "task_info", task.Info())
			return
		}
		log.CtxDebugw(ctx, "succeed to gc an object", "object_info", objectInfo, "deleted_at_block_id", currentGCBlockID)
		gcObjectNumber++
	}
	isSucceed = true
}

func (e *ExecuteModular) HandleGCZombiePieceTask(ctx context.Context, task coretask.GCZombiePieceTask) {
	var (
		err                             error
		waitingVerifyGCIntegrityObjects []*corespdb.IntegrityMeta
	)

	reportProgress := func() {
		reportErr := e.ReportTask(ctx, task)
		log.CtxDebugw(ctx, "gc zombie piece task report progress", "task_info", task.Info(), "error", reportErr)
	}

	defer func() {
		if err == nil { // succeed
			reportProgress()
		} else { // failed
			task.SetError(err)
			reportProgress()
		}
		log.CtxDebugw(ctx, "gc zombie piece task", "task_info", task.Info(),
			"waiting_gc_object_number", len(waitingVerifyGCIntegrityObjects), "error", err)
	}()

	// verify zombie piece via IntegrityMetaTable
	if err = e.gcZombiePieceFromIntegrityMeta(ctx, task); err != nil {
		log.CtxErrorw(ctx, "failed to gc zombie piece from integrity meta", "task_info", task.Info(), "error", err)
		return
	}
	// verify and delete zombie piece via piece hash
	if err = e.gcZombiePieceFromPieceHash(ctx, task); err != nil {
		log.CtxErrorw(ctx, "failed to gc zombie piece from piece hash", "task_info", task.Info(), "error", err)
		return
	}
}

func (e *ExecuteModular) HandleGCMetaTask(ctx context.Context, task coretask.GCMetaTask) {
	var (
		err error
	)
	reportProgress := func() {
		reportErr := e.ReportTask(ctx, task)
		log.CtxDebugw(ctx, "gc meta task report progress", "task_info", task.Info(), "error", reportErr)
	}

	defer func() {
		if err == nil { // succeed
			reportProgress()
		} else { // failed
			task.SetError(err)
			reportProgress()
		}
		log.CtxDebugw(ctx, "succeed to run gc meta task", "task_info", task.Info(), "error", err)
	}()

	go e.gcMetaBucketTraffic(ctx, task)
	go e.gcMetaReadRecord(ctx, task)
}

func (e *ExecuteModular) gcMetaBucketTraffic(ctx context.Context, task coretask.GCMetaTask) error {
	now := time.Now()
	daysAgo := now.Add(-time.Duration(e.bucketTrafficKeepLatestDay) * 24 * time.Hour)
	yearMonth := sqldb.TimeToYearMonth(daysAgo)
	err := e.baseApp.GfSpDB().DeleteExpiredBucketTraffic(yearMonth)
	if err != nil {
		log.CtxErrorw(ctx, "failed to delete expired bucket traffic", "error", err)
		return err
	}

	log.CtxInfow(ctx, "succeed to delete bucket traffic", "task", task, "days_ago", daysAgo, "year_month", yearMonth)
	return nil
}

func (e *ExecuteModular) gcMetaReadRecord(ctx context.Context, task coretask.GCMetaTask) error {
	now := time.Now()
	daysAgo := now.Add(-time.Duration(e.readRecordKeepLatestDay) * time.Hour * 24)

	err := e.baseApp.GfSpDB().DeleteExpiredReadRecord(uint64(daysAgo.UnixMicro()), e.readRecordDeleteLimit)
	if err != nil {
		log.CtxErrorw(ctx, "failed to delete expired read record", "error", err)
		return err
	}
	log.CtxInfow(ctx, "succeed to delete expired read record", "task", task, "daysAgo", daysAgo)
	return nil
}

func (e *ExecuteModular) HandleGCBucketMigrationBucket(ctx context.Context, task coretask.GCBucketMigrationTask) {
	var (
		gvgList         []*virtualgrouptypes.GlobalVirtualGroup
		err             error
		startAfter      uint64 = 0
		limit           uint32 = 100
		objectList      []*metadatatypes.ObjectDetails
		gcNum           uint64 = 0
		bucketInfo      *storagetypes.BucketInfo
		gcFinisheGvgNum uint64 = 0
		gvgTotalNum     uint64 = 0
	)
	bucketID := task.GetBucketID()

	reportProgress := func() bool {
		reportErr := e.ReportTask(ctx, task)
		log.CtxDebugw(ctx, "gc object task report progress", "task_info", task.Info(), "error", reportErr)
		return errors.Is(reportErr, manager.ErrCanceledTask)
	}
	defer func() {
		task.SetTotalGvgNum(gvgTotalNum)
		task.SetGCFinishedGvgNum(gcFinisheGvgNum)
		if err == nil { // succeed
			task.SetFinished(true)
			reportProgress()
		} else { // failed
			task.SetError(err)
			reportProgress()
		}
		log.CtxDebugw(ctx, "succeed to report gc bucket migration task", "task_info", task.Info(),
			"bucket_id", bucketID, "gc_num", gcNum, "error", err)
	}()

	// list gvg
	if gvgList, err = e.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(ctx, bucketID); err != nil {
		log.CtxErrorw(ctx, "failed to list global virtual group by bucket id", "bucket_id", bucketID, "error", err)
		return
	}

	// current chain's gvg info compare metadata info
	if bucketInfo, err = e.baseApp.Consensus().QueryBucketInfoById(ctx, bucketID); err != nil || bucketInfo == nil {
		log.Errorw("failed to get bucket by bucket name", "bucket_id", bucketID, "error", err)
		return
	}
	vgfID := bucketInfo.GetGlobalVirtualGroupFamilyId()
	log.CtxInfow(ctx, "begin to gc bucket migration by bucket id", "bucket_id", bucketID, "gvgList", gvgList, "vgfID", vgfID, "error", err)

	gvgTotalNum = uint64(len(gvgList))

	for _, gvg := range gvgList {
		if gvg.FamilyId != vgfID {
			log.CtxErrorw(ctx, "failed to check gvg's status with chain's status, the gvg may be old data", "bucket_id", bucketID, "gvg", gvg, "error", err)
			err = errors.New("gvg family id mismatch")
			return
		}
		for {
			if objectList, err = e.baseApp.GfSpClient().ListObjectsByGVGAndBucketForGC(ctx, gvg.GetId(), bucketID, startAfter, limit); err != nil {
				log.CtxErrorw(ctx, "failed to list objectList by gvg and bucket for gc", "bucket_id", bucketID, "gvg", gvg, "error", err)
				return
			}
			// can delete, verify
			for index, obj := range objectList {
				gcNum++
				objectInfo := obj.GetObject().GetObjectInfo()
				startAfter = obj.GetObject().GetObjectInfo().Id.Uint64()
				if e.gcWorker.checkGVGMatchSP(ctx, objectInfo, piecestore.PrimarySPRedundancyIndex) == ErrInvalidRedundancyIndex {
					err = e.gcWorker.deleteObjectSegmentsAndIntegrity(ctx, objectInfo)
					log.CtxInfow(ctx, "succeed to delete objects by gvg and bucket for gc", "object", objectInfo, "error", err)
				}
				if gcNum%reportProgressPerN == 0 || index == len(objectList)-1 {
					log.Infow("bucket migration gc report task", "current_gvg", gvg, "bucket_id", bucketID,
						"last_gc_object_id", obj.GetObject().GetObjectInfo().Id.Uint64())
					task.SetLastGCObjectID(obj.GetObject().GetObjectInfo().Id.Uint64())
					task.SetLastGCGvgID(uint64(gvg.GetId()))
					task.SetTotalGvgNum(gvgTotalNum)
					task.SetGCFinishedGvgNum(gcFinisheGvgNum)
					if err = e.ReportTask(ctx, task); err != nil { // report task is already automatically triggered.
						log.CtxErrorw(ctx, "failed to report bucket migration task gc progress", "task_info", task, "error", err)
						return
					}
				}
			}

			if len(objectList) < int(limit) {
				log.CtxInfow(ctx, "succeed to finish one gvg bucket migration gc", "bucket_id", bucketID, "gvg", gvg)
				gcFinisheGvgNum++
				break
			}
		}
	}

	log.CtxInfow(ctx, "succeed to gc bucket migration by bucket id", "bucket_id", bucketID, "gvgs", gvgList, "vgfID", vgfID, "gc_num", gcNum)
}

func (e *ExecuteModular) gcZombiePieceFromIntegrityMeta(ctx context.Context, task coretask.GCZombiePieceTask) error {
	var (
		err                            error
		waitingVerifyGCIntegrityPieces []*corespdb.IntegrityMeta
		objInfoFromMetaData            *storagetypes.ObjectInfo
		objInfoFromChain               *storagetypes.ObjectInfo
	)

	if waitingVerifyGCIntegrityPieces, err = e.baseApp.GfSpDB().ListIntegrityMetaByObjectIDRange(int64(task.GetStartObjectId()), int64(task.GetEndObjectId()), true); err != nil {
		log.CtxErrorw(ctx, "failed to query gc integrity pieces list", "task_info", task.Info(), "error", err)
		return err
	}

	if len(waitingVerifyGCIntegrityPieces) == 0 {
		log.CtxInfow(ctx, "no waiting gc integrity pieces", "task_info", task.Info())
		return nil
	}

	for _, integrityObject := range waitingVerifyGCIntegrityPieces {
		log.CtxDebugw(ctx, "gc zombie current waiting verify gc integrity meta piece", "integrity_meta", integrityObject)
		objID := integrityObject.ObjectID
		// If metadata service has object info, use objInfoFromMetaData, otherwise query from chain
		if objInfoFromMetaData, err = e.baseApp.GfSpClient().GetObjectByID(ctx, objID); err != nil {
			log.Errorf("failed to get object meta from meta data service, will check from chain", "error", err)
			if strings.Contains(err.Error(), "no such object from metadata") {
				// If deletion is possible, has integrity hash, lacks object info, perform a chain check, and verify against the chain.
				if objInfoFromChain, err = e.baseApp.Consensus().QueryObjectInfoByID(ctx, strconv.FormatUint(objID, 10)); err != nil {
					log.Errorf("failed to get object info from chain", "error", err)
					continue
				} else {
					// 2) query metadata error, but chain has the object info, gvg  primary sp should have integrity meta
					if e.gcWorker.checkGVGMatchSP(ctx, objInfoFromChain, integrityObject.RedundancyIndex) == ErrInvalidRedundancyIndex {
						e.gcWorker.deleteObjectPiecesAndIntegrityMeta(ctx, integrityObject, objInfoFromChain.Version)
					}
					continue
				}
			}
		} else {
			// 3) check integrity meta & object info
			if e.gcWorker.checkGVGMatchSP(ctx, objInfoFromMetaData, integrityObject.RedundancyIndex) == ErrInvalidRedundancyIndex {
				e.gcWorker.deleteObjectPiecesAndIntegrityMeta(ctx, integrityObject, objInfoFromMetaData.Version)
			}
		}
	}
	return nil
}

func (e *ExecuteModular) gcZombiePieceFromPieceHash(ctx context.Context, task coretask.GCZombiePieceTask) error {
	var (
		err                    error
		waitingVerifyGCPieces  []*spdb.GCPieceMeta
		objectInfoFromMetadata *storagetypes.ObjectInfo
		objInfoFromChain       *storagetypes.ObjectInfo
	)
	log.CtxDebugw(ctx, "start to gc zombie piece from piece hash", "task_info", task.Info())

	// replicate piece checksum must on secondary sp
	if waitingVerifyGCPieces, err = e.baseApp.GfSpDB().ListReplicatePieceChecksumByObjectIDRange(
		int64(task.GetStartObjectId()), int64(task.GetEndObjectId())); err != nil {
		log.CtxErrorw(ctx, "failed to query replicate piece checksum", "task_info", task.Info(), "error", err)
		return err
	}

	if len(waitingVerifyGCPieces) == 0 {
		log.CtxInfow(ctx, "no waiting gc pieces", "task_info", task.Info())
		return nil
	}

	for _, piece := range waitingVerifyGCPieces {
		log.CtxDebugw(ctx, "gc zombie current waiting verify gc meta piece", "piece", piece)
		objID := piece.ObjectID

		// Get object information from metadata
		// Note:Currently, waitingVerifyGCPieces is returned at the granularity of a piece. We may need a cache to
		// store already queried information if there are performance issues:
		//
		// 1) objectInfoFromMetadata, objInfoFromChain;
		// 2) bucketInfo, gvgInfo, and so on.
		if objectInfoFromMetadata, err = e.baseApp.GfSpClient().GetObjectByID(ctx, objID); err != nil {
			// If the object doesn't exist in metadata, recheck from the chain before proceeding with the deletion.
			if strings.Contains(err.Error(), "no such object from metadata") {
				if objInfoFromChain, err = e.baseApp.Consensus().QueryObjectInfoByID(ctx, strconv.FormatUint(objID, 10)); err != nil {
					if strings.Contains(err.Error(), "No such object") {
						// 1) This object does not exist on the chain
						log.Infof("the object doesn't exist in metadata and chain, the zombie piece should be deleted", "piece", piece)
						e.gcWorker.deletePieceAndPieceChecksum(ctx, piece)
					}
				} else {
					// an object under Updating might have records in piece hash table(during replication); since the SP can also be picked as
					// a secondary SP(in another GVG) for the updated object, the redundancyIndex would be different
					// from the prev one(onchain or metadata still have it until the object is sealed).
					// Need to make sure these updated pieces will not be GC.
					if objInfoFromChain.IsUpdating {
						return nil
					}
					// 2) If there is an error querying metadata but the chain contains object information, recheck the meta.
					if e.gcWorker.checkGVGMatchSP(ctx, objInfoFromChain, piece.RedundancyIndex) == ErrInvalidRedundancyIndex {
						e.gcWorker.deletePieceAndPieceChecksum(ctx, piece)
					}
				}
			}
		} else {
			// an object under Updating might have records in piece hash table(during replication); since the SP can also be picked as
			// a secondary SP(in another GVG) for the updated object, the redundancyIndex would be different
			// from the prev one(onchain or metadata still have it until the object is sealed).
			// Need to make sure these updated pieces will not be GC.
			if objectInfoFromMetadata.IsUpdating {
				return nil
			}
			// 3) Validate using Metadata information.
			if e.gcWorker.checkGVGMatchSP(ctx, objectInfoFromMetadata, piece.RedundancyIndex) == ErrInvalidRedundancyIndex {
				e.gcWorker.deletePieceAndPieceChecksum(ctx, piece)
			}
		}
	}
	return nil
}

func (e *ExecuteModular) HandleGCStaleVersionObjectTask(ctx context.Context, task coretask.GCStaleVersionObjectTask) {
	var err error
	reportProgress := func() {
		reportErr := e.ReportTask(ctx, task)
		log.CtxDebugw(ctx, "gc stale version object task report progress", "task_info", task.Info(), "error", reportErr)
	}

	defer func() {
		if err == nil { // succeed
			reportProgress()
		} else { // failed
			task.SetError(err)
			reportProgress()
		}
	}()
	log.CtxDebugw(ctx, "HandleGCStaleVersionObjectTask", "task_info", task.Info())

	if err = e.gcStaleVersionObjectFromShadowIntegrityMeta(ctx, task); err != nil {
		log.CtxErrorw(ctx, "failed to gc zombie piece from integrity meta", "task_info", task.Info(), "error", err)
		return
	}
}

func (e *ExecuteModular) gcStaleVersionObjectFromShadowIntegrityMeta(ctx context.Context, task coretask.GCStaleVersionObjectTask) error {
	var err error
	objectID := task.GetObjectId()
	metaSegmentCount := uint32(len(task.GetPieceChecksumList()))

	gcStaleVersionPieces := func(objectID uint64, segmentCount uint32, version int64, redundancyIndex int32) error {
		for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
			var pieceKey string
			if task.GetRedundancyIndex() == piecestore.PrimarySPRedundancyIndex {
				pieceKey = e.baseApp.PieceOp().SegmentPieceKey(objectID, segIdx, version)
			} else {
				pieceKey = e.baseApp.PieceOp().ECPieceKey(objectID, segIdx, uint32(redundancyIndex), version)
			}
			err = e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
			if err != nil {
				log.CtxErrorw(ctx, "failed to delete the stale sp pieces", "object_id", objectID,
					"piece_key", pieceKey, "error", err)
				if strings.Contains(err.Error(), "The specified key does not exist") {
					continue
				} else {
					return err
				}
			}
		}
		err = e.baseApp.GfSpDB().DeleteShadowObjectIntegrity(objectID, task.GetRedundancyIndex())
		if err != nil {
			log.CtxErrorw(ctx, "failed to delete shadow object integrity meta", "object_id", objectID, "error", err)
			return err
		}
		return nil
	}

	objectInfo, err := e.baseApp.Consensus().QueryObjectInfoByID(context.Background(), util.Uint64ToString(objectID))
	if err != nil {
		log.Errorw("failed to query object info", "object_id", task.GetObjectId(), "error", err)

		if strings.Contains(err.Error(), "No such object") {
			// if the object is deleted, can gc the piece according to the shadow integrity meta
			err = gcStaleVersionPieces(task.GetObjectId(), metaSegmentCount, task.GetVersion(), task.GetRedundancyIndex())
			if err != nil {
				log.Errorw("failed to gc stale pieces", "object_id", task.GetObjectId(), "version", task.GetVersion(),
					"segment_count", metaSegmentCount, "redundancy_index", task.GetRedundancyIndex(), "error", err)
				return err
			}
			return nil
		}
		return err
	}

	if task.GetVersion() < objectInfo.Version {
		err = gcStaleVersionPieces(task.GetObjectId(), metaSegmentCount, task.GetVersion(), task.GetRedundancyIndex())
		if err != nil {
			log.Errorw("failed to gc stale pieces", "object_id", task.GetObjectId(), "version", task.GetVersion(),
				"segment_count", metaSegmentCount, "redundancy_index", task.GetRedundancyIndex(), "error", err)
			return err
		}
	} else if task.GetVersion() == objectInfo.Version {
		// the shadowIntegrityMeta is stale, the object is experiencing another update
		if objectInfo.IsUpdating {
			err = gcStaleVersionPieces(task.GetObjectId(), metaSegmentCount, task.GetVersion(), task.GetRedundancyIndex())
			if err != nil {
				log.Errorw("failed to gc stale pieces", "object_id", task.GetObjectId(), "version", task.GetVersion(),
					"segment_count", metaSegmentCount, "redundancy_index", task.GetRedundancyIndex(), "error", err)
				return err
			}
		} else {
			var staleIntegrityMeta *spdb.IntegrityMeta
			if task.GetRedundancyIndex() == piecestore.PrimarySPRedundancyIndex {
				staleIntegrityMeta, err = e.baseApp.GfSpDB().GetObjectIntegrity(task.GetObjectId(), task.GetRedundancyIndex())
				if err != nil {
					log.Errorw("failed to get object integrity meta", "object_id", task.GetObjectId(), "error", err)
					return err
				}
				for segIdx := uint32(0); segIdx < uint32(len(staleIntegrityMeta.PieceChecksumList)); segIdx++ {
					pieceKey := e.baseApp.PieceOp().SegmentPieceKey(task.GetObjectId(), segIdx, task.GetVersion()-1)
					err = e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
					if err != nil {
						log.CtxErrorw(ctx, "failed to delete the stale sp pieces", "object_info", objectInfo,
							"piece_key", pieceKey, "error", err)
						if strings.Contains(err.Error(), "The specified key does not exist") {
							continue
						} else {
							return err
						}
					}
				}
			} else {
				// the SP might have 2 integrityMetas for an objet if SP play as primary and 1 of secondary in GVG
				staleIntegrityMetas, err := e.baseApp.GfSpDB().ListIntegrityMetaByObjectIDRange(int64(objectID), int64(objectID)+1, true)
				if err != nil {
					return err
				}
				for _, sim := range staleIntegrityMetas {
					if sim.RedundancyIndex != piecestore.PrimarySPRedundancyIndex {
						staleIntegrityMeta = sim
						break
					}
				}
				// it might be already GC by the GCZombie job
				if staleIntegrityMeta == nil {
					integrityMeta := &spdb.IntegrityMeta{
						ObjectID:          task.GetObjectId(),
						RedundancyIndex:   task.GetRedundancyIndex(),
						IntegrityChecksum: task.GetIntegrityChecksum(),
						PieceChecksumList: task.GetPieceChecksumList(),
					}
					err = e.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta)
					if err != nil {
						return err
					}
					err = e.baseApp.GfSpDB().DeleteShadowObjectIntegrity(task.GetObjectId(), task.GetRedundancyIndex())
					if err != nil {
						log.Errorw("failed to delete shadow object integrity meta", "object_id", task.GetObjectId())
						return err
					}
					return nil
				}
				for segIdx := uint32(0); segIdx < uint32(len(staleIntegrityMeta.PieceChecksumList)); segIdx++ {
					pieceKey := e.baseApp.PieceOp().ECPieceKey(task.GetObjectId(), segIdx, uint32(staleIntegrityMeta.RedundancyIndex), task.GetVersion()-1)
					err = e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
					if err != nil {
						log.CtxErrorw(ctx, "failed to delete the stale sp pieces", "object_info", objectInfo,
							"piece_key", pieceKey, "error", err)
						// the piece might have been gc the the gcZombiePiece Task if object is deleted and the gcStale tasks are not executed for a while
						if strings.Contains(err.Error(), "The specified key does not exist") {
							continue
						} else {
							return err
						}
					}
				}
			}
			err = e.baseApp.GfSpDB().DeleteObjectIntegrity(objectID, staleIntegrityMeta.RedundancyIndex)
			if err != nil {
				log.CtxErrorw(ctx, "failed to delete the object integrity meta", "object_id", objectID, "error", err)
				return err
			}
			integrityMeta := &spdb.IntegrityMeta{
				ObjectID:          task.GetObjectId(),
				RedundancyIndex:   task.GetRedundancyIndex(),
				IntegrityChecksum: task.GetIntegrityChecksum(),
				PieceChecksumList: task.GetPieceChecksumList(),
				ObjectSize:        task.GetObjectSize(),
			}
			err = e.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta)
			if err != nil {
				return err
			}
			err = e.baseApp.GfSpDB().DeleteShadowObjectIntegrity(task.GetObjectId(), task.GetRedundancyIndex())
			if err != nil {
				log.Errorw("failed to delete shadow object integrity meta", "object_id", task.GetObjectId())
				return err
			}
		}
	} else {
		// user cancels the update
		if !objectInfo.IsUpdating {
			err = gcStaleVersionPieces(task.GetObjectId(), metaSegmentCount, task.GetVersion(), task.GetRedundancyIndex())
			if err != nil {
				log.Errorw("failed to gc stale pieces", "object_id", task.GetObjectId(), "version", task.GetVersion(),
					"segment_count", metaSegmentCount, "redundancy_index", task.GetRedundancyIndex(), "error", err)
				return err
			}
		}
	}
	return nil
}
