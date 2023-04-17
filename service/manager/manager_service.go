package manager

import (
	"context"
	"time"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	tqueuetypes "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue/types"
	"github.com/bnb-chain/greenfield-storage-provider/service/manager/types"
)

var _ types.ManagerServiceServer = &Manager{}

// AskUploadObject asks to create object to SP manager.
func (m *Manager) AskUploadObject(ctx context.Context, req *types.AskUploadObjectRequest) (
	*types.AskUploadObjectResponse, error) {
	task := req.GetUploadObjectTask()
	if task == nil {
		return nil, merrors.ErrDanglingTaskPointer
	}
	ctx = log.WithValue(ctx, "object_id", string(task.Key()))
	resp := &types.AskUploadObjectResponse{}
	if m.pqueue.HasTask(task.Key()) {
		resp.Allow = false
		resp.RefuseReason = merrors.ErrRepeatedTask.Error()
		return resp, merrors.ErrRepeatedTask
	}
	uploading := m.pqueue.GetUploadingTasksCount()
	if uploading >= m.config.UploadParallel {
		resp.Allow = false
		resp.RefuseReason = merrors.ErrExceedUploadParallel.Error()
		return resp, merrors.ErrRepeatedTask
	} else {
		resp.Allow = true
	}
	return resp, nil
}

// CreateUploadObjectTask asks to upload object to SP manager.
func (m *Manager) CreateUploadObjectTask(ctx context.Context, req *types.CreateUploadObjectTaskRequest) (
	*types.CreateUploadObjectTaskResponse, error) {
	resp := &types.CreateUploadObjectTaskResponse{}
	task := req.GetUploadObjectTask()
	if task == nil {
		return resp, merrors.ErrDanglingTaskPointer
	}
	ctx = log.WithValue(ctx, "object_id", string(task.Key()))
	if m.pqueue.HasTask(task.Key()) {
		oldTask := m.pqueue.PopTask(task.Key())
		if oldTask != nil && !oldTask.Expired() {
			log.CtxErrorw(ctx, "failed to create upload object task, has repeated task")
			m.pqueue.PushTask(oldTask)
			return resp, merrors.ErrRepeatedTask
		}
	}
	err := m.pqueue.PushTask(task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push upload object task to queue", "error", err)
		return resp, err
	}
	log.CtxDebugw(ctx, "succeed to push upload object task to queue")
	return resp, err
}

// DoneUploadObjectTask notifies the manager the upload object task has been done.
func (m *Manager) DoneUploadObjectTask(ctx context.Context, req *types.DoneUploadObjectTaskRequest) (
	*types.DoneUploadObjectTaskResponse, error) {
	resp := &types.DoneUploadObjectTaskResponse{}
	task := req.GetUploadObjectTask()
	if task == nil {
		return resp, merrors.ErrDanglingTaskPointer
	}
	ctx = log.WithValue(ctx, "object_id", string(task.Key()))
	m.pqueue.PopTask(task.Key())
	replicateTask, err := tqueuetypes.NewReplicatePieceTask(task.GetObject())
	if err != nil {
		log.CtxErrorw(ctx, "failed to make replicate piece task", "error", err)
		return resp, err
	}
	err = m.pqueue.PushTask(replicateTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push replicate task to queue", "error", err)
		return resp, err
	}
	log.CtxDebugw(ctx, "succeed to push replicate task to queue")
	return resp, err
}

// DoneReplicatePieceTask notifies the manager the replicate piece task has been done.
func (m *Manager) DoneReplicatePieceTask(ctx context.Context, req *types.DoneReplicatePieceTaskRequest) (
	*types.DoneReplicatePieceTaskResponse, error) {
	resp := &types.DoneReplicatePieceTaskResponse{}
	task := req.GetReplicatePieceTask()
	if task == nil {
		return resp, merrors.ErrDanglingTaskPointer
	}
	ctx = log.WithValue(ctx, "object_id", string(task.Key()))
	if task.Error() != nil {
		log.CtxErrorw(ctx, "failed execute replicate piece task", "error", task.Error())
		if task.RetryExceed() {
			m.pqueue.PopTask(task.Key())
			log.CtxErrorw(ctx, "cancel replicate piece task", "retry", task.GetRetry(),
				"retry_limit", task.GetRetryLimit(), "time_out", task.GetTimeout())
			return resp, nil
		}
		task.SetUpdateTime(time.Now().Unix())
		m.pqueue.PopPushTask(task)
		return resp, nil
	}
	m.pqueue.PopTask(task.Key())
	sealTask, err := tqueuetypes.NewSealObjectTask(task.GetObject())
	if err != nil {
		log.CtxErrorw(ctx, "failed to make seal task", "error", err)
		return resp, err
	}
	err = m.pqueue.PushTask(sealTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push seal object task to queue", "error", err)
		return resp, err
	}
	log.CtxDebugw(ctx, "succeed to push seal object task to queue")
	return resp, err
}

// DoneSealObjectTask notifies the manager the seal object task has been done.
func (m *Manager) DoneSealObjectTask(ctx context.Context, req *types.DoneSealObjectTaskRequest) (
	*types.DoneSealObjectTaskResponse, error) {
	resp := &types.DoneSealObjectTaskResponse{}
	task := req.GetSealObjectTask()
	if task == nil {
		return resp, merrors.ErrDanglingTaskPointer
	}
	ctx = log.WithValue(ctx, "object_id", string(task.Key()))
	if task.Error() != nil {
		log.CtxErrorw(ctx, "failed execute seal object task", "error", task.Error())
		if task.RetryExceed() {
			m.pqueue.PopTask(task.Key())
			log.CtxErrorw(ctx, "cancel seal object task", "retry", task.GetRetry(),
				"retry_limit", task.GetRetryLimit(), "time_out", task.GetTimeout())
			return resp, nil
		}
		task.SetUpdateTime(time.Now().Unix())
		m.pqueue.PopPushTask(task)
		return resp, nil
	}
	m.pqueue.PopTask(task.Key())
	log.CtxDebugw(ctx, "succeed to seal object on chain")
	return resp, nil
}

// AskTask asks the task to execute
func (m *Manager) AskTask(ctx context.Context, req *types.AskTaskRequest) (*types.AskTaskResponse, error) {
	resp := &types.AskTaskResponse{}
	limit := req.GetLimit().TransferRcmgrLimits()
	task := m.pqueue.PopTaskByLimit(limit)
	if task == nil {
		return resp, nil
	}
	defer func() {
		task.IncRetry()
		task.SetUpdateTime(time.Now().Unix())
		m.pqueue.PushTask(task)
	}()
	switch t := (task).(type) {
	case *tqueuetypes.ReplicatePieceTask:
		resp.Task = &types.AskTaskResponse_ReplicatePieceTask{
			ReplicatePieceTask: t,
		}
	case *tqueuetypes.SealObjectTask:
		resp.Task = &types.AskTaskResponse_SealObjectTask{
			SealObjectTask: t,
		}
	case *tqueuetypes.GCObjectTask:
		resp.Task = &types.AskTaskResponse_GcObjectTask{
			GcObjectTask: t,
		}
	default:
		log.CtxErrorw(ctx, "task node does not support task type", "task_type", task.Type())
		return resp, merrors.ErrUnsupportedDispatchTaskType
	}
	resp.HasTask = true
	return resp, nil
}

// DoneGCObjectTask notifies the manager the gc object task has been done.
func (m *Manager) DoneGCObjectTask(ctx context.Context, req *types.DoneGCObjectTaskRequest) (
	*types.DoneGCObjectTaskResponse, error) {
	resp := &types.DoneGCObjectTaskResponse{}
	task := req.GetGcObjectTask()
	if task == nil {
		return resp, merrors.ErrDanglingTaskPointer
	}
	ctx = log.WithValue(ctx, "object_id", string(task.Key()))
	if task.Error() != nil {
		log.CtxErrorw(ctx, "failed execute gc object task", "error", task.Error())
		task.SetUpdateTime(time.Now().Unix())
		m.pqueue.PopPushTask(task)
		return resp, nil
	}
	m.pqueue.PopTask(task.Key())
	log.CtxInfow(ctx, "succeed to run gc object")
	return resp, nil
}

// ReportGCObjectProcess notifies the manager the gc object task process.
func (m *Manager) ReportGCObjectProcess(ctx context.Context, req *types.ReportGCObjectProcessRequest) (
	*types.ReportGCObjectProcessResponse, error) {
	resp := &types.ReportGCObjectProcessResponse{}
	task := req.GetGcObjectTask()
	if task == nil {
		return resp, merrors.ErrDanglingTaskPointer
	}
	ctx = log.WithValue(ctx, "object_id", string(task.Key()))
	if !m.pqueue.HasTask(task.Key()) {
		log.CtxErrorw(ctx, "failed to report gc object process", "error", merrors.ErrTaskCanceled)
		resp.Cancel = true
		return resp, merrors.ErrTaskCanceled
	}
	task.SetUpdateTime(time.Now().Unix())
	m.pqueue.PopPushTask(task)
	return resp, nil
}
