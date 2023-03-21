package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
)

// UploaderClient is an uploader gRPC service client wrapper
type UploaderClient struct {
	uploader types.UploaderServiceClient
	conn     *grpc.ClientConn
}

// NewUploaderClient return an UploaderClient instance
func NewUploaderClient(address string) (*UploaderClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithChainUnaryInterceptor(openmetrics.UnaryClientInterceptor(metrics.DefaultGRPCClientMetrics)),
		grpc.WithChainStreamInterceptor(openmetrics.StreamClientInterceptor(metrics.DefaultGRPCClientMetrics)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(model.MaxCallMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(model.MaxCallMsgSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
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
