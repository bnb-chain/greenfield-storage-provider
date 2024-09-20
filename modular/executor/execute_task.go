package executor

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/0xPolygon/polygon-edge/bls"
	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	storagetypes "github.com/evmos/evmos/v12/x/storage/types"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsperrors"
	"github.com/zkMeLabs/mechain-storage-provider/core/module"
	"github.com/zkMeLabs/mechain-storage-provider/core/spdb"
	coretask "github.com/zkMeLabs/mechain-storage-provider/core/task"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/metrics"
	"github.com/zkMeLabs/mechain-storage-provider/util"
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
	ErrInvalidReplicatePieceTask    = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45217, "invalid replicate piece task")
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

func (e *ExecuteModular) HandleSealObjectTask(ctx context.Context, task coretask.SealObjectTask) {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to handle seal object, task pointer dangling")
		task.SetError(ErrDanglingPointer)
		return
	}
	var blsSig bls.Signatures
	blsSigBts := task.GetSecondarySignatures()
	for _, sigBts := range blsSigBts {
		signature, err := bls.UnmarshalSignature(sigBts)
		if err != nil {
			return
		}
		blsSig = append(blsSig, signature)
	}
	blsAggSigs, _ := blsSig.Aggregate().Marshal()
	task.AppendLog("executor-begin-handle-seal-task")
	if task.GetIsAgentUpload() {
		checksums, makeErr := e.makeCheckSumsForAgentUpload(ctx, task.GetObjectInfo(), len(task.GetSecondaryAddresses()), task.GetStorageParams())
		if makeErr != nil {
			task.SetError(makeErr)
			return
		}
		sealMsg := &storagetypes.MsgSealObjectV2{
			Operator:                    e.baseApp.OperatorAddress(),
			BucketName:                  task.GetObjectInfo().GetBucketName(),
			ObjectName:                  task.GetObjectInfo().GetObjectName(),
			GlobalVirtualGroupId:        task.GetGlobalVirtualGroupId(),
			SecondarySpBlsAggSignatures: blsAggSigs,
			ExpectChecksums:             checksums,
		}
		task.SetError(e.sealObjectV2(ctx, task, sealMsg))
	} else {
		sealMsg := &storagetypes.MsgSealObject{
			Operator:                    e.baseApp.OperatorAddress(),
			BucketName:                  task.GetObjectInfo().GetBucketName(),
			ObjectName:                  task.GetObjectInfo().GetObjectName(),
			GlobalVirtualGroupId:        task.GetGlobalVirtualGroupId(),
			SecondarySpBlsAggSignatures: blsAggSigs,
		}
		task.SetError(e.sealObject(ctx, task, sealMsg))
	}

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

// sealObjectV2 Sends a new seal message to the chain
func (e *ExecuteModular) sealObjectV2(ctx context.Context, task coretask.ObjectTask, sealMsg *storagetypes.MsgSealObjectV2) error {
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
		txHash, err = e.baseApp.GfSpClient().SealObjectV2(ctx, sealMsg)
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
	if offChainObject.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED || offChainObject.IsUpdating {
		log.CtxError(ctx, "failed to confirm receive task, object is unsealed or still under updating")
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
				pieceKey = e.baseApp.PieceOp().ECPieceKey(offChainObject.Id.Uint64(), i, uint32(task.GetRedundancyIdx()), offChainObject.Version)
			} else {
				pieceKey = e.baseApp.PieceOp().SegmentPieceKey(offChainObject.Id.Uint64(), i, offChainObject.Version)
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

	// used by secondary SP or successor secondary SP for GVG
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
		// used by Primary SP or successor primary SP for VGF
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
	signature, err := e.baseApp.GfSpClient().SignRecoveryTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign recovery task", "object", task.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}
	task.SetSignature(signature)

	pieceData, err = e.doRecoveryPiece(ctx, task, primarySPEndpoint)
	if err != nil {
		log.CtxDebugw(ctx, "failed to recover secondary SP data from primary SP")
		return err
	}
	// compare integrity hash
	if !task.BySuccessorSP() {
		if err = e.checkRecoveryChecksum(ctx, task, hash.GenerateChecksum(pieceData)); err != nil {
			return err
		}
	}

	recoveryKey := e.baseApp.PieceOp().ECPieceKey(objectId, segmentIdx, uint32(task.GetEcIdx()), task.GetObjectInfo().GetVersion())
	// write the recovery segment key to keystore
	err = e.baseApp.PieceStore().PutPiece(ctx, recoveryKey, pieceData)
	if err != nil {
		log.CtxErrorw(ctx, "EC recover data write piece fail", "pieceKey:", recoveryKey, "error", err)
		return err
	}
	if task.BySuccessorSP() {
		err = e.setPieceMetadata(ctx, task, pieceData)
		if err != nil {
			log.CtxErrorw(ctx, "failed to set piece meta data to DB", "object_name:", task.GetObjectInfo().GetObjectName(), "segment_idx", task.GetSegmentIdx(), "redundancy_idx", task.GetEcIdx(), "error", err)
			return err
		}
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

	recoveryDataSources := make([][]byte, secondaryCount)
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

	signature, err := e.baseApp.GfSpClient().SignRecoveryTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign recovery task", "object", task.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}
	task.SetSignature(signature)

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	for ecIdx := 0; ecIdx < secondaryCount; ecIdx++ {
		recoveryDataSources[ecIdx] = nil
		go func(ctx context.Context, secondaryIndex int) {
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
		}(childCtx, ecIdx)
	}

loop:
	for {
		select {
		case <-doneCh:
			doneTaskNum++
			// it is enough to recovery data with minRecoveryPieces EC data, no need to wait
			if doneTaskNum >= minRecoveryPieces {
				cancel()
				break loop
			}
		case <-quitCh: // all the task finish
			cancel()
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
		if !task.BySuccessorSP() {
			if err = e.checkRecoveryChecksum(ctx, task, hash.GenerateChecksum(recoverySegData)); err != nil {
				return err
			}
		}
		// if the task is generated by primary SP, the recovery key is segment
		recoveryKey = e.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), task.GetSegmentIdx(), task.GetObjectInfo().GetVersion())
		recoveredPieceData = recoverySegData
	} else {
		redundancyIdx := task.GetEcIdx()
		// if the task is generated by a secondary SP, the recovery key is EC piece
		recoveryKey = e.baseApp.PieceOp().ECPieceKey(task.GetObjectInfo().Id.Uint64(), task.GetSegmentIdx(), uint32(redundancyIdx), task.GetObjectInfo().GetVersion())

		recoveredPieceData, err = e.getECPieceBySegment(ctx, task.GetEcIdx(), task.GetObjectInfo(), task.GetStorageParams(), recoverySegData, task.GetSegmentIdx())
		if err != nil {
			return err
		}
		// compare integrity hash
		if !task.BySuccessorSP() {
			if err = e.checkRecoveryChecksum(ctx, task, hash.GenerateChecksum(recoveredPieceData)); err != nil {
				return err
			}
		}
	}

	// write the recovery segment key to keystore
	if err = e.baseApp.PieceStore().PutPiece(ctx, recoveryKey, recoveredPieceData); err != nil {
		log.CtxErrorw(ctx, "EC decode data write piece fail", "pieceKey:", recoveryKey, "error", err)
		return err
	}

	if task.BySuccessorSP() {
		err = e.setPieceMetadata(ctx, task, recoveredPieceData)
		if err != nil {
			log.CtxErrorw(ctx, "failed to set piece meta data to DB", "object_name:", task.GetObjectInfo().GetObjectName(), "segment_idx", task.GetSegmentIdx(), "error", err)
			return err
		}
	}
	log.CtxDebugw(ctx, "finish recovery from secondary SPs", "object_name:", task.GetObjectInfo().GetObjectName(), "segment_idx", task.GetSegmentIdx())
	return nil
}

// getECPieceBySegment return the EC encodes data based on the redundancyIdx and the segment data
func (e *ExecuteModular) getECPieceBySegment(ctx context.Context, redundancyIdx int32, objectInfo *storagetypes.ObjectInfo,
	params *storagetypes.Params, recoverySegData []byte, segmentIdx uint32,
) ([]byte, error) {
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
	data []byte, err error,
) {
	var pieceData []byte
	// timeout for single piece recover
	ctxWithTimeout, cancel := context.WithTimeout(ctx, replicateTimeOut)
	defer cancel()
	// recovery primary sp segment or secondary piece
	respBody, err := e.baseApp.GfSpClient().GetPieceFromECChunks(ctxWithTimeout, endpoint, rTask)
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

func (e *ExecuteModular) setPieceMetadata(ctx context.Context, task coretask.RecoveryPieceTask, pieceData []byte) error {
	objectID := task.GetObjectInfo().Id.Uint64()
	segmentIdx := task.GetSegmentIdx()
	redundancyIdx := task.GetEcIdx()
	version := task.GetObjectInfo().Version
	pieceChecksum := hash.GenerateChecksum(pieceData)

	if err := e.baseApp.GfSpDB().SetReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx, pieceChecksum, version); err != nil {
		log.CtxErrorw(ctx, "failed to set replicate piece checksum", "object_id", objectID,
			"segment_index", segmentIdx, "redundancy_index", redundancyIdx, "error", err)
		detail := fmt.Sprintf("failed to set replicate piece checksum, object_id: %s, segment_index: %v, redundancy_index: %v, error: %s",
			task.GetObjectInfo().Id.String(), segmentIdx, task.GetEcIdx(), err.Error())
		return ErrGfSpDBWithDetail(detail)
	}
	segmentCount := e.baseApp.PieceOp().SegmentPieceCount(task.GetObjectInfo().GetPayloadSize(),
		task.GetStorageParams().VersionedParams.GetMaxSegmentSize())

	pieceChecksums, err := e.baseApp.GfSpDB().GetAllReplicatePieceChecksumOptimized(task.GetObjectInfo().Id.Uint64(), task.GetEcIdx(), segmentCount)
	if err != nil {
		log.CtxInfow(ctx, "failed to get recover piece checksum", "object_id", objectID,
			"segment_index", segmentIdx, "error", err)
		return err
	}
	if len(pieceChecksums) == int(segmentCount) {
		integrityChecksum := hash.GenerateIntegrityHash(pieceChecksums)
		integrityMeta := &spdb.IntegrityMeta{
			ObjectID:          task.GetObjectInfo().Id.Uint64(),
			RedundancyIndex:   task.GetEcIdx(),
			IntegrityChecksum: integrityChecksum,
			PieceChecksumList: pieceChecksums,
		}
		if err = e.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta); err != nil {
			log.CtxErrorw(ctx, "failed to set object integrity", "object_id", objectID,
				"segment_index", segmentIdx, "error", err)
			return err
		}
		err = e.baseApp.GfSpDB().DeleteAllReplicatePieceChecksum(objectID, task.GetEcIdx(), segmentCount)
		if err != nil {
			log.CtxErrorw(ctx, "failed to delete all recover piece checksum", "task", task, "error", err)
			return err
		}
	}
	return nil
}
