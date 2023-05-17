package executor

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var _ module.TaskExecutor = &ExecuteModular{}

type ExecuteModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   corercmgr.ResourceScope

	maxExecuteNum int64
	executingNum  int64

	askTaskInterval int

	askReplicateApprovalTimeout  int64
	askReplicateApprovalExFactor float64

	listenSealTimeoutHeight int
	listenSealRetryTimeout  int
	maxListenSealRetry      int
}

func (e *ExecuteModular) Name() string {
	return module.ExecuteModularName
}

func (e *ExecuteModular) Start(ctx context.Context) error {
	scope, err := e.baseApp.ResourceManager().OpenService(e.Name())
	if err != nil {
		return err
	}
	e.scope = scope
	go e.eventLoop(ctx)
	return nil
}

func (e *ExecuteModular) eventLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(e.askTaskInterval))
	logCnt := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logCnt++
			maxExecuteNum := atomic.LoadInt64(&e.maxExecuteNum)
			executingNum := atomic.LoadInt64(&e.executingNum)
			metrics.MaxTaskNumberGauge.WithLabelValues(e.Name()).Set(float64(maxExecuteNum))
			metrics.RunningTaskNumberGauge.WithLabelValues(e.Name()).Set(float64(executingNum))

			if executingNum >= maxExecuteNum {
				if logCnt%1000 == 0 {
					log.CtxErrorw(ctx, "failed to start ask task, executing task exceed",
						"executing_num", executingNum, "max_execute_num", maxExecuteNum)
				}
				continue
			}
			if logCnt%1000 == 0 {
				log.CtxDebugw(ctx, "start to ask task", "executing_num", executingNum,
					"max_execute_num", maxExecuteNum)
			}
			atomic.AddInt64(&e.executingNum, 1)
			go func() {
				defer atomic.AddInt64(&e.executingNum, -1)
				limit, err := e.scope.RemainingResource()
				if err != nil {
					log.CtxErrorw(ctx, "failed to get remaining resource", "error", err)
					return
				}
				metrics.RemainingMemoryGauge.WithLabelValues(e.Name()).Set(float64(limit.GetMemoryLimit()))
				metrics.RemainingTaskGauge.WithLabelValues(e.Name()).Set(float64(limit.GetTaskTotalLimit()))
				metrics.RemainingHighPriorityTaskGauge.WithLabelValues(e.Name()).Set(
					float64(limit.GetTaskLimit(corercmgr.ReserveTaskPriorityHigh)))
				metrics.RemainingMediumPriorityTaskGauge.WithLabelValues(e.Name()).Set(
					float64(limit.GetTaskLimit(corercmgr.ReserveTaskPriorityMedium)))
				metrics.RemainingLowTaskGauge.WithLabelValues(e.Name()).Set(
					float64(limit.GetTaskLimit(corercmgr.ReserveTaskPriorityLow)))
				e.AskTask(ctx, limit)
			}()
		}
	}
}

func (e *ExecuteModular) AskTask(ctx context.Context, limit corercmgr.Limit) {
	askTask, err := e.baseApp.GfSpClient().AskTask(ctx, limit)
	if err != nil {
		switch e := err.(type) {
		case *gfsperrors.GfSpError:
			if e.GetInnerCode() == 60005 || e.GetInnerCode() == 990603 {
				return
			}
		default:
		}
		log.CtxWarnw(ctx, "failed to ask task", "remaining", limit.String(), "error", err)
		return
	}
	if askTask == nil {
		log.CtxWarnw(ctx, "failed to ask task, dangling pointer",
			"remaining", limit.String(), "error", err)
		return
	}
	span, err := e.ReserveResource(ctx, askTask.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve resource", "task_require",
			askTask.EstimateLimit().String(), "remaining", limit.String(), "error", err)
	}
	defer e.ReleaseResource(ctx, span)
	defer e.ReportTask(ctx, askTask)
	ctx = log.WithValue(ctx, log.CtxKeyTask, askTask.Key().String())
	switch t := askTask.(type) {
	case *gfsptask.GfSpReplicatePieceTask:
		e.HandleReplicatePieceTask(ctx, t)
	case *gfsptask.GfSpSealObjectTask:
		e.HandleSealObjectTask(ctx, t)
	case *gfsptask.GfSpReceivePieceTask:
		e.HandleReceivePieceTask(ctx, t)
	case *gfsptask.GfSpGCObjectTask:
		e.HandleGCObjectTask(ctx, t)
	case *gfsptask.GfSpGCZombiePieceTask:
		e.HandleGCZombiePieceTask(ctx, t)
	case *gfsptask.GfSpGCMetaTask:
		e.HandleGCMetaTask(ctx, t)
	default:
		log.CtxErrorw(ctx, "unsupported task type")
	}
	log.CtxDebugw(ctx, "finish to handle task")
	return
}

func (e *ExecuteModular) ReportTask(
	ctx context.Context,
	task coretask.Task) error {
	err := e.baseApp.GfSpClient().ReportTask(ctx, task)
	log.CtxDebugw(ctx, "finish to report task", "error", err)
	return err
}

func (e *ExecuteModular) Stop(ctx context.Context) error {
	e.scope.Release()
	return nil
}

func (e *ExecuteModular) ReserveResource(
	ctx context.Context,
	st *corercmgr.ScopeStat) (
	corercmgr.ResourceScopeSpan, error) {
	span, err := e.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(st)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (e *ExecuteModular) ReleaseResource(
	ctx context.Context,
	span corercmgr.ResourceScopeSpan) {
	span.Done()
}
