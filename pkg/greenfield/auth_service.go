package greenfield

import (
	"context"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// AuthUploadObjectWithAccount verify the greenfield chain information for upload object.
func (greenfield *Greenfield) AuthUploadObjectWithAccount(ctx context.Context, bucket, object, account, sp string) (
	accountExist bool, bucketExist bool, objectInitStatue bool, paymentEnough bool,
	spBucket bool, ownerObject bool, err error) {
	accountExist, err = greenfield.HasAccount(ctx, account)
	if err != nil || !accountExist {
		return
	}
	var bucketInfo *storagetypes.BucketInfo
	bucketInfo, err = greenfield.QueryBucketInfo(ctx, bucket)
	if err != nil || bucketInfo == nil {
		bucketExist = false
		return
	}
	bucketExist = true

	var objectInfo *storagetypes.ObjectInfo
	objectInfo, err = greenfield.QueryObjectInfo(ctx, bucket, object)
	if err != nil || objectInfo == nil {
		objectInitStatue = false
		return
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_INIT {
		objectInitStatue = false
		return
	}
	if objectInfo.GetOwner() != account {
		ownerObject = false
		return
	}

	// TODO:: check payment address whether in arrears status
	paymentEnough = true

	if bucketInfo.GetPrimarySpAddress() == sp {
		spBucket = true
	} else {
		spBucket = false
	}
	return
}

// AuthDownloadObjectWithAccount verify the greenfield chain information for download object.
func (greenfield *Greenfield) AuthDownloadObjectWithAccount(ctx context.Context, bucket, object, account, sp string) (
	accountExist bool, bucketExist bool, objectServiceStatue bool, paymentEnough bool,
	spBucket bool, bucketID uint64, readQuota int32, ownerObject bool, err error) {

	accountExist, err = greenfield.HasAccount(ctx, account)
	if err != nil || !accountExist {
		return
	}
	var bucketInfo *storagetypes.BucketInfo
	bucketInfo, err = greenfield.QueryBucketInfo(ctx, bucket)
	if err != nil || bucketInfo == nil {
		bucketExist = false
		return
	}
	bucketExist = true

	var objectInfo *storagetypes.ObjectInfo
	objectInfo, err = greenfield.QueryObjectInfo(ctx, bucket, object)
	if err != nil || objectInfo == nil {
		objectServiceStatue = false
		return
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_IN_SERVICE {
		objectServiceStatue = false
		return
	}
	if objectInfo.GetOwner() != account {
		ownerObject = false
		return
	}

	// TODO:: check payment address whether in arrears status
	paymentEnough = true

	if bucketInfo.GetPrimarySpAddress() == sp {
		spBucket = true
	} else {
		spBucket = false
	}

	bucketID = bucketInfo.Id.Uint64()
	readQuota = int32(bucketInfo.GetReadQuota())
	return
}
