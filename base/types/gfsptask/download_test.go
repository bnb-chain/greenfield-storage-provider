package gfsptask

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	mockBucketInfo = &storagetypes.BucketInfo{
		Owner:      "mockOwner",
		BucketName: "mockBucketName",
		Visibility: 1,
		Id:         sdkmath.NewUint(1),
	}
)

func TestInitDownloadObjectTask(t *testing.T) {
	m := &GfSpDownloadObjectTask{}
	m.InitDownloadObjectTask(mockObjectInfo, mockBucketInfo, mockStorageParams, coretask.MaxTaskPriority, "mock", 1, 2, 0, 0)
}

func TestGfSpDownloadObjectTask_Key(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpDownloadObjectTask_Type(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskDownloadObject, result)
}

func TestGfSpDownloadObjectTask_Info(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpDownloadObjectTask_GetAddress(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpDownloadObjectTask_SetAddress(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpDownloadObjectTask_GetCreateTime(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpDownloadObjectTask_SetCreateTime(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpDownloadObjectTask_GetUpdateTime(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpDownloadObjectTask_SetUpdateTime(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpDownloadObjectTask_GetTimeout(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpDownloadObjectTask_SetTimeout(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpDownloadObjectTask_ExceedTimeout(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpDownloadObjectTask_GetRetry(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpDownloadObjectTask_IncRetry(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpDownloadObjectTask_SetRetry(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpDownloadObjectTask_GetMaxRetry(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpDownloadObjectTask_SetMaxRetry(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpDownloadObjectTask_ExceedRetry(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpDownloadObjectTask_Expired(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpDownloadObjectTask_GetPriority(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpDownloadObjectTask_SetPriority(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpDownloadObjectTask_EstimateLimit(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpDownloadObjectTask_Error(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpDownloadObjectTask_GetUserAddress(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpDownloadObjectTask_SetUserAddress(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpDownloadObjectTask_SetLogs(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpDownloadObjectTask_GetLogs(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpDownloadObjectTask_AppendLog(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpDownloadObjectTask_SetError(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpDownloadObjectTask_GetSize1(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
		Low:           10,
		High:          0,
	}
	result := m.GetSize()
	assert.Equal(t, int64(0), result)
}

func TestGfSpDownloadObjectTask_GetSize2(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
		Low:           1,
		High:          5,
	}
	result := m.GetSize()
	assert.Equal(t, int64(5), result)
}

func TestGfSpDownloadObjectTask_SetObjectInfo(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpDownloadObjectTask_SetStorageParams(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestGfSpDownloadObjectTask_SetBucketInfo(t *testing.T) {
	m := &GfSpDownloadObjectTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetBucketInfo(mockBucketInfo)
}

func TestInitDownloadPieceTask(t *testing.T) {
	m := &GfSpDownloadPieceTask{}
	m.InitDownloadPieceTask(mockObjectInfo, mockBucketInfo, mockStorageParams, coretask.MaxTaskPriority, false, "mock",
		10, "mockPieceKey", 1, 2, 0, 0)
}

func TestGfSpDownloadPieceTask_Key(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpDownloadPieceTask_Type(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskDownloadPiece, result)
}

func TestGfSpDownloadPieceTask_Info(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpDownloadPieceTask_GetAddress(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpDownloadPieceTask_SetAddress(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpDownloadPieceTask_GetCreateTime(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpDownloadPieceTask_SetCreateTime(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpDownloadPieceTask_GetUpdateTime(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpDownloadPieceTask_SetUpdateTime(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpDownloadPieceTask_GetTimeout(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpDownloadPieceTask_SetTimeout(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpDownloadPieceTask_ExceedTimeout(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpDownloadPieceTask_GetRetry(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpDownloadPieceTask_IncRetry(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpDownloadPieceTask_SetRetry(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpDownloadPieceTask_GetMaxRetry(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpDownloadPieceTask_SetMaxRetry(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpDownloadPieceTask_ExceedRetry(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpDownloadPieceTask_Expired(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpDownloadPieceTask_GetPriority(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpDownloadPieceTask_SetPriority(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpDownloadPieceTask_EstimateLimit(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpDownloadPieceTask_Error(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpDownloadPieceTask_GetUserAddress(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpDownloadPieceTask_SetUserAddress(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpDownloadPieceTask_SetLogs(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpDownloadPieceTask_GetLogs(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpDownloadPieceTask_AppendLog(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpDownloadPieceTask_SetError(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpDownloadPieceTask_GetSize(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetSize()
}

func TestGfSpDownloadPieceTask_SetObjectInfo(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpDownloadPieceTask_SetStorageParams(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestGfSpDownloadPieceTask_SetBucketInfo(t *testing.T) {
	m := &GfSpDownloadPieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetBucketInfo(mockBucketInfo)
}

func TestInitChallengePieceTask(t *testing.T) {
	m := &GfSpChallengePieceTask{}
	m.InitChallengePieceTask(mockObjectInfo, mockBucketInfo, mockStorageParams, coretask.MaxTaskPriority, "mock", 1,
		2, 1, 2)
}

func TestGfSpChallengePieceTask_Key(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}

func TestGfSpChallengePieceTask_Type(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskChallengePiece, result)
}

func TestGfSpChallengePieceTask_Info(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Info()
}

func TestGfSpChallengePieceTask_GetAddress(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetAddress()
}

func TestGfSpChallengePieceTask_SetAddress(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpChallengePieceTask_GetCreateTime(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetCreateTime()
}

func TestGfSpChallengePieceTask_SetCreateTime(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetCreateTime(1)
}

func TestGfSpChallengePieceTask_GetUpdateTime(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUpdateTime()
}

func TestGfSpChallengePieceTask_SetUpdateTime(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUpdateTime(1)
}

func TestGfSpChallengePieceTask_GetTimeout(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetTimeout()
}

func TestGfSpChallengePieceTask_SetTimeout(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetTimeout(1)
}

func TestGfSpChallengePieceTask_ExceedTimeout(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedTimeout()
}

func TestGfSpChallengePieceTask_GetRetry(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetRetry()
}

func TestGfSpChallengePieceTask_IncRetry(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.IncRetry()
}

func TestGfSpChallengePieceTask_SetRetry(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRetry(1)
}

func TestGfSpChallengePieceTask_GetMaxRetry(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetMaxRetry()
}

func TestGfSpChallengePieceTask_SetMaxRetry(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetMaxRetry(1)
}

func TestGfSpChallengePieceTask_ExceedRetry(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.ExceedRetry()
}

func TestGfSpChallengePieceTask_Expired(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.Expired()
}

func TestGfSpChallengePieceTask_GetPriority(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetPriority()
}

func TestGfSpChallengePieceTask_SetPriority(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPriority(1)
}

func TestGfSpChallengePieceTask_EstimateLimit1(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    &storagetypes.ObjectInfo{PayloadSize: 1},
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
		RedundancyIdx: -1,
	}
	m.EstimateLimit()
}

func TestGfSpChallengePieceTask_EstimateLimit2(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.EstimateLimit()
}

func TestGfSpChallengePieceTask_Error(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	_ = m.Error()
}

func TestGfSpChallengePieceTask_GetUserAddress(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetUserAddress()
}

func TestGfSpChallengePieceTask_SetUserAddress(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetUserAddress("mock")
}

func TestGfSpChallengePieceTask_SetLogs(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetLogs("mock")
}

func TestGfSpChallengePieceTask_GetLogs(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.GetLogs()
}

func TestGfSpChallengePieceTask_AppendLog(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.AppendLog("mock")
}

func TestGfSpChallengePieceTask_SetError(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetError(nil)
}

func TestGfSpChallengePieceTask_SetObjectInfo(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpChallengePieceTask_SetStorageParams(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestGfSpChallengePieceTask_SetBucketInfo(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetBucketInfo(mockBucketInfo)
}

func TestGfSpChallengePieceTask_SetSegmentIdx(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetSegmentIdx(2)
}

func TestGfSpChallengePieceTask_SetRedundancyIdx(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetRedundancyIdx(1)
}

func TestGfSpChallengePieceTask_SetIntegrityHash(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetIntegrityHash([]byte("test"))
}

func TestGfSpChallengePieceTask_SetPieceHash(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPieceHash([][]byte{[]byte("test")})
}

func TestGfSpChallengePieceTask_SetPieceDataSize(t *testing.T) {
	m := &GfSpChallengePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		BucketInfo:    mockBucketInfo,
		StorageParams: mockStorageParams,
	}
	m.SetPieceDataSize(1)
}
