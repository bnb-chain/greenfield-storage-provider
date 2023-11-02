package sqldb

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const (
	mockBlockHeight                       = 10
	mockSwapOutKey                        = "mock_swap_out_key"
	mockMigrateSubscribeProgressQuerySQL  = "SELECT * FROM `migrate_subscribe_progress` WHERE event_name = ? ORDER BY `migrate_subscribe_progress`.`event_name` LIMIT 1"
	mockMigrateSubscribeProgressInsertSQL = "INSERT INTO `migrate_subscribe_progress` (`event_name`,`last_subscribed_block_height`) VALUES (?,?)"
	mockMigrateSubscribeProgressUpdateSQL = "UPDATE `migrate_subscribe_progress` SET `event_name`=?,`last_subscribed_block_height`=? WHERE event_name = ?"
	mockQuerySwapOutUnitInSrcSPQuerySQL   = "SELECT * FROM `swap_out_unit` WHERE is_dest_sp = false and swap_out_key = ? ORDER BY `swap_out_unit`.`swap_out_key` LIMIT 1"
	mockListDestSPSwapOutUintsQuerySQL    = "SELECT * FROM `swap_out_unit` WHERE is_dest_sp = true"
	mockMigrateGVGQuerySQL                = "SELECT * FROM `migrate_gvg` WHERE migrate_key = ? ORDER BY `migrate_gvg`.`migrate_key` LIMIT 1"
	mockMigrateGVGDeleteSQL               = "DELETE FROM `migrate_gvg` WHERE `migrate_gvg`.`migrate_key` = ?"
)

func TestSpDBImpl_UpdateSPExitSubscribeProgressInsertSuccess(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SPExitProgressKey).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressInsertSQL).WithArgs(SPExitProgressKey, mockBlockHeight).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateSPExitSubscribeProgress(mockBlockHeight)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateSPExitSubscribeProgressUpdateSuccess(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SPExitProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(SPExitProgressKey, 5))
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressUpdateSQL).WithArgs(SPExitProgressKey, mockBlockHeight, SPExitProgressKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateSPExitSubscribeProgress(mockBlockHeight)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateSPExitSubscribeProgressQueryFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SPExitProgressKey).WillReturnError(mockDBInternalError)
	err := s.UpdateSPExitSubscribeProgress(mockBlockHeight)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateSPExitSubscribeProgressInsertFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SPExitProgressKey).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressInsertSQL).WithArgs(SPExitProgressKey, mockBlockHeight).
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateSPExitSubscribeProgress(mockBlockHeight)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateSPExitSubscribeProgressUpdateFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SPExitProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(SPExitProgressKey, 5))
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressUpdateSQL).WithArgs(SPExitProgressKey, mockBlockHeight, SPExitProgressKey).
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateSPExitSubscribeProgress(mockBlockHeight)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_QuerySPExitSubscribeProgressSuccess(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SPExitProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(SPExitProgressKey, mockBlockHeight))
	result, err := s.QuerySPExitSubscribeProgress()
	assert.Nil(t, err)
	assert.Equal(t, uint64(mockBlockHeight), result)
}

func TestSpDBImpl_QuerySPExitSubscribeProgressRecordNotFound(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WillReturnError(gorm.ErrRecordNotFound)
	result, err := s.QuerySPExitSubscribeProgress()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), result)
}

func TestSpDBImpl_QuerySPExitSubscribeProgressFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WillReturnError(mockDBInternalError)
	result, err := s.QuerySPExitSubscribeProgress()
	assert.Equal(t, mockDBInternalError, err)
	assert.Equal(t, uint64(0), result)
}

func TestSpDBImpl_UpdateSwapOutSubscribeProgressInsertSuccess(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SwapOutProgressKey).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressInsertSQL).WithArgs(SwapOutProgressKey, mockBlockHeight).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateSwapOutSubscribeProgress(mockBlockHeight)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateSwapOutSubscribeProgressUpdateSuccess(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SwapOutProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(SwapOutProgressKey, 5))
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressUpdateSQL).WithArgs(SwapOutProgressKey, mockBlockHeight, SwapOutProgressKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateSwapOutSubscribeProgress(mockBlockHeight)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateSwapOutSubscribeProgressQueryFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SwapOutProgressKey).WillReturnError(mockDBInternalError)
	err := s.UpdateSwapOutSubscribeProgress(mockBlockHeight)
	assert.Equal(t, mockDBInternalError, err)
}

func TestSpDBImpl_UpdateSwapOutSubscribeProgressInsertFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SwapOutProgressKey).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressInsertSQL).WithArgs(SwapOutProgressKey, mockBlockHeight).
		WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()
	err := s.UpdateSwapOutSubscribeProgress(mockBlockHeight)
	assert.Contains(t, err.Error(), "failed to insert record in subscribe progress table")
}

func TestSpDBImpl_UpdateSwapOutSubscribeProgressUpdateFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SwapOutProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(SwapOutProgressKey, 5))
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressUpdateSQL).WithArgs(SwapOutProgressKey, mockBlockHeight, SwapOutProgressKey).
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateSwapOutSubscribeProgress(mockBlockHeight)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_QuerySwapOutSubscribeProgress(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(SwapOutProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(SwapOutProgressKey, mockBlockHeight))
	result, err := s.QuerySwapOutSubscribeProgress()
	assert.Nil(t, err)
	assert.Equal(t, uint64(mockBlockHeight), result)
}

func TestSpDBImpl_QuerySwapOutSubscribeProgressRecordNotFound(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WillReturnError(gorm.ErrRecordNotFound)
	result, err := s.QuerySwapOutSubscribeProgress()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), result)
}

func TestSpDBImpl_QuerySwapOutSubscribeProgressFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WillReturnError(mockDBInternalError)
	result, err := s.QuerySwapOutSubscribeProgress()
	assert.Equal(t, mockDBInternalError, err)
	assert.Equal(t, uint64(0), result)
}

func TestSpDBImpl_UpdateBucketMigrateSubscribeProgressInsertSuccess(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(BucketMigrateProgressKey).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressInsertSQL).WithArgs(BucketMigrateProgressKey, mockBlockHeight).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateBucketMigrateSubscribeProgress(mockBlockHeight)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateBucketMigrateSubscribeProgressUpdateSuccess(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(BucketMigrateProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(BucketMigrateProgressKey, 5))
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressUpdateSQL).WithArgs(BucketMigrateProgressKey, mockBlockHeight, BucketMigrateProgressKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateBucketMigrateSubscribeProgress(mockBlockHeight)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateBucketMigrateSubscribeProgressQueryFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(BucketMigrateProgressKey).WillReturnError(mockDBInternalError)
	err := s.UpdateBucketMigrateSubscribeProgress(mockBlockHeight)
	assert.Equal(t, mockDBInternalError, err)
}

func TestSpDBImpl_UpdateBucketMigrateSubscribeProgressInsertFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(BucketMigrateProgressKey).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressInsertSQL).WithArgs(BucketMigrateProgressKey, mockBlockHeight).
		WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()
	err := s.UpdateBucketMigrateSubscribeProgress(mockBlockHeight)
	assert.Contains(t, err.Error(), "failed to insert record in subscribe progress table")
}

func TestSpDBImpl_UpdateBucketMigrateSubscribeProgressUpdateFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(BucketMigrateProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(BucketMigrateProgressKey, 5))
	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateSubscribeProgressUpdateSQL).WithArgs(BucketMigrateProgressKey, mockBlockHeight, BucketMigrateProgressKey).
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateBucketMigrateSubscribeProgress(mockBlockHeight)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_QueryBucketMigrateSubscribeProgress(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WithArgs(BucketMigrateProgressKey).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "last_subscribed_block_height"}).AddRow(BucketMigrateProgressKey, mockBlockHeight))
	result, err := s.QueryBucketMigrateSubscribeProgress()
	assert.Nil(t, err)
	assert.Equal(t, uint64(mockBlockHeight), result)
}

func TestSpDBImpl_QueryBucketMigrateSubscribeProgressRecordNotFound(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WillReturnError(gorm.ErrRecordNotFound)
	result, err := s.QueryBucketMigrateSubscribeProgress()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), result)
}

func TestSpDBImpl_QueryBucketMigrateSubscribeProgressFailure(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockMigrateSubscribeProgressQuerySQL).WillReturnError(mockDBInternalError)
	result, err := s.QueryBucketMigrateSubscribeProgress()
	assert.Equal(t, mockDBInternalError, err)
	assert.Equal(t, uint64(0), result)
}

func TestSpDBImpl_InsertSwapOutUnitSuccess(t *testing.T) {
	s, mock := setupDB(t)
	msgSwapOut := &virtualgrouptypes.MsgSwapOut{
		StorageProvider:            "0x1a2b3c4d5e67890654",
		GlobalVirtualGroupFamilyId: 1,
		GlobalVirtualGroupIds:      []uint32{1, 2, 3},
		SuccessorSpId:              2,
	}
	byteData, err := json.Marshal(msgSwapOut)
	assert.Nil(t, err)
	meta := &spdb.SwapOutMeta{
		SwapOutKey:    mockSwapOutKey,
		IsDestSP:      true,
		SwapOutMsg:    msgSwapOut,
		CompletedGVGs: []uint32{5, 6},
	}

	swapOutTable := &SwapOutTable{
		SwapOutKey:       meta.SwapOutKey,
		IsDestSP:         meta.IsDestSP,
		SwapOutMsg:       hex.EncodeToString(byteData),
		CompletedGVGList: util.Uint32SliceToString(meta.CompletedGVGs),
	}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `swap_out_unit` (`swap_out_key`,`is_dest_sp`,`swap_out_msg`,`completed_gvg_list`) VALUES (?,?,?,?)").
		WithArgs(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err = s.InsertSwapOutUnit(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertSwapOutUnitFailure(t *testing.T) {
	msgSwapOut := &virtualgrouptypes.MsgSwapOut{
		StorageProvider:            "0x1a2b3c4d5e67890654",
		GlobalVirtualGroupFamilyId: 1,
		GlobalVirtualGroupIds:      []uint32{1, 2, 3},
		SuccessorSpId:              2,
	}
	meta := &spdb.SwapOutMeta{
		SwapOutKey:    mockSwapOutKey,
		IsDestSP:      true,
		SwapOutMsg:    msgSwapOut,
		CompletedGVGs: []uint32{5, 6},
	}

	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `swap_out_unit` (`swap_out_key`,`is_dest_sp`,`swap_out_msg`,`completed_gvg_list`) VALUES (?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.InsertSwapOutUnit(meta)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateSwapOutUnitCompletedGVGListSuccess(t *testing.T) {
	var completedGVGList = []uint32{1, 2}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `swap_out_unit` SET `completed_gvg_list`=? WHERE swap_out_key = ? and is_dest_sp = 1").
		WithArgs(util.Uint32SliceToString(completedGVGList), mockSwapOutKey).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateSwapOutUnitCompletedGVGList(mockSwapOutKey, completedGVGList)
	assert.Nil(t, err)
}

func TestSpDBImpl_QuerySwapOutUnitInSrcSPSuccess(t *testing.T) {
	s, mock := setupDB(t)
	var completedGVGList = []uint32{1, 2}
	msgSwapOut := &virtualgrouptypes.MsgSwapOut{
		StorageProvider:            "0x1a2b3c4d5e67890654",
		GlobalVirtualGroupFamilyId: 1,
		GlobalVirtualGroupIds:      []uint32{1, 2, 3},
		SuccessorSpId:              2,
	}
	byteData, err := json.Marshal(msgSwapOut)
	assert.Nil(t, err)
	swapOutTable := &SwapOutTable{
		SwapOutKey:       mockSwapOutKey,
		IsDestSP:         true,
		SwapOutMsg:       hex.EncodeToString(byteData),
		CompletedGVGList: util.Uint32SliceToString(completedGVGList),
	}
	mock.ExpectQuery(mockQuerySwapOutUnitInSrcSPQuerySQL).WithArgs(mockSwapOutKey).
		WillReturnRows(sqlmock.NewRows([]string{"swap_out_key", "is_dest_sp", "swap_out_msg", "completed_gvg_list"}).
			AddRow(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList))
	meta, err := s.QuerySwapOutUnitInSrcSP(mockSwapOutKey)
	assert.Nil(t, err)
	assert.Equal(t, completedGVGList, meta.CompletedGVGs)
}

func TestSpDBImpl_QuerySwapOutUnitInSrcSPFailure1(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	s, mock := setupDB(t)
	mock.ExpectQuery(mockQuerySwapOutUnitInSrcSPQuerySQL).WillReturnError(mockDBInternalError)
	meta, err := s.QuerySwapOutUnitInSrcSP(mockSwapOutKey)
	assert.Equal(t, mockDBInternalError, err)
	assert.Nil(t, meta)
}

func TestSpDBImpl_QuerySwapOutUnitInSrcSPFailure2(t *testing.T) {
	t.Log("Failure case description: hex decode string returns error")
	s, mock := setupDB(t)
	swapOutTable := &SwapOutTable{
		SwapOutKey:       mockSwapOutKey,
		IsDestSP:         true,
		SwapOutMsg:       "test",
		CompletedGVGList: util.Uint32SliceToString([]uint32{1, 2}),
	}
	mock.ExpectQuery(mockQuerySwapOutUnitInSrcSPQuerySQL).WithArgs(mockSwapOutKey).
		WillReturnRows(sqlmock.NewRows([]string{"swap_out_key", "is_dest_sp", "swap_out_msg", "completed_gvg_list"}).
			AddRow(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList))
	meta, err := s.QuerySwapOutUnitInSrcSP(mockSwapOutKey)
	assert.NotNil(t, err)
	assert.Nil(t, meta)
}

func TestSpDBImpl_QuerySwapOutUnitInSrcFailure3(t *testing.T) {
	t.Log("Failure case description: json unmarshal returns error")
	s, mock := setupDB(t)
	swapOutTable := &SwapOutTable{
		SwapOutKey:       mockSwapOutKey,
		IsDestSP:         true,
		SwapOutMsg:       hex.EncodeToString([]byte(`{"name":what?}`)),
		CompletedGVGList: util.Uint32SliceToString([]uint32{1, 2}),
	}
	mock.ExpectQuery(mockQuerySwapOutUnitInSrcSPQuerySQL).WithArgs(mockSwapOutKey).
		WillReturnRows(sqlmock.NewRows([]string{"swap_out_key", "is_dest_sp", "swap_out_msg", "completed_gvg_list"}).
			AddRow(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList))
	meta, err := s.QuerySwapOutUnitInSrcSP(mockSwapOutKey)
	assert.NotNil(t, err)
	assert.Nil(t, meta)
}

func TestSpDBImpl_QuerySwapOutUnitInSrcSPFailure4(t *testing.T) {
	t.Log("Failure case description: string to uint32 slice returns error")
	s, mock := setupDB(t)
	msgSwapOut := &virtualgrouptypes.MsgSwapOut{
		StorageProvider:            "0x1a2b3c4d5e67890654",
		GlobalVirtualGroupFamilyId: 1,
		GlobalVirtualGroupIds:      []uint32{1, 2, 3},
		SuccessorSpId:              2,
	}
	byteData, err := json.Marshal(msgSwapOut)
	assert.Nil(t, err)
	swapOutTable := &SwapOutTable{
		SwapOutKey:       mockSwapOutKey,
		IsDestSP:         true,
		SwapOutMsg:       hex.EncodeToString(byteData),
		CompletedGVGList: "test",
	}
	mock.ExpectQuery(mockQuerySwapOutUnitInSrcSPQuerySQL).WithArgs(mockSwapOutKey).
		WillReturnRows(sqlmock.NewRows([]string{"swap_out_key", "is_dest_sp", "swap_out_msg", "completed_gvg_list"}).
			AddRow(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList))
	meta, err := s.QuerySwapOutUnitInSrcSP(mockSwapOutKey)
	assert.NotNil(t, err)
	assert.Nil(t, meta)
}

func TestSpDBImpl_ListDestSPSwapOutUnitsSuccess(t *testing.T) {
	var completedGVGList = []uint32{1, 2}
	msgSwapOut := &virtualgrouptypes.MsgSwapOut{
		StorageProvider:            "0x1a2b3c4d5e67890654",
		GlobalVirtualGroupFamilyId: 1,
		GlobalVirtualGroupIds:      []uint32{1, 2, 3},
		SuccessorSpId:              2,
	}
	byteData, err := json.Marshal(msgSwapOut)
	assert.Nil(t, err)
	swapOutTable := &SwapOutTable{
		SwapOutKey:       mockSwapOutKey,
		IsDestSP:         true,
		SwapOutMsg:       hex.EncodeToString(byteData),
		CompletedGVGList: util.Uint32SliceToString(completedGVGList),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery(mockListDestSPSwapOutUintsQuerySQL).
		WillReturnRows(sqlmock.NewRows([]string{"swap_out_key", "is_dest_sp", "swap_out_msg", "completed_gvg_list"}).
			AddRow(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList))
	result, err := s.ListDestSPSwapOutUnits()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, completedGVGList, result[0].CompletedGVGs)
}

func TestSpDBImpl_ListDestSPSwapOutUnitsFailure1(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	s, mock := setupDB(t)
	mock.ExpectQuery(mockListDestSPSwapOutUintsQuerySQL).WillReturnError(mockDBInternalError)
	result, err := s.ListDestSPSwapOutUnits()
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_ListDestSPSwapOutUnitsFailure2(t *testing.T) {
	t.Log("Failure case description: hex decode string returns error")
	swapOutTable := &SwapOutTable{
		SwapOutKey:       mockSwapOutKey,
		IsDestSP:         true,
		SwapOutMsg:       "test",
		CompletedGVGList: util.Uint32SliceToString([]uint32{1, 2}),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery(mockListDestSPSwapOutUintsQuerySQL).
		WillReturnRows(sqlmock.NewRows([]string{"swap_out_key", "is_dest_sp", "swap_out_msg", "completed_gvg_list"}).
			AddRow(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList))
	result, err := s.ListDestSPSwapOutUnits()
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_ListDestSPSwapOutUnitsFailure3(t *testing.T) {
	t.Log("Failure case description: json unmarshal returns error")
	swapOutTable := &SwapOutTable{
		SwapOutKey:       mockSwapOutKey,
		IsDestSP:         true,
		SwapOutMsg:       hex.EncodeToString([]byte(`{"name":what?}`)),
		CompletedGVGList: util.Uint32SliceToString([]uint32{1, 2}),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery(mockListDestSPSwapOutUintsQuerySQL).
		WillReturnRows(sqlmock.NewRows([]string{"swap_out_key", "is_dest_sp", "swap_out_msg", "completed_gvg_list"}).
			AddRow(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList))
	result, err := s.ListDestSPSwapOutUnits()
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_ListDestSPSwapOutUnitsFailure4(t *testing.T) {
	t.Log("Failure case description: string to uint32 slice returns error")
	msgSwapOut := &virtualgrouptypes.MsgSwapOut{
		StorageProvider:            "0x1a2b3c4d5e67890654",
		GlobalVirtualGroupFamilyId: 1,
		GlobalVirtualGroupIds:      []uint32{1, 2, 3},
		SuccessorSpId:              2,
	}
	byteData, err := json.Marshal(msgSwapOut)
	assert.Nil(t, err)
	swapOutTable := &SwapOutTable{
		SwapOutKey:       mockSwapOutKey,
		IsDestSP:         true,
		SwapOutMsg:       hex.EncodeToString(byteData),
		CompletedGVGList: "test",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery(mockListDestSPSwapOutUintsQuerySQL).
		WillReturnRows(sqlmock.NewRows([]string{"swap_out_key", "is_dest_sp", "swap_out_msg", "completed_gvg_list"}).
			AddRow(swapOutTable.SwapOutKey, swapOutTable.IsDestSP, swapOutTable.SwapOutMsg, swapOutTable.CompletedGVGList))
	result, err := s.ListDestSPSwapOutUnits()
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_InsertMigrateGVGUnitSuccess(t *testing.T) {
	s, mock := setupDB(t)
	meta := &spdb.MigrateGVGUnitMeta{
		MigrateGVGKey:            "mock_migrate_gvg_key",
		SwapOutKey:               "mock_swap_out_key",
		GlobalVirtualGroupID:     1,
		DestGlobalVirtualGroupID: 2,
		VirtualGroupFamilyID:     3,
		RedundancyIndex:          4,
		BucketID:                 5,
		SrcSPID:                  6,
		DestSPID:                 7,
		LastMigratedObjectID:     8,
		MigrateStatus:            9,
		RetryTime:                10,
	}
	m := &MigrateGVGTable{
		MigrateKey:               meta.MigrateGVGKey,
		SwapOutKey:               meta.SwapOutKey,
		GlobalVirtualGroupID:     meta.GlobalVirtualGroupID,
		DestGlobalVirtualGroupID: meta.DestGlobalVirtualGroupID,
		VirtualGroupFamilyID:     meta.VirtualGroupFamilyID,
		BucketID:                 meta.BucketID,
		RedundancyIndex:          meta.RedundancyIndex,
		SrcSPID:                  meta.SrcSPID,
		DestSPID:                 meta.DestSPID,
		LastMigratedObjectID:     meta.LastMigratedObjectID,
		MigrateStatus:            meta.MigrateStatus,
		RetryTime:                meta.RetryTime,
	}
	mock.ExpectQuery(mockMigrateGVGQuerySQL).WithArgs(meta.MigrateGVGKey).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `migrate_gvg` (`migrate_key`,`swap_out_key`,`global_virtual_group_id`,`dest_global_virtual_group_id`,`virtual_group_family_id`,`bucket_id`,`redundancy_index`,`src_sp_id`,`dest_sp_id`,`last_migrated_object_id`,`migrate_status`,`retry_time`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)").
		WithArgs(m.MigrateKey, m.SwapOutKey, m.GlobalVirtualGroupID, m.DestGlobalVirtualGroupID, m.VirtualGroupFamilyID, m.BucketID, m.RedundancyIndex, m.SrcSPID, m.DestSPID, m.LastMigratedObjectID, m.MigrateStatus, m.RetryTime).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertMigrateGVGUnit(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertMigrateGVGUnitFailure(t *testing.T) {
	s, mock := setupDB(t)
	meta := &spdb.MigrateGVGUnitMeta{
		MigrateGVGKey:            "mock_migrate_gvg_key",
		SwapOutKey:               "mock_swap_out_key",
		GlobalVirtualGroupID:     1,
		DestGlobalVirtualGroupID: 2,
		VirtualGroupFamilyID:     3,
		RedundancyIndex:          4,
		BucketID:                 5,
		SrcSPID:                  6,
		DestSPID:                 7,
		LastMigratedObjectID:     8,
		MigrateStatus:            9,
	}
	mock.ExpectQuery(mockMigrateGVGQuerySQL).WithArgs(meta.MigrateGVGKey).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `migrate_gvg` (`migrate_key`,`swap_out_key`,`global_virtual_group_id`,`dest_global_virtual_group_id`,`virtual_group_family_id`,`bucket_id`,`redundancy_index`,`src_sp_id`,`dest_sp_id`,`last_migrated_object_id`,`migrate_status`) VALUES (?,?,?,?,?,?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.InsertMigrateGVGUnit(meta)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_DeleteMigrateGVGUnit(t *testing.T) {
	s, mock := setupDB(t)
	meta := &spdb.MigrateGVGUnitMeta{
		MigrateGVGKey:            "mock_migrate_gvg_key",
		SwapOutKey:               "mock_swap_out_key",
		GlobalVirtualGroupID:     1,
		DestGlobalVirtualGroupID: 2,
		VirtualGroupFamilyID:     3,
		RedundancyIndex:          4,
		BucketID:                 5,
		SrcSPID:                  6,
		DestSPID:                 7,
		LastMigratedObjectID:     8,
		MigrateStatus:            9,
	}

	mock.ExpectBegin()
	mock.ExpectExec(mockMigrateGVGDeleteSQL).WithArgs(meta.MigrateGVGKey).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := s.DeleteMigrateGVGUnit(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateMigrateGVGUnitStatusSuccess(t *testing.T) {
	var (
		migrateKey    = "mockMigrateKey"
		migrateStatus = 2
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `migrate_gvg` SET `migrate_status`=? WHERE migrate_key = ?").
		WithArgs(migrateStatus, migrateKey).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateMigrateGVGUnitStatus(migrateKey, migrateStatus)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateMigrateGVGUnitStatusFailure(t *testing.T) {
	var (
		migrateKey    = "mockMigrateKey"
		migrateStatus = 2
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `migrate_gvg` SET `migrate_status`=? WHERE migrate_key = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateMigrateGVGUnitStatus(migrateKey, migrateStatus)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateMigrateGVGUnitLastMigrateObjectIDSuccess(t *testing.T) {
	var (
		migrateKey           = "mockMigrateKey"
		lastMigratedObjectID = uint64(25)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `migrate_gvg` SET `last_migrated_object_id`=? WHERE migrate_key = ?").
		WithArgs(lastMigratedObjectID, migrateKey).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateMigrateGVGUnitLastMigrateObjectID(migrateKey, lastMigratedObjectID)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateMigrateGVGUnitLastMigrateObjectIDFailure(t *testing.T) {
	var (
		migrateKey           = "mockMigrateKey"
		lastMigratedObjectID = uint64(25)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `migrate_gvg` SET `last_migrated_object_id`=? WHERE migrate_key = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateMigrateGVGUnitLastMigrateObjectID(migrateKey, lastMigratedObjectID)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_QueryMigrateGVGUnitSuccess(t *testing.T) {
	migrateKey := "mockMigrateKey"
	m := &spdb.MigrateGVGUnitMeta{
		GlobalVirtualGroupID:     1,
		DestGlobalVirtualGroupID: 2,
		VirtualGroupFamilyID:     3,
		RedundancyIndex:          4,
		BucketID:                 5,
		SrcSPID:                  6,
		DestSPID:                 7,
		LastMigratedObjectID:     8,
		MigrateStatus:            9,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `migrate_gvg` WHERE migrate_key = ? ORDER BY `migrate_gvg`.`migrate_key` LIMIT 1").
		WithArgs(migrateKey).WillReturnRows(sqlmock.NewRows([]string{"global_virtual_group_id", "dest_global_virtual_group_id",
		"virtual_group_family_id", "redundancy_index", "bucket_id", "src_sp_id", "dest_sp_id", "last_migrated_object_id", "migrate_status"}).
		AddRow(m.GlobalVirtualGroupID, m.DestGlobalVirtualGroupID, m.VirtualGroupFamilyID, m.RedundancyIndex, m.BucketID,
			m.SrcSPID, m.DestSPID, m.LastMigratedObjectID, m.MigrateStatus))
	result, err := s.QueryMigrateGVGUnit(migrateKey)
	assert.Nil(t, err)
	assert.Equal(t, m.BucketID, result.BucketID)
}

func TestSpDBImpl_QueryMigrateGVGUnitFailure(t *testing.T) {
	migrateKey := "mockMigrateKey"
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `migrate_gvg` WHERE migrate_key = ? ORDER BY `migrate_gvg`.`migrate_key` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.QueryMigrateGVGUnit(migrateKey)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_ListMigrateGVGUnitsByBucketIDSuccess(t *testing.T) {
	var bucketID = uint64(4)
	m := MigrateGVGTable{
		MigrateKey:               "mock_migrate_key",
		SwapOutKey:               "mock_swap_out_key",
		GlobalVirtualGroupID:     1,
		DestGlobalVirtualGroupID: 2,
		VirtualGroupFamilyID:     3,
		BucketID:                 4,
		RedundancyIndex:          5,
		SrcSPID:                  6,
		DestSPID:                 7,
		LastMigratedObjectID:     8,
		MigrateStatus:            9,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `migrate_gvg` WHERE bucket_id = ?").WithArgs(bucketID).
		WillReturnRows(sqlmock.NewRows([]string{"migrate_key", "swap_out_key", "global_virtual_group_id", "dest_global_virtual_group_id",
			"virtual_group_family_id", "redundancy_index", "bucket_id", "src_sp_id", "dest_sp_id", "last_migrated_object_id", "migrate_status"}).
			AddRow(m.MigrateKey, m.SwapOutKey, m.GlobalVirtualGroupID, m.DestGlobalVirtualGroupID, m.VirtualGroupFamilyID, m.BucketID,
				m.RedundancyIndex, m.SrcSPID, m.DestSPID, m.LastMigratedObjectID, m.MigrateStatus))
	result, err := s.ListMigrateGVGUnitsByBucketID(bucketID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, result[0].LastMigratedObjectID, uint64(8))
}

func TestSpDBImpl_ListMigrateGVGUnitsByBucketIDFailure(t *testing.T) {
	var bucketID = uint64(4)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `migrate_gvg` WHERE bucket_id = ?").WillReturnError(mockDBInternalError)
	result, err := s.ListMigrateGVGUnitsByBucketID(bucketID)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_DeleteMigrateGVGUnitsByBucketIDSuccess(t *testing.T) {
	var bucketID = uint64(4)
	m := MigrateGVGTable{
		MigrateKey:               "mock_migrate_key",
		SwapOutKey:               "mock_swap_out_key",
		GlobalVirtualGroupID:     1,
		DestGlobalVirtualGroupID: 2,
		VirtualGroupFamilyID:     3,
		BucketID:                 4,
		RedundancyIndex:          5,
		SrcSPID:                  6,
		DestSPID:                 7,
		LastMigratedObjectID:     8,
		MigrateStatus:            9,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `migrate_gvg` WHERE bucket_id = ?").WithArgs(bucketID).
		WillReturnRows(sqlmock.NewRows([]string{"migrate_key", "swap_out_key", "global_virtual_group_id", "dest_global_virtual_group_id",
			"virtual_group_family_id", "redundancy_index", "bucket_id", "src_sp_id", "dest_sp_id", "last_migrated_object_id", "migrate_status"}).
			AddRow(m.MigrateKey, m.SwapOutKey, m.GlobalVirtualGroupID, m.DestGlobalVirtualGroupID, m.VirtualGroupFamilyID, m.BucketID,
				m.RedundancyIndex, m.SrcSPID, m.DestSPID, m.LastMigratedObjectID, m.MigrateStatus))
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `migrate_gvg` WHERE bucket_id = ? AND `migrate_gvg`.`migrate_key` = ?").
		WithArgs(bucketID, m.MigrateKey).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.DeleteMigrateGVGUnitsByBucketID(bucketID)
	assert.Nil(t, err)
}

func TestSpDBImpl_DeleteMigrateGVGUnitsByBucketIDFailure(t *testing.T) {
	var bucketID = uint64(4)
	m := MigrateGVGTable{
		MigrateKey:               "mock_migrate_key",
		SwapOutKey:               "mock_swap_out_key",
		GlobalVirtualGroupID:     1,
		DestGlobalVirtualGroupID: 2,
		VirtualGroupFamilyID:     3,
		BucketID:                 4,
		RedundancyIndex:          5,
		SrcSPID:                  6,
		DestSPID:                 7,
		LastMigratedObjectID:     8,
		MigrateStatus:            9,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `migrate_gvg` WHERE bucket_id = ?").WithArgs(bucketID).
		WillReturnRows(sqlmock.NewRows([]string{"migrate_key", "swap_out_key", "global_virtual_group_id", "dest_global_virtual_group_id",
			"virtual_group_family_id", "redundancy_index", "bucket_id", "src_sp_id", "dest_sp_id", "last_migrated_object_id", "migrate_status"}).
			AddRow(m.MigrateKey, m.SwapOutKey, m.GlobalVirtualGroupID, m.DestGlobalVirtualGroupID, m.VirtualGroupFamilyID, m.BucketID,
				m.RedundancyIndex, m.SrcSPID, m.DestSPID, m.LastMigratedObjectID, m.MigrateStatus))
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `migrate_gvg` WHERE bucket_id = ? AND `migrate_gvg`.`migrate_key` = ?").WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.DeleteMigrateGVGUnitsByBucketID(bucketID)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}
