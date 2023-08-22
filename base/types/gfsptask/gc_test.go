package gfsptask

import (
	"testing"

	"github.com/stretchr/testify/assert"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func TestInitGCObjectTask(t *testing.T) {
	m := &GfSpGCObjectTask{}
	m.InitGCObjectTask(coretask.MaxTaskPriority, 1, 2, 0)
}

func TestGfSpGCObjectTask_Key(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.Key()
}

func TestGfSpGCObjectTask_Type(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskGCObject, result)
}

func TestGfSpGCObjectTask_Info(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.Info()
}

func TestGfSpGCObjectTask_GetAddress(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetAddress()
}

func TestGfSpGCObjectTask_SetAddress(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetAddress("mockAddress")
}

func TestGfSpGCObjectTask_GetCreateTime(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetCreateTime()
}

func TestGfSpGCObjectTask_SetCreateTime(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetCreateTime(1)
}

func TestGfSpGCObjectTask_GetUpdateTime(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetUpdateTime()
}

func TestGfSpGCObjectTask_SetUpdateTime(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetUpdateTime(1)
}

func TestGfSpGCObjectTask_GetTimeout(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetTimeout()
}

func TestGfSpGCObjectTask_SetTimeout(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetTimeout(1)
}

func TestGfSpGCObjectTask_ExceedTimeout(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.ExceedTimeout()
}

func TestGfSpGCObjectTask_GetRetry(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetRetry()
}

func TestGfSpGCObjectTask_IncRetry(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.IncRetry()
}

func TestGfSpGCObjectTask_SetRetry(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetRetry(1)
}

func TestGfSpGCObjectTask_GetMaxRetry(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetMaxRetry()
}

func TestGfSpGCObjectTask_SetMaxRetry(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetMaxRetry(1)
}

func TestGfSpGCObjectTask_ExceedRetry(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.ExceedRetry()
}

func TestGfSpGCObjectTask_Expired(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.Expired()
}

func TestGfSpGCObjectTask_GetPriority(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetPriority()
}

func TestGfSpGCObjectTask_SetPriority(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetPriority(1)
}

func TestGfSpGCObjectTask_EstimateLimit(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.EstimateLimit()
}

func TestGfSpGCObjectTask_Error(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	_ = m.Error()
}

func TestGfSpGCObjectTask_GetUserAddress(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetUserAddress()
}

func TestGfSpGCObjectTask_SetUserAddress(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetUserAddress("mock")
}

func TestGfSpGCObjectTask_SetLogs(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetLogs("mock")
}

func TestGfSpGCObjectTask_GetLogs(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetLogs()
}

func TestGfSpGCObjectTask_AppendLog(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.AppendLog("mock")
}

func TestGfSpGCObjectTask_SetError(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetError(nil)
}

func TestGfSpGCObjectTask_SetStartBlockNumber(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetStartBlockNumber(1)
}

func TestGfSpGCObjectTask_SetEndBlockNumber(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetEndBlockNumber(1)
}

func TestGfSpGCObjectTask_GetGCObjectProgress(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.GetGCObjectProgress()
}

func TestGfSpGCObjectTask_SetGCObjectProgress(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetGCObjectProgress(2, 2)
}

func TestGfSpGCObjectTask_SetCurrentBlockNumber(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetCurrentBlockNumber(1)
}

func TestGfSpGCObjectTask_SetLastDeletedObjectId(t *testing.T) {
	m := &GfSpGCObjectTask{Task: &GfSpTask{}}
	m.SetLastDeletedObjectId(1)
}

func TestGfSpGCZombiePieceTask_Key(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.Key()
}

func TestGfSpGCZombiePieceTask_Type(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskGCZombiePiece, result)
}

func TestGfSpGCZombiePieceTask_Info(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.Info()
}

func TestGfSpGCZombiePieceTask_GetAddress(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetAddress()
}

func TestGfSpGCZombiePieceTask_SetAddress(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetAddress("mockAddress")
}

func TestGfSpGCZombiePieceTask_GetCreateTime(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetCreateTime()
}

func TestGfSpGCZombiePieceTask_SetCreateTime(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetCreateTime(1)
}

func TestGfSpGCZombiePieceTask_GetUpdateTime(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetUpdateTime()
}

func TestGfSpGCZombiePieceTask_SetUpdateTime(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetUpdateTime(1)
}

func TestGfSpGCZombiePieceTask_GetTimeout(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetTimeout()
}

func TestGfSpGCZombiePieceTask_SetTimeout(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetTimeout(1)
}

func TestGfSpGCZombiePieceTask_ExceedTimeout(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.ExceedTimeout()
}

func TestGfSpGCZombiePieceTask_GetRetry(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetRetry()
}

func TestGfSpGCZombiePieceTask_IncRetry(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.IncRetry()
}

func TestGfSpGCZombiePieceTask_SetRetry(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetRetry(1)
}

func TestGfSpGCZombiePieceTask_GetMaxRetry(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetMaxRetry()
}

func TestGfSpGCZombiePieceTask_SetMaxRetry(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetMaxRetry(1)
}

func TestGfSpGCZombiePieceTask_ExceedRetry(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.ExceedRetry()
}

func TestGfSpGCZombiePieceTask_Expired(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.Expired()
}

func TestGfSpGCZombiePieceTask_GetPriority(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetPriority()
}

func TestGfSpGCZombiePieceTask_SetPriority(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetPriority(1)
}

func TestGfSpGCZombiePieceTask_EstimateLimit(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.EstimateLimit()
}

func TestGfSpGCZombiePieceTask_Error(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	_ = m.Error()
}

func TestGfSpGCZombiePieceTask_GetUserAddress(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetUserAddress()
}

func TestGfSpGCZombiePieceTask_SetUserAddress(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetUserAddress("mock")
}

func TestGfSpGCZombiePieceTask_SetLogs(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetLogs("mock")
}

func TestGfSpGCZombiePieceTask_GetLogs(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.GetLogs()
}

func TestGfSpGCZombiePieceTask_AppendLog(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.AppendLog("mock")
}

func TestGfSpGCZombiePieceTask_SetError(t *testing.T) {
	m := &GfSpGCZombiePieceTask{Task: &GfSpTask{}}
	m.SetError(nil)
}

// ===

func TestGfSpGCMetaTask_Key(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.Key()
}

func TestGfSpGCMetaTask_Type(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskGCMeta, result)
}

func TestGfSpGCMetaTask_Info(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.Info()
}

func TestGfSpGCMetaTask_GetAddress(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetAddress()
}

func TestGfSpGCMetaTask_SetAddress(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetAddress("mockAddress")
}

func TestGfSpGCMetaTask_GetCreateTime(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetCreateTime()
}

func TestGfSpGCMetaTask_SetCreateTime(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetCreateTime(1)
}

func TestGfSpGCMetaTask_GetUpdateTime(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetUpdateTime()
}

func TestGfSpGCMetaTask_SetUpdateTime(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetUpdateTime(1)
}

func TestGfSpGCMetaTask_GetTimeout(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetTimeout()
}

func TestGfSpGCMetaTask_SetTimeout(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetTimeout(1)
}

func TestGfSpGCMetaTask_ExceedTimeout(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.ExceedTimeout()
}

func TestGfSpGCMetaTask_GetRetry(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetRetry()
}

func TestGfSpGCMetaTask_IncRetry(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.IncRetry()
}

func TestGfSpGCMetaTask_SetRetry(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetRetry(1)
}

func TestGfSpGCMetaTask_GetMaxRetry(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetMaxRetry()
}

func TestGfSpGCMetaTask_SetMaxRetry(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetMaxRetry(1)
}

func TestGfSpGCMetaTask_ExceedRetry(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.ExceedRetry()
}

func TestGfSpGCMetaTask_Expired(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.Expired()
}

func TestGfSpGCMetaTask_GetPriority(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetPriority()
}

func TestGfSpGCMetaTask_SetPriority(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetPriority(1)
}

func TestGfSpGCMetaTask_EstimateLimit(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.EstimateLimit()
}

func TestGfSpGCMetaTask_Error(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	_ = m.Error()
}

func TestGfSpGCMetaTask_GetUserAddress(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetUserAddress()
}

func TestGfSpGCMetaTask_SetUserAddress(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetUserAddress("mock")
}

func TestGfSpGCMetaTask_SetLogs(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetLogs("mock")
}

func TestGfSpGCMetaTask_GetLogs(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetLogs()
}

func TestGfSpGCMetaTask_AppendLog(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.AppendLog("mock")
}

func TestGfSpGCMetaTask_SetError(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetError(nil)
}

func TestGfSpGCMetaTask_GetGCMetaStatus(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.GetGCMetaStatus()
}

func TestGfSpGCMetaTask_SetGCMetaStatus(t *testing.T) {
	m := &GfSpGCMetaTask{Task: &GfSpTask{}}
	m.SetGCMetaStatus(0, 0)
}
