package manager

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/service/types"
)

var _ module.Manager = &ManageModular{}

type ManageModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope

	uploadQueue    taskqueue.TQueueOnStrategy
	replicateQueue taskqueue.TQueueOnStrategyWithLimit
	sealQueue      taskqueue.TQueueOnStrategyWithLimit
	receiveQueue   taskqueue.TQueueOnStrategyWithLimit
	gcObjectQueue  taskqueue.TQueueOnStrategyWithLimit
	gcZombieQueue  taskqueue.TQueueOnStrategyWithLimit
	gcMetaQueue    taskqueue.TQueueOnStrategyWithLimit
	downloadQueue  taskqueue.TQueueOnStrategy
	challengeQueue taskqueue.TQueueOnStrategy

	maxUploadObjectNumber int

	gcObjectTimeInterval  int
	gcBlockHeight         uint64
	gcObjectBlockInterval uint64
	gcSafeBlockDistance   uint64

	syncConsensusInfoInterval uint64
}

func (m *ManageModular) Name() string {
	return module.ManageModularName
}

func (m *ManageModular) Start(ctx context.Context) error {
	m.uploadQueue.SetRetireTaskStrategy(m.GCUploadObjectQueue)
	m.replicateQueue.SetRetireTaskStrategy(m.GCReplicatePieceQueue)
	m.replicateQueue.SetFilterTaskStrategy(m.FilterUploadingTask)
	m.sealQueue.SetRetireTaskStrategy(m.GCSealObjectQueue)
	m.sealQueue.SetFilterTaskStrategy(m.FilterUploadingTask)
	m.receiveQueue.SetRetireTaskStrategy(m.GCReceiveQueue)
	m.receiveQueue.SetFilterTaskStrategy(m.FilterUploadingTask)
	m.gcObjectQueue.SetRetireTaskStrategy(m.ResetGCObjectTask)
	m.gcObjectQueue.SetFilterTaskStrategy(m.FilterGCTask)
	m.downloadQueue.SetRetireTaskStrategy(m.GCCacheQueue)
	m.challengeQueue.SetRetireTaskStrategy(m.GCCacheQueue)
	scope, err := m.baseApp.ResourceManager().OpenService(m.Name())
	if err != nil {
		return err
	}
	m.scope = scope
	err = m.LoadTaskFromDB()
	if err != nil {
		return err
	}

	go m.eventLoop(ctx)
	return nil
}

func (m *ManageModular) eventLoop(ctx context.Context) {
	gcObjectTicker := time.NewTicker(time.Duration(m.gcObjectTimeInterval) * time.Second)
	syncConsensusInfoTicker := time.NewTicker(time.Duration(m.syncConsensusInfoInterval) * time.Second)
	statisticsTicker := time.NewTicker(time.Duration(60) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-statisticsTicker.C:
			log.CtxDebugw(ctx, m.State())
		case <-syncConsensusInfoTicker.C:
			m.syncConsensusInfo(ctx)
		case <-gcObjectTicker.C:
			start := m.gcBlockHeight
			end := m.gcBlockHeight + m.gcObjectBlockInterval
			currentBlockHeight, err := m.baseApp.Consensus().CurrentHeight(ctx)
			if err != nil {
				log.CtxErrorw(ctx, "failed to get current block number for gc object")
				continue
			}
			if end+m.gcSafeBlockDistance < currentBlockHeight {
				log.CtxErrorw(ctx, "current block number less safe distance",
					"current_block_height", currentBlockHeight,
					"gc_block_height", m.gcBlockHeight,
					"safe_distance", m.gcSafeBlockDistance)
				continue
			}
			task := &gfsptask.GfSpGCObjectTask{}
			task.InitGCObjectTask(m.baseApp.TaskPriority(task), start, end, m.baseApp.TaskTimeout(task))
			err = m.baseApp.GfSpDB().SetGCObjectProgress(task.Key().String(), start, end)
			if err != nil {
				log.CtxErrorw(ctx, "failed to update gc object status", "error", err)
				continue
			}
			err = m.gcObjectQueue.Push(task)
			if err == nil {
				m.gcBlockHeight = end + 1
				metrics.GCBlockNumberGauge.WithLabelValues(m.Name()).Set(float64(m.gcBlockHeight))
			}
			log.CtxErrorw(ctx, "finish to generate gc object task", "task_key", task.Key().String(),
				"start_block_height", start, "end_block_height", end, "error", err)
		}
	}
}

func (m *ManageModular) Stop(ctx context.Context) error {
	m.scope.Release()
	return nil
}

func (m *ManageModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	span, err := m.scope.BeginSpan()
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin span", "error", err)
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve resource", "error", err)
		return nil, err
	}
	return span, nil
}

func (m *ManageModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
}

func (m *ManageModular) LoadTaskFromDB() error {
	return nil
}

func (m *ManageModular) TaskUploading(ctx context.Context, task task.Task) bool {
	if m.uploadQueue.Has(task.Key()) {
		log.CtxDebugw(ctx, "uploading object repeated")
		return true
	}
	if m.replicateQueue.Has(task.Key()) {
		log.CtxDebugw(ctx, "replicating object repeated")
		return true
	}
	if m.sealQueue.Has(task.Key()) {
		log.CtxDebugw(ctx, "sealing object repeated")
		return true
	}
	return false
}

func (m *ManageModular) UploadingObjectNumber() int {
	return m.uploadQueue.Len() + m.replicateQueue.Len() + m.sealQueue.Len()
}

func (m *ManageModular) GCUploadObjectQueue(qTask task.Task) bool {
	task := qTask.(task.UploadObjectTask)
	if task.Expired() {
		err := m.baseApp.GfSpDB().UpdateJobState(task.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR)
		if err != nil {
			log.Errorw("failed to update job state", "task_key", task.Key().String(), "error, err")
		}
		return true
	}
	return false
}

func (m *ManageModular) GCReplicatePieceQueue(qTask task.Task) bool {
	task := qTask.(task.ReplicatePieceTask)
	if task.Expired() {
		err := m.baseApp.GfSpDB().UpdateJobState(task.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
		if err != nil {
			log.Errorw("failed to update job state", "task_key", task.Key().String(), "error, err")
		}
		return true
	}
	return false
}

func (m *ManageModular) GCSealObjectQueue(qTask task.Task) bool {
	task := qTask.(task.SealObjectTask)
	if task.Expired() {
		err := m.baseApp.GfSpDB().UpdateJobState(task.GetObjectInfo().Id.Uint64(), types.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
		if err != nil {
			log.Errorw("failed to update job state", "task_key", task.Key().String(), "error, err")
		}
		return true
	}
	return false
}

func (m *ManageModular) GCReceiveQueue(qTask task.Task) bool {
	return qTask.Expired()
}

func (m *ManageModular) ResetGCObjectTask(qTask task.Task) bool {
	task := qTask.(task.GCObjectTask)
	if task.Expired() {
		log.Errorw("reset gc object task", "old_task_key", task.Key().String())
		task.SetRetry(0)
		log.Errorw("reset gc object task", "new_task_key", task.Key().String())
	}
	return false
}

func (m *ManageModular) GCCacheQueue(qTask task.Task) bool {
	return true
}

func (m *ManageModular) FilterGCTask(qTask task.Task) bool {
	return qTask.GetRetry() == 0
}

func (m *ManageModular) FilterUploadingTask(qTask task.Task) bool {
	return qTask.ExceedTimeout()
}

func (m *ManageModular) PickUpTask(ctx context.Context, tasks []task.Task) task.Task {
	if len(tasks) == 0 {
		return nil
	}
	if len(tasks) == 1 {
		log.CtxDebugw(ctx, "only one task for picking")
		return tasks[0]
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].GetPriority() < tasks[j].GetPriority()
	})
	var totalPriority int
	for _, task := range tasks {
		totalPriority += int(task.GetPriority())
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randPriority := r.Intn(totalPriority)
	totalPriority = 0

	for _, task := range tasks {
		totalPriority += int(task.GetPriority())
		if totalPriority >= randPriority {
			return task
		}
	}
	return nil
}

func (m *ManageModular) syncConsensusInfo(ctx context.Context) {
	spInfoList, err := m.baseApp.Consensus().QuerySPInfo(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to query sp info", "error", err)
		return
	}
	if err = m.baseApp.GfSpDB().UpdateAllSp(spInfoList); err != nil {
		log.CtxErrorw(ctx, "failed to update sp info", "error", err)
		return
	}
	for _, spInfo := range spInfoList {
		if strings.EqualFold(m.baseApp.OperateAddress(), spInfo.OperatorAddress) {
			if err = m.baseApp.GfSpDB().SetOwnSpInfo(spInfo); err != nil {
				log.Errorw("failed to set own sp info", "error", err)
				return
			}
		}
	}
}

func (m *ManageModular) State() string {
	return fmt.Sprintf(
		"upload[%d], replicate[%d], seal[%d], receive[%d], gcObject[%d], gcZombie[%d], gcMeta[%d], download[%d], challenge[%d], gcBlock[%d], gcSafeDistanc[%d]",
		m.uploadQueue.Len(), m.replicateQueue.Len(), m.sealQueue.Len(),
		m.receiveQueue.Len(), m.gcObjectQueue.Len(), m.gcZombieQueue.Len(),
		m.gcMetaQueue.Len(), m.downloadQueue.Len(), m.challengeQueue.Len(),
		m.gcBlockHeight, m.gcSafeBlockDistance)
}
