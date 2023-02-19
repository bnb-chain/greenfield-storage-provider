package client

import (
	"context"
	"errors"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ChallengeClient is a grpc client wrapper.
type ChallengeClient struct {
	address   string
	conn      *grpc.ClientConn
	challenge stypes.ChallengeServiceClient
}

// NewChallengeClient return a ChallengeClient.
func NewChallengeClient(address string) (*ChallengeClient, error) {
	conn, err := grpc.DialContext(context.Background(), address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke challenge service grpc.DialContext failed", "error", err)
		return nil, err
	}
	client := &ChallengeClient{
		address:   address,
		conn:      conn,
		challenge: stypes.NewChallengeServiceClient(conn),
	}
	return client, nil
}

func (client *ChallengeClient) ChallengePiece(ctx context.Context, in *stypes.ChallengeServiceChallengePieceRequest,
	opts ...grpc.CallOption) (*stypes.ChallengeServiceChallengePieceResponse, error) {
	resp, err := client.challenge.ChallengePiece(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "challenge piece failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "challenge piece response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	log.CtxInfow(ctx, "challenge piece succeed", "request", in)
	return resp, nil
}

func (client *ChallengeClient) Close() error {
	return client.conn.Close()
}
