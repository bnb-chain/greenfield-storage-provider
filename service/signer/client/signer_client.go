package client

import (
	"context"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer/types"
)

type SignerClient struct {
	address string
	signer  types.SignerServiceClient
	conn    *grpc.ClientConn
}

func NewSignerClient(address string) (*SignerClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("failed to dail signer", "error", err)
		return nil, err
	}
	client := &SignerClient{
		address: address,
		conn:    conn,
		signer:  types.NewSignerServiceClient(conn),
	}
	return client, nil
}

func (client *SignerClient) Close() error {
	return client.conn.Close()
}

func (client *SignerClient) SignBucketApproval(ctx context.Context,
	msg *storagetypes.MsgCreateBucket, opts ...grpc.CallOption) ([]byte, error) {
	resp, err := client.signer.SignBucketApproval(ctx,
		&types.SignBucketApprovalRequest{CreateBucketMsg: msg}, opts...)
	if err != nil {
		return nil, err
	}
	return resp.Signature, nil
}

func (client *SignerClient) VerifyBucketApproval(ctx context.Context,
	msg *storagetypes.MsgCreateBucket, opts ...grpc.CallOption) (bool, error) {
	resp, err := client.signer.VerifyBucketApproval(ctx,
		&types.VerifyBucketApprovalRequest{CreateBucketMsg: msg}, opts...)
	if err != nil {
		return false, err
	}
	return resp.GetResult(), nil
}

func (client *SignerClient) SignObjectApproval(ctx context.Context,
	msg *storagetypes.MsgCreateObject, opts ...grpc.CallOption) ([]byte, error) {
	resp, err := client.signer.SignObjectApproval(ctx,
		&types.SignObjectApprovalRequest{CreateObjectMsg: msg}, opts...)
	if err != nil {
		return nil, err
	}
	return resp.Signature, nil
}

func (client *SignerClient) VerifyObjectApproval(ctx context.Context,
	msg *storagetypes.MsgCreateObject, opts ...grpc.CallOption) (bool, error) {
	resp, err := client.signer.VerifyObjectApproval(ctx,
		&types.VerifyObjectApprovalRequest{CreateObjectMsg: msg}, opts...)
	if err != nil {
		return false, err
	}
	return resp.GetResult(), nil
}

func (client *SignerClient) SignIntegrityHash(ctx context.Context,
	checksum [][]byte, opts ...grpc.CallOption) ([]byte, []byte, error) {
	resp, err := client.signer.SignIntegrityHash(ctx,
		&types.SignIntegrityHashRequest{Data: checksum}, opts...)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return resp.GetIntegrityHash(), resp.GetSignature(), nil
}

func (client *SignerClient) SealObjectOnChain(ctx context.Context,
	sealObject *storagetypes.MsgSealObject, opts ...grpc.CallOption) ([]byte, error) {
	resp, err := client.signer.SealObjectOnChain(ctx,
		&types.SealObjectOnChainRequest{SealObject: sealObject}, opts...)
	if err != nil {
		return []byte{}, err
	}
	return resp.GetTxHash(), nil
}
