package database

import (
	"context"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"gorm.io/gorm"
	"time"

	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm/clause"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func (db *DB) SaveStat(ctx context.Context, stat *models.DataStat) error {
	return db.Db.Table((&models.DataStat{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "one_row_id"}},
		UpdateAll: true,
	}).Create(stat).Error
}

func (db *DB) GetObjectCount(needSeal bool, blockHeight int64) ([]int64, error) {
	result := make([]int64, 0, bsdb.ObjectsNumberOfShards)
	step := int64(1000)
	for i := 0; i < bsdb.ObjectsNumberOfShards; i++ {
		sum := int64(0)
		primaryKey := int64(0)
		count := int64(0)
		for {
			var err error
			tmpDB := db.Db.Table(bsdb.GetObjectsTableNameByShardNumber(i))
			if needSeal {
				err = tmpDB.Where("id > ? and id <= ? and status = ? and block_height <= ?", primaryKey, primaryKey+step, "OBJECT_STATUS_SEALED", blockHeight).Count(&count).Error
			} else {
				err = tmpDB.Where("id > ? and id <= ? and block_height <= ?", primaryKey, primaryKey+step, blockHeight).Count(&count).Error
			}
			if err != nil && err == gorm.ErrRecordNotFound {
				break
			}
			if err != nil {
				log.Errorw("failed to get object count", "error", err, "left", primaryKey, "right", primaryKey+step)
				return result, err
			}
			sum += count
			primaryKey += step
			time.Sleep(20 * time.Millisecond)
		}
		result = append(result, sum)
	}
	return result, nil
}

func (db *DB) SaveBlockResult(ctx context.Context, br *models.BlockResult) error {
	return db.Db.Table((&models.BlockResult{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "block_height"}},
		UpdateAll: true,
	}).Create(br).Error
}

func (db *DB) DeleteBlockResult(ctx context.Context, blockHeight int64) error {
	return db.Db.Table((&models.BlockResult{}).TableName()).Where("block_height < ?", blockHeight).Delete(&models.BlockResult{}).Error
}
