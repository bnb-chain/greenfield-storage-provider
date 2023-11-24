package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

// GetBucketNameByObjectID get bucket name info by an object id
func (b *BsDBImpl) GetBucketNameByObjectID(objectID common.Hash) (string, error) {
	var (
		objectIdMap *ObjectIDMap
		err         error
	)

	err = b.db.Table((&ObjectIDMap{}).TableName()).
		Select("*").
		Where("object_id = ?", objectID).
		Take(&objectIdMap).Error

	return objectIdMap.BucketName, err
}

// GetLatestObjectID get latest object id
func (b *BsDBImpl) GetLatestObjectID() (uint64, error) {
	var (
		objectIdMap *ObjectIDMap
		err         error
	)

	err = b.db.Table((&ObjectIDMap{}).TableName()).
		Select("*").
		Order("object_id desc").
		Limit(1).
		Take(&objectIdMap).Error

	return objectIdMap.ObjectID.Big().Uint64(), err
}
