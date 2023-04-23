package bsdb

import "github.com/forbole/juno/v4/common"

// ListObjectsByBucketName list objects info by a bucket name
// sorts the objects by object ID in descending order by default, which is equivalent to sorting by create_at in descending order
func (b *BsDBImpl) ListObjectsByBucketName(bucketName string, maxKeys int, startAfter common.Hash) ([]*Object, error) {
	var (
		objects []*Object
		err     error
		limit   int
	)
	// sets the default max keys value when user didn't input maxKeys
	if maxKeys == 0 {
		maxKeys = ListObjectsDefaultMaxKeys
	}
	// return NextContinuationToken by adding 1 additionally
	limit = maxKeys + 1
	// select latest objects when user didn't input startAfter
	if startAfter == common.HexToHash("") {
		err = b.db.Table((&Object{}).TableName()).
			Select("*").
			Where("bucket_name = ?", bucketName).
			Order("object_id desc").
			Limit(limit).
			Find(&objects).Error
		return objects, err
	}
	// select objects after a specific key.
	err = b.db.Table((&Object{}).TableName()).
		Select("*").
		Where("bucket_name = ? and object_id <= ?", bucketName, startAfter).
		Order("object_id desc").
		Limit(limit).
		Find(&objects).Error
	return objects, err
}

// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
func (b *BsDBImpl) ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, isFullList bool) ([]*Object, error) {
	var (
		objects []*Object
		err     error
	)

	if isFullList {
		err = b.db.Table((&Object{}).TableName()).
			Select("*").
			Where("update_at >= ? and update_at <= ? and removed = ?", startBlockNumber, endBlockNumber, true).
			Limit(DeletedObjectsDefaultSize).
			Order("update_at,object_id asc").
			Find(&objects).Error
		return objects, err
	}
	err = b.db.Table((&Bucket{}).TableName()).
		Select("objects.*").
		Joins("left join objects on buckets.bucket_id = objects.bucket_id").
		Where("objects.update_at >= ? and objects.update_at <= ? and objects.removed = ? and "+
			"(objects.visibility='VISIBILITY_TYPE_PUBLIC_READ') or (objects.visibility='VISIBILITY_TYPE_INHERIT' and buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ')",
			startBlockNumber, endBlockNumber, true).
		Limit(DeletedObjectsDefaultSize).
		Order("objects.update_at, objects.object_id asc").
		Find(&objects).Error
	return objects, err
}

// GetObjectByName get object info by an object name
func (b *BsDBImpl) GetObjectByName(objectName string, bucketName string, isFullList bool) (*Object, error) {
	var (
		object *Object
		err    error
	)

	if isFullList {
		err = b.db.Table((&Object{}).TableName()).
			Select("*").
			Where("object_name = ? and bucket_name = ?", objectName, bucketName).
			Take(&object).Error
		return object, err
	}

	err = b.db.Table((&Bucket{}).TableName()).
		Select("objects.*").
		Joins("left join objects on buckets.bucket_id = objects.bucket_id").
		Where("objects.object_name = ? and objects.bucket_name = ? and "+
			"(objects.visibility='VISIBILITY_TYPE_PUBLIC_READ') or (objects.visibility='VISIBILITY_TYPE_INHERIT' and buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ')",
			objectName, bucketName).
		Take(&object).Error
	return object, err
}
