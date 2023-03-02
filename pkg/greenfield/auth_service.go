package greenfield

import (
	"context"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// AuthUploadObjectWithAccount verify the greenfield chain information for upload object.
func (greenfield *Greenfield) AuthUploadObjectWithAccount(ctx context.Context, bucket, object, account, sp string) (
	accountExist bool, bucketExist bool, isInitStatus bool, paymentEnough bool,
	isSpBucket bool, ownerObject bool, err error) {
	accountExist, err = greenfield.HasAccount(ctx, account)
	if err != nil || !accountExist {
		log.Errorw("failed to query account", "bucket", bucket,
			"object", object, "account_exist", accountExist, "error", err)
		return
	}
	var bucketInfo *storagetypes.BucketInfo
	bucketInfo, err = greenfield.QueryBucketInfo(ctx, bucket)
	if err != nil || bucketInfo == nil {
		bucketExist = false
		log.Errorw("failed to query bucket info", "bucket", bucket,
			"object", object, "error", err)
		return
	}
	bucketExist = true

	var objectInfo *storagetypes.ObjectInfo
	objectInfo, err = greenfield.QueryObjectInfo(ctx, bucket, object)
	if err != nil || objectInfo == nil {
		isInitStatus = false
		log.Errorw("failed to query object info", "bucket", bucket,
			"object", object, "error", err)
		return
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_INIT {
		isInitStatus = false
		log.Errorw("object status is not equal to status_init",
			"status", objectInfo.GetObjectStatus())
		return
	}
	isInitStatus = true
	if objectInfo.GetOwner() != account {
		ownerObject = false
		log.Errorw("object owner is not equal to account",
			"owner", objectInfo.GetOwner(), "account", account)
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
			log.Errorw("object sp is not equal to primary sp", "owner_sp", bucketInfo.GetPrimarySpAddress(), "sp", sp)
			spBucket = false
		}
	*/
	isSpBucket = true
	return
}

// AuthDownloadObjectWithAccount verify the greenfield chain information for download object.
func (greenfield *Greenfield) AuthDownloadObjectWithAccount(ctx context.Context, bucket, object, account, sp string) (
	accountExist bool, bucketExist bool, isServiceStatus bool, paymentEnough bool,
	isSpBucket bool, bucketID uint64, readQuota int32, ownerObject bool, err error) {

	accountExist, err = greenfield.HasAccount(ctx, account)
	if err != nil || !accountExist {
		log.Errorw("failed to query account", "bucket", bucket,
			"object", object, "account_exist", accountExist, "error", err)
		return
	}
	var bucketInfo *storagetypes.BucketInfo
	bucketInfo, err = greenfield.QueryBucketInfo(ctx, bucket)
	if err != nil || bucketInfo == nil {
		bucketExist = false
		log.Errorw("failed to query bucket info",
			"bucket", bucket, "object", object, "error", err)
		return
	}
	bucketExist = true

	var objectInfo *storagetypes.ObjectInfo
	objectInfo, err = greenfield.QueryObjectInfo(ctx, bucket, object)
	if err != nil || objectInfo == nil {
		isServiceStatus = false
		log.Errorw("failed to query object info",
			"bucket", bucket, "object", object, "error", err)
		return
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_IN_SERVICE {
		isServiceStatus = false
		log.Errorw("object not in status_in_service",
			"status", objectInfo.GetObjectStatus())
		return
	}
	isServiceStatus = true
	if objectInfo.GetOwner() != account {
		ownerObject = false
		log.Errorw("object owner mismatch", "owner", objectInfo.GetOwner(), "account", account)
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
			log.Errorw("object sp is not equal to primary sp", "owner_sp", bucketInfo.GetPrimarySpAddress(), "sp", sp)
		}
	*/
	isSpBucket = true
	bucketID = bucketInfo.Id.Uint64()
	readQuota = int32(bucketInfo.GetReadQuota())
	return
}
