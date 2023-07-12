package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prysmaticlabs/prysm/crypto/bls"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (e *ExecuteModular) HandleReplicatePieceTask(ctx context.Context, task coretask.ReplicatePieceTask) {
	var (
		err    error
		blsSig []bls.Signature
	)
	startReplicateTime := time.Now()
	defer func() {
		task.SetError(err)
		metrics.PerfPutObjectTime.WithLabelValues("background_replicate_cost").Observe(time.Since(startReplicateTime).Seconds())
	}()
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		err = ErrDanglingPointer
		return
	}

	replicatePieceTotalTime := time.Now()
	err = e.handleReplicatePiece(ctx, task)
	metrics.PerfPutObjectTime.WithLabelValues("background_replicate_object_time").Observe(time.Since(replicatePieceTotalTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_task_replicate_object_end_time").Observe(time.Since(startReplicateTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "task_info", task.Info(), "error", err)
		return
	}
	log.CtxDebugw(ctx, "succeed to replicate all pieces", "task_info", task.Info())

	if blsSig, err = bls.MultipleSignaturesFromBytes(task.GetSecondarySignatures()); err != nil {
		log.CtxErrorw(ctx, "failed to generate multiple signatures",
			"origin_signature", task.GetSecondarySignatures(), "error", err)
		return
	} else {
		task.AppendLog("executor-end-replicate-object")
		metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessReplicateAllPiece).Inc()
		metrics.ExecutorTime.WithLabelValues(ExeutorSuccessReplicateAllPiece).Observe(time.Since(replicatePieceTotalTime).Seconds())
	}
	sealMsg := &storagetypes.MsgSealObject{
		Operator:                    e.baseApp.OperatorAddress(),
		BucketName:                  task.GetObjectInfo().GetBucketName(),
		ObjectName:                  task.GetObjectInfo().GetObjectName(),
		GlobalVirtualGroupId:        task.GetGlobalVirtualGroupId(),
		SecondarySpBlsAggSignatures: bls.AggregateSignatures(blsSig).Marshal(),
	}
	sealTime := time.Now()
	sealErr := e.sealObject(ctx, task, sealMsg)
	metrics.PerfPutObjectTime.WithLabelValues("background_seal_object_cost").Observe(time.Since(sealTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_task_seal_object_end").Observe(time.Since(startReplicateTime).Seconds())
	if sealErr == nil {
		task.SetSealed(true)
	}
	log.CtxDebugw(ctx, "finish combine seal object", "task", task, "error", sealErr)
}

func (e *ExecuteModular) handleReplicatePiece(ctx context.Context, rTask coretask.ReplicatePieceTask) (err error) {
	var (
		wg                sync.WaitGroup
		segmentPieceCount = e.baseApp.PieceOp().SegmentPieceCount(
			rTask.GetObjectInfo().GetPayloadSize(),
			rTask.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		replicateCount = rTask.GetStorageParams().VersionedParams.GetRedundantDataChunkNum() +
			rTask.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()
		secondarySignatures = make([][]byte, replicateCount)
	)

	log.Debugw("replicate task info", "task_sps", rTask.GetSecondaryEndpoints())

	doReplicateECPiece := func(pieceIdx uint32, data [][]byte) {
		log.Debugw("start to replicate ec piece")
		for rIdx, sp := range rTask.GetSecondaryEndpoints() {
			log.Debugw("start to replicate ec piece", "sp", sp)
			wg.Add(1)
			go e.doReplicatePiece(ctx, &wg, rTask, sp, uint32(rIdx), pieceIdx, data[rIdx])
		}
		wg.Wait()
		log.Debugw("finish to replicate ec piece")
	}
	doReplicateSegmentPiece := func(pieceIdx uint32, data []byte) {
		log.Debugw("start to replicate segment piece")
		for rIdx, sp := range rTask.GetSecondaryEndpoints() {
			log.Debugw("start to replicate segment piece", "sp", sp)
			wg.Add(1)
			go e.doReplicatePiece(ctx, &wg, rTask, sp, uint32(rIdx), pieceIdx, data)
		}
		wg.Wait()
		log.Debugw("finish to replicate segment piece")
	}
	doneReplicate := func() error {
		log.Debugw("start to done replicate")
		for rIdx, sp := range rTask.GetSecondaryEndpoints() {
			log.Debugw("start to done replicate", "sp", sp)
			signature, innerErr := e.doneReplicatePiece(ctx, rTask, sp, uint32(rIdx))
			if innerErr == nil {
				secondarySignatures[rIdx] = signature
				metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessDoneReplicatePiece).Inc()
			} else {
				metrics.ExecutorCounter.WithLabelValues(ExeutorFailureDoneReplicatePiece).Inc()
				return innerErr
			}
		}
		log.Debugw("finish to done replicate")
		return nil
	}

	startReplicatePieceTime := time.Now()
	for segIdx := uint32(0); segIdx < segmentPieceCount; segIdx++ {
		pieceKey := e.baseApp.PieceOp().SegmentPieceKey(rTask.GetObjectInfo().Id.Uint64(), segIdx)
		startGetPieceTime := time.Now()
		segData, err := e.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
		metrics.PerfPutObjectTime.WithLabelValues("background_get_piece_time").Observe(time.Since(startGetPieceTime).Seconds())
		metrics.PerfPutObjectTime.WithLabelValues("background_get_piece_end_time").Observe(time.Since(time.Unix(rTask.GetCreateTime(), 0)).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to get segment data from piece store", "error", err)
			rTask.SetError(err)
			return err
		}
		if rTask.GetObjectInfo().GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
			ecTime := time.Now()
			ecData, err := redundancy.EncodeRawSegment(segData,
				int(rTask.GetStorageParams().VersionedParams.GetRedundantDataChunkNum()),
				int(rTask.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()))
			metrics.PerfPutObjectTime.WithLabelValues("background_ec_time").Observe(time.Since(ecTime).Seconds())
			metrics.PerfPutObjectTime.WithLabelValues("background_ec_end_time").Observe(time.Since(time.Unix(rTask.GetCreateTime(), 0)).Seconds())
			if err != nil {
				log.CtxErrorw(ctx, "failed to ec encode data", "error", err)
				rTask.SetError(err)
				return err
			}
			doReplicateECPiece(segIdx, ecData)
		} else {
			doReplicateSegmentPiece(segIdx, segData)
		}
	}
	metrics.PerfPutObjectTime.WithLabelValues("background_replicate_all_piece_time").Observe(time.Since(startReplicatePieceTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_replicate_all_piece_end_time").Observe(time.Since(startReplicatePieceTime).Seconds())
	doneTime := time.Now()
	err = doneReplicate()
	metrics.PerfPutObjectTime.WithLabelValues("background_done_replicate_time").Observe(time.Since(doneTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_done_replicate_piece_end_time").Observe(time.Since(startReplicatePieceTime).Seconds())
	if err == nil {
		rTask.SetSecondarySignatures(secondarySignatures)
	}
	return err
}

func (e *ExecuteModular) doReplicatePiece(ctx context.Context, waitGroup *sync.WaitGroup, rTask coretask.ReplicatePieceTask,
	spEndpoint string, replicateIdx uint32, pieceIdx uint32, data []byte) (err error) {
	var signature []byte
	rTask.AppendLog(fmt.Sprintf("executor-begin-replicate-piece-sidx:%d-ridx-%d", pieceIdx, replicateIdx))
	startTime := time.Now()
	defer func() {
		if err != nil {
			rTask.AppendLog(fmt.Sprintf("executor-end-replicate-piece-sidx:%d-ridx-%d-error:%s-endpoint:%s", pieceIdx, replicateIdx, err.Error(), spEndpoint))
			metrics.ExecutorCounter.WithLabelValues(ExeutorFailureReplicateOnePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorFailureReplicateOnePiece).Observe(time.Since(startTime).Seconds())
		} else {
			rTask.AppendLog(fmt.Sprintf("executor-end-replicate-piece-sidx:%d-ridx-%d-endpoint:%s", pieceIdx, replicateIdx, spEndpoint))
			metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessReplicateOnePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorSuccessReplicateOnePiece).Observe(time.Since(startTime).Seconds())
		}
		waitGroup.Done()
	}()
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetGlobalVirtualGroupId(), rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), replicateIdx, int32(pieceIdx), int64(len(data)))
	receive.SetPieceChecksum(hash.GenerateChecksum(data))
	ctx = log.WithValue(ctx, log.CtxKeyTask, receive.Key().String())
	signTime := time.Now()
	signature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_cost").Observe(time.Since(signTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_end").Observe(time.Since(startTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign receive task", "replicate_idx", replicateIdx,
			"piece_idx", pieceIdx, "error", err)
		return
	}
	receive.SetSignature(signature)
	replicateOnePieceTime := time.Now()
	err = e.baseApp.GfSpClient().ReplicatePieceToSecondary(ctx, spEndpoint, receive, data)
	metrics.PerfPutObjectTime.WithLabelValues("background_replicate_one_piece_cost").Observe(time.Since(replicateOnePieceTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_replicate_one_piece_end").Observe(time.Since(startTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "replicate_idx", replicateIdx,
			"piece_idx", pieceIdx, "error", err)
		return
	}
	log.CtxDebugw(ctx, "success to replicate piece", "replicate_idx", replicateIdx,
		"piece_idx", pieceIdx)
	return
}

func (e *ExecuteModular) doneReplicatePiece(ctx context.Context, rTask coretask.ReplicatePieceTask,
	spEndpoint string, replicateIdx uint32) ([]byte, error) {
	var (
		err           error
		signature     []byte
		taskSignature []byte
	)
	rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-ridx-%d", replicateIdx))
	startTime := time.Now()
	defer func() {
		if err != nil {
			rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-ridx-%d-error:%s-endpoint:%s", replicateIdx, err.Error(), spEndpoint))
			metrics.ExecutorCounter.WithLabelValues(ExeutorFailureDoneReplicatePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorFailureDoneReplicatePiece).Observe(time.Since(startTime).Seconds())
		} else {
			rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-ridx-%d-endpoint:%s", replicateIdx, spEndpoint))
			metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessDoneReplicatePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorSuccessDoneReplicatePiece).Observe(time.Since(startTime).Seconds())
		}
	}()

	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetGlobalVirtualGroupId(), rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), replicateIdx, -1, 0)
	signTime := time.Now()
	taskSignature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_cost").Observe(time.Since(signTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_end").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign done receive task",
			"replicate_idx", replicateIdx, "error", err)
		return nil, err
	}
	receive.SetSignature(taskSignature)
	doneReplicateTime := time.Now()
	signature, err = e.baseApp.GfSpClient().DoneReplicatePieceToSecondary(ctx, spEndpoint, receive)
	metrics.PerfPutObjectTime.WithLabelValues("background_done_receive_http_cost").Observe(time.Since(doneReplicateTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_done_receive_http_end").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece", "endpoint", spEndpoint,
			"replicate_idx", replicateIdx, "error", err)
		return nil, err
	}
	if int(replicateIdx+1) >= len(rTask.GetObjectInfo().GetChecksums()) {
		log.CtxErrorw(ctx, "failed to done replicate piece, replicate idx out of bounds",
			"replicate_idx", replicateIdx)
		return nil, ErrReplicateIdsOutOfBounds
	}

	// TODO:
	// veritySignatureTime := time.Now()
	// TODO get gvgId and blsPubKey from task, bls pub key alreay injected via key manager for current sp
	// var blsPubKey bls.PublicKey
	// err = veritySignature(ctx, rTask.GetObjectInfo().Id.Uint64(), rTask.GetGlobalVirtualGroupId(), integrity,
	//	storagetypes.GenerateHash(rTask.GetObjectInfo().GetChecksums()[:]), signature, blsPubKey)
	// metrics.PerfUploadTimeHistogram.WithLabelValues("background_verity_seal_signature_time").Observe(time.Since(veritySignatureTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_verity_seal_signature_end_time").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed verify secondary signature", "endpoint", spEndpoint,
			"replicate_idx", replicateIdx, "error", err)
		return nil, err
	}
	log.CtxDebugw(ctx, "succeed to done replicate", "endpoint", spEndpoint, "replicate_idx", replicateIdx)
	return signature, nil
}

/*
func veritySignature(ctx context.Context, objectID uint64, gvgId uint32, integrity []byte, expectedIntegrity []byte, signature []byte, blsPubKey bls.PublicKey) error {
	if !bytes.Equal(expectedIntegrity, integrity) {
		log.CtxErrorw(ctx, "replicate sp invalid integrity", "integrity", hex.EncodeToString(integrity),
			"expect", hex.EncodeToString(expectedIntegrity))
		return ErrInvalidIntegrity
	}
	originMsgHash := storagetypes.NewSecondarySpSealObjectSignDoc(sdkmath.NewUint(objectID), gvgId, integrity).GetBlsSignHash()
	err := types.VerifyBlsSignature(blsPubKey, originMsgHash, signature)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify signature", "error", err)
		return err
	}
	return nil
}
*/
