package bsdb

import (
	"fmt"
	"sort"
	"time"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"
	"github.com/spaolacci/murmur3"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
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
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

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
			err = b.db.Table(GetObjectsTableName(bucketName)).
				Select("*").
				Where("bucket_name = ?", bucketName).
				Scopes(filters...).
				Limit(limit).
				Order("object_name asc").
				Find(&results).Error
		} else {
			err = b.db.Table(GetObjectsTableName(bucketName)).
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

type ByUpdateAtAndObjectID []*Object

func (a ByUpdateAtAndObjectID) Len() int { return len(a) }

// Less we want to sort as ascending here
func (a ByUpdateAtAndObjectID) Less(i, j int) bool {
	if a[i].UpdateAt == a[j].UpdateAt {
		return a[i].ObjectID.Big().Uint64() < a[j].ObjectID.Big().Uint64()
	}
	return a[i].UpdateAt < a[j].UpdateAt
}
func (a ByUpdateAtAndObjectID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
func (b *BsDBImpl) ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, includePrivate bool) ([]*Object, error) {
	var (
		totalObjects []*Object
		objects      []*Object
		err          error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

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

	sort.Sort(ByUpdateAtAndObjectID(totalObjects))

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
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if includePrivate {
		err = b.db.Table(GetObjectsTableName(bucketName)).
			Select("*").
			Where("object_name = ? and bucket_name = ? and removed = false", objectName, bucketName).
			Take(&object).Error
		return object, err
	}

	err = b.db.Table(GetObjectsTableName(bucketName)).
		Select("objects.*").
		Joins("left join objects on buckets.bucket_id = objects.bucket_id").
		Where("objects.object_name = ? and objects.bucket_name = ? and objects.removed = false and "+
			"((objects.visibility='VISIBILITY_TYPE_PUBLIC_READ') or (objects.visibility='VISIBILITY_TYPE_INHERIT' and buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ'))",
			objectName, bucketName).
		Take(&object).Error
	return object, err
}

// ListObjectsByIDs list objects by object ids
func (b *BsDBImpl) ListObjectsByIDs(ids []common.Hash, includeRemoved bool) ([]*Object, error) {
	var (
		objects []*Object
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if !includeRemoved {
		filters = append(filters, RemovedFilter(includeRemoved))
	}
	for _, id := range ids {
		var object *Object
		bucketName, err := b.GetBucketNameByObjectID(id)
		if err != nil {
			log.Errorw("failed to get bucket name by object id in ListObjectsByObjectID", "error", err)
			continue
		}
		err = b.db.Table(GetObjectsTableName(bucketName)).
			Select("*").
			Where("object_id = ?", id).
			Scopes(filters...).
			Take(&object).Error
		if err != nil {
			log.Errorw("failed to get object by object id in ListObjectsByObjectID", "error", err)
			continue
		}
		objects = append(objects, object)
	}

	return objects, err
}

func GetObjectsTableName(bucketName string) string {
	return GetObjectsTableNameByShardNumber(int(GetObjectsShardNumberByBucketName(bucketName)))
}

func GetObjectsShardNumberByBucketName(bucketName string) uint32 {
	return murmur3.Sum32([]byte(bucketName)) % ObjectsNumberOfShards
}

func GetObjectsTableNameByShardNumber(shard int) string {
	return fmt.Sprintf("%s_%02d", ObjectTableName, shard)
}

// GetObjectByID get object info by object id
func (b *BsDBImpl) GetObjectByID(objectID int64, includeRemoved bool) (*Object, error) {
	var (
		object       *Object
		err          error
		objectIDHash common.Hash
		bucketName   string
		filters      []func(*gorm.DB) *gorm.DB
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	objectIDHash = common.BigToHash(math.NewInt(objectID).BigInt())
	if !includeRemoved {
		filters = append(filters, RemovedFilter(includeRemoved))
	}

	bucketName, err = b.GetBucketNameByObjectID(objectIDHash)
	if err != nil {
		log.Errorw("failed to get bucket name by object id in GetObjectByID", "error", err)
		return nil, err
	}

	err = b.db.Table(GetObjectsTableName(bucketName)).
		Select("*").
		Where("object_id  = ?", objectIDHash).
		Scopes(filters...).
		Take(&object).Error
	return object, err
}

// ListObjectsInGVGAndBucket list objects by gvg and bucket id
func (b *BsDBImpl) ListObjectsInGVGAndBucket(bucketID common.Hash, gvgID uint32, startAfter common.Hash, limit int) ([]*Object, *Bucket, error) {
	var (
		localGroups []*LocalVirtualGroup
		objects     []*Object
		gvgIDs      []uint32
		lvgIDs      []uint32
		bucket      *Bucket
		err         error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	gvgIDs = append(gvgIDs, gvgID)

	localGroups, err = b.ListLvgByGvgAndBucketID(bucketID, gvgIDs)
	if err != nil || len(localGroups) == 0 {
		return nil, nil, err
	}

	lvgIDs = make([]uint32, len(localGroups))
	for i, group := range localGroups {
		lvgIDs[i] = group.LocalVirtualGroupId
	}

	objects, bucket, err = b.ListObjectsByLVGID(lvgIDs, bucketID, startAfter, limit)
	return objects, bucket, err
}

// ListObjectsByGVGAndBucketForGC list objects by gvg and bucket for gc
func (b *BsDBImpl) ListObjectsByGVGAndBucketForGC(bucketID common.Hash, gvgID uint32, startAfter common.Hash, limit int) ([]*Object, *Bucket, error) {
	var (
		localGroups   []*LocalVirtualGroup
		objects       []*Object
		gvgIDs        []uint32
		lvgIDs        []uint32
		completeEvent *EventCompleteMigrationBucket
		bucket        *Bucket
		filters       []func(*gorm.DB) *gorm.DB
		err           error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	gvgIDs = append(gvgIDs, gvgID)

	localGroups, err = b.ListLvgByGvgAndBucketID(bucketID, gvgIDs)
	if err != nil || len(localGroups) == 0 {
		return nil, nil, err
	}

	lvgIDs = make([]uint32, len(localGroups))
	for i, group := range localGroups {
		lvgIDs[i] = group.LocalVirtualGroupId
	}

	completeEvent, err = b.GetMigrateBucketEventByBucketID(bucketID)
	if err != nil {
		return nil, nil, err
	}
	if completeEvent != nil {
		filters = append(filters, CreateAtFilter(completeEvent.CreateAt))
	}

	objects, bucket, err = b.ListObjectsByLVGID(lvgIDs, bucketID, startAfter, limit, filters...)
	return objects, bucket, err
}

// ListObjectsByLVGID list objects by lvg id
func (b *BsDBImpl) ListObjectsByLVGID(lvgIDs []uint32, bucketID common.Hash, startAfter common.Hash, limit int, filters ...func(*gorm.DB) *gorm.DB) ([]*Object, *Bucket, error) {
	var (
		bucket  *Bucket
		objects []*Object
		err     error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	bucket, err = b.GetBucketByID(bucketID.Big().Int64(), true)
	if err != nil {
		log.Errorw("failed to get bucket name by bucket id in ListObjectsByLVGID", "error", err)
		return nil, nil, err
	}

	filters = append(filters, ObjectIDStartAfterFilter(startAfter), RemovedFilter(false), WithLimit(limit))
	err = b.db.Table(GetObjectsTableName(bucket.BucketName)).
		Select("*").
		Where("local_virtual_group_id in (?) and bucket_id = ? and status = 'OBJECT_STATUS_SEALED'", lvgIDs, bucketID).
		Scopes(filters...).
		Order("object_id").
		Find(&objects).Error

	return objects, bucket, err
}

// ListObjectsInGVG list objects by gvg and bucket id
func (b *BsDBImpl) ListObjectsInGVG(gvgID uint32, startAfter common.Hash, limit int) ([]*Object, []*Bucket, error) {
	var (
		localGroups  []*LocalVirtualGroup
		objects      []*Object
		buckets      []*Bucket
		bucketIDs    []common.Hash
		bucketIDsMap map[common.Hash]bool
		tableNameMap map[common.Hash]string
		gvgIDs       []uint32
		query        string
		err          error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	gvgIDs = append(gvgIDs, gvgID)

	localGroups, err = b.ListLvgByGvgID(gvgIDs)
	if err != nil || len(localGroups) == 0 {
		return nil, nil, err
	}

	bucketIDsMap = make(map[common.Hash]bool)
	bucketIDs = make([]common.Hash, 0)
	for _, group := range localGroups {
		if _, ok := bucketIDsMap[group.BucketID]; !ok {
			bucketIDs = append(bucketIDs, group.BucketID)
			bucketIDsMap[group.BucketID] = true
		}
	}

	buckets, err = b.ListBucketsByIDs(bucketIDs, false)
	if err != nil {
		return nil, nil, err
	}

	tableNameMap = make(map[common.Hash]string)
	for _, bucket := range buckets {
		tableNameMap[bucket.BucketID] = GetObjectsTableName(bucket.BucketName)
	}

	query = fmt.Sprintf("select * from ((select * from %s where local_virtual_group_id = %d and bucket_id = %s)", tableNameMap[localGroups[0].BucketID], localGroups[0].LocalVirtualGroupId, localGroups[0].BucketID.String())
	if len(localGroups) > 1 {
		for _, group := range localGroups[1:] {
			subQuery := fmt.Sprintf(" UNION ALL (select * from %s where local_virtual_group_id = %d and bucket_id = %s)", tableNameMap[group.BucketID], group.LocalVirtualGroupId, group.BucketID.String())
			query = query + subQuery
		}
	}

	query = query + fmt.Sprintf(") as combined where status = 'OBJECT_STATUS_SEALED' and object_id > %s and  removed = false order by object_id limit %d;", startAfter.String(), limit)
	err = b.db.Table((&Object{}).TableName()).Raw(query).Find(&objects).Error

	return objects, buckets, err
}
