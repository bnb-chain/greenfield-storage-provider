package gfspclient

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

func (s *GfSpClient) GetObject(ctx context.Context, downloadObjectTask coretask.DownloadObjectTask, opts ...grpc.DialOption) (
	[]byte, error) {
	conn, connErr := s.Connection(ctx, s.downloaderEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect downloader", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpDownloadObjectRequest{
		DownloadObjectTask: downloadObjectTask.(*gfsptask.GfSpDownloadObjectTask),
	}
	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpDownloadObject(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to download object", "error", err)
		return nil, ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetData(), nil
}

func (s *GfSpClient) GetPiece(ctx context.Context, downloadPieceTask coretask.DownloadPieceTask, opts ...grpc.DialOption) (
	[]byte, error) {
	conn, connErr := s.Connection(ctx, s.downloaderEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect downloader", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpDownloadPieceRequest{
		DownloadPieceTask: downloadPieceTask.(*gfsptask.GfSpDownloadPieceTask),
	}
	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpDownloadPiece(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to download piece", "error", err)
		return nil, ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetData(), nil
}

func (s *GfSpClient) GetChallengeInfo(ctx context.Context, challengePieceTask coretask.ChallengePieceTask, opts ...grpc.DialOption) (
	[]byte, [][]byte, []byte, error) {
	conn, connErr := s.Connection(ctx, s.downloaderEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect downloader", "error", connErr)
		return nil, nil, nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpGetChallengeInfoRequest{
		ChallengePieceTask: challengePieceTask.(*gfsptask.GfSpChallengePieceTask),
	}
	startTime := time.Now()
	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpGetChallengeInfo(ctx, req)
	metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_client_total_time").Observe(time.Since(startTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get challenge piece info", "error", err)
		return nil, nil, nil, ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return nil, nil, nil, resp.GetErr()
	}
	return resp.GetIntegrityHash(), resp.GetChecksums(), resp.GetData(), nil
}
