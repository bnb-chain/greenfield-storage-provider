package downloader

import (
	"context"
	"errors"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingPointer   = gfsperrors.Register(module.DownloadModularName, http.StatusInternalServerError, 30001, "OoooH.... request lost")
	ErrObjectUnsealed    = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30002, "object unsealed")
	ErrRepeatedTask      = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30003, "request repeated")
	ErrExceedBucketQuota = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30004, "bucket quota overflow")
	ErrExceedQueue       = gfsperrors.Register(module.DownloadModularName, http.StatusServiceUnavailable, 30005, "request exceed the limit, try again later")
	ErrInvalidParam      = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30006, "request params invalid")
	ErrNoSuchPiece       = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30007, "request params invalid, no such piece")
	ErrPieceStore        = gfsperrors.Register(module.DownloadModularName, http.StatusInternalServerError, 35101, "server slipped away, try again later")
	ErrGfSpDB            = gfsperrors.Register(module.DownloadModularName, http.StatusInternalServerError, 35201, "server slipped away, try again later")
)

func (d *DownloadModular) PreDownloadObject(ctx context.Context, task task.DownloadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed pre download object due to pointer dangling")
		return ErrDanglingPointer
	}
	if task.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to pre download object due to object unsealed")
		return ErrObjectUnsealed
	}
	if d.downloadQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "failed to pre download object due to task repeated")
		return ErrRepeatedTask
	}
	// TODO:: spilt check and add record steps, check in pre download, add record in post download
	if err := d.baseApp.GfSpDB().CheckQuotaAndAddReadRecord(
		&spdb.ReadRecord{
			BucketID:        task.GetBucketInfo().Id.Uint64(),
			ObjectID:        task.GetObjectInfo().Id.Uint64(),
			UserAddress:     task.GetUserAddress(),
			BucketName:      task.GetBucketInfo().GetBucketName(),
			ObjectName:      task.GetObjectInfo().GetObjectName(),
			ReadSize:        uint64(task.GetSize()),
			ReadTimestampUs: sqldb.GetCurrentTimestampUs(),
		},
		&spdb.BucketQuota{
			ReadQuotaSize: task.GetBucketInfo().GetChargedReadQuota() + d.bucketFreeQuota,
		},
	); err != nil {
		log.CtxErrorw(ctx, "failed to check bucket quota", "error", err)
		if errors.Is(err, sqldb.ErrCheckQuotaEnough) {
			return ErrExceedBucketQuota
		}
		// ignore the access db error, it is the system's inner error, will be let the request go.
	}
	// report the task to the manager for monitor the download task
	d.baseApp.GfSpClient().ReportTask(ctx, task)
	return nil
}

func (d *DownloadModular) HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) ([]byte, error) {
	var err error
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()
	if err = d.downloadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push download queue", "error", err)
		return nil, ErrExceedQueue
	}
	defer d.downloadQueue.PopByKey(task.Key())
	pieceInfos, err := SplitToSegmentPieceInfos(task, d.baseApp.PieceOp())
	if err != nil {
		log.CtxErrorw(ctx, "failed to generate piece info to download", "error", err)
		return nil, err
	}
	var data []byte
	for _, pInfo := range pieceInfos {
		piece, err := d.baseApp.PieceStore().GetPiece(ctx, pInfo.SegmentPieceKey,
			int64(pInfo.Offset), int64(pInfo.Length))
		if err != nil {
			log.CtxErrorw(ctx, "failed to get piece data from piece store", "error", err)
			return nil, ErrPieceStore
		}
		data = append(data, piece...)
	}
	return data, nil
}

type SegmentPieceInfo struct {
	SegmentPieceKey string
	Offset          uint64
	Length          uint64
}

func SplitToSegmentPieceInfos(
	task task.DownloadObjectTask,
	op piecestore.PieceOp) (
	[]*SegmentPieceInfo, error) {
	if task.GetObjectInfo().GetPayloadSize() == 0 ||
		task.GetLow() >= int64(task.GetObjectInfo().GetPayloadSize()) ||
		task.GetHigh() >= int64(task.GetObjectInfo().GetPayloadSize()) ||
		task.GetHigh() < task.GetLow() {
		log.Errorw("failed to parse params", "object_size",
			task.GetObjectInfo().GetPayloadSize(), "low", task.GetLow(), "high", task.GetHigh())
		return nil, ErrInvalidParam
	}
	segmentSize := task.GetStorageParams().VersionedParams.GetMaxSegmentSize()
	segmentCount := op.MaxSegmentSize(task.GetObjectInfo().GetPayloadSize(),
		task.GetStorageParams().VersionedParams.GetMaxSegmentSize())
	var (
		pieceInfos []*SegmentPieceInfo
		low        = uint64(task.GetLow())
		high       = uint64(task.GetHigh())
	)
	for segmentPieceIndex := uint64(0); segmentPieceIndex < uint64(segmentCount); segmentPieceIndex++ {
		currentStart := segmentPieceIndex * segmentSize
		currentEnd := (segmentPieceIndex+1)*segmentSize - 1
		if low > currentEnd {
			continue
		}
		if low > currentStart {
			currentStart = low
		}

		if high <= currentEnd {
			currentEnd = high
			offsetInPiece := currentStart - (segmentPieceIndex * segmentSize)
			lengthInPiece := currentEnd - currentStart + 1
			pieceInfos = append(pieceInfos, &SegmentPieceInfo{
				SegmentPieceKey: op.SegmentPieceKey(
					task.GetObjectInfo().Id.Uint64(),
					uint32(segmentPieceIndex)),
				Offset: offsetInPiece,
				Length: lengthInPiece,
			})
			// break to finish
			break
		} else {
			offsetInPiece := currentStart - (segmentPieceIndex * segmentSize)
			lengthInPiece := currentEnd - currentStart + 1
			pieceInfos = append(pieceInfos, &SegmentPieceInfo{
				SegmentPieceKey: op.SegmentPieceKey(
					task.GetObjectInfo().Id.Uint64(),
					uint32(segmentPieceIndex)),
				Offset: offsetInPiece,
				Length: lengthInPiece,
			})
		}
	}
	return pieceInfos, nil
}

func (d *DownloadModular) PostDownloadObject(
	ctx context.Context,
	task task.DownloadObjectTask) {
}

func (d *DownloadModular) PreDownloadPiece(
	ctx context.Context,
	task task.DownloadPieceTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed pre download piece due to pointer dangling")
		return ErrDanglingPointer
	}
	if task.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to pre download piece due to object unsealed")
		return ErrObjectUnsealed
	}
	if d.downloadQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "failed to pre download piece due to task repeated")
		return ErrRepeatedTask
	}

	if task.GetEnableCheck() {
		if err := d.baseApp.GfSpDB().CheckQuotaAndAddReadRecord(
			&spdb.ReadRecord{
				BucketID:        task.GetBucketInfo().Id.Uint64(),
				ObjectID:        task.GetObjectInfo().Id.Uint64(),
				UserAddress:     task.GetUserAddress(),
				BucketName:      task.GetBucketInfo().GetBucketName(),
				ObjectName:      task.GetObjectInfo().GetObjectName(),
				ReadSize:        task.GetTotalSize(),
				ReadTimestampUs: sqldb.GetCurrentTimestampUs(),
			},
			&spdb.BucketQuota{
				ReadQuotaSize: task.GetBucketInfo().GetChargedReadQuota() + d.bucketFreeQuota,
			},
		); err != nil {
			log.CtxErrorw(ctx, "failed to check bucket quota", "error", err)
			if errors.Is(err, sqldb.ErrCheckQuotaEnough) {
				return ErrExceedBucketQuota
			}
			// ignore the access db error, it is the system's inner error, will be let the request go.
		}
	}
	// report the task to the manager for monitor the download piece task
	d.baseApp.GfSpClient().ReportTask(ctx, task)
	return nil
}

func (d *DownloadModular) HandleDownloadPieceTask(
	ctx context.Context,
	task task.DownloadPieceTask) ([]byte, error) {
	var (
		err       error
		pieceData []byte
	)

	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()
	if err = d.downloadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push download queue", "error", err)
		return nil, ErrExceedQueue
	}
	defer d.downloadQueue.PopByKey(task.Key())

	if pieceData, err = d.baseApp.PieceStore().GetPiece(ctx, task.GetPieceKey(),
		int64(task.GetPieceOffset()), int64(task.GetPieceLength())); err != nil {
		log.CtxErrorw(ctx, "failed to get piece data from piece store", "task_info", task.Info(), "error", err)
		return nil, ErrPieceStore
	}
	return pieceData, nil
}

func (d *DownloadModular) PostDownloadPiece(
	ctx context.Context,
	task task.DownloadPieceTask) {
}

func (d *DownloadModular) PreChallengePiece(
	ctx context.Context,
	task task.ChallengePieceTask) error {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre challenge piece due to pointer dangling")
		return ErrDanglingPointer
	}
	if task.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to pre challenge piece due to object unsealed")
		return ErrObjectUnsealed
	}
	d.baseApp.GfSpClient().ReportTask(ctx, task)
	return nil
}

func (d *DownloadModular) HandleChallengePiece(
	ctx context.Context,
	task task.ChallengePieceTask) (
	[]byte, [][]byte, []byte, error) {
	var (
		err       error
		integrity *corespdb.IntegrityMeta
		data      []byte
	)
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()
	if err = d.challengeQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push challenge piece queue", "error", err)
		return nil, nil, nil, ErrExceedQueue
	}
	defer d.challengeQueue.PopByKey(task.Key())
	pieceKey := d.baseApp.PieceOp().ChallengePieceKey(
		task.GetObjectInfo().Id.Uint64(),
		task.GetSegmentIdx(),
		task.GetRedundancyIdx())
	integrity, err = d.baseApp.GfSpDB().GetObjectIntegrity(task.GetObjectInfo().Id.Uint64())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get integrity hash", "error", err)
		return nil, nil, nil, ErrGfSpDB
	}
	if int(task.GetSegmentIdx()) >= len(integrity.PieceChecksumList) {
		log.CtxErrorw(ctx, "failed to get challenge info due to segment index wrong")
		return nil, nil, nil, ErrNoSuchPiece
	}
	data, err = d.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get piece data", "error", err)
		return nil, nil, nil, ErrPieceStore
	}
	return integrity.IntegrityChecksum, integrity.PieceChecksumList, data, nil
}

func (d *DownloadModular) PostChallengePiece(
	ctx context.Context,
	task task.ChallengePieceTask) {
}

func (d *DownloadModular) QueryTasks(
	ctx context.Context,
	subKey task.TKey) (
	[]task.Task, error) {
	downloadTasks, _ := taskqueue.ScanTQueueBySubKey(d.downloadQueue, subKey)
	challengeTasks, _ := taskqueue.ScanTQueueBySubKey(d.challengeQueue, subKey)
	return append(downloadTasks, challengeTasks...), nil
}
