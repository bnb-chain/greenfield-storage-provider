package signer

import (
	"context"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

func (signer *SignerServer) Sign(ctx context.Context, req *stypes.SignRequest) (*stypes.SignResponse, error) {
	sig, err := signer.greenfieldChain.Sign(req.Msg)

	return &stypes.SignResponse{
		Signature:  sig,
		ErrMessage: merrors.MakeErrMsgResponse(err),
	}, nil
}

func (signer *SignerServer) SealObject(ctx context.Context, object *ptypes.ObjectInfo) (*stypes.SealObjectResponse, error) {
	txHash, err := signer.greenfieldChain.SealObject(ctx, object)

	return &stypes.SealObjectResponse{
		TxHash:     txHash,
		ErrMessage: merrors.MakeErrMsgResponse(err),
	}, nil
}
