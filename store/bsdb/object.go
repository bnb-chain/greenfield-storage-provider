package bsdb

import (
	"github.com/forbole/juno/v4/common"
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
			err = b.db.Table((&Object{}).TableName()).
				Select("*").
				Where("bucket_name = ?", bucketName).
				Scopes(filters...).
				Limit(limit).
				Order("object_name asc").
				Find(&results).Error
		} else {
			err = b.db.Table((&Object{}).TableName()).
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

// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
func (b *BsDBImpl) ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, includePrivate bool) ([]*Object, error) {
	var (
		objects []*Object
		err     error
	)

	if includePrivate {
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
func (b *BsDBImpl) GetObjectByName(objectName string, bucketName string, includePrivate bool) (*Object, error) {
	var (
		object *Object
		err    error
	)

	if includePrivate {
		err = b.db.Table((&Object{}).TableName()).
			Select("*").
			Where("object_name = ? and bucket_name = ? and removed = false", objectName, bucketName).
			Take(&object).Error
		return object, err
	}

	err = b.db.Table((&Bucket{}).TableName()).
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

	err = b.db.Table((&Object{}).TableName()).
		Select("*").
		Where("object_id in (?)", ids).
		Scopes(filters...).
		Find(&objects).Error
	return objects, err
}

// ListPrimaryObjects list objects by primary sp id
func (b *BsDBImpl) ListPrimaryObjects(spID uint32, startAfter common.Hash, limit int) ([]*Object, error) {
	var (
		groups      []*GlobalVirtualGroup
		localGroups []*LocalVirtualGroup
		objects     []*Object
		gvgIDs      []uint32
		lvgIDs      []uint32
		err         error
	)

	groups, err = b.ListGvgByPrimarySpID(spID)
	if err != nil || groups == nil {
		return nil, err
	}

	gvgIDs = make([]uint32, len(groups))
	for i, group := range groups {
		gvgIDs[i] = group.GlobalVirtualGroupId
	}

	localGroups, err = b.ListLvgByGvgID(gvgIDs)
	if err != nil || localGroups == nil {
		return nil, err
	}

	lvgIDs = make([]uint32, len(localGroups))
	for i, group := range localGroups {
		lvgIDs[i] = group.LocalVirtualGroupId
	}

	//TODO check the removed logic here
	objects, err = b.ListObjectsByLVGID(lvgIDs, startAfter, limit)
	return objects, err
}

// ListSecondaryObjects list objects by secondary sp id
func (b *BsDBImpl) ListSecondaryObjects(spID uint32, startAfter common.Hash, limit int) ([]*Object, error) {
	var (
		groups      []*GlobalVirtualGroup
		localGroups []*LocalVirtualGroup
		objects     []*Object
		gvgIDs      []uint32
		lvgIDs      []uint32
		err         error
	)

	groups, err = b.ListGvgBySecondarySpID(spID)
	if err != nil || groups == nil {
		return nil, err
	}

	gvgIDs = make([]uint32, len(groups))
	for i, group := range groups {
		gvgIDs[i] = group.GlobalVirtualGroupId
	}

	localGroups, err = b.ListLvgByGvgID(gvgIDs)
	if err != nil || localGroups == nil {
		return nil, err
	}

	lvgIDs = make([]uint32, len(localGroups))
	for i, group := range localGroups {
		lvgIDs[i] = group.LocalVirtualGroupId
	}

	objects, err = b.ListObjectsByLVGID(lvgIDs, startAfter, limit)
	return objects, err
}

// ListObjectsInGVG list objects by gvg id
func (b *BsDBImpl) ListObjectsInGVG(gvgID uint32, startAfter common.Hash, limit int) ([]*Object, error) {
	var (
		localGroups []*LocalVirtualGroup
		objects     []*Object
		gvgIDs      []uint32
		lvgIDs      []uint32
		err         error
	)
	gvgIDs = append(gvgIDs, gvgID)

	localGroups, err = b.ListLvgByGvgID(gvgIDs)
	if err != nil || localGroups == nil {
		return nil, err
	}

	lvgIDs = make([]uint32, len(localGroups))
	for i, group := range localGroups {
		lvgIDs[i] = group.LocalVirtualGroupId
	}

	objects, err = b.ListObjectsByLVGID(lvgIDs, startAfter, limit)
	return objects, err
}

// ListObjectsByLVGID list objects by lvg id
func (b *BsDBImpl) ListObjectsByLVGID(lvgIDs []uint32, startAfter common.Hash, limit int) ([]*Object, error) {
	var (
		objects []*Object
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)

	filters = append(filters, ObjectIDStartAfterFilter(startAfter), RemovedFilter(false), WithLimit(limit))
	err = b.db.Table((&Object{}).TableName()).
		Select("*").
		Where("local_virtual_group_id in (?)", lvgIDs).
		Scopes(filters...).
		Order("object_id").
		Find(&objects).Error
	return objects, err
}
