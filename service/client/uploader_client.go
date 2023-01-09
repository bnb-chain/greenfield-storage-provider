package client

import (
	"context"
	"errors"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UploaderClient is a grpc client wrapper.
type UploaderClient struct {
	address  string
	uploader service.UploaderServiceClient
	conn     *grpc.ClientConn
}

// NewUploaderClient return a UploaderClient.
func NewUploaderClient(address string) (*UploaderClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), ClientRpcTimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke uploader service grpc.DialContext failed", "error", err)
		return nil, err
	}
	client := &UploaderClient{
		address:  address,
		conn:     conn,
		uploader: service.NewUploaderServiceClient(conn),
	}
	return client, nil
}

// CreateObject invoke uploader service CreateObject interface.
func (client *UploaderClient) CreateObject(ctx context.Context, in *service.UploaderServiceCreateObjectRequest, opts ...grpc.CallOption) (*service.UploaderServiceCreateObjectResponse, error) {
	resp, err := client.uploader.CreateObject(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "send create object rpc failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "create object response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

// UploadPayload return grpc stream client, and be used to upload payload.
func (client *UploaderClient) UploadPayload(ctx context.Context, opts ...grpc.CallOption) (service.UploaderService_UploadPayloadClient, error) {
	return client.uploader.UploadPayload(ctx, opts...)
}

func (client *UploaderClient) Close() error {
	return client.conn.Close()
}
