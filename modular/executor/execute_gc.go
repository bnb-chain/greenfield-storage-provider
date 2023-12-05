package executor

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/modular/manager"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/forbole/juno/v4/common"
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
func (gc *GCWorker) deleteObjectPiecesAndIntegrityMeta(ctx context.Context, integrityMeta *corespdb.IntegrityMeta) error {
	objID := integrityMeta.ObjectID
	redundancyIdx := integrityMeta.RedundancyIndex
	maxSegment := len(integrityMeta.PieceChecksumList)

	// delete object pieces
	for segmentIdx := uint32(0); segmentIdx <= uint32(maxSegment); segmentIdx++ {
		gc.deletePiece(ctx, objID, segmentIdx, redundancyIdx)
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
		context.Background(), objectInfo.GetCreateAt()); err != nil {
		log.Errorw("failed to query storage params", "error", err)
		return errors.New("failed to query storage params")
	}
	segmentCount := gc.e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
		storageParams.VersionedParams.GetMaxSegmentSize())
	for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
		pieceKey := gc.e.baseApp.PieceOp().SegmentPieceKey(objectInfo.Id.Uint64(), segIdx)
		deleteErr := gc.e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
		log.CtxDebugw(ctx, "succeed to delete the primary sp segment", "object_info", objectInfo,
			"piece_key", pieceKey, "error", deleteErr)
	}
	deleteErr := gc.e.baseApp.GfSpDB().DeleteObjectIntegrity(objectInfo.Id.Uint64(), piecestore.PrimarySPRedundancyIndex)
	log.CtxDebugw(ctx, "delete the object and integrity meta", "object_info", objectInfo, "error", deleteErr)
	return nil
}

// deletePiece delete single piece if meta data or chain has object info
func (gc *GCWorker) deletePiece(ctx context.Context, objID uint64, segmentIdx uint32, redundancyIdx int32) {
	var pieceKey string
	if redundancyIdx != piecestore.PrimarySPRedundancyIndex {
		pieceKey = gc.e.baseApp.PieceOp().ECPieceKey(objID, segmentIdx, uint32(redundancyIdx))
	} else {
		pieceKey = gc.e.baseApp.PieceOp().SegmentPieceKey(objID, segmentIdx)
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
	log.CtxInfow(ctx, "start to delete piece and piece checksum", "object_id", objID, "segmentIdx", segmentIdx, "redundancyIdx", redundancyIdx)

	gc.deletePiece(ctx, piece.ObjectID, piece.SegmentIndex, piece.RedundancyIndex)
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
	bucketInfo, err := gc.e.baseApp.GfSpClient().GetBucketByBucketName(ctx, objectInfo.BucketName, true)
	if err != nil || bucketInfo == nil {
		log.Errorw("failed to get bucket by bucket name", "bucket", bucketInfo, "error", err)
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
			log.CtxInfow(ctx, "the piece isn't in correct location, will be delete",
				"object_info", objectInfo, "redundancy_index", redundancyIndex, "gvg", gvg, "sp_id", spID)
			return ErrInvalidRedundancyIndex
		}
	} else {
		if gvg.GetSecondarySpIds()[redundancyIndex] != spID {
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
		currentGCObjectID = objectInfo.Id.Uint64()
		if currentGCBlockID < task.GetCurrentBlockNumber() {
			log.Errorw("skip gc object", "object_info", objectInfo,
				"task_current_gc_block_id", task.GetCurrentBlockNumber())
			continue
		}
		segmentCount := e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
			storageParams.VersionedParams.GetMaxSegmentSize())
		for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
			pieceKey := e.baseApp.PieceOp().SegmentPieceKey(currentGCObjectID, segIdx)
			// ignore this delete api error, TODO: refine gc workflow by enrich metadata index.
			deleteErr := e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
			log.CtxDebugw(ctx, "delete the primary sp pieces", "object_info", objectInfo,
				"piece_key", pieceKey, "error", deleteErr)
		}
		bucketInfo, err := e.baseApp.GfSpClient().GetBucketByBucketName(ctx, objectInfo.BucketName, true)
		if err != nil || bucketInfo == nil {
			log.Errorw("failed to get bucket by bucket name", "error", err)
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
					pieceKey := e.baseApp.PieceOp().ECPieceKey(currentGCObjectID, segIdx, uint32(rIdx))
					if objectInfo.GetRedundancyType() == storagetypes.REDUNDANCY_REPLICA_TYPE {
						pieceKey = e.baseApp.PieceOp().SegmentPieceKey(objectInfo.Id.Uint64(), segIdx)
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
		groups     []*bsdb.GlobalVirtualGroup
		err        error
		startAfter uint64 = 0
		limit      uint32 = 100
		objects    []*metadatatypes.ObjectDetails
		gcNum      uint64 = 0
		bucketInfo *storagetypes.BucketInfo
	)
	bucketID := task.GetBucketID()

	reportProgress := func() bool {
		reportErr := e.ReportTask(ctx, task)
		log.CtxDebugw(ctx, "gc object task report progress", "task_info", task.Info(), "error", reportErr)
		return errors.Is(reportErr, manager.ErrCanceledTask)
	}
	defer func() {
		if err == nil { // succeed
			reportProgress()
		} else { // failed
			task.SetError(err)
			reportProgress()
		}
		log.CtxDebugw(ctx, "succeed to report gc bucket migration task", "task_info", task.Info(), "bucket_id", bucketID, "error", err)
	}()

	// list gvg
	if groups, err = e.baseApp.GfBsDB().ListGvgByBucketID(common.BigToHash(math.NewUint(bucketID).BigInt())); err != nil {
		log.CtxErrorw(ctx, "failed to list global virtual group by bucket id", "error", err)
		return
	}

	// current chain's gvg info compare metadata info
	if bucketInfo, err = e.baseApp.Consensus().QueryBucketInfoById(ctx, bucketID); err != nil || bucketInfo == nil {
		log.Errorw("failed to get bucket by bucket name", "error", err)
		return
	}
	vgfID := bucketInfo.GetGlobalVirtualGroupFamilyId()
	log.CtxInfow(ctx, "begin to gc bucket migration by bucket id", "bucket_id", bucketID, "gvgs", groups, "vgfID", vgfID, "error", err)

	for _, gvg := range groups {
		if gvg.FamilyId != vgfID {
			log.CtxErrorw(ctx, "failed to check gvg's status with chain's status, the gvg may be old data", "error", err)
			err = errors.New("gvg family id mismatch")
			return
		}
		for {
			if objects, err = e.baseApp.GfSpClient().ListObjectsByGVGAndBucketForGC(ctx, gvg.GlobalVirtualGroupId, bucketID, startAfter, limit); err != nil {
				log.CtxErrorw(ctx, "failed to list objects by gvg and bucket for gc", "error", err)
				return
			}
			// can delete, verify
			for _, obj := range objects {
				gcNum++
				objectInfo := obj.GetObject().GetObjectInfo()
				if e.gcWorker.checkGVGMatchSP(ctx, objectInfo, piecestore.PrimarySPRedundancyIndex) == ErrInvalidRedundancyIndex {
					err = e.gcWorker.deleteObjectSegmentsAndIntegrity(ctx, objectInfo)
					log.CtxInfow(ctx, "succeed to delete objects by gvg and bucket for gc", "object", objectInfo, "error", err)
				}
				if gcNum%reportProgressPerN == 0 {
					log.Infow("bucket migration gc report task", "current_gvg", gvg, "bucket_id", bucketID,
						"last_gc_object_id", obj.GetObject().GetObjectInfo().Id.Uint64())
					task.SetLastGCObjectID(obj.GetObject().GetObjectInfo().Id.Uint64())
					task.SetLastGCGvgID(gvg.ID)
					if err = e.ReportTask(ctx, task); err != nil { // report task is already automatically triggered.
						log.CtxErrorw(ctx, "failed to report bucket migration task gc progress", "task_info", task, "error", err)
						return
					}
				}
			}

			if len(objects) < int(limit) {
				log.CtxInfow(ctx, "succeed to finish one gvg bucket migration gc", "gvg", gvg)
				break
			}
		}
	}
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
					if strings.Contains(err.Error(), "No such object") {
						// 1) This object does not exist on the chain
						e.gcWorker.deleteObjectPiecesAndIntegrityMeta(ctx, integrityObject)
					}
				} else {
					// 2) query metadata error, but chain has the object info, gvg  primary sp should have integrity meta
					if e.gcWorker.checkGVGMatchSP(ctx, objInfoFromChain, integrityObject.RedundancyIndex) == ErrInvalidRedundancyIndex {
						e.gcWorker.deleteObjectPiecesAndIntegrityMeta(ctx, integrityObject)
					}
					continue
				}
			}
		} else {
			// 3) check integrity meta & object info
			if e.gcWorker.checkGVGMatchSP(ctx, objInfoFromMetaData, integrityObject.RedundancyIndex) == ErrInvalidRedundancyIndex {
				e.gcWorker.deleteObjectPiecesAndIntegrityMeta(ctx, integrityObject)
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
					// 2) If there is an error querying metadata but the chain contains object information, recheck the meta.
					if e.gcWorker.checkGVGMatchSP(ctx, objInfoFromChain, piece.RedundancyIndex) == ErrInvalidRedundancyIndex {
						e.gcWorker.deletePieceAndPieceChecksum(ctx, piece)
					}
				}
			}
		} else {
			// 3) Validate using Metadata information.
			if e.gcWorker.checkGVGMatchSP(ctx, objectInfoFromMetadata, piece.RedundancyIndex) == ErrInvalidRedundancyIndex {
				e.gcWorker.deletePieceAndPieceChecksum(ctx, piece)
			}
		}
	}
	return nil
}
