package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
)

// MetadataClient is an metadata gRPC service client wrapper
type MetadataClient struct {
	address  string
	metadata metatypes.MetadataServiceClient
	conn     *grpc.ClientConn
}

// NewMetadataClient return an MetadataClient instance
func NewMetadataClient(address string) (*MetadataClient, error) {
	conn, err := grpc.DialContext(context.Background(), address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("failed to dial metadata service", "error", err)
		return nil, err
	}
	client := &MetadataClient{
		address:  address,
		conn:     conn,
		metadata: metatypes.NewMetadataServiceClient(conn),
	}
	return client, nil
}

// Close the metadata gPRC client connection
func (client *MetadataClient) Close() error {
	return client.conn.Close()
}

// GetUserBuckets get buckets info by a user address
func (client *MetadataClient) GetUserBuckets(ctx context.Context, in *metatypes.GetUserBucketsRequest, opts ...grpc.CallOption) (*metatypes.GetUserBucketsResponse, error) {
	resp, err := client.metadata.GetUserBuckets(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get user buckets rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// ListObjectsByBucketName list objects info by a bucket name
func (client *MetadataClient) ListObjectsByBucketName(ctx context.Context, in *metatypes.ListObjectsByBucketNameRequest, opts ...grpc.CallOption) (*metatypes.ListObjectsByBucketNameResponse, error) {
	resp, err := client.metadata.ListObjectsByBucketName(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send list objects by bucket name rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// GetBucketByBucketName get bucket info by a bucket name
func (client *MetadataClient) GetBucketByBucketName(ctx context.Context, in *metatypes.GetBucketByBucketNameRequest, opts ...grpc.CallOption) (*metatypes.GetBucketByBucketNameResponse, error) {
	resp, err := client.metadata.GetBucketByBucketName(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get bucket rpc by bucket name", "error", err)
		return nil, err
	}
	return resp, nil
}

// GetBucketByBucketID get bucket info by a bucket id
func (client *MetadataClient) GetBucketByBucketID(ctx context.Context, in *metatypes.GetBucketByBucketIDRequest, opts ...grpc.CallOption) (*metatypes.GetBucketByBucketIDResponse, error) {
	resp, err := client.metadata.GetBucketByBucketID(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get bucket by bucket id rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block range
func (client *MetadataClient) ListDeletedObjectsByBlockNumberRange(ctx context.Context, in *metatypes.ListDeletedObjectsByBlockNumberRangeRequest, opts ...grpc.CallOption) (*metatypes.ListDeletedObjectsByBlockNumberRangeResponse, error) {
	resp, err := client.metadata.ListDeletedObjectsByBlockNumberRange(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send list deleted objects by block number rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// GetUserBucketsCount get buckets count by a user address
func (client *MetadataClient) GetUserBucketsCount(ctx context.Context, in *metatypes.GetUserBucketsCountRequest, opts ...grpc.CallOption) (*metatypes.GetUserBucketsCountResponse, error) {
	resp, err := client.metadata.GetUserBucketsCount(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get user buckets count rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// GetObjectByObjectNameAndBucketName get object info by an object name and a bucket name
func (client *MetadataClient) GetObjectByObjectNameAndBucketName(ctx context.Context, in *metatypes.GetObjectByObjectNameAndBucketNameRequest, opts ...grpc.CallOption) (*metatypes.GetObjectByObjectNameAndBucketNameResponse, error) {
	resp, err := client.metadata.GetObjectByObjectNameAndBucketName(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get object rpc by object name", "error", err)
		return nil, err
	}
	return resp, nil
}

// GetPaymentByBucketName get bucket payment info by a bucket name
func (client *MetadataClient) GetPaymentByBucketName(ctx context.Context, in *metatypes.GetPaymentByBucketNameRequest, opts ...grpc.CallOption) (*metatypes.GetPaymentByBucketNameResponse, error) {
	resp, err := client.metadata.GetPaymentByBucketName(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get payment by bucket name rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// GetPaymentByBucketID get bucket payment info by a bucket id
func (client *MetadataClient) GetPaymentByBucketID(ctx context.Context, in *metatypes.GetPaymentByBucketIDRequest, opts ...grpc.CallOption) (*metatypes.GetPaymentByBucketIDResponse, error) {
	resp, err := client.metadata.GetPaymentByBucketID(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get payment by bucket id rpc", "error", err)
		return nil, err
	}
	return resp, nil
}
