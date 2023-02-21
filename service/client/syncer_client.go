package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var _ SyncerAPI = &SyncerClient{}

// SyncerAPI provides an interface to enable mocking the
// SyncerClient's API operation. This makes unit test to test your code easier.
//
//go:generate mockgen -source=./syncer_client.go -destination=./mock/syncer_mock.go -package=mock
type SyncerAPI interface {
	SyncPiece(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error)
	Close() error
}

type SyncerClient struct {
	address string
	syncer  stypes.SyncerServiceClient
	conn    *grpc.ClientConn
}

func NewSyncerClient(address string) (*SyncerClient, error) {
	conn, err := grpc.DialContext(context.Background(), address, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(model.MaxCallMsgSize), grpc.MaxCallSendMsgSize(model.MaxCallMsgSize)))
	if err != nil {
		log.Errorw("invoke syncer service failed", "error", err)
		return nil, err
	}
	client := &SyncerClient{
		address: address,
		conn:    conn,
		syncer:  stypes.NewSyncerServiceClient(conn),
	}
	return client, nil
}

// UploadECPiece return SyncerService_UploadECPieceClient, need to be closed by caller
func (client *SyncerClient) SyncPiece(ctx context.Context, opts ...grpc.CallOption) (
	stypes.SyncerService_SyncPieceClient, error) {
	return client.syncer.SyncPiece(ctx, opts...)
}

func (client *SyncerClient) Close() error {
	return client.conn.Close()
}
