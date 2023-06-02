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
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	// DiscontinueBucketReason defines the reason for stop serving
	DiscontinueBucketReason = "testnet cleanup"

	// DiscontinueBucketLimit define the max buckets to fetch in a single request
	DiscontinueBucketLimit = int64(500)
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
	gcBlockHeight         uint64 // TODO: load from db?
	gcObjectBlockInterval uint64
	gcSafeBlockDistance   uint64

	syncConsensusInfoInterval uint64
	statisticsOutputInterval  int

	discontinueBucketEnabled       bool
	discontinueBucketTimeInterval  int
	discontinueBucketKeepAliveDays int
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
	m.syncConsensusInfo(ctx)
	gcObjectTicker := time.NewTicker(time.Duration(m.gcObjectTimeInterval) * time.Second)
	syncConsensusInfoTicker := time.NewTicker(time.Duration(m.syncConsensusInfoInterval) * time.Second)
	statisticsTicker := time.NewTicker(time.Duration(m.statisticsOutputInterval) * time.Second)
	discontinueBucketTicker := time.NewTicker(time.Duration(m.discontinueBucketTimeInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-statisticsTicker.C:
			log.CtxDebug(ctx, m.Statistics())
		case <-syncConsensusInfoTicker.C:
			m.syncConsensusInfo(ctx)
		case <-gcObjectTicker.C:
			start := m.gcBlockHeight
			end := m.gcBlockHeight + m.gcObjectBlockInterval
			currentBlockHeight, err := m.baseApp.Consensus().CurrentHeight(ctx)
			if err != nil {
				log.CtxErrorw(ctx, "failed to get current block height for gc object and try again later", "error", err)
				continue
			}
			if end+m.gcSafeBlockDistance > currentBlockHeight {
				log.CtxErrorw(ctx, "current block number less safe distance and try again later",
					"start_gc_block_height", start,
					"end_gc_block_height", end,
					"safe_distance", m.gcSafeBlockDistance,
					"current_block_height", currentBlockHeight)
				continue
			}
			task := &gfsptask.GfSpGCObjectTask{}
			task.InitGCObjectTask(m.baseApp.TaskPriority(task), start, end, m.baseApp.TaskTimeout(task, 0))
			if err = m.baseApp.GfSpDB().InsertGCObjectProgress(task.Key().String()); err != nil {
				log.CtxErrorw(ctx, "failed to init the gc object task", "error", err)
				continue
			}
			err = m.gcObjectQueue.Push(task)
			if err == nil {
				metrics.GCBlockNumberGauge.WithLabelValues(m.Name()).Set(float64(m.gcBlockHeight))
				m.gcBlockHeight = end + 1
			}
			log.CtxErrorw(ctx, "generate a gc object task", "task_info", task.Info(), "error", err)
		case <-discontinueBucketTicker.C:
			if !m.discontinueBucketEnabled {
				continue
			}
			m.discontinueBuckets(ctx)
			log.Infof("finish to discontinue buckets", "time", time.Now())
		}
	}
}

func (m *ManageModular) discontinueBuckets(ctx context.Context) {
	createAt := time.Now().AddDate(0, 0, -m.discontinueBucketKeepAliveDays)
	buckets, err := m.baseApp.GfSpClient().ListExpiredBucketsBySp(context.Background(),
		createAt.Unix(), m.baseApp.OperateAddress(), DiscontinueBucketLimit)
	if err != nil {
		log.Errorw("failed to query expired buckets", "error", err)
		return
	}

	for _, bucket := range buckets {
		time.Sleep(1 * time.Second)
		log.Infow("start to discontinue bucket", "bucket_name", bucket.BucketInfo.BucketName)
		discontinueBucket := &storagetypes.MsgDiscontinueBucket{
			BucketName: bucket.BucketInfo.BucketName,
			Reason:     DiscontinueBucketReason,
		}
		err = m.baseApp.GfSpClient().DiscontinueBucket(ctx, discontinueBucket)
		if err != nil {
			log.Errorw("failed to discontinue bucket on chain", "bucket_name",
				discontinueBucket.BucketName, "error", err)
			continue
		} else {
			log.Infow("succeed to discontinue bucket", "bucket_name",
				discontinueBucket.BucketName)
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
		if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:         task.GetObjectInfo().Id.Uint64(),
			TaskState:        types.TaskState_TASK_STATE_UPLOAD_OBJECT_ERROR,
			ErrorDescription: "expired",
		}); err != nil {
			log.Errorw("failed to update task state", "task_key", task.Key().String(), "error", err)
		}
		return true
	}
	return false
}

func (m *ManageModular) GCReplicatePieceQueue(qTask task.Task) bool {
	task := qTask.(task.ReplicatePieceTask)
	if task.Expired() {
		if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:         task.GetObjectInfo().Id.Uint64(),
			TaskState:        types.TaskState_TASK_STATE_REPLICATE_OBJECT_ERROR,
			ErrorDescription: "expired",
		}); err != nil {
			log.Errorw("failed to update task state", "task_key", task.Key().String(), "error", err)
		}
		return true
	}
	return false
}

func (m *ManageModular) GCSealObjectQueue(qTask task.Task) bool {
	task := qTask.(task.SealObjectTask)
	if task.Expired() {
		if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
			ObjectID:         task.GetObjectInfo().Id.Uint64(),
			TaskState:        types.TaskState_TASK_STATE_SEAL_OBJECT_ERROR,
			ErrorDescription: "expired",
		}); err != nil {
			log.Errorw("failed to update task state", "task_key", task.Key().String(), "error", err)
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
	return !qTask.Expired() && (qTask.GetRetry() == 0 || qTask.ExceedTimeout())
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
	spList, err := m.baseApp.Consensus().ListSPs(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list sps", "error", err)
		return
	}
	if err = m.baseApp.GfSpDB().UpdateAllSp(spList); err != nil {
		log.CtxErrorw(ctx, "failed to update all sp list", "error", err)
		return
	}
	for _, sp := range spList {
		if strings.EqualFold(m.baseApp.OperateAddress(), sp.OperatorAddress) {
			if err = m.baseApp.GfSpDB().SetOwnSpInfo(sp); err != nil {
				log.Errorw("failed to set own sp info", "error", err)
				return
			}
		}
	}
}

func (m *ManageModular) Statistics() string {
	return fmt.Sprintf(
		"upload[%d], replicate[%d], seal[%d], receive[%d], gcObject[%d], gcZombie[%d], gcMeta[%d], download[%d], challenge[%d], gcBlockHeight[%d], gcSafeDistance[%d]",
		m.uploadQueue.Len(), m.replicateQueue.Len(), m.sealQueue.Len(),
		m.receiveQueue.Len(), m.gcObjectQueue.Len(), m.gcZombieQueue.Len(),
		m.gcMetaQueue.Len(), m.downloadQueue.Len(), m.challengeQueue.Len(),
		m.gcBlockHeight, m.gcSafeBlockDistance)
}
