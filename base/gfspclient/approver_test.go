package gfspclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestGfSpClient_AskCreateBucketApproval(t *testing.T) {
	cases := []struct {
		name          string
		task          coretask.ApprovalCreateBucketTask
		wantedResult1 bool
		wantedIsErr   bool
		wantedErr     error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpCreateBucketApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName3},
			},
			wantedResult1: true,
			wantedIsErr:   false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpCreateBucketApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName1},
			},
			wantedResult1: false,
			wantedIsErr:   true,
			wantedErr:     mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpCreateBucketApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName2},
			},
			wantedResult1: false,
			wantedIsErr:   true,
			wantedErr:     ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			defer s.Close()
			result1, result2, err := s.AskCreateBucketApproval(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult1, result1)
				assert.Nil(t, result2)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult1, result1)
				assert.NotNil(t, result2)
			}
		})
	}
}

func TestGfSpClient_AskCreateBucketApprovalFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect approver")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result1, result2, err := s.AskCreateBucketApproval(ctx, &gfsptask.GfSpCreateBucketApprovalTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result1)
	assert.Nil(t, result2)
}

func TestGfSpClient_AskMigrateBucketApproval(t *testing.T) {
	cases := []struct {
		name          string
		task          coretask.ApprovalMigrateBucketTask
		wantedResult1 bool
		wantedIsErr   bool
		wantedErr     error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpMigrateBucketApprovalTask{
				Task:              &gfsptask.GfSpTask{},
				MigrateBucketInfo: &storagetypes.MsgMigrateBucket{BucketName: mockBucketName3},
			},
			wantedResult1: true,
			wantedIsErr:   false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpMigrateBucketApprovalTask{
				Task:              &gfsptask.GfSpTask{},
				MigrateBucketInfo: &storagetypes.MsgMigrateBucket{BucketName: mockBucketName1},
			},
			wantedResult1: false,
			wantedIsErr:   true,
			wantedErr:     mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpMigrateBucketApprovalTask{
				Task:              &gfsptask.GfSpTask{},
				MigrateBucketInfo: &storagetypes.MsgMigrateBucket{BucketName: mockBucketName2},
			},
			wantedResult1: false,
			wantedIsErr:   true,
			wantedErr:     ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			defer s.Close()
			result1, result2, err := s.AskMigrateBucketApproval(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult1, result1)
				assert.Nil(t, result2)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult1, result1)
				assert.NotNil(t, result2)
			}
		})
	}
}

func TestGfSpClient_AskMigrateBucketApprovalFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect approver")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result1, result2, err := s.AskMigrateBucketApproval(ctx, &gfsptask.GfSpMigrateBucketApprovalTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result1)
	assert.Nil(t, result2)
}

func TestGfSpClient_AskCreateObjectApproval(t *testing.T) {
	cases := []struct {
		name          string
		task          coretask.ApprovalCreateObjectTask
		wantedResult1 bool
		wantedIsErr   bool
		wantedErr     error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpCreateObjectApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateObjectInfo: &storagetypes.MsgCreateObject{ObjectName: mockObjectName3},
			},
			wantedResult1: true,
			wantedIsErr:   false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpCreateObjectApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateObjectInfo: &storagetypes.MsgCreateObject{ObjectName: mockObjectName1},
			},
			wantedResult1: false,
			wantedIsErr:   true,
			wantedErr:     mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpCreateObjectApprovalTask{
				Task:             &gfsptask.GfSpTask{},
				CreateObjectInfo: &storagetypes.MsgCreateObject{ObjectName: mockObjectName2},
			},
			wantedResult1: false,
			wantedIsErr:   true,
			wantedErr:     ErrNoSuchObject,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			defer s.Close()
			result1, result2, err := s.AskCreateObjectApproval(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult1, result1)
				assert.Nil(t, result2)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult1, result1)
				assert.NotNil(t, result2)
			}
		})
	}
}

func TestGfSpClient_AskCreateObjectApprovalFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect approver")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result1, result2, err := s.AskCreateObjectApproval(ctx, &gfsptask.GfSpCreateObjectApprovalTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result1)
	assert.Nil(t, result2)
}
