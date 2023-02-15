package signer

import (
	"context"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
)

const (
	APITokenMD = "API-KEY"
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

// IPWhitelistInterceptor returns a new unary server interceptors that performs per-request ip whitelist.
func (signer *SignerServer) IPWhitelistInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ip := util.GetIPFromGRPCContext(ctx)
		if !signer.whitelist.Permitted(ip) {
			return nil, merrors.ErrIPBlocked
		}

		return handler(ctx, req)
	}
}

// AuthInterceptor returns a new unary server interceptors that performs per-request auth.
func (signer *SignerServer) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		apiKey := metautils.ExtractIncoming(ctx).Get(APITokenMD)
		if apiKey != signer.config.APIKey {
			return nil, merrors.ErrAPIKey
		}
		return handler(ctx, req)
	}
}
