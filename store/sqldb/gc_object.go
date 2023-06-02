package sqldb

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
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

// GetAllGCObjectTask is unused.
// TODO: will be implemented in the future, may be used in startup.
func (s *SpDBImpl) GetAllGCObjectTask(taskKey string) []task.GCObjectTask { return nil }
