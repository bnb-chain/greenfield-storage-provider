package gfspapp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

var (
	ErrDownloadTaskDangling    = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 990301, "OoooH... request lost")
	ErrDownloadExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 990302, "server overload, try again later")
)

var _ gfspserver.GfSpDownloadServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpDownloadObject(ctx context.Context, req *gfspserver.GfSpDownloadObjectRequest) (
	*gfspserver.GfSpDownloadObjectResponse, error) {
	downloadObjectTask := req.GetDownloadObjectTask()
	if downloadObjectTask == nil {
		log.Error("failed to download object due to task pointer dangling")
		return &gfspserver.GfSpDownloadObjectResponse{Err: ErrDownloadTaskDangling}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, downloadObjectTask.Key().String())
	span, err := g.downloader.ReserveResource(ctx, downloadObjectTask.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve download object resource", "error", err)
		return &gfspserver.GfSpDownloadObjectResponse{Err: ErrDownloadExhaustResource}, nil
	}
	defer span.Done()
	data, err := g.OnDownloadObjectTask(ctx, downloadObjectTask)
	log.CtxDebugw(ctx, "finished to download object", "len", len(data), "error", err)
	return &gfspserver.GfSpDownloadObjectResponse{
		Err:  gfsperrors.MakeGfSpError(err),
		Data: data,
	}, nil
}

func (g *GfSpBaseApp) OnDownloadObjectTask(ctx context.Context, downloadObjectTask task.DownloadObjectTask) ([]byte, error) {
	if downloadObjectTask == nil || downloadObjectTask.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to download object due to task pointer dangling")
		return nil, ErrDownloadTaskDangling
	}
	err := g.downloader.PreDownloadObject(ctx, downloadObjectTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre download object", "task_info", downloadObjectTask.Info(), "error", err)
		return nil, err
	}
	data, err := g.downloader.HandleDownloadObjectTask(ctx, downloadObjectTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to download object", "error", err)
		return nil, err
	}
	g.downloader.PostDownloadObject(ctx, downloadObjectTask)
	log.CtxDebugw(ctx, "succeed to download object")
	return data, nil
}

func (g *GfSpBaseApp) GfSpDownloadPiece(ctx context.Context, req *gfspserver.GfSpDownloadPieceRequest) (
	*gfspserver.GfSpDownloadPieceResponse, error) {
	downloadPieceTask := req.GetDownloadPieceTask()
	startTime := time.Now()
	defer func() {
		metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_server_total_time").Observe(time.Since(startTime).Seconds())
	}()
	if downloadPieceTask == nil {
		log.Error("failed to download piece due to task pointer dangling")
		return &gfspserver.GfSpDownloadPieceResponse{Err: ErrDownloadTaskDangling}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, downloadPieceTask.Key().String())
	span, err := g.downloader.ReserveResource(ctx, downloadPieceTask.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve download piece resource", "error", err)
		return &gfspserver.GfSpDownloadPieceResponse{Err: ErrDownloadExhaustResource}, nil
	}
	defer span.Done()
	data, err := g.OnDownloadPieceTask(ctx, downloadPieceTask)
	log.CtxDebugw(ctx, "finished to download piece", "len", len(data), "error", err)
	return &gfspserver.GfSpDownloadPieceResponse{
		Err:  gfsperrors.MakeGfSpError(err),
		Data: data,
	}, nil
}

func (g *GfSpBaseApp) OnDownloadPieceTask(ctx context.Context, downloadPieceTask task.DownloadPieceTask) (
	data []byte, err error) {
	if downloadPieceTask == nil || downloadPieceTask.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to download piece due to task pointer dangling")
		return nil, ErrDownloadTaskDangling
	}
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(DownloaderFailureGetPiece).Inc()
			metrics.ReqTime.WithLabelValues(DownloaderFailureGetPiece).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(DownloaderSuccessGetPiece).Inc()
			metrics.ReqTime.WithLabelValues(DownloaderSuccessGetPiece).Observe(time.Since(startTime).Seconds())
		}
	}()

	err = g.downloader.PreDownloadPiece(ctx, downloadPieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre download piece", "task_info", downloadPieceTask.Info(), "error", err)
		return nil, err
	}
	data, err = g.downloader.HandleDownloadPieceTask(ctx, downloadPieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to download piece", "error", err)
		return nil, err
	}
	g.downloader.PostDownloadPiece(ctx, downloadPieceTask)
	log.CtxDebugw(ctx, "succeed to download piece")
	return data, nil
}

func (g *GfSpBaseApp) GfSpGetChallengeInfo(ctx context.Context, req *gfspserver.GfSpGetChallengeInfoRequest) (
	*gfspserver.GfSpGetChallengeInfoResponse, error) {
	startTime := time.Now()
	defer func() {
		metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_server_total_time").Observe(time.Since(startTime).Seconds())
	}()
	challengePieceTask := req.GetChallengePieceTask()
	if challengePieceTask == nil {
		log.CtxError(ctx, "failed to challenge piece due to task pointer dangling")
		return &gfspserver.GfSpGetChallengeInfoResponse{Err: ErrDownloadTaskDangling}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, challengePieceTask.Key().String())
	span, err := g.downloader.ReserveResource(ctx, challengePieceTask.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve challenge resource", "error", err)
		return &gfspserver.GfSpGetChallengeInfoResponse{Err: ErrDownloadExhaustResource}, nil
	}
	defer span.Done()
	integrity, checksums, data, err := g.OnChallengePieceTask(ctx, challengePieceTask)
	log.CtxDebugw(ctx, "finished to get object challenge info", "len", len(data), "error", err)
	return &gfspserver.GfSpGetChallengeInfoResponse{
		Err:           gfsperrors.MakeGfSpError(err),
		IntegrityHash: integrity,
		Checksums:     checksums,
		Data:          data,
	}, nil
}

func (g *GfSpBaseApp) OnChallengePieceTask(ctx context.Context, challengePieceTask task.ChallengePieceTask) (
	integrity []byte, checksums [][]byte, data []byte, err error) {
	if challengePieceTask == nil || challengePieceTask.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to challenge piece due to task pointer dangling")
		return nil, nil, nil, ErrDownloadTaskDangling
	}

	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(DownloaderFailureGetChallengeInfo).Inc()
			metrics.ReqTime.WithLabelValues(DownloaderFailureGetChallengeInfo).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(DownloaderSuccessGetChallengeInfo).Inc()
			metrics.ReqTime.WithLabelValues(DownloaderSuccessGetChallengeInfo).Observe(time.Since(startTime).Seconds())
		}
	}()

	err = g.downloader.PreChallengePiece(ctx, challengePieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre challenge piece", "task_info", challengePieceTask.Info(), "error", err)
		return nil, nil, nil, err
	}
	integrity, checksums, data, err = g.downloader.HandleChallengePiece(ctx, challengePieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to challenge piece", "error", err)
		return nil, nil, nil, err
	}
	g.downloader.PostChallengePiece(ctx, challengePieceTask)
	log.CtxDebugw(ctx, "succeed to challenge piece")
	return integrity, checksums, data, nil
}

func (g *GfSpBaseApp) GfSpReimburseQuota(ctx context.Context, fixRequest *gfspserver.GfSpReimburseQuotaRequest) (*gfspserver.GfSpReimburseQuotaResponse, error) {
	if fixRequest == nil {
		log.CtxError(ctx, "failed to reimburse quota due to task pointer dangling")
		return &gfspserver.GfSpReimburseQuotaResponse{Err: ErrDownloadTaskDangling}, nil
	}
	err := g.GfSpDB().UpdateExtraQuota(fixRequest.GetBucketId(), fixRequest.GetExtraQuota(), fixRequest.YearMonth)
	if err != nil {
		log.CtxErrorw(ctx, "failed to reimburse extra quota", "error", err, "bucketID:", fixRequest.GetBucketId())
	}

	log.CtxDebugw(ctx, "succeed to reimburse extra quota", "bucketID:", fixRequest.GetBucketId(), "extra quota:", fixRequest.ExtraQuota)
	return &gfspserver.GfSpReimburseQuotaResponse{
		Err: gfsperrors.MakeGfSpError(err),
	}, nil
}

func (g *GfSpBaseApp) GfSpDeductQuotaForBucketMigrate(ctx context.Context, deductQuotaRequest *gfspserver.GfSpDeductQuotaForBucketMigrateRequest) (*gfspserver.GfSpDeductQuotaForBucketMigrateResponse, error) {
	// ReadRecord for bucket migration
	readRecord := &spdb.ReadRecord{
		BucketID:        deductQuotaRequest.GetBucketId(),
		ReadSize:        deductQuotaRequest.GetDeductQuota(),
		ReadTimestampUs: sqldb.GetCurrentTimestampUs(),
	}

	if dbErr := g.GfSpDB().CheckQuotaAndAddReadRecord(
		readRecord,
		&spdb.BucketQuota{
			ChargedQuotaSize: deductQuotaRequest.GetDeductQuota(),
		},
	); dbErr != nil {
		log.CtxErrorw(ctx, "failed to check bucket quota", "error", dbErr)
		if errors.Is(dbErr, sqldb.ErrCheckQuotaEnough) {
			return &gfspserver.GfSpDeductQuotaForBucketMigrateResponse{
				Err: gfsperrors.MakeGfSpError(dbErr),
			}, nil
		}
		// ignore the access db error, it is the system's inner error, will be let the request go.
	}
	log.CtxDebugw(ctx, "succeed to deduct quota for bucket migrate", "bucket_id", deductQuotaRequest.GetBucketId(), "deduct_quota", deductQuotaRequest.GetDeductQuota())
	return &gfspserver.GfSpDeductQuotaForBucketMigrateResponse{
		Err: gfsperrors.MakeGfSpError(nil),
	}, nil
}
