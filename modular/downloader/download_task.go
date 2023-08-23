package downloader

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingPointer   = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30001, "OoooH.... request lost")
	ErrObjectUnsealed    = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30002, "object unsealed")
	ErrExceedRequest     = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30003, "get piece request exceed")
	ErrExceedBucketQuota = gfsperrors.Register(module.DownloadModularName, http.StatusNotAcceptable, 30004, "bucket quota overflow")
	ErrInvalidParam      = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30005, "request params invalid")
	ErrNoSuchPiece       = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30006, "request params invalid, no such piece")
	ErrKeyFormat         = gfsperrors.Register(module.DownloadModularName, http.StatusBadRequest, 30007, "invalid key format")
)

func ErrPieceStoreWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 85101, detail)
}

func ErrGfSpDBWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 85201, detail)
}

func ErrConsensusWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.ReceiveModularName, http.StatusInternalServerError, 85201, detail)
}

func (d *DownloadModular) PreDownloadObject(ctx context.Context, downloadObjectTask task.DownloadObjectTask) error {
	var (
		err           error
		freeQuotaSize uint64
		bucketTraffic *spdb.BucketTraffic
	)

	if downloadObjectTask == nil || downloadObjectTask.GetObjectInfo() == nil || downloadObjectTask.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed pre download object due to pointer dangling")
		return ErrDanglingPointer
	}
	if downloadObjectTask.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to pre download object due to object unsealed")
		return ErrObjectUnsealed
	}

	readRecord := &spdb.ReadRecord{
		BucketID:        downloadObjectTask.GetBucketInfo().Id.Uint64(),
		ObjectID:        downloadObjectTask.GetObjectInfo().Id.Uint64(),
		UserAddress:     downloadObjectTask.GetUserAddress(),
		BucketName:      downloadObjectTask.GetBucketInfo().GetBucketName(),
		ObjectName:      downloadObjectTask.GetObjectInfo().GetObjectName(),
		ReadSize:        uint64(downloadObjectTask.GetSize()),
		ReadTimestampUs: sqldb.GetCurrentTimestampUs(),
	}

	yearMonth := sqldb.TimestampYearMonth(readRecord.ReadTimestampUs)
	bucketID := downloadObjectTask.GetBucketInfo().Id.Uint64()
	bucketTraffic, err = d.baseApp.GfSpDB().GetBucketTraffic(bucketID, yearMonth)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.CtxErrorw(ctx, "failed to get bucket traffic", "bucket_id", bucketID, "error", err)
		return err
	}

	// init the bucket traffic table when checking quota for the first time
	if bucketTraffic == nil {
		freeQuotaSize, err = d.baseApp.Consensus().QuerySPFreeQuota(ctx, d.baseApp.OperatorAddress())
		if err != nil {
			return ErrConsensusWithDetail("QuerySPFreeQuota error: " + err.Error())
		}
		// only need to set the free quota when init the traffic table for every month
		err = d.baseApp.GfSpDB().InitBucketTraffic(readRecord, &spdb.BucketQuota{
			ChargedQuotaSize: downloadObjectTask.GetBucketInfo().GetChargedReadQuota(),
			FreeQuotaSize:    freeQuotaSize,
		})
		if err != nil {
			log.CtxErrorw(ctx, "failed to init bucket traffic", "error", err)
			return ErrGfSpDBWithDetail("failed to init bucket traffic, error: " + err.Error())
		}
	}

	// TODO:: spilt check and add record steps, check in pre download, add record in post download
	if err = d.baseApp.GfSpDB().CheckQuotaAndAddReadRecord(
		readRecord,
		&spdb.BucketQuota{
			ChargedQuotaSize: downloadObjectTask.GetBucketInfo().GetChargedReadQuota(),
			FreeQuotaSize:    freeQuotaSize,
		},
	); err != nil {
		log.CtxErrorw(ctx, "failed to check bucket quota", "error", err)
		if errors.Is(err, sqldb.ErrCheckQuotaEnough) {
			return ErrExceedBucketQuota
		}
		// ignore the access db error, it is the system's inner error, will be let the request go.
	}
	// report the task to the manager for monitor the download task
	go func() {
		_ = d.baseApp.GfSpClient().ReportTask(context.Background(), downloadObjectTask)
	}()
	return nil
}

func (d *DownloadModular) HandleDownloadObjectTask(ctx context.Context, downloadObjectTask task.DownloadObjectTask) ([]byte, error) {
	var err error
	defer func() {
		if err != nil {
			downloadObjectTask.SetError(err)
		}
		log.CtxDebugw(ctx, downloadObjectTask.Info())
	}()
	defer func() {
		atomic.AddInt64(&d.downloading, -1)
	}()
	if atomic.AddInt64(&d.downloading, 1) >= atomic.LoadInt64(&d.downloadParallel) {
		log.CtxErrorw(ctx, "failed to download object due to max download concurrent",
			"current_download_concurrent", d.downloading, "max_download_concurrent", d.downloadParallel,
			"task_info", downloadObjectTask.Info())
		err = ErrExceedRequest
		return nil, err
	}

	pieceInfos, err := SplitToSegmentPieceInfos(downloadObjectTask, d.baseApp.PieceOp())
	if err != nil {
		log.CtxErrorw(ctx, "failed to generate piece info to download", "error", err)
		return nil, err
	}
	var data []byte
	for _, pInfo := range pieceInfos {
		key := cacheKey(pInfo.SegmentPieceKey, int64(pInfo.Offset), int64(pInfo.Length))
		pieceData, has := d.pieceCache.Get(key)
		if has {
			data = append(data, pieceData.([]byte)...)
			continue
		}
		piece, getPieceErr := d.baseApp.PieceStore().GetPiece(ctx, pInfo.SegmentPieceKey,
			int64(pInfo.Offset), int64(pInfo.Length))
		if getPieceErr != nil {
			log.CtxErrorw(ctx, "failed to get piece data from piece store", "task_info", downloadObjectTask.Info(), "piece_info", pInfo, "error", getPieceErr)
			err = ErrPieceStoreWithDetail("failed to get piece data from piece store, error: " + getPieceErr.Error())
			return nil, err
		}
		d.pieceCache.Add(key, piece)
		data = append(data, piece...)
	}
	return data, nil
}

type SegmentPieceInfo struct {
	SegmentPieceKey string
	Offset          uint64
	Length          uint64
}

func SplitToSegmentPieceInfos(downloadObjectTask task.DownloadObjectTask, op piecestore.PieceOp) ([]*SegmentPieceInfo, error) {
	if downloadObjectTask.GetObjectInfo().GetPayloadSize() == 0 ||
		downloadObjectTask.GetLow() >= int64(downloadObjectTask.GetObjectInfo().GetPayloadSize()) ||
		downloadObjectTask.GetHigh() >= int64(downloadObjectTask.GetObjectInfo().GetPayloadSize()) ||
		downloadObjectTask.GetHigh() < downloadObjectTask.GetLow() {
		log.Errorw("failed to parse params", "object_size",
			downloadObjectTask.GetObjectInfo().GetPayloadSize(), "low", downloadObjectTask.GetLow(), "high", downloadObjectTask.GetHigh())
		return nil, ErrInvalidParam
	}
	segmentSize := downloadObjectTask.GetStorageParams().VersionedParams.GetMaxSegmentSize()
	segmentCount := op.SegmentPieceCount(downloadObjectTask.GetObjectInfo().GetPayloadSize(),
		downloadObjectTask.GetStorageParams().VersionedParams.GetMaxSegmentSize())
	var (
		pieceInfos []*SegmentPieceInfo
		low        = uint64(downloadObjectTask.GetLow())
		high       = uint64(downloadObjectTask.GetHigh())
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
					downloadObjectTask.GetObjectInfo().Id.Uint64(),
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
					downloadObjectTask.GetObjectInfo().Id.Uint64(),
					uint32(segmentPieceIndex)),
				Offset: offsetInPiece,
				Length: lengthInPiece,
			})
		}
	}
	return pieceInfos, nil
}

func (d *DownloadModular) PostDownloadObject(context.Context, task.DownloadObjectTask) {
}

func (d *DownloadModular) PreDownloadPiece(ctx context.Context, downloadPieceTask task.DownloadPieceTask) error {
	defer func() {
		// report the task to the manager for monitor the download piece task
		go func() {
			_ = d.baseApp.GfSpClient().ReportTask(context.Background(), downloadPieceTask)
		}()
	}()
	var (
		err           error
		freeQuotaSize uint64
		bucketTraffic *spdb.BucketTraffic
	)

	if downloadPieceTask == nil || downloadPieceTask.GetObjectInfo() == nil || downloadPieceTask.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed pre download piece due to pointer dangling")
		return ErrDanglingPointer
	}

	if downloadPieceTask.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to pre download piece due to object unsealed")
		return ErrObjectUnsealed
	}

	// if it is a request from the primary SP of the object, no need to check quota
	bucketInfo := downloadPieceTask.GetBucketInfo()
	bucketSPID, err := util.GetBucketPrimarySPID(ctx, d.baseApp.Consensus(), bucketInfo)
	if err != nil {
		return err
	}
	bucketPrimarySp, err := d.baseApp.Consensus().QuerySPByID(ctx, bucketSPID)
	if err != nil {
		return err
	}
	if downloadPieceTask.GetUserAddress() == bucketPrimarySp.OperatorAddress {
		return nil
	}
	if downloadPieceTask.GetEnableCheck() {
		checkQuotaTime := time.Now()

		bucketID := downloadPieceTask.GetBucketInfo().Id.Uint64()
		bucketName := downloadPieceTask.GetBucketInfo().GetBucketName()
		readRecord := &spdb.ReadRecord{
			BucketID:        bucketID,
			ObjectID:        downloadPieceTask.GetObjectInfo().Id.Uint64(),
			UserAddress:     downloadPieceTask.GetUserAddress(),
			BucketName:      bucketName,
			ObjectName:      downloadPieceTask.GetObjectInfo().GetObjectName(),
			ReadSize:        downloadPieceTask.GetTotalSize(),
			ReadTimestampUs: sqldb.GetCurrentTimestampUs(),
		}

		yearMonth := sqldb.TimestampYearMonth(readRecord.ReadTimestampUs)
		bucketTraffic, err = d.baseApp.GfSpDB().GetBucketTraffic(bucketID, yearMonth)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.CtxErrorw(ctx, "failed to get bucket traffic", "task_info", downloadPieceTask.Info(), "error", err)
			return err
		}

		// init the bucket traffic table when checking quota for the first time
		if bucketTraffic == nil {
			freeQuotaSize, err = d.baseApp.Consensus().QuerySPFreeQuota(ctx, d.baseApp.OperatorAddress())
			if err != nil {
				return ErrConsensusWithDetail("QuerySPFreeQuota error: " + err.Error())
			}
			log.CtxDebugw(ctx, "finish init bucket traffic table", "charged_quota", downloadPieceTask.GetBucketInfo().GetChargedReadQuota(),
				"free_quota", freeQuotaSize)

			// only need to set the free quota when init the traffic table for every month
			err = d.baseApp.GfSpDB().InitBucketTraffic(readRecord, &spdb.BucketQuota{
				ChargedQuotaSize: downloadPieceTask.GetBucketInfo().GetChargedReadQuota(),
				FreeQuotaSize:    freeQuotaSize,
			})
			if err != nil {
				log.CtxErrorw(ctx, "failed to init bucket traffic", "error", err)
				return ErrGfSpDBWithDetail("failed to init bucket traffic, error: " + err.Error())
			}
		}

		if dbErr := d.baseApp.GfSpDB().CheckQuotaAndAddReadRecord(
			readRecord,
			&spdb.BucketQuota{
				ChargedQuotaSize: downloadPieceTask.GetBucketInfo().GetChargedReadQuota(),
			},
		); dbErr != nil {
			metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_check_quota_time").Observe(time.Since(checkQuotaTime).Seconds())
			log.CtxErrorw(ctx, "failed to check bucket quota", "error", dbErr)
			if errors.Is(dbErr, sqldb.ErrCheckQuotaEnough) {
				return ErrExceedBucketQuota
			}
			// ignore the access db error, it is the system's inner error, will be let the request go.
		}
		metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_check_quota_time").Observe(time.Since(checkQuotaTime).Seconds())

	}

	return nil
}

func (d *DownloadModular) HandleDownloadPieceTask(ctx context.Context, downloadPieceTask task.DownloadPieceTask) ([]byte, error) {
	var (
		err       error
		pieceData []byte
	)

	defer func() {
		if err != nil {
			downloadPieceTask.SetError(err)
		}
		log.CtxDebugw(ctx, downloadPieceTask.Info())
	}()

	defer func() {
		atomic.AddInt64(&d.downloading, -1)
	}()
	if atomic.AddInt64(&d.downloading, 1) >= atomic.LoadInt64(&d.downloadParallel) {
		log.CtxErrorw(ctx, "failed to download object due to max download concurrent",
			"current_download_concurrent", d.downloading, "max_download_concurrent", d.downloadParallel,
			"task_info", downloadPieceTask.Info())
		err = ErrExceedRequest
		return nil, err
	}

	key := cacheKey(downloadPieceTask.GetPieceKey(),
		int64(downloadPieceTask.GetPieceOffset()),
		int64(downloadPieceTask.GetPieceLength()))
	data, has := d.pieceCache.Get(key)
	if has {
		return data.([]byte), nil
	}

	putPieceTime := time.Now()
	if pieceData, err = d.baseApp.PieceStore().GetPiece(ctx, downloadPieceTask.GetPieceKey(),
		int64(downloadPieceTask.GetPieceOffset()), int64(downloadPieceTask.GetPieceLength())); err != nil {
		metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_put_piece_time").Observe(time.Since(putPieceTime).Seconds())
		log.CtxErrorw(ctx, "failed to get piece data from piece store", "task_info", downloadPieceTask.Info(), "error", err)
		return nil, ErrPieceStoreWithDetail("failed to get piece data from piece store, task_info: " + downloadPieceTask.Info() + ", error: " + err.Error())
	}
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_put_piece_time").Observe(time.Since(putPieceTime).Seconds())
	return pieceData, nil
}

func (d *DownloadModular) PostDownloadPiece(context.Context, task.DownloadPieceTask) {
}

func (d *DownloadModular) PreChallengePiece(ctx context.Context, challengePieceTask task.ChallengePieceTask) error {
	if challengePieceTask == nil || challengePieceTask.GetObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre challenge piece due to pointer dangling")
		return ErrDanglingPointer
	}
	if challengePieceTask.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.CtxErrorw(ctx, "failed to pre challenge piece due to object unsealed")
		return ErrObjectUnsealed
	}
	go func() {
		_ = d.baseApp.GfSpClient().ReportTask(context.Background(), challengePieceTask)
	}()
	return nil
}

func (d *DownloadModular) HandleChallengePiece(ctx context.Context, challengePieceTask task.ChallengePieceTask) (
	[]byte, [][]byte, []byte, error) {
	var (
		err       error
		integrity *corespdb.IntegrityMeta
		data      []byte
	)
	defer func() {
		if err != nil {
			challengePieceTask.SetError(err)
		}
		log.CtxDebugw(ctx, challengePieceTask.Info())
	}()

	defer func() {
		atomic.AddInt64(&d.challenging, -1)
	}()
	if atomic.AddInt64(&d.challenging, 1) >= atomic.LoadInt64(&d.challengeParallel) {
		log.CtxErrorw(ctx, "failed to get challenge piece info due to max challenge concurrent",
			"current_challenge_concurrent", d.challenging, "max_challenge_concurrent", d.challengeParallel,
			"task_info", challengePieceTask.Info())
		err = ErrExceedRequest
		return nil, nil, nil, err
	}

	pieceKey := d.baseApp.PieceOp().ChallengePieceKey(
		challengePieceTask.GetObjectInfo().Id.Uint64(),
		challengePieceTask.GetSegmentIdx(),
		challengePieceTask.GetRedundancyIdx())
	getIntegrityTime := time.Now()
	integrity, err = d.baseApp.GfSpDB().GetObjectIntegrity(challengePieceTask.GetObjectInfo().Id.Uint64(), challengePieceTask.GetRedundancyIdx())
	metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_get_integrity_time").Observe(time.Since(getIntegrityTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get integrity hash", "task", challengePieceTask, "error", err)
		return nil, nil, nil, ErrGfSpDBWithDetail("failed to get integrity hash, error: " + err.Error())
	}
	if int(challengePieceTask.GetSegmentIdx()) >= len(integrity.PieceChecksumList) {
		log.CtxErrorw(ctx, "failed to get challenge info due to segment index wrong", "task", challengePieceTask, "integrity_meta", integrity)
		return nil, nil, nil, ErrNoSuchPiece
	}

	key := cacheKey(pieceKey, int64(0), int64(-1))
	piece, has := d.pieceCache.Get(key)
	if has {
		return integrity.IntegrityChecksum, integrity.PieceChecksumList, piece.([]byte), nil
	}

	getPieceTime := time.Now()
	data, err = d.baseApp.PieceStore().GetPiece(ctx, pieceKey, 0, -1)
	metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_get_piece_time").Observe(time.Since(getPieceTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get piece data", "task", challengePieceTask, "error", err)
		return nil, nil, nil, ErrPieceStoreWithDetail("failed to get piece data, error: " + err.Error())
	}

	return integrity.IntegrityChecksum, integrity.PieceChecksumList, data, nil
}

func (d *DownloadModular) PostChallengePiece(context.Context, task.ChallengePieceTask) {
}

func (d *DownloadModular) QueryTasks(context.Context, task.TKey) ([]task.Task, error) {
	return nil, nil
}
