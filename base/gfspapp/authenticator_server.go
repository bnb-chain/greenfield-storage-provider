package gfspapp

import (
	"context"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var _ gfspserver.GfSpAuthenticationServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpVerifyAuthentication(ctx context.Context, req *gfspserver.GfSpAuthenticationRequest) (
	*gfspserver.GfSpAuthenticationResponse, error) {
	ctx = log.WithValue(ctx, log.CtxKeyBucketName, req.GetBucketName())
	ctx = log.WithValue(ctx, log.CtxKeyObjectName, req.GetObjectName())
	log.CtxDebugw(ctx, "begin to authenticate", "user", req.GetUserAccount(), "auth_type", req.GetAuthType())
	startTime := time.Now()
	allow, err := g.authenticator.VerifyAuthentication(ctx, coremodule.AuthOpType(req.GetAuthType()),
		req.GetUserAccount(), req.GetBucketName(), req.GetObjectName())
	if err != nil || !allow {
		metrics.ReqCounter.WithLabelValues(AuthFailure).Inc()
		metrics.ReqTime.WithLabelValues(AuthFailure).Observe(time.Since(startTime).Seconds())
	} else {
		metrics.ReqCounter.WithLabelValues(AuthSuccess).Inc()
		metrics.ReqTime.WithLabelValues(AuthSuccess).Observe(time.Since(startTime).Seconds())
	}
	log.CtxDebugw(ctx, "finish to authenticate", "user", req.GetUserAccount(), "auth_type", req.GetAuthType(),
		"allow", allow, "error", err)
	return &gfspserver.GfSpAuthenticationResponse{
		Err:     gfsperrors.MakeGfSpError(err),
		Allowed: allow,
	}, nil
}

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (g *GfSpBaseApp) GetAuthNonce(ctx context.Context, req *gfspserver.GetAuthNonceRequest) (*gfspserver.GetAuthNonceResponse, error) {
	log.CtxDebugw(ctx, "begin to get auth nonce", "user", req.GetAccountId(), "domain", req.GetDomain())
	resp, err := g.authenticator.GetAuthNonce(ctx, req.AccountId, req.Domain)
	log.CtxDebugw(ctx, "finish to get auth nonce", "user", req.GetAccountId(), "domain", req.GetDomain(), "error", err)
	if err != nil {
		return &gfspserver.GetAuthNonceResponse{
			Err: gfsperrors.MakeGfSpError(err),
		}, nil
	}

	return &gfspserver.GetAuthNonceResponse{
		Err:              gfsperrors.MakeGfSpError(err),
		CurrentNonce:     resp.CurrentNonce,
		NextNonce:        resp.NextNonce,
		CurrentPublicKey: resp.CurrentPublicKey,
		ExpiryDate:       resp.ExpiryDate.UnixMilli(),
	}, nil
}

// UpdateUserPublicKey updates the user public key once the Dapp or client generates the EDDSA key pairs.
func (g *GfSpBaseApp) UpdateUserPublicKey(ctx context.Context, req *gfspserver.UpdateUserPublicKeyRequest) (*gfspserver.UpdateUserPublicKeyResponse, error) {
	log.CtxDebugw(ctx, "begin to update user public key", "user", req.GetAccountId(), "domain", req.GetDomain(),
		"public_key", req.UserPublicKey)
	resp, err := g.authenticator.UpdateUserPublicKey(ctx, req.AccountId, req.Domain, req.CurrentNonce, req.Nonce, req.UserPublicKey, req.ExpiryDate)
	log.CtxDebugw(ctx, "finish to update user public key", "user", req.GetAccountId(), "domain", req.GetDomain(),
		"error", err)
	return &gfspserver.UpdateUserPublicKeyResponse{
		Err:    gfsperrors.MakeGfSpError(err),
		Result: resp,
	}, nil
}

// VerifyGNFD1EddsaSignature verifies the signature signed by user's EDDSA private key.
func (g *GfSpBaseApp) VerifyGNFD1EddsaSignature(ctx context.Context, req *gfspserver.VerifyGNFD1EddsaSignatureRequest) (*gfspserver.VerifyGNFD1EddsaSignatureResponse, error) {
	log.CtxDebugw(ctx, "begin to verify off-chain signature", "user", req.GetAccountId(), "domain", req.GetDomain(),
		"off_chain_sig", req.OffChainSig, "real_msg_to_sign", req.RealMsgToSign)
	resp, err := g.authenticator.VerifyGNFD1EddsaSignature(ctx, req.AccountId, req.Domain, req.OffChainSig, req.RealMsgToSign)
	log.CtxDebugw(ctx, "finish to verify off-chain signature", "user", req.GetAccountId(), "domain", req.GetDomain(),
		"error", err)
	return &gfspserver.VerifyGNFD1EddsaSignatureResponse{
		Err:    gfsperrors.MakeGfSpError(err),
		Result: resp,
	}, nil
}
