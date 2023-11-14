package executor

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var _ coremodule.TaskExecutor = &ExecuteModular{}

type ExecuteModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   corercmgr.ResourceScope

	maxExecuteNum int64
	executingNum  int64

	askTaskInterval int

	askReplicateApprovalTimeout  int64
	askReplicateApprovalExFactor float64

	listenSealTimeoutHeight     int
	listenSealRetryTimeout      int
	maxListenSealRetry          int
	maxObjectMigrationRetry     int
	objectMigrationRetryTimeout int

	statisticsOutputInterval   int
	doingReplicatePieceTaskCnt int64
	doingSpSealObjectTaskCnt   int64
	doingReceivePieceTaskCnt   int64
	doingGCObjectTaskCnt       int64
	doingGCZombiePieceTaskCnt  int64
	doingGCGCMetaTaskCnt       int64
	doingRecoveryPieceTaskCnt  int64
	doingMigrationGVGTaskCnt   int64

	enableSkipFailedToMigrateObject bool // only for debugging, and online config can only be false

	spID  uint32
	spMap map[uint32]*sptypes.StorageProvider
	mutex sync.RWMutex
}

func (e *ExecuteModular) Name() string {
	return coremodule.ExecuteModularName
}

func (e *ExecuteModular) Start(ctx context.Context) error {
	scope, err := e.baseApp.ResourceManager().OpenService(e.Name())
	if err != nil {
		return err
	}
	e.scope = scope
	sps, err := e.baseApp.Consensus().ListSPs(ctx)
	if err != nil {
		return err
	}
	e.spMap = make(map[uint32]*sptypes.StorageProvider)
	for _, sp := range sps {
		e.spMap[sp.Id] = sp
	}
	go e.eventLoop(ctx)
	return nil
}

func (e *ExecuteModular) eventLoop(ctx context.Context) {
	for i := int64(0); i < e.maxExecuteNum; i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
				default:
					err := e.AskTask(ctx)
					if err != nil {
						rand.New(rand.NewSource(time.Now().Unix()))
						sleep := rand.Intn(int(e.maxExecuteNum)) + 1
						time.Sleep(time.Duration(sleep) * time.Millisecond)
					}
				}
			}
		}(ctx)
	}

	statisticsTicker := time.NewTicker(time.Duration(e.statisticsOutputInterval) * time.Second)
	updateSpTicker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-statisticsTicker.C:
			log.CtxInfo(ctx, e.Statistics())
		case <-updateSpTicker.C:
			sps, err := e.baseApp.Consensus().ListSPs(ctx)
			if err != nil {
				continue
			}
			e.mutex.Lock()
			for _, sp := range sps {
				e.spMap[sp.Id] = sp
			}
			e.mutex.Unlock()
		}
	}
}

func (e *ExecuteModular) getSpByID(id uint32) *sptypes.StorageProvider {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.spMap[id]
}

func (e *ExecuteModular) omitError(err error) bool {
	switch realErr := err.(type) {
	case *gfsperrors.GfSpError:
		if realErr.GetInnerCode() == gfspapp.ErrNoTaskMatchLimit.GetInnerCode() {
			return true
		}
	}
	return false
}

func (e *ExecuteModular) AskTask(ctx context.Context) error {
	startTime := time.Now()
	limit, err := e.scope.RemainingResource()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get remaining resource", "error", err)
		return err
	}

	metrics.RemainingMemoryGauge.WithLabelValues(e.Name()).Set(float64(limit.GetMemoryLimit()))
	metrics.RemainingTaskGauge.WithLabelValues(e.Name()).Set(float64(limit.GetTaskTotalLimit()))
	metrics.RemainingHighPriorityTaskGauge.WithLabelValues(e.Name()).Set(
		float64(limit.GetTaskLimit(corercmgr.ReserveTaskPriorityHigh)))
	metrics.RemainingMediumPriorityTaskGauge.WithLabelValues(e.Name()).Set(
		float64(limit.GetTaskLimit(corercmgr.ReserveTaskPriorityMedium)))
	metrics.RemainingLowTaskGauge.WithLabelValues(e.Name()).Set(
		float64(limit.GetTaskLimit(corercmgr.ReserveTaskPriorityLow)))

	askTask, err := e.baseApp.GfSpClient().AskTask(ctx, limit)
	if err != nil {
		metrics.ReqCounter.WithLabelValues(ExecutorFailureAskNoTask).Inc()
		metrics.ReqTime.WithLabelValues(ExecutorFailureAskNoTask).Observe(time.Since(startTime).Seconds())
		if e.omitError(err) {
			return err
		}
		log.CtxErrorw(ctx, "failed to ask task", "remaining", limit.String(), "error", err)
		return err
	}
	// double confirm the safe task
	if askTask == nil {
		metrics.ReqCounter.WithLabelValues(ExecutorFailureAskTask).Inc()
		metrics.ReqTime.WithLabelValues(ExecutorFailureAskTask).Observe(time.Since(startTime).Seconds())
		log.CtxErrorw(ctx, "failed to ask task due to dangling pointer", "remaining", limit.String(),
			"error", err)
		return ErrDanglingPointer
	}
	metrics.ReqCounter.WithLabelValues(ExecutorSuccessAskTask).Inc()
	metrics.ReqTime.WithLabelValues(ExecutorSuccessAskTask).Observe(time.Since(startTime).Seconds())

	atomic.AddInt64(&e.executingNum, 1)
	defer atomic.AddInt64(&e.executingNum, -1)
	span, err := e.ReserveResource(ctx, askTask.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve resource", "task_require",
			askTask.EstimateLimit().String(), "remaining", limit.String(), "error", err)
		return err
	}
	metrics.RunningTaskNumberGauge.WithLabelValues("running_task_num").Set(float64(atomic.LoadInt64(&e.executingNum)))
	metrics.MaxTaskNumberGauge.WithLabelValues("max_task_num").Set(float64(atomic.LoadInt64(&e.maxExecuteNum)))

	defer func() {
		e.ReleaseResource(ctx, span)
		go e.ReportTask(context.Background(), askTask)
	}()

	runTime := time.Now()
	defer func() {
		metrics.ReqCounter.WithLabelValues(ExecutorRunTask).Inc()
		metrics.ReqTime.WithLabelValues(ExecutorRunTask).Observe(time.Since(runTime).Seconds())
	}()

	ctx = log.WithValue(ctx, log.CtxKeyTask, askTask.Key().String())
	switch t := askTask.(type) {
	case *gfsptask.GfSpReplicatePieceTask:
		atomic.AddInt64(&e.doingReplicatePieceTaskCnt, 1)
		defer atomic.AddInt64(&e.doingReplicatePieceTaskCnt, -1)
		e.HandleReplicatePieceTask(ctx, t)
		if t.Error() != nil {
			metrics.ReqCounter.WithLabelValues(ExecutorFailureReplicateTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorFailureReplicateTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ExecutorSuccessReplicateTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorSuccessReplicateTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfsptask.GfSpSealObjectTask:
		atomic.AddInt64(&e.doingSpSealObjectTaskCnt, 1)
		defer atomic.AddInt64(&e.doingSpSealObjectTaskCnt, -1)
		e.HandleSealObjectTask(ctx, t)
		if t.Error() != nil {
			metrics.ReqCounter.WithLabelValues(ExecutorFailureSealObjectTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorFailureSealObjectTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ExecutorSuccessSealObjectTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorSuccessSealObjectTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfsptask.GfSpReceivePieceTask:
		atomic.AddInt64(&e.doingReceivePieceTaskCnt, 1)
		defer atomic.AddInt64(&e.doingReceivePieceTaskCnt, -1)
		e.HandleReceivePieceTask(ctx, t)
		if t.Error() != nil {
			metrics.ReqCounter.WithLabelValues(ExecutorFailureReceiveTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorFailureReceiveTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ExecutorSuccessReceiveTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorSuccessReceiveTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfsptask.GfSpGCObjectTask:
		atomic.AddInt64(&e.doingGCObjectTaskCnt, 1)
		defer atomic.AddInt64(&e.doingGCObjectTaskCnt, -1)
		e.HandleGCObjectTask(ctx, t)
	case *gfsptask.GfSpGCZombiePieceTask:
		atomic.AddInt64(&e.doingGCZombiePieceTaskCnt, 1)
		defer atomic.AddInt64(&e.doingGCZombiePieceTaskCnt, -1)
		e.HandleGCZombiePieceTask(ctx, t)
	case *gfsptask.GfSpGCMetaTask:
		atomic.AddInt64(&e.doingGCGCMetaTaskCnt, 1)
		defer atomic.AddInt64(&e.doingGCGCMetaTaskCnt, -1)
		e.HandleGCMetaTask(ctx, t)
	case *gfsptask.GfSpRecoverPieceTask:
		atomic.AddInt64(&e.doingRecoveryPieceTaskCnt, 1)
		defer atomic.AddInt64(&e.doingRecoveryPieceTaskCnt, -1)
		e.HandleRecoverPieceTask(ctx, t)
		if t.Error() != nil {
			metrics.ReqCounter.WithLabelValues(ExecutorFailureRecoveryTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorFailureRecoveryTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ExecutorSuccessRecoveryTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorSuccessRecoveryTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfsptask.GfSpMigrateGVGTask:
		atomic.AddInt64(&e.doingMigrationGVGTaskCnt, 1)
		defer atomic.AddInt64(&e.doingMigrationGVGTaskCnt, -1)
		e.HandleMigrateGVGTask(ctx, t)
	default:
		log.CtxError(ctx, "unsupported task type")
	}
	log.CtxDebug(ctx, "finished to handle task")
	return nil
}

func (e *ExecuteModular) ReportTask(ctx context.Context, task coretask.Task) (err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(ExecutorFailureReportTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorFailureReportTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ExecutorSuccessReportTask).Inc()
			metrics.ReqTime.WithLabelValues(ExecutorSuccessReportTask).Observe(time.Since(startTime).Seconds())
		}
	}()

	err = e.baseApp.GfSpClient().ReportTask(ctx, task)
	log.CtxDebugw(ctx, "finished to report task", "task_info", task.Info(), "error", err)
	return err
}

func (e *ExecuteModular) Stop(ctx context.Context) error {
	e.scope.Release()
	return nil
}

func (e *ExecuteModular) ReserveResource(ctx context.Context, st *corercmgr.ScopeStat) (corercmgr.ResourceScopeSpan, error) {
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

func (e *ExecuteModular) ReleaseResource(ctx context.Context, span corercmgr.ResourceScopeSpan) {
	span.Done()
}

func (e *ExecuteModular) Statistics() string {
	return fmt.Sprintf(
		"maxAsk[%d], asking[%d], replicate[%d], seal[%d], receive[%d], gcObject[%d], gcZombie[%d], gcMeta[%d], migrateGVG[%d]",
		atomic.LoadInt64(&e.maxExecuteNum), atomic.LoadInt64(&e.executingNum),
		atomic.LoadInt64(&e.doingReplicatePieceTaskCnt),
		atomic.LoadInt64(&e.doingSpSealObjectTaskCnt),
		atomic.LoadInt64(&e.doingReceivePieceTaskCnt),
		atomic.LoadInt64(&e.doingGCObjectTaskCnt),
		atomic.LoadInt64(&e.doingGCZombiePieceTaskCnt),
		atomic.LoadInt64(&e.doingGCGCMetaTaskCnt),
		atomic.LoadInt64(&e.doingMigrationGVGTaskCnt))
}

func (e *ExecuteModular) getSPID() (uint32, error) {
	if e.spID != 0 {
		return e.spID, nil
	}
	spInfo, err := e.baseApp.Consensus().QuerySP(context.Background(), e.baseApp.OperatorAddress())
	if err != nil {
		return 0, err
	}
	e.spID = spInfo.GetId()
	return e.spID, nil
}
