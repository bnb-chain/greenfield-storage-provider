package approver

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	DefaultBlockInterval          = 3
	DefaultApprovalExpiredTimeout = int64(DefaultBlockInterval * 20)
)

var _ module.Approver = &ApprovalModular{}

type managerTasksStats struct {
	uploadTaskCount          uint32
	replicateTaskCount       uint32
	sealTaskCount            uint32
	resumableUploadTaskCount uint32
	maxUploadingCount        uint32
	migrateGVGCount          uint32
}

func (s *managerTasksStats) totalUploadTasks() uint32 {
	return s.uploadTaskCount + s.replicateTaskCount + s.sealTaskCount + s.resumableUploadTaskCount
}

type ApprovalModular struct {
	baseApp     *gfspapp.GfSpBaseApp
	scope       rcmgr.ResourceScope
	bucketQueue taskqueue.TQueueOnStrategy
	objectQueue taskqueue.TQueueOnStrategy

	currentBlockHeight uint64
	// defines the max bucket number per account, approver refuses the ask approval
	// request if account own the bucket number greater the value
	accountBucketNumber int64
	// defines the creation of bucket/object approval timeout height the approval
	// expired height equal to current block height + timeout height
	bucketApprovalTimeoutHeight uint64
	objectApprovalTimeoutHeight uint64

	// the maximum number of GVGs migrating to current SP concurrently is allowed
	migrateGVGLimit int

	statsMutex sync.Mutex
	tasksStats *managerTasksStats

	spID uint32
}

func (a *ApprovalModular) Name() string {
	return module.ApprovalModularName
}

func (a *ApprovalModular) Start(ctx context.Context) error {
	a.bucketQueue.SetRetireTaskStrategy(a.GCApprovalQueue)
	a.objectQueue.SetRetireTaskStrategy(a.GCApprovalQueue)
	scope, err := a.baseApp.ResourceManager().OpenService(a.Name())
	if err != nil {
		return err
	}
	a.scope = scope
	a.tasksStats = &managerTasksStats{}
	go a.eventLoop(ctx)
	return nil
}

func (a *ApprovalModular) Stop(ctx context.Context) error {
	a.scope.Release()
	return nil
}

func (a *ApprovalModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	span, err := a.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (a *ApprovalModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
}

func (a *ApprovalModular) eventLoop(ctx context.Context) {
	getCurrentBlockHeightTicker := time.NewTicker(time.Duration(DefaultBlockInterval) * time.Second)
	updateManagerTasksStatsTicket := time.NewTicker(time.Duration(DefaultBlockInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-getCurrentBlockHeightTicker.C:
			current, err := a.baseApp.Consensus().CurrentHeight(context.Background())
			if err != nil {
				log.CtxErrorw(ctx, "failed to get current block number", "error", err)
				current = atomic.LoadUint64(&a.currentBlockHeight) + 1
			}
			a.SetCurrentBlockHeight(current)
		case <-updateManagerTasksStatsTicket.C:
			stats, err := a.baseApp.GfSpClient().GetTasksStats(context.Background())
			if err == nil {
				a.statsMutex.Lock()
				a.tasksStats = &managerTasksStats{
					stats.GetUploadCount(),
					stats.GetReplicateCount(),
					stats.GetSealCount(),
					stats.GetResumableUploadCount(),
					stats.GetMaxUploading(),
					stats.GetMigrateGvgCount(),
				}
				a.statsMutex.Unlock()
			}
		}
	}
}

// GCApprovalQueue defines the strategy of gc approval queue when the queue is full.
// if the approval is expired, it can be deleted.
func (a *ApprovalModular) GCApprovalQueue(qTask task.Task) bool {
	approvalTask := qTask.(task.ApprovalTask)
	if approvalTask.GetCreateTime()+DefaultApprovalExpiredTimeout < time.Now().Unix() {
		log.Debugw("expire approval task", "info", approvalTask.Info())
		return true
	}
	log.Debugw("approval task not expired", "expired_height", approvalTask.GetExpiredHeight(),
		"current_height", a.GetCurrentBlockHeight())
	return false
}

func (a *ApprovalModular) GetCurrentBlockHeight() uint64 {
	return atomic.LoadUint64(&a.currentBlockHeight)
}

func (a *ApprovalModular) SetCurrentBlockHeight(height uint64) {
	atomic.StoreUint64(&a.currentBlockHeight, height)
}

func (a *ApprovalModular) getSPID() (uint32, error) {
	if a.spID != 0 {
		return a.spID, nil
	}
	spInfo, err := a.baseApp.Consensus().QuerySP(context.Background(), a.baseApp.OperatorAddress())
	if err != nil {
		return 0, err
	}
	a.spID = spInfo.GetId()
	return a.spID, nil
}

func (a *ApprovalModular) exceedCreateObjectLimit() bool {
	a.statsMutex.Lock()
	defer a.statsMutex.Unlock()
	return a.tasksStats.totalUploadTasks() >= a.tasksStats.maxUploadingCount
}

func (a *ApprovalModular) exceedMigrateGVGLimit() bool {
	a.statsMutex.Lock()
	defer a.statsMutex.Unlock()
	return a.tasksStats.migrateGVGCount >= uint32(a.migrateGVGLimit)
}
