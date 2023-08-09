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

func (s *GfSpClient) ReplicatePiece(ctx context.Context, task coretask.ReceivePieceTask, data []byte,
	opts ...grpc.DialOption) error {
	conn, connErr := s.Connection(ctx, s.receiverEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect receiver", "error", connErr)
		return ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpReplicatePieceRequest{
		ReceivePieceTask: task.(*gfsptask.GfSpReceivePieceTask),
		PieceData:        data,
	}
	startTime := time.Now()
	resp, err := gfspserver.NewGfSpReceiveServiceClient(conn).GfSpReplicatePiece(ctx, req)
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_client_total_time").Observe(time.Since(startTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "client failed to replicate piece", "error", err)
		return ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return resp.GetErr()
	}
	return nil
}

func (s *GfSpClient) DoneReplicatePiece(ctx context.Context, task coretask.ReceivePieceTask, opts ...grpc.DialOption) (
	[]byte, []byte, error) {
	conn, connErr := s.Connection(ctx, s.receiverEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect receiver", "error", connErr)
		return nil, nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpDoneReplicatePieceRequest{
		ReceivePieceTask: task.(*gfsptask.GfSpReceivePieceTask),
	}
	startTime := time.Now()
	resp, err := gfspserver.NewGfSpReceiveServiceClient(conn).GfSpDoneReplicatePiece(ctx, req)
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_done_client_total_time").Observe(time.Since(startTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "client failed to done replicate piece", "error", err)
		return nil, nil, ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return nil, nil, resp.GetErr()
	}
	return resp.GetIntegrityHash(), resp.GetSignature(), nil
}
