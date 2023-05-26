package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"google.golang.org/grpc"
)

func (s *GfSpClient) VerifyAuthorize(ctx context.Context,
	auth coremodule.AuthOpType, account, bucket, object string) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authorizerEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authorizer", "error", connErr)
		return false, ErrRpcUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpAuthorizeRequest{
		AuthType:    int32(auth),
		UserAccount: account,
		BucketName:  bucket,
		ObjectName:  object,
	}
	resp, err := gfspserver.NewGfSpAuthorizationServiceClient(conn).GfSpVerifyAuthorize(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to verify authorize", "error", err)
		return false, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return false, resp.GetErr()
	}
	return resp.GetAllowed(), nil
}

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (s *GfSpClient) GetAuthNonce(ctx context.Context, in *gfspserver.GetAuthNonceRequest, opts ...grpc.CallOption) (*gfspserver.GetAuthNonceResponse, error) {
	conn, connErr := s.Connection(ctx, s.authorizerEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authorizer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	resp, err := gfspserver.NewGfSpAuthorizationServiceClient(conn).GetAuthNonce(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get auth nonce rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// UpdateUserPublicKey updates the user public key once the Dapp or client generates the EDDSA key pairs.
func (s *GfSpClient) UpdateUserPublicKey(ctx context.Context, in *gfspserver.UpdateUserPublicKeyRequest, opts ...grpc.CallOption) (*gfspserver.UpdateUserPublicKeyResponse, error) {
	conn, connErr := s.Connection(ctx, s.authorizerEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authorizer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	resp, err := gfspserver.NewGfSpAuthorizationServiceClient(conn).UpdateUserPublicKey(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update user public key rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
func (s *GfSpClient) VerifyOffChainSignature(ctx context.Context, in *gfspserver.VerifyOffChainSignatureRequest, opts ...grpc.CallOption) (*gfspserver.VerifyOffChainSignatureResponse, error) {
	conn, connErr := s.Connection(ctx, s.authorizerEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authorizer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	resp, err := gfspserver.NewGfSpAuthorizationServiceClient(conn).VerifyOffChainSignature(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify off-chain signature rpc", "error", err)
		return nil, err
	}
	return resp, nil
}
