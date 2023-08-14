package sqldb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	"github.com/stretchr/testify/assert"
)

func TestSpDBImpl_InsertUploadProgressSuccess(t *testing.T) {
	var objectID = uint64(10)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `upload_object_progress` (`task_state`,`global_virtual_group_id`,`task_state_description`,`error_description`,`secondary_endpoints`,`secondary_signatures`,`create_timestamp_second`,`update_timestamp_second`,`object_id`) VALUES (?,?,?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertUploadProgress(objectID)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertUploadProgressFailure(t *testing.T) {
	var objectID = uint64(10)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `upload_object_progress` (`task_state`,`global_virtual_group_id`,`task_state_description`,`error_description`,`secondary_endpoints`,`secondary_signatures`,`create_timestamp_second`,`update_timestamp_second`,`object_id`) VALUES (?,?,?,?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.InsertUploadProgress(objectID)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_DeleteUploadProgressSuccess(t *testing.T) {
	var objectID = uint64(10)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `upload_object_progress` WHERE `upload_object_progress`.`object_id` = ?").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.DeleteUploadProgress(objectID)
	assert.Nil(t, err)
}

func TestSpDBImpl_DeleteUploadProgressFailure(t *testing.T) {
	var objectID = uint64(10)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `upload_object_progress` WHERE `upload_object_progress`.`object_id` = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectCommit()
	err := s.DeleteUploadProgress(objectID)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateUploadProgressSuccess1(t *testing.T) {
	t.Log("Success case description: SecondaryEndpoints is not equal to 0")
	uploadMeta := &corespdb.UploadObjectMeta{
		ObjectID:             10,
		TaskState:            storetypes.TaskState_TASK_STATE_ALLOC_SECONDARY_DONE,
		GlobalVirtualGroupID: 1,
		SecondaryEndpoints:   []string{"a", "b"},
		SecondarySignatures:  [][]byte{[]byte("mock")},
		ErrorDescription:     "no error",
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `upload_object_progress` SET `task_state`=?,`global_virtual_group_id`=?,`task_state_description`=?,`error_description`=?,`secondary_endpoints`=?,`secondary_signatures`=?,`update_timestamp_second`=? WHERE object_id = ?").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateUploadProgress(uploadMeta)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateUploadProgressSuccess2(t *testing.T) {
	t.Log("Success case description: SecondaryEndpoints is equal to 0")
	uploadMeta := &corespdb.UploadObjectMeta{
		ObjectID:             10,
		TaskState:            storetypes.TaskState_TASK_STATE_ALLOC_SECONDARY_DONE,
		GlobalVirtualGroupID: 1,
		SecondarySignatures:  [][]byte{[]byte("mock")},
		ErrorDescription:     "no error",
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `upload_object_progress` SET `task_state`=?,`task_state_description`=?,`error_description`=?,`update_timestamp_second`=? WHERE object_id = ?").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateUploadProgress(uploadMeta)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateUploadProgressFailure1(t *testing.T) {
	t.Log("Failure case description: SecondaryEndpoints is not equal to 0 returns error")
	uploadMeta := &corespdb.UploadObjectMeta{
		ObjectID:             10,
		TaskState:            storetypes.TaskState_TASK_STATE_ALLOC_SECONDARY_DONE,
		GlobalVirtualGroupID: 1,
		SecondaryEndpoints:   []string{"a", "b"},
		SecondarySignatures:  [][]byte{[]byte("mock")},
		ErrorDescription:     "no error",
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `upload_object_progress` SET `task_state`=?,`global_virtual_group_id`=?,`task_state_description`=?,`error_description`=?,`secondary_endpoints`=?,`secondary_signatures`=?,`update_timestamp_second`=? WHERE object_id = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateUploadProgress(uploadMeta)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateUploadProgressFailure2(t *testing.T) {
	t.Log("Failure case description: SecondaryEndpoints is equal to 0 returns error")
	uploadMeta := &corespdb.UploadObjectMeta{
		ObjectID:             10,
		TaskState:            storetypes.TaskState_TASK_STATE_ALLOC_SECONDARY_DONE,
		GlobalVirtualGroupID: 1,
		SecondarySignatures:  [][]byte{[]byte("mock")},
		ErrorDescription:     "no error",
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `upload_object_progress` SET `task_state`=?,`task_state_description`=?,`error_description`=?,`update_timestamp_second`=? WHERE object_id = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateUploadProgress(uploadMeta)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_GetUploadStateSuccess(t *testing.T) {
	var objectID = uint64(1)
	u := &UploadObjectProgressTable{
		ObjectID:              objectID,
		TaskState:             2,
		GlobalVirtualGroupID:  1,
		TaskStateDescription:  "mockTaskStateDescription",
		ErrorDescription:      "no error",
		SecondaryEndpoints:    "mockEndpoint",
		SecondarySignatures:   "mockSig",
		CreateTimestampSecond: 1,
		UpdateTimestampSecond: 1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE object_id = ? ORDER BY `upload_object_progress`.`object_id` LIMIT 1").
		WithArgs(objectID).WillReturnRows(sqlmock.NewRows([]string{"object_id", "task_state", "global_virtual_group_id", "task_state_description",
		"error_description", "secondary_endpoints", "secondary_signatures", "create_timestamp_second", "update_timestamp_second"}).
		AddRow(u.ObjectID, u.TaskState, u.GlobalVirtualGroupID, u.TaskStateDescription, u.ErrorDescription, u.SecondaryEndpoints,
			u.SecondarySignatures, u.CreateTimestampSecond, u.UpdateTimestampSecond))
	result1, result2, err := s.GetUploadState(objectID)
	assert.Nil(t, err)
	assert.Equal(t, storetypes.TaskState(2), result1)
	assert.Equal(t, "no error", result2)
}

func TestSpDBImpl_GetUploadStateFailure(t *testing.T) {
	var objectID = uint64(1)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE object_id = ? ORDER BY `upload_object_progress`.`object_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result1, result2, err := s.GetUploadState(objectID)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Equal(t, storetypes.TaskState(0), result1)
	assert.Equal(t, "failed to query upload table", result2)
}

func TestSpDBImpl_GetUploadMetasToReplicateSuccess1(t *testing.T) {
	t.Log("Success case description: not expired")
	var (
		limit         = 1
		timeoutSecond = GetCurrentUnixTime()
		u             = &UploadObjectProgressTable{
			ObjectID:              10,
			TaskState:             2,
			GlobalVirtualGroupID:  1,
			TaskStateDescription:  "mockTaskStateDescription",
			ErrorDescription:      "no error",
			SecondaryEndpoints:    "mockEndpoint",
			SecondarySignatures:   "mockSig",
			CreateTimestampSecond: 1,
			UpdateTimestampSecond: 1,
		}
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE task_state IN (?,?) ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "task_state", "global_virtual_group_id", "task_state_description",
			"error_description", "secondary_endpoints", "secondary_signatures", "create_timestamp_second", "update_timestamp_second"}).AddRow(
			u.ObjectID, u.TaskState, u.GlobalVirtualGroupID, u.TaskStateDescription, u.ErrorDescription, u.SecondaryEndpoints,
			u.SecondarySignatures, u.CreateTimestampSecond, u.UpdateTimestampSecond))
	result, err := s.GetUploadMetasToReplicate(limit, timeoutSecond)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetUploadMetasToReplicateSuccess2(t *testing.T) {
	t.Log("Success case description: expired")
	var (
		limit         = 1
		timeoutSecond = int64(1)
		u             = &UploadObjectProgressTable{
			ObjectID:              10,
			TaskState:             2,
			GlobalVirtualGroupID:  1,
			TaskStateDescription:  "mockTaskStateDescription",
			ErrorDescription:      "no error",
			SecondaryEndpoints:    "mockEndpoint",
			SecondarySignatures:   "mockSig",
			CreateTimestampSecond: 1,
			UpdateTimestampSecond: 1,
		}
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE task_state IN (?,?) ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "task_state", "global_virtual_group_id", "task_state_description",
			"error_description", "secondary_endpoints", "secondary_signatures", "create_timestamp_second", "update_timestamp_second"}).AddRow(
			u.ObjectID, u.TaskState, u.GlobalVirtualGroupID, u.TaskStateDescription, u.ErrorDescription, u.SecondaryEndpoints,
			u.SecondarySignatures, u.CreateTimestampSecond, u.UpdateTimestampSecond))
	result, err := s.GetUploadMetasToReplicate(limit, timeoutSecond)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))
}

func TestSpDBImpl_GetUploadMetasToReplicateFailure(t *testing.T) {
	var (
		limit         = 1
		timeoutSecond = int64(1)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE task_state IN (?,?) ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetUploadMetasToReplicate(limit, timeoutSecond)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetUploadMetasToSealSuccess1(t *testing.T) {
	t.Log("Success case description: not expired")
	var (
		limit         = 1
		timeoutSecond = GetCurrentUnixTime()
		u             = &UploadObjectProgressTable{
			ObjectID:              10,
			TaskState:             2,
			GlobalVirtualGroupID:  1,
			TaskStateDescription:  "mockTaskStateDescription",
			ErrorDescription:      "no error",
			SecondaryEndpoints:    "mockEndpoint",
			SecondarySignatures:   "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
			CreateTimestampSecond: 1,
			UpdateTimestampSecond: 1,
		}
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE task_state IN (?,?) ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "task_state", "global_virtual_group_id", "task_state_description",
			"error_description", "secondary_endpoints", "secondary_signatures", "create_timestamp_second", "update_timestamp_second"}).AddRow(
			u.ObjectID, u.TaskState, u.GlobalVirtualGroupID, u.TaskStateDescription, u.ErrorDescription, u.SecondaryEndpoints,
			u.SecondarySignatures, u.CreateTimestampSecond, u.UpdateTimestampSecond))
	result, err := s.GetUploadMetasToSeal(limit, timeoutSecond)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetUploadMetasToSealSuccess2(t *testing.T) {
	t.Log("Success case description: expired")
	var (
		limit         = 1
		timeoutSecond = int64(1)
		u             = &UploadObjectProgressTable{
			ObjectID:              10,
			TaskState:             2,
			GlobalVirtualGroupID:  1,
			TaskStateDescription:  "mockTaskStateDescription",
			ErrorDescription:      "no error",
			SecondaryEndpoints:    "mockEndpoint",
			SecondarySignatures:   "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
			CreateTimestampSecond: 1,
			UpdateTimestampSecond: 1,
		}
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE task_state IN (?,?) ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "task_state", "global_virtual_group_id", "task_state_description",
			"error_description", "secondary_endpoints", "secondary_signatures", "create_timestamp_second", "update_timestamp_second"}).AddRow(
			u.ObjectID, u.TaskState, u.GlobalVirtualGroupID, u.TaskStateDescription, u.ErrorDescription, u.SecondaryEndpoints,
			u.SecondarySignatures, u.CreateTimestampSecond, u.UpdateTimestampSecond))
	result, err := s.GetUploadMetasToSeal(limit, timeoutSecond)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))
}

func TestSpDBImpl_GetUploadMetasToSealFailure1(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	var (
		limit         = 1
		timeoutSecond = int64(1)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE task_state IN (?,?) ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetUploadMetasToSeal(limit, timeoutSecond)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetUploadMetasToSealFailure2(t *testing.T) {
	t.Log("Failure case description: convert string to bytes slice returns error")
	var (
		limit         = 1
		timeoutSecond = GetCurrentUnixTime()
		u             = &UploadObjectProgressTable{
			ObjectID:              10,
			TaskState:             2,
			GlobalVirtualGroupID:  1,
			TaskStateDescription:  "mockTaskStateDescription",
			ErrorDescription:      "no error",
			SecondaryEndpoints:    "mockEndpoint",
			SecondarySignatures:   "mockSig",
			CreateTimestampSecond: 1,
			UpdateTimestampSecond: 1,
		}
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `upload_object_progress` WHERE task_state IN (?,?) ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "task_state", "global_virtual_group_id", "task_state_description",
			"error_description", "secondary_endpoints", "secondary_signatures", "create_timestamp_second", "update_timestamp_second"}).AddRow(
			u.ObjectID, u.TaskState, u.GlobalVirtualGroupID, u.TaskStateDescription, u.ErrorDescription, u.SecondaryEndpoints,
			u.SecondarySignatures, u.CreateTimestampSecond, u.UpdateTimestampSecond))
	result, err := s.GetUploadMetasToSeal(limit, timeoutSecond)
	assert.Equal(t, err.Error(), "encoding/hex: invalid byte: U+006D 'm'")
	assert.Nil(t, result)
}
