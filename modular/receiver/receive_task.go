package receiver

import (
	"bytes"
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	ErrDanglingTask        = gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 80001, "OoooH... request lost, try again later")
	ErrRepeatedTask        = gfsperrors.Register(module.ReceiveModularName, http.StatusBadRequest, 80002, "request repeated")
	ErrExceedTask          = gfsperrors.Register(module.ReceiveModularName, http.StatusServiceUnavailable, 80003, "OoooH... request exceed, try again later")
	ErrUnfinishedTask      = gfsperrors.Register(module.ReceiveModularName, http.StatusForbidden, 80004, "replicate piece unfinished")
	ErrInvalidDataChecksum = gfsperrors.Register(module.ReceiveModularName, http.StatusNotAcceptable, 80005, "verify data checksum failed")
	ErrPieceStore          = gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 85101, "server slipped away, try again later")
	ErrGfSpDB              = gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 85201, "server slipped away, try again later")
)

func (r *ReceiveModular) HandleReceivePieceTask(
	ctx context.Context,
	task task.ReceivePieceTask,
	data []byte) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to pre receive piece, task pointer dangling")
		return ErrDanglingTask
	}
	if r.receiveQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "has repeat receive task")
		return ErrRepeatedTask
	}
	err := r.receiveQueue.Push(task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push receive task to queue", "error", err)
		return ErrExceedTask
	}
	defer r.receiveQueue.PopByKey(task.Key())
	checksum := hash.GenerateChecksum(data)
	if !bytes.Equal(checksum, task.GetPieceChecksum()) {
		return ErrInvalidDataChecksum
	}
	pieceKey := r.baseApp.PieceOp().ECPieceKey(task.GetObjectInfo().Id.Uint64(),
		uint32(task.GetPieceIdx()), task.GetReplicateIdx())
	if err = r.baseApp.GfSpDB().SetReplicatePieceChecksum(task.GetObjectInfo().Id.Uint64(),
		task.GetReplicateIdx(), uint32(task.GetPieceIdx()), task.GetPieceChecksum()); err != nil {
		log.CtxErrorw(ctx, "failed to set checksum to db", "error", err)
		return ErrGfSpDB
	}
	if err = r.baseApp.PieceStore().PutPiece(ctx, pieceKey, data); err != nil {
		return ErrPieceStore
	}
	log.CtxDebugw(ctx, "succeed to receive piece data")
	return nil
}

func (r *ReceiveModular) HandleDoneReceivePieceTask(
	ctx context.Context,
	task task.ReceivePieceTask) (
	[]byte, []byte, error) {
	if err := r.receiveQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push receive task", "error", err)
		return nil, nil, ErrExceedTask
	}
	defer r.receiveQueue.PopByKey(task.Key())
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to done receive task, task pointer dangling")
		return nil, nil, ErrDanglingTask
	}
	segmentCount := r.baseApp.PieceOp().SegmentCount(task.GetObjectInfo().GetPayloadSize(),
		task.GetStorageParams().VersionedParams.GetMaxSegmentSize())
	checksums, err := r.baseApp.GfSpDB().GetAllReplicatePieceChecksum(
		task.GetObjectInfo().Id.Uint64(), task.GetReplicateIdx(), segmentCount)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get checksum from db", "error", err)
		return nil, nil, ErrGfSpDB
	}
	if len(checksums) != int(segmentCount) {
		log.CtxErrorw(ctx, "replicate piece unfinished")
		return nil, nil, ErrUnfinishedTask
	}
	signature, integrity, err := r.baseApp.GfSpClient().SignIntegrityHash(ctx,
		task.GetObjectInfo().Id.Uint64(), checksums)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign the integrity hash", "error", err)
		return nil, nil, err
	}
	integrityMeta := &corespdb.IntegrityMeta{
		ObjectID:          task.GetObjectInfo().Id.Uint64(),
		IntegrityChecksum: integrity,
		PieceChecksumList: checksums,
		Signature:         signature,
	}
	err = r.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta)
	if err != nil {
		log.CtxErrorw(ctx, "failed to write integrity hash to db", "error", err)
		return nil, nil, ErrGfSpDB
	}
	if err = r.baseApp.GfSpDB().DeleteAllReplicatePieceChecksum(
		task.GetObjectInfo().Id.Uint64(), task.GetReplicateIdx(), segmentCount); err != nil {
		log.CtxErrorw(ctx, "failed to delete all replicate piece checksum", "error", err)
	}
	// the manager dispatch the reset task to confirm whether seal on chain as secondary sp.
	task.SetError(nil)
	if err = r.baseApp.GfSpClient().ReportTask(ctx, task); err != nil {
		log.CtxErrorw(ctx, "failed to report receive task for confirming seal", "error", err)
	}
	log.CtxDebugw(ctx, "succeed to done receive piece")
	return integrity, signature, nil
}
