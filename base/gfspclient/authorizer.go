package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"google.golang.org/grpc"
)

func (s *GfSpClient) VerifyAuthorize(ctx context.Context, auth coremodule.AuthOpType, account, bucket, object string) (bool, error) {
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
func (s *GfSpClient) GetAuthNonce(ctx context.Context, account string, domain string, opts ...grpc.CallOption) (currentNonce int32, nextNonce int32, currentPublicKey string, expiryDate int64, err error) {
	conn, connErr := s.Connection(ctx, s.authorizerEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authorizer", "error", connErr)
		return 0, 0, "", 0, ErrRpcUnknown
	}
	defer conn.Close()
	req := &gfspserver.GetAuthNonceRequest{
		AccountId: account,
		Domain:    domain,
	}
	resp, err := gfspserver.NewGfSpAuthorizationServiceClient(conn).GetAuthNonce(ctx, req, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get auth nonce rpc", "error", err)
		return 0, 0, "", 0, err
	}
	if resp.GetErr() != nil {
		return 0, 0, "", 0, resp.GetErr()
	}
	return resp.GetCurrentNonce(), resp.GetNextNonce(), resp.GetCurrentPublicKey(), resp.GetExpiryDate(), nil
}

// UpdateUserPublicKey updates the user public key once the Dapp or client generates the EDDSA key pairs.
func (s *GfSpClient) UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32, userPublicKey string, expiryDate int64, opts ...grpc.CallOption) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authorizerEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authorizer", "error", connErr)
		return false, ErrRpcUnknown
	}
	req := &gfspserver.UpdateUserPublicKeyRequest{
		AccountId:     account,
		Domain:        domain,
		CurrentNonce:  currentNonce,
		Nonce:         nonce,
		UserPublicKey: userPublicKey,
		ExpiryDate:    expiryDate,
	}
	resp, err := gfspserver.NewGfSpAuthorizationServiceClient(conn).UpdateUserPublicKey(ctx, req, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update user public key rpc", "error", err)
		return false, err
	}
	if resp.GetErr() != nil {
		return false, resp.GetErr()
	}
	return resp.Result, nil
}

// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
func (s *GfSpClient) VerifyOffChainSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign string, opts ...grpc.CallOption) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authorizerEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authorizer", "error", connErr)
		return false, ErrRpcUnknown
	}
	req := &gfspserver.VerifyOffChainSignatureRequest{
		AccountId:     account,
		Domain:        domain,
		OffChainSig:   offChainSig,
		RealMsgToSign: realMsgToSign,
	}
	resp, err := gfspserver.NewGfSpAuthorizationServiceClient(conn).VerifyOffChainSignature(ctx, req, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify off-chain signature rpc", "error", err)
		return false, err
	}
	if resp.GetErr() != nil {
		return false, resp.GetErr()
	}
	return resp.Result, nil
}
