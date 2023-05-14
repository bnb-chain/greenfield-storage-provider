package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"google.golang.org/grpc"
)

func (s *GfSpClient) GetObject(
	ctx context.Context,
	task coretask.DownloadObjectTask,
	opts ...grpc.DialOption) (
	[]byte, error) {
	conn, err := s.Connection(ctx, s.downloaderEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	req := &gfspserver.GfSpDownloadObjectRequest{
		DownLoadTask: task.(*gfsptask.GfSpDownloadObjectTask),
	}
	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpDownloadObject(ctx, req)
	if err != nil {
		return nil, ErrRpcUnknown
	}
	return resp.GetData(), resp.GetErr()
	return nil, nil
}

func (s *GfSpClient) GetChallengeInfo(
	ctx context.Context,
	task coretask.ChallengePieceTask,
	opts ...grpc.DialOption) (
	[]byte, [][]byte, []byte, error) {
	conn, err := s.Connection(ctx, s.downloaderEndpoint, opts...)
	if err != nil {
		return nil, nil, nil, err
	}
	defer conn.Close()
	req := &gfspserver.GfSpGetChallengeInfoRequest{
		ChallengePieceTask: task.(*gfsptask.GfSpChallengePieceTask),
	}
	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpGetChallengeInfo(ctx, req)
	if err != nil {
		return nil, nil, nil, ErrRpcUnknown
	}
	return resp.GetIntegrityHash(), resp.GetChecksums(), resp.GetData(), resp.GetErr()
}
