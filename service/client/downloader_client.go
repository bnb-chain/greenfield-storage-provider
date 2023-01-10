package client

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type DownloaderClient struct {
	address    string
	downloader service.DownloaderServiceClient
	conn       *grpc.ClientConn
}

func NewDownloaderClient(address string) (*DownloaderClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), ClientRPCTimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke downloader service dail failed", "error", err)
		return nil, err
	}
	client := &DownloaderClient{
		address:    address,
		conn:       conn,
		downloader: service.NewDownloaderServiceClient(conn),
	}
	return client, nil
}

func (client *DownloaderClient) Close() error {
	return client.conn.Close()
}

func (client *DownloaderClient) DownloaderObject(ctx context.Context, req *service.DownloaderServiceDownloaderObjectRequest, opts ...grpc.CallOption) (data []byte, err error) {
	ctx = log.Context(context.Background(), req)
	var (
		stream service.DownloaderService_DownloaderObjectClient
		resp   *service.DownloaderServiceDownloaderObjectResponse
	)
	defer func() {
		if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() == service.ErrCode_ERR_CODE_ERROR {
			err = errors.New(resp.GetErrMessage().GetErrMsg())
		}
		log.CtxErrorw(ctx, "downloader object completed", "error", err)
	}()
	stream, err = client.downloader.DownloaderObject(ctx, req, opts...)
	if err != nil {
		return data, err
	}
	for {
		resp, err = stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			return data, err
		}
		if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() == service.ErrCode_ERR_CODE_ERROR {
			err = errors.New(resp.GetErrMessage().GetErrMsg())
			return
		}
		data = append(data, resp.GetData()...)
	}
	return
}

func (client *DownloaderClient) DownloaderSegment(ctx context.Context, in *service.DownloaderServiceDownloaderSegmentRequest, opts ...grpc.CallOption) (*service.DownloaderServiceDownloaderSegmentResponse, error) {
	resp, err := client.downloader.DownloaderSegment(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "downloader segment failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "downloader segment response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}
