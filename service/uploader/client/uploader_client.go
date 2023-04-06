package client

import (
	"context"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

// UploaderClient is an uploader gRPC service client wrapper
type UploaderClient struct {
	uploader types.UploaderServiceClient
	conn     *grpc.ClientConn
}

// NewUploaderClient return an UploaderClient instance
func NewUploaderClient(address string) (*UploaderClient, error) {
	options := utilgrpc.GetDefaultClientOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
	conn, err := grpc.DialContext(context.Background(), address, options...)
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

// Close the uploader gRPC client connection
func (client *UploaderClient) Close() error {
	return client.conn.Close()
}

// QueryPuttingObject query a putting object info with object id
func (client *UploaderClient) QueryPuttingObject(ctx context.Context, objectID uint64, opts ...grpc.CallOption) (
	*servicetypes.SegmentInfo, error) {
	resp, err := client.uploader.QueryPuttingObject(ctx,
		&types.QueryPuttingObjectRequest{ObjectId: objectID}, opts...)
	if err != nil {
		return nil, err
	}
	return resp.GetSegmentInfo(), nil
}

// PutObject return grpc stream client, and be used to upload object payload.
func (client *UploaderClient) PutObject(ctx context.Context, opts ...grpc.CallOption) (types.UploaderService_PutObjectClient, error) {
	return client.uploader.PutObject(ctx, opts...)
}
