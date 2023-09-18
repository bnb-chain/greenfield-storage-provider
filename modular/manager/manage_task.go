package manager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/util"

	"cosmossdk.io/math"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingTask         = gfsperrors.Register(module.ManageModularName, http.StatusBadRequest, 60001, "OoooH... request lost")
	ErrRepeatedTask         = gfsperrors.Register(module.ManageModularName, http.StatusNotAcceptable, 60002, "request repeated")
	ErrExceedTask           = gfsperrors.Register(module.ManageModularName, http.StatusNotAcceptable, 60003, "OoooH... request exceed, try again later")
	ErrCanceledTask         = gfsperrors.Register(module.ManageModularName, http.StatusBadRequest, 60004, "task canceled")
	ErrFutureSupport        = gfsperrors.Register(module.ManageModularName, http.StatusNotFound, 60005, "future support")
	ErrNotifyMigrateSwapOut = gfsperrors.Register(module.ManageModularName, http.StatusNotAcceptable, 60006, "failed to notify swap out start")
)

func ErrGfSpDBWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.ManageModularName, http.StatusInternalServerError, 65201, detail)
}

func (m *ManageModular) DispatchTask(ctx context.Context, limit rcmgr.Limit) (task.Task, error) {
	for {
		select {
		case <-ctx.Done():
			log.CtxErrorw(ctx, "dispatch task context is canceled")
			return nil, nil
		case dispatchTask := <-m.taskCh:
			atomic.AddInt64(&m.backupTaskNum, -1)
			if !limit.NotLess(dispatchTask.EstimateLimit()) {
				log.CtxErrorw(ctx, "resource exceed", "executor_limit", limit.String(), "task_limit", dispatchTask.EstimateLimit().String(), "task_info", dispatchTask.Info())
				go func() {
					m.taskCh <- dispatchTask
					atomic.AddInt64(&m.backupTaskNum, 1)
				}()
				continue
			}
			dispatchTask.IncRetry()
			dispatchTask.SetError(nil)
			dispatchTask.SetUpdateTime(time.Now().Unix())
			dispatchTask.SetAddress(util.GetRPCRemoteAddress(ctx))
			m.repushTask(dispatchTask)
			log.CtxDebugw(ctx, "dispatch task to executor", "key_info", dispatchTask.Info())
			return dispatchTask, nil
		}
	}
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
		return ErrGfSpDBWithDetail("failed to create upload object progress, task_info: " + task.Info() + ", error: " + err.Error())
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
	return m.pickGVGAndReplicate(ctx, task.GetVirtualGroupFamilyId(), task)
}

func (m *ManageModular) pickGVGAndReplicate(ctx context.Context, vgfID uint32, task task.ObjectTask) error {
	startPickGVGTime := time.Now()
	gvgMeta, err := m.pickGlobalVirtualGroup(ctx, vgfID, task.GetStorageParams())
	log.CtxInfow(ctx, "pick global virtual group", "time_cost", time.Since(startPickGVGTime).Seconds(), "gvg_meta", gvgMeta, "error", err)
	if err != nil {
		return err
	}

	replicateTask := &gfsptask.GfSpReplicatePieceTask{}
	replicateTask.InitReplicatePieceTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(replicateTask),
		m.baseApp.TaskTimeout(replicateTask, task.GetObjectInfo().GetPayloadSize()),
		m.baseApp.TaskMaxRetry(replicateTask))
	replicateTask.GlobalVirtualGroupId = gvgMeta.ID
	replicateTask.SecondaryEndpoints = gvgMeta.SecondarySPEndpoints
	log.Debugw("replicate task info", "task", replicateTask, "gvg_meta", gvgMeta)
	replicateTask.SetCreateTime(task.GetCreateTime())
	replicateTask.SetLogs(task.GetLogs())
	replicateTask.AppendLog("manager-create-replicate-task")
	err = m.replicateQueue.Push(replicateTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push replicate piece task to queue", "error", err)
		return err
	}
	go m.backUpTask()
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
			"replicating", m.replicateQueue.Len(), "sealing", m.sealQueue.Len(), "resumable uploading", m.resumableUploadQueue.Len())
		return ErrExceedTask
	}
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "uploading object repeated", "task_info", task.Info())
		return ErrRepeatedTask
	}
	if err := m.resumableUploadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push resumable upload object task to queue", "task_info", task.Info(), "error", err)
		return err
	}
	if err := m.baseApp.GfSpDB().InsertUploadProgress(task.GetObjectInfo().Id.Uint64()); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil
		} else {
			log.CtxErrorw(ctx, "failed to create resumable upload object progress", "task_info", task.Info(), "error", err)
			return ErrGfSpDBWithDetail("failed to create resumable upload object progress, task_info: " + task.Info() + ", error: " + err.Error())
		}
	}
	return nil
}

func (m *ManageModular) HandleDoneResumableUploadObjectTask(ctx context.Context, task task.ResumableUploadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to handle done upload object, pointer dangling")
		return ErrDanglingTask
	}
	m.resumableUploadQueue.PopByKey(task.Key())

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
		metrics.ManagerCounter.WithLabelValues(ManagerFailureUpload).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerFailureUpload).Observe(
			time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
		return nil
	} else {
		metrics.ManagerCounter.WithLabelValues(ManagerSuccessUpload).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerSuccessUpload).Observe(
			time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
	}

	// During a resumable upload, the uploader reports each uploaded segment to the manager.
	// Once all segments are reported as completed, the replication process can begin.
	if !task.GetCompleted() {
		return nil
	}

	startPickGVGTime := time.Now()
	gvgMeta, err := m.pickGlobalVirtualGroup(ctx, task.GetVirtualGroupFamilyId(), task.GetStorageParams())
	if err != nil {
		log.CtxErrorw(ctx, "failed to pick global virtual group", "time_cost", time.Since(startPickGVGTime).Seconds(), "error", err)
		return err
	}

	replicateTask := &gfsptask.GfSpReplicatePieceTask{}
	replicateTask.InitReplicatePieceTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(replicateTask),
		m.baseApp.TaskTimeout(replicateTask, task.GetObjectInfo().GetPayloadSize()),
		m.baseApp.TaskMaxRetry(replicateTask))
	replicateTask.GlobalVirtualGroupId = gvgMeta.ID
	replicateTask.SecondaryEndpoints = gvgMeta.SecondarySPEndpoints

	err = m.replicateQueue.Push(replicateTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push replicate piece task to queue", "error", err)
		return err
	}
	go m.backUpTask()
	go func() error {
		err = m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:  task.GetObjectInfo().Id.Uint64(),
			TaskState: types.TaskState_TASK_STATE_REPLICATE_OBJECT_DOING,
		})
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
			return ErrGfSpDBWithDetail("failed to update object task state, error: " + err.Error())
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
		go func() {
			_ = m.baseApp.GfSpDB().InsertPutEvent(task)
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
	sealObject.InitSealObjectTask(task.GetGlobalVirtualGroupId(), task.GetObjectInfo(), task.GetStorageParams(),
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
	go m.backUpTask()
	go func() {
		if err = m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:             task.GetObjectInfo().Id.Uint64(),
			TaskState:            types.TaskState_TASK_STATE_SEAL_OBJECT_DOING,
			GlobalVirtualGroupID: task.GetGlobalVirtualGroupId(),
			SecondaryEndpoints:   task.GetSecondaryEndpoints(),
			SecondarySignatures:  task.GetSecondarySignatures(),
			ErrorDescription:     "",
		}); err != nil {
			log.Errorw("failed to update object task state", "task_info", task.Info(), "task_info", task.Info(), "error", err)
			return
		}
		log.Debugw("succeed to done replicate piece and waiting for scheduling to seal object", "task_info", task.Info())
	}()
	return nil
}

func (m *ManageModular) handleFailedReplicatePieceTask(ctx context.Context, handleTask task.ReplicatePieceTask) error {
	if handleTask.GetNotAvailableSpIdx() != -1 {
		gvgID := handleTask.GetGlobalVirtualGroupId()
		gvg, err := m.baseApp.Consensus().QueryGlobalVirtualGroup(context.Background(), gvgID)
		if err != nil {
			log.Errorw("failed to query global virtual group from chain, ", "gvgID", gvgID, "error", err)
			return err
		}
		sspID := gvg.GetSecondarySpIds()[handleTask.GetNotAvailableSpIdx()]
		sspJoinGVGs, err := m.baseApp.GfSpClient().ListGlobalVirtualGroupsBySecondarySP(ctx, sspID)
		if err != nil {
			log.Errorw("failed to list GVGs by secondary sp", "spID", sspID, "error", err)
			return err
		}
		shouldFreezeGVGs := make([]*virtualgrouptypes.GlobalVirtualGroup, 0)
		selfSPID, err := m.getSPID()
		if err != nil {
			log.CtxErrorw(ctx, "failed to get self sp id", "error", err)
			return err
		}
		for _, g := range sspJoinGVGs {
			if g.GetPrimarySpId() == selfSPID {
				shouldFreezeGVGs = append(shouldFreezeGVGs, g)
			}
		}
		m.virtualGroupManager.FreezeSPAndGVGs(sspID, shouldFreezeGVGs)
		log.CtxDebugw(ctx, "add sp to freeze pool", "spID", sspID, "excludedGVGs", shouldFreezeGVGs)
		m.replicateQueue.PopByKey(handleTask.Key())

		return m.pickGVGAndReplicate(ctx, gvg.FamilyId, handleTask)
	}

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
		metrics.ManagerCounter.WithLabelValues(ManagerCancelReplicate).Inc()
		metrics.ManagerTime.WithLabelValues(ManagerCancelReplicate).Observe(
			time.Since(time.Unix(handleTask.GetCreateTime(), 0)).Seconds())
		go func() {
			_ = m.baseApp.GfSpDB().InsertPutEvent(shadowTask)
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         handleTask.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_REPLICATE_OBJECT_ERROR,
				ErrorDescription: "exceed_replicate_retry",
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
				ErrorDescription: "exceed_seal_retry",
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
			if err == nil {
				go m.backUpTask()
			}
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

func (m *ManageModular) HandleMigrateGVGTask(ctx context.Context, task task.MigrateGVGTask) error {
	if task == nil {
		log.CtxErrorw(ctx, "failed to handle migrate gvg due to pointer dangling")
		return ErrDanglingTask
	}
	var err, pushErr error
	cancelTask := false

	if task.GetBucketID() != 0 {
		// if there is no execute plan, we should cancel this task
		if _, err = m.bucketMigrateScheduler.getExecutePlanByBucketID(task.GetBucketID()); err != nil {
			cancelTask = true
		}
	}

	pushErr = m.migrateGVGQueuePopByLimitAndPushAgain(task, !cancelTask)
	if pushErr != nil {
		log.CtxErrorw(ctx, "failed to push task to migrate gvg queue", "task", task, "error", pushErr)
		return pushErr
	}

	if task.GetBucketID() != 0 {
		err = m.bucketMigrateScheduler.UpdateMigrateProgress(task)
	} else {
		err = m.spExitScheduler.UpdateMigrateProgress(task)
	}

	log.CtxInfow(ctx, "succeed to handle migrate gvg task", "task", task, "error", err)
	return err
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
	migrateGVGTasks, _ := taskqueue.ScanTQueueWithLimitBySubKey(m.migrateGVGQueue, subKey)

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
	tasks = append(tasks, migrateGVGTasks...)
	return tasks, nil
}

func (m *ManageModular) QueryBucketMigrate(ctx context.Context) (res *gfspserver.GfSpQueryBucketMigrateResponse, err error) {
	if m.bucketMigrateScheduler != nil {
		res, err = m.bucketMigrateScheduler.listExecutePlan()
	} else {
		res, err = nil, errors.New("bucketMigrateScheduler not exit")
	}

	return res, err
}

func (m *ManageModular) QuerySpExit(ctx context.Context) (res *gfspserver.GfSpQuerySpExitResponse, err error) {
	if m.spExitScheduler != nil {
		res, err = m.spExitScheduler.ListSPExitPlan()
	} else {
		res, err = nil, errors.New("spExitScheduler not exit")
	}

	return res, err
}

// PickVirtualGroupFamily is used to pick a suitable vgf for creating bucket.
func (m *ManageModular) PickVirtualGroupFamily(ctx context.Context, task task.ApprovalCreateBucketTask) (uint32, error) {
	var (
		err error
		vgf *vgmgr.VirtualGroupFamilyMeta
	)

	if vgf, err = m.virtualGroupManager.PickVirtualGroupFamily(); err != nil {
		// create a new gvg, and retry pick.
		if err = m.createGlobalVirtualGroup(0, nil); err != nil {
			log.CtxErrorw(ctx, "failed to create global virtual group", "task_info", task.Info(), "error", err)
			return 0, err
		}
		m.virtualGroupManager.ForceRefreshMeta()
		if vgf, err = m.virtualGroupManager.PickVirtualGroupFamily(); err != nil {
			log.CtxErrorw(ctx, "failed to pick vgf", "task_info", task.Info(), "error", err)
			return 0, err
		}
		return vgf.ID, nil
	}
	return vgf.ID, nil
}

var _ vgmgr.GenerateGVGSecondarySPsPolicy = &GenerateGVGSecondarySPsPolicyByPrefer{}

type GenerateGVGSecondarySPsPolicyByPrefer struct {
	expectedSecondarySPNumber int
	preferSPIDMap             map[uint32]bool
	preferSPIDList            []uint32
	backupSPIDList            []uint32
}

func NewGenerateGVGSecondarySPsPolicyByPrefer(p *storagetypes.Params, preferSPIDList []uint32) *GenerateGVGSecondarySPsPolicyByPrefer {
	policy := &GenerateGVGSecondarySPsPolicyByPrefer{
		expectedSecondarySPNumber: int(p.GetRedundantDataChunkNum() + p.GetRedundantParityChunkNum()),
		preferSPIDMap:             make(map[uint32]bool),
		preferSPIDList:            make([]uint32, 0),
		backupSPIDList:            make([]uint32, 0),
	}
	for _, spID := range preferSPIDList {
		policy.preferSPIDMap[spID] = true
	}
	return policy
}

func (p *GenerateGVGSecondarySPsPolicyByPrefer) AddCandidateSP(spID uint32) {
	if _, found := p.preferSPIDMap[spID]; found {
		p.preferSPIDList = append(p.preferSPIDList, spID)
	} else {
		p.backupSPIDList = append(p.backupSPIDList, spID)
	}

}
func (p *GenerateGVGSecondarySPsPolicyByPrefer) GenerateGVGSecondarySPs() ([]uint32, error) {
	if p.expectedSecondarySPNumber > len(p.preferSPIDList)+len(p.backupSPIDList) {
		return nil, fmt.Errorf("no enough sp")
	}
	resultSPList := make([]uint32, 0)
	resultSPList = append(resultSPList, p.preferSPIDList...)
	resultSPList = append(resultSPList, p.backupSPIDList...)
	return resultSPList[0:p.expectedSecondarySPNumber], nil
}

func (m *ManageModular) createGlobalVirtualGroup(vgfID uint32, params *storagetypes.Params) error {
	var err error
	if params == nil {
		if params, err = m.baseApp.Consensus().QueryStorageParamsByTimestamp(context.Background(), time.Now().Unix()); err != nil {
			return err
		}
	}
	gvgMeta, err := m.virtualGroupManager.GenerateGlobalVirtualGroupMeta(NewGenerateGVGSecondarySPsPolicyByPrefer(params, m.gvgPreferSPList))
	if err != nil {
		return err
	}
	log.Infow("begin to create a gvg", "gvg_meta", gvgMeta)
	virtualGroupParams, err := m.baseApp.Consensus().QueryVirtualGroupParams(context.Background())
	if err != nil {
		return err
	}
	return m.baseApp.GfSpClient().CreateGlobalVirtualGroup(context.Background(), &gfspserver.GfSpCreateGlobalVirtualGroup{
		VirtualGroupFamilyId: vgfID,
		PrimarySpAddress:     m.baseApp.OperatorAddress(), // it is useless
		SecondarySpIds:       gvgMeta.SecondarySPIDs,
		Deposit: &sdk.Coin{
			Denom:  virtualGroupParams.GetDepositDenom(),
			Amount: virtualGroupParams.GvgStakingPerBytes.Mul(math.NewIntFromUint64(gvgMeta.StakingStorageSize)),
		},
	})
}

// pickGlobalVirtualGroup is used to pick a suitable gvg for replicating object.
func (m *ManageModular) pickGlobalVirtualGroup(ctx context.Context, vgfID uint32, param *storagetypes.Params) (*vgmgr.GlobalVirtualGroupMeta, error) {
	var (
		err error
		gvg *vgmgr.GlobalVirtualGroupMeta
	)

	if gvg, err = m.virtualGroupManager.PickGlobalVirtualGroup(vgfID); err != nil {
		// create a new gvg, and retry pick.
		if err = m.createGlobalVirtualGroup(vgfID, param); err != nil {
			log.CtxErrorw(ctx, "failed to create global virtual group", "vgf_id", vgfID, "error", err)
			return gvg, err
		}
		m.virtualGroupManager.ForceRefreshMeta()
		if gvg, err = m.virtualGroupManager.PickGlobalVirtualGroup(vgfID); err != nil {
			log.CtxErrorw(ctx, "failed to pick gvg", "vgf_id", vgfID, "error", err)
			return gvg, err
		}
		return gvg, nil
	}
	log.CtxDebugw(ctx, "succeed to pick gvg", "gvg", gvg)
	return gvg, nil
}
