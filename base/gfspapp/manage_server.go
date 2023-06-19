package gfspapp

import (
	"context"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrUploadTaskDangling  = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 990601, "OoooH... request lost")
	ErrUnsupportedTaskType = gfsperrors.Register(BaseCodeSpace, http.StatusNotFound, 990602, "unsupported task type")
	ErrNoTaskMatchLimit    = gfsperrors.Register(BaseCodeSpace, http.StatusNotFound, 990603, "no task to dispatch below the require limits")
)

var _ gfspserver.GfSpManageServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpBeginTask(ctx context.Context, req *gfspserver.GfSpBeginTaskRequest) (*gfspserver.GfSpBeginTaskResponse, error) {
	if req.GetRequest() == nil {
		log.Error("failed to begin task due to pointer dangling")
		return &gfspserver.GfSpBeginTaskResponse{Err: ErrUploadTaskDangling}, nil
	}
	switch task := req.GetRequest().(type) {
	case *gfspserver.GfSpBeginTaskRequest_UploadObjectTask:
		err := g.OnBeginUploadObjectTask(ctx, task.UploadObjectTask)
		return &gfspserver.GfSpBeginTaskResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	default:
		return &gfspserver.GfSpBeginTaskResponse{Err: ErrUnsupportedTaskType}, nil
	}
}

func (g *GfSpBaseApp) OnBeginUploadObjectTask(ctx context.Context, task coretask.UploadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to begin upload object task due to object info pointer dangling")
		return ErrUploadTaskDangling
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	err := g.manager.HandleCreateUploadObjectTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin upload object task", "info", task.Info(), "error", err)
		return err
	}
	log.CtxDebugw(ctx, "succeed to begin upload object task", "info", task.Info())
	return nil
}

func (g *GfSpBaseApp) GfSpAskTask(ctx context.Context, req *gfspserver.GfSpAskTaskRequest) (*gfspserver.GfSpAskTaskResponse, error) {
	gfspTask, err := g.OnAskTask(ctx, req.GetNodeLimit())
	if err != nil {
		log.CtxErrorw(ctx, "failed to dispatch task", "error", err)
		return &gfspserver.GfSpAskTaskResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	}
	if gfspTask == nil {
		return &gfspserver.GfSpAskTaskResponse{Err: ErrNoTaskMatchLimit}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, gfspTask.Key().String())
	resp := &gfspserver.GfSpAskTaskResponse{}
	switch t := gfspTask.(type) {
	case *gfsptask.GfSpReplicatePieceTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_ReplicatePieceTask{
			ReplicatePieceTask: t,
		}
		metrics.DispatchReplicatePieceTaskCounter.WithLabelValues(g.manager.Name()).Inc()
	case *gfsptask.GfSpSealObjectTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_SealObjectTask{
			SealObjectTask: t,
		}
		metrics.DispatchSealObjectTaskCounter.WithLabelValues(g.manager.Name()).Inc()
	case *gfsptask.GfSpReceivePieceTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_ReceivePieceTask{
			ReceivePieceTask: t,
		}
		metrics.DispatchReceivePieceTaskCounter.WithLabelValues(g.manager.Name()).Inc()
	case *gfsptask.GfSpGCObjectTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_GcObjectTask{
			GcObjectTask: t,
		}
		metrics.DispatchGcObjectTaskCounter.WithLabelValues(g.manager.Name()).Inc()
	case *gfsptask.GfSpGCZombiePieceTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_GcZombiePieceTask{
			GcZombiePieceTask: t,
		}
	case *gfsptask.GfSpGCMetaTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_GcMetaTask{
			GcMetaTask: t,
		}
	case *gfsptask.GfSpRecoveryPieceTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_RecoveryPieceTask{
			RecoveryPieceTask: t,
		}
	default:
		log.CtxErrorw(ctx, "[BUG] Unsupported task type to dispatch")
		return &gfspserver.GfSpAskTaskResponse{Err: ErrUnsupportedTaskType}, nil
	}
	log.CtxDebugw(ctx, "succeed to response ask task")
	return resp, nil
}

func (g *GfSpBaseApp) OnAskTask(ctx context.Context, limit corercmgr.Limit) (coretask.Task, error) {
	gfspTask, err := g.manager.DispatchTask(ctx, limit)
	if err != nil {
		return nil, gfsperrors.MakeGfSpError(err)
	}
	if gfspTask == nil {
		return nil, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, gfspTask.Key().String())
	gfspTask.IncRetry()
	gfspTask.SetError(nil)
	gfspTask.SetUpdateTime(time.Now().Unix())
	gfspTask.SetAddress(GetRPCRemoteAddress(ctx))
	log.CtxDebugw(ctx, "succeed to dispatch task", "info", gfspTask.Info())
	return gfspTask, nil
}

func (g *GfSpBaseApp) GfSpReportTask(ctx context.Context, req *gfspserver.GfSpReportTaskRequest) (
	*gfspserver.GfSpReportTaskResponse, error) {
	var (
		reportTask = req.GetRequest()
		err        error
	)
	if reportTask == nil {
		log.CtxError(ctx, "failed to receive report task due to object info pointer dangling")
		return &gfspserver.GfSpReportTaskResponse{Err: ErrUploadTaskDangling}, nil
	}
	switch t := reportTask.(type) {
	case *gfspserver.GfSpReportTaskRequest_UploadObjectTask:
		task := t.UploadObjectTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		metrics.UploadObjectTaskTimeHistogram.WithLabelValues(g.manager.Name()).Observe(
			time.Since(time.Unix(t.UploadObjectTask.GetCreateTime(), 0)).Seconds())
		if t.UploadObjectTask.Error() != nil {
			metrics.UploadObjectTaskFailedCounter.WithLabelValues(g.manager.Name()).Inc()
		}

		startReportDoneUploadTask := time.Now()
		err = g.manager.HandleDoneUploadObjectTask(ctx, t.UploadObjectTask)
		metrics.PerfUploadTimeHistogram.WithLabelValues("report_upload_task_done_server").
			Observe(time.Since(startReportDoneUploadTask).Seconds())
	case *gfspserver.GfSpReportTaskRequest_ReplicatePieceTask:
		task := t.ReplicatePieceTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		metrics.ReplicateAndSealTaskTimeHistogram.WithLabelValues(g.manager.Name()).Observe(
			time.Since(time.Unix(t.ReplicatePieceTask.GetCreateTime(), 0)).Seconds())
		if t.ReplicatePieceTask.Error() != nil {
			metrics.ReplicatePieceTaskFailedCounter.WithLabelValues(g.manager.Name()).Inc()
		}
		if !t.ReplicatePieceTask.GetSealed() {
			metrics.ReplicateCombineSealTaskFailedCounter.WithLabelValues(g.manager.Name()).Inc()
		}

		err = g.manager.HandleReplicatePieceTask(ctx, t.ReplicatePieceTask)
	case *gfspserver.GfSpReportTaskRequest_SealObjectTask:
		task := t.SealObjectTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		metrics.SealObjectTaskTimeHistogram.WithLabelValues(g.manager.Name()).Observe(
			time.Since(time.Unix(t.SealObjectTask.GetCreateTime(), 0)).Seconds())
		if t.SealObjectTask.Error() != nil {
			metrics.SealObjectTaskFailedCounter.WithLabelValues(g.manager.Name()).Inc()
		}

		err = g.manager.HandleSealObjectTask(ctx, t.SealObjectTask)
	case *gfspserver.GfSpReportTaskRequest_ReceivePieceTask:
		task := t.ReceivePieceTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		metrics.ReceiveTaskTimeHistogram.WithLabelValues(g.manager.Name()).Observe(
			time.Since(time.Unix(t.ReceivePieceTask.GetCreateTime(), 0)).Seconds())
		if t.ReceivePieceTask.Error() != nil {
			metrics.ReceivePieceTaskFailedCounter.WithLabelValues(g.manager.Name()).Inc()
		}

		err = g.manager.HandleReceivePieceTask(ctx, t.ReceivePieceTask)
	case *gfspserver.GfSpReportTaskRequest_GcObjectTask:
		task := t.GcObjectTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		err = g.manager.HandleGCObjectTask(ctx, t.GcObjectTask)
	case *gfspserver.GfSpReportTaskRequest_GcZombiePieceTask:
		task := t.GcZombiePieceTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		err = g.manager.HandleGCZombiePieceTask(ctx, t.GcZombiePieceTask)
	case *gfspserver.GfSpReportTaskRequest_GcMetaTask:
		task := t.GcMetaTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		err = g.manager.HandleGCMetaTask(ctx, t.GcMetaTask)
	case *gfspserver.GfSpReportTaskRequest_DownloadObjectTask:
		task := t.DownloadObjectTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		err = g.manager.HandleDownloadObjectTask(ctx, t.DownloadObjectTask)
	case *gfspserver.GfSpReportTaskRequest_ChallengePieceTask:
		task := t.ChallengePieceTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		err = g.manager.HandleChallengePieceTask(ctx, t.ChallengePieceTask)
	case *gfspserver.GfSpReportTaskRequest_RecoveryPieceTask:
		task := t.RecoveryPieceTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())

		err = g.manager.HandleRecoveryPieceTask(ctx, t.RecoveryPieceTask)
	default:
		log.CtxErrorw(ctx, "receive unsupported task type")
		return &gfspserver.GfSpReportTaskResponse{Err: ErrUnsupportedTaskType}, nil
	}
	if err != nil {
		log.CtxErrorw(ctx, "failed to report task", "error", err)
		return &gfspserver.GfSpReportTaskResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	}
	log.CtxInfow(ctx, "succeed to handle reported task")
	return &gfspserver.GfSpReportTaskResponse{}, nil
}
