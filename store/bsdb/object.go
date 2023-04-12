package bsdb

// ListObjectsByBucketName list objects info by a bucket name
func (b *BsDBImpl) ListObjectsByBucketName(bucketName string) ([]*Object, error) {
	var (
		objects []*Object
		err     error
	)

	err = b.db.Table((&Object{}).TableName()).
		Select("*").
		Where("bucket_name = ?", bucketName).
		Order("create_at desc").
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

<<<<<<< HEAD
// GetObjectInfo get object info by an object and a bucket name
func (b *BsDBImpl) GetObjectInfo(objectName, bucketName string) (*Object, error) {
=======
// GetObjectByName get object info by an object name
func (b *BsDBImpl) GetObjectByName(objectName string, bucketName string, isFullList bool) (*Object, error) {
>>>>>>> ad1f2984 (feat: download without auth)
	var (
		object *Object
		err    error
	)

<<<<<<< HEAD
	err = b.db.Table((&Object{}).TableName()).
		Select("*").
		Where("object_name = ? and bucket_name = ?", objectName, bucketName).
		Find(&object).Error
=======
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
>>>>>>> ad1f2984 (feat: download without auth)
	return object, err
}
