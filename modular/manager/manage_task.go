package manager

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/types"
)

var (
	ErrDanglingTask  = gfsperrors.Register(module.ManageModularName, http.StatusBadRequest, 60001, "OoooH... request lost")
	ErrRepeatedTask  = gfsperrors.Register(module.ManageModularName, http.StatusNotAcceptable, 60002, "request repeated")
	ErrExceedTask    = gfsperrors.Register(module.ManageModularName, http.StatusNotAcceptable, 60003, "OoooH... request exceed, try again later")
	ErrCanceledTask  = gfsperrors.Register(module.ManageModularName, http.StatusBadRequest, 60004, "task canceled")
	ErrFutureSupport = gfsperrors.Register(module.ManageModularName, http.StatusNotFound, 60005, "future support")
	ErrGfSpDB        = gfsperrors.Register(module.DownloadModularName, http.StatusInternalServerError, 65201, "server slipped away, try again later")
)

func (m *ManageModular) DispatchTask(ctx context.Context, limit rcmgr.Limit) (task.Task, error) {
	var (
		backupTasks   []task.Task
		reservedTasks []task.Task
		task          task.Task
	)
	task = m.replicateQueue.PopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add replicate piece task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.sealQueue.PopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add seal object task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.gcObjectQueue.PopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc object task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.gcZombieQueue.PopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc zombie piece task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.gcMetaQueue.PopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc meta task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.receiveQueue.PopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add confirm receive piece to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.recoveryQueue.PopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add confirm recovery piece to backup set", "recovery task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task, reservedTasks = m.PickUpTask(ctx, backupTasks)

	go func() {
		if len(reservedTasks) == 0 {
			return
		}
		for _, reservedTask := range reservedTasks {
			switch t := reservedTask.(type) {
			case *gfsptask.GfSpReplicatePieceTask:
				err := m.replicateQueue.Push(t)
				log.Errorw("failed to retry push replicate task to queue after dispatch", "error", err)
			case *gfsptask.GfSpSealObjectTask:
				err := m.sealQueue.Push(t)
				log.Errorw("failed to retry push seal task to queue after dispatch", "error", err)
			case *gfsptask.GfSpReceivePieceTask:
				err := m.receiveQueue.Push(t)
				log.Errorw("failed to retry push receive task to queue after dispatch", "error", err)
			case *gfsptask.GfSpGCObjectTask:
				err := m.gcObjectQueue.Push(t)
				log.Errorw("failed to retry push gc object task to queue after dispatch", "error", err)
			case *gfsptask.GfSpGCZombiePieceTask:
				err := m.gcZombieQueue.Push(t)
				log.Errorw("failed to retry push gc zombie task to queue after dispatch", "error", err)
			case *gfsptask.GfSpGCMetaTask:
				err := m.gcMetaQueue.Push(t)
				log.Errorw("failed to retry push gc meta task to queue after dispatch", "error", err)
			case *gfsptask.GfSpRecoverPieceTask:
				err := m.recoveryQueue.Push(t)
				log.Errorw("failed to retry push recovery task to queue after dispatch", "error", err)
			}
		}
	}()
	if task == nil {
		return nil, nil
	}
	return task, nil
}

func (m *ManageModular) HandleCreateUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle begin upload object due to task pointer dangling")
		return ErrDanglingTask
	}
	if m.UploadingObjectNumber() >= m.maxUploadObjectNumber {
		log.CtxErrorw(ctx, "uploading object exceed", "uploading", m.uploadQueue.Len(),
			"replicating", m.replicateQueue.Len(), "sealing", m.sealQueue.Len())
		return ErrExceedTask
	}
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "uploading object repeated", "task_info", task.Info())
		return ErrRepeatedTask
	}
	if err := m.uploadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push upload object task to queue", "task_info", task.Info(), "error", err)
		return err
	}
	if err := m.baseApp.GfSpDB().InsertUploadProgress(task.GetObjectInfo().Id.Uint64()); err != nil {
		log.CtxErrorw(ctx, "failed to create upload object progress", "task_info", task.Info(), "error", err)
		return ErrGfSpDB
	}
	return nil
}

func (m *ManageModular) HandleDoneUploadObjectTask(ctx context.Context, task task.UploadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle done upload object due to pointer dangling")
		return ErrDanglingTask
	}
	m.uploadQueue.PopByKey(task.Key())
	uploading := m.TaskUploading(ctx, task)
	if uploading {
		log.CtxErrorw(ctx, "uploading object repeated")
		return ErrRepeatedTask
	}
	if task.Error() != nil {
		go func() {
			err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         task.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_UPLOAD_OBJECT_ERROR,
				ErrorDescription: task.Error().Error(),
			})
			if err != nil {
				log.Errorw("failed to update object task state", "task_info", task.Info(), "error", err)
			}
			log.Errorw("reports failed update object task", "task_info", task.Info(), "error", task.Error())
		}()
		metrics.ManagerCounter.WithLabelValues(ManagerFailureUpload).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerFailureUpload).Observe(
			time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
		return nil
	} else {
		metrics.ManagerCounter.WithLabelValues(ManagerSuccessUpload).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerSuccessUpload).Observe(
			time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
	}
	replicateTask := &gfsptask.GfSpReplicatePieceTask{}
	replicateTask.InitReplicatePieceTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(replicateTask),
		m.baseApp.TaskTimeout(replicateTask, task.GetObjectInfo().GetPayloadSize()),
		m.baseApp.TaskMaxRetry(replicateTask))
	replicateTask.SetCreateTime(task.GetCreateTime())
	replicateTask.SetLogs(task.GetLogs())
	replicateTask.AppendLog("manager-create-replicate-task")
	err := m.replicateQueue.Push(replicateTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push replicate piece task to queue", "error", err)
		return err
	}
	go func() {
		err = m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:  task.GetObjectInfo().Id.Uint64(),
			TaskState: types.TaskState_TASK_STATE_REPLICATE_OBJECT_DOING,
		})
		if err != nil {
			log.Errorw("failed to update object task state", "task_info", task.Info(), "error", err)
			return
		}
		log.Debugw("succeed to done upload object and waiting for scheduling to replicate piece", "task_info", task.Info())
	}()
	return nil
}

func (m *ManageModular) HandleCreateResumableUploadObjectTask(ctx context.Context, task task.ResumableUploadObjectTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle begin upload object due to task pointer dangling")
		return ErrDanglingTask
	}
	if m.UploadingObjectNumber() >= m.maxUploadObjectNumber {
		log.CtxErrorw(ctx, "uploading object exceed", "uploading", m.uploadQueue.Len(),
			"replicating", m.replicateQueue.Len(), "sealing", m.sealQueue.Len(), "resumable uploading", m.resumeableUploadQueue.Len())
		return ErrExceedTask
	}
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "uploading object repeated", "task_info", task.Info())
		return ErrRepeatedTask
	}
	if err := m.resumeableUploadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push resumable upload object task to queue", "task_info", task.Info(), "error", err)
		return err
	}
	if err := m.baseApp.GfSpDB().InsertUploadProgress(task.GetObjectInfo().Id.Uint64()); err != nil {
		log.CtxErrorw(ctx, "failed to create resumable upload object progress", "task_info", task.Info(), "error", err)
		// TODO(chris)
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil
		} else {
			return ErrGfSpDB
		}
	}
	return nil
}

func (m *ManageModular) HandleDoneResumableUploadObjectTask(ctx context.Context, task task.ResumableUploadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle done upload object, pointer dangling")
		return ErrDanglingTask
	}
	m.resumeableUploadQueue.PopByKey(task.Key())

	uploading := m.TaskUploading(ctx, task)
	if uploading {
		log.CtxErrorw(ctx, "uploading object repeated")
		return ErrRepeatedTask
	}
	if task.Error() != nil {
		go func() error {
			err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         task.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_UPLOAD_OBJECT_ERROR,
				ErrorDescription: task.Error().Error(),
			})
			if err != nil {
				log.CtxErrorw(ctx, "failed to resumable update object task state", "error", err)
			}
			log.CtxErrorw(ctx, "reports failed resumable update object task", "task_info", task.Info(), "error", task.Error())
			return nil
		}()
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
		return err
	}
	go func() error {
		err = m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:  task.GetObjectInfo().Id.Uint64(),
			TaskState: types.TaskState_TASK_STATE_REPLICATE_OBJECT_DOING,
		})
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
			return ErrGfSpDB
		}
		log.CtxDebugw(ctx, "succeed to done upload object and waiting for scheduling to replicate piece")
		return nil
	}()
	return nil
}

func (m *ManageModular) HandleReplicatePieceTask(ctx context.Context, task task.ReplicatePieceTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle replicate piece due to pointer dangling")
		return ErrDanglingTask
	}
	if task.Error() != nil {
		log.CtxErrorw(ctx, "handler error replicate piece task", "task_info", task.Info(), "error", task.Error())
		_ = m.handleFailedReplicatePieceTask(ctx, task)
		metrics.ManagerCounter.WithLabelValues(ManagerFailureReplicate).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerFailureReplicate).Observe(
			time.Since(time.Unix(task.GetUpdateTime(), 0)).Seconds())
		return nil
	} else {
		metrics.ManagerCounter.WithLabelValues(ManagerSuccessReplicate).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerSuccessReplicate).Observe(
			time.Since(time.Unix(task.GetUpdateTime(), 0)).Seconds())
	}
	m.replicateQueue.PopByKey(task.Key())
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "replicate piece object task repeated")
		return ErrRepeatedTask
	}
	if task.GetSealed() {
		task.AppendLog(fmt.Sprintf("manager-handle-succeed-replicate-task-retry:%d", task.GetRetry()))
		_ = m.baseApp.GfSpDB().InsertPutEvent(task)
		go func() {
			log.Debugw("replicate piece object task has combined seal object task", "task_info", task.Info())
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:  task.GetObjectInfo().Id.Uint64(),
				TaskState: types.TaskState_TASK_STATE_SEAL_OBJECT_DONE,
			}); err != nil {
				log.Errorw("failed to update object task state", "task_info", task.Info(), "error", err)
			}
			log.Errorw("succeed to update object task state", "task_info", task.Info())
			// TODO: delete this upload db record?
		}()
		metrics.ManagerCounter.WithLabelValues(ManagerSuccessReplicateAndSeal).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerSuccessReplicateAndSeal).Observe(
			time.Since(time.Unix(task.GetUpdateTime(), 0)).Seconds())
		return nil
	} else {
		task.AppendLog("manager-handle-succeed-replicate-failed-seal")
		metrics.ManagerCounter.WithLabelValues(ManagerFailureReplicateAndSeal).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerFailureReplicateAndSeal).Observe(
			time.Since(time.Unix(task.GetUpdateTime(), 0)).Seconds())
	}

	log.CtxDebugw(ctx, "replicate piece object task fails to combine seal object task", "task_info", task.Info())
	sealObject := &gfsptask.GfSpSealObjectTask{}
	sealObject.InitSealObjectTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(sealObject), task.GetSecondaryAddresses(), task.GetSecondarySignatures(),
		m.baseApp.TaskTimeout(sealObject, 0), m.baseApp.TaskMaxRetry(sealObject))
	sealObject.SetCreateTime(task.GetCreateTime())
	sealObject.SetLogs(task.GetLogs())
	sealObject.AppendLog("manager-create-seal-task")
	err := m.sealQueue.Push(sealObject)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push seal object task to queue", "task_info", task.Info(), "error", err)
		return err
	}
	go func() {
		if err = m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:            task.GetObjectInfo().Id.Uint64(),
			TaskState:           types.TaskState_TASK_STATE_SEAL_OBJECT_DOING,
			SecondaryAddresses:  task.GetSecondaryAddresses(),
			SecondarySignatures: task.GetSecondarySignatures(),
			ErrorDescription:    "",
		}); err != nil {
			log.Errorw("failed to update object task state", "task_info", task.Info(), "task_info", task.Info(), "error", err)
			return
		}
		log.Debugw("succeed to done replicate piece and waiting for scheduling to seal object", "task_info", task.Info())
	}()
	return nil
}

func (m *ManageModular) handleFailedReplicatePieceTask(ctx context.Context, handleTask task.ReplicatePieceTask) error {
	shadowTask := handleTask
	oldTask := m.replicateQueue.PopByKey(handleTask.Key())
	if m.TaskUploading(ctx, handleTask) {
		log.CtxErrorw(ctx, "replicate piece task repeated", "task_info", handleTask.Info())
		return ErrRepeatedTask
	}
	if oldTask == nil {
		log.CtxErrorw(ctx, "task has been canceled", "task_info", handleTask.Info())
		return ErrCanceledTask
	}
	handleTask = oldTask.(task.ReplicatePieceTask)
	if !handleTask.ExceedRetry() {
		handleTask.AppendLog(fmt.Sprintf("manager-handle-failed-replicate-task-repush:%d", shadowTask.GetRetry()))
		handleTask.AppendLog(shadowTask.GetLogs())
		handleTask.SetUpdateTime(time.Now().Unix())
		err := m.replicateQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry", "task_info", handleTask.Info(), "error", err)
	} else {
		shadowTask.AppendLog(fmt.Sprintf("manager-handle-failed-replicate-task-error:%s-retry:%d", shadowTask.Error().Error(), shadowTask.GetRetry()))
		_ = m.baseApp.GfSpDB().InsertPutEvent(shadowTask)
		metrics.ManagerCounter.WithLabelValues(ManagerCancelReplicate).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerCancelReplicate).Observe(
			time.Since(time.Unix(handleTask.GetCreateTime(), 0)).Seconds())
		go func() {
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         handleTask.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_REPLICATE_OBJECT_ERROR,
				ErrorDescription: "exceed_retry",
			}); err != nil {
				log.Errorw("failed to update object task state", "task_info", handleTask.Info(), "error", err)
				return
			}
			log.Errorw("succeed to update object task state", "task_info", handleTask.Info())
		}()
		log.CtxWarnw(ctx, "delete expired replicate piece task", "task_info", handleTask.Info())
	}
	return nil
}

func (m *ManageModular) HandleSealObjectTask(ctx context.Context, task task.SealObjectTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle seal object due to task pointer dangling")
		return ErrDanglingTask
	}
	if task.Error() != nil {
		log.CtxErrorw(ctx, "handler error seal object task", "task_info", task.Info(), "error", task.Error())
		_ = m.handleFailedSealObjectTask(ctx, task)
		metrics.ManagerCounter.WithLabelValues(ManagerFailureSeal).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerFailureSeal).Observe(
			time.Since(time.Unix(task.GetUpdateTime(), 0)).Seconds())
		return nil
	} else {
		metrics.ManagerCounter.WithLabelValues(ManagerSuccessSeal).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerSuccessSeal).Observe(
			time.Since(time.Unix(task.GetUpdateTime(), 0)).Seconds())
	}
	go func() {
		m.sealQueue.PopByKey(task.Key())
		task.AppendLog(fmt.Sprintf("manager-handle-succeed-seal-task-retry:%d", task.GetRetry()))
		_ = m.baseApp.GfSpDB().InsertPutEvent(task)
		if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:  task.GetObjectInfo().Id.Uint64(),
			TaskState: types.TaskState_TASK_STATE_SEAL_OBJECT_DONE,
		}); err != nil {
			log.Errorw("failed to update object task state", "task_info", task.Info(), "error", err)
			return
		}
		// TODO: delete this upload db record?
		log.Debugw("succeed to seal object on chain", "task_info", task.Info())
	}()
	return nil
}

func (m *ManageModular) handleFailedSealObjectTask(ctx context.Context, handleTask task.SealObjectTask) error {
	shadowTask := handleTask
	oldTask := m.sealQueue.PopByKey(handleTask.Key())
	if m.TaskUploading(ctx, handleTask) {
		log.CtxErrorw(ctx, "seal object task repeated", "task_info", handleTask.Info())
		return ErrRepeatedTask
	}
	if oldTask == nil {
		log.CtxErrorw(ctx, "task has been canceled", "task_info", handleTask.Info())
		return ErrCanceledTask
	}
	handleTask = oldTask.(task.SealObjectTask)
	if !handleTask.ExceedRetry() {
		handleTask.AppendLog(fmt.Sprintf("manager-handle-failed-seal-task-error:%s-repush:%d", shadowTask.Error().Error(), shadowTask.GetRetry()))
		handleTask.AppendLog(shadowTask.GetLogs())
		handleTask.SetUpdateTime(time.Now().Unix())
		err := m.sealQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry", "task_info", handleTask.Info(), "error", err)
		return nil
	} else {
		shadowTask.AppendLog(fmt.Sprintf("manager-handle-failed-seal-task-error:%s-retry:%d", shadowTask.Error().Error(), handleTask.GetRetry()))
		_ = m.baseApp.GfSpDB().InsertPutEvent(shadowTask)
		metrics.ManagerCounter.WithLabelValues(ManagerCancelSeal).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerCancelSeal).Observe(
			time.Since(time.Unix(handleTask.GetCreateTime(), 0)).Seconds())
		go func() {
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         handleTask.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_SEAL_OBJECT_ERROR,
				ErrorDescription: "exceed_retry",
			}); err != nil {
				log.Errorw("failed to update object task state", "task_info", handleTask.Info(), "error", err)
				return
			}
			log.Errorw("succeed to update object task state", "task_info", handleTask.Info())
		}()
		log.CtxWarnw(ctx, "delete expired seal object task", "task_info", handleTask.Info())
	}
	return nil
}

func (m *ManageModular) HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) error {
	if task.GetSealed() {
		go m.receiveQueue.PopByKey(task.Key())
		metrics.ManagerCounter.WithLabelValues(ManagerSuccessConfirmReceive).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerSuccessConfirmReceive).Observe(
			time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
		log.CtxDebugw(ctx, "succeed to confirm receive piece seal on chain")
	} else if task.Error() != nil {
		_ = m.handleFailedReceivePieceTask(ctx, task)
		metrics.ManagerCounter.WithLabelValues(ManagerFailureConfirmReceive).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerFailureConfirmReceive).Observe(
			time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
		return nil
	} else {
		go func() {
			task.SetRetry(0)
			task.SetMaxRetry(m.baseApp.TaskMaxRetry(task))
			task.SetTimeout(m.baseApp.TaskTimeout(task, 0))
			task.SetPriority(m.baseApp.TaskPriority(task))
			task.SetUpdateTime(time.Now().Unix())
			err := m.receiveQueue.Push(task)
			log.CtxErrorw(ctx, "push receive task to queue", "error", err)
		}()
	}
	return nil
}

func (m *ManageModular) handleFailedReceivePieceTask(ctx context.Context, handleTask task.ReceivePieceTask) error {
	oldTask := m.receiveQueue.PopByKey(handleTask.Key())
	if oldTask == nil {
		log.CtxErrorw(ctx, "task has been canceled", "task_info", handleTask.Info())
		return ErrCanceledTask
	}
	handleTask = oldTask.(task.ReceivePieceTask)
	if !handleTask.ExceedRetry() {
		handleTask.SetUpdateTime(time.Now().Unix())
		err := m.receiveQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry", "task_info", handleTask.Info(), "error", err)
	} else {
		log.CtxErrorw(ctx, "delete expired confirm receive piece task", "task_info", handleTask.Info())
		// TODO: confirm it
	}
	return nil
}

func (m *ManageModular) HandleGCObjectTask(ctx context.Context, gcTask task.GCObjectTask) error {
	if gcTask == nil {
		log.CtxErrorw(ctx, "failed to handle gc object due to task pointer dangling")
		return ErrDanglingTask
	}
	if !m.gcObjectQueue.Has(gcTask.Key()) {
		log.CtxErrorw(ctx, "task is not in the gc queue", "task_info", gcTask.Info())
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
	err := m.gcObjectQueue.Push(gcTask)
	log.CtxInfow(ctx, "push gc object task to queue again", "from", oldTask, "to", gcTask, "error", err)
	currentGCBlockID, deletedObjectID := gcTask.GetGCObjectProgress()
	err = m.baseApp.GfSpDB().UpdateGCObjectProgress(&spdb.GCObjectMeta{
		TaskKey:             gcTask.Key().String(),
		CurrentBlockHeight:  currentGCBlockID,
		LastDeletedObjectID: deletedObjectID,
	})
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

func (m *ManageModular) HandleRecoverPieceTask(ctx context.Context, task task.RecoveryPieceTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle recovery piece due to pointer dangling")
		return ErrDanglingTask
	}

	if task.GetRecovered() {
		m.recoveryQueue.PopByKey(task.Key())
		log.CtxErrorw(ctx, "finished recovery", "task_info", task.Info())
		return nil
	}

	if task.Error() != nil {
		log.CtxErrorw(ctx, "handler error recovery piece task", "task_info", task.Info(), "error", task.Error())
		return m.handleFailedRecoverPieceTask(ctx, task)
	}

	if m.TaskRecovering(ctx, task) {
		log.CtxErrorw(ctx, "recovering object repeated", "task_info", task.Info())
		return ErrRepeatedTask
	}

	task.SetUpdateTime(time.Now().Unix())
	if err := m.recoveryQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push recovery object task to queue", "task_info", task.Info(), "error", err)
		return err
	}

	return nil
}

func (m *ManageModular) handleFailedRecoverPieceTask(ctx context.Context, handleTask task.RecoveryPieceTask) error {
	oldTask := m.recoveryQueue.PopByKey(handleTask.Key())
	if oldTask == nil {
		log.CtxErrorw(ctx, "task has been canceled", "task_info", handleTask.Info())
		return ErrCanceledTask
	}
	handleTask = oldTask.(task.RecoveryPieceTask)
	if !handleTask.ExceedRetry() {
		handleTask.SetUpdateTime(time.Now().Unix())
		err := m.recoveryQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry", "task_info", handleTask.Info(), "error", err)
	} else {
		log.CtxErrorw(ctx, "delete expired confirm recovery piece task", "task_info", handleTask.Info())
	}
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
	recoveryTasks, _ := taskqueue.ScanTQueueWithLimitBySubKey(m.recoveryQueue, subKey)

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
	tasks = append(tasks, recoveryTasks...)
	return tasks, nil
}
