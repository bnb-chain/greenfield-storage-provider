package gfspapp

import (
	"context"
	"net/http"
	"time"

	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsperrors"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfspserver"
	"github.com/zkMeLabs/mechain-storage-provider/core/task"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/metrics"
)

var (
	ErrApprovalTaskDangling    = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 990101, "OoooH... request lost")
	ErrApprovalExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 990102, "server overload, try again later")
)

var _ gfspserver.GfSpApprovalServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpAskApproval(ctx context.Context, req *gfspserver.GfSpAskApprovalRequest) (
	*gfspserver.GfSpAskApprovalResponse, error,
) {
	if req == nil || req.GetRequest() == nil {
		log.Error("failed to ask approval due to pointer dangling")
		return &gfspserver.GfSpAskApprovalResponse{Err: ErrApprovalTaskDangling}, nil
	}
	switch taskType := req.GetRequest().(type) {
	case *gfspserver.GfSpAskApprovalRequest_CreateBucketApprovalTask:
		approvalTask := taskType.CreateBucketApprovalTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, approvalTask.Key().String())
		span, err := g.approver.ReserveResource(ctx, approvalTask.EstimateLimit().ScopeStat())
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
			},
		}, nil
	case *gfspserver.GfSpAskApprovalRequest_MigrateBucketApprovalTask:
		approvalTask := taskType.MigrateBucketApprovalTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, approvalTask.Key().String())
		span, err := g.approver.ReserveResource(ctx, approvalTask.EstimateLimit().ScopeStat())
		if err != nil {
			log.CtxErrorw(ctx, "failed to reserve approval resource", "error", err)
			return &gfspserver.GfSpAskApprovalResponse{Err: ErrApprovalExhaustResource}, nil
		}
		defer span.Done()
		allow, err := g.OnAskMigrateBucketApproval(ctx, approvalTask)
		return &gfspserver.GfSpAskApprovalResponse{
			Err:     gfsperrors.MakeGfSpError(err),
			Allowed: allow,
			Response: &gfspserver.GfSpAskApprovalResponse_MigrateBucketApprovalTask{
				MigrateBucketApprovalTask: approvalTask,
			},
		}, nil
	case *gfspserver.GfSpAskApprovalRequest_CreateObjectApprovalTask:
		approvalTask := taskType.CreateObjectApprovalTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, approvalTask.Key().String())
		span, err := g.approver.ReserveResource(ctx, approvalTask.EstimateLimit().ScopeStat())
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
	case *gfspserver.GfSpAskApprovalRequest_DelegateCreateObjectApprovalTask:
		approvalTask := taskType.DelegateCreateObjectApprovalTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, approvalTask.Key().String())
		span, err := g.approver.ReserveResource(ctx, approvalTask.EstimateLimit().ScopeStat())
		if err != nil {
			log.CtxErrorw(ctx, "failed to reserve approval resource", "error", err)
			return &gfspserver.GfSpAskApprovalResponse{Err: ErrApprovalExhaustResource}, nil
		}
		defer span.Done()
		allow, err := g.OnAskDelegateCreateObjectApproval(ctx, approvalTask)
		return &gfspserver.GfSpAskApprovalResponse{
			Err:     gfsperrors.MakeGfSpError(err),
			Allowed: allow,
			Response: &gfspserver.GfSpAskApprovalResponse_DelegateCreateObjectApprovalTask{
				DelegateCreateObjectApprovalTask: approvalTask,
			},
		}, nil
	default:
		return &gfspserver.GfSpAskApprovalResponse{Err: ErrUnsupportedTaskType}, nil
	}
}

func (g *GfSpBaseApp) OnAskCreateBucketApproval(ctx context.Context, task task.ApprovalCreateBucketTask) (allow bool, err error) {
	if task == nil || task.GetCreateBucketInfo() == nil {
		log.CtxError(ctx, "failed to ask create bucket approval due to bucket info pointer dangling")
		return false, ErrApprovalTaskDangling
	}
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(ApproverFailureGetBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(ApproverFailureGetBucketApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ApproverSuccessGetBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(ApproverSuccessGetBucketApproval).Observe(time.Since(startTime).Seconds())
		}
	}()

	err = g.approver.PreCreateBucketApproval(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre create bucket approval", "info", task.Info(), "error", err)
		return false, err
	}
	allow, err = g.approver.HandleCreateBucketApprovalTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to ask create bucket approval", "error", err)
		return false, err
	}
	g.approver.PostCreateBucketApproval(ctx, task)
	log.CtxDebugw(ctx, "succeed to ask create bucket approval")
	return allow, nil
}

func (g *GfSpBaseApp) OnAskMigrateBucketApproval(ctx context.Context, task task.ApprovalMigrateBucketTask) (bool, error) {
	if task == nil || task.GetMigrateBucketInfo() == nil {
		log.CtxError(ctx, "failed to ask migrate bucket approval due to bucket info pointer dangling")
		return false, ErrApprovalTaskDangling
	}

	err := g.approver.PreMigrateBucketApproval(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre migrate bucket approval", "info", task.Info(), "error", err)
		return false, err
	}
	allow, err := g.approver.HandleMigrateBucketApprovalTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to ask  migrate bucket approval", "error", err)
		return false, err
	}
	g.approver.PostMigrateBucketApproval(ctx, task)
	log.CtxDebugw(ctx, "succeed to ask migrate bucket approval")
	return allow, nil
}

func (g *GfSpBaseApp) OnAskCreateObjectApproval(ctx context.Context, task task.ApprovalCreateObjectTask) (allow bool, err error) {
	if task == nil || task.GetCreateObjectInfo() == nil {
		log.CtxError(ctx, "failed to ask create object approval due to object info pointer dangling")
		return false, ErrApprovalTaskDangling
	}
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(ApproverFailureGetObjectApproval).Inc()
			metrics.ReqTime.WithLabelValues(ApproverFailureGetObjectApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ApproverSuccessGetObjectApproval).Inc()
			metrics.ReqTime.WithLabelValues(ApproverSuccessGetObjectApproval).Observe(time.Since(startTime).Seconds())
		}
	}()

	err = g.approver.PreCreateObjectApproval(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre create object approval", "info", task.Info(), "error", err)
		return false, err
	}
	allow, err = g.approver.HandleCreateObjectApprovalTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to ask create object approval", "error", err)
		return false, err
	}
	g.approver.PostCreateObjectApproval(ctx, task)
	log.CtxDebugw(ctx, "succeed to ask create object approval")
	return allow, nil
}

func (g *GfSpBaseApp) OnAskDelegateCreateObjectApproval(ctx context.Context, task task.ApprovalDelegateCreateObjectTask) (allow bool, err error) {
	if task == nil || task.GetDelegateCreateObject() == nil {
		log.CtxError(ctx, "failed to ask create object approval due to object info pointer dangling")
		return false, ErrApprovalTaskDangling
	}
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(ApproverFailureGetObjectApproval).Inc()
			metrics.ReqTime.WithLabelValues(ApproverFailureGetObjectApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ApproverSuccessGetObjectApproval).Inc()
			metrics.ReqTime.WithLabelValues(ApproverSuccessGetObjectApproval).Observe(time.Since(startTime).Seconds())
		}
	}()

	err = g.approver.PreCreateObjectApproval(ctx, nil)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre create object approval", "info", task.Info(), "error", err)
		return false, err
	}
	allow, err = g.approver.HandleDelegateCreateObjectApprovalTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to ask create object approval", "error", err)
		return false, err
	}
	log.CtxDebugw(ctx, "succeed to ask create object approval")
	return allow, nil
}
