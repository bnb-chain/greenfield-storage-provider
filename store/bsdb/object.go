package bsdb

import (
	"fmt"
	"github.com/forbole/juno/v4/common"
	"github.com/spaolacci/murmur3"
	"gorm.io/gorm"
	"sort"
)

const ObjectsNumberOfShards = 64

// ListObjectsByBucketName lists objects information by a bucket name.
// The function takes the following parameters:
// - bucketName: The name of the bucket to search for objects.
// - continuationToken: A token to paginate through the list of objects.
// - prefix: A prefix to filter the objects by their object names.
// - delimiter: A delimiter to group objects that share a common prefix. An empty delimiter means no grouping.
// - maxKeys: The maximum number of objects to return in the result.
//
// The function returns a slice of ListObjectsResult, which contains information about the objects and their types (object or common_prefix).
// If there is a delimiter specified, the function will group objects that share a common prefix and return them as common_prefix in the result.
// If the delimiter is empty, the function will return all objects without grouping them by a common prefix.
func (b *BsDBImpl) ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter string, maxKeys int, includeRemoved bool) ([]*ListObjectsResult, error) {
	var (
		err     error
		limit   int
		results []*ListObjectsResult
		filters []func(*gorm.DB) *gorm.DB
	)

	// return NextContinuationToken by adding 1 additionally
	limit = maxKeys + 1

	// If delimiter is specified, execute a raw SQL query to:
	// 1. Retrieve objects from the given bucket with matching prefix and continuationToken
	// 2. Find common prefixes based on the delimiter
	// 3. Limit results
	if delimiter != "" {
		results, err = b.ListObjects(bucketName, continuationToken, prefix, maxKeys)
	} else {
		// If delimiter is not specified, retrieve objects directly

		if continuationToken != "" {
			filters = append(filters, ContinuationTokenFilter(continuationToken))
		}
		if prefix != "" {
			filters = append(filters, PrefixFilter(prefix))
		}

		if includeRemoved {
			err = b.db.Scopes(ReadObjectsTable(bucketName)).
				Select("*").
				Where("bucket_name = ?", bucketName).
				Scopes(filters...).
				Limit(limit).
				Order("object_name asc").
				Find(&results).Error
		} else {
			err = b.db.Scopes(ReadObjectsTable(bucketName)).
				Select("*").
				Where("bucket_name = ? and removed = false", bucketName).
				Scopes(filters...).
				Limit(limit).
				Order("object_name asc").
				Find(&results).Error
		}
	}
	return results, err
}

type ByUpdateAtAndID []*Object

func (a ByUpdateAtAndID) Len() int { return len(a) }
func (a ByUpdateAtAndID) Less(i, j int) bool {
	if a[i].UpdateAt == a[j].UpdateAt {
		return a[i].ID < a[j].ID
	}
	return a[i].UpdateAt < a[j].UpdateAt
}
func (a ByUpdateAtAndID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
func (b *BsDBImpl) ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, includePrivate bool) ([]*Object, error) {
	var (
		totalObjects []*Object
		objects      []*Object
		err          error
	)

	if includePrivate {
		for i := 0; i < ObjectsNumberOfShards; i++ {
			err = b.db.Table(GetObjectsTableNameByShardNumber(i)).
				Select("*").
				Where("update_at >= ? and update_at <= ? and removed = ?", startBlockNumber, endBlockNumber, true).
				Limit(DeletedObjectsDefaultSize).
				Order("update_at,object_id asc").
				Find(&objects).Error
			totalObjects = append(totalObjects, objects...)
		}
	} else {
		for i := 0; i < ObjectsNumberOfShards; i++ {
			objectTableName := GetObjectsTableNameByShardNumber(i)
			joins := fmt.Sprintf("right join buckets on buckets.bucket_id = %s.bucket_id", objectTableName)
			order := fmt.Sprintf("%s.update_at, %s.object_id asc", objectTableName, objectTableName)
			where := fmt.Sprintf("%s.update_at >= ? and %s.update_at <= ? and %s.removed = ? and "+
				"((%s.visibility='VISIBILITY_TYPE_PUBLIC_READ') or "+
				"(%s.visibility='VISIBILITY_TYPE_INHERIT' and buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ'))",
				objectTableName, objectTableName, objectTableName, objectTableName, objectTableName)

			err = b.db.Table(objectTableName).
				Select(objectTableName+".*").
				Joins(joins).
				Where(where, startBlockNumber, endBlockNumber, true).
				Limit(DeletedObjectsDefaultSize).
				Order(order).
				Find(&objects).Error
			totalObjects = append(totalObjects, objects...)
		}
	}

	sort.Sort(ByUpdateAtAndID(totalObjects))

	if len(totalObjects) > DeletedObjectsDefaultSize {
		totalObjects = totalObjects[0:DeletedObjectsDefaultSize]
	}
	return totalObjects, err
}

// GetObjectByName get object info by an object name
func (b *BsDBImpl) GetObjectByName(objectName string, bucketName string, includePrivate bool) (*Object, error) {
	var (
		object *Object
		err    error
	)

	if includePrivate {
		err = b.db.Scopes(ReadObjectsTable(bucketName)).
			Select("*").
			Where("object_name = ? and bucket_name = ? and removed = false", objectName, bucketName).
			Take(&object).Error
		return object, err
	}

	err = b.db.Scopes(ReadObjectsTable(bucketName)).
		Select("objects.*").
		Joins("left join objects on buckets.bucket_id = objects.bucket_id").
		Where("objects.object_name = ? and objects.bucket_name = ? and objects.removed = false and "+
			"((objects.visibility='VISIBILITY_TYPE_PUBLIC_READ') or (objects.visibility='VISIBILITY_TYPE_INHERIT' and buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ'))",
			objectName, bucketName).
		Take(&object).Error
	return object, err
}

// ListObjectsByObjectID list objects by object ids
func (b *BsDBImpl) ListObjectsByObjectID(ids []common.Hash, includeRemoved bool) ([]*Object, error) {
	var (
		objects []*Object
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)

	if !includeRemoved {
		filters = append(filters, RemovedFilter(includeRemoved))
	}
	//
	//for idx, id := range ids {
	//	getobjectbyid
	//}

	err = b.db.Table((&Object{}).TableName()).
		Select("*").
		Where("object_id in (?)", ids).
		Scopes(filters...).
		Find(&objects).Error
	return objects, err
}

func GetObjectsTableName(bucketName string) string {
	return GetObjectsTableNameByShardNumber(int(GetObjectsShardNumberByUID(bucketName)))
}

func GetObjectsShardNumberByUID(bucketName string) uint32 {
	return murmur3.Sum32([]byte(bucketName)) % ObjectsNumberOfShards
}

func GetObjectsTableNameByShardNumber(shard int) string {
	return fmt.Sprintf("objects_%02d", shard)
}

var ReadObjectsTable = func(uid string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Table(GetObjectsTableName(uid))
	}
}
