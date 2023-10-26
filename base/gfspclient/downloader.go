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
		return nil, ErrRPCUnknownWithDetail("client failed to connect downloader, error: " + connErr.Error())
	}
	defer conn.Close()
	req := &gfspserver.GfSpDownloadObjectRequest{
		DownloadObjectTask: downloadObjectTask.(*gfsptask.GfSpDownloadObjectTask),
	}
	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpDownloadObject(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to download object", "error", err)
		return nil, ErrRPCUnknownWithDetail("client failed to download object, error: " + err.Error())
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
		return nil, ErrRPCUnknownWithDetail("client failed to connect downloader, error: " + connErr.Error())
	}
	defer conn.Close()
	req := &gfspserver.GfSpDownloadPieceRequest{
		DownloadPieceTask: downloadPieceTask.(*gfsptask.GfSpDownloadPieceTask),
	}
	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpDownloadPiece(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to download piece", "error", err)
		return nil, ErrRPCUnknownWithDetail("client failed to download piece, error: " + err.Error())
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetData(), nil
}

func (s *GfSpClient) RecoupQuota(ctx context.Context, bucketID, extraQuota uint64, yearMonth string, opts ...grpc.DialOption) error {
	conn, connErr := s.Connection(ctx, s.downloaderEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect downloader", "error", connErr)
		return ErrRPCUnknownWithDetail("client failed to connect downloader, error: " + connErr.Error())
	}
	defer conn.Close()
	req := &gfspserver.GfSpReimburseQuotaRequest{
		BucketId:   bucketID,
		ExtraQuota: extraQuota,
		YearMonth:  yearMonth,
	}

	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpReimburseQuota(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to recoup the extra quota", "error", err)
		return ErrRPCUnknownWithDetail("client failed to recoup extra quota, error: " + err.Error())
	}
	return resp.GetErr()
}

func (s *GfSpClient) DeductQuotaForBucketMigrate(ctx context.Context, bucketID, deductQuota uint64, yearMonth string, opts ...grpc.DialOption) error {
	conn, connErr := s.Connection(ctx, s.downloaderEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect downloader", "error", connErr)
		return ErrRPCUnknownWithDetail("client failed to connect downloader, error: " + connErr.Error())
	}
	defer conn.Close()
	req := &gfspserver.GfSpDeductQuotaForBucketMigrateRequest{
		BucketId:    bucketID,
		DeductQuota: deductQuota,
		YearMonth:   yearMonth,
	}

	resp, err := gfspserver.NewGfSpDownloadServiceClient(conn).GfSpDeductQuotaForBucketMigrate(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to deduct the quota for bucket migrate", "request", req, "error", err)
		return ErrRPCUnknownWithDetail("client failed to deduct the quota for bucket migrate, error: " + err.Error())
	}
	return resp.GetErr()
}

func (s *GfSpClient) GetChallengeInfo(ctx context.Context, challengePieceTask coretask.ChallengePieceTask, opts ...grpc.DialOption) (
	[]byte, [][]byte, []byte, error) {
	conn, connErr := s.Connection(ctx, s.downloaderEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect downloader", "error", connErr)
		return nil, nil, nil, ErrRPCUnknownWithDetail("client failed to connect downloader, error: " + connErr.Error())
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
		return nil, nil, nil, ErrRPCUnknownWithDetail("client failed to get challenge piece info, error: " + err.Error())
	}
	if resp.GetErr() != nil {
		return nil, nil, nil, resp.GetErr()
	}
	return resp.GetIntegrityHash(), resp.GetChecksums(), resp.GetData(), nil
}
