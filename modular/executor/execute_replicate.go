package executor

import (
	"context"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/prysmaticlabs/prysm/crypto/bls"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

//ExecutorBeginSealTx                 = "executor_begin_seal_tx"
//ExecutorEndSealTx                   = "executor_end_seal_tx"
//ExecutorBeginConfirmSeal            = "executor_begin_confirm_seal"
//ExecutorEndConfirmSeal              = "executor_end_confirm_seal"

func (e *ExecuteModular) HandleReplicatePieceTask(ctx context.Context, task coretask.ReplicatePieceTask) {
	var (
		err error
	)
	startReplicateTime := time.Now()
	defer func() {
		task.SetError(err)
		metrics.PerfUploadTimeHistogram.WithLabelValues("background_replicate_time").Observe(time.Since(startReplicateTime).Seconds())
	}()
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		err = ErrDanglingPointer
		return
	}

	replicatePieceTotalTime := time.Now()
	err = e.handleReplicatePiece(ctx, task)
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_replicate_object_time").Observe(time.Since(replicatePieceTotalTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_task_replicate_object_end_time").Observe(time.Since(startReplicateTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "task_info", task.Info(), "error", err)
		return
	}
	log.CtxDebugw(ctx, "succeed to replicate all pieces", "task_info", task.Info())

	blsSig, err := bls.MultipleSignaturesFromBytes(task.GetSecondarySignatures())
	if err != nil {
		log.CtxErrorw(ctx, "failed to generate multiple signatures", "error", err)
		return
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
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_seal_object_time").Observe(time.Since(sealTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_task_seal_object_end_time").Observe(time.Since(startReplicateTime).Seconds())
	if sealErr == nil {
		task.SetSealed(true)
	}
	log.CtxDebugw(ctx, "finish combine seal object", "error", sealErr)
}

func (e *ExecuteModular) handleReplicatePiece(ctx context.Context, rTask coretask.ReplicatePieceTask) (err error) {
	var (
		wg                sync.WaitGroup
		segmentPieceCount = e.baseApp.PieceOp().SegmentPieceCount(
			rTask.GetObjectInfo().GetPayloadSize(),
			rTask.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		replicateCount = rTask.GetStorageParams().VersionedParams.GetRedundantDataChunkNum() +
			rTask.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()
		secondaryAddresses  = make([]string, replicateCount)
		secondarySignatures = make([][]byte, replicateCount)
	)

	doReplicateECPiece := func(pieceIdx uint32, data [][]byte) {
		for rIdx, sp := range rTask.GetSecondarySps() {
			wg.Add(1)
			go e.doReplicatePiece(ctx, &wg, rTask, sp, uint32(rIdx), pieceIdx, data[rIdx])
		}
		wg.Wait()
	}
	doReplicateSegmentPiece := func(pieceIdx uint32, data []byte) {
		for rIdx, sp := range rTask.GetSecondarySps() {
			wg.Add(1)
			go e.doReplicatePiece(ctx, &wg, rTask, sp, uint32(rIdx), pieceIdx, data)
		}
		wg.Wait()
	}
	doneReplicate := func() error {
		for rIdx, sp := range rTask.GetSecondarySps() {
			_, signature, innerErr := e.doneReplicatePiece(ctx, rTask, sp, uint32(rIdx))
			if innerErr == nil {
				secondaryAddresses[rIdx] = sp.GetOperatorAddress()
				secondarySignatures[rIdx] = signature
				metrics.ReplicateSucceedCounter.WithLabelValues(e.Name()).Inc()
			} else {
				metrics.ReplicateFailedCounter.WithLabelValues(e.Name()).Inc()
				return innerErr
			}
		}
		return nil
	}

	startReplicatePieceTime := time.Now()
	for pIdx := uint32(0); pIdx < segmentPieceCount; pIdx++ {
		pieceKey := e.baseApp.PieceOp().SegmentPieceKey(rTask.GetObjectInfo().Id.Uint64(), pIdx)
		startGetPieceTime := time.Now()
		segData, err := e.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
		metrics.PerfUploadTimeHistogram.WithLabelValues("background_get_piece_time").Observe(time.Since(startGetPieceTime).Seconds())
		metrics.PerfUploadTimeHistogram.WithLabelValues("background_get_piece_end_time").Observe(time.Since(time.Unix(rTask.GetCreateTime(), 0)).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to get segment data form piece store", "error", err)
			rTask.SetError(err)
			return err
		}
		if rTask.GetObjectInfo().GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
			ecTime := time.Now()
			ecData, err := redundancy.EncodeRawSegment(segData,
				int(rTask.GetStorageParams().VersionedParams.GetRedundantDataChunkNum()),
				int(rTask.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()))
			metrics.PerfUploadTimeHistogram.WithLabelValues("background_ec_time").Observe(time.Since(ecTime).Seconds())
			metrics.PerfUploadTimeHistogram.WithLabelValues("background_ec_end_time").Observe(time.Since(time.Unix(rTask.GetCreateTime(), 0)).Seconds())
			if err != nil {
				log.CtxErrorw(ctx, "failed to ec encode data", "error", err)
				rTask.SetError(err)
				return err
			}
			doReplicateECPiece(pIdx, ecData)
		} else {
			doReplicateSegmentPiece(pIdx, segData)
		}
	}
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_replicate_all_piece_time").Observe(time.Since(startReplicatePieceTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_replicate_all_piece_end_time").Observe(time.Since(startReplicatePieceTime).Seconds())
	doneTime := time.Now()
	err = doneReplicate()
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_done_replicate_time").Observe(time.Since(doneTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_done_replicate_piece_end_time").Observe(time.Since(startReplicatePieceTime).Seconds())
	return err
}

func (e *ExecuteModular) doReplicatePiece(ctx context.Context, waitGroup *sync.WaitGroup, rTask coretask.ReplicatePieceTask,
	sp *sptypes.StorageProvider, replicateIdx uint32, pieceIdx uint32, data []byte) (err error) {
	var signature []byte
	metrics.ReplicatePieceSizeCounter.WithLabelValues(e.Name()).Add(float64(len(data)))
	startTime := time.Now()
	defer func() {
		metrics.ReplicatePieceTimeHistogram.WithLabelValues(e.Name()).Observe(time.Since(startTime).Seconds())
		waitGroup.Done()
	}()
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetGlobalVirtualGroupId(), rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), replicateIdx, int32(pieceIdx), int64(len(data)))
	receive.SetPieceChecksum(hash.GenerateChecksum(data))
	ctx = log.WithValue(ctx, log.CtxKeyTask, receive.Key().String())
	signTime := time.Now()
	signature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_sign_receive_time").Observe(time.Since(signTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_sign_receive_end_time").Observe(time.Since(startTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign receive task", "replicate_idx", replicateIdx,
			"piece_idx", pieceIdx, "error", err)
		return
	}
	receive.SetSignature(signature)
	replicateOnePieceTime := time.Now()
	e.baseApp.GfSpDB().InsertUploadEvent(rTask.GetObjectInfo().Id.Uint64(), spdb.ExecutorBeginReplicateOnePiece, receive.Info())
	err = e.baseApp.GfSpClient().ReplicatePieceToSecondary(ctx, sp.GetEndpoint(), receive, data)
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_replicate_one_piece_time").Observe(time.Since(replicateOnePieceTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_replicate_one_piece_end_time").Observe(time.Since(startTime).Seconds())
	if err != nil {
		e.baseApp.GfSpDB().InsertUploadEvent(rTask.GetObjectInfo().Id.Uint64(), spdb.ExecutorEndReplicateOnePiece, receive.Info()+":"+err.Error())
		log.CtxErrorw(ctx, "failed to replicate piece", "replicate_idx", replicateIdx,
			"piece_idx", pieceIdx, "error", err)
		return
	}
	e.baseApp.GfSpDB().InsertUploadEvent(rTask.GetObjectInfo().Id.Uint64(), spdb.ExecutorEndReplicateOnePiece, receive.Info())
	log.CtxDebugw(ctx, "success to replicate piece", "replicate_idx", replicateIdx,
		"piece_idx", pieceIdx)
	return
}

func (e *ExecuteModular) doneReplicatePiece(ctx context.Context, rTask coretask.ReplicatePieceTask,
	sp *sptypes.StorageProvider, replicateIdx uint32) ([]byte, []byte, error) {
	var (
		err           error
		integrity     []byte
		signature     []byte
		taskSignature []byte
	)
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetGlobalVirtualGroupId(), rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), replicateIdx, -1, 0)
	signTime := time.Now()
	taskSignature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_sign_receive_time").Observe(time.Since(signTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_sign_receive_end_time").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign done receive task",
			"replicate_idx", replicateIdx, "error", err)
		return nil, nil, err
	}
	receive.SetSignature(taskSignature)
	doneReplicateTime := time.Now()
	e.baseApp.GfSpDB().InsertUploadEvent(rTask.GetObjectInfo().Id.Uint64(), spdb.ExecutorBeginDoneReplicatePiece, receive.Info())
	integrity, signature, err = e.baseApp.GfSpClient().DoneReplicatePieceToSecondary(ctx,
		sp.GetEndpoint(), receive)
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_done_receive_http_time").Observe(time.Since(doneReplicateTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_done_receive_http_end_time").Observe(time.Since(signTime).Seconds())
	if err != nil {
		e.baseApp.GfSpDB().InsertUploadEvent(rTask.GetObjectInfo().Id.Uint64(), spdb.ExecutorEndDoneReplicatePiece, receive.Info()+":"+err.Error())
		log.CtxErrorw(ctx, "failed to done replicate piece",
			"endpoint", sp.GetEndpoint(),
			"replicate_idx", replicateIdx, "error", err)
		return nil, nil, err
	}
	e.baseApp.GfSpDB().InsertUploadEvent(rTask.GetObjectInfo().Id.Uint64(), spdb.ExecutorEndDoneReplicatePiece, receive.Info())
	if int(replicateIdx+1) >= len(rTask.GetObjectInfo().GetChecksums()) {
		log.CtxErrorw(ctx, "failed to done replicate piece, replicate idx out of bounds",
			"replicate_idx", replicateIdx)
		return nil, nil, ErrReplicateIdsOutOfBounds
	}

	// TODO:
	// veritySignatureTime := time.Now()
	// TODO get gvgId and blsPubKey from task, bls pub key alreay injected via key manager for current sp
	// var blsPubKey bls.PublicKey
	// err = veritySignature(ctx, rTask.GetObjectInfo().Id.Uint64(), rTask.GetGlobalVirtualGroupId(), integrity,
	//	storagetypes.GenerateHash(rTask.GetObjectInfo().GetChecksums()[:]), signature, blsPubKey)
	// metrics.PerfUploadTimeHistogram.WithLabelValues("background_verity_seal_signature_time").Observe(time.Since(veritySignatureTime).Seconds())
	metrics.PerfUploadTimeHistogram.WithLabelValues("background_verity_seal_signature_end_time").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed verify secondary signature",
			"endpoint", sp.GetEndpoint(),
			"replicate_idx", replicateIdx, "error", err)
		return nil, nil, err
	}
	log.CtxDebugw(ctx, "succeed to done replicate",
		"endpoint", sp.GetEndpoint(),
		"replicate_idx", replicateIdx)
	return integrity, signature, nil
}

/*
func veritySignature(ctx context.Context, objectID uint64, gvgId uint32, integrity []byte, expectedIntegrity []byte, signature []byte, blsPubKey bls.PublicKey) error {
	if !bytes.Equal(expectedIntegrity, integrity) {
		log.CtxErrorw(ctx, "replicate sp invalid integrity", "integrity", hex.EncodeToString(integrity),
			"expect", hex.EncodeToString(expectedIntegrity))
		return ErrInvalidIntegrity
	}
	originMsgHash := storagetypes.NewSecondarySpSealObjectSignDoc(sdkmath.NewUint(objectID), gvgId, integrity).GetSignBytes()
	err := types.VerifyBlsSignature(blsPubKey, originMsgHash, signature)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify signature", "error", err)
		return err
	}
	return nil
}
*/
