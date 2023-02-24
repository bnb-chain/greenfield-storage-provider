package greenfield

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// AuthUploadObjectWithAccount verify the greenfield chain information for upload object.
func (greenfield *Greenfield) AuthUploadObjectWithAccount(ctx context.Context, bucket, object, account, sp string) (
	accountExist bool, bucketExist bool, objectInitStatue bool, paymentEnough bool,
	spBucket bool, ownerObject bool, err error) {
	accountExist, err = greenfield.HasAccount(ctx, account)
	if err != nil || !accountExist {
		log.Warnw("failed to query account", "error", err, "is_account_exist", accountExist)
		return
	}
	var bucketInfo *storagetypes.BucketInfo
	bucketInfo, err = greenfield.QueryBucketInfo(ctx, bucket)
	if err != nil || bucketInfo == nil {
		bucketExist = false
		log.Warnw("failed to query bucket info", "error", err, "bucket_info", bucketInfo)
		return
	}
	bucketExist = true

	var objectInfo *storagetypes.ObjectInfo
	objectInfo, err = greenfield.QueryObjectInfo(ctx, bucket, object)
	if err != nil || objectInfo == nil {
		objectInitStatue = false
		log.Warnw("failed to query object info", "error", err, "object_info", objectInfo)
		return
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_INIT {
		objectInitStatue = false
		log.Warnw("object status is not equal to status_init", "status", objectInfo.GetObjectStatus())
		return
	}
	objectInitStatue = true
	if objectInfo.GetOwner() != account {
		ownerObject = false
		log.Warnw("object owner is not equal to account", "owner", objectInfo.GetOwner(), "account", account)
		return
	}
	ownerObject = true

	// TODO:: check payment address whether in arrears status
	paymentEnough = true
	// TODO: sp should be operator address
	/*
		if bucketInfo.GetPrimarySpAddress() == sp {
			spBucket = true
		} else {
			log.Warnw("object sp is not equal to primary sp", "owner_sp", bucketInfo.GetPrimarySpAddress(), "sp", sp)
			spBucket = false
		}
	*/
	spBucket = true
	return
}

// AuthDownloadObjectWithAccount verify the greenfield chain information for download object.
func (greenfield *Greenfield) AuthDownloadObjectWithAccount(ctx context.Context, bucket, object, account, sp string) (
	accountExist bool, bucketExist bool, objectServiceStatue bool, paymentEnough bool,
	spBucket bool, bucketID uint64, readQuota int32, ownerObject bool, err error) {

	accountExist, err = greenfield.HasAccount(ctx, account)
	if err != nil || !accountExist {
		log.Warnw("failed to query account", "error", err, "is_account_exist", accountExist)
		return
	}
	var bucketInfo *storagetypes.BucketInfo
	bucketInfo, err = greenfield.QueryBucketInfo(ctx, bucket)
	if err != nil || bucketInfo == nil {
		bucketExist = false
		log.Warnw("failed to query bucket info", "error", err, "bucket_info", bucketInfo)
		return
	}
	bucketExist = true

	var objectInfo *storagetypes.ObjectInfo
	objectInfo, err = greenfield.QueryObjectInfo(ctx, bucket, object)
	if err != nil || objectInfo == nil {
		objectServiceStatue = false
		log.Warnw("failed to query object info", "error", err, "object_info", objectInfo)
		return
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_IN_SERVICE {
		objectServiceStatue = false
		log.Warnw("object status is not equal to status_in_service", "status", objectInfo.GetObjectStatus())
		return
	}
	objectServiceStatue = true
	if objectInfo.GetOwner() != account {
		ownerObject = false
		return
	}
	ownerObject = true

	// TODO:: check payment address whether in arrears status
	paymentEnough = true
	// TODO: sp should be operator address
	/*
		if bucketInfo.GetPrimarySpAddress() == sp {
			spBucket = true
		} else {
			spBucket = false
			log.Warnw("object sp is not equal to primary sp", "owner_sp", bucketInfo.GetPrimarySpAddress(), "sp", sp)
		}
	*/
	spBucket = true
	bucketID = bucketInfo.Id.Uint64()
	readQuota = int32(bucketInfo.GetReadQuota())
	return
}
