package client

import (
	"context"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UploaderClient is an uploader gRPC service client wrapper
type UploaderClient struct {
	uploader types.UploaderServiceClient
	conn     *grpc.ClientConn
}

const uploaderRPCServiceName = "service.uploader.types.UploaderService"

// NewUploaderClient return an UploaderClient instance
func NewUploaderClient(address string) (*UploaderClient, error) {
	options := utilgrpc.GetDefaultClientOptions()
	retryOption, err := utilgrpc.GetDefaultGRPCRetryPolicy(uploaderRPCServiceName)
	if err != nil {
		log.Errorw("failed to get uploader client retry option", "error", err)
		return nil, err
	}
	options = append(options, retryOption)
	//if metrics.GetMetrics().Enabled() {
	//	options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	//}
	conn, err := grpc.DialContext(context.Background(), address, options...)
	if err != nil {
		log.Errorw("failed to invoke uploader service client", "error", err)
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

// QueryPuttingObject queries a putting object info with object id
func (client *UploaderClient) QueryPuttingObject(ctx context.Context, objectID uint64, opts ...grpc.CallOption) (
	*servicetypes.PieceInfo, error) {
	resp, err := client.uploader.QueryPuttingObject(ctx,
		&types.QueryPuttingObjectRequest{ObjectId: objectID}, opts...)
	if err != nil {
		return nil, err
	}
	return resp.GetPieceInfo(), nil
}

// PutObject returns grpc stream client, and be used to upload object payload.
func (client *UploaderClient) PutObject(ctx context.Context, opts ...grpc.CallOption) (types.UploaderService_PutObjectClient, error) {
	return client.uploader.PutObject(ctx, opts...)
}

// QueryUploadProgress is used to query upload object progress
func (client *UploaderClient) QueryUploadProgress(ctx context.Context, objectID uint64, opts ...grpc.CallOption) (servicetypes.JobState, error) {
	resp, err := client.uploader.QueryUploadProgress(ctx, &types.QueryUploadProgressRequest{ObjectId: objectID}, opts...)
	if err != nil {
		errStatus, _ := status.FromError(err)
		if codes.NotFound == errStatus.Code() {
			return servicetypes.JobState_JOB_STATE_INIT_UNSPECIFIED, merrors.ErrNoSuchObject
		}
		return servicetypes.JobState_JOB_STATE_INIT_UNSPECIFIED, err
	}
	return resp.GetState(), nil
}
