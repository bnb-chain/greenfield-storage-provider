package sqldb

import (
	"fmt"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"gorm.io/gorm"
)

func (s *SpDBImpl) InsertUploadProgress(objectID uint64) error {
	var result *gorm.DB
	taskState := storetypes.TaskState_TASK_STATE_INIT_UNSPECIFIED
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

func (s *SpDBImpl) UpdateUploadProgress(uploadMeta *corespdb.UploadObjectMeta) error {
	if len(uploadMeta.SecondaryAddresses) != 0 {
		if result := s.db.Model(&UploadObjectProgressTable{}).Where("object_id = ?", uploadMeta.ObjectID).
			Updates(&UploadObjectProgressTable{
				TaskState:             int32(uploadMeta.TaskState),
				TaskStateDescription:  uploadMeta.TaskState.String(),
				ErrorDescription:      uploadMeta.ErrorDescription,
				SecondaryAddresses:    util.JoinWithComma(uploadMeta.SecondaryAddresses),
				SecondarySignatures:   util.BytesSliceToString(uploadMeta.SecondarySignatures),
				UpdateTimestampSecond: GetCurrentUnixTime(),
			}); result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to update upload task record: %s", result.Error)
		}
	} else {
		if result := s.db.Model(&UploadObjectProgressTable{}).Where("object_id = ?", uploadMeta.ObjectID).
			Updates(&UploadObjectProgressTable{
				TaskState:             int32(uploadMeta.TaskState),
				TaskStateDescription:  uploadMeta.TaskState.String(),
				ErrorDescription:      uploadMeta.ErrorDescription,
				UpdateTimestampSecond: GetCurrentUnixTime(),
			}); result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to update upload task record: %s", result.Error)
		}
	}
	return nil
}

func (s *SpDBImpl) GetUploadState(objectID uint64) (storetypes.TaskState, error) {
	queryReturn := &UploadObjectProgressTable{}
	result := s.db.First(queryReturn, "object_id = ?", objectID)
	if result.Error != nil {
		return storetypes.TaskState_TASK_STATE_INIT_UNSPECIFIED, fmt.Errorf("failed to query upload table: %s", result.Error)
	}
	return storetypes.TaskState(queryReturn.TaskState), nil
}

func (s *SpDBImpl) GetUploadMetasToReplicate(limit int) ([]*corespdb.UploadObjectMeta, error) {
	var (
		result        *gorm.DB
		uploadObjects []UploadObjectProgressTable
	)
	result = s.db.Where("task_state IN ?", []string{
		util.Uint32ToString(uint32(storetypes.TaskState_TASK_STATE_UPLOAD_OBJECT_DONE)),
		util.Uint32ToString(uint32(storetypes.TaskState_TASK_STATE_REPLICATE_OBJECT_DOING)),
	}).Order("update_timestamp_second DESC").Limit(limit).Find(&uploadObjects)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query upload table: %s", result.Error)
	}
	// TODO: impl
	return nil, nil
}

func (s *SpDBImpl) GetUploadMetasToSeal(limit int) ([]*corespdb.UploadObjectMeta, error) {
	var (
		result        *gorm.DB
		uploadObjects []UploadObjectProgressTable
	)
	result = s.db.Where("task_state IN ?", []string{
		util.Uint32ToString(uint32(storetypes.TaskState_TASK_STATE_REPLICATE_OBJECT_DONE)),
		util.Uint32ToString(uint32(storetypes.TaskState_TASK_STATE_SEAL_OBJECT_DOING)),
	}).Order("update_timestamp_second DESC").Limit(limit).Find(&uploadObjects)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query upload table: %s", result.Error)
	}
	// TODO: impl
	return nil, nil
}
