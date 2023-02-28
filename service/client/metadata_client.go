package client

import (
	"context"
	"errors"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MetadataClient struct {
	address  string
	metadata stypes.MetadataServiceClient
	conn     *grpc.ClientConn
}

func NewMetadataClient(address string) (*MetadataClient, error) {
	conn, err := grpc.DialContext(context.Background(), address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke metadata service grpc.DialContext failed", "error", err)
		return nil, err
	}
	client := &MetadataClient{
		address:  address,
		conn:     conn,
		metadata: stypes.NewMetadataServiceClient(conn),
	}
	return client, nil
}

func (client *MetadataClient) GetUserBuckets(ctx context.Context, in *stypes.MetadataServiceGetUserBucketsRequest, opts ...grpc.CallOption) (*stypes.MetadataServiceGetUserBucketsResponse, error) {
	resp, err := client.metadata.GetUserBuckets(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "send get user buckets  rpc failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "get user buckets response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *MetadataClient) ListObjectsByBucketName(ctx context.Context, in *stypes.MetadataServiceListObjectsByBucketNameRequest, opts ...grpc.CallOption) (*stypes.MetadataServiceListObjectsByBucketNameResponse, error) {
	resp, err := client.metadata.ListObjectsByBucketName(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "send list object by bucket name rpc failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "list object by bucket name response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *MetadataClient) Close() error {
	return client.conn.Close()
}
