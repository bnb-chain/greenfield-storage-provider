package database

import (
	"context"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// CreatePrefixTree create prefix tree nodes by input slice
func (db *DB) CreatePrefixTree(ctx context.Context, prefixTree []*bsdb.SlashPrefixTreeNode) error {
	// because the passed in prefixTree is one object, the array members have same bucketName, we can use the first one
	shardTableName := bsdb.GetPrefixesTableName(prefixTree[0].BucketName)
	err := db.Db.WithContext(ctx).Table(shardTableName).Create(&prefixTree).Error
	if err != nil {
		return err
	}
	return nil
}

// DeletePrefixTree delete prefix tree nodes by given conditions
func (db *DB) DeletePrefixTree(ctx context.Context, prefixTree []*bsdb.SlashPrefixTreeNode) error {
	if len(prefixTree) == 0 {
		return nil
	}
	// because the passed in prefixTree is one object, the array members have same bucketName, we can use the first one
	shardTableName := bsdb.GetPrefixesTableName(prefixTree[0].BucketName)
	tx := db.Db.WithContext(ctx).Table(shardTableName)
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
	err := stmt.Unscoped().Delete(&bsdb.SlashPrefixTreeNode{}).Error
	if err != nil {
		return err
	}
	return nil
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
