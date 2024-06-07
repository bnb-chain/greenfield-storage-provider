package gfspclient

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfspserver"
	coremodule "github.com/zkMeLabs/mechain-storage-provider/core/module"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/metrics"
)

func (s *GfSpClient) VerifyAuthentication(ctx context.Context, auth coremodule.AuthOpType, account, bucket, object string,
	opts ...grpc.DialOption,
) (bool, error) {
	startTime := time.Now()
	defer func() {
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_client_total_time").Observe(time.Since(startTime).Seconds())
	}()
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	metrics.PerfAuthTimeHistogram.WithLabelValues("auth_client_create_conn_time").Observe(time.Since(startTime).Seconds())
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknownWithDetail("client failed to connect authenticator, error:", connErr)
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
		return false, ErrRPCUnknownWithDetail("client failed to verify authentication, error: ", err)
	}
	if resp.GetErr() != nil {
		return false, resp.GetErr()
	}
	return resp.GetAllowed(), nil
}

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (s *GfSpClient) GetAuthNonce(ctx context.Context, account string, domain string, opts ...grpc.DialOption) (
	currentNonce int32, nextNonce int32, currentPublicKey string, expiryDate int64, err error,
) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return 0, 0, "", 0, ErrRPCUnknownWithDetail("client failed to connect authenticator, error: ", connErr)
	}
	defer conn.Close()
	req := &gfspserver.GetAuthNonceRequest{
		AccountId: account,
		Domain:    domain,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).GetAuthNonce(ctx, req)
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
func (s *GfSpClient) UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32,
	userPublicKey string, expiryDate int64, opts ...grpc.DialOption,
) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknownWithDetail("client failed to connect authenticator, error: ", connErr)
	}
	defer conn.Close()
	req := &gfspserver.UpdateUserPublicKeyRequest{
		AccountId:     account,
		Domain:        domain,
		CurrentNonce:  currentNonce,
		Nonce:         nonce,
		UserPublicKey: userPublicKey,
		ExpiryDate:    expiryDate,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).UpdateUserPublicKey(ctx, req)
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

// VerifyGNFD1EddsaSignature verifies the signature signed by user's EDDSA private key.
func (s *GfSpClient) VerifyGNFD1EddsaSignature(ctx context.Context, account string, domain string, offChainSig string,
	realMsgToSign []byte, opts ...grpc.DialOption,
) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknownWithDetail("client failed to connect authenticator, error: ", connErr)
	}
	defer conn.Close()
	req := &gfspserver.VerifyGNFD1EddsaSignatureRequest{
		AccountId:     account,
		Domain:        domain,
		OffChainSig:   offChainSig,
		RealMsgToSign: realMsgToSign,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).VerifyGNFD1EddsaSignature(ctx, req)
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

func (s *GfSpClient) GetAuthKeyV2(ctx context.Context, account string, domain string, userPublicKey string, opts ...grpc.DialOption) (publicKey string, expiryDate int64, err error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return "", 0, ErrRPCUnknownWithDetail("client failed to connect authenticator, error: ", connErr)
	}
	defer conn.Close()
	req := &gfspserver.GetAuthKeyV2Request{
		AccountId:     account,
		Domain:        domain,
		UserPublicKey: userPublicKey,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).GetAuthKeyV2(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get GetAuthKeyV2 rpc", "error", err)
		return "", 0, err
	}
	if resp.GetErr() != nil {
		return "", 0, resp.GetErr()
	}
	return resp.GetPublicKey(), resp.GetExpiryDate(), nil
}

func (s *GfSpClient) UpdateUserPublicKeyV2(ctx context.Context, account string, domain string, userPublicKey string, expiryDate int64, opts ...grpc.DialOption) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknownWithDetail("client failed to connect authenticator, error: ", connErr)
	}
	defer conn.Close()
	req := &gfspserver.UpdateUserPublicKeyV2Request{
		AccountId:     account,
		Domain:        domain,
		UserPublicKey: userPublicKey,
		ExpiryDate:    expiryDate,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).UpdateUserPublicKeyV2(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to UpdateUserPublicKeyV2 rpc", "error", err)
		return false, err
	}
	if resp.GetErr() != nil {
		return false, resp.GetErr()
	}
	return resp.Result, nil
}

func (s *GfSpClient) VerifyGNFD2EddsaSignature(ctx context.Context, account string, domain string, userPublicKey string, offChainSig string, realMsgToSign []byte, opts ...grpc.DialOption) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknownWithDetail("client failed to connect authenticator, error: ", connErr)
	}
	defer conn.Close()
	req := &gfspserver.VerifyGNFD2EddsaSignatureRequest{
		AccountId:     account,
		Domain:        domain,
		UserPublicKey: userPublicKey,
		OffChainSig:   offChainSig,
		RealMsgToSign: realMsgToSign,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).VerifyGNFD2EddsaSignature(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to VerifyGNFD2EddsaSignature rpc", "error", err)
		return false, err
	}
	if resp.GetErr() != nil {
		return false, resp.GetErr()
	}
	return resp.Result, nil
}

func (s *GfSpClient) ListAuthKeysV2(ctx context.Context, account string, domain string, opts ...grpc.DialOption) ([]string, error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return nil, ErrRPCUnknownWithDetail("client failed to connect authenticator, error: ", connErr)
	}
	defer conn.Close()
	req := &gfspserver.ListAuthKeysV2Request{
		AccountId: account,
		Domain:    domain,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).ListAuthKeysV2(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get ListAuthKeysV2 rpc", "error", err)
		return nil, err
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetPublicKeys(), nil
}

func (s *GfSpClient) DeleteAuthKeysV2(ctx context.Context, account string, domain string, userPublicKeys []string, opts ...grpc.DialOption) (bool, error) {
	conn, connErr := s.Connection(ctx, s.authenticatorEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect authenticator", "error", connErr)
		return false, ErrRPCUnknownWithDetail("client failed to connect authenticator, error: ", connErr)
	}
	defer conn.Close()
	req := &gfspserver.DeleteAuthKeysV2Request{
		AccountId:  account,
		Domain:     domain,
		PublicKeys: userPublicKeys,
	}
	resp, err := gfspserver.NewGfSpAuthenticationServiceClient(conn).DeleteAuthKeysV2(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get DeleteAuthKeysV2 rpc", "error", err)
		return false, err
	}
	if resp.GetErr() != nil {
		return false, resp.GetErr()
	}
	return resp.Result, nil
}
