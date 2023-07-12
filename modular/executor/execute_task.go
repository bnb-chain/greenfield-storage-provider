package executor

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/modular/manager"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingPointer         = gfsperrors.Register(module.ExecuteModularName, http.StatusBadRequest, 40001, "OoooH.... request lost")
	ErrInsufficientApproval    = gfsperrors.Register(module.ExecuteModularName, http.StatusNotFound, 40002, "insufficient approvals from p2p")
	ErrUnsealed                = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 40003, "seal object on chain failed")
	ErrExhaustedApproval       = gfsperrors.Register(module.ExecuteModularName, http.StatusNotFound, 40004, "approvals exhausted")
	ErrInvalidIntegrity        = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40005, "secondary integrity hash verification failed")
	ErrSecondaryMismatch       = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40006, "secondary sp mismatch")
	ErrReplicateIdsOutOfBounds = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40007, "replicate idx out of bounds")
	ErrGfSpDB                  = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45201, "server slipped away, try again later")
	ErrRecoveryRedundancyType  = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45202, "recovery only support EC redundancy type")
	ErrRecoveryPieceNotEnough  = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45203, "fail to get enough piece data to recovery")
	ErrRecoveryDecode          = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45204, "EC decode error")
	ErrPieceStore              = gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 45205, "server slipped away, try again later")
	ErrRecoveryPieceChecksum   = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45206, "recovery checksum not correct")
	ErrRecoveryPieceLength     = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45207, "get secondary piece data length error")
	ErrPrimaryNotFound         = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 45208, "primary sp endpoint not found when recoverying")
	ErrRecoveryPieceIndex      = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 45209, "recovery piece index invalid")
)

func (e *ExecuteModular) HandleSealObjectTask(ctx context.Context, task coretask.SealObjectTask) {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to handle seal object, task pointer dangling")
		task.SetError(ErrDanglingPointer)
		return
	}
	task.AppendLog("executor-begin-handle-seal-task")
	sealMsg := &storagetypes.MsgSealObject{
		Operator:              e.baseApp.OperatorAddress(),
		BucketName:            task.GetObjectInfo().GetBucketName(),
		ObjectName:            task.GetObjectInfo().GetObjectName(),
		SecondarySpAddresses:  task.GetSecondaryAddresses(),
		SecondarySpSignatures: task.GetSecondarySignatures(),
	}
	task.SetError(e.sealObject(ctx, task, sealMsg))
	task.AppendLog("executor-end-handle-seal-task")
	log.CtxDebugw(ctx, "finish to handle seal object task", "error", task.Error())
}

func (e *ExecuteModular) sealObject(ctx context.Context, task coretask.ObjectTask, sealMsg *storagetypes.MsgSealObject) error {
	var (
		err    error
		txHash string
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ExecutorCounter.WithLabelValues(ExeutorFailureSealObject).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorFailureSealObject).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessSealObject).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorSuccessSealObject).Observe(time.Since(startTime).Seconds())
		}
	}()
	for retry := int64(0); retry <= task.GetMaxRetry(); retry++ {
		txHash, err = e.baseApp.GfSpClient().SealObject(ctx, sealMsg)
		if err != nil {
			task.AppendLog(fmt.Sprintf("executor-seal-tx-failed-error:%s-retry:%d", err.Error(), retry))
			log.CtxErrorw(ctx, "failed to seal object", "retry", retry,
				"max_retry", task.GetMaxRetry(), "error", err)
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
		sealed, err = e.baseApp.Consensus().ListenObjectSeal(ctx,
			object.Id.Uint64(), e.listenSealTimeoutHeight)
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
		log.CtxErrorw(ctx, "failed to handle receive piece confirm, task pointer dangling")
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
		log.CtxErrorw(ctx, "failed to confirm receive task, object is unsealed")
		task.SetError(ErrUnsealed)
		return
	}
	// regardless of whether the sp is as secondary or not, it needs to be set to the
	// sealed state to let the manager clear the task.
	task.SetSealed(true)
	if int(task.GetReplicateIdx()) >= len(offChainObject.GetSecondarySpAddresses()) {
		log.CtxErrorw(ctx, "failed to confirm receive task, replicate idx out of bounds",
			"replicate_idx", task.GetReplicateIdx(),
			"secondary_sp_len", len(offChainObject.GetSecondarySpAddresses()))
		task.SetError(ErrReplicateIdsOutOfBounds)
		return
	}
	if offChainObject.GetSecondarySpAddresses()[int(task.GetReplicateIdx())] != e.baseApp.OperatorAddress() {
		log.CtxErrorw(ctx, "failed to confirm receive task, secondary sp mismatch",
			"expect", offChainObject.GetSecondarySpAddresses()[int(task.GetReplicateIdx())],
			"current", e.baseApp.OperatorAddress())
		task.SetError(ErrSecondaryMismatch)
		// TODO:: gc zombie task will gc the zombie piece, it is a conservative plan
		err = e.baseApp.GfSpDB().DeleteObjectIntegrity(task.GetObjectInfo().Id.Uint64())
		if err != nil {
			log.CtxErrorw(ctx, "failed to delete integrity")
		}
		var pieceKey string
		segmentCount := e.baseApp.PieceOp().SegmentPieceCount(offChainObject.GetPayloadSize(),
			task.GetStorageParams().GetMaxPayloadSize())
		for i := uint32(0); i < segmentCount; i++ {
			if task.GetObjectInfo().GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
				pieceKey = e.baseApp.PieceOp().ECPieceKey(offChainObject.Id.Uint64(),
					i, task.GetReplicateIdx())
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
	log.CtxDebugw(ctx, "succeed to handle confirm receive piece task")
}

func (e *ExecuteModular) HandleGCObjectTask(ctx context.Context, task coretask.GCObjectTask) {
	var (
		err                error
		waitingGCObjects   []*types.Object
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
		log.CtxDebugw(ctx, "gc object task",
			"task_info", task.Info(), "is_succeed", isSucceed,
			"response_end_block_id", responseEndBlockID, "waiting_gc_object_number", len(waitingGCObjects),
			"has_gc_object_number", gcObjectNumber, "try_again_later", tryAgainLater,
			"task_is_canceled", taskIsCanceled, "has_no_object", hasNoObject, "error", err)
	}()

	if waitingGCObjects, responseEndBlockID, err = e.baseApp.GfSpClient().ListDeletedObjectsByBlockNumberRange(
		ctx, e.baseApp.OperatorAddress(), task.GetStartBlockNumber(),
		task.GetEndBlockNumber(), true); err != nil {
		log.CtxErrorw(ctx, "failed to query deleted object list", "task_info", task.Info(), "error", err)
		return
	}
	if responseEndBlockID < task.GetStartBlockNumber() || responseEndBlockID < task.GetEndBlockNumber() {
		tryAgainLater = true
		log.CtxInfow(ctx, "metadata is not latest, try again later",
			"response_end_block_id", responseEndBlockID, "task_info", task.Info())
		return
	}
	if len(waitingGCObjects) == 0 {
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
		segmentCount := e.baseApp.PieceOp().SegmentPieceCount(
			objectInfo.GetPayloadSize(), storageParams.VersionedParams.GetMaxSegmentSize())
		for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
			pieceKey := e.baseApp.PieceOp().SegmentPieceKey(currentGCObjectID, segIdx)
			// ignore this delete api error, TODO: refine gc workflow by enrich metadata index.
			deleteErr := e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
			log.CtxDebugw(ctx, "delete the primary sp pieces",
				"object_info", objectInfo, "piece_key", pieceKey, "error", deleteErr)
		}
		for rIdx, address := range objectInfo.GetSecondarySpAddresses() {
			if strings.Compare(e.baseApp.OperatorAddress(), address) == 0 {
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
		// ignore this delete api error, TODO: refine gc workflow by enrich metadata index.
		deleteErr := e.baseApp.GfSpDB().DeleteObjectIntegrity(objectInfo.Id.Uint64())
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
	log.CtxWarn(ctx, "gc zombie piece future support")
}

func (e *ExecuteModular) HandleGCMetaTask(ctx context.Context, task coretask.GCMetaTask) {
	log.CtxWarn(ctx, "gc meta future support")
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
	primarySPEndpoint, err = e.getObjectPrimarySPEndpoint(ctx, task.GetObjectInfo().BucketName)
	if err != nil {
		return err
	}

	pieceData, err = e.doRecoveryPiece(ctx, task, primarySPEndpoint)
	if err != nil {
		log.CtxDebugw(ctx, "fail to recovery secondary SP data from secondary SPs")
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
func (e *ExecuteModular) getECPieceBySegment(ctx context.Context, redundancyIdx int32, objectInfo *storagetypes.ObjectInfo, params *storagetypes.Params, recoverySegData []byte, segmentIdx uint32) ([]byte, error) {
	dataShards := params.GetRedundantDataChunkNum()
	parityShards := params.GetRedundantParityChunkNum()
	if redundancyIdx < 0 || redundancyIdx > int32(dataShards+parityShards-1) {
		return nil, fmt.Errorf("invaild redundancyIdx ")
	}
	// if it is the data shards of ec-encoded pieces, just get the ec data by offset
	if redundancyIdx > 0 && redundancyIdx < int32(dataShards)-1 {
		ECPieceSize := e.baseApp.PieceOp().ECPieceSize(objectInfo.PayloadSize, segmentIdx, params.GetMaxSegmentSize(), params.GetRedundantDataChunkNum())

		startPos := int64(redundancyIdx) * ECPieceSize
		endPos := int64(redundancyIdx+1)*ECPieceSize - 1
		return recoverySegData[startPos:endPos], nil
	}

	// if it is the parity shard, it needs to encode again to compute the parity shards
	ECEncodeData, err := redundancy.EncodeRawSegment(recoverySegData,
		int(dataShards),
		int(parityShards))
	if err != nil {
		log.CtxErrorw(ctx, "failed to ec encode data when recovering secondary SP", "error", err)
		return nil, err
	}
	return ECEncodeData[redundancyIdx], nil
}

func (e *ExecuteModular) checkRecoveryChecksum(ctx context.Context, task coretask.RecoveryPieceTask, recoveryChecksum []byte) error {
	integrityMeta, err := e.baseApp.GfSpDB().GetObjectIntegrity(task.GetObjectInfo().Id.Uint64())
	if err != nil {
		log.CtxErrorw(ctx, "search integrity hash in db error when recovery", "objectName:", task.GetObjectInfo().ObjectName, "error", err)
		return ErrGfSpDB
	}

	expectedHash := integrityMeta.PieceChecksumList[task.GetSegmentIdx()]
	if !bytes.Equal(recoveryChecksum, expectedHash) {
		log.CtxErrorw(ctx, "check integrity hash of recovery data err", "objectName:", task.GetObjectInfo().ObjectName,
			"expected value", hex.EncodeToString(expectedHash), "actual value", recoveryChecksum, "error", ErrRecoveryPieceChecksum)
		return ErrRecoveryPieceChecksum
	}
	return nil
}

func (e *ExecuteModular) doRecoveryPiece(ctx context.Context, rTask coretask.RecoveryPieceTask, endpoint string) (data []byte, err error) {
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
	// recovery primary sp segment pr secondary piece
	respBody, err := e.baseApp.GfSpClient().GetPieceFromECChunks(ctx, endpoint, rTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to recovery piece", "objectID", rTask.GetObjectInfo().Id,
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

	log.CtxDebugw(ctx, "success to recovery piece from sp", "objectID", rTask.GetObjectInfo().Id,
		"segment_idx", rTask.GetSegmentIdx(), "secondary endpoint", endpoint)

	return pieceData, nil
}

// getObjectSecondaryEndpoints return the secondary sp endpoints list of the specific object
func (e *ExecuteModular) getObjectSecondaryEndpoints(ctx context.Context, objectInfo *storagetypes.ObjectInfo) ([]string, int, error) {
	secondarySPAddrs := objectInfo.GetSecondarySpAddresses()
	spList, err := e.baseApp.Consensus().ListSPs(ctx)
	if err != nil {
		return nil, 0, err
	}

	var secondaryCount int
	secondaryEndpointList := make([]string, len(secondarySPAddrs))
	for idx, addr := range secondarySPAddrs {
		for _, info := range spList {
			if addr == info.GetOperatorAddress() {
				secondaryCount++
				secondaryEndpointList[idx] = info.Endpoint
			}
		}
	}

	return secondaryEndpointList, secondaryCount, nil
}

func (e *ExecuteModular) getObjectPrimarySPEndpoint(ctx context.Context, bucketName string) (string, error) {
	spList, err := e.baseApp.Consensus().ListSPs(ctx)
	if err != nil {
		return "", err
	}

	bucketInfo, err := e.baseApp.Consensus().QueryBucketInfo(ctx, bucketName)
	if err != nil {
		return "", err
	}

	primarySP := bucketInfo.GetPrimarySpAddress()
	for _, info := range spList {
		if primarySP == info.GetOperatorAddress() {
			return info.Endpoint, nil
		}
	}

	return "", ErrPrimaryNotFound
}
