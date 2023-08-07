package gfspclient

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

// AuthenticatorAPI for mock use
type AuthenticatorAPI interface {
	VerifyAuthentication(ctx context.Context, auth coremodule.AuthOpType, account, bucket, object string) (bool, error)
	GetAuthNonce(ctx context.Context, account string, domain string, opts ...grpc.CallOption) (currentNonce int32, nextNonce int32, currentPublicKey string, expiryDate int64, err error)
	UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32, userPublicKey string, expiryDate int64, opts ...grpc.CallOption) (bool, error)
	VerifyOffChainSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign string, opts ...grpc.CallOption) (bool, error)
	VerifyGNFD1EddsaSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign []byte, opts ...grpc.CallOption) (bool, error)
}

func (s *GfSpClient) VerifyAuthentication(ctx context.Context, auth coremodule.AuthOpType, account, bucket, object string) (bool, error) {
	startTime := time.Now()
	defer metrics.PerfAuthTimeHistogram.WithLabelValues("auth_client_total_time").Observe(time.Since(startTime).Seconds())
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint)
	metrics.PerfAuthTimeHistogram.WithLabelValues("auth_client_create_conn_time").Observe(time.Since(startTime).Seconds())
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpAuthenticationRequest{
		AuthType:    int32(auth),
		UserAccount: account,
		BucketName:  bucket,
		ObjectName:  object,
	}
	startRequestTime := time.Now()
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).GfSpVerifyAuthentication(ctx, req)
	metrics.PerfAuthTimeHistogram.WithLabelValues("auth_client_network_time").Observe(time.Since(startRequestTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "client failed to verify authentication", "error", err)
		return false, ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return false, resp.GetErr()
	}
	return resp.GetAllowed(), nil
}

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (s *GfSpClient) GetAuthNonce(ctx context.Context, account string, domain string, opts ...grpc.CallOption) (currentNonce int32, nextNonce int32, currentPublicKey string, expiryDate int64, err error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return 0, 0, "", 0, ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GetAuthNonceRequest{
		AccountId: account,
		Domain:    domain,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).GetAuthNonce(ctx, req, opts...)
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
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknown
	}
	req := &gfspserver.UpdateUserPublicKeyRequest{
		AccountId:     account,
		Domain:        domain,
		CurrentNonce:  currentNonce,
		Nonce:         nonce,
		UserPublicKey: userPublicKey,
		ExpiryDate:    expiryDate,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).UpdateUserPublicKey(ctx, req, opts...)
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

// Deprecated: This method will be deleted in future versions, once most SP and clients migrates to GNFD1 Auth.
// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
func (s *GfSpClient) VerifyOffChainSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign string, opts ...grpc.CallOption) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknown
	}
	req := &gfspserver.VerifyOffChainSignatureRequest{
		AccountId:     account,
		Domain:        domain,
		OffChainSig:   offChainSig,
		RealMsgToSign: realMsgToSign,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).VerifyOffChainSignature(ctx, req, opts...)
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

// VerifyGNFD1EddsaSignature verifies the signature signed by user's EDDSA private key.
func (s *GfSpClient) VerifyGNFD1EddsaSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign []byte, opts ...grpc.CallOption) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknown
	}
	req := &gfspserver.VerifyGNFD1EddsaSignatureRequest{
		AccountId:     account,
		Domain:        domain,
		OffChainSig:   offChainSig,
		RealMsgToSign: realMsgToSign,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).VerifyGNFD1EddsaSignature(ctx, req, opts...)
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
