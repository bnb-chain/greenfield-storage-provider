package gfsptask

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	mockGVG = &virtualgrouptypes.GlobalVirtualGroup{
		Id:                    1,
		FamilyId:              2,
		PrimarySpId:           3,
		SecondarySpIds:        []uint32{0, 1, 2},
		StoredSize:            10000,
		VirtualPaymentAddress: "mockVirtualPaymentAddress",
		TotalDeposit:          sdkmath.NewInt(1),
	}
	mockSP = &sptypes.StorageProvider{
		Id:                 1,
		OperatorAddress:    "mockOperatorAddress",
		FundingAddress:     "mockFundingAddress",
		SealAddress:        "mockSealAddress",
		ApprovalAddress:    "mockApprovalAddress",
		GcAddress:          "mockGcAddress",
		MaintenanceAddress: "mockMaintenanceAddress",
		TotalDeposit:       sdkmath.NewInt(1),
		Endpoint:           "mockEndpoint",
	}
)

func TestInitMigrateGVGTask(t *testing.T) {
	m := &GfSpMigrateGVGTask{}
	m.InitMigrateGVGTask(coretask.MaxTaskPriority, 1, mockGVG, 0, mockSP, 0, 0)
}

func TestGfSpMigrateGVGTask_Key(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.Key()
}

func TestGfSpMigrateGVGTask_Type(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskMigrateGVG, result)
}

func TestGfSpMigrateGVGTask_Info(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.Info()
}

func TestGfSpMigrateGVGTask_GetAddress(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetAddress()
}

func TestGfSpMigrateGVGTask_SetAddress(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetAddress("mockAddress")
}

func TestGfSpMigrateGVGTask_GetCreateTime(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetCreateTime()
}

func TestGfSpMigrateGVGTask_SetCreateTime(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetCreateTime(1)
}

func TestGfSpMigrateGVGTask_GetUpdateTime(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetUpdateTime()
}

func TestGfSpMigrateGVGTask_SetUpdateTime(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetUpdateTime(1)
}

func TestGfSpMigrateGVGTask_GetTimeout(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetTimeout()
}

func TestGfSpMigrateGVGTask_SetTimeout(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetTimeout(1)
}

func TestGfSpMigrateGVGTask_ExceedTimeout(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.ExceedTimeout()
}

func TestGfSpMigrateGVGTask_GetRetry(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetRetry()
}

func TestGfSpMigrateGVGTask_IncRetry(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.IncRetry()
}

func TestGfSpMigrateGVGTask_SetRetry(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetRetry(1)
}

func TestGfSpMigrateGVGTask_GetMaxRetry(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetMaxRetry()
}

func TestGfSpMigrateGVGTask_SetMaxRetry(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetMaxRetry(1)
}

func TestGfSpMigrateGVGTask_ExceedRetry(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.ExceedRetry()
}

func TestGfSpMigrateGVGTask_Expired(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.Expired()
}

func TestGfSpMigrateGVGTask_GetPriority(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetPriority()
}

func TestGfSpMigrateGVGTask_SetPriority(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetPriority(1)
}

func TestGfSpMigrateGVGTask_EstimateLimit(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.EstimateLimit()
}

func TestGfSpMigrateGVGTask_Error(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	_ = m.Error()
}

func TestGfSpMigrateGVGTask_GetUserAddress(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetUserAddress()
}

func TestGfSpMigrateGVGTask_SetUserAddress(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetUserAddress("mock")
}

func TestGfSpMigrateGVGTask_SetLogs(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetLogs("mock")
}

func TestGfSpMigrateGVGTask_GetLogs(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetLogs()
}

func TestGfSpMigrateGVGTask_AppendLog(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.AppendLog("mock")
}

func TestGfSpMigrateGVGTask_SetError(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetError(nil)
}

func TestGfSpMigrateGVGTask_SetSrcGvg(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetSrcGvg(mockGVG)
}

func TestGfSpMigrateGVGTask_SetDestGvg(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetDestGvg(mockGVG)
}

func TestGfSpMigrateGVGTask_SetSrcSp(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetSrcSp(mockSP)
}

func TestGfSpMigrateGVGTask_GetBucketID(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetBucketID()
}

func TestGfSpMigrateGVGTask_SetBucketID(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetBucketID(1)
}

func TestGfSpMigrateGVGTask_SetRedundancyIdx(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetRedundancyIdx(1)
}

func TestGfSpMigrateGVGTask_GetLastMigratedObjectID(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetLastMigratedObjectID()
}

func TestGfSpMigrateGVGTask_SetLastMigratedObjectID(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetLastMigratedObjectID(1)
}

func TestGfSpMigrateGVGTask_SetFinished(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.SetFinished(true)
}

func TestGfSpMigrateGVGTask_GetSignBytes(t *testing.T) {
	m := &GfSpMigrateGVGTask{
		Task:     &GfSpTask{},
		BucketId: 1,
		SrcGvg:   mockGVG,
		DestGvg:  mockGVG,
		SrcSp:    mockSP,
	}
	m.GetSignBytes()
}

func TestGfSpMigratePieceTask_Key(t *testing.T) {
	m := &GfSpMigratePieceTask{
		Task:          &GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	m.Key()
}
