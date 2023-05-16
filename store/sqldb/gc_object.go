package sqldb

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"gorm.io/gorm"
)

// SetGCObjectProgress is used to set gc object progress.
func (s *SpDBImpl) SetGCObjectProgress(taskKey string, curDeletingBlockID uint64, lastDeletedObjectID uint64) error {
	var (
		result             *gorm.DB
		insertGCObjectTask *GCObjectTaskTable
	)
	insertGCObjectTask = &GCObjectTaskTable{
		TaskKey:                taskKey,
		CurrentDeletingBlockID: curDeletingBlockID,
		LastDeletedObjectID:    lastDeletedObjectID,
	}
	result = s.db.Create(insertGCObjectTask)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert gc task record: %s", result.Error)
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
