package manager

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
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
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	// DiscontinueBucketReason defines the reason for stop serving
	DiscontinueBucketReason = "testnet cleanup"

	// DiscontinueBucketLimit define the max buckets to fetch in a single request
	DiscontinueBucketLimit = int64(500)

	// RejectUnSealObjectRetry defines the retry number of sending reject unseal object tx.
	RejectUnSealObjectRetry = 3

	// RejectUnSealObjectTimeout defines the timeout of sending reject unseal object tx.
	RejectUnSealObjectTimeout = 3

	// DefaultBackupTaskTimeout defines the timeout of backing up task for dispatching
	DefaultBackupTaskTimeout = 1
)

var _ module.Manager = &ManageModular{}

type ManageModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope

	taskCh        chan task.Task
	backupTaskNum int64
	backupTaskMux sync.Mutex

	// loading task at startup.
	enableLoadTask           bool
	loadTaskLimitToReplicate int
	loadTaskLimitToSeal      int
	loadTaskLimitToGC        int

	uploadQueue          taskqueue.TQueueOnStrategy
	resumableUploadQueue taskqueue.TQueueOnStrategy
	replicateQueue       taskqueue.TQueueOnStrategyWithLimit
	sealQueue            taskqueue.TQueueOnStrategyWithLimit
	receiveQueue         taskqueue.TQueueOnStrategyWithLimit
	gcObjectQueue        taskqueue.TQueueOnStrategyWithLimit
	gcZombieQueue        taskqueue.TQueueOnStrategyWithLimit
	gcMetaQueue          taskqueue.TQueueOnStrategyWithLimit
	downloadQueue        taskqueue.TQueueOnStrategy
	challengeQueue       taskqueue.TQueueOnStrategy
	recoveryQueue        taskqueue.TQueueOnStrategyWithLimit
	migrateGVGQueue      taskqueue.TQueueOnStrategyWithLimit
	migrateGVGQueueMux   sync.Mutex

	maxUploadObjectNumber int

	gcObjectTimeInterval  int
	gcBlockHeight         uint64
	gcObjectBlockInterval uint64
	gcSafeBlockDistance   uint64

	syncConsensusInfoInterval uint64
	statisticsOutputInterval  int

	discontinueBucketEnabled       bool
	discontinueBucketTimeInterval  int
	discontinueBucketKeepAliveDays int

	spID                   uint32
	virtualGroupManager    vgmgr.VirtualGroupManager
	bucketMigrateScheduler *BucketMigrateScheduler
	spExitScheduler        *SPExitScheduler

	subscribeSPExitEventInterval        uint
	subscribeBucketMigrateEventInterval uint
	subscribeSwapOutEventInterval       uint

	loadReplicateTimeout int64
	loadSealTimeout      int64

	gvgPreferSPList []uint32
}

func (m *ManageModular) Name() string {
	return module.ManageModularName
}

func (m *ManageModular) Start(ctx context.Context) error {
	m.uploadQueue.SetRetireTaskStrategy(m.GCUploadObjectQueue)
	m.resumableUploadQueue.SetRetireTaskStrategy(m.GCResumableUploadObjectQueue)
	m.replicateQueue.SetRetireTaskStrategy(m.GCReplicatePieceQueue)
	m.replicateQueue.SetFilterTaskStrategy(m.FilterUploadingTask)
	m.sealQueue.SetRetireTaskStrategy(m.GCSealObjectQueue)
	m.sealQueue.SetFilterTaskStrategy(m.FilterUploadingTask)
	m.receiveQueue.SetRetireTaskStrategy(m.GCReceiveQueue)
	m.receiveQueue.SetFilterTaskStrategy(m.FilterReceiveTask)
	m.gcObjectQueue.SetRetireTaskStrategy(m.ResetGCObjectTask)
	m.gcObjectQueue.SetFilterTaskStrategy(m.FilterGCTask)
	m.downloadQueue.SetRetireTaskStrategy(m.GCCacheQueue)
	m.challengeQueue.SetRetireTaskStrategy(m.GCCacheQueue)
	m.recoveryQueue.SetRetireTaskStrategy(m.GCRecoverQueue)
	m.recoveryQueue.SetFilterTaskStrategy(m.FilterUploadingTask)
	m.migrateGVGQueue.SetRetireTaskStrategy(m.GCMigrateGVGQueue)
	m.migrateGVGQueue.SetFilterTaskStrategy(m.FilterGVGTask)

	scope, err := m.baseApp.ResourceManager().OpenService(m.Name())
	if err != nil {
		return err
	}
	m.scope = scope
	err = m.LoadTaskFromDB()
	if err != nil {
		return err
	}

	if err = m.LoadTaskFromDB(); err != nil {
		return err
	}

	go m.delayStartMigrateScheduler()
	go m.eventLoop(ctx)
	return nil
}

func (m *ManageModular) delayStartMigrateScheduler() {
	// delay start to wait metadata service ready.
	// migrate scheduler init depend metadata.
	for {
		time.Sleep(5 * time.Second)
		var err error
		if m.bucketMigrateScheduler == nil {
			if m.bucketMigrateScheduler, err = NewBucketMigrateScheduler(m); err != nil {
				log.Errorw("failed to new bucket migrate scheduler", "error", err)
				continue
			}
		}
		if m.spExitScheduler == nil {
			if m.spExitScheduler, err = NewSPExitScheduler(m); err != nil {
				log.Errorw("failed to new sp exit scheduler", "error", err)
				continue
			}
		}
		log.Info("succeed to start migrate scheduler")
		return
	}
}

func (m *ManageModular) eventLoop(ctx context.Context) {
	m.syncConsensusInfo(ctx)
	gcObjectTicker := time.NewTicker(time.Duration(m.gcObjectTimeInterval) * time.Second)
	syncConsensusInfoTicker := time.NewTicker(time.Duration(m.syncConsensusInfoInterval) * time.Second)
	statisticsTicker := time.NewTicker(time.Duration(m.statisticsOutputInterval) * time.Second)
	discontinueBucketTicker := time.NewTicker(time.Duration(m.discontinueBucketTimeInterval) * time.Second)
	backupTaskTicker := time.NewTicker(time.Duration(DefaultBackupTaskTimeout) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-statisticsTicker.C:
			log.CtxDebug(ctx, m.Statistics())
		case <-syncConsensusInfoTicker.C:
			m.syncConsensusInfo(ctx)
		case <-backupTaskTicker.C:
			m.backUpTask()
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
			err = m.gcObjectQueue.Push(task)
			if err == nil {
				metrics.GCBlockNumberGauge.WithLabelValues(ManagerGCBlockNumber).Set(float64(m.gcBlockHeight))
				m.gcBlockHeight = end + 1

				if err = m.baseApp.GfSpDB().InsertGCObjectProgress(&spdb.GCObjectMeta{
					TaskKey:          task.Key().String(),
					StartBlockHeight: start,
					EndBlockHeight:   end,
				}); err != nil {
					log.CtxErrorw(ctx, "failed to init the gc object task", "error", err)
					continue
				}
			}
			log.CtxErrorw(ctx, "generate a gc object task", "task_info", task.Info(), "error", err)
		case <-discontinueBucketTicker.C:
			if !m.discontinueBucketEnabled {
				continue
			}
			go m.discontinueBuckets(ctx)
			log.Infow("finished to discontinue buckets", "time", time.Now())
		}
	}
}

func (m *ManageModular) discontinueBuckets(ctx context.Context) {
	createAt := time.Now().AddDate(0, 0, -m.discontinueBucketKeepAliveDays)
	spID, err := m.getSPID()
	if err != nil {
		log.Errorw("failed to query sp id", "error", err)
		return
	}
	buckets, err := m.baseApp.GfSpClient().ListExpiredBucketsBySp(context.Background(),
		createAt.Unix(), spID, DiscontinueBucketLimit)
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
		_, err = m.baseApp.GfSpClient().DiscontinueBucket(ctx, discontinueBucket)
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
	if !m.enableLoadTask {
		log.Info("skip load tasks from db")
		return nil
	}

	var (
		err                          error
		replicateMetas               []*spdb.UploadObjectMeta
		generateReplicateTaskCounter int
		sealMetas                    []*spdb.UploadObjectMeta
		generateSealTaskCounter      int
		gcObjectMetas                []*spdb.GCObjectMeta
		generateGCObjectTaskCounter  int
	)

	log.Info("start to load task from sp db")

	replicateMetas, err = m.baseApp.GfSpDB().GetUploadMetasToReplicate(m.loadTaskLimitToReplicate, m.loadReplicateTimeout)
	if err != nil {
		log.Errorw("failed to load replicate task from sp db", "error", err)
		return err
	}
	for _, meta := range replicateMetas {
		objectInfo, queryErr := m.baseApp.Consensus().QueryObjectInfoByID(context.Background(), util.Uint64ToString(meta.ObjectID))
		if queryErr != nil {
			log.Errorw("failed to query object info and continue", "object_id", meta.ObjectID, "error", queryErr)
			continue
		}

		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.Infow("object is not in create status and continue", "object_info", objectInfo)
			continue
		}
		storageParams, queryErr := m.baseApp.Consensus().QueryStorageParamsByTimestamp(context.Background(), objectInfo.GetCreateAt())
		if queryErr != nil {
			log.Errorw("failed to query storage param and continue", "object_id", meta.ObjectID, "error", queryErr)
			continue
		}
		replicateTask := &gfsptask.GfSpReplicatePieceTask{}
		replicateTask.InitReplicatePieceTask(objectInfo, storageParams, m.baseApp.TaskPriority(replicateTask),
			m.baseApp.TaskTimeout(replicateTask, objectInfo.GetPayloadSize()), m.baseApp.TaskMaxRetry(replicateTask))
		replicateTask.SetSecondaryAddresses(meta.SecondaryEndpoints)
		replicateTask.SetSecondarySignatures(meta.SecondarySignatures)
		replicateTask.GlobalVirtualGroupId = meta.GlobalVirtualGroupID
		pushErr := m.replicateQueue.Push(replicateTask)
		if pushErr != nil {
			log.Errorw("failed to push replicate piece task to queue", "object_info", objectInfo, "error", pushErr)
			continue
		}
		generateReplicateTaskCounter++
	}

	sealMetas, err = m.baseApp.GfSpDB().GetUploadMetasToSeal(m.loadTaskLimitToSeal, m.loadSealTimeout)
	if err != nil {
		log.Errorw("failed to load seal task from sp db", "error", err)
		return err
	}
	for _, meta := range sealMetas {
		objectInfo, queryErr := m.baseApp.Consensus().QueryObjectInfoByID(context.Background(), util.Uint64ToString(meta.ObjectID))
		if queryErr != nil {
			log.Errorw("failed to query object info and continue", "object_id", meta.ObjectID, "error", queryErr)
			continue
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.Infow("object is not in create status and continue", "object_info", objectInfo)
			continue
		}
		storageParams, queryErr := m.baseApp.Consensus().QueryStorageParamsByTimestamp(context.Background(), objectInfo.GetCreateAt())
		if queryErr != nil {
			log.Errorw("failed to query storage param and continue", "object_id", meta.ObjectID, "error", queryErr)
			continue
		}
		sealTask := &gfsptask.GfSpSealObjectTask{}
		sealTask.InitSealObjectTask(meta.GlobalVirtualGroupID, objectInfo, storageParams, m.baseApp.TaskPriority(sealTask),
			meta.SecondaryEndpoints, meta.SecondarySignatures, m.baseApp.TaskTimeout(sealTask, 0), m.baseApp.TaskMaxRetry(sealTask))
		pushErr := m.sealQueue.Push(sealTask)
		if pushErr != nil {
			log.Errorw("failed to push seal object task to queue", "object_info", objectInfo, "error", pushErr)
			continue
		}
		generateSealTaskCounter++
	}

	gcObjectMetas, err = m.baseApp.GfSpDB().GetGCMetasToGC(m.loadTaskLimitToGC)
	if err != nil {
		log.Errorw("failed to load gc task from sp db", "error", err)
		return err
	}
	for _, meta := range gcObjectMetas {
		gcObjectTask := &gfsptask.GfSpGCObjectTask{}
		gcObjectTask.InitGCObjectTask(m.baseApp.TaskPriority(gcObjectTask), meta.StartBlockHeight, meta.EndBlockHeight, m.baseApp.TaskTimeout(gcObjectTask, 0))
		gcObjectTask.SetGCObjectProgress(meta.CurrentBlockHeight, meta.LastDeletedObjectID)
		pushErr := m.gcObjectQueue.Push(gcObjectTask)
		if pushErr != nil {
			log.Errorw("failed to push gc object task to queue", "gc_object_task_meta", meta, "error", pushErr)
			continue
		}
		generateGCObjectTaskCounter++
		if meta.EndBlockHeight >= m.gcBlockHeight {
			m.gcBlockHeight = meta.EndBlockHeight + 1
		}
	}

	log.Infow("end to load task from sp db", "replicate_task_number", generateReplicateTaskCounter,
		"seal_task_number", generateSealTaskCounter, "gc_object_task_number", generateGCObjectTaskCounter)
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
	if m.resumableUploadQueue.Has(task.Key()) {
		log.CtxDebugw(ctx, "resumable uploading object repeated")
		return true
	}
	return false
}

func (m *ManageModular) TaskRecovering(ctx context.Context, task task.Task) bool {
	if m.recoveryQueue.Has(task.Key()) {
		log.CtxDebugw(ctx, "recovery object repeated")
		return true
	}

	return false
}

func (m *ManageModular) UploadingObjectNumber() int {
	return m.uploadQueue.Len() + m.replicateQueue.Len() + m.sealQueue.Len() + m.resumableUploadQueue.Len()
}

func (m *ManageModular) GCUploadObjectQueue(qTask task.Task) bool {
	task := qTask.(task.UploadObjectTask)
	if task.Expired() {
		go func() {
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         task.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_UPLOAD_OBJECT_ERROR,
				ErrorDescription: "expired",
			}); err != nil {
				log.Errorw("failed to update task state", "task_key", task.Key().String(), "error", err)
			}
		}()
		return true
	}
	return false
}

func (m *ManageModular) GCResumableUploadObjectQueue(qTask task.Task) bool {
	task := qTask.(task.ResumableUploadObjectTask)
	if task.Expired() {
		go func() {
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         task.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_UPLOAD_OBJECT_ERROR,
				ErrorDescription: "expired",
			}); err != nil {
				log.Errorw("failed to update task state", "task_key", task.Key().String(), "error", err)
			}
		}()
		return true
	}
	return false
}

func (m *ManageModular) GCReplicatePieceQueue(qTask task.Task) bool {
	task := qTask.(task.ReplicatePieceTask)
	if task.Expired() {
		go func() {
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         task.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_REPLICATE_OBJECT_ERROR,
				ErrorDescription: "expired",
			}); err != nil {
				log.Errorw("failed to update task state", "task_key", task.Key().String(), "error", err)
			}
		}()
		return true
	}
	return false
}

func (m *ManageModular) GCSealObjectQueue(qTask task.Task) bool {
	task := qTask.(task.SealObjectTask)
	if task.Expired() {
		go func() {
			if err := m.baseApp.GfSpDB().UpdateUploadProgress(&spdb.UploadObjectMeta{
				ObjectID:         task.GetObjectInfo().Id.Uint64(),
				TaskState:        types.TaskState_TASK_STATE_SEAL_OBJECT_ERROR,
				ErrorDescription: "expired",
			}); err != nil {
				log.Errorw("failed to update task state", "task_key", task.Key().String(), "error", err)
			}
		}()
		return true
	}
	return false
}

func (m *ManageModular) GCReceiveQueue(qTask task.Task) bool {
	return qTask.ExceedRetry()
}

func (m *ManageModular) GCRecoverQueue(qTask task.Task) bool {
	return qTask.ExceedRetry() || qTask.ExceedTimeout()
}

func (m *ManageModular) GCMigrateGVGQueue(qTask task.Task) bool {
	task := qTask.(task.MigrateGVGTask)
	return task.GetFinished()
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
	if qTask.ExceedRetry() {
		return false
	}
	if qTask.ExceedTimeout() {
		return true
	}
	if qTask.GetRetry() == 0 {
		return true
	}
	return false
}

func (m *ManageModular) FilterGVGTask(qTask task.Task) bool {
	if qTask.GetRetry() == 0 {
		return true
	}
	if qTask.ExceedTimeout() {
		return true
	}
	return false
}

func (m *ManageModular) FilterReceiveTask(qTask task.Task) bool {
	if qTask.ExceedRetry() {
		return false
	}
	if qTask.ExceedTimeout() {
		return true
	}
	return false
}

func (m *ManageModular) PickUpTask(ctx context.Context, tasks []task.Task) (task.Task, []task.Task) {
	if len(tasks) == 0 {
		return nil, nil
	}
	if len(tasks) == 1 {
		log.CtxDebugw(ctx, "only one task for picking")
		return tasks[0], nil
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].GetPriority() < tasks[j].GetPriority()
	})
	var totalPriority int
	for _, t := range tasks {
		totalPriority += int(t.GetPriority())
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randPriority := r.Intn(totalPriority)
	log.Debugw("pick up task", "total_priority", totalPriority, "rand_priority", randPriority)

	totalPriority = 0
	for i, t := range tasks {
		totalPriority += int(t.GetPriority())
		if totalPriority >= randPriority {
			t.AppendLog("pickup-to-backup-task-pool")
			return t, append(tasks[:i], tasks[i+1:]...)
		}
	}
	return nil, tasks
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
		if strings.EqualFold(m.baseApp.OperatorAddress(), sp.OperatorAddress) {
			if err = m.baseApp.GfSpDB().SetOwnSpInfo(sp); err != nil {
				log.Errorw("failed to set own sp info", "error", err)
				return
			}
		}
	}
}

func (m *ManageModular) RejectUnSealObject(ctx context.Context, object *storagetypes.ObjectInfo) error {
	rejectUnSealObjectMsg := &storagetypes.MsgRejectSealObject{
		BucketName: object.GetBucketName(),
		ObjectName: object.GetObjectName(),
	}

	var err error
	for i := 0; i < RejectUnSealObjectRetry; i++ {
		_, err = m.baseApp.GfSpClient().RejectUnSealObject(ctx, rejectUnSealObjectMsg)
		if err != nil {
			time.Sleep(RejectUnSealObjectTimeout * time.Second)
		} else {
			log.CtxDebugw(ctx, "succeed to reject unseal object")
			reject, err := m.baseApp.Consensus().ListenRejectUnSealObject(ctx, object.Id.Uint64(), DefaultListenRejectUnSealTimeoutHeight)
			if err != nil {
				log.CtxErrorw(ctx, "failed to reject unseal object", "error", err)
				continue
			}
			if !reject {
				log.CtxErrorw(ctx, "failed to reject unseal object")
				continue
			}
			return nil
		}
	}
	log.CtxErrorw(ctx, "failed to reject unseal object", "error", err)
	return err
}

func (m *ManageModular) Statistics() string {
	return fmt.Sprintf(
		"upload[%d], replicate[%d], seal[%d], receive[%d], recovery[%d] gcObject[%d], gcZombie[%d], gcMeta[%d], download[%d], challenge[%d], migrateGVG[%d], gcBlockHeight[%d], gcSafeDistance[%d], backupTaskNum[%d]",
		m.uploadQueue.Len(), m.replicateQueue.Len(), m.sealQueue.Len(),
		m.receiveQueue.Len(), m.recoveryQueue.Len(), m.gcObjectQueue.Len(), m.gcZombieQueue.Len(),
		m.gcMetaQueue.Len(), m.downloadQueue.Len(), m.challengeQueue.Len(), m.migrateGVGQueue.Len(),
		m.gcBlockHeight, m.gcSafeBlockDistance, m.backupTaskNum)
}

func (m *ManageModular) backUpTask() {
	m.backupTaskMux.Lock()
	defer m.backupTaskMux.Unlock()

	startPopTime := time.Now().String()
	var (
		backupTasks   []task.Task
		reservedTasks []task.Task
		targetTask    task.Task

		ctx   = context.Background()
		limit = &rcmgr.Unlimited{}
	)

	targetTask = m.replicateQueue.PopByLimit(limit)
	if targetTask != nil {
		log.CtxDebugw(ctx, "add replicate piece task to backup set", "task_key", targetTask.Key().String(),
			"task_limit", targetTask.EstimateLimit().String())
		backupTasks = append(backupTasks, targetTask)
	}
	targetTask = m.sealQueue.PopByLimit(limit)
	if targetTask != nil {
		log.CtxDebugw(ctx, "add seal object task to backup set", "task_key", targetTask.Key().String(),
			"task_limit", targetTask.EstimateLimit().String())
		backupTasks = append(backupTasks, targetTask)
	}
	targetTask = m.gcObjectQueue.PopByLimit(limit)
	if targetTask != nil {
		log.CtxDebugw(ctx, "add gc object task to backup set", "task_key", targetTask.Key().String(),
			"task_limit", targetTask.EstimateLimit().String())
		backupTasks = append(backupTasks, targetTask)
	}
	targetTask = m.gcZombieQueue.PopByLimit(limit)
	if targetTask != nil {
		log.CtxDebugw(ctx, "add gc zombie piece task to backup set", "task_key", targetTask.Key().String(),
			"task_limit", targetTask.EstimateLimit().String())
		backupTasks = append(backupTasks, targetTask)
	}
	targetTask = m.gcMetaQueue.PopByLimit(limit)
	if targetTask != nil {
		log.CtxDebugw(ctx, "add gc meta task to backup set", "task_key", targetTask.Key().String(),
			"task_limit", targetTask.EstimateLimit().String())
		backupTasks = append(backupTasks, targetTask)
	}
	targetTask = m.receiveQueue.PopByLimit(limit)
	if targetTask != nil {
		log.CtxDebugw(ctx, "add confirm receive piece to backup set", "task_key", targetTask.Key().String(),
			"task_limit", targetTask.EstimateLimit().String())
		backupTasks = append(backupTasks, targetTask)
	}
	targetTask = m.recoveryQueue.PopByLimit(limit)
	if targetTask != nil {
		log.CtxDebugw(ctx, "add confirm recovery piece to backup set", "task_key", targetTask.Key().String(),
			"task_limit", targetTask.EstimateLimit().String())
		backupTasks = append(backupTasks, targetTask)
	}
	targetTask = m.migrateGVGQueuePopByLimit(limit)
	if targetTask != nil {
		log.CtxDebugw(ctx, "add confirm migrate gvg to backup set", "task_key", targetTask.Key().String())
		backupTasks = append(backupTasks, targetTask)
	}
	endPopTime := time.Now().String()

	startPickUpTime := time.Now().String()
	targetTask, reservedTasks = m.PickUpTask(ctx, backupTasks)
	if targetTask != nil {
		targetTask.AppendLog("start-pop-task-from-queue:" + startPopTime)
		targetTask.AppendLog("end-pop-task-from-queue:" + endPopTime)
		targetTask.AppendLog("start-pickup-task-to-dispatch: " + startPickUpTime)
		targetTask.AppendLog("end-pickup-task-to-dispatch")

		atomic.AddInt64(&m.backupTaskNum, 1)
		m.taskCh <- targetTask
	}

	for _, reservedTask := range reservedTasks {
		m.repushTask(reservedTask)
	}
}

func (m *ManageModular) repushTask(reserved task.Task) {
	switch t := reserved.(type) {
	case *gfsptask.GfSpReplicatePieceTask:
		err := m.replicateQueue.Push(t)
		log.Errorw("failed to retry push replicate task to queue after dispatching", "error", err)
	case *gfsptask.GfSpSealObjectTask:
		err := m.sealQueue.Push(t)
		log.Errorw("failed to retry push seal task to queue after dispatching", "error", err)
	case *gfsptask.GfSpReceivePieceTask:
		err := m.receiveQueue.Push(t)
		log.Errorw("failed to retry push receive task to queue after dispatching", "error", err)
	case *gfsptask.GfSpGCObjectTask:
		err := m.gcObjectQueue.Push(t)
		log.Errorw("failed to retry push gc object task to queue after dispatching", "error", err)
	case *gfsptask.GfSpGCZombiePieceTask:
		err := m.gcZombieQueue.Push(t)
		log.Errorw("failed to retry push gc zombie task to queue after dispatching", "error", err)
	case *gfsptask.GfSpGCMetaTask:
		err := m.gcMetaQueue.Push(t)
		log.Errorw("failed to retry push gc meta task to queue after dispatching", "error", err)
	case *gfsptask.GfSpRecoverPieceTask:
		err := m.recoveryQueue.Push(t)
		log.Errorw("failed to retry push recovery task to queue after dispatching", "error", err)
	case *gfsptask.GfSpMigrateGVGTask:
		err := m.migrateGVGQueuePush(t)
		log.Errorw("failed to retry push migration gvg task to queue after dispatching", "error", err)
	}
}

func (m *ManageModular) migrateGVGQueuePush(task task.Task) error {
	m.migrateGVGQueueMux.Lock()
	defer m.migrateGVGQueueMux.Unlock()

	err := m.migrateGVGQueue.Push(task)

	return err
}

func (m *ManageModular) migrateGVGQueuePopByLimit(limit rcmgr.Limit) task.Task {
	m.migrateGVGQueueMux.Lock()
	defer m.migrateGVGQueueMux.Unlock()
	task := m.migrateGVGQueue.PopByLimit(limit)

	return task
}

func (m *ManageModular) migrateGVGQueuePopByKey(key task.TKey) {
	m.migrateGVGQueueMux.Lock()
	defer m.migrateGVGQueueMux.Unlock()
	m.migrateGVGQueue.PopByKey(key)
}

func (m *ManageModular) migrateGVGQueuePopByLimitAndPushAgain(task task.MigrateGVGTask, push bool) error {
	m.migrateGVGQueueMux.Lock()
	defer m.migrateGVGQueueMux.Unlock()

	var pushErr error

	m.migrateGVGQueue.PopByKey(task.Key())
	task.SetUpdateTime(time.Now().Unix())
	if !task.GetFinished() || push {
		if pushErr = m.migrateGVGQueue.Push(task); pushErr != nil {
			log.Errorw("failed to push gvg task queue", "task", task, "error", pushErr)
		}
	}
	log.Debugw("succeed to push gvg task queue", "task", task, "queue", m.migrateGVGQueue, "push", push, "error", pushErr)

	return pushErr
}
