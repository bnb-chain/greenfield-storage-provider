package gfspclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func TestGfSpClient_CreateUploadObject(t *testing.T) {
	cases := []struct {
		name         string
		task         coretask.UploadObjectTask
		wantedResult []byte
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpUploadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3},
			},
			wantedResult: []byte(mockBufNet),
			wantedIsErr:  false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpUploadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName1},
			},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpUploadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName2},
			},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			err := s.CreateUploadObject(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpClient_CreateUploadObjectFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect manager")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	err := s.CreateUploadObject(ctx, &gfsptask.GfSpUploadObjectTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_CreateResumableUploadObject(t *testing.T) {
	cases := []struct {
		name         string
		task         coretask.ResumableUploadObjectTask
		wantedResult []byte
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpResumableUploadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3},
			},
			wantedResult: []byte(mockBufNet),
			wantedIsErr:  false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpResumableUploadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName1},
			},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpResumableUploadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName2},
			},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			err := s.CreateResumableUploadObject(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpClient_CreateResumableUploadObjectFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect manager")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	err := s.CreateResumableUploadObject(ctx, &gfsptask.GfSpResumableUploadObjectTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_AskTask(t *testing.T) {
	cases := []struct {
		name         string
		limit        corercmgr.Limit
		wantedResult coretask.Task
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "success: ReplicatePieceTask",
			limit:        &gfsplimit.GfSpLimit{Memory: 0},
			wantedResult: &gfsptask.GfSpReplicatePieceTask{},
			wantedIsErr:  false,
		},
		{
			name:         "success: SealObjectTask",
			limit:        &gfsplimit.GfSpLimit{Memory: 1},
			wantedResult: &gfsptask.GfSpSealObjectTask{},
			wantedIsErr:  false,
		},
		{
			name:         "success: ReceivePieceTask",
			limit:        &gfsplimit.GfSpLimit{Memory: 2},
			wantedResult: &gfsptask.GfSpReceivePieceTask{},
			wantedIsErr:  false,
		},
		{
			name:         "success: GCObjectTask",
			limit:        &gfsplimit.GfSpLimit{Memory: 3},
			wantedResult: &gfsptask.GfSpGCObjectTask{},
			wantedIsErr:  false,
		},
		{
			name:         "success: GCZombiePieceTask",
			limit:        &gfsplimit.GfSpLimit{Memory: 4},
			wantedResult: &gfsptask.GfSpGCZombiePieceTask{},
			wantedIsErr:  false,
		},
		{
			name:         "success: GCMetaTask",
			limit:        &gfsplimit.GfSpLimit{Memory: 5},
			wantedResult: &gfsptask.GfSpGCMetaTask{},
			wantedIsErr:  false,
		},
		{
			name:         "success: RecoverPieceTask",
			limit:        &gfsplimit.GfSpLimit{Memory: 6},
			wantedResult: &gfsptask.GfSpRecoverPieceTask{},
			wantedIsErr:  false,
		},
		{
			name:         "success: MigrateGvgTask",
			limit:        &gfsplimit.GfSpLimit{Memory: 7},
			wantedResult: &gfsptask.GfSpMigrateGVGTask{},
			wantedIsErr:  false,
		},
		{
			name:         "Failure: ErrTypeMismatch",
			limit:        &gfsplimit.GfSpLimit{Memory: 8},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    ErrTypeMismatch,
		},
		{
			name:         "mock rpc error",
			limit:        &gfsplimit.GfSpLimit{Memory: -2},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name:         "mock response returns error",
			limit:        &gfsplimit.GfSpLimit{Memory: -1},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.AskTask(ctx, tt.limit)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

func TestGfSpClient_AskTaskFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect manager")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.AskTask(ctx, &gfsplimit.GfSpLimit{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_ReportTask(t *testing.T) {
	cases := []struct {
		name        string
		task        coretask.Task
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success: UploadObjectTask",
			task:        &gfsptask.GfSpUploadObjectTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: ResumableUploadObjectTask",
			task:        &gfsptask.GfSpResumableUploadObjectTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: ReplicatePieceTask",
			task:        &gfsptask.GfSpReplicatePieceTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: ReceivePieceTask",
			task:        &gfsptask.GfSpReceivePieceTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: SealObjectTask",
			task:        &gfsptask.GfSpSealObjectTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: GcObjectTask",
			task:        &gfsptask.GfSpGCObjectTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: GcZombiePieceTask",
			task:        &gfsptask.GfSpGCZombiePieceTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: GcMetaTask",
			task:        &gfsptask.GfSpGCMetaTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: DownloadObjectTask",
			task:        &gfsptask.GfSpDownloadObjectTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: ChallengePieceTask",
			task:        &gfsptask.GfSpChallengePieceTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: RecoverPieceTask",
			task:        &gfsptask.GfSpRecoverPieceTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name:        "success: MigrateGvgTask",
			task:        &gfsptask.GfSpMigrateGVGTask{Task: &gfsptask.GfSpTask{}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpUploadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName1},
			},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpUploadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName2},
			},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			err := s.ReportTask(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpClient_ReportTaskFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect manager")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	err := s.ReportTask(ctx, &gfsptask.GfSpUploadObjectTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_PickVirtualGroupFamilyID(t *testing.T) {
	cases := []struct {
		name         string
		task         coretask.ApprovalCreateBucketTask
		wantedResult uint32
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpCreateBucketApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName3},
			},
			wantedResult: 1,
			wantedIsErr:  false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpCreateBucketApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName1},
			},
			wantedResult: 0,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpCreateBucketApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName2},
			},
			wantedResult: 0,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.PickVirtualGroupFamilyID(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, uint32(0), result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, uint32(1), result)
			}
		})
	}
}

func TestGfSpClient_PickVirtualGroupFamilyIDFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect manager")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.PickVirtualGroupFamilyID(ctx, &gfsptask.GfSpCreateBucketApprovalTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, uint32(0), result)
}

func TestGfSpClient_NotifyMigrateSwapOut(t *testing.T) {
	cases := []struct {
		name        string
		swapOut     *virtualgrouptypes.MsgSwapOut
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			swapOut:     &virtualgrouptypes.MsgSwapOut{GlobalVirtualGroupFamilyId: 2},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			swapOut:     &virtualgrouptypes.MsgSwapOut{GlobalVirtualGroupFamilyId: 0},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			swapOut:     &virtualgrouptypes.MsgSwapOut{GlobalVirtualGroupFamilyId: 1},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			err := s.NotifyMigrateSwapOut(ctx, tt.swapOut)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpClient_NotifyMigrateSwapOutFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect manager")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	err := s.NotifyMigrateSwapOut(ctx, &virtualgrouptypes.MsgSwapOut{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
}
