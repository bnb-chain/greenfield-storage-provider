package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func (s *GfSpClient) AskSecondaryReplicatePieceApproval(
	ctx context.Context,
	task coretask.ApprovalReplicatePieceTask,
	low, high int,
	timeout int64) (
	[]*gfsptask.GfSpReplicatePieceApprovalTask, error) {
	conn, err := s.P2PConn(ctx)
	if err != nil {
		return nil, err
	}
	req := &gfspserver.GfSpAskSecondaryReplicatePieceApprovalRequest{
		ReplicatePieceApprovalTask: task.(*gfsptask.GfSpReplicatePieceApprovalTask),
		Min:                        int32(low),
		Max:                        int32(high),
		Timeout:                    timeout,
	}
	resp, err := gfspserver.NewGfSpP2PServiceClient(conn).GfSpAskSecondaryReplicatePieceApproval(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetApprovedTasks(), nil
}

func (s *GfSpClient) QueryP2PBootstrap(ctx context.Context) ([]string, error) {
	conn, err := s.P2PConn(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := gfspserver.NewGfSpP2PServiceClient(conn).GfSpQueryP2PBootstrap(ctx, &gfspserver.GfSpQueryP2PNodeRequest{})
	if err != nil {
		return nil, err
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetNodes(), nil
}
