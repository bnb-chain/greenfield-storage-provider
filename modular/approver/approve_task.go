package approver

import (
	"context"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrDanglingPointer    = gfsperrors.Register(module.ApprovalModularName, http.StatusBadRequest, 10001, "OoooH.... request lost")
	ErrExceedBucketNumber = gfsperrors.Register(module.ApprovalModularName, http.StatusNotAcceptable, 10002, "account buckets exceed the limit")
	ErrSigner             = gfsperrors.Register(module.ApprovalModularName, http.StatusInternalServerError, 11001, "server slipped away, try again later")
)

func (a *ApprovalModular) PreCreateBucketApproval(ctx context.Context, task coretask.ApprovalCreateBucketTask) error {
	return nil
}

func (a *ApprovalModular) HandleCreateBucketApprovalTask(ctx context.Context, task coretask.ApprovalCreateBucketTask) (bool, error) {
	var (
		err           error
		signature     []byte
		currentHeight uint64
	)
	if task == nil || task.GetCreateBucketInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre create bucket approval due to pointer nil")
		return false, ErrDanglingPointer
	}
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()
	startQueryQueue := time.Now()
	if a.bucketQueue.Has(task.Key()) {
		shadowTask := a.bucketQueue.PopByKey(task.Key())
		task.SetCreateBucketInfo(shadowTask.(coretask.ApprovalCreateBucketTask).GetCreateBucketInfo())
		_ = a.bucketQueue.Push(shadowTask)
		metrics.PerfGetApprovalTimeHistogram.WithLabelValues("check_repeated_in_create_bucket_approval").
			Observe(time.Since(startQueryQueue).Seconds())
		log.CtxErrorw(ctx, "repeated create bucket approval task is returned")
		return true, nil
	}
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("check_repeated_in_create_bucket_approval").
		Observe(time.Since(startQueryQueue).Seconds())

	startQueryMetadata := time.Now()
	buckets, err := a.baseApp.GfSpClient().GetUserBucketsCount(ctx, task.GetCreateBucketInfo().GetCreator(), false)
	if err != nil {
		metrics.PerfGetApprovalTimeHistogram.WithLabelValues("check_counter_in_create_bucket_approval").
			Observe(time.Since(startQueryMetadata).Seconds())
		log.CtxErrorw(ctx, "failed to get account owns max bucket number", "error", err)
		return false, err
	}
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("check_counter_in_create_bucket_approval").
		Observe(time.Since(startQueryMetadata).Seconds())
	if buckets >= a.accountBucketNumber {
		log.CtxErrorw(ctx, "account owns bucket number exceed")
		err = ErrExceedBucketNumber
		return false, err
	}

	startPickVGF := time.Now()
	vgfID, err := a.baseApp.GfSpClient().PickVirtualGroupFamilyID(ctx, task)
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("pick_vfg").Observe(time.Since(startPickVGF).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to pick virtual group family", "error", err)
		return false, err
	}

	task.GetCreateBucketInfo().PrimarySpApproval.GlobalVirtualGroupFamilyId = vgfID
	// begin to sign the new approval task
	startQueryChain := time.Now()
	currentHeight = a.GetCurrentBlockHeight()
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("query_current_height_in_create_bucket_approval").
		Observe(time.Since(startQueryChain).Seconds())
	task.SetExpiredHeight(currentHeight + a.bucketApprovalTimeoutHeight)
	startSignApproval := time.Now()
	signature, err = a.baseApp.GfSpClient().SignCreateBucketApproval(ctx, task.GetCreateBucketInfo())
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("sign_in_create_bucket_approval").
		Observe(time.Since(startSignApproval).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign the create bucket approval", "error", err)
		return false, ErrSigner
	}
	task.GetCreateBucketInfo().GetPrimarySpApproval().Sig = signature
	startPushQueue := time.Now()
	if err = a.bucketQueue.Push(task); err != nil {
		metrics.PerfGetApprovalTimeHistogram.WithLabelValues("update_queue_in_create_bucket_approval").
			Observe(time.Since(startPushQueue).Seconds())
		log.CtxErrorw(ctx, "failed to push the create bucket approval to queue", "error", err)
		return false, err
	}
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("update_queue_in_create_bucket_approval").
		Observe(time.Since(startPushQueue).Seconds())
	return true, nil
}

func (a *ApprovalModular) PostCreateBucketApproval(ctx context.Context, task coretask.ApprovalCreateBucketTask) {
}

func (a *ApprovalModular) PreCreateObjectApproval(ctx context.Context, task coretask.ApprovalCreateObjectTask) error {
	return nil
}

func (a *ApprovalModular) HandleCreateObjectApprovalTask(ctx context.Context, task coretask.ApprovalCreateObjectTask) (bool, error) {
	var (
		err           error
		signature     []byte
		currentHeight uint64
	)
	if task == nil || task.GetCreateObjectInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre create object approval due to pointer nil")
		return false, ErrDanglingPointer
	}
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()

	startQueryQueue := time.Now()
	if a.objectQueue.Has(task.Key()) {
		shadowTask := a.objectQueue.PopByKey(task.Key())
		task.SetCreateObjectInfo(shadowTask.(coretask.ApprovalCreateObjectTask).GetCreateObjectInfo())
		_ = a.objectQueue.Push(shadowTask)
		metrics.PerfGetApprovalTimeHistogram.WithLabelValues("check_repeated_in_create_object_approval").
			Observe(time.Since(startQueryQueue).Seconds())
		log.CtxErrorw(ctx, "repeated create object approval task is returned")
		return true, nil
	}
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("check_repeated_in_create_object_approval").
		Observe(time.Since(startQueryQueue).Seconds())

	// begin to sign the new approval task
	startQueryChain := time.Now()
	currentHeight = a.GetCurrentBlockHeight()
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("query_current_height_in_create_object_approval").
		Observe(time.Since(startQueryChain).Seconds())
	task.SetExpiredHeight(currentHeight + a.objectApprovalTimeoutHeight)
	startSignApproval := time.Now()
	signature, err = a.baseApp.GfSpClient().SignCreateObjectApproval(ctx, task.GetCreateObjectInfo())
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("sign_in_create_object_approval").
		Observe(time.Since(startSignApproval).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign the create object approval", "error", err)
		return false, err
	}
	task.GetCreateObjectInfo().GetPrimarySpApproval().Sig = signature
	startPushQueue := time.Now()
	if err = a.objectQueue.Push(task); err != nil {
		metrics.PerfGetApprovalTimeHistogram.WithLabelValues("update_queue_in_create_object_approval").
			Observe(time.Since(startPushQueue).Seconds())
		log.CtxErrorw(ctx, "failed to push the create object task to queue", "error", err)
		return false, err
	}
	metrics.PerfGetApprovalTimeHistogram.WithLabelValues("update_queue_in_create_object_approval").
		Observe(time.Since(startPushQueue).Seconds())
	return true, nil
}

func (a *ApprovalModular) PostCreateObjectApproval(ctx context.Context, task coretask.ApprovalCreateObjectTask) {
}

func (a *ApprovalModular) QueryTasks(ctx context.Context, subKey coretask.TKey) ([]coretask.Task, error) {
	bucketApprovalTasks, _ := taskqueue.ScanTQueueBySubKey(a.bucketQueue, subKey)
	objectApprovalTasks, _ := taskqueue.ScanTQueueBySubKey(a.objectQueue, subKey)
	return append(bucketApprovalTasks, objectApprovalTasks...), nil
}
