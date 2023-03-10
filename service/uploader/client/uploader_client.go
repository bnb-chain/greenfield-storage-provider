package client

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	types "github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
)

// UploaderClient is an uploader gRPC service client wrapper
type UploaderClient struct {
	uploader types.UploaderServiceClient
	conn     *grpc.ClientConn
}

// NewUploaderClient return an UploaderClient instance
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

// Close the uploader gPRC client connection
func (client *UploaderClient) Close() error {
	return client.conn.Close()
}

// QueryUploadingObject query an uploading object info with object id
func (client *UploaderClient) QueryUploadingObject(ctx context.Context, objectID sdkmath.Uint, opts ...grpc.CallOption) (
	*servicetypes.SegmentInfo, error) {
	resp, err := client.uploader.QueryUploadingObject(ctx,
		&types.QueryUploadingObjectRequest{ObjectId: objectID}, opts...)
	if err != nil {
		return nil, err
	}
	return resp.GetSegmentInfo(), nil
}

// UploadObject return grpc stream client, and be used to upload object payload.
func (client *UploaderClient) UploadObject(ctx context.Context, opts ...grpc.CallOption) (types.UploaderService_UploadObjectClient, error) {
	return client.uploader.UploadObject(ctx, opts...)
}
