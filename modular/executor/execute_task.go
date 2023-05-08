package executor

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/modular/manager"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingPointer         = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 40001, "OoooH.... request lost")
	ErrInsufficientApproval    = gfsperrors.Register(module.ExecuteModularName, http.StatusNotFound, 40002, "insufficient approvals from p2p")
	ErrUnsealed                = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 40003, "seal object on chain failed")
	ErrExhaustedApproval       = gfsperrors.Register(module.ExecuteModularName, http.StatusNotFound, 40004, "approvals exhausted")
	ErrInvalidIntegrity        = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40005, "secondary integrity hash verification failed")
	ErrSecondaryMismatch       = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40006, "secondary sp mismatch")
	ErrReplicateIdsOutOfBounds = gfsperrors.Register(module.ExecuteModularName, http.StatusNotAcceptable, 40007, "replicate idx out of bounds")
	ErrGfSpDB                  = gfsperrors.Register(module.ExecuteModularName, http.StatusInternalServerError, 45201, "server slipped away, try again later")
)

func (e *ExecuteModular) HandleSealObjectTask(
	ctx context.Context,
	task coretask.SealObjectTask) {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to handle seal object, task pointer dangling")
		task.SetError(ErrDanglingPointer)
		return
	}
	sealMsg := &storagetypes.MsgSealObject{
		Operator:              e.baseApp.OperateAddress(),
		BucketName:            task.GetObjectInfo().GetBucketName(),
		ObjectName:            task.GetObjectInfo().GetObjectName(),
		SecondarySpAddresses:  task.GetObjectInfo().GetSecondarySpAddresses(),
		SecondarySpSignatures: task.GetSecondarySignature(),
	}
	task.SetError(e.sealObject(ctx, task, sealMsg))
	log.CtxDebugw(ctx, "finish to handle seal object task", "error", task.Error())
}

func (e *ExecuteModular) sealObject(
	ctx context.Context,
	task coretask.ObjectTask,
	sealMsg *storagetypes.MsgSealObject) error {
	var err error
	for retry := int64(0); retry <= task.GetMaxRetry(); retry++ {
		err = e.baseApp.GfSpClient().SealObject(ctx, sealMsg)
		if err != nil {
			log.CtxErrorw(ctx, "failed to seal object", "retry", retry,
				"max_retry", task.GetMaxRetry(), "error", err)
			time.Sleep(time.Duration(e.listenSealRetryTimeout) * time.Second)
		}
	}
	// even though signer return error, maybe seal on chain successfully because
	// signer use the async mode, so ignore the error and listen directly
	err = e.listenSealObject(ctx, task.GetObjectInfo())
	if err != nil {
		metrics.SealObjectSucceedCounter.WithLabelValues(e.Name()).Inc()
	} else {
		metrics.SealObjectFailedCounter.WithLabelValues(e.Name()).Inc()
	}
	return err
}

func (e *ExecuteModular) listenSealObject(
	ctx context.Context,
	object *storagetypes.ObjectInfo) error {
	var err error
	for retry := 0; retry < e.maxListenSealRetry; retry++ {
		sealed, innerErr := e.baseApp.Consensus().ListenObjectSeal(ctx,
			object.Id.Uint64(), e.listenSealTimeoutHeight)
		if innerErr != nil {
			log.CtxErrorw(ctx, "failed to listen object seal", "retry", retry,
				"max_retry", e.maxListenSealRetry, "error", err)
			time.Sleep(time.Duration(e.listenSealRetryTimeout) * time.Second)
			err = innerErr
			continue
		}
		if !sealed {
			log.CtxErrorw(ctx, "failed to seal object on chain", "retry", retry,
				"max_retry", e.maxListenSealRetry, "error", err)
			err = ErrUnsealed
			continue
		}
		err = nil
		break
	}
	return err
}

func (e *ExecuteModular) HandleReceivePieceTask(
	ctx context.Context,
	task coretask.ReceivePieceTask) {
	if task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to handle receive piece confirm, task pointer dangling")
		return
	}
	var (
		err           error
		onChainObject *storagetypes.ObjectInfo
	)
	err = e.listenSealObject(ctx, task.GetObjectInfo())
	if err == nil {
		task.SetSealed(true)
	}
	log.CtxDebugw(ctx, "finish to listen seal object for receive piece task, "+
		"begin to check secondary sp", "error", err)

	onChainObject, err = e.baseApp.Consensus().QueryObjectInfo(ctx, task.GetObjectInfo().GetBucketName(),
		task.GetObjectInfo().GetObjectName())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get object info", "error", err)
		task.SetError(err)
		return
	}
	if onChainObject.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to confirm receive task, object is unsealed")
		task.SetError(ErrUnsealed)
		return
	}
	// regardless of whether the sp is as secondary or not, it needs to be set to the
	// sealed state to let the manager clear the task.
	task.SetSealed(true)
	if int(task.GetReplicateIdx()) >= len(onChainObject.GetSecondarySpAddresses()) {
		log.CtxErrorw(ctx, "failed to confirm receive task, replicate idx out of bounds",
			"replicate_idx", task.GetReplicateIdx(),
			"secondary_sp_len", len(onChainObject.GetSecondarySpAddresses()))
		task.SetError(ErrReplicateIdsOutOfBounds)
		return
	}
	if onChainObject.GetSecondarySpAddresses()[int(task.GetReplicateIdx())] != e.baseApp.OperateAddress() {
		log.CtxErrorw(ctx, "failed to confirm receive task, secondary sp mismatch",
			"expect", onChainObject.GetSecondarySpAddresses()[int(task.GetReplicateIdx())],
			"current", e.baseApp.OperateAddress())
		task.SetError(ErrSecondaryMismatch)
		err = e.baseApp.GfSpDB().DeleteObjectIntegrity(task.GetObjectInfo().Id.Uint64())
		if err != nil {
			log.CtxErrorw(ctx, "failed to delete integrity")
		}
		var pieceKey string
		segmentCount := e.baseApp.PieceOp().SegmentCount(onChainObject.GetPayloadSize(),
			task.GetStorageParams().GetMaxPayloadSize())
		for i := uint32(0); i < segmentCount; i++ {
			if task.GetObjectInfo().GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
				pieceKey = e.baseApp.PieceOp().ECPieceKey(onChainObject.Id.Uint64(),
					i, task.GetReplicateIdx())
			} else {
				pieceKey = e.baseApp.PieceOp().SegmentPieceKey(onChainObject.Id.Uint64(), i)
			}
			err = e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
			if err != nil {
				log.CtxErrorw(ctx, "failed to delete piece data", "piece_key", pieceKey)
			}
		}
		return
	}
	log.CtxDebugw(ctx, "success to handle confirm receive piece task")
}

func (e *ExecuteModular) HandleGCObjectTask(
	ctx context.Context,
	task coretask.GCObjectTask) {
	var err error
	defer func() {
		task.SetError(err)
	}()

	objects, endBlockNumber, err := e.baseApp.GfSpClient().ListDeletedObjectsByBlockNumberRange(
		ctx, e.baseApp.OperateAddress(), task.GetStartBlockNumber(),
		task.GetEndBlockNumber(), true)
	if err != nil {
		return
	}
	if len(objects) == 0 {
		task.SetCurrentBlockNumber(task.GetEndBlockNumber() + 1)
		return
	}

	cancel := func() bool {
		return errors.Is(e.ReportTask(ctx, task), manager.ErrCanceledTask)
	}

	var (
		objectInfo       *storagetypes.ObjectInfo
		deletingBlock    uint64
		deletingObjectID uint64
		deletingIdx      int
		pieceKey         string
		segmentCount     uint32
		params           *storagetypes.Params
	)
	params, err = e.baseApp.GfSpDB().GetStorageParams()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get params", "error", err)
		return
	}
	for {
		log.CtxDebugw(ctx, "report gc object process", "info", task.Info())
		if cancel() {
			return
		}
		if deletingIdx >= len(objects) {
			deletingBlock = task.GetEndBlockNumber() + 1
		}
		if deletingBlock > endBlockNumber {
			deletingBlock = task.GetEndBlockNumber() + 1
		}
		task.SetCurrentBlockNumber(deletingBlock)
		if task.GetCurrentBlockNumber() > task.GetEndBlockNumber() {
			return
		}
		object := objects[deletingIdx]
		objectInfo = object.GetObjectInfo()
		deletingIdx++
		deletingObjectID = objectInfo.Id.Uint64()
		deletingBlock = uint64(object.GetDeleteAt())

		segmentCount = e.baseApp.PieceOp().SegmentCount(
			objectInfo.GetPayloadSize(), params.VersionedParams.GetMaxSegmentSize())
		for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
			pieceKey = e.baseApp.PieceOp().SegmentPieceKey(deletingObjectID, segIdx)
			err = e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
			if err != nil {
				log.CtxErrorw(ctx, "failed to delete segment piece",
					"piece_key", pieceKey, "error", err)
				return
			}
			log.CtxDebugw(ctx, "succeed to delete primary payload", "piece_key", pieceKey)
		}
		for rIdx, address := range objectInfo.GetSecondarySpAddresses() {
			if strings.Compare(e.baseApp.OperateAddress(), address) != 0 {
				continue
			}
			for segIdx := uint32(0); segIdx < segmentCount; segIdx++ {
				if objectInfo.GetRedundancyType() == storagetypes.REDUNDANCY_REPLICA_TYPE {
					pieceKey = e.baseApp.PieceOp().SegmentPieceKey(deletingObjectID, segIdx)
				} else {
					pieceKey = e.baseApp.PieceOp().ECPieceKey(deletingObjectID, segIdx, uint32(rIdx))
				}
				pieceKey = e.baseApp.PieceOp().ECPieceKey(deletingObjectID, segIdx, uint32(rIdx))
				err = e.baseApp.PieceStore().DeletePiece(ctx, pieceKey)
				if err != nil {
					log.CtxErrorw(ctx, "failed to delete replicate piece",
						"piece_key", pieceKey, "error", err)
					return
				}
				log.CtxDebugw(ctx, "succeed to delete secondary piece", "piece_key", pieceKey)
			}
			break
		}
		err = e.baseApp.GfSpDB().DeleteObjectIntegrity(deletingObjectID)
		if err != nil {
			log.CtxErrorw(ctx, "failed to delete integrity meta", "error", err)
			return
		}
		task.SetCurrentBlockNumber(deletingBlock)
		task.SetLastDeletedObjectId(deletingObjectID)
		metrics.GCObjectCounter.WithLabelValues(e.Name()).Inc()
	}
}

func (e *ExecuteModular) HandleGCZombiePieceTask(
	ctx context.Context,
	task coretask.GCZombiePieceTask) {
	log.CtxWarn(ctx, "gc zombie piece future support")
	return
}

func (e *ExecuteModular) HandleGCMetaTask(
	ctx context.Context,
	task coretask.GCMetaTask) {
	log.CtxWarn(ctx, "gc meta future support")
	return
}
