package client

import (
	"context"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	"github.com/bnb-chain/greenfield-storage-provider/service/p2p/types"
)

// P2PClient is a p2p server gRPC service client wrapper
type P2PClient struct {
	address string
	conn    *grpc.ClientConn
	p2p     types.P2PServiceClient
}

// NewP2PClient return a P2PClient instance
func NewP2PClient(address string) (*P2PClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(model.MaxCallMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(model.MaxCallMsgSize)))
	if err != nil {
		log.Errorw("failed to dial p2p server", "error", err)
		return nil, err
	}
	client := &P2PClient{
		address: address,
		conn:    conn,
		p2p:     types.NewP2PServiceClient(conn),
	}
	return client, nil
}

// Close the p2p server gPRC connection
func (p *P2PClient) Close() error {
	return p.conn.Close()
}

// GetApproval asks the approval to other SP.
func (p *P2PClient) GetApproval(ctx context.Context, object *storagetypes.ObjectInfo, expected int64, timeout int64, opts ...grpc.CallOption) (
	map[string]*p2ptypes.GetApprovalResponse, map[string]*p2ptypes.GetApprovalResponse, error) {
	req := &types.GetApprovalRequest{
		Approval:       &p2ptypes.GetApprovalRequest{ObjectInfo: object},
		ExpectedAccept: expected,
		Timeout:        timeout,
	}
	resp, err := p.p2p.GetApproval(ctx, req, opts...)
	if err != nil {
		return nil, nil, err
	}
	return resp.GetAccept(), resp.GetRefuse(), nil
}
