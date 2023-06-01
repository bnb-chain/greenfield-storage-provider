package gfspapp

import (
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var _ gfspserver.GfSpReceiveServiceServer = &GfSpBaseApp{}

var (
	ErrReceiveTaskDangling    = gfsperrors.Register(BaseCodeSpace, http.StatusInternalServerError, 990801, "OoooH... request lost")
	ErrReceiveExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusServiceUnavailable, 990802, "server overload, try again later")
)

func (g *GfSpBaseApp) GfSpReplicatePiece(ctx context.Context, req *gfspserver.GfSpReplicatePieceRequest) (
	*gfspserver.GfSpReplicatePieceResponse, error) {
	task := req.GetReceivePieceTask()
	if task == nil {
		log.Error("failed to receive piece due to task pointer dangling")
		return &gfspserver.GfSpReplicatePieceResponse{Err: ErrReceiveTaskDangling}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	span, err := g.receiver.ReserveResource(ctx, task.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve resource", "error", err)
		return &gfspserver.GfSpReplicatePieceResponse{Err: ErrReceiveExhaustResource}, nil
	}
	defer span.Done()
	metrics.ReceivePieceSizeHistogram.WithLabelValues(
		g.receiver.Name()).Observe(float64(task.GetPieceSize()))
	err = g.receiver.HandleReceivePieceTask(ctx, task, req.GetPieceData())
	if err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "error", err)
		return &gfspserver.GfSpReplicatePieceResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	}
	log.CtxDebugw(ctx, "succeed to replicate piece")
	return &gfspserver.GfSpReplicatePieceResponse{}, nil
}

func (g *GfSpBaseApp) GfSpDoneReplicatePiece(ctx context.Context, req *gfspserver.GfSpDoneReplicatePieceRequest) (
	*gfspserver.GfSpDoneReplicatePieceResponse, error) {
	task := req.GetReceivePieceTask()
	if task == nil {
		log.Error("failed to done receive piece due to task pointer dangling")
		return &gfspserver.GfSpDoneReplicatePieceResponse{Err: ErrReceiveTaskDangling}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	integrity, signature, err := g.receiver.HandleDoneReceivePieceTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece", "error", err)
		return &gfspserver.GfSpDoneReplicatePieceResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	}
	log.CtxDebugw(ctx, "succeed to done replicate pieces")
	return &gfspserver.GfSpDoneReplicatePieceResponse{
		IntegrityHash: integrity,
		Signature:     signature,
	}, nil
}
