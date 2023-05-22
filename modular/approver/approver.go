package approver

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var _ module.Approver = &ApprovalModular{}

type ApprovalModular struct {
	baseApp     *gfspapp.GfSpBaseApp
	scope       rcmgr.ResourceScope
	bucketQueue taskqueue.TQueueOnStrategy
	objectQueue taskqueue.TQueueOnStrategy

	// defines the max bucket number per account, approver refuses the ask approval
	// request if account own the bucket number greater the value
	accountBucketNumber int64
	// defines the creation of bucket/object approval timeout height the approval
	// expired height equal to current block height + timeout height
	bucketApprovalTimeoutHeight uint64
	objectApprovalTimeoutHeight uint64
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

// GCApprovalQueue defines the strategy of gc approval queue when the queue is full.
// if the approval is expired, it can be deleted.
func (a *ApprovalModular) GCApprovalQueue(qTask task.Task) bool {
	task := qTask.(task.ApprovalTask)
	ctx := log.WithValue(context.Background(), log.CtxKeyTask, task.Key().String())
	current, err := a.baseApp.Consensus().CurrentHeight(context.Background())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get current height", "error", err)
		return false
	}
	if task.GetExpiredHeight() < current {
		log.CtxDebugw(ctx, "expire approval task", "info", task.Info())
		return true
	}
	log.CtxDebugw(ctx, "approval task not expired", "current_height", current,
		"expired_height", task.GetExpiredHeight())
	return false
}
