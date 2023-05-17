package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"google.golang.org/grpc"
)

func (s *GfSpClient) ReplicatePiece(
	ctx context.Context,
	task coretask.ReceivePieceTask,
	data []byte,
	opts ...grpc.DialOption) error {
	conn, err := s.Connection(ctx, s.receiverEndpoint, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	req := &gfspserver.GfSpReplicatePieceRequest{
		ReceivePieceTask: task.(*gfsptask.GfSpReceivePieceTask),
		PieceData:        data,
	}
	resp, err := gfspserver.NewGfSpReceiveServiceClient(conn).GfSpReplicatePiece(ctx, req)
	if err != nil {
		return err
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
		return nil, nil, err
	}
	defer conn.Close()
	req := &gfspserver.GfSpDoneReplicatePieceRequest{
		ReceivePieceTask: task.(*gfsptask.GfSpReceivePieceTask),
	}
	resp, err := gfspserver.NewGfSpReceiveServiceClient(conn).GfSpDoneReplicatePiece(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	if resp.GetErr() != nil {
		return nil, nil, resp.GetErr()
	}
	return resp.GetIntegrityHash(), resp.GetSignature(), nil
}
