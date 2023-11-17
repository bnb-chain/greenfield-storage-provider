package executor

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
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
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/forbole/juno/v4/common"
	"github.com/prysmaticlabs/prysm/crypto/bls"
)

var (
	ErrDanglingPointer              = gfsperrors.Register(module.ExecuteModularName, http.StatusBadRequest, 40001, "OoooH.... request lost")
	ErrInsufficientApproval         = gfsperrors.Register(module.ExecuteModularName, http.StatusNotFound, 40002, "insufficient approvals from p2p")
	ErrUnsealed                     = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 40003, "seal object on chain failed")
	ErrExhaustedApproval            = gfsperrors.Register(module.ExecuteModularName, http.StatusNotFound, 40004, "approvals exhausted")
	ErrInvalidIntegrity             = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40005, "secondary integrity hash verification failed")
	ErrSecondaryMismatch            = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40006, "secondary sp mismatch")
	ErrReplicateIdsOutOfBounds      = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40007, "replicate idx out of bounds")
	ErrRecoveryRedundancyType       = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45202, "recovery only support EC redundancy type")
	ErrRecoveryPieceNotEnough       = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45203, "failed to get enough piece data to recovery")
	ErrRecoveryDecode               = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45204, "EC decode error")
	ErrRecoveryPieceChecksum        = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45206, "recovery checksum not correct")
	ErrRecoveryPieceLength          = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45207, "get secondary piece data length error")
	ErrPrimaryNotFound              = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 45208, "primary sp endpoint not found when recovering")
	ErrRecoveryPieceIndex           = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 45209, "recovery piece index invalid")
	ErrMigratedPieceChecksum        = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45210, "migrate piece checksum is not correct")
	ErrInvalidRedundancyIndex       = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45212, "invalid redundancy index")
	ErrSetObjectIntegrity           = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45213, "failed to set object integrity into spdb")
	ErrInvalidPieceChecksumLength   = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45214, "invalid piece checksum length")
	ErrRecoveryObjectStatus         = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 45215, "the recovered object has not been sealed state")
	ErrInvalidSecondaryBlsSignature = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 45216, "primary receive invalid bls signature from secondary SP")
)

func ErrGfSpDBWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45201, detail)
}

func ErrPieceStoreWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 45205, detail)
}

func ErrConsensusWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.ExecuteModularName, http.StatusBadRequest, 45211, detail)
}

type GCChecker struct {
	e *ExecuteModular
}

// NewGCChecker returns a family conflicted checker instance.
func NewGCChecker(e *ExecuteModular) *GCChecker {
	return &GCChecker{
		e: e,
	}
}

// deleteObjectPiecesAndIntegrityMetaByObjectInfo used by gcZombiePiece
func (gc *GCChecker) deleteObjectPiecesAndIntegrityMeta(ctx context.Context, integrityMeta *corespdb.IntegrityMeta) error {
	objID := integrityMeta.ObjectID
	redundancyIdx := integrityMeta.RedundancyIndex
	maxSegment := len(integrityMeta.PieceChecksumList)

	// delete object pieces
	for segmentIdx := uint32(0); segmentIdx <= uint32(maxSegment); segmentIdx++ {
		gc.deletePiece(ctx, objID, segmentIdx, redundancyIdx)
	}

	// delete integrity meta
	_ = gc.e.baseApp.GfSpDB().DeleteObjectIntegrity(objID, redundancyIdx)
	log.CtxDebugw(ctx, "succeed to delete all object segment and integrity meta", "object_id", objID, "integrity_meta", integrityMeta)

	return nil
}

func (gc *GCChecker) deleteObjectSegments(ctx context.Context, objectInfo *storagetypes.ObjectInfo) error {
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
		// ignore this delete api error, TODO: refine gc workflow by enrich metadata index.
		deleteErr := gc.e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
		log.CtxDebugw(ctx, "succeed to delete the primary sp segment", "object_info", objectInfo,
			"piece_key", pieceKey, "error", deleteErr)
	}
	return nil
}

// deletePiece delete single piece if meta data or chain has object info
func (gc *GCChecker) deletePiece(ctx context.Context, objID uint64, segmentIdx uint32, redundancyIdx int32) {
	var pieceKey string
	// TODO -1 means REDUNDANCY_EC_TYPE
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
func (gc *GCChecker) deletePieceAndPieceChecksum(ctx context.Context, piece *spdb.GCPieceMeta) error {
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
func (gc *GCChecker) isAllowGCCheck(objectInfo *storagetypes.ObjectInfo, bucketInfo *metadatatypes.Bucket) bool {
	// object not in sealed status
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.Infow("the object isn't sealed, do not need to gc check")
		return false
	}
	// bucket migrating
	if bucketInfo.GetBucketInfo().GetBucketStatus() == storagetypes.BUCKET_STATUS_MIGRATING {
		log.Infow("bucket is migrating, do not need to gc check")
		return false
	}
	log.Infow("the object is sealed and the bucket is not migrating, the object can gc", "object", objectInfo, "bucket", bucketInfo)
	return true
}

func (gc *GCChecker) getGvgAndSpId(ctx context.Context, objectInfo *storagetypes.ObjectInfo) (*metadatatypes.Bucket, *virtualgrouptypes.GlobalVirtualGroup, uint32, error) {
	bucketInfo, err := gc.e.baseApp.GfSpClient().GetBucketByBucketName(ctx, objectInfo.BucketName, true)
	if err != nil || bucketInfo == nil {
		log.Errorw("failed to get bucket by bucket name", "error", err)
		return nil, nil, 0, err
	}

	// piece
	gvg, err := gc.e.baseApp.GfSpClient().GetGlobalVirtualGroup(ctx, bucketInfo.BucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
	if err != nil {
		log.Errorw("failed to get global virtual group", "error", err)
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
func (gc *GCChecker) checkGVGMatchSP(ctx context.Context, objectInfo *storagetypes.ObjectInfo, redundancyIndex int32) error {
	bucketInfo, gvg, spID, err := gc.getGvgAndSpId(ctx, objectInfo)
	if err != nil {
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

func (e *ExecuteModular) HandleSealObjectTask(ctx context.Context, task coretask.SealObjectTask) {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to handle seal object, task pointer dangling")
		task.SetError(ErrDanglingPointer)
		return
	}
	blsSig, err := bls.MultipleSignaturesFromBytes(task.GetSecondarySignatures())
	if err != nil {
		return
	}
	task.AppendLog("executor-begin-handle-seal-task")
	sealMsg := &storagetypes.MsgSealObject{
		Operator:                    e.baseApp.OperatorAddress(),
		BucketName:                  task.GetObjectInfo().GetBucketName(),
		ObjectName:                  task.GetObjectInfo().GetObjectName(),
		GlobalVirtualGroupId:        task.GetGlobalVirtualGroupId(),
		SecondarySpBlsAggSignatures: bls.AggregateSignatures(blsSig).Marshal(),
	}
	task.SetError(e.sealObject(ctx, task, sealMsg))
	metrics.PerfPutObjectTime.WithLabelValues("seal_object_total_time_from_uploading_to_sealing").Observe(time.Since(
		time.Unix(task.GetObjectInfo().GetCreateAt(), 0)).Seconds())
	task.AppendLog("executor-end-handle-seal-task")
	log.CtxDebugw(ctx, "finished to handle seal object task", "error", task.Error())
}

func (e *ExecuteModular) sealObject(ctx context.Context, task coretask.ObjectTask, sealMsg *storagetypes.MsgSealObject) error {
	var (
		err    error
		txHash string
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ExecutorCounter.WithLabelValues(ExecutorFailureSealObject).Inc()
			metrics.ExecutorTime.WithLabelValues(ExecutorFailureSealObject).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ExecutorCounter.WithLabelValues(ExecutorSuccessSealObject).Inc()
			metrics.ExecutorTime.WithLabelValues(ExecutorSuccessSealObject).Observe(time.Since(startTime).Seconds())
		}
	}()
	for retry := int64(0); retry <= task.GetMaxRetry(); retry++ {
		txHash, err = e.baseApp.GfSpClient().SealObject(ctx, sealMsg)
		if err != nil {
			task.AppendLog(fmt.Sprintf("executor-seal-tx-failed-error:%s-retry:%d", err.Error(), retry))
			log.CtxErrorw(ctx, "failed to seal object", "retry", retry, "max_retry", task.GetMaxRetry(),
				"error", err)
			time.Sleep(time.Duration(e.listenSealRetryTimeout) * time.Second)
		} else {
			task.AppendLog(fmt.Sprintf("executor-seal-tx-succeed-retry:%d-txHash:%s", retry, txHash))
			err = nil
			break
		}
	}
	// even though signer return error, maybe seal on chain successfully because
	// signer use the async mode, so ignore the error and listen directly
	err = e.listenSealObject(ctx, task, task.GetObjectInfo())
	return err
}

func (e *ExecuteModular) listenSealObject(ctx context.Context, task coretask.ObjectTask, object *storagetypes.ObjectInfo) error {
	var (
		err    error
		sealed bool
	)
	for retry := 0; retry < e.maxListenSealRetry; retry++ {
		sealed, err = e.baseApp.Consensus().ListenObjectSeal(ctx, object.Id.Uint64(), e.listenSealTimeoutHeight)
		if err != nil {
			task.AppendLog(fmt.Sprintf("executor-listen-seal-failed-error:%s-retry:%d", err.Error(), retry))
			log.CtxErrorw(ctx, "failed to listen object seal", "retry", retry,
				"max_retry", e.maxListenSealRetry, "error", err)
			time.Sleep(time.Duration(e.listenSealRetryTimeout) * time.Second)
			continue
		}
		if !sealed {
			task.AppendLog(fmt.Sprintf("executor-listen-seal-failed(unseal)-retry:%d", retry))
			log.CtxErrorw(ctx, "failed to seal object on chain", "retry", retry,
				"max_retry", e.maxListenSealRetry, "error", err)
			err = ErrUnsealed
			continue
		}
		task.AppendLog(fmt.Sprintf("executor-listen-seal-succeed-retry:%d", retry))
		err = nil
		break
	}
	return err
}

func (e *ExecuteModular) HandleReceivePieceTask(ctx context.Context, task coretask.ReceivePieceTask) {
	if task.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to handle receive piece confirm, task pointer dangling")
		return
	}
	var (
		err            error
		offChainObject *storagetypes.ObjectInfo
	)
	offChainObject, err = e.baseApp.GfSpClient().GetObjectByID(ctx, task.GetObjectInfo().Id.Uint64())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get object info", "error", err)
		task.SetError(err)
		return
	}
	if offChainObject.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxError(ctx, "failed to confirm receive task, object is unsealed")
		task.SetError(ErrUnsealed)
		return
	}

	// regardless of whether the sp is as secondary or not, it needs to be set to the
	// sealed state to let the manager clear the task.
	task.SetSealed(true)
	// TODO: might add api GetGvgByObjectId in meta
	bucketInfo, err := e.baseApp.GfSpClient().GetBucketByBucketName(ctx, offChainObject.BucketName, true)
	if err != nil || bucketInfo == nil {
		log.Errorw("failed to get bucket by bucket name", "error", err)
		return
	}
	gvg, err := e.baseApp.GfSpClient().GetGlobalVirtualGroup(ctx, bucketInfo.BucketInfo.Id.Uint64(), offChainObject.LocalVirtualGroupId)
	if err != nil {
		log.Errorw("failed to get global virtual group", "error", err)
		return
	}
	if int(task.GetRedundancyIdx()) >= len(gvg.GetSecondarySpIds()) {
		log.CtxErrorw(ctx, "failed to confirm receive task, replicate idx out of bounds",
			"redundancy_idx", task.GetRedundancyIdx(), "secondary_sp_len", len(gvg.GetSecondarySpIds()))
		task.SetError(ErrReplicateIdsOutOfBounds)
		return
	}

	spID, err := e.getSPID()
	if err != nil {
		log.Errorw("failed to get sp id", "error", err)
		return
	}
	if gvg.GetSecondarySpIds()[int(task.GetRedundancyIdx())] != spID {
		log.CtxErrorw(ctx, "failed to confirm receive task, secondary sp mismatch", "expect",
			gvg.GetSecondarySpIds()[int(task.GetRedundancyIdx())], "current", e.baseApp.OperatorAddress())
		task.SetError(ErrSecondaryMismatch)
		err = e.baseApp.GfSpDB().DeleteObjectIntegrity(task.GetObjectInfo().Id.Uint64(), task.GetRedundancyIdx())
		if err != nil {
			log.CtxError(ctx, "failed to delete integrity")
		}
		var pieceKey string
		segmentCount := e.baseApp.PieceOp().SegmentPieceCount(offChainObject.GetPayloadSize(),
			task.GetStorageParams().GetMaxPayloadSize())
		for i := uint32(0); i < segmentCount; i++ {
			if task.GetObjectInfo().GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
				pieceKey = e.baseApp.PieceOp().ECPieceKey(offChainObject.Id.Uint64(), i, uint32(task.GetRedundancyIdx()))
			} else {
				pieceKey = e.baseApp.PieceOp().SegmentPieceKey(offChainObject.Id.Uint64(), i)
			}
			err = e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
			if err != nil {
				log.CtxErrorw(ctx, "failed to delete piece data", "piece_key", pieceKey)
			}
		}
		return
	}
	log.CtxDebug(ctx, "succeed to handle confirm receive piece task")
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
		log.Error("no waiting gc objects")
		hasNoObject = true
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
		// TODO get sp id from config
		spId, err := e.getSPID()
		if err != nil {
			log.Errorw("failed to get sp id", "error", err)
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

	reportProgress := func() bool {
		reportErr := e.ReportTask(ctx, task)
		log.CtxDebugw(ctx, "gc zombie piece task report progress", "task_info", task.Info(), "error", reportErr)
		return errors.Is(reportErr, manager.ErrCanceledTask)
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
	reportProgress := func() bool {
		reportErr := e.ReportTask(ctx, task)
		log.CtxDebugw(ctx, "gc meta task report progress", "task_info", task.Info(), "error", reportErr)
		return errors.Is(reportErr, manager.ErrCanceledTask)
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
	// TODO list buckets when large dataset
	now := time.Now()
	daysAgo := now.Add(-time.Duration(e.bucketTrafficKeepLatestDay) * time.Hour * 24)
	yearMonth := sqldb.TimeToYearMonth(daysAgo)

	err := e.baseApp.GfSpDB().DeleteAllBucketTrafficExpired(yearMonth)
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

	err := e.baseApp.GfSpDB().DeleteAllReadRecordExpired(uint64(daysAgo.UnixMicro()))
	if err != nil {
		log.CtxErrorw(ctx, "failed to delete expired read record", "error", err)
		return err
	}
	log.CtxInfow(ctx, "succeed to delete read record", "task", task, "daysAgo", daysAgo)
	return nil
}

func (e *ExecuteModular) HandleGCBucketMigrationBucket(ctx context.Context, task coretask.GCBucketMigrationTask) {
	// TODO gc progress persist in db
	var (
		groups     []*bsdb.GlobalVirtualGroup
		err        error
		startAfter uint64 = 0
		limit      uint32
		objects    []*metadatatypes.ObjectDetails
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
	groups, err = e.baseApp.GfBsDB().ListGvgByBucketID(common.BigToHash(math.NewUint(bucketID).BigInt()))
	if err != nil {
		log.CtxErrorw(ctx, "failed to list global virtual group by bucket id", "error", err)
		return
	}

	// current chain's gvg info compare metadata info
	bucketInfo, err := e.baseApp.Consensus().QueryBucketInfoById(ctx, bucketID)
	if err != nil || bucketInfo == nil {
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
		objects, err = e.baseApp.GfSpClient().ListObjectsByGVGAndBucketForGC(ctx, gvg.GlobalVirtualGroupId, bucketID, startAfter, limit)
		if err != nil {
			log.CtxErrorw(ctx, "failed to list objects by gvg and bucket for gc", "error", err)
			return
		}
		// can delete, verify
		for _, obj := range objects {
			objectInfo := obj.GetObject().GetObjectInfo()
			if e.gcChecker.checkGVGMatchSP(ctx, objectInfo, piecestore.PrimarySPRedundancyIndex) == ErrInvalidRedundancyIndex {
				e.gcChecker.deleteObjectSegments(ctx, objectInfo)
				log.CtxInfow(ctx, "succeed to delete objects by gvg and bucket for gc", "object", objectInfo, "error", err)
			}
		}
	}

}

// HandleRecoverPieceTask handle the recovery piece task, it will send request to other SPs to get piece data to recovery,
// recovery the original data, and write the recovered data to piece store
func (e *ExecuteModular) HandleRecoverPieceTask(ctx context.Context, task coretask.RecoveryPieceTask) {
	var (
		dataShards         = task.GetStorageParams().VersionedParams.GetRedundantDataChunkNum()
		parityShards       = task.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()
		ecPieceCount       = dataShards + parityShards
		recoveryKey        string
		recoveryMinEcIndex = -1
		err                error
		finishRecovery     = false
	)
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		if task.Error() != nil {
			log.CtxErrorw(ctx, "recovery task failed", "error", task.Error())
		}
		if finishRecovery {
			task.SetRecoverDone()
		}
	}()

	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		err = ErrDanglingPointer
		return
	}

	if task.GetObjectInfo().ObjectStatus != storagetypes.OBJECT_STATUS_SEALED {
		err = ErrRecoveryObjectStatus
		return
	}

	if task.GetObjectInfo().GetRedundancyType() != storagetypes.REDUNDANCY_EC_TYPE {
		err = ErrRecoveryRedundancyType
		return
	}

	redundancyIdx := task.GetEcIdx()
	maxRedundancyIndex := int32(ecPieceCount) - 1

	if redundancyIdx < int32(recoveryMinEcIndex) || redundancyIdx > maxRedundancyIndex {
		err = ErrRecoveryPieceIndex
		return
	}

	if redundancyIdx >= 0 {
		log.Info(1)
		// recover secondary SP data by the primary SP
		if err = e.recoverByPrimarySP(ctx, task); err != nil {
			// if failed to recover by the primary SP, try to recovery secondary SP data from the other secondary SPs
			recoverErr := e.recoverBySecondarySP(ctx, task, true)
			if recoverErr != nil {
				err = recoverErr
				return
			}
		}

		log.CtxDebugw(ctx, "secondary SP recovery successfully", "pieceKey:", recoveryKey)
		finishRecovery = true
	} else {
		// recover primarySP data by secondary SPs
		if recoverErr := e.recoverBySecondarySP(ctx, task, false); recoverErr != nil {
			err = recoverErr
			return
		}
		log.CtxDebugw(ctx, "primary SP recovery successfully", "pieceKey:", recoveryKey)
		finishRecovery = true
	}
}

// recoverByPrimarySP recover secondary SP by the corresponding primary SP
func (e *ExecuteModular) recoverByPrimarySP(ctx context.Context, task coretask.RecoveryPieceTask) error {
	log.CtxDebugw(ctx, "begin to recovery by the primary SP", "objectName:", task.GetObjectInfo().GetObjectName())
	var (
		err               error
		primarySPEndpoint string
		pieceData         []byte
	)
	objectId := task.GetObjectInfo().Id.Uint64()
	segmentIdx := task.GetSegmentIdx()
	primarySPEndpoint, err = e.getBucketPrimarySPEndpoint(ctx, task.GetObjectInfo().BucketName)
	if err != nil {
		return err
	}

	pieceData, err = e.doRecoveryPiece(ctx, task, primarySPEndpoint)
	if err != nil {
		log.CtxDebugw(ctx, "failed to recover secondary SP data from secondary SPs")
		return err
	}
	// compare integrity hash
	if err = e.checkRecoveryChecksum(ctx, task, hash.GenerateChecksum(pieceData)); err != nil {
		return err
	}

	recoveryKey := e.baseApp.PieceOp().ECPieceKey(objectId, segmentIdx, uint32(task.GetEcIdx()))
	// write the recovery segment key to keystore
	err = e.baseApp.PieceStore().PutPiece(ctx, recoveryKey, pieceData)
	if err != nil {
		log.CtxErrorw(ctx, "EC recover data write piece fail", "pieceKey:", recoveryKey, "error", err)
		return err
	}

	return nil
}

// recoverBySecondarySP recovery primarySP or recovery Secondary from secondary SPs
func (e *ExecuteModular) recoverBySecondarySP(ctx context.Context, task coretask.RecoveryPieceTask, isMyselfSecondary bool) error {
	log.CtxDebugw(ctx, "begin to recovery from secondary SPs", "objectName:", task.GetObjectInfo().GetObjectName())
	var (
		dataShards         = task.GetStorageParams().VersionedParams.GetRedundantDataChunkNum()
		maxSegmentSize     = task.GetStorageParams().VersionedParams.GetMaxSegmentSize()
		parityShards       = task.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()
		minRecoveryPieces  = dataShards
		ecPieceCount       = dataShards + parityShards
		doneTaskNum        = uint32(0)
		err                error
		secondaryEndpoints []string
		secondaryCount     int
		totalTaskNum       int32
		executeEndpoint    string
		recoveredPieceData []byte
		recoveryKey        string
	)

	secondaryEndpoints, secondaryCount, err = e.getObjectSecondaryEndpoints(ctx, task.GetObjectInfo())
	if err != nil {
		return err
	}

	if uint32(secondaryCount) != ecPieceCount {
		return ErrRecoveryPieceNotEnough
	}

	var recoveryDataSources = make([][]byte, secondaryCount)
	doneCh := make(chan bool, secondaryCount)
	quitCh := make(chan bool)

	totalTaskNum = int32(secondaryCount)
	if isMyselfSecondary {
		totalTaskNum = totalTaskNum - 1
	}
	downLoadPieceSize := 0
	segmentSize := e.baseApp.PieceOp().SegmentPieceSize(task.GetObjectInfo().PayloadSize, task.GetSegmentIdx(), maxSegmentSize)

	if isMyselfSecondary {
		operator := e.baseApp.OperatorAddress()
		spInfo, dbErr := e.baseApp.GfSpDB().GetSpByAddress(operator, spdb.OperatorAddressType)
		if dbErr != nil {
			return dbErr
		}

		executeEndpoint = spInfo.Endpoint
	}

	for ecIdx := 0; ecIdx < secondaryCount; ecIdx++ {
		recoveryDataSources[ecIdx] = nil
		go func(secondaryIndex int) {
			secondaryEndpoint := secondaryEndpoints[secondaryIndex]
			// if myself is secondary, bypass to send request to myself
			if isMyselfSecondary && secondaryEndpoint == executeEndpoint {
				if atomic.AddInt32(&totalTaskNum, -1) == 0 {
					quitCh <- true
				}
				return
			}
			pieceData, recoverErr := e.doRecoveryPiece(ctx, task, secondaryEndpoints[secondaryIndex])
			if recoverErr == nil {
				recoveryDataSources[secondaryIndex] = pieceData
				log.Debugf("get one piece from ", "piece length:%d ", len(pieceData), "secondary sp:", secondaryEndpoints[secondaryIndex])
				doneCh <- true
				downLoadPieceSize = len(pieceData)
			}
			// finish all the task, send signal to quitCh
			if atomic.AddInt32(&totalTaskNum, -1) == 0 {
				quitCh <- true
			}
		}(ecIdx)
	}

loop:
	for {
		select {
		case <-doneCh:
			doneTaskNum++
			// it is enough to recovery data with minRecoveryPieces EC data, no need to wait
			if doneTaskNum >= minRecoveryPieces {
				break loop
			}
		case <-quitCh: // all the task finish
			if doneTaskNum < minRecoveryPieces { // finish task num not enough
				log.CtxErrorw(ctx, "get piece from secondary not enough", "get secondary piece num:", doneTaskNum, "error", ErrRecoveryPieceNotEnough)
				return ErrRecoveryPieceNotEnough
			}
			ecTotalSize := int64(uint32(downLoadPieceSize) * dataShards)
			if ecTotalSize < segmentSize || ecTotalSize > segmentSize+int64(dataShards) {
				log.CtxErrorw(ctx, "get secondary piece data length error")
				return ErrRecoveryPieceLength
			}
		}
	}

	recoverySegData, recoverErr := redundancy.DecodeRawSegment(recoveryDataSources, segmentSize, int(dataShards), int(parityShards))
	if recoverErr != nil {
		log.CtxErrorw(ctx, "EC decode error when recovery", "objectName:", task.GetObjectInfo().ObjectName, "segIndex:", task.GetSegmentIdx(), "error", err)
		return ErrRecoveryDecode
	}

	// compare integrity hash
	if !isMyselfSecondary {
		if err = e.checkRecoveryChecksum(ctx, task, hash.GenerateChecksum(recoverySegData)); err != nil {
			return err
		}
		// if the task is generated by primary SP, the recovery key is segment
		recoveryKey = e.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), task.GetSegmentIdx())
		recoveredPieceData = recoverySegData
	} else {
		redundancyIdx := task.GetEcIdx()
		// if the task is generated by a secondary SP, the recovery key is EC piece
		recoveryKey = e.baseApp.PieceOp().ECPieceKey(task.GetObjectInfo().Id.Uint64(), task.GetSegmentIdx(), uint32(redundancyIdx))

		recoveredPieceData, err = e.getECPieceBySegment(ctx, task.GetEcIdx(), task.GetObjectInfo(), task.GetStorageParams(), recoverySegData, task.GetSegmentIdx())
		if err != nil {
			return err
		}
		// compare integrity hash
		if err = e.checkRecoveryChecksum(ctx, task, hash.GenerateChecksum(recoveredPieceData)); err != nil {
			return err
		}
	}

	// write the recovery segment key to keystore
	if err = e.baseApp.PieceStore().PutPiece(ctx, recoveryKey, recoveredPieceData); err != nil {
		log.CtxErrorw(ctx, "EC decode data write piece fail", "pieceKey:", recoveryKey, "error", err)
		return err
	}

	log.CtxDebugw(ctx, "finish recovery from secondary SPs", "objectName:", task.GetObjectInfo().GetObjectName())
	return nil
}

// getECPieceBySegment return the EC encodes data based on the redundancyIdx and the segment data
func (e *ExecuteModular) getECPieceBySegment(ctx context.Context, redundancyIdx int32, objectInfo *storagetypes.ObjectInfo,
	params *storagetypes.Params, recoverySegData []byte, segmentIdx uint32) ([]byte, error) {
	dataShards := params.GetRedundantDataChunkNum()
	parityShards := params.GetRedundantParityChunkNum()
	if redundancyIdx < 0 || redundancyIdx > int32(dataShards+parityShards-1) {
		return nil, fmt.Errorf("invalid redundancyIdx")
	}
	// if it is the data shards of ec-encoded pieces, just get the ec data by offset
	if redundancyIdx > 0 && redundancyIdx < int32(dataShards)-1 {
		ECPieceSize := e.baseApp.PieceOp().ECPieceSize(objectInfo.PayloadSize, segmentIdx, params.GetMaxSegmentSize(), params.GetRedundantDataChunkNum())

		startPos := int64(redundancyIdx) * ECPieceSize
		endPos := int64(redundancyIdx+1)*ECPieceSize - 1
		return recoverySegData[startPos:endPos], nil
	}

	// if it is the parity shard, it needs to encode again to compute the parity shards
	ECEncodeData, err := redundancy.EncodeRawSegment(recoverySegData, int(dataShards), int(parityShards))
	if err != nil {
		log.CtxErrorw(ctx, "failed to ec encode data when recovering secondary SP", "error", err)
		return nil, err
	}
	return ECEncodeData[redundancyIdx], nil
}

func (e *ExecuteModular) checkRecoveryChecksum(ctx context.Context, task coretask.RecoveryPieceTask, recoveryChecksum []byte) error {
	integrityMeta, err := e.baseApp.GfSpDB().GetObjectIntegrity(task.GetObjectInfo().Id.Uint64(), task.GetEcIdx())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get object integrity hash in db when recovery", "objectName:",
			task.GetObjectInfo().ObjectName, "error", err)
		return ErrGfSpDBWithDetail("failed to get object integrity hash in db when recovery, objectName: " +
			task.GetObjectInfo().ObjectName + ",error: " + err.Error())
	}

	expectedHash := integrityMeta.PieceChecksumList[task.GetSegmentIdx()]
	if !bytes.Equal(recoveryChecksum, expectedHash) {
		log.CtxErrorw(ctx, "check integrity hash of recovery data err", "objectName:", task.GetObjectInfo().ObjectName,
			"expected value", hex.EncodeToString(expectedHash), "actual value", recoveryChecksum, "error", ErrRecoveryPieceChecksum)
		return ErrRecoveryPieceChecksum
	}
	return nil
}

func (e *ExecuteModular) doRecoveryPiece(ctx context.Context, rTask coretask.RecoveryPieceTask, endpoint string) (
	data []byte, err error) {
	var (
		signature []byte
		pieceData []byte
	)
	signature, err = e.baseApp.GfSpClient().SignRecoveryTask(ctx, rTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign recovery task", "object", rTask.GetObjectInfo().GetObjectName(), "error", err)
		return
	}
	rTask.SetSignature(signature)
	// recovery primary sp segment or secondary piece
	respBody, err := e.baseApp.GfSpClient().GetPieceFromECChunks(ctx, endpoint, rTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get piece from ec chunks", "objectID", rTask.GetObjectInfo().Id,
			"segment_idx", rTask.GetSegmentIdx(), "secondary endpoint", endpoint, "error", err)
		return
	}

	defer respBody.Close()
	pieceData, err = io.ReadAll(respBody)
	if err != nil {
		log.CtxErrorw(ctx, "failed to read recovery piece data from sp", "objectID", rTask.GetObjectInfo().Id,
			"segment idx", rTask.GetSegmentIdx(), "secondary endpoint", endpoint, "error", err)
		return nil, err
	}

	log.CtxDebugw(ctx, "succeed to recovery piece from sp", "objectID", rTask.GetObjectInfo().Id,
		"segment_idx", rTask.GetSegmentIdx(), "secondary endpoint", endpoint)

	return pieceData, nil
}

// getObjectSecondaryEndpoints return the secondary sp endpoints list of the specific object
func (e *ExecuteModular) getObjectSecondaryEndpoints(ctx context.Context, objectInfo *storagetypes.ObjectInfo) ([]string, int, error) {
	// TODO: might add api GetGvgByObjectID in meta
	bucketInfo, err := e.baseApp.GfSpClient().GetBucketByBucketName(ctx, objectInfo.BucketName, true)
	if err != nil {
		log.Errorf("failed to get bucket by bucket name", "error", err)
		return nil, 0, err
	}
	gvg, err := e.baseApp.GfSpClient().GetGlobalVirtualGroup(ctx, bucketInfo.BucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
	if err != nil {
		return nil, 0, err
	}
	secondarySPIds := gvg.GetSecondarySpIds()

	spList, err := e.baseApp.Consensus().ListSPs(ctx)
	if err != nil {
		return nil, 0, err
	}
	var secondaryCount int
	secondaryEndpointList := make([]string, len(secondarySPIds))
	for idx, sspId := range secondarySPIds {
		for _, info := range spList {
			if sspId == info.Id {
				secondaryCount++
				secondaryEndpointList[idx] = info.Endpoint
			}
		}
	}

	return secondaryEndpointList, secondaryCount, nil
}

func (e *ExecuteModular) getBucketPrimarySPEndpoint(ctx context.Context, bucketName string) (string, error) {
	bucketMeta, _, err := e.baseApp.GfSpClient().GetBucketMeta(ctx, bucketName, true)
	if err != nil {
		return "", err
	}
	bucketSPID, err := util.GetBucketPrimarySPID(ctx, e.baseApp.Consensus(), bucketMeta.GetBucketInfo())
	if err != nil {
		return "", err
	}
	spList, err := e.baseApp.Consensus().ListSPs(ctx)
	if err != nil {
		return "", err
	}
	for _, info := range spList {
		if bucketSPID == info.Id {
			return info.Endpoint, nil
		}
	}
	return "", ErrPrimaryNotFound
}

func (e *ExecuteModular) gcZombiePieceFromIntegrityMeta(ctx context.Context, task coretask.GCZombiePieceTask) error {
	var (
		err                            error
		waitingVerifyGCIntegrityPieces []*corespdb.IntegrityMeta
		objInfoFromMetaData            *storagetypes.ObjectInfo
		objInfoFromChain               *storagetypes.ObjectInfo
	)

	if waitingVerifyGCIntegrityPieces, err = e.baseApp.GfSpDB().ListIntegrityMetaByObjectIDRange(int64(task.GetStartObjectId()), int64(task.GetEndObjectId()), true); err != nil {
		log.CtxErrorw(ctx, "failed to query deleted object list", "task_info", task.Info(), "error", err)
		return err
	}

	if len(waitingVerifyGCIntegrityPieces) == 0 {
		log.CtxInfow(ctx, "no waiting gc objects", "task_info", task.Info())
		return nil
	}

	for _, integrityObject := range waitingVerifyGCIntegrityPieces {
		log.CtxDebugw(ctx, "gc zombie current waiting verify gc integrity meta piece", "integrity_meta", integrityObject)
		// error
		objID := integrityObject.ObjectID
		// If metadata service has object info, use objInfoFromMetaData, otherwise query from chain
		if objInfoFromMetaData, err = e.baseApp.GfSpClient().GetObjectByID(ctx, objID); err != nil {
			log.Errorf("failed to get object meta from meta data service, will check from chain", "error", err)
			if strings.Contains(err.Error(), "no such object from metadata") {
				// could delete, has integrity hash, do not have object info, chain check verify check chain
				if objInfoFromChain, err = e.baseApp.Consensus().QueryObjectInfoByID(ctx, strconv.FormatUint(objID, 10)); err != nil {
					if strings.Contains(err.Error(), "No such object") {
						// 1) This object does not exist on the chain
						e.gcChecker.deleteObjectPiecesAndIntegrityMeta(ctx, integrityObject)
					}
				} else {
					// 2) query metadata error, but chain has the object info, gvg  primary sp should has integrity meta
					if e.gcChecker.checkGVGMatchSP(ctx, objInfoFromChain, integrityObject.RedundancyIndex) == ErrInvalidRedundancyIndex {
						e.gcChecker.deleteObjectPiecesAndIntegrityMeta(ctx, integrityObject)
					}
					continue
				}
			}
		} else {
			// check integrity meta & object info
			if e.gcChecker.checkGVGMatchSP(ctx, objInfoFromMetaData, integrityObject.RedundancyIndex) == ErrInvalidRedundancyIndex {
				e.gcChecker.deleteObjectPiecesAndIntegrityMeta(ctx, integrityObject)
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

	// replicate piece must be secondary sp
	if waitingVerifyGCPieces, err = e.baseApp.GfSpDB().ListReplicatePieceChecksumByObjectIDRange(
		int64(task.GetStartObjectId()), int64(task.GetEndObjectId())); err != nil {
		log.CtxErrorw(ctx, "failed to query replicate piece checksum", "task_info", task.Info(), "error", err)
		return err
	}

	if len(waitingVerifyGCPieces) == 0 {
		log.CtxErrorw(ctx, "no waiting gc pieces", "task_info", task.Info())
		return nil
	}

	for _, piece := range waitingVerifyGCPieces {
		log.CtxDebugw(ctx, "gc zombie current waiting verify gc meta piece", "piece", piece)
		// error
		// migrating
		// sealed ? timeout ?
		objID := piece.ObjectID
		// object info
		if objectInfoFromMetadata, err = e.baseApp.GfSpClient().GetObjectByID(ctx, objID); err != nil {
			// check from chain delete ?
			if strings.Contains(err.Error(), "no such object from metadata") {
				log.Infof("failed to get object meta, the zombie piece should be deleted", "error", err)
				if objInfoFromChain, err = e.baseApp.Consensus().QueryObjectInfoByID(ctx, strconv.FormatUint(objID, 10)); err != nil {
					if strings.Contains(err.Error(), "No such object") {
						// 1) This object does not exist on the chain
						e.gcChecker.deletePieceAndPieceChecksum(ctx, piece)
					}
				} else {
					// 2) query metadata error, but chain has the object info, gvg  primary sp should has integrity meta
					// TODO: refine
					if e.gcChecker.checkGVGMatchSP(ctx, objInfoFromChain, piece.RedundancyIndex) == ErrInvalidRedundancyIndex {
						e.gcChecker.deletePieceAndPieceChecksum(ctx, piece)
					}
				}
			}
		} else {
			if e.gcChecker.checkGVGMatchSP(ctx, objectInfoFromMetadata, piece.RedundancyIndex) == ErrInvalidRedundancyIndex {
				//task.SetError(ErrSecondaryMismatch)
				// ignore error
				e.gcChecker.deletePieceAndPieceChecksum(ctx, piece)
			}
		}
	}
	return nil
}
