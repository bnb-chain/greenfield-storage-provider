package gfspapp

import (
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	ErrApprovalTaskDangling    = gfsperrors.Register(BaseCodeSpace, http.StatusInternalServerError, 990101, "OoooH... request lost")
	ErrApprovalExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusServiceUnavailable, 990102, "server overload, try again later")
)

var _ gfspserver.GfSpApprovalServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpAskApproval(
	ctx context.Context,
	req *gfspserver.GfSpAskApprovalRequest) (
	*gfspserver.GfSpAskApprovalResponse, error) {
	if req.GetRequest() == nil {
		log.Error("failed to ask approval, approval task pointer dangling")
		return &gfspserver.GfSpAskApprovalResponse{Err: ErrApprovalTaskDangling}, nil
	}
	switch task := req.GetRequest().(type) {
	case *gfspserver.GfSpAskApprovalRequest_CreateBucketApprovalTask:
		approvalTask := task.CreateBucketApprovalTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, approvalTask.Key().String())
		span, err := g.receiver.ReserveResource(ctx, approvalTask.EstimateLimit().ScopeStat())
		if err != nil {
			log.CtxErrorw(ctx, "failed to reserve approval resource", "error", err)
			return &gfspserver.GfSpAskApprovalResponse{Err: ErrApprovalExhaustResource}, nil
		}
		defer span.Done()
		allow, err := g.OnAskCreateBucketApproval(ctx, approvalTask)
		return &gfspserver.GfSpAskApprovalResponse{
			Err:     gfsperrors.MakeGfSpError(err),
			Allowed: allow,
			Response: &gfspserver.GfSpAskApprovalResponse_CreateBucketApprovalTask{
				CreateBucketApprovalTask: approvalTask,
			}}, nil
	case *gfspserver.GfSpAskApprovalRequest_CreateObjectApprovalTask:
		approvalTask := task.CreateObjectApprovalTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, approvalTask.Key().String())
		span, err := g.receiver.ReserveResource(ctx, approvalTask.EstimateLimit().ScopeStat())
		if err != nil {
			log.CtxErrorw(ctx, "failed to reserve approval resource", "error", err)
			return &gfspserver.GfSpAskApprovalResponse{Err: ErrApprovalExhaustResource}, nil
		}
		defer span.Done()
		allow, err := g.OnAskCreateObjectApproval(ctx, approvalTask)
		return &gfspserver.GfSpAskApprovalResponse{
			Err:     gfsperrors.MakeGfSpError(err),
			Allowed: allow,
			Response: &gfspserver.GfSpAskApprovalResponse_CreateObjectApprovalTask{
				CreateObjectApprovalTask: approvalTask,
			},
		}, nil
	default:
		return &gfspserver.GfSpAskApprovalResponse{Err: ErrUnsupportedTaskType}, nil
	}
}

func (g *GfSpBaseApp) OnAskCreateBucketApproval(
	ctx context.Context,
	task task.ApprovalCreateBucketTask) (
	bool, error) {
	if task == nil || task.GetCreateBucketInfo() == nil {
		log.CtxError(ctx, "failed to ask create bucket approval, bucket info pointer dangling")
		return false, ErrApprovalTaskDangling
	}

	err := g.approver.PreCreateBucketApproval(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre create bucket approval", "info", task.Info(), "error", err)
		return false, err
	}
	allow, err := g.approver.HandleCreateBucketApprovalTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to ask create bucket approval", "error", err)
		return false, err
	}
	g.approver.PostCreateBucketApproval(ctx, task)
	log.CtxDebugw(ctx, "succeed to ask create bucket approval")
	return allow, nil
}

func (g *GfSpBaseApp) OnAskCreateObjectApproval(
	ctx context.Context,
	task task.ApprovalCreateObjectTask) (
	bool, error) {
	if task == nil || task.GetCreateObjectInfo() == nil {
		log.CtxError(ctx, "failed to ask create object approval, object info pointer dangling")
		return false, ErrApprovalTaskDangling
	}

	err := g.approver.PreCreateObjectApproval(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre create object approval", "info", task.Info(), "error", err)
		return false, err
	}
	allow, err := g.approver.HandleCreateObjectApprovalTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to ask create object approval", "error", err)
		return false, err
	}
	g.approver.PostCreateObjectApproval(ctx, task)

	log.CtxDebugw(ctx, "succeed to ask create object approval")
	return allow, nil
}
