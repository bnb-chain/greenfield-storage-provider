package client

import (
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var ClientRPCTimeout = time.Second * 5

var _ io.Closer = &StoneHubClient{}

// StoneHubAPI provides an interface to enable mocking the
// StoneHubClient's API operation. This makes unit test to test your code easier.
//
//go:generate mockgen -source=./stone_hub_client.go -destination=./mock/stone_hub_mock.go -package=mock
type StoneHubAPI interface {
	CreateObject(ctx context.Context, in *service.StoneHubServiceCreateObjectRequest, opts ...grpc.CallOption) (*service.StoneHubServiceCreateObjectResponse, error)
	SetObjectCreateInfo(ctx context.Context, in *service.StoneHubServiceSetObjectCreateInfoRequest, opts ...grpc.CallOption) (*service.StoneHubServiceSetSetObjectCreateInfoResponse, error)
	BeginUploadPayload(ctx context.Context, in *service.StoneHubServiceBeginUploadPayloadRequest, opts ...grpc.CallOption) (*service.StoneHubServiceBeginUploadPayloadResponse, error)
	DonePrimaryPieceJob(ctx context.Context, in *service.StoneHubServiceDonePrimaryPieceJobRequest, opts ...grpc.CallOption) (*service.StoneHubServiceDonePrimaryPieceJobResponse, error)
	AllocStoneJob(ctx context.Context, opts ...grpc.CallOption) (*service.StoneHubServiceAllocStoneJobResponse, error)
	DoneSecondaryPieceJob(ctx context.Context, in *service.StoneHubServiceDoneSecondaryPieceJobRequest, opts ...grpc.CallOption) (*service.StoneHubServiceDoneSecondaryPieceJobResponse, error)
	Close() error
}

type StoneHubClient struct {
	address  string
	stoneHub service.StoneHubServiceClient
	conn     *grpc.ClientConn
}

func NewStoneHubClient(address string) (*StoneHubClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), ClientRPCTimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke stoneHub service grpc.DialContext failed", "error", err)
		return nil, err
	}
	client := &StoneHubClient{
		address:  address,
		conn:     conn,
		stoneHub: service.NewStoneHubServiceClient(conn),
	}
	return client, nil
}

func (client *StoneHubClient) CreateObject(ctx context.Context, in *service.StoneHubServiceCreateObjectRequest,
	opts ...grpc.CallOption) (*service.StoneHubServiceCreateObjectResponse, error) {
	resp, err := client.stoneHub.CreateObject(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "create object failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "create object response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) SetObjectCreateInfo(ctx context.Context, in *service.StoneHubServiceSetObjectCreateInfoRequest,
	opts ...grpc.CallOption) (*service.StoneHubServiceSetSetObjectCreateInfoResponse, error) {
	resp, err := client.stoneHub.SetObjectCreateInfo(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "set object height and object id failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "set object height and object id response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) BeginUploadPayload(ctx context.Context, in *service.StoneHubServiceBeginUploadPayloadRequest,
	opts ...grpc.CallOption) (*service.StoneHubServiceBeginUploadPayloadResponse, error) {
	resp, err := client.stoneHub.BeginUploadPayload(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "begin upload stone failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "begin upload stone response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) DonePrimaryPieceJob(ctx context.Context, in *service.StoneHubServiceDonePrimaryPieceJobRequest,
	opts ...grpc.CallOption) (*service.StoneHubServiceDonePrimaryPieceJobResponse, error) {
	resp, err := client.stoneHub.DonePrimaryPieceJob(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "done primary piece job failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "done primary piece job response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) AllocStoneJob(ctx context.Context, opts ...grpc.CallOption) (
	*service.StoneHubServiceAllocStoneJobResponse, error) {
	req := &service.StoneHubServiceAllocStoneJobRequest{TraceId: util.GenerateRequestID()}
	resp, err := client.stoneHub.AllocStoneJob(ctx, req, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "alloc stone job failed", "error", err)
		return nil, err
	}
	if resp.PieceJob == nil {
		log.CtxErrorw(ctx, "alloc stone job empty.")
		return nil, merrors.ErrEmptyJob
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "alloc stone job failed", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) DoneSecondaryPieceJob(ctx context.Context, in *service.StoneHubServiceDoneSecondaryPieceJobRequest,
	opts ...grpc.CallOption) (*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	resp, err := client.stoneHub.DoneSecondaryPieceJob(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "done secondary piece job failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "done secondary piece job response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) Close() error {
	return client.conn.Close()
}
