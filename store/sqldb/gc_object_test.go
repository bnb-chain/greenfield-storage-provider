package sqldb

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

const taskKey = "mockTaskKey"

func TestSpDBImpl_InsertGCObjectProgressSuccess(t *testing.T) {
	gcMeta := &spdb.GCObjectMeta{
		TaskKey:          taskKey,
		StartBlockHeight: 1,
		EndBlockHeight:   10,
	}
	gcTable := &GCObjectProgressTable{
		TaskKey:               gcMeta.TaskKey,
		StartGCBlockID:        gcMeta.StartBlockHeight,
		EndGCBlockID:          gcMeta.EndBlockHeight,
		CurrentGCBlockID:      0,
		LastDeletedObjectID:   0,
		CreateTimestampSecond: GetCurrentUnixTime(),
		UpdateTimestampSecond: GetCurrentUnixTime(),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `gc_object_progress` (`task_key`,`start_gc_block_id`,`end_gc_block_id`,`current_gc_block_id`,"+
		"`last_deleted_object_id`,`create_timestamp_second`,`update_timestamp_second`) VALUES (?,?,?,?,?,?,?)").
		WithArgs(gcTable.TaskKey, gcTable.StartGCBlockID, gcTable.EndGCBlockID, gcTable.CurrentGCBlockID, gcTable.LastDeletedObjectID,
			gcTable.CreateTimestampSecond, gcTable.UpdateTimestampSecond).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertGCObjectProgress(gcMeta)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertGCObjectProgressFailure(t *testing.T) {
	gcMeta := &spdb.GCObjectMeta{
		TaskKey:          taskKey,
		StartBlockHeight: 1,
		EndBlockHeight:   10,
	}
	gcTable := &GCObjectProgressTable{
		TaskKey:               gcMeta.TaskKey,
		StartGCBlockID:        gcMeta.StartBlockHeight,
		EndGCBlockID:          gcMeta.EndBlockHeight,
		CurrentGCBlockID:      0,
		LastDeletedObjectID:   0,
		CreateTimestampSecond: GetCurrentUnixTime(),
		UpdateTimestampSecond: GetCurrentUnixTime(),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `gc_object_progress` (`task_key`,`start_gc_block_id`,`end_gc_block_id`,`current_gc_block_id`,"+
		"`last_deleted_object_id`,`create_timestamp_second`,`update_timestamp_second`) VALUES (?,?,?,?,?,?,?)").
		WithArgs(gcTable.TaskKey, gcTable.StartGCBlockID, gcTable.EndGCBlockID, gcTable.CurrentGCBlockID, gcTable.LastDeletedObjectID,
			gcTable.CreateTimestampSecond, gcTable.UpdateTimestampSecond).WillReturnError(errors.New("failed to insert"))
	mock.ExpectCommit()
	err := s.InsertGCObjectProgress(gcMeta)
	assert.Contains(t, err.Error(), "failed to insert gc record: failed to insert")
}

func TestSpDBImpl_DeleteGCObjectProgress(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `gc_object_progress` WHERE `gc_object_progress`.`task_key` = ?").WithArgs(taskKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.DeleteGCObjectProgress(taskKey)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateGCObjectProgressSuccess(t *testing.T) {
	gcMeta := &spdb.GCObjectMeta{
		TaskKey:             taskKey,
		CurrentBlockHeight:  20,
		LastDeletedObjectID: 15,
	}
	gcTable := &GCObjectProgressTable{
		TaskKey:               gcMeta.TaskKey,
		CurrentGCBlockID:      gcMeta.CurrentBlockHeight,
		LastDeletedObjectID:   gcMeta.LastDeletedObjectID,
		UpdateTimestampSecond: GetCurrentUnixTime(),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `gc_object_progress` SET `current_gc_block_id`=?,`last_deleted_object_id`=?,`update_timestamp_second`=? WHERE task_key = ?").
		WithArgs(gcTable.CurrentGCBlockID, gcTable.LastDeletedObjectID, gcTable.UpdateTimestampSecond, gcTable.TaskKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateGCObjectProgress(gcMeta)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateGCObjectProgressFailure(t *testing.T) {
	gcMeta := &spdb.GCObjectMeta{
		TaskKey:             taskKey,
		CurrentBlockHeight:  20,
		LastDeletedObjectID: 15,
	}
	gcTable := &GCObjectProgressTable{
		TaskKey:               gcMeta.TaskKey,
		CurrentGCBlockID:      gcMeta.CurrentBlockHeight,
		LastDeletedObjectID:   gcMeta.LastDeletedObjectID,
		UpdateTimestampSecond: GetCurrentUnixTime(),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `gc_object_progress` SET `current_gc_block_id`=?,`last_deleted_object_id`=?,`update_timestamp_second`=? WHERE task_key = ?").
		WithArgs(gcTable.CurrentGCBlockID, gcTable.LastDeletedObjectID, gcTable.UpdateTimestampSecond, gcTable.TaskKey).
		WillReturnError(errors.New("failed to update"))
	mock.ExpectCommit()
	err := s.UpdateGCObjectProgress(gcMeta)
	assert.Contains(t, err.Error(), "failed to update gc task record: failed to update")
}

func TestSpDBImpl_GetGCMetasToGCSuccess(t *testing.T) {
	gcMeta1 := GCObjectProgressTable{
		TaskKey:               taskKey,
		StartGCBlockID:        10,
		EndGCBlockID:          100,
		CurrentGCBlockID:      30,
		LastDeletedObjectID:   20,
		CreateTimestampSecond: GetCurrentUnixTime(),
		UpdateTimestampSecond: GetCurrentUnixTime(),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `gc_object_progress` ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"task_key", "start_gc_block_id", "end_gc_block_id", "current_gc_block_id",
			"last_deleted_object_id", "create_timestamp_second", "update_timestamp_second"}).AddRow(gcMeta1.TaskKey, gcMeta1.StartGCBlockID,
			gcMeta1.EndGCBlockID, gcMeta1.CurrentGCBlockID, gcMeta1.LastDeletedObjectID, gcMeta1.CreateTimestampSecond, gcMeta1.UpdateTimestampSecond))
	limit := 1
	data, err := s.GetGCMetasToGC(limit)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(data))
}

func TestSpDBImpl_GetGCMetasToGCFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `gc_object_progress` ORDER BY update_timestamp_second DESC LIMIT 1").
		WillReturnError(errors.New("failed to query"))
	limit := 1
	data, err := s.GetGCMetasToGC(limit)
	assert.Contains(t, err.Error(), "failed to query gc table: failed to query")
	assert.Equal(t, 0, len(data))
}
