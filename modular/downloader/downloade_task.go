package downloader

import (
	"context"
	"net/http"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

var (
	ErrDanglingPointer   = gfsperrors.Register(DownloadModularName, http.StatusNotFound, 30001, "OoooH.... request lost")
	ErrObjectState       = gfsperrors.Register(DownloadModularName, http.StatusBadRequest, 30002, "object unsealed")
	ErrRepeatedTask      = gfsperrors.Register(DownloadModularName, http.StatusBadRequest, 30004, "request repeated")
	ErrExceedBucketQuota = gfsperrors.Register(DownloadModularName, http.StatusBadRequest, 30005, "bucket quota overflow")
	ErrExceedQueue       = gfsperrors.Register(DownloadModularName, http.StatusServiceUnavailable, 30006, "request exceed the limit, try again later")
	ErrInvalidParam      = gfsperrors.Register(DownloadModularName, http.StatusBadRequest, 30007, "request params invalid")
	ErrPieceStore        = gfsperrors.Register(DownloadModularName, http.StatusInternalServerError, 35101, "server slipped away, try again later")
	ErrGfSpDB            = gfsperrors.Register(DownloadModularName, http.StatusInternalServerError, 35201, "server slipped away, try again later")
)

func (d *DownloadModular) PreDownloadObject(
	ctx context.Context,
	task task.DownloadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed pre download object, task pointer dangling")
		return ErrDanglingPointer
	}
	if task.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to pre download object, object unsealed")
		return ErrObjectState
	}
	if d.downloadQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "failed to pre download object, task repeated")
		return ErrRepeatedTask
	}
	// TODO:: spilt check and add record, check in pre download, add record in post download
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
		return ErrExceedBucketQuota
	}
	d.baseApp.GfSpClient().ReportTask(ctx, task)
	return nil
}

func (d *DownloadModular) HandleDownloadObjectTask(
	ctx context.Context,
	task task.DownloadObjectTask) ([]byte, error) {
	if err := d.downloadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push download queue", "error", err)
		return nil, ErrExceedQueue
	}
	defer d.downloadQueue.PopByKey(task.Key())
	pieceInfos, err := d.SplitToSegmentPieceInfos(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to generate piece info to download", "error", err)
		return nil, err
	}
	var data []byte
	for _, pInfo := range pieceInfos {
		piece, err := d.baseApp.PieceStore().GetPiece(ctx, pInfo.segmentPieceKey,
			int64(pInfo.offset), int64(pInfo.length))
		if err != nil {
			return nil, ErrPieceStore
		}
		data = append(data, piece...)
	}
	return data, nil
}

type segmentPieceInfo struct {
	segmentPieceKey string
	offset          uint64
	length          uint64
}

func (d *DownloadModular) SplitToSegmentPieceInfos(
	ctx context.Context,
	task task.DownloadObjectTask) (
	[]*segmentPieceInfo, error) {
	if task.GetObjectInfo().GetPayloadSize() == 0 ||
		task.GetLow() >= int64(task.GetObjectInfo().GetPayloadSize()) ||
		task.GetHigh() >= int64(task.GetObjectInfo().GetPayloadSize()) ||
		task.GetHigh() <= task.GetLow() {
		log.CtxErrorw(ctx, "failed to parser params", "object_size",
			task.GetObjectInfo().GetPayloadSize(), "low", task.GetLow(), "high", task.GetHigh())
		return nil, ErrInvalidParam
	}
	segmentSize := task.GetStorageParams().VersionedParams.GetMaxSegmentSize()
	segmentCount := d.baseApp.PieceOp().MaxSegmentSize(task.GetObjectInfo().GetPayloadSize(),
		task.GetStorageParams().VersionedParams.GetMaxSegmentSize())
	var (
		pieceInfos []*segmentPieceInfo
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
			pieceInfos = append(pieceInfos, &segmentPieceInfo{
				segmentPieceKey: d.baseApp.PieceOp().SegmentPieceKey(
					task.GetObjectInfo().Id.Uint64(),
					uint32(segmentPieceIndex)),
				offset: offsetInPiece,
				length: lengthInPiece,
			})
			// break to finish
			break
		} else {
			offsetInPiece := currentStart - (segmentPieceIndex * segmentSize)
			lengthInPiece := currentEnd - currentStart + 1
			pieceInfos = append(pieceInfos, &segmentPieceInfo{
				segmentPieceKey: d.baseApp.PieceOp().SegmentPieceKey(
					task.GetObjectInfo().Id.Uint64(),
					uint32(segmentPieceIndex)),
				offset: offsetInPiece,
				length: lengthInPiece,
			})
		}
	}
	return pieceInfos, nil
}

func (d *DownloadModular) PostDownloadObject(
	ctx context.Context,
	task task.DownloadObjectTask) {
	d.baseApp.GfSpClient().ReportTask(ctx, task)
}

func (d *DownloadModular) PreChallengePiece(
	ctx context.Context,
	task task.ChallengePieceTask) error {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre challenge piece, task or object pointer dangling")
		return ErrDanglingPointer
	}
	if task.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to pre download object, object unsealed")
		return ErrObjectState
	}
	d.baseApp.GfSpClient().ReportTask(ctx, task)
	return nil
}

func (d *DownloadModular) HandleChallengePiece(
	ctx context.Context,
	task task.ChallengePieceTask) (
	[]byte, [][]byte, []byte, error) {
	d.challengeQueue.Push(task)
	if err := d.challengeQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push challenge piece queue", "error", err)
		return nil, nil, nil, ErrExceedQueue
	}
	defer d.challengeQueue.PopByKey(task.Key())
	pieceKey := d.baseApp.PieceOp().ChallengePieceKey(
		task.GetObjectInfo().Id.Uint64(),
		task.GetSegmentIdx(),
		task.GetRedundancyIdx())
	integrity, err := d.baseApp.GfSpDB().GetObjectIntegrity(task.GetObjectInfo().Id.Uint64())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get integrity hash", "error", err)
		return nil, nil, nil, ErrGfSpDB
	}
	data, err := d.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get piece data", "error", err)
		return nil, nil, nil, ErrPieceStore
	}
	return integrity.IntegrityHash, integrity.Checksum, data, nil
}

func (d *DownloadModular) PostChallengePiece(
	ctx context.Context,
	task task.ChallengePieceTask) {
	d.baseApp.GfSpClient().ReportTask(ctx, task)
}
