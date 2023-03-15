package client

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
)

// DownloaderClient is a downloader gRPC service client wrapper
type DownloaderClient struct {
	address    string
	downloader types.DownloaderServiceClient
	conn       *grpc.ClientConn
}

// NewDownloaderClient return a DownloaderClient instance
func NewDownloaderClient(address string) (*DownloaderClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(model.MaxCallMsgSize)))
	if err != nil {
		log.Errorw("failed to dial downloader", "error", err)
		return nil, err
	}
	client := &DownloaderClient{
		address:    address,
		conn:       conn,
		downloader: types.NewDownloaderServiceClient(conn),
	}
	return client, nil
}

// Close the download gPRC connection
func (client *DownloaderClient) Close() error {
	return client.conn.Close()
}

// GetObject download the payload of the object
func (client *DownloaderClient) GetObject(ctx context.Context, req *types.GetObjectRequest,
	opts ...grpc.CallOption) (types.DownloaderService_GetObjectClient, error) {
	return client.downloader.GetObject(ctx, req, opts...)
}
