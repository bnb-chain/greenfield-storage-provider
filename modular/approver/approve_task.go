package approver

import (
	"context"
	"fmt"
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
	ErrDanglingPointer     = gfsperrors.Register(module.ApprovalModularName, http.StatusBadRequest, 10001, "OoooH.... request lost")
	ErrExceedBucketNumber  = gfsperrors.Register(module.ApprovalModularName, http.StatusNotAcceptable, 10002, "account buckets exceed the limit")
	ErrExceedApprovalLimit = gfsperrors.Register(module.ApprovalModularName, http.StatusNotAcceptable, 10003, "SP is too busy to approve the request, please come back later")
)

func ErrSignerWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.ApprovalModularName, http.StatusInternalServerError, 11001, detail)
}

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
	has := a.bucketQueue.Has(task.Key())
	metrics.PerfApprovalTime.WithLabelValues("approval_bucket_check_repeated_cost").Observe(time.Since(startQueryQueue).Seconds())
	if has {
		shadowTask := a.bucketQueue.PopByKey(task.Key())
		task.SetCreateBucketInfo(shadowTask.(coretask.ApprovalCreateBucketTask).GetCreateBucketInfo())
		_ = a.bucketQueue.Push(shadowTask)
		log.CtxErrorw(ctx, "repeated create bucket approval task is returned")
		return true, nil
	}
	startQueryMetadata := time.Now()
	buckets, err := a.baseApp.GfSpClient().GetUserBucketsCount(ctx, task.GetCreateBucketInfo().GetCreator(), false)
	metrics.PerfApprovalTime.WithLabelValues("approval_bucket_get_bucket_count_cost").Observe(time.Since(startQueryMetadata).Seconds())
	metrics.PerfApprovalTime.WithLabelValues("approval_bucket_get_bucket_count_end").Observe(time.Since(startQueryQueue).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get account owns max bucket number", "error", err)
		return false, err
	}
	if buckets >= a.accountBucketNumber {
		log.CtxErrorw(ctx, "account owns bucket number exceed")
		err = ErrExceedBucketNumber
		return false, err
	}

	startPickVGF := time.Now()
	vgfID, err := a.baseApp.GfSpClient().PickVirtualGroupFamilyID(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pick virtual group family", "time_cost", time.Since(startPickVGF).Seconds(), "error", err)
		return false, err
	}
	log.Debugw("succeed to pick vgf id", "vgf_id", vgfID, "time_cost", time.Since(startPickVGF).Seconds())

	task.GetCreateBucketInfo().PrimarySpApproval.GlobalVirtualGroupFamilyId = vgfID
	// begin to sign the new approval task
	startQueryChain := time.Now()
	currentHeight = a.GetCurrentBlockHeight()
	metrics.PerfApprovalTime.WithLabelValues("approval_bucket_query_block_height_cost").Observe(time.Since(startQueryChain).Seconds())
	metrics.PerfApprovalTime.WithLabelValues("approval_bucket_query_block_height_end").Observe(time.Since(startQueryQueue).Seconds())
	task.SetExpiredHeight(currentHeight + a.bucketApprovalTimeoutHeight)
	startSignApproval := time.Now()
	signature, err = a.baseApp.GfSpClient().SignCreateBucketApproval(ctx, task.GetCreateBucketInfo())
	metrics.PerfApprovalTime.WithLabelValues("approval_bucket_sign_create_bucket_cost").Observe(time.Since(startSignApproval).Seconds())
	metrics.PerfApprovalTime.WithLabelValues("approval_bucket_sign_create_bucket_end").Observe(time.Since(startQueryQueue).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign the create bucket approval", "error", err)
		return false, ErrSignerWithDetail("failed to sign the create bucket approval, error: " + err.Error())
	}
	task.GetCreateBucketInfo().GetPrimarySpApproval().Sig = signature
	go a.bucketQueue.Push(task)
	return true, nil
}

func (a *ApprovalModular) PostCreateBucketApproval(ctx context.Context, task coretask.ApprovalCreateBucketTask) {
}

func (a *ApprovalModular) PreMigrateBucketApproval(ctx context.Context, task coretask.ApprovalMigrateBucketTask) error {
	var err error
	if task == nil || task.GetMigrateBucketInfo() == nil {
		log.CtxError(ctx, "failed to ask migrate bucket approval due to bucket info pointer dangling")
		return ErrDanglingPointer
	}
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()
	selfSPID, err := a.getSPID()
	if err != nil {
		return err
	}
	if selfSPID != task.GetMigrateBucketInfo().GetDstPrimarySpId() {
		return fmt.Errorf("current SP is not the correct one to ask for approval")
	}
	if a.exceedMigrateGVGLimit() {
		log.CtxErrorw(ctx, "Exceeding SP concurrent GVGs migration limit", "limit", a.migrateGVGLimit)
		return ErrExceedApprovalLimit
	}
	return nil
}

func (a *ApprovalModular) HandleMigrateBucketApprovalTask(ctx context.Context, task coretask.ApprovalMigrateBucketTask) (bool, error) {
	var (
		err           error
		signature     []byte
		currentHeight uint64
	)
	if task == nil || task.GetMigrateBucketInfo() == nil {
		log.CtxErrorw(ctx, "failed to pre migrate bucket approval due to pointer nil")
		return false, ErrDanglingPointer
	}
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, task.Info())
	}()
	if a.bucketQueue.Has(task.Key()) {
		shadowTask := a.bucketQueue.PopByKey(task.Key())
		task.SetMigrateBucketInfo(shadowTask.(coretask.ApprovalMigrateBucketTask).GetMigrateBucketInfo())
		_ = a.bucketQueue.Push(shadowTask)
		log.CtxErrorw(ctx, "repeated migrate bucket approval task is returned")
		return true, nil
	}

	// begin to sign the new approval task
	currentHeight = a.GetCurrentBlockHeight()
	task.SetExpiredHeight(currentHeight + a.bucketApprovalTimeoutHeight)
	signature, err = a.baseApp.GfSpClient().SignMigrateBucketApproval(ctx, task.GetMigrateBucketInfo())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign migrate bucket approval", "error", err)
		return false, ErrSignerWithDetail("failed to sign migrate bucket approval, error: " + err.Error())
	}
	task.GetMigrateBucketInfo().GetDstPrimarySpApproval().Sig = signature
	_ = a.bucketQueue.Push(task)
	return true, nil
}

func (a *ApprovalModular) PostMigrateBucketApproval(ctx context.Context, task coretask.ApprovalMigrateBucketTask) {
}

func (a *ApprovalModular) PreCreateObjectApproval(_ context.Context, _ coretask.ApprovalCreateObjectTask) error {
	if a.exceedCreateObjectLimit() {
		return ErrExceedApprovalLimit
	}
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
	has := a.objectQueue.Has(task.Key())
	metrics.PerfApprovalTime.WithLabelValues("approval_object_check_repeated_cost").Observe(time.Since(startQueryQueue).Seconds())
	if has {
		shadowTask := a.objectQueue.PopByKey(task.Key())
		task.SetCreateObjectInfo(shadowTask.(coretask.ApprovalCreateObjectTask).GetCreateObjectInfo())
		_ = a.objectQueue.Push(shadowTask)
		log.CtxErrorw(ctx, "repeated create object approval task is returned")
		return true, nil
	}

	// begin to sign the new approval task
	startQueryChain := time.Now()
	currentHeight = a.GetCurrentBlockHeight()
	metrics.PerfApprovalTime.WithLabelValues("approval_object_query_block_height_cost").Observe(time.Since(startQueryChain).Seconds())
	metrics.PerfApprovalTime.WithLabelValues("approval_object_query_block_height_end").Observe(time.Since(startQueryQueue).Seconds())
	task.SetExpiredHeight(currentHeight + a.objectApprovalTimeoutHeight)
	startSignApproval := time.Now()
	signature, err = a.baseApp.GfSpClient().SignCreateObjectApproval(ctx, task.GetCreateObjectInfo())
	metrics.PerfApprovalTime.WithLabelValues("approval_object_sign_create_bucket_cost").Observe(time.Since(startSignApproval).Seconds())
	metrics.PerfApprovalTime.WithLabelValues("approval_object_sign_create_bucket_end").Observe(time.Since(startQueryQueue).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign create object approval", "error", err)
		return false, err
	}
	task.GetCreateObjectInfo().GetPrimarySpApproval().Sig = signature
	go a.objectQueue.Push(task)
	return true, nil
}

func (a *ApprovalModular) PostCreateObjectApproval(ctx context.Context, task coretask.ApprovalCreateObjectTask) {
}

func (a *ApprovalModular) QueryTasks(ctx context.Context, subKey coretask.TKey) ([]coretask.Task, error) {
	bucketApprovalTasks, _ := taskqueue.ScanTQueueBySubKey(a.bucketQueue, subKey)
	objectApprovalTasks, _ := taskqueue.ScanTQueueBySubKey(a.objectQueue, subKey)
	return append(bucketApprovalTasks, objectApprovalTasks...), nil
}
