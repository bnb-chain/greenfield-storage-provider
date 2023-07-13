package receiver

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-common/go/hash"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrDanglingTask        = gfsperrors.Register(module.ReceiveModularName, http.StatusBadRequest, 80001, "OoooH... request lost, try again later")
	ErrRepeatedTask        = gfsperrors.Register(module.ReceiveModularName, http.StatusNotAcceptable, 80002, "request repeated")
	ErrUnfinishedTask      = gfsperrors.Register(module.ReceiveModularName, http.StatusForbidden, 80003, "replicate piece unfinished")
	ErrInvalidDataChecksum = gfsperrors.Register(module.ReceiveModularName, http.StatusNotAcceptable, 80004, "verify data checksum failed")
	ErrPieceStore          = gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 85101, "server slipped away, try again later")
	ErrGfSpDB              = gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 85201, "server slipped away, try again later")
)

func (r *ReceiveModular) HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask, data []byte) error {
	var (
		err error
	)
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()

	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre receive piece due to pointer dangling")
		err = ErrDanglingTask
		return ErrDanglingTask
	}
	checkHasTime := time.Now()
	if r.receiveQueue.Has(task.Key()) {
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_check_has_time").Observe(time.Since(checkHasTime).Seconds())
		log.CtxErrorw(ctx, "has repeat receive task")
		err = ErrRepeatedTask
		return ErrRepeatedTask
	}
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_check_has_time").Observe(time.Since(checkHasTime).Seconds())

	pushTime := time.Now()
	err = r.receiveQueue.Push(task)
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_push_time").Observe(time.Since(pushTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to push receive task to queue", "error", err)
		return err
	}
	defer r.receiveQueue.PopByKey(task.Key())
	checksum := hash.GenerateChecksum(data)
	if !bytes.Equal(checksum, task.GetPieceChecksum()) {
		err = ErrInvalidDataChecksum
		return ErrInvalidDataChecksum
	}
	var pieceKey string
	if task.GetObjectInfo().GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
		pieceKey = r.baseApp.PieceOp().ECPieceKey(task.GetObjectInfo().Id.Uint64(),
			uint32(task.GetPieceIdx()), task.GetReplicateIdx())
	} else {
		pieceKey = r.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(),
			uint32(task.GetPieceIdx()))
	}
	setDBTime := time.Now()
	if err = r.baseApp.GfSpDB().SetReplicatePieceChecksum(task.GetObjectInfo().Id.Uint64(),
		task.GetReplicateIdx(), uint32(task.GetPieceIdx()), task.GetPieceChecksum()); err != nil {
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_set_mysql_time").Observe(time.Since(setDBTime).Seconds())
		log.CtxErrorw(ctx, "failed to set checksum to db", "error", err)
		err = ErrGfSpDB
		return ErrGfSpDB
	}
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_set_mysql_time").Observe(time.Since(setDBTime).Seconds())

	setPieceTime := time.Now()
	if err = r.baseApp.PieceStore().PutPiece(ctx, pieceKey, data); err != nil {
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_set_piece_time").Observe(time.Since(setPieceTime).Seconds())
		err = ErrPieceStore
		return ErrPieceStore
	}
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_set_piece_time").Observe(time.Since(setPieceTime).Seconds())
	log.CtxDebugw(ctx, "succeed to receive piece data")
	return nil
}

func (r *ReceiveModular) HandleDoneReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, error) {
	var err error
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()

	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre receive piece due to pointer dangling")
		err = ErrDanglingTask
		return nil, ErrDanglingTask
	}
	pushTime := time.Now()
	if err = r.receiveQueue.Push(task); err != nil {
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_push_time").Observe(time.Since(pushTime).Seconds())
		log.CtxErrorw(ctx, "failed to push receive task", "error", err)
		return nil, err
	}
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_push_time").Observe(time.Since(pushTime).Seconds())

	defer r.receiveQueue.PopByKey(task.Key())
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to done receive task, pointer dangling")
		err = ErrDanglingTask
		return nil, ErrDanglingTask
	}
	segmentCount := r.baseApp.PieceOp().SegmentPieceCount(task.GetObjectInfo().GetPayloadSize(),
		task.GetStorageParams().VersionedParams.GetMaxSegmentSize())

	getChecksumsTime := time.Now()
	pieceChecksums, err := r.baseApp.GfSpDB().GetAllReplicatePieceChecksum(
		task.GetObjectInfo().Id.Uint64(), task.GetReplicateIdx(), segmentCount)
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_get_checksums_time").Observe(time.Since(getChecksumsTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get checksum from db", "error", err)
		err = ErrGfSpDB
		return nil, ErrGfSpDB
	}
	if len(pieceChecksums) != int(segmentCount) {
		log.CtxErrorw(ctx, "replicate piece unfinished")
		err = ErrUnfinishedTask
		return nil, ErrUnfinishedTask
	}
	signTime := time.Now()
	signature, err := r.baseApp.GfSpClient().SignSecondarySealBls(ctx,
		task.GetObjectInfo().Id.Uint64(), task.GetGlobalVirtualGroupId(), task.GetObjectInfo().GetChecksums())

	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_sign_time").Observe(time.Since(signTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign the integrity hash", "error", err)
		return nil, err
	}

	integrityMeta := &corespdb.IntegrityMeta{
		ObjectID:          task.GetObjectInfo().Id.Uint64(),
		RedundancyIndex:   int32(task.GetReplicateIdx()),
		IntegrityChecksum: hash.GenerateIntegrityHash(pieceChecksums),
		PieceChecksumList: pieceChecksums,
	}
	setIntegrityTime := time.Now()
	err = r.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta)
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_set_integrity_time").Observe(time.Since(setIntegrityTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to write integrity meta to db", "error", err)
		err = ErrGfSpDB
		return nil, ErrGfSpDB
	}
	deletePieceHashTime := time.Now()
	if err = r.baseApp.GfSpDB().DeleteAllReplicatePieceChecksum(
		task.GetObjectInfo().Id.Uint64(), task.GetReplicateIdx(), segmentCount); err != nil {
		log.CtxErrorw(ctx, "failed to delete all replicate piece checksum", "error", err)
		// ignore the error,let the request go, the background task will gc the meta again later
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_delete_piece_hash_time").Observe(time.Since(deletePieceHashTime).Seconds())
	}
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_delete_piece_hash_time").Observe(time.Since(deletePieceHashTime).Seconds())
	// the manager dispatch the task to confirm whether seal on chain as secondary sp.
	task.SetError(nil)

	// bucket migration doesn't need to report task, because it's not scheduled by manager
	if task.GetBucketMigration() {
		return signature, nil
	}
	reportTime := time.Now()
	if err = r.baseApp.GfSpClient().ReportTask(ctx, task); err != nil {
		log.CtxErrorw(ctx, "failed to report receive task for confirming seal", "error", err)
		// ignore the error,let the request go, the background task will gc the unsealed data later
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_report_time").Observe(time.Since(reportTime).Seconds())
	}
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_server_done_report_time").Observe(time.Since(reportTime).Seconds())
	log.CtxDebugw(ctx, "succeed to done receive piece")
	return signature, nil
}

func (r *ReceiveModular) QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error) {
	receiveTasks, _ := taskqueue.ScanTQueueBySubKey(r.receiveQueue, subKey)
	return receiveTasks, nil
}
