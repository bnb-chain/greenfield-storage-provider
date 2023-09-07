package gfspclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestGfSpClient_QueryTasks(t *testing.T) {
	cases := []struct {
		name        string
		subKey      string
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name:        "success",
			subKey:      mockObjectName3,
			wantedIsErr: false,
		},
		{
			name:        "mock rpc error",
			subKey:      mockObjectName1,
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name:        "mock response returns error",
			subKey:      mockObjectName2,
			wantedIsErr: true,
			wantedErr:   ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := mockBufClient()
			result, err := s.QueryTasks(ctx, mockAddress, tt.subKey, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, []string{mockBufNet}, result)
			}
		})
	}
}

func TestGfSpClient_QueryTasksFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect gfsp server")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.QueryTasks(ctx, mockAddress, mockObjectName1)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_QueryBucketMigrate(t *testing.T) {
	cases := []struct {
		name        string
		value       string
		wantedIsErr bool
		wantedErr   error
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
			s := mockBufClient()
			md := metadata.Pairs(mockBufNet, tt.value)
			ctx1 := metadata.NewOutgoingContext(ctx, md)
			result, err := s.QueryBucketMigrate(ctx1, mockAddress, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestGfSpClient_QueryBucketMigrateFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect gfsp server")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.QueryBucketMigrate(ctx, mockAddress)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}

func TestGfSpClient_QuerySPExit(t *testing.T) {
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
			s := mockBufClient()
			md := metadata.Pairs(mockBufNet, tt.value)
			ctx1 := metadata.NewOutgoingContext(context.Background(), md)
			result, err := s.QuerySPExit(ctx1, mockAddress, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestGfSpClient_QuerySPExitFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect gfsp server")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.QuerySPExit(ctx, mockAddress)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Empty(t, result)
}
