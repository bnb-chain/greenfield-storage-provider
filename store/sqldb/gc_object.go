package sqldb

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"gorm.io/gorm"
)

// InsertGCObjectProgress is used to insert/update gc object progress.
func (s *SpDBImpl) InsertGCObjectProgress(taskKey string) error {
	if result := s.db.Create(&GCObjectProgressTable{
		TaskKey:                taskKey,
		CurrentDeletingBlockID: 0,
		LastDeletedObjectID:    0,
		CreateTimestampSecond:  GetCurrentUnixTime(),
		UpdateTimestampSecond:  GetCurrentUnixTime(),
	}); result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert gc task record: %s", result.Error)
	}
	return nil

}

// DeleteGCObjectProgress is used to delete gc object task.
func (s *SpDBImpl) DeleteGCObjectProgress(taskKey string) error {
	return s.db.Delete(&GCObjectProgressTable{
		TaskKey: taskKey, // should be the primary key
	}).Error
}

func (s *SpDBImpl) UpdateGCObjectProgress(gcMeta *spdb.GCObjectMeta) error {
	if result := s.db.Model(&GCObjectProgressTable{}).Where("task_key = ?", gcMeta.TaskKey).Updates(&GCObjectProgressTable{
		CurrentDeletingBlockID: gcMeta.CurrentBlockHeight,
		LastDeletedObjectID:    gcMeta.LastDeletedObjectID,
		UpdateTimestampSecond:  GetCurrentUnixTime(),
	}); result.Error != nil {
		return fmt.Errorf("failed to update gc task record: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) GetGCMetasToGC(limit int) ([]*spdb.GCObjectMeta, error) {
	var (
		result       *gorm.DB
		gcProgresses []GCObjectProgressTable
	)
	result = s.db.Order("update_timestamp_second DESC").Limit(limit).Find(&gcProgresses)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query gc table: %s", result.Error)
	}
	// TODO: impl
	return nil, nil
}
