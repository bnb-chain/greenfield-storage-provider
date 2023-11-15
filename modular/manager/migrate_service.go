package manager

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (m *ManageModular) getSPID() (uint32, error) {
	if m.spID != 0 {
		return m.spID, nil
	}
	spInfo, err := m.baseApp.Consensus().QuerySP(context.Background(), m.baseApp.OperatorAddress())
	if err != nil {
		return 0, err
	}
	m.spID = spInfo.GetId()
	return m.spID, nil
}

// NotifyMigrateSwapOut is used to receive migrate gvg task from src sp.
func (m *ManageModular) NotifyMigrateSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) error {
	var (
		err      error
		selfSPID uint32
	)
	if m.spExitScheduler == nil {
		log.CtxError(ctx, "sp exit scheduler has no init")
		return ErrNotifyMigrateSwapOut
	}
	selfSPID, err = m.getSPID()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get self sp id", "error", err)
		return err
	}
	if selfSPID != swapOut.SuccessorSpId {
		log.CtxErrorw(ctx, "successor sp id is mismatch", "self_sp_id", selfSPID, "swap_out_successor_sp_id", swapOut.SuccessorSpId)
		return fmt.Errorf("successor sp id is mismatch")
	}

	return m.spExitScheduler.AddSwapOutToTaskRunner(swapOut)
}

// NotifyPreMigrateBucketAndDeductQuota is used to notify record bucket is migrating
func (m *ManageModular) NotifyPreMigrateBucketAndDeductQuota(ctx context.Context, bucketID uint64) (gfsptask.GfSpBucketQuotaInfo, error) {
	var (
		state      int
		err        error
		bucketSize uint64
	)

	if state, err = m.baseApp.GfSpDB().QueryMigrateBucketState(bucketID); err != nil {
		log.CtxErrorw(ctx, "failed to query migrate bucket state", "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	if state != int(Init) {
		log.CtxInfow(ctx, "the bucket has already notified", "bucket_id", bucketID)
		return gfsptask.GfSpBucketQuotaInfo{}, fmt.Errorf("the bucket has already notified")
	}

	// get bucket quota and check, lock quota
	bucketSize, err = m.getBucketTotalSize(ctx, bucketID)
	if err != nil {
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	quota, err := m.baseApp.GfSpClient().GetLatestBucketReadQuota(
		ctx, bucketID)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket read quota", "bucket_id",
			bucketID, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	// reduce quota, sql db
	err = m.baseApp.GfSpClient().DeductQuotaForBucketMigrate(
		ctx, bucketID, bucketSize, quota.GetMonth())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket read quota", "bucket_id",
			bucketID, "error", err)
		return quota, err
	}

	// update state
	if err = m.baseApp.GfSpDB().UpdateBucketMigrationPreDeductedQuota(bucketID, bucketSize, int(SrcSPPreDeductQuotaDone)); err != nil {
		log.CtxErrorw(ctx, "failed to update migrate bucket state", "error", err)
		// if failed to update migrate bucket state, recoup quota and return error
		log.CtxErrorw(ctx, "failed to pre done migrate bucket due to update db failed", "bucket_id",
			bucketID, "error", err)
		quotaUpdateErr := m.baseApp.GfSpClient().RecoupQuota(ctx, bucketID, bucketSize, quota.GetMonth())
		if quotaUpdateErr != nil {
			log.CtxErrorw(ctx, "failed to recoup extra quota to user", "error", err)
		}
		return quota, err
	}

	return quota, nil
}

// NotifyPostMigrateBucketAndRecoupQuota is used to notify src sp confirm that only one Post migrate bucket is allowed.
func (m *ManageModular) NotifyPostMigrateBucketAndRecoupQuota(ctx context.Context, bmInfo *gfsptask.GfSpBucketMigrationInfo) (gfsptask.GfSpBucketQuotaInfo, error) {

	var (
		err        error
		extraQuota uint64
		bucketSize uint64
	)

	bucketID := bmInfo.GetBucketId()

	// ignore this error, this is in the last post phase
	//if err = m.baseApp.GfSpDB().DeleteMigrateBucket(bucketID); err != nil {
	//	log.CtxErrorw(ctx, "failed to delete the migrate bucket status", "error", err)
	//}

	latestQuota, err := m.baseApp.GfSpClient().GetLatestBucketReadQuota(
		ctx, bucketID)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket read quota", "bucket_id",
			bucketID, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	bucketSizeStr, err := m.baseApp.GfSpClient().GetBucketSize(ctx, bucketID)
	if err != nil {
		log.Errorf("failed to get bucket total object size", "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	bucketSize, err = util.StringToUint64(bucketSizeStr)
	if err != nil {
		log.CtxErrorw(ctx, "failed to convert bucket size to uint64", "bucket_id",
			bucketID, "bucket_size", bucketSize, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	if bmInfo.GetFinished() {
		// bucket migration gc
		//go m.GenerateGCBucketMigrationTask(ctx, bucketID, bucketSize)
	} else {
		// get bucket quota and check TODO month check
		bucketSize, err = m.getBucketTotalSize(ctx, bucketID)
		if err != nil {
			return gfsptask.GfSpBucketQuotaInfo{}, err
		}
		migratedBytes := bmInfo.GetMigratedBytesSize()
		if migratedBytes >= bucketSize {
			// If the data migrated surpasses the total bucket size, quota recoup is skipped.
			// This situation may arise due to deletions in the bucket migration process.
			log.CtxErrorw(ctx, "failed to recoup extra quota to user", "error", err)
		} else {
			extraQuota = bucketSize - migratedBytes
			quotaUpdateErr := m.baseApp.GfSpClient().RecoupQuota(ctx, bmInfo.GetBucketId(), extraQuota, latestQuota.GetMonth())
			// no need to return the db error to user
			if quotaUpdateErr != nil {
				log.CtxErrorw(ctx, "failed to recoup extra quota to user", "error", err)
			}
			if err = m.baseApp.GfSpDB().UpdateBucketMigrationRecoupQuota(bucketID, extraQuota, int(Init)); err != nil {
				log.CtxErrorw(ctx, "failed to update bucket migrate progress recoup quota", "error", err)
			}
		}
		log.CtxDebugw(ctx, "succeed to recoup extra quota to user", "extra_quote", extraQuota)
	}

	return gfsptask.GfSpBucketQuotaInfo{}, nil
}

// getBucketTotalSize return the total size of the bucket
func (m *ManageModular) getBucketTotalSize(ctx context.Context, bucketID uint64) (uint64, error) {
	bucketSize, err := m.baseApp.GfSpClient().GetBucketSize(
		ctx, bucketID)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket size", "bucket_id", bucketID, "error", err)
		return 0, err
	}
	quotaNeed, err := util.StringToUint64(bucketSize)
	if err != nil {
		log.CtxErrorw(ctx, "failed to convert bucket size to uint64", "bucket_id",
			bucketID, "bucket_size", bucketSize, "error", err)
		return 0, err
	}
	return quotaNeed, nil
}
