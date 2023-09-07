package gfsptask

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield/types/common"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	mockCreateBucketInfo = &storagetypes.MsgCreateBucket{
		Creator:    "mockCreator",
		BucketName: "mockBucketName",
	}
	mockFingerprint = []byte("mockFingerprint")
)

func TestInitApprovalCreateBucketTask(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{}
	m.InitApprovalCreateBucketTask("mockUserAddress", mockCreateBucketInfo, mockFingerprint, coretask.MaxTaskPriority)
}

func TestGfSpCreateBucketApprovalTask_Key(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.Key()
}

func TestGfSpCreateBucketApprovalTask_Type(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskCreateBucketApproval, result)
}

func TestGfSpCreateBucketApprovalTask_Info(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.Info()
}

func TestGfSpCreateBucketApprovalTask_GetAddress(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetAddress()
}

func TestGfSpCreateBucketApprovalTask_SetAddress(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpCreateBucketApprovalTask_GetCreateTime(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetCreateTime()
}

func TestGfSpCreateBucketApprovalTask_SetCreateTime(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetCreateTime(1)
}

func TestGfSpCreateBucketApprovalTask_GetUpdateTime(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetUpdateTime()
}

func TestGfSpCreateBucketApprovalTask_SetUpdateTime(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetUpdateTime(1)
}

func TestGfSpCreateBucketApprovalTask_GetTimeout(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetTimeout()
}

func TestGfSpCreateBucketApprovalTask_SetTimeout(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetTimeout(1)
}

func TestGfSpCreateBucketApprovalTask_ExceedTimeout(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.ExceedTimeout()
}

func TestGfSpCreateBucketApprovalTask_GetRetry(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetRetry()
}

func TestGfSpCreateBucketApprovalTask_IncRetry(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.IncRetry()
}

func TestGfSpCreateBucketApprovalTask_SetRetry(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetRetry(1)
}

func TestGfSpCreateBucketApprovalTask_GetMaxRetry(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetMaxRetry()
}

func TestGfSpCreateBucketApprovalTask_SetMaxRetry(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetMaxRetry(1)
}

func TestGfSpCreateBucketApprovalTask_ExceedRetry(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.ExceedRetry()
}

func TestGfSpCreateBucketApprovalTask_Expired(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.Expired()
}

func TestGfSpCreateBucketApprovalTask_GetPriority(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetPriority()
}

func TestGfSpCreateBucketApprovalTask_SetPriority(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetPriority(1)
}

func TestGfSpCreateBucketApprovalTask_EstimateLimit(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.EstimateLimit()
}

func TestGfSpCreateBucketApprovalTask_Error(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	_ = m.Error()
}

func TestGfSpCreateBucketApprovalTask_GetUserAddress(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetUserAddress()
}

func TestGfSpCreateBucketApprovalTask_SetUserAddress(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetUserAddress("mock")
}

func TestGfSpCreateBucketApprovalTask_SetLogs(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetLogs("mock")
}

func TestGfSpCreateBucketApprovalTask_GetLogs(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetLogs()
}

func TestGfSpCreateBucketApprovalTask_AppendLog(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.AppendLog("mock")
}

func TestGfSpCreateBucketApprovalTask_SetError(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetError(nil)
}

func TestGfSpCreateBucketApprovalTask_SetExpiredHeight(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task: &GfSpTask{},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator:           "mockCreator",
			BucketName:        "mockBucketName",
			PrimarySpApproval: &common.Approval{},
		},
		Fingerprint: mockFingerprint,
	}
	m.SetExpiredHeight(1)
}

func TestGfSpCreateBucketApprovalTask_GetExpiredHeight(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:             &GfSpTask{},
		CreateBucketInfo: mockCreateBucketInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetExpiredHeight()
}

func TestGfSpCreateBucketApprovalTask_SetCreateBucketInfo(t *testing.T) {
	m := &GfSpCreateBucketApprovalTask{
		Task:        &GfSpTask{},
		Fingerprint: mockFingerprint,
	}
	m.SetCreateBucketInfo(mockCreateBucketInfo)
}

var (
	mockMigrateBucketInfo = &storagetypes.MsgMigrateBucket{
		Operator:             "mockOperator",
		BucketName:           "mockBucketName",
		DstPrimarySpId:       01,
		DstPrimarySpApproval: &common.Approval{},
	}
)

func TestInitApprovalMigrateBucketTask(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{}
	m.InitApprovalMigrateBucketTask(mockMigrateBucketInfo, coretask.MaxTaskPriority)
}

func TestGfSpMigrateBucketApprovalTask_Key(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.Key()
}

func TestGfSpMigrateBucketApprovalTask_Type(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskMigrateBucketApproval, result)
}

func TestGfSpMigrateBucketApprovalTask_Info(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.Info()
}

func TestGfSpMigrateBucketApprovalTask_GetAddress(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetAddress()
}

func TestGfSpMigrateBucketApprovalTask_SetAddress(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpMigrateBucketApprovalTask_GetCreateTime(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetCreateTime()
}

func TestGfSpMigrateBucketApprovalTask_SetCreateTime(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetCreateTime(1)
}

func TestGfSpMigrateBucketApprovalTask_GetUpdateTime(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetUpdateTime()
}

func TestGfSpMigrateBucketApprovalTask_SetUpdateTime(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetUpdateTime(1)
}

func TestGfSpMigrateBucketApprovalTask_GetTimeout(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetTimeout()
}

func TestGfSpMigrateBucketApprovalTask_SetTimeout(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetTimeout(1)
}

func TestGfSpMigrateBucketApprovalTask_ExceedTimeout(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.ExceedTimeout()
}

func TestGfSpMigrateBucketApprovalTask_GetRetry(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetRetry()
}

func TestGfSpMigrateBucketApprovalTask_IncRetry(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.IncRetry()
}

func TestGfSpMigrateBucketApprovalTask_SetRetry(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetRetry(1)
}

func TestGfSpMigrateBucketApprovalTask_GetMaxRetry(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetMaxRetry()
}

func TestGfSpMigrateBucketApprovalTask_SetMaxRetry(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetMaxRetry(1)
}

func TestGfSpMigrateBucketApprovalTask_ExceedRetry(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.ExceedRetry()
}

func TestGfSpMigrateBucketApprovalTask_Expired(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.Expired()
}

func TestGfSpMigrateBucketApprovalTask_GetPriority(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetPriority()
}

func TestGfSpMigrateBucketApprovalTask_SetPriority(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetPriority(1)
}

func TestGfSpMigrateBucketApprovalTask_EstimateLimit(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.EstimateLimit()
}

func TestGfSpMigrateBucketApprovalTask_Error(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	_ = m.Error()
}

func TestGfSpMigrateBucketApprovalTask_GetUserAddress(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetUserAddress()
}

func TestGfSpMigrateBucketApprovalTask_SetUserAddress(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetUserAddress("mock")
}

func TestGfSpMigrateBucketApprovalTask_SetLogs(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetLogs("mock")
}

func TestGfSpMigrateBucketApprovalTask_GetLogs(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetLogs()
}

func TestGfSpMigrateBucketApprovalTask_AppendLog(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.AppendLog("mock")
}

func TestGfSpMigrateBucketApprovalTask_SetError(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetError(nil)
}

func TestGfSpMigrateBucketApprovalTask_SetExpiredHeight(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.SetExpiredHeight(1)
}

func TestGfSpMigrateBucketApprovalTask_GetExpiredHeight(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task:              &GfSpTask{},
		MigrateBucketInfo: mockMigrateBucketInfo,
	}
	m.GetExpiredHeight()
}

func TestGfSpMigrateBucketApprovalTask_SetMigrateBucketInfo(t *testing.T) {
	m := &GfSpMigrateBucketApprovalTask{
		Task: &GfSpTask{},
	}
	m.SetMigrateBucketInfo(mockMigrateBucketInfo)
}

var (
	mockCreateObjectInfo = &storagetypes.MsgCreateObject{
		Creator:           "mockCreator",
		BucketName:        "mockBucketName",
		ObjectName:        "mockObjectName",
		PayloadSize:       10,
		Visibility:        1,
		ContentType:       "application/json",
		PrimarySpApproval: &common.Approval{},
		ExpectChecksums:   [][]byte{[]byte("test")},
		RedundancyType:    1,
	}
)

func TestInitApprovalCreateObjectTask(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{}
	m.InitApprovalCreateObjectTask("mockUserAddress", mockCreateObjectInfo, mockFingerprint, coretask.MaxTaskPriority)
}

func TestGfSpCreateObjectApprovalTask_Key(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.Key()
}

func TestGfSpCreateObjectApprovalTask_Type(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskCreateObjectApproval, result)
}

func TestGfSpCreateObjectApprovalTask_Info(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.Info()
}

func TestGfSpCreateObjectApprovalTask_GetAddress(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetAddress()
}

func TestGfSpCreateObjectApprovalTask_SetAddress(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpCreateObjectApprovalTask_GetCreateTime(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetCreateTime()
}

func TestGfSpCreateObjectApprovalTask_SetCreateTime(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetCreateTime(1)
}

func TestGfSpCreateObjectApprovalTask_GetUpdateTime(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetUpdateTime()
}

func TestGfSpCreateObjectApprovalTask_SetUpdateTime(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetUpdateTime(1)
}

func TestGfSpCreateObjectApprovalTask_GetTimeout(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetTimeout()
}

func TestGfSpCreateObjectApprovalTask_SetTimeout(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetTimeout(1)
}

func TestGfSpCreateObjectApprovalTask_ExceedTimeout(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.ExceedTimeout()
}

func TestGfSpCreateObjectApprovalTask_GetRetry(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetRetry()
}

func TestGfSpCreateObjectApprovalTask_IncRetry(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.IncRetry()
}

func TestGfSpCreateObjectApprovalTask_SetRetry(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetRetry(1)
}

func TestGfSpCreateObjectApprovalTask_GetMaxRetry(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetMaxRetry()
}

func TestGfSpCreateObjectApprovalTask_SetMaxRetry(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetMaxRetry(1)
}

func TestGfSpCreateObjectApprovalTask_ExceedRetry(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.ExceedRetry()
}

func TestGfSpCreateObjectApprovalTask_Expired(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.Expired()
}

func TestGfSpCreateObjectApprovalTask_GetPriority(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetPriority()
}

func TestGfSpCreateObjectApprovalTask_SetPriority(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetPriority(1)
}

func TestGfSpCreateObjectApprovalTask_EstimateLimit(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.EstimateLimit()
}

func TestGfSpCreateObjectApprovalTask_Error(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	_ = m.Error()
}

func TestGfSpCreateObjectApprovalTask_GetUserAddress(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetUserAddress()
}

func TestGfSpCreateObjectApprovalTask_SetUserAddress(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetUserAddress("mock")
}

func TestGfSpCreateObjectApprovalTask_SetLogs(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetLogs("mock")
}

func TestGfSpCreateObjectApprovalTask_GetLogs(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetLogs()
}

func TestGfSpCreateObjectApprovalTask_AppendLog(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.AppendLog("mock")
}

func TestGfSpCreateObjectApprovalTask_SetError(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetError(nil)
}

func TestGfSpCreateObjectApprovalTask_SetExpiredHeight(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetExpiredHeight(1)
}

func TestGfSpCreateObjectApprovalTask_GetExpiredHeight(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.GetExpiredHeight()
}

func TestGfSpCreateObjectApprovalTask_SetCreateObjectInfo(t *testing.T) {
	m := &GfSpCreateObjectApprovalTask{
		Task:             &GfSpTask{},
		CreateObjectInfo: mockCreateObjectInfo,
		Fingerprint:      mockFingerprint,
	}
	m.SetCreateObjectInfo(mockCreateObjectInfo)
}

var (
	mockObjectInfo = &storagetypes.ObjectInfo{
		Owner:               "mockOwner",
		Creator:             "mockCreator",
		BucketName:          "mockBucketName",
		ObjectName:          "mockObjectName",
		Id:                  sdkmath.NewUint(1),
		LocalVirtualGroupId: 1,
		PayloadSize:         10,
		Visibility:          1,
		ContentType:         "application/json",
		Checksums:           [][]byte{[]byte("1")},
	}
	mockStorageParams = &storagetypes.Params{
		VersionedParams: storagetypes.VersionedParams{
			MaxSegmentSize:          100,
			RedundantDataChunkNum:   3,
			RedundantParityChunkNum: 2,
			MinChargeSize:           10,
		},
		MaxPayloadSize: 10000,
	}
)

func TestInitApprovalReplicatePieceTask(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{}
	m.InitApprovalReplicatePieceTask(mockObjectInfo, mockStorageParams, coretask.MaxTaskPriority, "mockAskSpOperatorAddress")
}

func TestGfSpReplicatePieceApprovalTask_GetSignBytes(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetSignBytes()
}

func TestGfSpReplicatePieceApprovalTask_Key(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.Key()
}

func TestGfSpReplicatePieceApprovalTask_Type(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskReplicatePieceApproval, result)
}

func TestGfSpReplicatePieceApprovalTask_Info(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.Info()
}

func TestGfSpReplicatePieceApprovalTask_GetAddress(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetAddress()
}

func TestGfSpReplicatePieceApprovalTask_SetAddress(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpReplicatePieceApprovalTask_GetCreateTime(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetCreateTime()
}

func TestGfSpReplicatePieceApprovalTask_SetCreateTime(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetCreateTime(1)
}

func TestGfSpReplicatePieceApprovalTask_GetUpdateTime(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetUpdateTime()
}

func TestGfSpReplicatePieceApprovalTask_SetUpdateTime(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetUpdateTime(1)
}

func TestGfSpReplicatePieceApprovalTask_GetTimeout(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetTimeout()
}

func TestGfSpReplicatePieceApprovalTask_SetTimeout(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetTimeout(1)
}

func TestGfSpReplicatePieceApprovalTask_ExceedTimeout(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.ExceedTimeout()
}

func TestGfSpReplicatePieceApprovalTask_GetRetry(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetRetry()
}

func TestGfSpReplicatePieceApprovalTask_IncRetry(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.IncRetry()
}

func TestGfSpReplicatePieceApprovalTask_SetRetry(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetRetry(1)
}

func TestGfSpReplicatePieceApprovalTask_GetMaxRetry(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetMaxRetry()
}

func TestGfSpReplicatePieceApprovalTask_SetMaxRetry(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetMaxRetry(1)
}

func TestGfSpReplicatePieceApprovalTask_ExceedRetry(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.ExceedRetry()
}

func TestGfSpReplicatePieceApprovalTask_Expired(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.Expired()
}

func TestGfSpReplicatePieceApprovalTask_GetPriority(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetPriority()
}

func TestGfSpReplicatePieceApprovalTask_SetPriority(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetPriority(1)
}

func TestGfSpReplicatePieceApprovalTask_EstimateLimit(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.EstimateLimit()
}

func TestGfSpReplicatePieceApprovalTask_Error(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	_ = m.Error()
}

func TestGfSpReplicatePieceApprovalTask_GetUserAddress(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetUserAddress()
}

func TestGfSpReplicatePieceApprovalTask_SetUserAddress(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetUserAddress("mock")
}

func TestGfSpReplicatePieceApprovalTask_SetLogs(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetLogs("mock")
}

func TestGfSpReplicatePieceApprovalTask_GetLogs(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.GetLogs()
}

func TestGfSpReplicatePieceApprovalTask_AppendLog(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.AppendLog("mock")
}

func TestGfSpReplicatePieceApprovalTask_SetError(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetError(nil)
}

func TestGfSpReplicatePieceApprovalTask_SetExpiredHeight(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetExpiredHeight(1)
}

func TestGfSpReplicatePieceApprovalTask_SetObjectInfo(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetObjectInfo(mockObjectInfo)
}

func TestGfSpReplicatePieceApprovalTask_SetStorageParams(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetStorageParams(mockStorageParams)
}

func TestGfSpReplicatePieceApprovalTask_SetAskSpOperatorAddress(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetAskSpOperatorAddress("mock")
}

func TestGfSpReplicatePieceApprovalTask_SetAskSignature(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetAskSignature([]byte("mock"))
}

func TestGfSpReplicatePieceApprovalTask_SetApprovedSpOperatorAddress(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetApprovedSpOperatorAddress("mock")
}

func TestGfSpReplicatePieceApprovalTask_SetApprovedSignature(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetApprovedSignature([]byte("mock"))
}

func TestGfSpReplicatePieceApprovalTask_SetApprovedSpEndpoint(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetApprovedSpEndpoint("mock")
}

func TestGfSpReplicatePieceApprovalTask_SetApprovedSpApprovalAddress(t *testing.T) {
	m := &GfSpReplicatePieceApprovalTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
		ExpiredHeight: 10,
	}
	m.SetApprovedSpApprovalAddress("mock")
}
