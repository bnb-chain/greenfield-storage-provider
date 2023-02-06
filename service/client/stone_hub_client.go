package client

import (
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var ClientRPCTimeout = time.Second * 5

var _ io.Closer = &StoneHubClient{}

// StoneHubAPI provides an interface to enable mocking the
// StoneHubClient's API operation. This makes unit test to test your code easier.
//
//go:generate mockgen -source=./stone_hub_client.go -destination=./mock/stone_hub_mock.go -package=mock
type StoneHubAPI interface {
	CreateObject(ctx context.Context, in *stypes.StoneHubServiceCreateObjectRequest, opts ...grpc.CallOption) (*stypes.StoneHubServiceCreateObjectResponse, error)
	SetObjectCreateInfo(ctx context.Context, in *stypes.StoneHubServiceSetObjectCreateInfoRequest, opts ...grpc.CallOption) (*stypes.StoneHubServiceSetObjectCreateInfoResponse, error)
	BeginUploadPayload(ctx context.Context, in *stypes.StoneHubServiceBeginUploadPayloadRequest, opts ...grpc.CallOption) (*stypes.StoneHubServiceBeginUploadPayloadResponse, error)
	BeginUploadPayloadV2(ctx context.Context, in *stypes.StoneHubServiceBeginUploadPayloadV2Request, opts ...grpc.CallOption) (*stypes.StoneHubServiceBeginUploadPayloadV2Response, error)
	DonePrimaryPieceJob(ctx context.Context, in *stypes.StoneHubServiceDonePrimaryPieceJobRequest, opts ...grpc.CallOption) (*stypes.StoneHubServiceDonePrimaryPieceJobResponse, error)
	AllocStoneJob(ctx context.Context, opts ...grpc.CallOption) (*stypes.StoneHubServiceAllocStoneJobResponse, error)
	DoneSecondaryPieceJob(ctx context.Context, in *stypes.StoneHubServiceDoneSecondaryPieceJobRequest, opts ...grpc.CallOption) (*stypes.StoneHubServiceDoneSecondaryPieceJobResponse, error)
	Close() error
}

type StoneHubClient struct {
	address  string
	stoneHub stypes.StoneHubServiceClient
	conn     *grpc.ClientConn
}

func NewStoneHubClient(address string) (*StoneHubClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), ClientRPCTimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke stoneHub service dail failed", "error", err)
		return nil, err
	}
	client := &StoneHubClient{
		address:  address,
		conn:     conn,
		stoneHub: stypes.NewStoneHubServiceClient(conn),
	}
	return client, nil
}

func (client *StoneHubClient) CreateObject(ctx context.Context, in *stypes.StoneHubServiceCreateObjectRequest,
	opts ...grpc.CallOption) (*stypes.StoneHubServiceCreateObjectResponse, error) {
	resp, err := client.stoneHub.CreateObject(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "create object failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "create object response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	log.CtxInfow(ctx, "create object succeed", "request", in)
	return resp, nil
}

func (client *StoneHubClient) SetObjectCreateInfo(ctx context.Context, in *stypes.StoneHubServiceSetObjectCreateInfoRequest,
	opts ...grpc.CallOption) (*stypes.StoneHubServiceSetObjectCreateInfoResponse, error) {
	resp, err := client.stoneHub.SetObjectCreateInfo(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "set object height and object id failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "set object height and object id response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) BeginUploadPayload(ctx context.Context, in *stypes.StoneHubServiceBeginUploadPayloadRequest,
	opts ...grpc.CallOption) (*stypes.StoneHubServiceBeginUploadPayloadResponse, error) {
	resp, err := client.stoneHub.BeginUploadPayload(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "begin upload stone failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "begin upload stone response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) BeginUploadPayloadV2(ctx context.Context, in *stypes.StoneHubServiceBeginUploadPayloadV2Request,
	opts ...grpc.CallOption) (*stypes.StoneHubServiceBeginUploadPayloadV2Response, error) {
	resp, err := client.stoneHub.BeginUploadPayloadV2(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "begin upload stone failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "begin upload stone response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) DonePrimaryPieceJob(ctx context.Context, in *stypes.StoneHubServiceDonePrimaryPieceJobRequest,
	opts ...grpc.CallOption) (*stypes.StoneHubServiceDonePrimaryPieceJobResponse, error) {
	resp, err := client.stoneHub.DonePrimaryPieceJob(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "done primary piece job failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "done primary piece job response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) AllocStoneJob(ctx context.Context, opts ...grpc.CallOption) (
	*stypes.StoneHubServiceAllocStoneJobResponse, error) {
	req := &stypes.StoneHubServiceAllocStoneJobRequest{TraceId: util.GenerateRequestID()}
	resp, err := client.stoneHub.AllocStoneJob(ctx, req, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "alloc stone job failed", "error", err)
		return nil, err
	}
	if resp.PieceJob == nil {
		log.CtxDebugw(ctx, "alloc stone job is empty")
		return nil, merrors.ErrEmptyJob
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "alloc stone job failed", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) DoneSecondaryPieceJob(ctx context.Context, in *stypes.StoneHubServiceDoneSecondaryPieceJobRequest,
	opts ...grpc.CallOption) (*stypes.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	resp, err := client.stoneHub.DoneSecondaryPieceJob(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "done secondary piece job failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "done secondary piece job response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

func (client *StoneHubClient) Close() error {
	return client.conn.Close()
}
