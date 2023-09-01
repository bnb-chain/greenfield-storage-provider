package gfspclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
)

func TestGfSpClient_AskSecondaryReplicatePieceApproval(t *testing.T) {
	cases := []struct {
		name        string
		low         int
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			low:         0,
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			low:         -2,
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			low:         -1,
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			ta := &gfsptask.GfSpReplicatePieceApprovalTask{}
			result, err := s.AskSecondaryReplicatePieceApproval(ctx, ta, tt.low, 0, 0)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, 1, len(result))
			}
		})
	}
}

func TestGfSpClient_AskSecondaryReplicatePieceApprovalFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect p2p")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.AskSecondaryReplicatePieceApproval(ctx, &gfsptask.GfSpReplicatePieceApprovalTask{}, 0, 0, 0)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_QueryP2PBootstrap(t *testing.T) {
	cases := []struct {
		name         string
		value        string
		wantedResult []string
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:        "success",
			value:       mockObjectName3,
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			value:       mockObjectName1,
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			value:       mockObjectName2,
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := setup(t, ctx)
			md := metadata.Pairs(mockBufNet, tt.value)
			ctx1 := metadata.NewOutgoingContext(ctx, md)
			result, err := s.QueryP2PBootstrap(ctx1)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, 1, len(result))
			}
		})
	}
}

func TestGfSpClient_QueryP2PBootstrapFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect p2p")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.QueryP2PBootstrap(ctx)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}
