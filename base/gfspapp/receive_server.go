package gfspapp

import (
	"context"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var _ gfspserver.GfSpReceiveServiceServer = &GfSpBaseApp{}

var (
	ErrReceiveTaskDangling    = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 990801, "OoooH... request lost")
	ErrReceiveExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 990802, "server overload, try again later")
)

func (g *GfSpBaseApp) GfSpReplicatePiece(ctx context.Context, req *gfspserver.GfSpReplicatePieceRequest) (
	resp *gfspserver.GfSpReplicatePieceResponse, err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(ReceiverFailureReplicatePiece).Inc()
			metrics.ReqTime.WithLabelValues(ReceiverFailureReplicatePiece).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ReceiverSuccessReplicatePiece).Inc()
			metrics.ReqTime.WithLabelValues(ReceiverSuccessReplicatePiece).Observe(time.Since(startTime).Seconds())
		}
	}()

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
	err = g.receiver.HandleReceivePieceTask(ctx, task, req.GetPieceData())
	if err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece", "error", err)
		return &gfspserver.GfSpReplicatePieceResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	}
	log.CtxDebugw(ctx, "succeed to replicate piece")
	return &gfspserver.GfSpReplicatePieceResponse{}, nil
}

func (g *GfSpBaseApp) GfSpDoneReplicatePiece(ctx context.Context, req *gfspserver.GfSpDoneReplicatePieceRequest) (
	resp *gfspserver.GfSpDoneReplicatePieceResponse, err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(ReceiverFailureDoneReplicatePiece).Inc()
			metrics.ReqTime.WithLabelValues(ReceiverFailureDoneReplicatePiece).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ReceiverSuccessDoneReplicatePiece).Inc()
			metrics.ReqTime.WithLabelValues(ReceiverSuccessDoneReplicatePiece).Observe(time.Since(startTime).Seconds())
		}
	}()

	task := req.GetReceivePieceTask()
	if task == nil {
		log.Error("failed to done receive piece due to task pointer dangling")
		return &gfspserver.GfSpDoneReplicatePieceResponse{Err: ErrReceiveTaskDangling}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	signature, err := g.receiver.HandleDoneReceivePieceTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate piece", "error", err)
		return &gfspserver.GfSpDoneReplicatePieceResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	}
	log.CtxDebugw(ctx, "succeed to done replicate pieces")
	return &gfspserver.GfSpDoneReplicatePieceResponse{
		Signature: signature,
	}, nil
}
