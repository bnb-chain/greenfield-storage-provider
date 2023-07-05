package executor

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (e *ExecuteModular) HandleReplicatePieceTask(ctx context.Context, task coretask.ReplicatePieceTask) {
	var (
		err       error
		approvals []*gfsptask.GfSpReplicatePieceApprovalTask
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
	low := task.GetStorageParams().VersionedParams.GetRedundantDataChunkNum() +
		task.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()
	high := math.Ceil(float64(low) * e.askReplicateApprovalExFactor)
	rAppTask := &gfsptask.GfSpReplicatePieceApprovalTask{}
	rAppTask.InitApprovalReplicatePieceTask(task.GetObjectInfo(), task.GetStorageParams(),
		e.baseApp.TaskPriority(rAppTask), e.baseApp.OperatorAddress())
	askReplicateApprovalTime := time.Now()
	task.AppendLog("executor-begin-p2p")
	approvals, err = e.AskReplicatePieceApproval(ctx, rAppTask, int(low),
		int(high), e.askReplicateApprovalTimeout)
	task.AppendLog("executor-end-p2p")
	metrics.PerfPutObjectTime.WithLabelValues("background_ask_p2p_approval_cost").Observe(time.Since(askReplicateApprovalTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_ask_p2p_approval_end").Observe(time.Since(startReplicateTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed get approvals", "error", err)
		metrics.ExecutorCounter.WithLabelValues(ExeutorFailureP2P).Inc()
		metrics.ExecutorTime.WithLabelValues(ExeutorFailureP2P).Observe(time.Since(askReplicateApprovalTime).Seconds())
		return
	} else {
		metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessP2P).Inc()
		metrics.ExecutorTime.WithLabelValues(ExeutorSuccessP2P).Observe(time.Since(askReplicateApprovalTime).Seconds())
	}
	replicatePieceTotalTime := time.Now()
	task.AppendLog("executor-begin-replicate-object")
	err = e.handleReplicatePiece(ctx, task, approvals)
	metrics.PerfPutObjectTime.WithLabelValues("background_replicate_object_cost").Observe(time.Since(replicatePieceTotalTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_task_replicate_object_end").Observe(time.Since(startReplicateTime).Seconds())
	if err != nil {
		task.AppendLog(fmt.Sprintf("executor-end-replicate-object-error%s", err.Error()))
		log.CtxErrorw(ctx, "failed to replicate piece", "error", err)
		metrics.ExecutorCounter.WithLabelValues(ExeutorFailureReplicateAllPiece).Inc()
		metrics.ExecutorTime.WithLabelValues(ExeutorFailureReplicateAllPiece).Observe(time.Since(replicatePieceTotalTime).Seconds())
		return
	} else {
		task.AppendLog("executor-end-replicate-object")
		metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessReplicateAllPiece).Inc()
		metrics.ExecutorTime.WithLabelValues(ExeutorSuccessReplicateAllPiece).Observe(time.Since(replicatePieceTotalTime).Seconds())
	}
	log.CtxDebugw(ctx, "succeed to replicate all pieces")
	// combine seal object
	sealMsg := &storagetypes.MsgSealObject{
		Operator:              e.baseApp.OperatorAddress(),
		BucketName:            task.GetObjectInfo().GetBucketName(),
		ObjectName:            task.GetObjectInfo().GetObjectName(),
		SecondarySpAddresses:  task.GetSecondaryAddresses(),
		SecondarySpSignatures: task.GetSecondarySignatures(),
	}
	sealTime := time.Now()
	sealErr := e.sealObject(ctx, task, sealMsg)
	metrics.PerfPutObjectTime.WithLabelValues("background_seal_object_cost").Observe(time.Since(sealTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_task_seal_object_end").Observe(time.Since(startReplicateTime).Seconds())
	if sealErr == nil {
		task.SetSealed(true)
	}
	log.CtxDebugw(ctx, "finish combine seal object", "error", sealErr)
}

func (e *ExecuteModular) AskReplicatePieceApproval(ctx context.Context, task coretask.ApprovalReplicatePieceTask,
	low, high int, timeout int64) (
	[]*gfsptask.GfSpReplicatePieceApprovalTask, error) {
	var (
		err       error
		approvals []*gfsptask.GfSpReplicatePieceApprovalTask
		spInfo    *sptypes.StorageProvider
	)
	approvals, err = e.baseApp.GfSpClient().AskSecondaryReplicatePieceApproval(ctx, task, low, high, timeout)
	if err != nil {
		return nil, err
	}
	if len(approvals) < low {
		log.CtxErrorw(ctx, "failed to get sufficient sp approval from p2p protocol")
		return nil, ErrInsufficientApproval
	}
	for _, approval := range approvals {
		spInfo, err = e.baseApp.GfSpDB().GetSpByAddress(
			approval.GetApprovedSpOperatorAddress(),
			spdb.OperatorAddressType)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get sp info from db", "error", err)
			continue
		}
		approval.SetApprovedSpEndpoint(spInfo.GetEndpoint())
		approval.SetApprovedSpApprovalAddress(spInfo.GetApprovalAddress())
	}
	if len(approvals) < low {
		log.CtxErrorw(ctx, "failed to get sufficient sp info from db")
		return nil, ErrGfSpDB
	}
	return approvals, nil
}

func (e *ExecuteModular) handleReplicatePiece(ctx context.Context, rTask coretask.ReplicatePieceTask,
	backUpApprovals []*gfsptask.GfSpReplicatePieceApprovalTask) (err error) {
	var (
		wg       sync.WaitGroup
		pieceKey string
		segCount = e.baseApp.PieceOp().SegmentPieceCount(
			rTask.GetObjectInfo().GetPayloadSize(),
			rTask.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		replCount = rTask.GetStorageParams().VersionedParams.GetRedundantDataChunkNum() +
			rTask.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()
		record              = make([]bool, replCount)
		secondaryAddresses  = make([]string, replCount)
		secondarySignatures = make([][]byte, replCount)
		approvals           = make([]coretask.ApprovalReplicatePieceTask, replCount)
		finish              bool
	)
	resetApprovals := func() (bool, error) {
		doneAll := true
		for rIdx, done := range record {
			if !done {
				doneAll = false
				if len(backUpApprovals) == 0 {
					return finish, ErrExhaustedApproval
				}
				approvals[rIdx] = backUpApprovals[0]
				backUpApprovals = backUpApprovals[1:]
			}
		}
		return doneAll, nil
	}
	doReplicateECPiece := func(pieceIdx uint32, data [][]byte) {
		for rIdx, done := range record {
			if !done {
				wg.Add(1)
				go e.doReplicatePiece(ctx, &wg, rTask, approvals[rIdx],
					uint32(rIdx), pieceIdx, data[rIdx])
			}
		}
		wg.Wait()
	}
	doReplicateSegmentPiece := func(pieceIdx uint32, data []byte) {
		for rIdx, done := range record {
			if !done {
				wg.Add(1)
				go e.doReplicatePiece(ctx, &wg, rTask, approvals[rIdx],
					uint32(rIdx), pieceIdx, data)
			}
		}
		wg.Wait()
	}
	doneReplicate := func() {
		for rIdx, done := range record {
			if !done {
				_, signature, innerErr := e.doneReplicatePiece(ctx, rTask, approvals[rIdx], uint32(rIdx))
				if innerErr == nil {
					secondaryAddresses[rIdx] = approvals[rIdx].GetApprovedSpOperatorAddress()
					secondarySignatures[rIdx] = signature
					record[rIdx] = true
				}
			}
		}
	}
	startPieceTime := time.Now()
	for {
		finish, err = resetApprovals()
		if err != nil {
			log.CtxErrorw(ctx, "failed to pick up sp", "error", err)
			return err
		}
		if finish {
			rTask.SetSecondaryAddresses(secondaryAddresses)
			rTask.SetSecondarySignatures(secondarySignatures)
			log.CtxDebugw(ctx, "success to replicate all pieces")
			return nil
		}
		for pIdx := uint32(0); pIdx < segCount; pIdx++ {
			pieceKey = e.baseApp.PieceOp().SegmentPieceKey(rTask.GetObjectInfo().Id.Uint64(), pIdx)
			getPieceTime := time.Now()
			segData, err := e.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
			metrics.PerfPutObjectTime.WithLabelValues("background_get_piece_cost").Observe(time.Since(getPieceTime).Seconds())
			metrics.PerfPutObjectTime.WithLabelValues("background_get_piece_end").Observe(time.Since(startPieceTime).Seconds())
			if err != nil {
				log.CtxErrorw(ctx, "failed to get segment data form piece store", "error", err)
				rTask.SetError(err)
				return err
			}
			if rTask.GetObjectInfo().GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
				ecData, err := redundancy.EncodeRawSegment(segData,
					int(rTask.GetStorageParams().VersionedParams.GetRedundantDataChunkNum()),
					int(rTask.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()))
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
		doneReplicate()
	}
}

func (e *ExecuteModular) doReplicatePiece(ctx context.Context, waitGroup *sync.WaitGroup, rTask coretask.ReplicatePieceTask,
	approval coretask.ApprovalReplicatePieceTask, replicateIdx uint32, pieceIdx uint32, data []byte) (err error) {
	var signature []byte
	rTask.AppendLog(fmt.Sprintf("executor-begin-replicate-piece-sidx:%d-ridx-%d", pieceIdx, replicateIdx))
	startTime := time.Now()
	defer func() {
		if err != nil {
			rTask.AppendLog(fmt.Sprintf("executor-end-replicate-piece-sidx:%d-ridx-%d-error:%s-endpoint:%s", pieceIdx, replicateIdx, err.Error(), approval.GetApprovedSpEndpoint()))
			metrics.ExecutorCounter.WithLabelValues(ExeutorFailureReplicateOnePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorFailureReplicateOnePiece).Observe(time.Since(startTime).Seconds())
		} else {
			rTask.AppendLog(fmt.Sprintf("executor-end-replicate-piece-sidx:%d-ridx-%d-endpoint:%s", pieceIdx, replicateIdx, approval.GetApprovedSpEndpoint()))
			metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessReplicateOnePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorSuccessReplicateOnePiece).Observe(time.Since(startTime).Seconds())
		}
		waitGroup.Done()
	}()
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetObjectInfo(), rTask.GetStorageParams(),
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
	err = e.baseApp.GfSpClient().ReplicatePieceToSecondary(ctx,
		approval.GetApprovedSpEndpoint(), approval, receive, data)
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
	approval coretask.ApprovalReplicatePieceTask, replicateIdx uint32) ([]byte, []byte, error) {
	var (
		err           error
		integrity     []byte
		signature     []byte
		taskSignature []byte
	)
	rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-ridx-%d", replicateIdx))
	startTime := time.Now()
	defer func() {
		if err != nil {
			rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-ridx-%d-error:%s-endpoint:%s", replicateIdx, err.Error(), approval.GetApprovedSpEndpoint()))
			metrics.ExecutorCounter.WithLabelValues(ExeutorFailureDoneReplicatePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorFailureDoneReplicatePiece).Observe(time.Since(startTime).Seconds())
		} else {
			rTask.AppendLog(fmt.Sprintf("executor-begin-done_replicate-piece-ridx-%d-endpoint:%s", replicateIdx, approval.GetApprovedSpEndpoint()))
			metrics.ExecutorCounter.WithLabelValues(ExeutorSuccessDoneReplicatePiece).Inc()
			metrics.ExecutorTime.WithLabelValues(ExeutorSuccessDoneReplicatePiece).Observe(time.Since(startTime).Seconds())
		}
	}()

	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), replicateIdx, -1, 0)
	signTime := time.Now()
	taskSignature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_cost").Observe(time.Since(signTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_sign_receive_end").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign done receive task",
			"replicate_idx", replicateIdx, "error", err)
		return nil, nil, err
	}
	receive.SetSignature(taskSignature)
	doneReplicateTime := time.Now()
	integrity, signature, err = e.baseApp.GfSpClient().DoneReplicatePieceToSecondary(ctx,
		approval.GetApprovedSpEndpoint(), approval, receive)
	metrics.PerfPutObjectTime.WithLabelValues("background_done_receive_http_cost").Observe(time.Since(doneReplicateTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("background_done_receive_http_end").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece",
			"endpoint", approval.GetApprovedSpEndpoint(),
			"replicate_idx", replicateIdx, "error", err)
		return nil, nil, err
	}
	if int(replicateIdx+1) >= len(rTask.GetObjectInfo().GetChecksums()) {
		log.CtxErrorw(ctx, "failed to done replicate piece, replicate idx out of bounds",
			"replicate_idx", replicateIdx,
			"secondary_sp_len", len(rTask.GetObjectInfo().GetSecondarySpAddresses()))
		return nil, nil, ErrReplicateIdsOutOfBounds
	}
	err = veritySignature(ctx, rTask.GetObjectInfo().Id.Uint64(), integrity,
		rTask.GetObjectInfo().GetChecksums()[replicateIdx+1],
		approval.GetApprovedSpOperatorAddress(),
		approval.GetApprovedSpApprovalAddress(), signature)
	if err != nil {
		log.CtxErrorw(ctx, "failed verify secondary signature",
			"endpoint", approval.GetApprovedSpEndpoint(),
			"replicate_idx", replicateIdx, "error", err)
		return nil, nil, err
	}
	log.CtxDebugw(ctx, "succeed to done replicate",
		"endpoint", approval.GetApprovedSpEndpoint(),
		"replicate_idx", replicateIdx)
	return integrity, signature, nil
}

func veritySignature(ctx context.Context, objectID uint64, integrity []byte, expectedIntegrity []byte,
	signOpAddress string, signApprovalAddress string, signature []byte) error {
	if !bytes.Equal(expectedIntegrity, integrity) {
		log.CtxErrorw(ctx, "replicate sp invalid integrity", "integrity", hex.EncodeToString(integrity),
			"expect", hex.EncodeToString(expectedIntegrity))
		return ErrInvalidIntegrity
	}
	signOp, err := sdk.AccAddressFromHexUnsafe(signOpAddress)
	if err != nil {
		log.CtxErrorw(ctx, "failed to parse sign op address", "error", err)
		return err
	}
	signApproval, err := sdk.AccAddressFromHexUnsafe(signApprovalAddress)
	if err != nil {
		log.CtxErrorw(ctx, "failed to parse sign approval address", "error", err)
		return err
	}
	originMsgHash := storagetypes.NewSecondarySpSignDoc(signOp,
		sdkmath.NewUint(objectID), integrity).GetSignBytes()
	err = storagetypes.VerifySignature(signApproval, sdk.Keccak256(originMsgHash), signature)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify signature", "error", err)
		return err
	}
	return nil
}
