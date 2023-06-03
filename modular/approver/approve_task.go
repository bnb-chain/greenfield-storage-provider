package approver

import (
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	ErrDanglingPointer    = gfsperrors.Register(module.ApprovalModularName, http.StatusInternalServerError, 10001, "OoooH.... request lost")
	ErrExceedBucketNumber = gfsperrors.Register(module.ApprovalModularName, http.StatusServiceUnavailable, 10002, "account buckets exceed the limit")
	ErrRepeatedTask       = gfsperrors.Register(module.ApprovalModularName, http.StatusBadRequest, 10003, "ask approval request repeated")
	ErrExceedQueue        = gfsperrors.Register(module.ApprovalModularName, http.StatusServiceUnavailable, 10004, "ask approval request exceed the limit, try again later")
	ErrSigner             = gfsperrors.Register(module.ApprovalModularName, http.StatusInternalServerError, 11001, "server slipped away, try again later")
	ErrConsensus          = gfsperrors.Register(module.ApprovalModularName, http.StatusInternalServerError, 15001, "server slipped away, try again later")
)

func (a *ApprovalModular) PreCreateBucketApproval(ctx context.Context, task coretask.ApprovalCreateBucketTask) error {
	if task == nil || task.GetCreateBucketInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre create bucket approval, pointer nil")
		return ErrDanglingPointer
	}
	buckets, err := a.baseApp.GfSpClient().GetUserBucketsCount(ctx, task.GetCreateBucketInfo().GetCreator())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get account owns max bucket number", "error", err)
		return err
	}
	if buckets >= a.accountBucketNumber {
		log.CtxErrorw(ctx, "account owns bucket number exceed")
		return ErrExceedBucketNumber
	}
	if a.bucketQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "failed to pre create bucket approval, task repeated")
		return ErrRepeatedTask
	}
	return nil
}

func (a *ApprovalModular) HandleCreateBucketApprovalTask(ctx context.Context, task coretask.ApprovalCreateBucketTask) (bool, error) {
	var (
		err           error
		signature     []byte
		currentHeight uint64
	)
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()
	currentHeight, err = a.baseApp.Consensus().CurrentHeight(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get current height", "error", err)
		return false, ErrConsensus
	}
	task.SetExpiredHeight(currentHeight + a.bucketApprovalTimeoutHeight)
	signature, err = a.baseApp.GfSpClient().SignCreateBucketApproval(ctx, task.GetCreateBucketInfo())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign the create bucket approval", "error", err)
		return false, ErrSigner
	}
	task.GetCreateBucketInfo().GetPrimarySpApproval().Sig = signature
	if err = a.bucketQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push the create bucket approval to queue", "error", err)
		return false, ErrExceedQueue
	}
	return true, nil
}

func (a *ApprovalModular) PostCreateBucketApproval(ctx context.Context, task coretask.ApprovalCreateBucketTask) {
}

func (a *ApprovalModular) PreCreateObjectApproval(ctx context.Context, task coretask.ApprovalCreateObjectTask) error {
	if task == nil || task.GetCreateObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre create object approval, pointer nil")
		return ErrDanglingPointer
	}
	if a.objectQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "failed to pre create object approval, task repeated")
		return ErrRepeatedTask
	}
	return nil
}

func (a *ApprovalModular) HandleCreateObjectApprovalTask(ctx context.Context, task coretask.ApprovalCreateObjectTask) (bool, error) {
	var (
		err           error
		signature     []byte
		currentHeight uint64
	)
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()
	currentHeight, err = a.baseApp.Consensus().CurrentHeight(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get current height", "error", err)
		return false, ErrConsensus
	}
	task.SetExpiredHeight(currentHeight + a.objectApprovalTimeoutHeight)
	signature, err = a.baseApp.GfSpClient().SignCreateObjectApproval(ctx, task.GetCreateObjectInfo())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign the create object approval", "error", err)
		return false, err
	}
	task.GetCreateObjectInfo().GetPrimarySpApproval().Sig = signature
	if err = a.objectQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push the create object task to queue", "error", err)
		return false, ErrExceedQueue
	}
	return true, nil
}

func (a *ApprovalModular) PostCreateObjectApproval(ctx context.Context, task coretask.ApprovalCreateObjectTask) {
}

func (a *ApprovalModular) QueryTasks(
	ctx context.Context,
	subKey coretask.TKey) (
	[]coretask.Task, error) {
	bucketApprovalTasks, _ := taskqueue.ScanTQueueBySubKey(a.bucketQueue, subKey)
	objectApprovalTasks, _ := taskqueue.ScanTQueueBySubKey(a.objectQueue, subKey)
	return append(bucketApprovalTasks, objectApprovalTasks...), nil
}
