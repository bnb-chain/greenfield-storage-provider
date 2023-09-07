package gfspclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestGfSpClient_ReplicatePiece(t *testing.T) {
	cases := []struct {
		name        string
		task        coretask.ReceivePieceTask
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName3}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName1}},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName2}},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := mockBufClient()
			err := s.ReplicatePiece(ctx, tt.task, []byte("mockData"), grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpClient_ReplicatePieceFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect receiver")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	err := s.ReplicatePiece(ctx, &gfsptask.GfSpReceivePieceTask{}, []byte("mockData"))
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_DoneReplicatePiece(t *testing.T) {
	cases := []struct {
		name        string
		task        coretask.ReceivePieceTask
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName3}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName1}},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpReceivePieceTask{ObjectInfo: &storagetypes.ObjectInfo{
				ObjectName: mockObjectName2}},
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := mockBufClient()
			_, result, err := s.DoneReplicatePiece(ctx, tt.task, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func TestGfSpClient_DoneReplicatePieceFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect receiver")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, _, err := s.DoneReplicatePiece(ctx, &gfsptask.GfSpReceivePieceTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}
