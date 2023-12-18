package approver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
)

var (
	ErrDanglingPointer     = gfsperrors.Register(module.ApprovalModularName, http.StatusBadRequest, 10001, "OoooH.... request lost")
	ErrExceedBucketNumber  = gfsperrors.Register(module.ApprovalModularName, http.StatusNotAcceptable, 10002, "account buckets exceed the limit")
	ErrExceedApprovalLimit = gfsperrors.Register(module.ApprovalModularName, http.StatusNotAcceptable, 10003, "SP is too busy to approve the request, please come back later")
)

const (
	SigExpireTimeSecond = 60 * 60
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
		log.CtxErrorw(ctx, "failed to create bucket approval due to pointer nil")
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
		state         int
	)
	if task == nil || task.GetMigrateBucketInfo() == nil {
		log.CtxErrorw(ctx, "failed to migrate bucket approval due to pointer nil")
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

	migrateBucketMsg := task.GetMigrateBucketInfo()
	bucketMeta, _, err := a.baseApp.GfSpClient().GetBucketMeta(ctx, migrateBucketMsg.GetBucketName(), true)
	if err != nil {
		return false, err
	}
	bucketID := bucketMeta.GetBucketInfo().Id.Uint64()

	// If the destination SP is still performing garbage collection for the bucket being migrated, the migration action is not allowed.
	if state, err = a.baseApp.GfSpDB().QueryMigrateBucketState(bucketID); err != nil {
		log.CtxErrorw(ctx, "failed to query migrate bucket state", "error", err)
		return false, err
	}
	if state == int(storetypes.BucketMigrationState_BUCKET_MIGRATION_STATE_MIGRATION_FINISHED) {
		// delete the last finished migrate bucket progress record
		if err = a.baseApp.GfSpDB().DeleteMigrateBucket(bucketID); err != nil {
			log.CtxErrorw(ctx, "failed to delete migrate bucket state", "bucket_id", bucketID, "error", err)
			return false, err
		}
	} else if state != int(storetypes.BucketMigrationState_BUCKET_MIGRATION_STATE_INIT_UNSPECIFIED) {
		log.CtxInfow(ctx, "the bucket is migrating or gc, migrated to this sp should be reject", "bucket_id", bucketID)
		return false, fmt.Errorf("the bucket is migrating or gc, try it after gc done")
	}

	// check src sp has enough quota
	allow, err := a.migrateBucketQuotaCheck(ctx, task)
	if err != nil || !allow {
		return allow, err
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
	log.CtxInfow(ctx, "succeed to hand migrate bucket approval", "task", task, "state", state)
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
		log.CtxErrorw(ctx, "failed to create object approval due to pointer nil")
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
	metrics.PerfApprovalTime.WithLabelValues("approval_object_sign_create_object_cost").Observe(time.Since(startSignApproval).Seconds())
	metrics.PerfApprovalTime.WithLabelValues("approval_object_sign_create_object_end").Observe(time.Since(startQueryQueue).Seconds())
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

func (a *ApprovalModular) migrateBucketQuotaCheck(ctx context.Context, task coretask.ApprovalMigrateBucketTask) (bool, error) {
	var (
		signature []byte
		err       error
	)
	migrateBucketMsg := task.GetMigrateBucketInfo()
	log.CtxDebugw(ctx, "start to check source sp whether has enough quota to execute bucket migration", "migrate_bucket_msg", migrateBucketMsg)
	bucketMeta, _, err := a.baseApp.GfSpClient().GetBucketMeta(ctx, migrateBucketMsg.GetBucketName(), true)
	if err != nil {
		return false, err
	}
	bucketID := bucketMeta.GetBucketInfo().Id.Uint64()
	spID := bucketMeta.GetVgf().GetPrimarySpId()

	srcSP, err := a.baseApp.Consensus().QuerySPByID(ctx, spID)
	if err != nil {
		log.CtxErrorw(ctx, "failed to query sp info", "sp_id", spID, "error", err)
		return false, ErrSignerWithDetail("failed to query sp info, error: " + err.Error())
	}

	bucketMigrationInfo := &gfsptask.GfSpBucketMigrationInfo{BucketId: bucketID}
	bucketMigrationInfo.ExpireTime = time.Now().Unix() + SigExpireTimeSecond
	signature, err = a.baseApp.GfSpClient().SignBucketMigrationInfo(context.Background(), bucketMigrationInfo)
	if err != nil {
		log.Errorw("failed to sign migrate bucket", "bucket_migration_info", bucketMigrationInfo, "error", err)
		return false, err
	} else {
		bucketMigrationInfo.SetSignature(signature)
	}
	err = a.baseApp.GfSpClient().QuerySPHasEnoughQuotaForMigrateBucket(ctx, srcSP.GetEndpoint(), bucketMigrationInfo)
	if err != nil {
		log.CtxErrorw(ctx, "failed to check src SP migrate bucket quota", "src_sp", srcSP, "bucket_id", bucketID, "error", err)
		return false, ErrSignerWithDetail("failed to check src SP migrate bucket quota, error: " + err.Error())
	}

	log.CtxDebugw(ctx, "succeed to check bucket source sp quota for bucket migration", "migrate_bucket_msg", migrateBucketMsg)
	return true, nil
}
