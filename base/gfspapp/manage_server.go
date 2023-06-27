package gfspapp

import (
	"context"
	"fmt"
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
	case *gfspserver.GfSpBeginTaskRequest_ResumableUploadObjectTask:
		err := g.OnBeginResumableUploadObjectTask(ctx, task.ResumableUploadObjectTask)
		return &gfspserver.GfSpBeginTaskResponse{Err: gfsperrors.MakeGfSpError(err)}, nil
	default:
		return &gfspserver.GfSpBeginTaskResponse{Err: ErrUnsupportedTaskType}, nil
	}
}

func (g *GfSpBaseApp) OnBeginUploadObjectTask(ctx context.Context, task coretask.UploadObjectTask) (err error) {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to begin upload object task due to object info pointer dangling")
		return ErrUploadTaskDangling
	}
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(ManagerFailureBeginUpload).Inc()
			metrics.ReqTime.WithLabelValues(ManagerFailureBeginUpload).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(ManagerBeginUpload).Inc()
			metrics.ReqTime.WithLabelValues(ManagerBeginUpload).Observe(time.Since(startTime).Seconds())
		}
	}()

	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	err = g.manager.HandleCreateUploadObjectTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin upload object task", "info", task.Info(), "error", err)
		return err
	}
	log.CtxDebugw(ctx, "succeed to begin upload object task", "info", task.Info())
	return nil
}

func (g *GfSpBaseApp) OnBeginResumableUploadObjectTask(ctx context.Context, task coretask.ResumableUploadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil {
		log.CtxError(ctx, "failed to begin resumable upload object task due to object info pointer dangling")
		return ErrUploadTaskDangling
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	err := g.manager.HandleCreateResumableUploadObjectTask(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin resumable upload object task", "info", task.Info(), "error", err)
		return err
	}
	log.CtxDebugw(ctx, "succeed to begin resumable upload object task", "info", task.Info())
	return nil
}

func (g *GfSpBaseApp) GfSpAskTask(ctx context.Context, req *gfspserver.GfSpAskTaskRequest) (*gfspserver.GfSpAskTaskResponse, error) {
	startTime := time.Now()
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

	defer func() {
		metrics.ReqCounter.WithLabelValues(ManagerSuccessDispatchTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerSuccessDispatchTask).Observe(time.Since(startTime).Seconds())
	}()

	switch t := gfspTask.(type) {
	case *gfsptask.GfSpReplicatePieceTask:
		t.AppendLog(fmt.Sprintf("manager-dispatch-replicate-task-retry:%d", t.GetRetry()))
		resp.Response = &gfspserver.GfSpAskTaskResponse_ReplicatePieceTask{
			ReplicatePieceTask: t,
		}
		if t.GetRetry() == 1 {
			metrics.PerfPutObjectTime.WithLabelValues("manager_put_object_replicate_schedule_cost").Observe(
				time.Since(time.Unix(t.GetCreateTime(), 0)).Seconds())
		} else {
			metrics.PerfPutObjectTime.WithLabelValues("manager_put_object_replicate_retry_schedule_cost").Observe(
				time.Since(time.Unix(t.GetCreateTime(), 0)).Seconds())
		}
		metrics.ReqCounter.WithLabelValues(ManagerDispatchReplicateTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerDispatchReplicateTask).Observe(time.Since(startTime).Seconds())
	case *gfsptask.GfSpSealObjectTask:
		t.AppendLog(fmt.Sprintf("manager-dispatch-seal-task-retry:%d", t.GetRetry()))
		resp.Response = &gfspserver.GfSpAskTaskResponse_SealObjectTask{
			SealObjectTask: t,
		}
		if t.GetRetry() == 1 {
			metrics.PerfPutObjectTime.WithLabelValues("manager_put_object_seal_schedule_cost").Observe(
				time.Since(time.Unix(t.GetCreateTime(), 0)).Seconds())
		} else {
			metrics.PerfPutObjectTime.WithLabelValues("manager_put_object_replicate_retry_schedule_cost").Observe(
				time.Since(time.Unix(t.GetCreateTime(), 0)).Seconds())
		}
		metrics.ReqCounter.WithLabelValues(ManagerDispatchSealTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerDispatchSealTask).Observe(time.Since(startTime).Seconds())
	case *gfsptask.GfSpReceivePieceTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_ReceivePieceTask{
			ReceivePieceTask: t,
		}
		if t.GetRetry() == 1 {
			metrics.PerfPutObjectTime.WithLabelValues("manager_put_object_receive_schedule_cost").Observe(
				time.Since(time.Unix(t.GetCreateTime(), 0)).Seconds())
		} else {
			metrics.PerfPutObjectTime.WithLabelValues("manager_put_object_replicate_retry_schedule_cost").Observe(
				time.Since(time.Unix(t.GetCreateTime(), 0)).Seconds())
		}
		metrics.ReqCounter.WithLabelValues(ManagerDispatchReceiveTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerDispatchReceiveTask).Observe(time.Since(startTime).Seconds())
	case *gfsptask.GfSpGCObjectTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_GcObjectTask{
			GcObjectTask: t,
		}
		metrics.ReqCounter.WithLabelValues(ManagerDispatchGCObjectTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerDispatchGCObjectTask).Observe(time.Since(startTime).Seconds())
	case *gfsptask.GfSpGCZombiePieceTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_GcZombiePieceTask{
			GcZombiePieceTask: t,
		}
	case *gfsptask.GfSpGCMetaTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_GcMetaTask{
			GcMetaTask: t,
		}
	case *gfsptask.GfSpRecoverPieceTask:
		resp.Response = &gfspserver.GfSpAskTaskResponse_RecoverPieceTask{
			RecoverPieceTask: t,
		}
		metrics.ReqCounter.WithLabelValues(ManagerDispatchRecoveryTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerDispatchRecoveryTask).Observe(time.Since(startTime).Seconds())
	default:
		log.CtxErrorw(ctx, "[BUG] Unsupported task type to dispatch")
		return &gfspserver.GfSpAskTaskResponse{Err: ErrUnsupportedTaskType}, nil
	}
	log.CtxDebugw(ctx, "succeed to response ask task")
	return resp, nil
}

func (g *GfSpBaseApp) OnAskTask(ctx context.Context, limit corercmgr.Limit) (coretask.Task, error) {
	startTime := time.Now()
	gfspTask, err := g.manager.DispatchTask(ctx, limit)
	if err != nil {
		metrics.ReqCounter.WithLabelValues(ManagerFailureDispatchTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerFailureDispatchTask).Observe(time.Since(startTime).Seconds())
		return nil, gfsperrors.MakeGfSpError(err)
	}
	if gfspTask == nil {
		metrics.ReqCounter.WithLabelValues(ManagerNoDispatchTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerNoDispatchTask).Observe(time.Since(startTime).Seconds())
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

	startTime := time.Now()
	defer func() {
		metrics.ReqCounter.WithLabelValues(ManagerReportTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerReportTask).Observe(time.Since(startTime).Seconds())
	}()

	switch t := reportTask.(type) {
	case *gfspserver.GfSpReportTaskRequest_UploadObjectTask:
		task := t.UploadObjectTask
		task.AppendLog(fmt.Sprintf("manager-receive-upload-task-error:%s-retry:%d", task.Error().Error(), task.GetRetry()))
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		_ = g.GfSpDB().InsertPutEvent(task)
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())
		err = g.manager.HandleDoneUploadObjectTask(ctx, t.UploadObjectTask)
		metrics.ReqCounter.WithLabelValues(ManagerReportUploadTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerReportUploadTask).Observe(time.Since(startTime).Seconds())
	case *gfspserver.GfSpReportTaskRequest_ResumableUploadObjectTask:
		task := t.ResumableUploadObjectTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())
		err = g.manager.HandleDoneResumableUploadObjectTask(ctx, t.ResumableUploadObjectTask)
	case *gfspserver.GfSpReportTaskRequest_ReplicatePieceTask:
		task := t.ReplicatePieceTask
		task.AppendLog(fmt.Sprintf("manager-receive-replicate-task-error:%s-retry:%d", task.Error().Error(), task.GetRetry()))
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())
		err = g.manager.HandleReplicatePieceTask(ctx, t.ReplicatePieceTask)
		metrics.ReqCounter.WithLabelValues(ManagerReportReplicateTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerReportReplicateTask).Observe(time.Since(startTime).Seconds())
	case *gfspserver.GfSpReportTaskRequest_SealObjectTask:
		task := t.SealObjectTask
		task.AppendLog(fmt.Sprintf("manager-receive-seal-task-error:%s-retry:%d", task.Error().Error(), task.GetRetry()))
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())
		err = g.manager.HandleSealObjectTask(ctx, t.SealObjectTask)
		metrics.ReqCounter.WithLabelValues(ManagerReportSealTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerReportSealTask).Observe(time.Since(startTime).Seconds())
	case *gfspserver.GfSpReportTaskRequest_ReceivePieceTask:
		task := t.ReceivePieceTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())
		err = g.manager.HandleReceivePieceTask(ctx, t.ReceivePieceTask)
		metrics.ReqCounter.WithLabelValues(ManagerReportReceiveTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerReportReceiveTask).Observe(time.Since(startTime).Seconds())
	case *gfspserver.GfSpReportTaskRequest_GcObjectTask:
		task := t.GcObjectTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle reported task", "task_info", task.Info())
		err = g.manager.HandleGCObjectTask(ctx, t.GcObjectTask)
		metrics.ReqCounter.WithLabelValues(ManagerReportGCObjectTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerReportGCObjectTask).Observe(time.Since(startTime).Seconds())
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
	case *gfspserver.GfSpReportTaskRequest_RecoverPieceTask:
		task := t.RecoverPieceTask
		ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
		task.SetAddress(GetRPCRemoteAddress(ctx))
		log.CtxInfow(ctx, "begin to handle recovery reported task", "task_info", task.Info())
		err = g.manager.HandleRecoverPieceTask(ctx, t.RecoverPieceTask)
		metrics.ReqCounter.WithLabelValues(ManagerReportRecoveryTask).Inc()
		metrics.ReqTime.WithLabelValues(ManagerReportRecoveryTask).Observe(time.Since(startTime).Seconds())
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
