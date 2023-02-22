package client

import (
	"context"
	"errors"

	gtypes "github.com/bnb-chain/greenfield/x/storage/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

type P2PServiceRpcClient struct {
	address string
	pclient stypes.P2PServiceClient
	conn    *grpc.ClientConn
}

func NewP2PServiceRpcClient(address string) (*P2PServiceRpcClient, error) {
	conn, err := grpc.DialContext(context.Background(), address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke stoneHub service dail failed", "error", err)
		return nil, err
	}
	client := &P2PServiceRpcClient{
		address: address,
		conn:    conn,
		pclient: stypes.NewP2PServiceClient(conn),
	}
	return client, nil
}

func (client *P2PServiceRpcClient) Close() error {
	return client.conn.Close()
}

func (client *P2PServiceRpcClient) AskSecondaryApproval(
	ctx context.Context,
	msg *gtypes.MsgCreateObject,
	min, max int32) (
	[]*gtypes.MsgCreateObject, error) {
	resp, err := client.pclient.AskSecondaryApproval(ctx, &stypes.P2PServiceAskSecondaryApprovalRequest{
		CreateObjectMsg:  msg,
		MinApprovalCount: min,
		MaxApprovalCount: max,
	})
	if err != nil {
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "fail to ask secondary approval", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp.GetCreateObjectMsg(), nil
}
