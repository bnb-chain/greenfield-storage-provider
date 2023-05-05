package bsdb

// ListObjectsByBucketName list objects info by a bucket name
// sorts the objects by object ID in descending order by default, which is equivalent to sorting by create_at in descending order
func (b *BsDBImpl) ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter string, maxKeys int) ([]*Object, []string, error) {
	var (
		objects            []*Object
		err                error
		limit              int
		commonPrefixes     []string
		listObjectsResults []ListObjectsResult
	)

	// return NextContinuationToken by adding 1 additionally
	limit = maxKeys + 1
	err = b.db.Raw(
		"SELECT object_name, 'object' as result_type "+
			"FROM objects "+
			"WHERE bucket_name = ? AND object_name LIKE CONCAT(?, '%') AND object_name > IF(? = '', '', ?) AND LOCATE(?, SUBSTRING(object_name, LENGTH(?) + 1)) = 0 "+
			"UNION "+
			"SELECT DISTINCT CONCAT(SUBSTRING(object_name, 1, LENGTH(?) + LOCATE(?, SUBSTRING(object_name, LENGTH(?) + 1)) - 1), ?) as object_name, 'common_prefix' as result_type "+
			"FROM objects "+
			"WHERE bucket_name = ? AND object_name LIKE CONCAT(?, '%') AND object_name > IF(? = '', '', ?) AND LOCATE(?, SUBSTRING(object_name, LENGTH(?) + 1)) > 0 "+
			"ORDER BY object_name "+
			"LIMIT ?",
		bucketName, prefix, continuationToken, continuationToken, delimiter, prefix,
		prefix, delimiter, prefix, delimiter, bucketName, prefix, continuationToken, continuationToken, delimiter, prefix,
		limit,
	).Scan(&listObjectsResults).Error

	return objects, commonPrefixes, err
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
