package sqldb

import (
	"fmt"

	servicetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	"gorm.io/gorm"
)

func (s *SpDBImpl) InsertUploadProgress(objectID uint64) error {
	var result *gorm.DB
	taskState := servicetypes.TaskState_TASK_STATE_INIT_UNSPECIFIED
	result = s.db.Create(&UploadObjectProgressTable{
		ObjectID:              objectID,
		TaskState:             int32(taskState),
		TaskStateDescription:  taskState.String(),
		CreateTimestampSecond: GetCurrentUnixTime(),
		UpdateTimestampSecond: GetCurrentUnixTime(),
	})
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert upload task record: %s", result.Error)
	}
	return nil
}
func (s *SpDBImpl) DeleteUploadProgress(objectID uint64) error {
	return s.db.Delete(&UploadObjectProgressTable{
		ObjectID: objectID, // should be the primary key
	}).Error
}

func (s *SpDBImpl) UpdateUploadProgress(objectID uint64, taskState servicetypes.TaskState, errorDescription string) error {
	if result := s.db.Model(&UploadObjectProgressTable{}).Where("object_id = ?", objectID).Updates(&UploadObjectProgressTable{
		TaskState:             int32(taskState),
		TaskStateDescription:  taskState.String(),
		ErrorDescription:      errorDescription,
		UpdateTimestampSecond: GetCurrentUnixTime(),
	}); result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to update upload task record: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) GetUploadState(objectID uint64) (servicetypes.TaskState, error) {
	queryReturn := &UploadObjectProgressTable{}
	result := s.db.First(queryReturn, "object_id = ?", objectID)
	if result.Error != nil {
		return servicetypes.TaskState_TASK_STATE_INIT_UNSPECIFIED, fmt.Errorf("failed to query upload table: %s", result.Error)
	}
	return servicetypes.TaskState(queryReturn.TaskState), nil
}
