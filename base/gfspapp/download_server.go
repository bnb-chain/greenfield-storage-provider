package gfspapp

import (
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrDownloadTaskDangling    = gfsperrors.Register(BaseCodeSpace, http.StatusInternalServerError, 990301, "OoooH... request lost")
	ErrDownloadExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusServiceUnavailable, 990302, "server overload, try again later")
)

var _ gfspserver.GfSpDownloadServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpDownloadObject(
	ctx context.Context,
	req *gfspserver.GfSpDownloadObjectRequest) (
	*gfspserver.GfSpDownloadObjectResponse, error) {
	task := req.GetDownLoadTask()
	if task == nil {
		log.Error("failed to download object, download task pointer dangling")
		return &gfspserver.GfSpDownloadObjectResponse{Err: ErrDownloadTaskDangling}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	span, err := g.downloader.ReserveResource(ctx, task.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve download resource", "error", err)
		return &gfspserver.GfSpDownloadObjectResponse{Err: ErrDownloadExhaustResource}, nil
	}
	defer span.Done()
	metrics.DownloadObjectSizeHistogram.WithLabelValues(
		g.downloader.Name()).Observe(float64(task.GetSize()))
	data, err := g.OnDownloadObjectTask(ctx, task)
	log.CtxDebugw(ctx, "finished to download object", "len", len(data), "error", err)
	return &gfspserver.GfSpDownloadObjectResponse{
		Err:  gfsperrors.MakeGfSpError(err),
		Data: data}, nil
}

func (g *GfSpBaseApp) OnDownloadObjectTask(
	ctx context.Context,
	task task.DownloadObjectTask) (
	[]byte, error) {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to download object, download task pointer dangling")
		return nil, ErrDownloadTaskDangling
	}
	err := g.downloader.PreDownloadObject(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre download object", "info", task.Info(), "error", err)
		return nil, err
	}
	data, err := g.downloader.HandleDownloadObjectTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to download object", "error", err)
		return nil, err
	}
	g.downloader.PostDownloadObject(ctx, task)
	log.CtxDebugw(ctx, "succeed to download object")
	return data, nil
}

func (g *GfSpBaseApp) GfSpGetChallengeInfo(
	ctx context.Context,
	req *gfspserver.GfSpGetChallengeInfoRequest) (
	*gfspserver.GfSpGetChallengeInfoResponse, error) {
	task := req.GetChallengePieceTask()
	if task == nil {
		log.CtxError(ctx, "failed to challenge piece, task pointer dangling")
		return &gfspserver.GfSpGetChallengeInfoResponse{Err: ErrDownloadTaskDangling}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	span, err := g.downloader.ReserveResource(ctx, task.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve challenge resource", "error", err)
		return &gfspserver.GfSpGetChallengeInfoResponse{Err: ErrDownloadExhaustResource}, nil
	}
	defer span.Done()
	metrics.ChallengePieceSizeHistogram.WithLabelValues(g.downloader.Name()).Observe(
		float64(task.EstimateLimit().GetMemoryLimit()))
	integrity, checksums, data, err := g.OnChallengePieceTask(ctx, task)
	log.CtxDebugw(ctx, "finished to get object challenge info", "len", len(data), "error", err)
	return &gfspserver.GfSpGetChallengeInfoResponse{
		Err:           gfsperrors.MakeGfSpError(err),
		IntegrityHash: integrity,
		Checksums:     checksums,
		Data:          data}, nil
}

func (g *GfSpBaseApp) OnChallengePieceTask(
	ctx context.Context,
	task task.ChallengePieceTask) (
	[]byte, [][]byte, []byte, error) {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to challenge piece, challenge piece task pointer dangling")
		return nil, nil, nil, ErrDownloadTaskDangling
	}
	err := g.downloader.PreChallengePiece(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre challenge piece", "info", task.Info(), "error", err)
		return nil, nil, nil, err
	}
	integrity, checksums, data, err := g.downloader.HandleChallengePiece(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to challenge piece", "error", err)
		return nil, nil, nil, err
	}
	g.downloader.PostChallengePiece(ctx, task)
	log.CtxDebugw(ctx, "succeed to challenge piece")
	return integrity, checksums, data, nil
}
