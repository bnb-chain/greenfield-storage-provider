package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/prysmaticlabs/prysm/crypto/bls"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	RtyAttem = retry.Attempts(uint(5))
	RtyDelay = retry.Delay(time.Millisecond * 500)
	RtyErr   = retry.LastErrorOnly(true)
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
		metrics.ExecutorCounter.WithLabelValues(ExecutorSuccessReplicateAllPiece).Inc()
		metrics.ExecutorTime.WithLabelValues(ExecutorSuccessReplicateAllPiece).Observe(time.Since(replicatePieceTotalTime).Seconds())
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

	doReplicateECPiece := func(ctx context.Context, segIdx uint32, data [][]byte, errChan chan error) {
		log.Debug("start to replicate ec piece")
		for redundancyIdx, sp := range rTask.GetSecondaryEndpoints() {
			log.Debugw("start to replicate ec piece", "sp", sp)
			wg.Add(1)
			go func(redundancyIdx int, sp string) {
				err = e.doReplicatePiece(ctx, &wg, rTask, sp, segIdx, int32(redundancyIdx), data[redundancyIdx])
				if err != nil {
					rTask.SetNotAvailableSpIdx(int32(redundancyIdx))
					errChan <- err
				}
			}(redundancyIdx, sp)
		}
		wg.Wait()
		log.Debug("finish to replicate ec piece")
	}
	doReplicateSegmentPiece := func(ctx context.Context, segIdx uint32, data []byte, errChan chan error) {
		log.Debug("start to replicate segment piece")
		for redundancyIdx, sp := range rTask.GetSecondaryEndpoints() {
			log.Debugw("start to replicate segment piece", "sp", sp)
			wg.Add(1)
			go func(redundancyIdx int, sp string) {
				err = e.doReplicatePiece(ctx, &wg, rTask, sp, segIdx, int32(redundancyIdx), data)
				if err != nil {
					rTask.SetNotAvailableSpIdx(int32(redundancyIdx))
					errChan <- err
				}
			}(redundancyIdx, sp)
		}
		wg.Wait()
		log.Debug("finish to replicate segment piece")
	}
	doneReplicate := func(ctx context.Context) error {
		log.Debug("start to done replicate")
		var gvg *virtualgrouptypes.GlobalVirtualGroup
		gvg, err = e.baseApp.GfSpClient().GetGlobalVirtualGroupByGvgID(ctx, rTask.GetGlobalVirtualGroupId())
		if err != nil {
			return ErrConsensusWithDetail("QueryGVGInfo error: " + err.Error())
		}
		for rIdx, spEp := range rTask.GetSecondaryEndpoints() {
			log.Debugw("start to done replicate", "sp", spEp)
			signature, innerErr := e.doneReplicatePiece(ctx, rTask, spEp, int32(rIdx))
			if innerErr == nil {
				msg := storagetypes.NewSecondarySpSealObjectSignDoc(e.baseApp.ChainID(), gvg.Id, rTask.GetObjectInfo().Id, storagetypes.GenerateHash(rTask.GetObjectInfo().GetChecksums()[:])).GetBlsSignHash()
				err = veritySecondarySpBlsSignature(e.getSpByID(gvg.GetSecondarySpIds()[rIdx]), signature, msg[:])
				if err != nil {
					rTask.SetNotAvailableSpIdx(int32(rIdx))
					log.CtxErrorw(ctx, "failed to verify secondary SP bls signature", "secondary_sp_id", gvg.GetSecondarySpIds()[rIdx], "error", err.Error())
					return ErrInvalidSecondaryBlsSignature
				}
				secondarySignatures[rIdx] = signature
				metrics.ExecutorCounter.WithLabelValues(ExecutorSuccessDoneReplicatePiece).Inc()
			} else {
				rTask.SetNotAvailableSpIdx(int32(rIdx))
				metrics.ExecutorCounter.WithLabelValues(ExecutorFailureDoneReplicatePiece).Inc()
				return innerErr
			}
		}
		log.Debug("finish to done replicate")
		return nil
	}
	startReplicatePieceTime := time.Now()
	errChan := make(chan error)
	quitChan := make(chan struct{})
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		for segIdx := uint32(0); segIdx < segmentPieceCount; segIdx++ {
			pieceKey := e.baseApp.PieceOp().SegmentPieceKey(rTask.GetObjectInfo().Id.Uint64(), segIdx)
			startGetPieceTime := time.Now()
			segData, err := e.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
			metrics.PerfPutObjectTime.WithLabelValues("background_get_piece_time").Observe(time.Since(startGetPieceTime).Seconds())
			metrics.PerfPutObjectTime.WithLabelValues("background_get_piece_end_time").Observe(time.Since(time.Unix(rTask.GetCreateTime(), 0)).Seconds())
			if err != nil {
				log.CtxErrorw(ctx, "failed to get segment data from piece store", "error", err)
				rTask.SetError(err)
				errChan <- err
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
					errChan <- err
				}
				doReplicateECPiece(childCtx, segIdx, ecData, errChan)
			} else {
				doReplicateSegmentPiece(childCtx, segIdx, segData, errChan)
			}
		}
		metrics.PerfPutObjectTime.WithLabelValues("background_replicate_all_piece_time").Observe(time.Since(startReplicatePieceTime).Seconds())
		metrics.PerfPutObjectTime.WithLabelValues("background_replicate_all_piece_end_time").Observe(time.Since(startReplicatePieceTime).Seconds())
		doneTime := time.Now()
		err = doneReplicate(childCtx)
		metrics.PerfPutObjectTime.WithLabelValues("background_done_replicate_time").Observe(time.Since(doneTime).Seconds())
		metrics.PerfPutObjectTime.WithLabelValues("background_done_replicate_piece_end_time").Observe(time.Since(startReplicatePieceTime).Seconds())
		if err == nil {
			rTask.SetSecondarySignatures(secondarySignatures)
		}
		quitChan <- struct{}{}
	}()
	var replicateErr error
	for {
		select {
		case err = <-errChan:
			if replicateErr == nil {
				replicateErr = err
			}
			cancel()
		case <-quitChan:
			if replicateErr != nil {
				return replicateErr
			}
			return err
		}
	}
}

func (e *ExecuteModular) doReplicatePiece(ctx context.Context, waitGroup *sync.WaitGroup, rTask coretask.ReplicatePieceTask,
	spEndpoint string, segmentIdx uint32, redundancyIdx int32, data []byte) (err error) {
	var signature []byte
	rTask.AppendLog(fmt.Sprintf("executor-begin-replicate-piece-sIdx:%d-rIdx-%d", segmentIdx, redundancyIdx))
	startTime := time.Now()
	defer func() {
		if err != nil {
			rTask.AppendLog(fmt.Sprintf("executor-end-replicate-piece-sIdx:%d-rIdx-%d-error:%s-endpoint:%s",
				segmentIdx, redundancyIdx, err.Error(), spEndpoint))
			metrics.ExecutorCounter.WithLabelValues(ExecutorFailureReplicateOnePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExecutorFailureReplicateOnePiece).Observe(time.Since(startTime).Seconds())
		} else {
			rTask.AppendLog(fmt.Sprintf("executor-end-replicate-piece-sIdx:%d-rIdx-%d-endpoint:%s",
				segmentIdx, redundancyIdx, spEndpoint))
			metrics.ExecutorCounter.WithLabelValues(ExecutorSuccessReplicateOnePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExecutorSuccessReplicateOnePiece).Observe(time.Since(startTime).Seconds())
		}
		waitGroup.Done()
	}()
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetGlobalVirtualGroupId(), rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), segmentIdx, redundancyIdx, int64(len(data)))
	receive.SetPieceChecksum(hash.GenerateChecksum(data))
	ctx = log.WithValue(ctx, log.CtxKeyTask, receive.Key().String())
	signTime := time.Now()
	signature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_cost").Observe(time.Since(signTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_end").Observe(time.Since(startTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign receive task", "segment_idx", segmentIdx,
			"redundancy_idx", redundancyIdx, "error", err)
		return
	}
	receive.SetSignature(signature)
	replicateOnePieceTime := time.Now()
	if err = retry.Do(func() error {
		return e.baseApp.GfSpClient().ReplicatePieceToSecondary(ctx, spEndpoint, receive, data)
	}, retry.Context(ctx), RtyAttem, RtyDelay, RtyErr); err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "segment_idx", segmentIdx,
			"redundancy_idx", redundancyIdx, "error", err)
		return err
	}
	metrics.PerfPutObjectTime.WithLabelValues("background_replicate_one_piece_cost").Observe(time.Since(replicateOnePieceTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_replicate_one_piece_end").Observe(time.Since(startTime).Seconds())
	log.CtxDebugw(ctx, "succeed to replicate piece", "segment_idx", segmentIdx,
		"redundancy_idx", redundancyIdx)
	return
}

func (e *ExecuteModular) doneReplicatePiece(ctx context.Context, rTask coretask.ReplicatePieceTask,
	spEndpoint string, redundancyIdx int32) ([]byte, error) {
	var (
		err           error
		signature     []byte
		taskSignature []byte
	)
	rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-rIdx-%d", redundancyIdx))
	startTime := time.Now()
	defer func() {
		if err != nil {
			rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-rIdx-%d-error:%s-endpoint:%s",
				redundancyIdx, err.Error(), spEndpoint))
			metrics.ExecutorCounter.WithLabelValues(ExecutorFailureDoneReplicatePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExecutorFailureDoneReplicatePiece).Observe(time.Since(startTime).Seconds())
		} else {
			rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-rIdx-%d-endpoint:%s",
				redundancyIdx, spEndpoint))
			metrics.ExecutorCounter.WithLabelValues(ExecutorSuccessDoneReplicatePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExecutorSuccessDoneReplicatePiece).Observe(time.Since(startTime).Seconds())
		}
	}()

	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetGlobalVirtualGroupId(), rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), 0, redundancyIdx, 0)
	receive.SetFinished(true)
	signTime := time.Now()
	taskSignature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_cost").Observe(time.Since(signTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_end").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign done receive task", "redundancy_idx", redundancyIdx, "error", err)
		return nil, err
	}
	receive.SetSignature(taskSignature)
	doneReplicateTime := time.Now()
	if err = retry.Do(func() error {
		signature, err = e.baseApp.GfSpClient().DoneReplicatePieceToSecondary(ctx, spEndpoint, receive)
		return err
	}, retry.Context(ctx), RtyAttem, RtyDelay, RtyErr); err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece", "endpoint", spEndpoint,
			"redundancy_idx", redundancyIdx, "error", err)
		return nil, err
	}
	metrics.PerfPutObjectTime.WithLabelValues("background_done_receive_http_cost").Observe(time.Since(doneReplicateTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_done_receive_http_end").Observe(time.Since(signTime).Seconds())
	if int(redundancyIdx+1) >= len(rTask.GetObjectInfo().GetChecksums()) {
		log.CtxErrorw(ctx, "failed to done replicate piece, replicate idx out of bounds", "redundancy_idx", redundancyIdx)
		return nil, ErrReplicateIdsOutOfBounds
	}

	// TODO:
	// veritySignatureTime := time.Now()
	// TODO get gvgID and blsPubKey from task, bls pub key already injected via key manager for current sp
	// var blsPubKey bls.PublicKey
	// err = veritySignature(ctx, rTask.GetObjectInfo().Id.Uint64(), rTask.GetGlobalVirtualGroupId(), integrity,
	//	storagetypes.GenerateHash(rTask.GetObjectInfo().GetChecksums()[:]), signature, blsPubKey)
	// metrics.PerfUploadTimeHistogram.WithLabelValues("background_verity_seal_signature_time").Observe(time.Since(veritySignatureTime).Seconds())
	// metrics.PerfPutObjectTime.WithLabelValues("background_verity_seal_signature_end_time").Observe(time.Since(signTime).Seconds())
	// if err != nil {
	// 	log.CtxErrorw(ctx, "failed verify secondary signature", "endpoint", spEndpoint,
	// 		"redundancy_idx", redundancyIdx, "error", err)
	// 	return nil, err
	// }
	log.CtxDebugw(ctx, "succeed to done replicate", "endpoint", spEndpoint, "redundancy_idx", redundancyIdx)
	return signature, nil
}

func veritySecondarySpBlsSignature(secondarySp *sptypes.StorageProvider, signature, sigDoc []byte) error {
	publicKey, err := bls.PublicKeyFromBytes(secondarySp.BlsKey)
	if err != nil {
		return err
	}
	sig, err := bls.SignatureFromBytes(signature)
	if err != nil {
		return err
	}
	if !sig.Verify(publicKey, sigDoc) {
		return fmt.Errorf("failed to verify SP[%d] bls signature", secondarySp.Id)
	}
	return nil
}
