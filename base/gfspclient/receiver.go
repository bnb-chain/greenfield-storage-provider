package gfspclient

import (
	"context"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func (s *GfSpClient) ReplicatePiece(
	ctx context.Context,
	task coretask.ReceivePieceTask,
	data []byte,
	opts ...grpc.DialOption) error {
	conn, err := s.Connection(ctx, s.receiverEndpoint, opts...)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect receiver", "error", err)
		return err
	}
	defer conn.Close()
	req := &gfspserver.GfSpReplicatePieceRequest{
		ReceivePieceTask: task.(*gfsptask.GfSpReceivePieceTask),
		PieceData:        data,
	}
	resp, err := gfspserver.NewGfSpReceiveServiceClient(conn).GfSpReplicatePiece(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to replicate piece", "error", err)
		return ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return resp.GetErr()
	}
	return nil
}

func (s *GfSpClient) DoneReplicatePiece(
	ctx context.Context,
	task coretask.ReceivePieceTask,
	opts ...grpc.DialOption) (
	[]byte, []byte, error) {
	conn, err := s.Connection(ctx, s.receiverEndpoint, opts...)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect receiver", "error", err)
		return nil, nil, err
	}
	defer conn.Close()
	req := &gfspserver.GfSpDoneReplicatePieceRequest{
		ReceivePieceTask: task.(*gfsptask.GfSpReceivePieceTask),
	}
	resp, err := gfspserver.NewGfSpReceiveServiceClient(conn).GfSpDoneReplicatePiece(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to done replicate piece", "error", err)
		return nil, nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, nil, resp.GetErr()
	}
	return resp.GetIntegrityHash(), resp.GetSignature(), nil
}
