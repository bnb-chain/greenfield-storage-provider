package signer

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-common/go/hash"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

/* signer_service.go implement SignerServiceServer grpc interface.
 *
 * SignBucketApproval, SignObjectApproval implement the signature request for approval.
 * VerifyBucketApproval, VerifyObjectApproval implement the signature verification for approval.
 * SignIntegrityHash implement the SP signature request of the integrity hash and signature.
 * SealObjectOnChain implement the primary SP to submit a SealObject transaction request.
 */

var _ types.SignerServiceServer = &SignerServer{}

const (
	APITokenMD = "API-KEY"
)

// SignBucketApproval implements v1.SignerServiceServer
func (signer *SignerServer) SignBucketApproval(ctx context.Context, req *types.SignBucketApprovalRequest) (*types.SignBucketApprovalResponse, error) {
	msg := req.CreateBucketMsg.GetApprovalBytes()
	sig, err := signer.client.Sign(client.SignApproval, msg)
	if err != nil {
		return nil, err
	}

	return &types.SignBucketApprovalResponse{
		Signature: sig,
	}, nil
}

// SignObjectApproval implements v1.SignerServiceServer
func (signer *SignerServer) SignObjectApproval(ctx context.Context, req *types.SignObjectApprovalRequest) (*types.SignObjectApprovalResponse, error) {
	msg := req.CreateObjectMsg.GetApprovalBytes()
	sig, err := signer.client.Sign(client.SignApproval, msg)
	if err != nil {
		return nil, err
	}

	return &types.SignObjectApprovalResponse{
		Signature: sig,
	}, nil
}

// VerifyBucketApproval implements v1.SignerServiceServer
func (signer *SignerServer) VerifyBucketApproval(ctx context.Context, req *types.VerifyBucketApprovalRequest) (*types.VerifyBucketApprovalResponse, error) {
	sig := req.CreateBucketMsg.GetPrimarySpApproval().GetSig()
	msg := req.CreateBucketMsg.GetApprovalBytes()

	return &types.VerifyBucketApprovalResponse{
		Result: signer.client.VerifySignature(client.SignApproval,
			msg, sig),
	}, nil
}

// VerifyObjectApproval implements v1.SignerServiceServer
func (signer *SignerServer) VerifyObjectApproval(ctx context.Context, req *types.VerifyObjectApprovalRequest) (*types.VerifyObjectApprovalResponse, error) {
	sig := req.CreateObjectMsg.GetPrimarySpApproval().GetSig()
	msg := req.CreateObjectMsg.GetApprovalBytes()

	return &types.VerifyObjectApprovalResponse{
		Result: signer.client.VerifySignature(client.SignApproval,
			msg, sig),
	}, nil
}

// SignIntegrityHash implements v1.SignerServiceServer
func (signer *SignerServer) SignIntegrityHash(ctx context.Context, req *types.SignIntegrityHashRequest) (*types.SignIntegrityHashResponse, error) {
	integrityHash := hash.GenerateIntegrityHash(req.Data)
	opAddr, err := signer.client.GetAddr(client.SignOperator)
	if err != nil {
		return nil, err
	}

	msg := storagetypes.NewSecondarySpSignDoc(opAddr, sdkmath.NewUint(req.ObjectId), integrityHash).GetSignBytes()
	sig, err := signer.client.Sign(client.SignApproval, msg)
	if err != nil {
		return nil, err
	}

	return &types.SignIntegrityHashResponse{
		Signature:     sig,
		IntegrityHash: integrityHash,
	}, nil
}

// SealObjectOnChain implements v1.SignerServiceServer
func (signer *SignerServer) SealObjectOnChain(ctx context.Context, req *types.SealObjectOnChainRequest) (*types.SealObjectOnChainResponse, error) {
	txHash, err := signer.client.SealObject(ctx, client.SignSeal, req.SealObject)
	if err != nil {
		return nil, err
	}

	return &types.SealObjectOnChainResponse{
		TxHash: txHash,
	}, nil

}

// SignPingMsg signs the ping msg for p2p node
func (signer *SignerServer) SignPingMsg(ctx context.Context, req *types.SignPingMsgRequest) (*types.SignPingMsgResponse, error) {
	msg := req.GetPing()
	sig, err := signer.client.Sign(client.SignOperator, msg.GetSignBytes())
	if err != nil {
		return nil, err
	}
	msg.Signature = sig
	return &types.SignPingMsgResponse{
		Ping: msg,
	}, nil
}

// SignPongMsg signs the pong msg for p2p node
func (signer *SignerServer) SignPongMsg(ctx context.Context, req *types.SignPongMsgRequest) (*types.SignPongMsgResponse, error) {
	msg := req.GetPong()
	sig, err := signer.client.Sign(client.SignOperator, msg.GetSignBytes())
	if err != nil {
		return nil, err
	}
	msg.Signature = sig
	return &types.SignPongMsgResponse{
		Pong: msg,
	}, nil
}

// SignReplicateApprovalReqMsg signs the get approval request msg for p2p node
func (signer *SignerServer) SignReplicateApprovalReqMsg(ctx context.Context, req *types.SignReplicateApprovalReqMsgRequest) (*types.SignReplicateApprovalReqMsgResponse, error) {
	msg := req.GetApproval()
	sig, err := signer.client.Sign(client.SignOperator, msg.GetSignBytes())
	if err != nil {
		return nil, err
	}
	msg.Signature = sig
	return &types.SignReplicateApprovalReqMsgResponse{
		Approval: msg,
	}, nil
}

// SignReplicateApprovalRspMsg signs the get approval response msg for p2p node
func (signer *SignerServer) SignReplicateApprovalRspMsg(ctx context.Context, req *types.SignReplicateApprovalRspMsgRequest) (*types.SignReplicateApprovalRspMsgResponse, error) {
	msg := req.GetApproval()
	sig, err := signer.client.Sign(client.SignOperator, msg.GetSignBytes())
	if err != nil {
		return nil, err
	}
	msg.Signature = sig
	return &types.SignReplicateApprovalRspMsgResponse{
		Approval: msg,
	}, nil
}

// IPWhitelistInterceptor returns a new unary server interceptors that performs per-request ip whitelist.
func (signer *SignerServer) IPWhitelistInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ip := util.GetIPFromGRPCContext(ctx)
		if !signer.svcWhitelist.Permitted(ip) {
			return nil, merrors.ErrIPBlocked
		}

		return handler(ctx, req)
	}
}

// AuthInterceptor returns a new unary server interceptors that performs per-request auth.
func (signer *SignerServer) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// TODO: add it in future
		apiKey := metautils.ExtractIncoming(ctx).Get(APITokenMD)
		if apiKey != signer.config.APIKey {
			return nil, merrors.ErrAPIKey
		}
		return handler(ctx, req)
	}
}
