package manager

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/types"
)

var (
	ErrDanglingTask     = gfsperrors.Register(module.ManageModularName, http.StatusInternalServerError, 60001, "OoooH... request lost")
	ErrRepeatedTask     = gfsperrors.Register(module.ManageModularName, http.StatusBadRequest, 60002, "request repeated")
	ErrExceedTask       = gfsperrors.Register(module.ManageModularName, http.StatusServiceUnavailable, 60003, "OoooH... request exceed, try again later")
	ErrCanceledTask     = gfsperrors.Register(module.ManageModularName, http.StatusBadRequest, 60004, "task canceled")
	ErrNoTaskMatchLimit = gfsperrors.Register(module.ManageModularName, http.StatusNotFound, 60005, "no task to dispatch below the require limits")
	ErrFutureSupport    = gfsperrors.Register(module.ManageModularName, http.StatusNotFound, 60006, "future support")
)

func (m *ManageModular) DispatchTask(
	ctx context.Context,
	limit rcmgr.Limit) (
	task.Task, error) {
	var (
		backUpTasks []task.Task
		task        task.Task
		mux         sync.Mutex
	)
	mux.Lock()
	defer mux.Unlock()
	task = m.replicateQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add replicate piece task to backup set", "task_key", task.Key().String(),
			"task_limit", "task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.sealQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add seal object task to backup set", "task_key", task.Key().String(),
			"task_limit", "task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.gcObjectQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc object task to backup set", "task_key", task.Key().String(),
			"task_limit", "task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.gcZombieQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc zombie piece task to backup set", "task_key", task.Key().String(),
			"task_limit", "task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.gcMetaQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc meta task to backup set", "task_key", task.Key().String(),
			"task_limit", "task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.receiveQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add confirm receive piece to backup set", "task_key", task.Key().String(),
			"task_limit", "task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.PickUpTask(ctx, backUpTasks)
	if task == nil {
		return nil, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
	log.CtxDebugw(ctx, "success to dispatch task", "require_limit", limit.String(),
		"task_limit", task.EstimateLimit().String())
	return task, nil
}

func (m *ManageModular) QueryTask(
	ctx context.Context,
	key task.TKey) (
	task.Task, error) {
	var qTask task.Task
	if qTask = m.uploadQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	if qTask = m.replicateQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	if qTask = m.sealQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	if qTask = m.receiveQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	if qTask = m.gcObjectQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	if qTask = m.gcZombieQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	if qTask = m.gcMetaQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	if qTask = m.downloadQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	if qTask = m.challengeQueue.PopByKey(key); qTask != nil {
		return qTask, nil
	}
	return nil, nil
}

func (m *ManageModular) HandleCreateUploadObjectTask(
	ctx context.Context,
	task task.UploadObjectTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle begin upload object, task pointer dangling")
		return ErrDanglingTask
	}
	if m.UploadingObjectNumber() >= m.maxUploadObjectNumber {
		log.CtxErrorw(ctx, "uploading object exceed", "uploading", m.uploadQueue.Len(),
			"replicating", m.replicateQueue.Len(), "sealing", m.sealQueue.Len())
		return ErrExceedTask
	}
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "uploading object repeated")
		return ErrRepeatedTask
	}
	if err := m.uploadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push upload object task to queue", "error", err)
		return ErrExceedTask
	}
	_, err := m.baseApp.GfSpDB().CreateUploadJob(task.GetObjectInfo())
	if err != nil {
		log.CtxErrorw(ctx, "failed to create object job", "error", err)
	}
	return nil
}

func (m *ManageModular) HandleDoneUploadObjectTask(
	ctx context.Context,
	task task.UploadObjectTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle done upload object, task pointer dangling")
		return ErrDanglingTask
	}
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle done upload object, task pointer dangling")
		return ErrDanglingTask
	}
	m.uploadQueue.PopByKey(task.Key())
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "uploading object repeated")
		return ErrRepeatedTask
	}
	if task.Error() != nil {
		err := m.baseApp.GfSpDB().UpdateJobState(task.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
		}
		log.CtxErrorw(ctx, "reports failed update object task", "error", task.Error())
		return nil
	}
	replicateTask := &gfsptask.GfSpReplicatePieceTask{}
	replicateTask.InitReplicatePieceTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(task), m.baseApp.TaskTimeout(task), m.baseApp.TaskMaxRetry(task))
	err := m.replicateQueue.Push(replicateTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push replicate piece task to queue", "error", err)
		return ErrExceedTask
	}
	err = m.baseApp.GfSpDB().UpdateJobState(task.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_REPLICATE_OBJECT_DOING)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update object task state", "error", err)
	}
	log.CtxDebugw(ctx, "succeed to done upload object and waiting for scheduling to replicate piece")
	return nil
}

func (m *ManageModular) HandleReplicatePieceTask(
	ctx context.Context,
	task task.ReplicatePieceTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle replicate piece, task pointer dangling")
		return ErrDanglingTask
	}
	if task.Error() != nil {
		log.CtxErrorw(ctx, "handler error replicate piece task", "error", task.Error())
		return m.handleFailedReplicatePieceTask(ctx, task)
	}
	m.replicateQueue.PopByKey(task.Key())
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "replicate piece object task repeated")
		return ErrRepeatedTask
	}
	if task.GetSealed() {
		log.CtxDebugw(ctx, "replicate piece object task has combined seal object task")
		err := m.baseApp.GfSpDB().UpdateJobState(task.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_SEAL_OBJECT_DONE)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
		}
		return nil
	}
	log.CtxDebugw(ctx, "replicate piece object task fails to combine seal object task")
	sealObject := &gfsptask.GfSpSealObjectTask{}
	sealObject.InitSealObjectTask(task.GetObjectInfo(), task.GetStorageParams(), m.baseApp.TaskPriority(task),
		task.GetSecondarySignature(), m.baseApp.TaskTimeout(task), m.baseApp.TaskMaxRetry(task))
	err := m.sealQueue.Push(sealObject)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push seal object task to queue", "error", err)
		return ErrExceedTask
	}
	err = m.baseApp.GfSpDB().UpdateJobState(task.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_SEAL_OBJECT_DOING)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update object task state", "error", err)
	}
	log.CtxDebugw(ctx, "succeed to done replicate piece and waiting for scheduling to seal object")
	return nil
}

func (m *ManageModular) handleFailedReplicatePieceTask(
	ctx context.Context,
	handleTask task.ReplicatePieceTask) error {
	oldTask := m.replicateQueue.PopByKey(handleTask.Key())
	if m.TaskUploading(ctx, handleTask) {
		log.CtxErrorw(ctx, "replicate piece task repeated")
		return ErrRepeatedTask
	}
	if oldTask == nil {
		log.CtxErrorw(ctx, "task has been canceled")
		return ErrCanceledTask
	}
	handleTask = oldTask.(task.ReplicatePieceTask)
	if !handleTask.Expired() {
		handleTask.SetUpdateTime(time.Now().Unix())
		m.replicateQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry")
	} else {
		err := m.baseApp.GfSpDB().UpdateJobState(handleTask.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
		}
		log.CtxWarnw(ctx, "delete expired replicate piece task", "max_retry", handleTask.GetMaxRetry(),
			"retry", handleTask.GetRetry(), "time_out", handleTask.GetTimeout())
	}
	return nil
}

func (m *ManageModular) HandleSealObjectTask(
	ctx context.Context,
	task task.SealObjectTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle seal object, task pointer dangling")
		return ErrDanglingTask
	}
	if task.Error() != nil {
		log.CtxErrorw(ctx, "handler error seal object task", "error", task.Error())
		return m.handleFailedSealObjectTask(ctx, task)
	}
	m.sealQueue.PopByKey(task.Key())
	err := m.baseApp.GfSpDB().UpdateJobState(task.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_SEAL_OBJECT_DONE)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update object task state", "error", err)
	}
	log.CtxDebugw(ctx, "succeed to seal object on chain")
	return nil
}

func (m *ManageModular) handleFailedSealObjectTask(
	ctx context.Context,
	handleTask task.SealObjectTask) error {
	oldTask := m.sealQueue.PopByKey(handleTask.Key())
	if m.TaskUploading(ctx, handleTask) {
		log.CtxErrorw(ctx, "seal object task repeated")
		return ErrRepeatedTask
	}
	if oldTask == nil {
		log.CtxErrorw(ctx, "task has been canceled")
		return ErrCanceledTask
	}
	handleTask = oldTask.(task.SealObjectTask)
	if !handleTask.Expired() {
		handleTask.SetUpdateTime(time.Now().Unix())
		m.sealQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry")
		return nil
	} else {
		err := m.baseApp.GfSpDB().UpdateJobState(handleTask.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
		}
		log.CtxWarnw(ctx, "delete expired seal object task", "max_retry", handleTask.GetMaxRetry(),
			"retry", handleTask.GetRetry(), "time_out", handleTask.GetTimeout())
	}
	return nil
}

func (m *ManageModular) HandleReceivePieceTask(
	ctx context.Context,
	task task.ReceivePieceTask) error {
	if task.Error() != nil {
		return m.handleFailedReceivePieceTask(ctx, task)
	}
	if task.GetSealed() {
		m.receiveQueue.PopByKey(task.Key())
		log.CtxDebugw(ctx, "succeed to confirm receive piece seal on chain")
	} else {
		task.SetRetry(0)
		task.SetMaxRetry(m.baseApp.TaskMaxRetry(task))
		task.SetTimeout(m.baseApp.TaskTimeout(task))
		task.SetPriority(m.baseApp.TaskPriority(task))
		task.SetUpdateTime(time.Now().Unix())
		if err := m.receiveQueue.Push(task); err != nil {
			log.CtxErrorw(ctx, "failed to receive task to queue", "error", err)
		}
	}
	return nil
}

func (m *ManageModular) handleFailedReceivePieceTask(
	ctx context.Context,
	handleTask task.ReceivePieceTask) error {
	oldTask := m.receiveQueue.PopByKey(handleTask.Key())
	if oldTask == nil {
		log.CtxErrorw(ctx, "task has been canceled")
		return ErrCanceledTask
	}
	handleTask = oldTask.(task.ReceivePieceTask)
	if !handleTask.Expired() {
		handleTask.SetUpdateTime(time.Now().Unix())
		m.receiveQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry")
	} else {
		log.CtxErrorw(ctx, "delete expired confirm receive piece task", "max_retry", handleTask.GetMaxRetry(),
			"retry", handleTask.GetRetry(), "time_out", handleTask.GetTimeout())
		return ErrCanceledTask
	}
	return nil
}

func (m *ManageModular) HandleGCObjectTask(
	ctx context.Context,
	task task.GCObjectTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle gc object due to task pointer dangling")
		return ErrDanglingTask
	}
	if !m.gcObjectQueue.Has(task.Key()) {
		return ErrCanceledTask
	}
	if task.GetEndBlockNumber() < task.GetCurrentBlockNumber() {
		log.CtxInfow(ctx, "succeed to gc object task", "end_block_number",
			task.GetEndBlockNumber(), "last_delete_objectId", task.GetLastDeletedObjectId())
		m.gcObjectQueue.PopByKey(task.Key())
		return nil
	}
	task.SetUpdateTime(time.Now().Unix())
	m.gcObjectQueue.PopByKey(task.Key())
	m.gcObjectQueue.Push(task)
	block, object := task.GetGCObjectProgress()
	err := m.baseApp.GfSpDB().SetGCObjectProgress(task.Key().String(), block, object)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update gc object status", "error", err)
	}
	return nil
}

func (m *ManageModular) HandleGCZombiePieceTask(
	ctx context.Context,
	task task.GCZombiePieceTask) error {
	return ErrFutureSupport
}

func (m *ManageModular) HandleGCMetaTask(
	ctx context.Context,
	task task.GCMetaTask) error {
	return ErrFutureSupport
}

func (m *ManageModular) HandleDownloadObjectTask(
	ctx context.Context,
	task task.DownloadObjectTask) error {
	m.downloadQueue.Push(task)
	log.CtxDebugw(ctx, "add download object task to queue")
	return nil
}

func (m *ManageModular) HandleChallengePieceTask(
	ctx context.Context,
	task task.ChallengePieceTask) error {
	m.challengeQueue.Push(task)
	log.CtxDebugw(ctx, "add challenge piece task to queue")
	return nil
}
