package p2p

import (
	"context"
	"net/http"

	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsperrors"
	"github.com/zkMeLabs/mechain-storage-provider/core/module"
	"github.com/zkMeLabs/mechain-storage-provider/core/task"
	"github.com/zkMeLabs/mechain-storage-provider/core/taskqueue"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
)

var (
	ErrRepeatedTask         = gfsperrors.Register(module.P2PModularName, http.StatusBadRequest, 70001, "request repeated")
	ErrInsufficientApproval = gfsperrors.Register(module.P2PModularName, http.StatusNotFound, 70002, "insufficient approvals as secondary sp")
)

func (p *P2PModular) HandleReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask,
	min, max int32, timeout int64,
) ([]task.ApprovalReplicatePieceTask, error) {
	if p.replicateApprovalQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "repeated task")
		return nil, ErrRepeatedTask
	}
	if err := p.replicateApprovalQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push replicate piece approval task to queue", "error", err)
		return nil, err
	}
	defer p.replicateApprovalQueue.PopByKey(task.Key())
	approvals, err := p.node.GetSecondaryReplicatePieceApproval(ctx, task, int(max), timeout)
	if err != nil {
		log.CtxErrorw(ctx, "failed to ask secondary replicate piece approval", "error", err)
		return nil, err
	}
	if len(approvals) < int(min) {
		log.CtxErrorw(ctx, "failed to get sufficient approvals as secondary sp",
			"accept", len(approvals), "min", min, "max", max)
		return nil, ErrInsufficientApproval
	}
	return approvals, nil
}

func (p *P2PModular) HandleQueryBootstrap(ctx context.Context) ([]string, error) {
	return p.node.Bootstrap(), nil
}

func (p *P2PModular) QueryTasks(
	ctx context.Context,
	subKey task.TKey) (
	[]task.Task, error,
) {
	approvalTasks, _ := taskqueue.ScanTQueueBySubKey(p.replicateApprovalQueue, subKey)
	return approvalTasks, nil
}
