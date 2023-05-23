package bsdb

import (
	"gorm.io/gorm"
)

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
func (b *BsDBImpl) ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter string, maxKeys int) ([]*ListObjectsResult, error) {
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
		err = b.db.Raw(
			`SELECT path_name, result_type, o.*
				FROM (
					SELECT DISTINCT object_name as path_name, 'object' as result_type, id
					FROM objects
					WHERE bucket_name = ? AND object_name LIKE ? AND object_name >= IF(? = '', '', ?) AND LOCATE(?, SUBSTRING(object_name, LENGTH(?) + 1)) = 0
					UNION
					SELECT CONCAT(SUBSTRING(object_name, 1, LENGTH(?) + LOCATE(?, SUBSTRING(object_name, LENGTH(?) + 1)) - 1), ?) as path_name, 'common_prefix' as result_type, MIN(id)
					FROM objects
					WHERE bucket_name = ? AND object_name LIKE ? AND object_name >= IF(? = '', '', ?) AND LOCATE(?, SUBSTRING(object_name, LENGTH(?) + 1)) > 0
					GROUP BY path_name
				) AS subquery
				JOIN objects o ON subquery.id = o.id
				ORDER BY path_name
				LIMIT ?;`,
			bucketName, prefix+"%", continuationToken, continuationToken, delimiter, prefix,
			prefix, delimiter, prefix, delimiter,
			bucketName, prefix+"%", continuationToken, continuationToken, delimiter, prefix,
			limit).Scan(&results).Error
	} else {
		// If delimiter is not specified, retrieve objects directly

		if continuationToken != "" {
			filters = append(filters, ContinuationTokenFilter(continuationToken))
		}
		if prefix != "" {
			filters = append(filters, PrefixFilter(prefix))
		}

		err = b.db.Table((&Object{}).TableName()).
			Select("*").
			Where("bucket_name = ?", bucketName).
			Scopes(filters...).
			Limit(limit).
			Order("object_name asc").
			Find(&results).Error
	}
	return results, err
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
			"((objects.visibility='VISIBILITY_TYPE_PUBLIC_READ') or (objects.visibility='VISIBILITY_TYPE_INHERIT' and buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ'))",
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
			"((objects.visibility='VISIBILITY_TYPE_PUBLIC_READ') or (objects.visibility='VISIBILITY_TYPE_INHERIT' and buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ'))",
			objectName, bucketName).
		Take(&object).Error
	return object, err
}
