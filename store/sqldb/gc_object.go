package sqldb

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/core/task"
)

// SetGCObjectProgress is used to insert/update gc object progress.
func (s *SpDBImpl) SetGCObjectProgress(taskKey string, curDeletingBlockID uint64, lastDeletedObjectID uint64) error {
	var (
		result      *gorm.DB
		queryReturn *GCObjectTaskTable
	)

	queryReturn = &GCObjectTaskTable{}
	result = s.db.First(queryReturn, "task_key = ?", taskKey)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		result = s.db.Create(&GCObjectTaskTable{
			TaskKey:                taskKey,
			CurrentDeletingBlockID: curDeletingBlockID,
			LastDeletedObjectID:    lastDeletedObjectID,
		})
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert gc task record: %s", result.Error)
		}
		return nil
	}
	if result.Error != nil {
		return fmt.Errorf("failed to set gc task record: %s", result.Error)
	}
	if queryReturn.CurrentDeletingBlockID == curDeletingBlockID &&
		queryReturn.LastDeletedObjectID == lastDeletedObjectID {
		return nil
	}
	result = s.db.Model(&GCObjectTaskTable{}).Where("task_key = ?", taskKey).Updates(&GCObjectTaskTable{
		CurrentDeletingBlockID: curDeletingBlockID,
		LastDeletedObjectID:    lastDeletedObjectID,
	})
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to update gc task record: %s", result.Error)
	}
	return nil
}

// DeleteGCObjectProgress is used to delete gc object task.
func (s *SpDBImpl) DeleteGCObjectProgress(taskKey string) error {
	return s.db.Delete(&GCObjectTaskTable{
		TaskKey: taskKey, // should be the primary key
	}).Error
}

// GetAllGCObjectTask is unused.
// TODO: will be implemented in the future, may be used in startup.
func (s *SpDBImpl) GetAllGCObjectTask(taskKey string) []task.GCObjectTask { return nil }
