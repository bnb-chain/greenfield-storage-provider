package client

import (
	"context"
	"errors"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

type SignerClient struct {
	address string
	signer  stypes.SignerServiceClient
	conn    *grpc.ClientConn
}

func NewSignerClient(address string) (*SignerClient, error) {
	conn, err := grpc.DialContext(context.Background(), address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke stoneHub service dail failed", "error", err)
		return nil, err
	}
	client := &SignerClient{
		address: address,
		conn:    conn,
		signer:  stypes.NewSignerServiceClient(conn),
	}
	return client, nil
}

func (client *SignerClient) Close() error {
	return client.conn.Close()
}

func (client *SignerClient) SignBucketApproval(ctx context.Context, msg *types.MsgCreateBucket, opts ...grpc.CallOption) ([]byte, error) {
	resp, err := client.signer.SignBucketApproval(ctx, &stypes.SignBucketApprovalRequest{CreateBucketMsg: msg}, opts...)
	if err != nil {
		log.CtxErrorw(ctx, "sign bucket approval failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "sign bucket approval failed", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp.Signature, nil
}

func (client *SignerClient) VerifyBucketApproval(ctx context.Context, msg *types.MsgCreateBucket, opts ...grpc.CallOption) (bool, error) {
	resp, err := client.signer.VerifyBucketApproval(ctx, &stypes.VerifyBucketApprovalRequest{CreateBucketMsg: msg}, opts...)
	if err != nil {
		log.CtxErrorw(ctx, "verify bucket approval failed", "error", err)
		return false, err
	}
	return resp.GetResult(), nil
}

func (client *SignerClient) SignObjectApproval(ctx context.Context, msg *types.MsgCreateObject, opts ...grpc.CallOption) ([]byte, error) {
	resp, err := client.signer.SignObjectApproval(ctx, &stypes.SignObjectApprovalRequest{CreateObjectMsg: msg}, opts...)
	if err != nil {
		log.CtxErrorw(ctx, "sign object approval failed", "error", err)
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "sign object approval failed", "error", resp.GetErrMessage().GetErrMsg())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp.Signature, nil
}

func (client *SignerClient) VerifyObjectApproval(ctx context.Context, msg *types.MsgCreateObject, opts ...grpc.CallOption) (bool, error) {
	resp, err := client.signer.VerifyObjectApproval(ctx, &stypes.VerifyObjectApprovalRequest{CreateObjectMsg: msg}, opts...)
	if err != nil {
		log.CtxErrorw(ctx, "verify bucket approval failed", "error", err)
		return false, err
	}
	return resp.GetResult(), nil
}

func (client *SignerClient) SignIntegrityHash(ctx context.Context, checksum [][]byte, opts ...grpc.CallOption) ([]byte, []byte, error) {
	resp, err := client.signer.SignIntegrityHash(ctx, &stypes.SignIntegrityHashRequest{Data: checksum}, opts...)
	if err != nil {
		log.CtxErrorw(ctx, "sign integrity hash failed", "error", err)
		return []byte{}, []byte{}, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "sign integrity hash  failed", "error", resp.GetErrMessage().GetErrMsg())
		return []byte{}, []byte{}, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp.GetIntegrityHash(), resp.GetSignature(), nil
}

func (client *SignerClient) SealObjectOnChain(ctx context.Context, object *ptypes.ObjectInfo, opts ...grpc.CallOption) ([]byte, error) {
	resp, err := client.signer.SealObjectOnChain(ctx, &stypes.SealObjectOnChainRequest{ObjectInfo: object}, opts...)
	if err != nil {
		log.CtxErrorw(ctx, "failed to seal object on chain", "error", err, "object_info", object)
		return []byte{}, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "failed to seal object on chain", "error", resp.GetErrMessage().GetErrMsg(), "object_info", object)
		return []byte{}, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp.GetTxHash(), nil
}
