package executor

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"math"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
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
	defer func() {
		task.SetError(err)
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
		e.baseApp.TaskPriority(rAppTask), e.baseApp.OperateAddress())
	approvals, err = e.AskReplicatePieceApproval(ctx, rAppTask, int(low),
		int(high), e.askReplicateApprovalTimeout)
	if err != nil {
		log.CtxErrorw(ctx, "failed get approvals", "error", err)
		return
	}
	err = e.handleReplicatePiece(ctx, task, approvals)
	if err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "error", err)
		return
	}
	log.CtxDebugw(ctx, "succeed to replicate all pieces")

	blsSig, err := bls.MultipleSignaturesFromBytes(task.GetSecondarySignatures())
	if err != nil {
		return
	}
	// combine seal object
	sealMsg := &storagetypes.MsgSealObject{
		Operator:                    e.baseApp.OperateAddress(),
		BucketName:                  task.GetObjectInfo().GetBucketName(),
		ObjectName:                  task.GetObjectInfo().GetObjectName(),
		GlobalVirtualGroupId:        task.GetObjectInfo().GetLocalVirtualGroupId(),
		SecondarySpBlsAggSignatures: bls.AggregateSignatures(blsSig).Marshal(),
	}
	sealErr := e.sealObject(ctx, task, sealMsg)
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
					metrics.ReplicateSucceedCounter.WithLabelValues(e.Name()).Inc()
				} else {
					metrics.ReplicateFailedCounter.WithLabelValues(e.Name()).Inc()
				}
			}
		}
	}
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
			segData, err := e.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
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
	metrics.ReplicatePieceSizeCounter.WithLabelValues(e.Name()).Add(float64(len(data)))
	startTime := time.Now()
	defer func() {
		metrics.ReplicatePieceTimeHistogram.WithLabelValues(e.Name()).Observe(time.Since(startTime).Seconds())
		waitGroup.Done()
	}()
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), replicateIdx, int32(pieceIdx), int64(len(data)))
	receive.SetPieceChecksum(hash.GenerateChecksum(data))
	ctx = log.WithValue(ctx, log.CtxKeyTask, receive.Key().String())
	signature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign receive task", "replicate_idx", replicateIdx,
			"piece_idx", pieceIdx, "error", err)
		return
	}
	receive.SetSignature(signature)
	err = e.baseApp.GfSpClient().ReplicatePieceToSecondary(ctx,
		approval.GetApprovedSpEndpoint(), approval, receive, data)
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
	receive := &gfsptask.GfSpReceivePieceTask{}
	receive.InitReceivePieceTask(rTask.GetObjectInfo(), rTask.GetStorageParams(),
		e.baseApp.TaskPriority(rTask), replicateIdx, -1, 0)
	taskSignature, err = e.baseApp.GfSpClient().SignReceiveTask(ctx, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign done receive task",
			"replicate_idx", replicateIdx, "error", err)
		return nil, nil, err
	}
	receive.SetSignature(taskSignature)
	integrity, signature, err = e.baseApp.GfSpClient().DoneReplicatePieceToSecondary(ctx,
		approval.GetApprovedSpEndpoint(), approval, receive)
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece",
			"endpoint", approval.GetApprovedSpEndpoint(),
			"replicate_idx", replicateIdx, "error", err)
		return nil, nil, err
	}

	// TODO get gvgId and blsPubKey from task
	var gvgId uint32
	var blsPubKey bls.PublicKey

	err = veritySignature(ctx, rTask.GetObjectInfo().Id.Uint64(), gvgId, integrity,
		storagetypes.GenerateHash(rTask.GetObjectInfo().GetChecksums()[:]), signature, blsPubKey)
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

func veritySignature(ctx context.Context, objectID uint64, gvgId uint32, integrity []byte, expectedIntegrity []byte, signature []byte, blsPubKey bls.PublicKey) error {
	if !bytes.Equal(expectedIntegrity, integrity) {
		log.CtxErrorw(ctx, "replicate sp invalid integrity", "integrity", hex.EncodeToString(integrity),
			"expect", hex.EncodeToString(expectedIntegrity))
		return ErrInvalidIntegrity
	}
	originMsgHash := storagetypes.NewSecondarySpSignDoc(sdkmath.NewUint(objectID), gvgId, integrity).GetSignBytes()
	err := storagetypes.VerifyBlsSignature(blsPubKey, originMsgHash, signature)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify signature", "error", err)
		return err
	}
	return nil
}
