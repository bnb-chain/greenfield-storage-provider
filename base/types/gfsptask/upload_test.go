package gfsptask

import (
	"testing"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/assert"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func TestInitUploadObjectTask(t *testing.T) {
	m := &GfSpUploadObjectTask{}
	m.InitUploadObjectTask(1, mockObjectInfo, mockStorageParams, 0)
}

func TestGfSpUploadObjectTask_Key(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpUploadObjectTask_Type(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskUpload, result)
}

func TestGfSpUploadObjectTask_Info(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpUploadObjectTask_GetAddress(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpUploadObjectTask_SetAddress(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpUploadObjectTask_GetCreateTime(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpUploadObjectTask_SetCreateTime(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpUploadObjectTask_GetUpdateTime(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpUploadObjectTask_SetUpdateTime(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpUploadObjectTask_GetTimeout(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpUploadObjectTask_SetTimeout(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpUploadObjectTask_ExceedTimeout(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpUploadObjectTask_GetRetry(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpUploadObjectTask_IncRetry(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpUploadObjectTask_SetRetry(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpUploadObjectTask_GetMaxRetry(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpUploadObjectTask_SetMaxRetry(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpUploadObjectTask_ExceedRetry(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpUploadObjectTask_Expired(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpUploadObjectTask_GetPriority(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpUploadObjectTask_SetPriority(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpUploadObjectTask_EstimateLimit1(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpUploadObjectTask_EstimateLimit2(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task: &GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			PayloadSize: 10,
		},
		StorageParams: &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 2}},
	}
	m.EstimateLimit()
}

func TestGfSpUploadObjectTask_Error(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpUploadObjectTask_GetUserAddress(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpUploadObjectTask_SetUserAddress(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpUploadObjectTask_SetLogs(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpUploadObjectTask_GetLogs(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpUploadObjectTask_AppendLog(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpUploadObjectTask_SetError(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpUploadObjectTask_SetObjectInfo(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpUploadObjectTask_SetStorageParams(t *testing.T) {
	m := &GfSpUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestInitResumableUploadObjectTask(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{}
	m.InitResumableUploadObjectTask(1, mockObjectInfo, mockStorageParams, 0, true, 1)
}

func TestGfSpResumableUploadObjectTask_GetResumeOffset(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetResumeOffset()
}

func TestGfSpResumableUploadObjectTask_SetResumeOffset(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetResumeOffset(1)
}

func TestGfSpResumableUploadObjectTask_SetCompleted(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCompleted(true)
}

func TestGfSpResumableUploadObjectTask_Key(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpResumableUploadObjectTask_Type(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskUpload, result)
}

func TestGfSpResumableUploadObjectTask_Info(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpResumableUploadObjectTask_GetAddress(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpResumableUploadObjectTask_SetAddress(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpResumableUploadObjectTask_GetCreateTime(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpResumableUploadObjectTask_SetCreateTime(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpResumableUploadObjectTask_GetUpdateTime(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpResumableUploadObjectTask_SetUpdateTime(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpResumableUploadObjectTask_GetTimeout(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpResumableUploadObjectTask_SetTimeout(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpResumableUploadObjectTask_ExceedTimeout(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpResumableUploadObjectTask_GetRetry(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpResumableUploadObjectTask_IncRetry(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpResumableUploadObjectTask_SetRetry(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpResumableUploadObjectTask_GetMaxRetry(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpResumableUploadObjectTask_SetMaxRetry(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpResumableUploadObjectTask_ExceedRetry(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpResumableUploadObjectTask_Expired(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpResumableUploadObjectTask_GetPriority(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpResumableUploadObjectTask_SetPriority(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpResumableUploadObjectTask_EstimateLimit1(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpResumableUploadObjectTask_EstimateLimit2(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task: &GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			PayloadSize: 10,
		},
		StorageParams: &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 2}},
	}
	m.EstimateLimit()
}

func TestGfSpResumableUploadObjectTask_Error(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpResumableUploadObjectTask_GetUserAddress(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpResumableUploadObjectTask_SetUserAddress(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpResumableUploadObjectTask_SetLogs(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpResumableUploadObjectTask_GetLogs(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpResumableUploadObjectTask_AppendLog(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpResumableUploadObjectTask_SetError(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpResumableUploadObjectTask_SetObjectInfo(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpResumableUploadObjectTask_SetStorageParams(t *testing.T) {
	m := &GfSpResumableUploadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestInitReplicatePieceTask(t *testing.T) {
	m := &GfSpReplicatePieceTask{}
	m.InitReplicatePieceTask(mockObjectInfo, mockStorageParams, coretask.MaxTaskPriority, 0, 0)
}

func TestGfSpReplicatePieceTask_Key(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpReplicatePieceTask_Type(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskReplicatePiece, result)
}

func TestGfSpReplicatePieceTask_Info(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpReplicatePieceTask_GetAddress(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpReplicatePieceTask_SetAddress(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpReplicatePieceTask_GetCreateTime(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpReplicatePieceTask_SetCreateTime(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpReplicatePieceTask_GetUpdateTime(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpReplicatePieceTask_SetUpdateTime(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpReplicatePieceTask_GetTimeout(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpReplicatePieceTask_SetTimeout(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpReplicatePieceTask_ExceedTimeout(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpReplicatePieceTask_GetRetry(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpReplicatePieceTask_IncRetry(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpReplicatePieceTask_SetRetry(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpReplicatePieceTask_GetMaxRetry(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpReplicatePieceTask_SetMaxRetry(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpReplicatePieceTask_ExceedRetry(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpReplicatePieceTask_Expired(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpReplicatePieceTask_GetPriority(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpReplicatePieceTask_SetPriority(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpReplicatePieceTask_EstimateLimit1(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task: &GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			PayloadSize:    10,
			RedundancyType: 1,
		},
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpResumableReplicateObjectTask_EstimateLimit2(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task: &GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			PayloadSize:    10,
			RedundancyType: 1,
		},
		StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
	}
	m.EstimateLimit()
}

func TestGfSpReplicatePieceTask_EstimateLimit3(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpResumableReplicateObjectTask_EstimateLimit4(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task: &GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			PayloadSize: 10,
		},
		StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
	}
	m.EstimateLimit()
}

func TestGfSpReplicatePieceTask_Error(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpReplicatePieceTask_GetUserAddress(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpReplicatePieceTask_SetUserAddress(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpReplicatePieceTask_SetLogs(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpReplicatePieceTask_GetLogs(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpReplicatePieceTask_AppendLog(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpReplicatePieceTask_SetError(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpReplicatePieceTask_SetObjectInfo(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpReplicatePieceTask_SetStorageParams(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestGfSpReplicatePieceTask_SetNotAvailableSpIdx(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetNotAvailableSpIdx(1)
}

func TestGfSpReplicatePieceTask_SetSealed(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSealed(true)
}

func TestGfSpReplicatePieceTask_SetSecondarySignatures(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSecondarySignatures([][]byte{[]byte("1")})
}

func TestGfSpReplicatePieceTask_SetSecondaryAddresses(t *testing.T) {
	m := &GfSpReplicatePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSecondaryAddresses([]string{"1"})
}

func TestInitSealObjectTask(t *testing.T) {
	m := &GfSpSealObjectTask{}
	m.InitSealObjectTask(1, mockObjectInfo, mockStorageParams, coretask.MaxTaskPriority, []string{"1"},
		[][]byte{[]byte("1")}, 0, 0)
}

func TestGfSpSealObjectTask_Key(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpSealObjectTask_Type(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskSealObject, result)
}

func TestGfSpSealObjectTask_Info(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpSealObjectTask_GetAddress(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpSealObjectTask_SetAddress(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpSealObjectTask_GetCreateTime(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpSealObjectTask_SetCreateTime(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpSealObjectTask_GetUpdateTime(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpSealObjectTask_SetUpdateTime(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpSealObjectTask_GetTimeout(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpSealObjectTask_SetTimeout(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpSealObjectTask_ExceedTimeout(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpSealObjectTask_GetRetry(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpSealObjectTask_IncRetry(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpSealObjectTask_SetRetry(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpSealObjectTask_GetMaxRetry(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpSealObjectTask_SetMaxRetry(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpSealObjectTask_ExceedRetry(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpSealObjectTask_Expired(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpSealObjectTask_GetPriority(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpSealObjectTask_SetPriority(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpSealObjectTask_EstimateLimit(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpSealObjectTask_SetSecondaryAddresses(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSecondaryAddresses([]string{"1"})
}

func TestGfSpSealObjectTask_SetSecondarySignatures(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSecondarySignatures([][]byte{[]byte("1")})
}

func TestGfSpSealObjectTask_Error(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpSealObjectTask_GetUserAddress(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpSealObjectTask_SetUserAddress(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpSealObjectTask_SetLogs(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpSealObjectTask_GetLogs(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpSealObjectTask_AppendLog(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpSealObjectTask_SetError(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpSealObjectTask_SetObjectInfo(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpSealObjectTask_SetStorageParams(t *testing.T) {
	m := &GfSpSealObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestInitReceivePieceTask(t *testing.T) {
	m := &GfSpReceivePieceTask{}
	m.InitReceivePieceTask(1, mockObjectInfo, mockStorageParams, coretask.MaxTaskPriority, 1, 0, 0)
}

func TestGfSpReceivePieceTask_GetSignBytes(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetSignBytes()
}

func TestGfSpReceivePieceTask_Key(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpReceivePieceTask_Type(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskReceivePiece, result)
}

func TestGfSpReceivePieceTask_Info(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpReceivePieceTask_GetAddress(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpReceivePieceTask_SetAddress(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpReceivePieceTask_GetCreateTime(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpReceivePieceTask_SetCreateTime(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpReceivePieceTask_GetUpdateTime(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpReceivePieceTask_SetUpdateTime(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpReceivePieceTask_GetTimeout(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpReceivePieceTask_SetTimeout(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpReceivePieceTask_ExceedTimeout(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpReceivePieceTask_GetRetry(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpReceivePieceTask_IncRetry(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpReceivePieceTask_SetRetry(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpReceivePieceTask_GetMaxRetry(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpReceivePieceTask_SetMaxRetry(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpReceivePieceTask_ExceedRetry(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpReceivePieceTask_Expired(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpReceivePieceTask_GetPriority(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpReceivePieceTask_SetPriority(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpReceivePieceTask_EstimateLimit(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpReceivePieceTask_Error(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpReceivePieceTask_GetUserAddress(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpReceivePieceTask_SetUserAddress(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpReceivePieceTask_SetLogs(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpReceivePieceTask_GetLogs(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpReceivePieceTask_AppendLog(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpReceivePieceTask_SetError(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpReceivePieceTask_SetObjectInfo(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpReceivePieceTask_SetStorageParams(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestGfSpReceivePieceTask_SetSegmentIdx(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSegmentIdx(1)
}

func TestGfSpReceivePieceTask_SetRedundancyIdx(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRedundancyIdx(1)
}

func TestGfSpReceivePieceTask_SetPieceSize(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPieceSize(1)
}

func TestGfSpReceivePieceTask_SetPieceChecksum(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPieceChecksum([]byte("test"))
}

func TestGfSpReceivePieceTask_SetSealed(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSealed(true)
}

func TestGfSpReceivePieceTask_SetFinished(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetFinished(true)
}

func TestGfSpReceivePieceTask_SetSignature(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSignature([]byte("test"))
}

func TestGfSpReceivePieceTask_GetGlobalVirtualGroupID(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.GetGlobalVirtualGroupID()
}

func TestGfSpReceivePieceTask_SetGlobalVirtualGroupID(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetGlobalVirtualGroupID(1)
}

func TestGfSpReceivePieceTask_SetBucketMigration(t *testing.T) {
	m := &GfSpReceivePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.SetBucketMigration(true)
}
