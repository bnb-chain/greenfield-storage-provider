package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// UploaderClient is a grpc client wrapper.
type UploaderClient struct {
	address  string
	uploader stypes.UploaderServiceClient
	conn     *grpc.ClientConn
}

// NewUploaderClient return a UploaderClient.
func NewUploaderClient(address string) (*UploaderClient, error) {
	conn, err := grpc.DialContext(context.Background(), address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke uploader service grpc.DialContext failed", "error", err)
		return nil, err
	}
	client := &UploaderClient{
		address:  address,
		conn:     conn,
		uploader: stypes.NewUploaderServiceClient(conn),
	}
	return client, nil
}

// UploadPayload return grpc stream client, and be used to upload payload.
func (client *UploaderClient) UploadPayload(ctx context.Context, opts ...grpc.CallOption) (stypes.UploaderService_UploadPayloadClient, error) {
	return client.uploader.UploadPayload(ctx, opts...)
}

func (client *UploaderClient) Close() error {
	return client.conn.Close()
}
