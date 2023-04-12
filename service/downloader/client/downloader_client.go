package client

import (
	"context"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

// DownloaderClient is a downloader gRPC service client wrapper
type DownloaderClient struct {
	address    string
	conn       *grpc.ClientConn
	downloader types.DownloaderServiceClient
}

const downloaderRPCServiceName = "service.downloader.types.DownloaderService"

// NewDownloaderClient returns a DownloaderClient instance
func NewDownloaderClient(address string) (*DownloaderClient, error) {
	options := utilgrpc.GetDefaultClientOptions()
	retryOption, err := utilgrpc.GetDefaultGRPCRetryPolicy(downloaderRPCServiceName)
	if err != nil {
		log.Errorw("failed to get downloader client retry option", "error", err)
		return nil, err
	}
	options = append(options, retryOption)
	if metrics.GetMetrics().Enabled() {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
	conn, err := grpc.DialContext(context.Background(), address, options...)
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

// GetObject downloads the payload of the object
func (client *DownloaderClient) GetObject(ctx context.Context, req *types.GetObjectRequest,
	opts ...grpc.CallOption) (types.DownloaderService_GetObjectClient, error) {
	return client.downloader.GetObject(ctx, req, opts...)
}

// GetBucketReadQuota gets the quota info of the specified month
func (client *DownloaderClient) GetBucketReadQuota(ctx context.Context, bucketInfo *storagetypes.BucketInfo, yearMonth string, opts ...grpc.CallOption) (*types.GetBucketReadQuotaResponse, error) {
	resp, err := client.downloader.GetBucketReadQuota(ctx,
		&types.GetBucketReadQuotaRequest{
			BucketInfo: bucketInfo,
			YearMonth:  yearMonth,
		}, opts...)
	if err != nil {
		log.Errorw("failed to get bucket read quota", "error", err)
		return nil, err
	}
	return resp, nil
}

// ListBucketReadRecord get read record list of the specified time range
func (client *DownloaderClient) ListBucketReadRecord(ctx context.Context, bucketInfo *storagetypes.BucketInfo, startTimestampUs, endTimestampUs, maxRecordNum int64, opts ...grpc.CallOption) (*types.ListBucketReadRecordResponse, error) {
	resp, err := client.downloader.ListBucketReadRecord(ctx,
		&types.ListBucketReadRecordRequest{
			BucketInfo:       bucketInfo,
			StartTimestampUs: startTimestampUs,
			EndTimestampUs:   endTimestampUs,
			MaxRecordNum:     maxRecordNum,
		}, opts...)
	if err != nil {
		log.Errorw("failed to list bucket read records", "error", err)
		return nil, err
	}
	return resp, nil
}

// GetSpByAddress get sp by sp address
func (client *DownloaderClient) GetSpByAddress(ctx context.Context, req *types.GetSpByAddressRequest, opts ...grpc.CallOption) (*types.GetSpByAddressResponse, error) {
	resp, err := client.downloader.GetSpByAddress(ctx, req, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get sp by address rpc", "error", err)
		return nil, err
	}
	return resp, nil
}
