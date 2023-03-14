package sqldb

import "github.com/bnb-chain/greenfield-storage-provider/model/metadata"

// ListObjectsByBucketName list objects info by a bucket name
func (s *SpDBImpl) ListObjectsByBucketName(bucketName string) ([]*metadata.Object, error) {
	var (
		objects []*metadata.Object
		err     error
	)

	err = s.db.Table((&metadata.Object{}).TableName()).
		Select("*").
		Where("bucket_name = ?", bucketName).
		Find(&objects).Error
	return objects, err
}

// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
func (s *SpDBImpl) ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, isFullList bool) ([]*metadata.Object, error) {
	var (
		objects []*metadata.Object
		err     error
	)

	err = s.db.Table((&metadata.Object{}).TableName()).
		Select("*").
		Where("create_at >= ? and create_at <= ? and is_public = ?", startBlockNumber, endBlockNumber, true).
		Limit(DeletedObjectsDefaultSize).
		Order("create_at asc").
		Find(&objects).Error
	return objects, err
}
