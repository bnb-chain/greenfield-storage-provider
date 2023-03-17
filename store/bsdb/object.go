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
			Order("update_at asc").
			Find(&objects).Error
		return objects, err
	}
	err = b.db.Table((&Object{}).TableName()).
		Select("*").
		Where("update_at >= ? and update_at <= ? and removed = ? and is_public = ?", startBlockNumber, endBlockNumber, true, true).
		Limit(DeletedObjectsDefaultSize).
		Order("update_at asc").
		Find(&objects).Error
	return objects, err
}
