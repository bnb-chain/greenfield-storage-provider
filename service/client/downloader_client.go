package client

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

type DownloaderClient struct {
	address    string
	downloader stypes.DownloaderServiceClient
	conn       *grpc.ClientConn
}

func NewDownloaderClient(address string) (*DownloaderClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(20*1024*1024)))
	if err != nil {
		log.Errorw("invoke downloader service dail failed", "error", err)
		return nil, err
	}
	client := &DownloaderClient{
		address:    address,
		conn:       conn,
		downloader: stypes.NewDownloaderServiceClient(conn),
	}
	return client, nil
}

func (client *DownloaderClient) Close() error {
	return client.conn.Close()
}

func (client *DownloaderClient) DownloaderObject(ctx context.Context, req *stypes.DownloaderServiceDownloaderObjectRequest,
	opts ...grpc.CallOption) (stypes.DownloaderService_DownloaderObjectClient, error) {
	ctx = log.Context(context.Background(), req)
	return client.downloader.DownloaderObject(ctx, req, opts...)
}

func (client *DownloaderClient) DownloaderSegment(ctx context.Context, in *stypes.DownloaderServiceDownloaderSegmentRequest,
	opts ...grpc.CallOption) (*stypes.DownloaderServiceDownloaderSegmentResponse, error) {
	resp, err := client.downloader.DownloaderSegment(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "downloader segment failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "downloader segment response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}
