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
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/types"
)

var (
	ErrDanglingTask  = gfsperrors.Register(module.ManageModularName, http.StatusBadRequest, 60001, "OoooH... request lost")
	ErrRepeatedTask  = gfsperrors.Register(module.ManageModularName, http.StatusNotAcceptable, 60002, "request repeated")
	ErrExceedTask    = gfsperrors.Register(module.ManageModularName, http.StatusNotAcceptable, 60003, "OoooH... request exceed, try again later")
	ErrCanceledTask  = gfsperrors.Register(module.ManageModularName, http.StatusBadRequest, 60004, "task canceled")
	ErrFutureSupport = gfsperrors.Register(module.ManageModularName, http.StatusNotFound, 60005, "future support")
)

func (m *ManageModular) DispatchTask(ctx context.Context, limit rcmgr.Limit) (task.Task, error) {
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
			"task_limit", task.EstimateLimit().String())
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
			"task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.gcZombieQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc zombie piece task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.gcMetaQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc meta task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.receiveQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add confirm receive piece to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backUpTasks = append(backUpTasks, task)
	}
	task = m.PickUpTask(ctx, backUpTasks)
	if task == nil {
		return nil, nil
	}
	return task, nil
}

func (m *ManageModular) HandleCreateUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error {
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

func (m *ManageModular) HandleDoneUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle done upload object, pointer dangling")
		return ErrDanglingTask
	}
	m.uploadQueue.PopByKey(task.Key())
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "uploading object repeated")
		return ErrRepeatedTask
	}
	if task.Error() != nil {
		err := m.baseApp.GfSpDB().UpdateJobState(
			task.GetObjectInfo().Id.Uint64(),
			types.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
		}
		log.CtxErrorw(ctx, "reports failed update object task", "error", task.Error())
		return nil
	}
	replicateTask := &gfsptask.GfSpReplicatePieceTask{}
	replicateTask.InitReplicatePieceTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(replicateTask),
		m.baseApp.TaskTimeout(replicateTask, task.GetObjectInfo().GetPayloadSize()),
		m.baseApp.TaskMaxRetry(replicateTask))
	err := m.replicateQueue.Push(replicateTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push replicate piece task to queue", "error", err)
		return ErrExceedTask
	}
	err = m.baseApp.GfSpDB().UpdateJobState(
		task.GetObjectInfo().Id.Uint64(),
		types.JobState_JOB_STATE_REPLICATE_OBJECT_DOING)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update object task state", "error", err)
	}
	log.CtxDebugw(ctx, "succeed to done upload object and waiting for scheduling to replicate piece")
	return nil
}

func (m *ManageModular) HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle replicate piece, pointer dangling")
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
		err := m.baseApp.GfSpDB().UpdateJobState(
			task.GetObjectInfo().Id.Uint64(),
			types.JobState_JOB_STATE_SEAL_OBJECT_DONE)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
		}
		return nil
	}
	log.CtxDebugw(ctx, "replicate piece object task fails to combine seal object task")
	sealObject := &gfsptask.GfSpSealObjectTask{}
	sealObject.InitSealObjectTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(sealObject), task.GetSecondarySignature(),
		m.baseApp.TaskTimeout(sealObject, 0), m.baseApp.TaskMaxRetry(sealObject))
	err := m.sealQueue.Push(sealObject)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push seal object task to queue", "error", err)
		return ErrExceedTask
	}
	err = m.baseApp.GfSpDB().UpdateJobState(
		task.GetObjectInfo().Id.Uint64(),
		types.JobState_JOB_STATE_SEAL_OBJECT_DOING)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update object task state", "error", err)
	}
	log.CtxDebugw(ctx, "succeed to done replicate piece and waiting for scheduling to seal object")
	return nil
}

func (m *ManageModular) handleFailedReplicatePieceTask(ctx context.Context, handleTask task.ReplicatePieceTask) error {
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
	if !handleTask.ExceedRetry() {
		handleTask.SetUpdateTime(time.Now().Unix())
		m.replicateQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry", "info", handleTask.Info())
	} else {
		err := m.baseApp.GfSpDB().UpdateJobState(
			handleTask.GetObjectInfo().Id.Uint64(),
			types.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
		}
		log.CtxWarnw(ctx, "delete expired replicate piece task", "info", handleTask.Info())
	}
	return nil
}

func (m *ManageModular) HandleSealObjectTask(ctx context.Context, task task.SealObjectTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle seal object, task pointer dangling")
		return ErrDanglingTask
	}
	if task.Error() != nil {
		log.CtxErrorw(ctx, "handler error seal object task", "error", task.Error())
		return m.handleFailedSealObjectTask(ctx, task)
	}
	m.sealQueue.PopByKey(task.Key())
	err := m.baseApp.GfSpDB().UpdateJobState(
		task.GetObjectInfo().Id.Uint64(),
		types.JobState_JOB_STATE_SEAL_OBJECT_DONE)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update object task state", "error", err)
	}
	log.CtxDebugw(ctx, "succeed to seal object on chain")
	return nil
}

func (m *ManageModular) handleFailedSealObjectTask(ctx context.Context, handleTask task.SealObjectTask) error {
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
	if !handleTask.ExceedRetry() {
		handleTask.SetUpdateTime(time.Now().Unix())
		m.sealQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry")
		return nil
	} else {
		err := m.baseApp.GfSpDB().UpdateJobState(
			handleTask.GetObjectInfo().Id.Uint64(),
			types.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
		}
		log.CtxWarnw(ctx, "delete expired seal object task", "info", handleTask.Info())
	}
	return nil
}

func (m *ManageModular) HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) error {
	if task.GetSealed() {
		m.receiveQueue.PopByKey(task.Key())
		log.CtxDebugw(ctx, "succeed to confirm receive piece seal on chain")
	} else if task.Error() != nil {
		return m.handleFailedReceivePieceTask(ctx, task)
	} else {
		task.SetRetry(0)
		task.SetMaxRetry(m.baseApp.TaskMaxRetry(task))
		task.SetTimeout(m.baseApp.TaskTimeout(task, 0))
		task.SetPriority(m.baseApp.TaskPriority(task))
		task.SetUpdateTime(time.Now().Unix())
		if err := m.receiveQueue.Push(task); err != nil {
			log.CtxErrorw(ctx, "failed to receive task to queue", "error", err)
		}
	}
	return nil
}

func (m *ManageModular) handleFailedReceivePieceTask(ctx context.Context, handleTask task.ReceivePieceTask) error {
	oldTask := m.receiveQueue.PopByKey(handleTask.Key())
	if oldTask == nil {
		log.CtxErrorw(ctx, "task has been canceled")
		return ErrCanceledTask
	}
	handleTask = oldTask.(task.ReceivePieceTask)
	if !handleTask.ExceedRetry() {
		handleTask.SetUpdateTime(time.Now().Unix())
		m.receiveQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry")
	} else {
		log.CtxErrorw(ctx, "delete expired confirm receive piece task", "info", handleTask.Info())
	}
	return nil
}

func (m *ManageModular) HandleGCObjectTask(ctx context.Context, gcTask task.GCObjectTask) error {
	if gcTask == nil {
		log.CtxErrorw(ctx, "failed to handle gc object due to task pointer dangling")
		return ErrDanglingTask
	}
	if !m.gcObjectQueue.Has(gcTask.Key()) {
		return ErrCanceledTask
	}
	if gcTask.GetCurrentBlockNumber() > gcTask.GetEndBlockNumber() {
		log.CtxInfow(ctx, "succeed to finish the gc object task", "task_info", gcTask.Info())
		m.gcObjectQueue.PopByKey(gcTask.Key())
		m.baseApp.GfSpDB().DeleteGCObjectProgress(gcTask.Key().String())
		return nil
	}
	gcTask.SetUpdateTime(time.Now().Unix())
	oldTask := m.gcObjectQueue.PopByKey(gcTask.Key())
	if oldTask != nil {
		if oldTask.(task.GCObjectTask).GetCurrentBlockNumber() > gcTask.GetCurrentBlockNumber() ||
			(oldTask.(task.GCObjectTask).GetCurrentBlockNumber() == gcTask.GetCurrentBlockNumber() &&
				oldTask.(task.GCObjectTask).GetLastDeletedObjectId() > gcTask.GetLastDeletedObjectId()) {
			log.CtxErrorw(ctx, "the reported gc object task is expired", "report_info", gcTask.Info(),
				"current_info", oldTask.Info())
			return ErrCanceledTask
		}
	} else {
		log.CtxErrorw(ctx, "the reported gc object task is canceled", "report_info", gcTask.Info())
		return ErrCanceledTask
	}
	m.gcObjectQueue.Push(gcTask)
	currentGCBlockID, deletedObjectID := gcTask.GetGCObjectProgress()
	err := m.baseApp.GfSpDB().SetGCObjectProgress(gcTask.Key().String(), currentGCBlockID, deletedObjectID)
	log.CtxInfow(ctx, "update the gc object task progress", "from", oldTask, "to", gcTask, "error", err)
	return nil
}

func (m *ManageModular) HandleGCZombiePieceTask(ctx context.Context, task task.GCZombiePieceTask) error {
	return ErrFutureSupport
}

func (m *ManageModular) HandleGCMetaTask(ctx context.Context, task task.GCMetaTask) error {
	return ErrFutureSupport
}

func (m *ManageModular) HandleDownloadObjectTask(ctx context.Context, task task.DownloadObjectTask) error {
	m.downloadQueue.Push(task)
	log.CtxDebugw(ctx, "add download object task to queue")
	return nil
}

func (m *ManageModular) HandleChallengePieceTask(ctx context.Context, task task.ChallengePieceTask) error {
	m.challengeQueue.Push(task)
	log.CtxDebugw(ctx, "add challenge piece task to queue")
	return nil
}

func (m *ManageModular) QueryTasks(ctx context.Context, subKey task.TKey) ([]task.Task, error) {
	uploadTasks, _ := taskqueue.ScanTQueueBySubKey(m.uploadQueue, subKey)
	replicateTasks, _ := taskqueue.ScanTQueueWithLimitBySubKey(m.replicateQueue, subKey)
	sealTasks, _ := taskqueue.ScanTQueueWithLimitBySubKey(m.sealQueue, subKey)
	receiveTasks, _ := taskqueue.ScanTQueueWithLimitBySubKey(m.receiveQueue, subKey)
	gcObjectTasks, _ := taskqueue.ScanTQueueWithLimitBySubKey(m.gcObjectQueue, subKey)
	gcZombieTasks, _ := taskqueue.ScanTQueueWithLimitBySubKey(m.gcZombieQueue, subKey)
	gcMetaTasks, _ := taskqueue.ScanTQueueWithLimitBySubKey(m.gcMetaQueue, subKey)
	downloadTasks, _ := taskqueue.ScanTQueueBySubKey(m.downloadQueue, subKey)
	challengeTasks, _ := taskqueue.ScanTQueueBySubKey(m.challengeQueue, subKey)

	var tasks []task.Task
	tasks = append(tasks, uploadTasks...)
	tasks = append(tasks, replicateTasks...)
	tasks = append(tasks, receiveTasks...)
	tasks = append(tasks, sealTasks...)
	tasks = append(tasks, gcObjectTasks...)
	tasks = append(tasks, gcZombieTasks...)
	tasks = append(tasks, gcMetaTasks...)
	tasks = append(tasks, downloadTasks...)
	tasks = append(tasks, challengeTasks...)
	return tasks, nil
}
