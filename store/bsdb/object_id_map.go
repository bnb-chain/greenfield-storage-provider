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
