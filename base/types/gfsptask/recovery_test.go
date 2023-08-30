package gfsptask

import (
	"testing"

	"github.com/stretchr/testify/assert"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func TestInitRecoverPieceTask(t *testing.T) {
	m := &GfSpRecoverPieceTask{}
	m.InitRecoverPieceTask(mockObjectInfo, mockStorageParams, coretask.MaxTaskPriority, 1, 2, 10, 1, 0)
}

func TestGfSpRecoverPieceTask_Key(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpRecoverPieceTask_Type(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskRecoverPiece, result)
}

func TestGfSpRecoverPieceTask_Info(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpRecoverPieceTask_GetAddress(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpRecoverPieceTask_SetAddress(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpRecoverPieceTask_GetCreateTime(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpRecoverPieceTask_SetCreateTime(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpRecoverPieceTask_GetUpdateTime(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpRecoverPieceTask_SetUpdateTime(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpRecoverPieceTask_GetTimeout(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpRecoverPieceTask_SetTimeout(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpRecoverPieceTask_ExceedTimeout(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpRecoverPieceTask_GetRetry(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpRecoverPieceTask_IncRetry(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpRecoverPieceTask_SetRetry(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpRecoverPieceTask_GetMaxRetry(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpRecoverPieceTask_SetMaxRetry(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpRecoverPieceTask_ExceedRetry(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpRecoverPieceTask_Expired(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpRecoverPieceTask_GetPriority(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpRecoverPieceTask_SetPriority(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpRecoverPieceTask_EstimateLimit(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpRecoverPieceTask_Error(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpRecoverPieceTask_GetUserAddress(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpRecoverPieceTask_SetUserAddress(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpRecoverPieceTask_SetLogs(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpRecoverPieceTask_GetLogs(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpRecoverPieceTask_AppendLog(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpRecoverPieceTask_SetError(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpRecoverPieceTask_SetObjectInfo(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpRecoverPieceTask_SetStorageParams(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestGfSpRecoverPieceTask_SetSignature(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSignature([]byte("mock"))
}

func TestGfSpRecoverPieceTask_SetPieceSize(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPieceSize(1)
}

func TestGfSpRecoverPieceTask_SetRecoverDone(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRecoverDone()
}

func TestGfSpRecoverPieceTask_GetSignBytes(t *testing.T) {
	m := &GfSpRecoverPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetSignBytes()
}
