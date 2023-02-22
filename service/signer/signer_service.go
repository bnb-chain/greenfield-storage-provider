package signer

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
)

/* signer_service.go implement SignerServiceServer grpc interface.
 *
 * SignBucketApproval, SignObjectApproval implement the signature request for approval.
 * VerifyBucketApproval, VerifyObjectApproval implement the signature verification for approval.
 * SignIntegrityHash implement the SP signature request of the integrity hash and signature.
 * SealObjectOnChain implement the primary SP to submit a SealObject transaction request.
 */

var _ stypes.SignerServiceServer = &SignerServer{}

const (
	APITokenMD = "API-KEY"
)

// SignBucketApproval implements v1.SignerServiceServer
func (signer *SignerServer) SignBucketApproval(ctx context.Context, req *stypes.SignBucketApprovalRequest) (*stypes.SignBucketApprovalResponse, error) {
	msg := req.CreateBucketMsg.GetApprovalBytes()
	sig, err := signer.client.Sign(client.SignApproval, msg)
	if err != nil {
		return &stypes.SignBucketApprovalResponse{
			ErrMessage: merrors.MakeErrMsgResponse(merrors.ErrSignMsg),
		}, nil
	}

	return &stypes.SignBucketApprovalResponse{
		Signature: sig,
	}, nil
}

// SignObjectApproval implements v1.SignerServiceServer
func (signer *SignerServer) SignObjectApproval(ctx context.Context, req *stypes.SignObjectApprovalRequest) (*stypes.SignObjectApprovalResponse, error) {
	msg := req.CreateObjectMsg.GetApprovalBytes()
	sig, err := signer.client.Sign(client.SignApproval, msg)
	if err != nil {
		return &stypes.SignObjectApprovalResponse{
			ErrMessage: merrors.MakeErrMsgResponse(merrors.ErrSignMsg),
		}, nil
	}

	return &stypes.SignObjectApprovalResponse{
		Signature: sig,
	}, nil
}

// VerifyBucketApproval implements v1.SignerServiceServer
func (signer *SignerServer) VerifyBucketApproval(ctx context.Context, req *stypes.VerifyBucketApprovalRequest) (*stypes.VerifyBucketApprovalResponse, error) {
	sig := req.CreateBucketMsg.GetPrimarySpApproval().GetSig()
	msg := req.CreateBucketMsg.GetApprovalBytes()

	return &stypes.VerifyBucketApprovalResponse{
		Result: signer.client.VerifySignature(client.SignApproval,
			msg, sig),
	}, nil
}

// VerifyObjectApproval implements v1.SignerServiceServer
func (signer *SignerServer) VerifyObjectApproval(ctx context.Context, req *stypes.VerifyObjectApprovalRequest) (*stypes.VerifyObjectApprovalResponse, error) {
	sig := req.CreateObjectMsg.GetPrimarySpApproval().GetSig()
	msg := req.CreateObjectMsg.GetApprovalBytes()

	return &stypes.VerifyObjectApprovalResponse{
		Result: signer.client.VerifySignature(client.SignApproval,
			msg, sig),
	}, nil
}

// SignIntegrityHash implements v1.SignerServiceServer
func (signer *SignerServer) SignIntegrityHash(ctx context.Context, req *stypes.SignIntegrityHashRequest) (*stypes.SignIntegrityHashResponse, error) {
	integrityHash := hash.GenerateIntegrityHash(req.Data)

	sig, err := signer.client.Sign(client.SignApproval, integrityHash)
	if err != nil {
		return &stypes.SignIntegrityHashResponse{
			ErrMessage: merrors.MakeErrMsgResponse(merrors.ErrSignMsg),
		}, nil
	}

	return &stypes.SignIntegrityHashResponse{
		Signature:     sig,
		IntegrityHash: integrityHash,
	}, nil
}

// SealObjectOnChain implements v1.SignerServiceServer
func (signer *SignerServer) SealObjectOnChain(ctx context.Context, req *stypes.SealObjectOnChainRequest) (*stypes.SealObjectOnChainResponse, error) {
	txHash, err := signer.client.SealObject(ctx, client.SignSeal, req.ObjectInfo)

	return &stypes.SealObjectOnChainResponse{
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
