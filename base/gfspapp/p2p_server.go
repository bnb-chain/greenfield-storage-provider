package gfspapp

import (
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	ErrReplicatePieceApprovalTaskDangling = gfsperrors.Register(BaseCodeSpace, http.StatusInternalServerError, 990701, "OoooH... request lost")
)
var _ gfspserver.GfSpP2PServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpAskSecondaryReplicatePieceApproval(
	ctx context.Context,
	req *gfspserver.GfSpAskSecondaryReplicatePieceApprovalRequest) (
	*gfspserver.GfSpAskSecondaryReplicatePieceApprovalResponse,
	error) {
	task := req.GetReplicatePieceApprovalTask()
	if task == nil {
		log.CtxError(ctx, "failed to ask replicate piece approval, task pointer dangling")
		return nil, ErrReplicatePieceApprovalTaskDangling
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	approvals, err := g.p2p.HandleReplicatePieceApproval(ctx, task, req.GetMin(), req.GetMax(), req.GetTimeout())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get replicate piece approval from p2p", "error", err)
		return &gfspserver.GfSpAskSecondaryReplicatePieceApprovalResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	}
	resp := &gfspserver.GfSpAskSecondaryReplicatePieceApprovalResponse{}
	for _, approval := range approvals {
		resp.ApprovedTasks = append(resp.ApprovedTasks, approval.(*gfsptask.GfSpReplicatePieceApprovalTask))
	}
	return resp, nil
}

func (g *GfSpBaseApp) GfSpQueryP2PBootstrap(ctx context.Context,
	req *gfspserver.GfSpQueryP2PNodeRequest) (
	*gfspserver.GfSpQueryP2PNodeResponse, error) {
	nodes, err := g.p2p.HandleQueryBootstrap(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to query p2p bootstrap", "error", err)
		return &gfspserver.GfSpQueryP2PNodeResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	}
	return &gfspserver.GfSpQueryP2PNodeResponse{Nodes: nodes}, nil
}
