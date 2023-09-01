package gfspclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func TestGfSpClient_SignCreateBucketApproval(t *testing.T) {
	cases := []struct {
		name        string
		bucket      *storagetypes.MsgCreateBucket
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			bucket:      &storagetypes.MsgCreateBucket{BucketName: mockBucketName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			bucket:      &storagetypes.MsgCreateBucket{BucketName: mockBucketName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			bucket:      &storagetypes.MsgCreateBucket{BucketName: mockBucketName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignCreateBucketApproval(ctx, tt.bucket)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignCreateBucketApprovalFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignCreateBucketApproval(ctx, &storagetypes.MsgCreateBucket{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SignMigrateBucketApproval(t *testing.T) {
	cases := []struct {
		name        string
		bucket      *storagetypes.MsgMigrateBucket
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			bucket:      &storagetypes.MsgMigrateBucket{BucketName: mockBucketName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			bucket:      &storagetypes.MsgMigrateBucket{BucketName: mockBucketName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			bucket:      &storagetypes.MsgMigrateBucket{BucketName: mockBucketName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignMigrateBucketApproval(ctx, tt.bucket)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignMigrateBucketApprovalFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignMigrateBucketApproval(ctx, &storagetypes.MsgMigrateBucket{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SignCreateObjectApproval(t *testing.T) {
	cases := []struct {
		name        string
		object      *storagetypes.MsgCreateObject
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			object:      &storagetypes.MsgCreateObject{ObjectName: mockObjectName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			object:      &storagetypes.MsgCreateObject{ObjectName: mockObjectName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			object:      &storagetypes.MsgCreateObject{ObjectName: mockObjectName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignCreateObjectApproval(ctx, tt.object)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignCreateObjectApprovalFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignCreateObjectApproval(ctx, &storagetypes.MsgCreateObject{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SealObject(t *testing.T) {
	cases := []struct {
		name        string
		object      *storagetypes.MsgSealObject
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			object:      &storagetypes.MsgSealObject{ObjectName: mockObjectName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			object:      &storagetypes.MsgSealObject{ObjectName: mockObjectName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			object:      &storagetypes.MsgSealObject{ObjectName: mockObjectName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SealObject(ctx, tt.object)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_SealObjectFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SealObject(ctx, &storagetypes.MsgSealObject{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_UpdateSPPrice(t *testing.T) {
	cases := []struct {
		name        string
		price       *sptypes.MsgUpdateSpStoragePrice
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			price:       &sptypes.MsgUpdateSpStoragePrice{FreeReadQuota: 2},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			price:       &sptypes.MsgUpdateSpStoragePrice{FreeReadQuota: 0},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			price:       &sptypes.MsgUpdateSpStoragePrice{FreeReadQuota: 1},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.UpdateSPPrice(ctx, tt.price)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_UpdateSPPriceFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.UpdateSPPrice(ctx, &sptypes.MsgUpdateSpStoragePrice{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_CreateGlobalVirtualGroup(t *testing.T) {
	cases := []struct {
		name        string
		group       *gfspserver.GfSpCreateGlobalVirtualGroup
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			group:       &gfspserver.GfSpCreateGlobalVirtualGroup{VirtualGroupFamilyId: 2},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			group:       &gfspserver.GfSpCreateGlobalVirtualGroup{VirtualGroupFamilyId: 0},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			group:       &gfspserver.GfSpCreateGlobalVirtualGroup{VirtualGroupFamilyId: 1},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			err := s.CreateGlobalVirtualGroup(ctx, tt.group)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpClient_CreateGlobalVirtualGroupFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	err := s.CreateGlobalVirtualGroup(ctx, &gfspserver.GfSpCreateGlobalVirtualGroup{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_RejectUnSealObject(t *testing.T) {
	cases := []struct {
		name        string
		object      *storagetypes.MsgRejectSealObject
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			object:      &storagetypes.MsgRejectSealObject{ObjectName: mockObjectName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			object:      &storagetypes.MsgRejectSealObject{ObjectName: mockObjectName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			object:      &storagetypes.MsgRejectSealObject{ObjectName: mockObjectName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.RejectUnSealObject(ctx, tt.object)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_RejectUnSealObjectFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.RejectUnSealObject(ctx, &storagetypes.MsgRejectSealObject{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_DiscontinueBucket(t *testing.T) {
	cases := []struct {
		name        string
		bucket      *storagetypes.MsgDiscontinueBucket
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			bucket:      &storagetypes.MsgDiscontinueBucket{BucketName: mockBucketName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			bucket:      &storagetypes.MsgDiscontinueBucket{BucketName: mockBucketName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			bucket:      &storagetypes.MsgDiscontinueBucket{BucketName: mockBucketName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.DiscontinueBucket(ctx, tt.bucket)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_DiscontinueBucketFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.DiscontinueBucket(ctx, &storagetypes.MsgDiscontinueBucket{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_SignReplicatePieceApproval(t *testing.T) {
	cases := []struct {
		name        string
		task        coretask.ApprovalReplicatePieceTask
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpReplicatePieceApprovalTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName3}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpReplicatePieceApprovalTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName1}},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpReplicatePieceApprovalTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName2}},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignReplicatePieceApproval(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignReplicatePieceApprovalFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignReplicatePieceApproval(ctx, &gfsptask.GfSpReplicatePieceApprovalTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SignSecondarySealBls(t *testing.T) {
	cases := []struct {
		name        string
		objectID    uint64
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			objectID:    2,
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			objectID:    0,
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			objectID:    1,
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignSecondarySealBls(ctx, tt.objectID, 0, nil)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignSecondarySealBlsFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignSecondarySealBls(ctx, 0, 0, nil)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_SignReceiveTask(t *testing.T) {
	cases := []struct {
		name        string
		receiveTask coretask.ReceivePieceTask
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			receiveTask: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName3}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			receiveTask: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName1}},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			receiveTask: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName2}},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignReceiveTask(ctx, tt.receiveTask)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignReceiveTaskFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignReceiveTask(ctx, &gfsptask.GfSpReceivePieceTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SignRecoveryTask(t *testing.T) {
	cases := []struct {
		name         string
		recoveryTask coretask.RecoveryPieceTask
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name: "success",
			recoveryTask: &gfsptask.GfSpRecoverPieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName3}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			recoveryTask: &gfsptask.GfSpRecoverPieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName1}},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			recoveryTask: &gfsptask.GfSpRecoverPieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName2}},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignRecoveryTask(ctx, tt.recoveryTask)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignRecoveryTaskFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignRecoveryTask(ctx, &gfsptask.GfSpRecoverPieceTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SignP2PPingMsg(t *testing.T) {
	cases := []struct {
		name        string
		ping        *gfspp2p.GfSpPing
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			ping:        &gfspp2p.GfSpPing{SpOperatorAddress: mockObjectName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			ping:        &gfspp2p.GfSpPing{SpOperatorAddress: mockObjectName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			ping:        &gfspp2p.GfSpPing{SpOperatorAddress: mockObjectName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignP2PPingMsg(ctx, tt.ping)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignP2PPingMsgFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignP2PPingMsg(ctx, &gfspp2p.GfSpPing{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SignP2PPongMsg(t *testing.T) {
	cases := []struct {
		name        string
		pong        *gfspp2p.GfSpPong
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			pong:        &gfspp2p.GfSpPong{SpOperatorAddress: mockObjectName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			pong:        &gfspp2p.GfSpPong{SpOperatorAddress: mockObjectName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			pong:        &gfspp2p.GfSpPong{SpOperatorAddress: mockObjectName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignP2PPongMsg(ctx, tt.pong)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignP2PPongMsgFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignP2PPongMsg(ctx, &gfspp2p.GfSpPong{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SignMigratePiece(t *testing.T) {
	cases := []struct {
		name        string
		task        *gfsptask.GfSpMigratePieceTask
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpMigratePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName3}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpMigratePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName1}},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpMigratePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName2}},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignMigratePiece(ctx, tt.task)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignMigratePieceFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignMigratePiece(ctx, &gfsptask.GfSpMigratePieceTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_CompleteMigrateBucket(t *testing.T) {
	cases := []struct {
		name          string
		migrateBucket *storagetypes.MsgCompleteMigrateBucket
		wantedIsErr   bool
		wantedErr     error
	}{
		{
			name:          "success",
			migrateBucket: &storagetypes.MsgCompleteMigrateBucket{BucketName: mockBucketName3},
			wantedIsErr:   false,
		},
		{
			name:          "mock rpc error",
			migrateBucket: &storagetypes.MsgCompleteMigrateBucket{BucketName: mockBucketName1},
			wantedIsErr:   true,
			wantedErr:     mockRPCErr,
		},
		{
			name:          "mock response returns error",
			migrateBucket: &storagetypes.MsgCompleteMigrateBucket{BucketName: mockBucketName2},
			wantedIsErr:   true,
			wantedErr:     ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.CompleteMigrateBucket(ctx, tt.migrateBucket)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_CompleteMigrateBucketFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.CompleteMigrateBucket(ctx, &storagetypes.MsgCompleteMigrateBucket{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_SignSecondarySPMigrationBucket(t *testing.T) {
	cases := []struct {
		name        string
		signDoc     *storagetypes.SecondarySpMigrationBucketSignDoc
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			signDoc:     &storagetypes.SecondarySpMigrationBucketSignDoc{DstPrimarySpId: 2},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			signDoc:     &storagetypes.SecondarySpMigrationBucketSignDoc{DstPrimarySpId: 0},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			signDoc:     &storagetypes.SecondarySpMigrationBucketSignDoc{DstPrimarySpId: 1},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SignSecondarySPMigrationBucket(ctx, tt.signDoc)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignSecondarySPMigrationBucketFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignSecondarySPMigrationBucket(ctx, &storagetypes.SecondarySpMigrationBucketSignDoc{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SignSwapOut(t *testing.T) {
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
			result, err := s.SignSwapOut(ctx, tt.swapOut)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockSignature, result)
			}
		})
	}
}

func TestGfSpClient_SignSwapOutFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SignSwapOut(ctx, &virtualgrouptypes.MsgSwapOut{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_SwapOut(t *testing.T) {
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
			result, err := s.SwapOut(ctx, tt.swapOut)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_SwapOutFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SwapOut(ctx, &virtualgrouptypes.MsgSwapOut{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_CompleteSwapOut(t *testing.T) {
	cases := []struct {
		name            string
		completeSwapOut *virtualgrouptypes.MsgCompleteSwapOut
		wantedIsErr     bool
		wantedErr       error
	}{
		{
			name:            "success",
			completeSwapOut: &virtualgrouptypes.MsgCompleteSwapOut{GlobalVirtualGroupFamilyId: 2},
			wantedIsErr:     false,
		},
		{
			name:            "mock rpc error",
			completeSwapOut: &virtualgrouptypes.MsgCompleteSwapOut{GlobalVirtualGroupFamilyId: 0},
			wantedIsErr:     true,
			wantedErr:       mockRPCErr,
		},
		{
			name:            "mock response returns error",
			completeSwapOut: &virtualgrouptypes.MsgCompleteSwapOut{GlobalVirtualGroupFamilyId: 1},
			wantedIsErr:     true,
			wantedErr:       ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.CompleteSwapOut(ctx, tt.completeSwapOut)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_CompleteSwapOutFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.CompleteSwapOut(ctx, &virtualgrouptypes.MsgCompleteSwapOut{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_SPExit(t *testing.T) {
	cases := []struct {
		name        string
		spExit      *virtualgrouptypes.MsgStorageProviderExit
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			spExit:      &virtualgrouptypes.MsgStorageProviderExit{StorageProvider: mockObjectName3},
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			spExit:      &virtualgrouptypes.MsgStorageProviderExit{StorageProvider: mockObjectName1},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			spExit:      &virtualgrouptypes.MsgStorageProviderExit{StorageProvider: mockObjectName2},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.SPExit(ctx, tt.spExit)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_SPExitFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.SPExit(ctx, &virtualgrouptypes.MsgStorageProviderExit{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_CompleteSPExit(t *testing.T) {
	cases := []struct {
		name           string
		completeSPExit *virtualgrouptypes.MsgCompleteStorageProviderExit
		wantedIsErr    bool
		wantedErr      error
	}{
		{
			name:           "success",
			completeSPExit: &virtualgrouptypes.MsgCompleteStorageProviderExit{StorageProvider: mockObjectName3},
			wantedIsErr:    false,
		},
		{
			name:           "mock rpc error",
			completeSPExit: &virtualgrouptypes.MsgCompleteStorageProviderExit{StorageProvider: mockObjectName1},
			wantedIsErr:    true,
			wantedErr:      mockRPCErr,
		},
		{
			name:           "mock response returns error",
			completeSPExit: &virtualgrouptypes.MsgCompleteStorageProviderExit{StorageProvider: mockObjectName2},
			wantedIsErr:    true,
			wantedErr:      ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			result, err := s.CompleteSPExit(ctx, tt.completeSPExit)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, mockTxHash, result)
			}
		})
	}
}

func TestGfSpClient_CompleteSPExitFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect signer")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.CompleteSPExit(ctx, &virtualgrouptypes.MsgCompleteStorageProviderExit{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}
