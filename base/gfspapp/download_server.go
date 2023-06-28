package gfspapp

import (
	"context"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
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
	metrics.DownloadObjectSizeHistogram.WithLabelValues(
		g.downloader.Name()).Observe(float64(downloadObjectTask.GetSize()))
	data, err := g.OnDownloadObjectTask(ctx, downloadObjectTask)
	log.CtxDebugw(ctx, "finished to download object", "len", len(data), "error", err)
	return &gfspserver.GfSpDownloadObjectResponse{
		Err:  gfsperrors.MakeGfSpError(err),
		Data: data}, nil
}

func (g *GfSpBaseApp) OnDownloadObjectTask(ctx context.Context, downloadObjectTask task.DownloadObjectTask) (
	[]byte, error) {
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
	defer metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_server_total_time").Observe(time.Since(startTime).Seconds())
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
	metrics.DownloadPieceSizeHistogram.WithLabelValues(
		g.downloader.Name()).Observe(float64(downloadPieceTask.GetSize()))
	data, err := g.OnDownloadPieceTask(ctx, downloadPieceTask)
	log.CtxDebugw(ctx, "finished to download piece", "len", len(data), "error", err)
	return &gfspserver.GfSpDownloadPieceResponse{
		Err:  gfsperrors.MakeGfSpError(err),
		Data: data}, nil
}

func (g *GfSpBaseApp) OnDownloadPieceTask(ctx context.Context, downloadPieceTask task.DownloadPieceTask) (
	[]byte, error) {
	if downloadPieceTask == nil || downloadPieceTask.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to download piece due to task pointer dangling")
		return nil, ErrDownloadTaskDangling
	}

	err := g.downloader.PreDownloadPiece(ctx, downloadPieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre download piece", "task_info", downloadPieceTask.Info(), "error", err)
		return nil, err
	}
	data, err := g.downloader.HandleDownloadPieceTask(ctx, downloadPieceTask)
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
	defer metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_server_total_time").Observe(time.Since(startTime).Seconds())
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
	metrics.ChallengePieceSizeHistogram.WithLabelValues(g.downloader.Name()).Observe(
		float64(challengePieceTask.EstimateLimit().GetMemoryLimit()))
	integrity, checksums, data, err := g.OnChallengePieceTask(ctx, challengePieceTask)
	log.CtxDebugw(ctx, "finished to get object challenge info", "len", len(data), "error", err)
	return &gfspserver.GfSpGetChallengeInfoResponse{
		Err:           gfsperrors.MakeGfSpError(err),
		IntegrityHash: integrity,
		Checksums:     checksums,
		Data:          data}, nil
}

func (g *GfSpBaseApp) OnChallengePieceTask(ctx context.Context, challengePieceTask task.ChallengePieceTask) (
	[]byte, [][]byte, []byte, error) {
	if challengePieceTask == nil || challengePieceTask.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to challenge piece due to task pointer dangling")
		return nil, nil, nil, ErrDownloadTaskDangling
	}
	err := g.downloader.PreChallengePiece(ctx, challengePieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre challenge piece", "task_info", challengePieceTask.Info(), "error", err)
		return nil, nil, nil, err
	}
	integrity, checksums, data, err := g.downloader.HandleChallengePiece(ctx, challengePieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to challenge piece", "error", err)
		return nil, nil, nil, err
	}
	g.downloader.PostChallengePiece(ctx, challengePieceTask)
	log.CtxDebugw(ctx, "succeed to challenge piece")
	return integrity, checksums, data, nil
}
