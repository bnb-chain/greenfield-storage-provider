package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	types "github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// UploaderClient is a grpc client wrapper.
type UploaderClient struct {
	uploader types.UploaderServiceClient
	conn     *grpc.ClientConn
}

// NewUploaderClient return an uploader gPRC client instance.
func NewUploaderClient(address string) (*UploaderClient, error) {
	conn, err := grpc.DialContext(context.Background(), address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("fail to invoke uploader service client", "error", err)
		return nil, err
	}
	client := &UploaderClient{
		conn:     conn,
		uploader: types.NewUploaderServiceClient(conn),
	}
	return client, nil
}

// Close the uploader gPRC client instance.
func (client *UploaderClient) Close() error {
	return client.conn.Close()
}

// QueryUploadingObject query an uploading object info with object id.
func (client *UploaderClient) QueryUploadingObject(
	ctx context.Context,
	objectId uint64,
	opts ...grpc.CallOption) (*servicetypes.SegmentInfo, error) {
	resp, err := client.uploader.QueryUploadingObject(ctx,
		&types.QueryUploadingObjectRequest{ObjectId: objectId}, opts...)
	if err != nil {
		return nil, err
	}
	return resp.GetSegmentInfo(), nil
}
