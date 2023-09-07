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

func TestGfSpClient_GetObject(t *testing.T) {
	cases := []struct {
		name         string
		task         coretask.DownloadObjectTask
		wantedResult []byte
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpDownloadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3},
			},
			wantedResult: []byte(mockBufNet),
			wantedIsErr:  false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpDownloadObjectTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName1},
			},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpDownloadObjectTask{
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
			s := mockBufClient()
			ctx := context.Background()
			result, err := s.GetObject(ctx, tt.task, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func TestGfSpClient_GetObjectFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect downloader")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.GetObject(ctx, &gfsptask.GfSpDownloadObjectTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_GetPiece(t *testing.T) {
	cases := []struct {
		name         string
		task         coretask.DownloadPieceTask
		wantedResult []byte
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpDownloadPieceTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3},
			},
			wantedResult: []byte(mockBufNet),
			wantedIsErr:  false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpDownloadPieceTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName1},
			},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpDownloadPieceTask{
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
			s := mockBufClient()
			ctx := context.Background()
			result, err := s.GetPiece(ctx, tt.task, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func TestGfSpClient_GetPieceFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect downloader")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.GetPiece(ctx, &gfsptask.GfSpDownloadPieceTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}

func TestGfSpClient_GetChallengeInfo(t *testing.T) {
	cases := []struct {
		name         string
		task         coretask.ChallengePieceTask
		wantedResult []byte
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpChallengePieceTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3},
			},
			wantedResult: []byte(mockBufNet),
			wantedIsErr:  false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpChallengePieceTask{
				Task:       &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName1},
			},
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpChallengePieceTask{
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
			s := mockBufClient()
			ctx := context.Background()
			_, _, result, err := s.GetChallengeInfo(ctx, tt.task, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func TestGfSpClient_GetChallengeInfoFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect downloader")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	_, _, result, err := s.GetChallengeInfo(ctx, &gfsptask.GfSpChallengePieceTask{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Nil(t, result)
}
