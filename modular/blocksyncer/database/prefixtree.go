package database

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// CreatePrefixTree create prefix tree nodes by input slice
func (db *DB) CreatePrefixTree(ctx context.Context, prefixTree []*bsdb.SlashPrefixTreeNode) (string, []interface{}) {
	// because the passed in prefixTree is one object, the array members have same bucketName, we can use the first one
	shardTableName := bsdb.GetPrefixesTableName(prefixTree[0].BucketName)
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table(shardTableName).Create(&prefixTree).Statement
	return stat.SQL.String(), stat.Vars
}

// DeletePrefixTree delete prefix tree nodes by given conditions
func (db *DB) DeletePrefixTree(ctx context.Context, prefixTree []*bsdb.SlashPrefixTreeNode) (string, []interface{}) {
	if len(prefixTree) == 0 {
		return "", nil
	}
	// because the passed in prefixTree is one object, the array members have same bucketName, we can use the first one
	shardTableName := bsdb.GetPrefixesTableName(prefixTree[0].BucketName)
	tx := db.Db.Session(&gorm.Session{DryRun: true}).Table(shardTableName)
	stmt := tx.Where("bucket_name = ? AND full_name = ? AND is_object = ?",
		prefixTree[0].BucketName,
		prefixTree[0].FullName,
		prefixTree[0].IsObject)

	for _, object := range prefixTree[1:] {
		stmt = stmt.Or("bucket_name = ? AND full_name = ? AND is_object = ?",
			object.BucketName,
			object.FullName,
			object.IsObject)
	}
	stmt2 := stmt.Unscoped().Delete(&bsdb.SlashPrefixTreeNode{}).Statement
	return stmt2.SQL.String(), stmt2.Vars

}

// GetPrefixTree get prefix tree node by full name and bucket name
func (db *DB) GetPrefixTree(ctx context.Context, fullName, bucketName string) (*bsdb.SlashPrefixTreeNode, error) {
	var prefixTreeNode *bsdb.SlashPrefixTreeNode
	shardTableName := bsdb.GetPrefixesTableName(bucketName)
	err := db.Db.WithContext(ctx).Table(shardTableName).
		Where("full_name = ? AND bucket_name = ? AND is_object = ?", fullName, bucketName, false).Take(&prefixTreeNode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return prefixTreeNode, nil
}

// GetPrefixTreesByBucketAndPathName get prefix tree node by bucket name and path name
func (db *DB) GetPrefixTreesByBucketAndPathName(ctx context.Context, bucketName, pathName string) ([]*bsdb.SlashPrefixTreeNode, error) {
	var prefixTreeNode []*bsdb.SlashPrefixTreeNode
	shardTableName := bsdb.GetPrefixesTableName(bucketName)
	err := db.Db.WithContext(ctx).Table(shardTableName).
		Where("bucket_name = ? AND path_name = ?", bucketName, pathName).Find(&prefixTreeNode).Error
	return prefixTreeNode, err
}

// GetPrefixTreesByBucketAndPathNameAndFullName get prefix tree node by bucket name and path name and full name
func (db *DB) GetPrefixTreesByBucketAndPathNameAndFullName(ctx context.Context, bucketName, pathName, fullName string) (*bsdb.SlashPrefixTreeNode, error) {
	var prefixTreeNode *bsdb.SlashPrefixTreeNode
	shardTableName := bsdb.GetPrefixesTableName(bucketName)
	err := db.Db.WithContext(ctx).Table(shardTableName).
		Where("bucket_name = ? AND path_name = ? and full_name = ?", bucketName, pathName, fullName).Take(&prefixTreeNode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return prefixTreeNode, nil
}

// GetPrefixTreeObject get prefix tree node object by object id
func (db *DB) GetPrefixTreeObject(ctx context.Context, objectID common.Hash, bucketName string) (*bsdb.SlashPrefixTreeNode, error) {
	var prefixTreeNode *bsdb.SlashPrefixTreeNode
	shardTableName := bsdb.GetPrefixesTableName(bucketName)
	err := db.Db.WithContext(ctx).Table(shardTableName).
		Where("object_id = ?", objectID).Take(&prefixTreeNode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return prefixTreeNode, nil
}

// GetPrefixTreeCount get prefix tree nodes count by path and bucket name
func (db *DB) GetPrefixTreeCount(ctx context.Context, pathName, bucketName string) (int64, error) {
	var count int64
	shardTableName := bsdb.GetPrefixesTableName(bucketName)
	err := db.Db.WithContext(ctx).Table(shardTableName).Where("bucket_name = ? AND path_name = ?", bucketName, pathName).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
