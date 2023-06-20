package manager

import (
	"context"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/core/vmmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		backupTasks []task.Task
		task        task.Task
	)
	m.mux.Lock()
	defer m.mux.Unlock()
	task = m.replicateQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add replicate piece task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.sealQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add seal object task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.gcObjectQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc object task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.gcZombieQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc zombie piece task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.gcMetaQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add gc meta task to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.receiveQueue.TopByLimit(limit)
	if task != nil {
		log.CtxDebugw(ctx, "add confirm receive piece to backup set", "task_key", task.Key().String(),
			"task_limit", task.EstimateLimit().String())
		backupTasks = append(backupTasks, task)
	}
	task = m.PickUpTask(ctx, backupTasks)
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

	startCheckUploadingTime := time.Now()
	uploading := m.TaskUploading(ctx, task)
	metrics.PerfUploadTimeHistogram.WithLabelValues("report_upload_task_check_uploading").
		Observe(time.Since(startCheckUploadingTime).Seconds())
	if uploading {
		log.CtxErrorw(ctx, "uploading object repeated")
		return ErrRepeatedTask
	}
	if task.Error() != nil {
		go func() error {
			startUpdateSPDBTime := time.Now()
			err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         task.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_UPLOAD_OBJECT_ERROR,
				ErrorDescription: task.Error().Error(),
			})
			metrics.PerfUploadTimeHistogram.WithLabelValues("report_upload_task_update_spdb").
				Observe(time.Since(startUpdateSPDBTime).Seconds())
			if err != nil {
				log.CtxErrorw(ctx, "failed to update object task state", "error", err)
				return ErrGfSpDB
			}
			log.CtxErrorw(ctx, "reports failed update object task", "task_info", task.Info(), "error", task.Error())
			return nil
		}()
		return nil
	}
	// TODO: refine it.
	startPickGVGTime := time.Now()
	gvgMeta, err := m.pickGlobalVirtualGroup(ctx, task.GetVirtualGroupFamilyId(), task.GetStorageParams())
	metrics.PerfUploadTimeHistogram.WithLabelValues("pick_global_virtual_group").
		Observe(time.Since(startPickGVGTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to pick global virtual group", "error", err)
		return err
	}

	replicateTask := &gfsptask.GfSpReplicatePieceTask{}
	replicateTask.InitReplicatePieceTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(replicateTask),
		m.baseApp.TaskTimeout(replicateTask, task.GetObjectInfo().GetPayloadSize()),
		m.baseApp.TaskMaxRetry(replicateTask))
	replicateTask.GlobalVirtualGroupId = task.GetVirtualGroupFamilyId()
	replicateTask.SecondarySps = gvgMeta.SecondarySPs

	startPushReplicateQueueTime := time.Now()
	err = m.replicateQueue.Push(replicateTask)
	metrics.PerfUploadTimeHistogram.WithLabelValues("report_upload_task_push_replicate_queue").
		Observe(time.Since(startPushReplicateQueueTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to push replicate piece task to queue", "error", err)
		return err
	}
	go func() error {
		startUpdateSPDBTime := time.Now()
		err = m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:  task.GetObjectInfo().Id.Uint64(),
			TaskState: types.TaskState_TASK_STATE_REPLICATE_OBJECT_DOING,
		})
		metrics.PerfUploadTimeHistogram.WithLabelValues("report_upload_task_update_spdb").
			Observe(time.Since(startUpdateSPDBTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "error", err)
			return ErrGfSpDB
		}
		log.CtxDebugw(ctx, "succeed to done upload object and waiting for scheduling to replicate piece", "task_info", task.Info())
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
		go m.handleFailedReplicatePieceTask(ctx, task)
		return nil
	}
	m.replicateQueue.PopByKey(task.Key())
	if m.TaskUploading(ctx, task) {
		log.CtxErrorw(ctx, "replicate piece object task repeated")
		return ErrRepeatedTask
	}
	if task.GetSealed() {
		go func() error {
			metrics.SealObjectSucceedCounter.WithLabelValues(m.Name()).Inc()
			log.CtxDebugw(ctx, "replicate piece object task has combined seal object task")
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:  task.GetObjectInfo().Id.Uint64(),
				TaskState: types.TaskState_TASK_STATE_SEAL_OBJECT_DONE,
			}); err != nil {
				log.CtxErrorw(ctx, "failed to update object task state", "task_info", task.Info(), "error", err)
				// succeed, ignore this error
				// return ErrGfSpDB
			}
			// TODO: delete this upload db record?
			return nil
		}()
		return nil
	}
	log.CtxDebugw(ctx, "replicate piece object task fails to combine seal object task", "task_info", task.Info())
	sealObject := &gfsptask.GfSpSealObjectTask{}
	sealObject.InitSealObjectTask(task.GetObjectInfo(), task.GetStorageParams(),
		m.baseApp.TaskPriority(sealObject), task.GetSecondaryAddresses(), task.GetSecondarySignatures(),
		m.baseApp.TaskTimeout(sealObject, 0), m.baseApp.TaskMaxRetry(sealObject))
	err := m.sealQueue.Push(sealObject)
	if err != nil {
		log.CtxErrorw(ctx, "failed to push seal object task to queue", "task_info", task.Info(), "error", err)
		return err
	}
	go func() error {
		if err = m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:            task.GetObjectInfo().Id.Uint64(),
			TaskState:           types.TaskState_TASK_STATE_SEAL_OBJECT_DOING,
			SecondaryAddresses:  task.GetSecondaryAddresses(),
			SecondarySignatures: task.GetSecondarySignatures(),
			ErrorDescription:    "",
		}); err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "task_info", task.Info(), "error", err)
			return ErrGfSpDB
		}
		log.CtxDebugw(ctx, "succeed to done replicate piece and waiting for scheduling to seal object", "task_info", task.Info())
		return nil
	}()
	return nil
}

func (m *ManageModular) handleFailedReplicatePieceTask(ctx context.Context, handleTask task.ReplicatePieceTask) error {
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
		handleTask.SetUpdateTime(time.Now().Unix())
		err := m.replicateQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry", "task_info", handleTask.Info(), "error", err)
	} else {
		if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:         handleTask.GetObjectInfo().Id.Uint64(),
			TaskState:        types.TaskState_TASK_STATE_REPLICATE_OBJECT_ERROR,
			ErrorDescription: "exceed_retry",
		}); err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "task_info", handleTask.Info(), "error", err)
			return ErrGfSpDB
		}
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
		go m.handleFailedSealObjectTask(ctx, task)
		return nil
	}
	go func() error {
		metrics.SealObjectSucceedCounter.WithLabelValues(m.Name()).Inc()
		m.sealQueue.PopByKey(task.Key())
		if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:  task.GetObjectInfo().Id.Uint64(),
			TaskState: types.TaskState_TASK_STATE_SEAL_OBJECT_DONE,
		}); err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "task_info", task.Info(), "error", err)
			// succeed, ignore this error
			// return ErrGfSpDB
		}
		// TODO: delete this upload db record?
		log.CtxDebugw(ctx, "succeed to seal object on chain", "task_info", task.Info())
		return nil
	}()
	return nil
}

func (m *ManageModular) handleFailedSealObjectTask(ctx context.Context, handleTask task.SealObjectTask) error {
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
		handleTask.SetUpdateTime(time.Now().Unix())
		err := m.sealQueue.Push(handleTask)
		log.CtxDebugw(ctx, "push task again to retry", "task_info", handleTask.Info(), "error", err)
		return nil
	} else {
		if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:         handleTask.GetObjectInfo().Id.Uint64(),
			TaskState:        types.TaskState_TASK_STATE_SEAL_OBJECT_ERROR,
			ErrorDescription: "exceed_retry",
		}); err != nil {
			log.CtxErrorw(ctx, "failed to update object task state", "task_info", handleTask.Info(), "error", err)
		}
		log.CtxWarnw(ctx, "delete expired seal object task", "task_info", handleTask.Info())
	}
	return nil
}

func (m *ManageModular) HandleReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) error {
	if task.GetSealed() {
		go m.receiveQueue.PopByKey(task.Key())
		log.CtxDebugw(ctx, "succeed to confirm receive piece seal on chain")
	} else if task.Error() != nil {
		go m.handleFailedReceivePieceTask(ctx, task)
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

// PickVirtualGroupFamily is used to pick a suitable vgf for creating bucket.
func (m *ManageModular) PickVirtualGroupFamily(ctx context.Context, task task.ApprovalCreateBucketTask) (uint32, error) {
	var (
		err error
		vgf *vmmgr.VirtualGroupFamilyMeta
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

func (m *ManageModular) createGlobalVirtualGroup(vgfID uint32, params *storagetypes.Params) error {
	var err error
	if params == nil {
		if params, err = m.baseApp.Consensus().QueryStorageParamsByTimestamp(context.Background(), time.Now().Unix()); err != nil {
			return err
		}
	}
	gvgMeta, err := m.virtualGroupManager.GenerateGlobalVirtualGroupMeta(params)
	if err != nil {
		return err
	}
	virtualGroupParams, err := m.baseApp.Consensus().QueryVirtualGroupParams(context.Background())
	if err != nil {
		return err
	}
	return m.baseApp.GfSpClient().CreateGlobalVirtualGroup(context.Background(), &gfspserver.GfSpCreateGlobalVirtualGroup{
		VirtualGroupFamilyId: vgfID,
		PrimarySpAddress:     m.baseApp.OperatorAddress(),
		SecondarySpIds:       gvgMeta.SecondarySPIDs,
		Deposit: &sdk.Coin{
			Denom:  virtualGroupParams.GetDepositDenom(),
			Amount: sdk.NewInt(int64(gvgMeta.StakingStorageSize * virtualGroupParams.GvgStakingPrice.BigInt().Uint64())),
		},
	})
}

// pickGlobalVirtualGroup is used to pick a suitable gvg for replicating object.
func (m *ManageModular) pickGlobalVirtualGroup(ctx context.Context, vgfID uint32, param *storagetypes.Params) (*vmmgr.GlobalVirtualGroupMeta, error) {
	var (
		err error
		gvg *vmmgr.GlobalVirtualGroupMeta
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
	return gvg, nil
}
