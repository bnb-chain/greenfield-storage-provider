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

//func GetObjectIDMapsTableName(objectID common.Hash) string {
//	return GetObjectsTableNameByShardNumber(int(GetObjectIDMapsShardNumberByUID(objectID)))
//}
//
//func GetObjectIDMapsShardNumberByUID(objectID common.Hash) uint64 {
//	return objectID.Big().Uint64() / 4;
//}
//
//func GetObjectIDMapsTableNameByShardNumber(shard int) string {
//	return fmt.Sprintf("object_id_maps_%02d", shard)
//}

//var ReadObjectsTable = func(uid string) func(db *gorm.DB) *gorm.DB {
//	return func(db *gorm.DB) *gorm.DB {
//		return db.Table(GetObjectsTableName(uid))
//	}
//}
