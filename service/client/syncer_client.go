package client

import (
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/inscription-storage-provider/util/log"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
)

var _ io.Closer = &SyncerClient{}

// SyncerAPI provides an interface to enable mocking the
// SyncerClient's API operation. This makes unit test to test your code easier.
//
//go:generate mockgen -source=./syncer_client.go -destination=./mock/syncer_mock.go -package=mock
type SyncerAPI interface {
	UploadECPiece(ctx context.Context, opts ...grpc.CallOption) (service.SyncerService_UploadECPieceClient, error)
	Close() error
}

type SyncerClient struct {
	address string
	syncer  service.SyncerServiceClient
	conn    *grpc.ClientConn
}

func NewSyncerClient(address string) (*SyncerClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), ClientRPCTimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke stoneHub service grpc.DialContext failed", "error", err)
		return nil, err
	}
	client := &SyncerClient{
		address: address,
		conn:    conn,
		syncer:  service.NewSyncerServiceClient(conn),
	}
	return client, nil
}

// UploadECPiece return SyncerService_UploadECPieceClient, need to be closed by caller
func (client *SyncerClient) UploadECPiece(ctx context.Context, opts ...grpc.CallOption) (
	service.SyncerService_UploadECPieceClient, error) {
	return client.syncer.UploadECPiece(ctx, opts...)
}

func (client *SyncerClient) Close() error {
	return client.conn.Close()
}
