package gfspapp

import (
	"context"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrAuthenticatorTaskDangling = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 990101, "OoooH... request lost")
)

var _ gfspserver.GfSpAuthenticationServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpVerifyAuthentication(ctx context.Context, req *gfspserver.GfSpAuthenticationRequest) (
	*gfspserver.GfSpAuthenticationResponse, error) {
	if req == nil {
		log.Error("failed to verify authentication due to pointer dangling")
		return &gfspserver.GfSpAuthenticationResponse{
			Err: ErrAuthenticatorTaskDangling,
		}, nil
	}
	ctx = log.WithValue(ctx, log.CtxKeyBucketName, req.GetBucketName())
	ctx = log.WithValue(ctx, log.CtxKeyObjectName, req.GetObjectName())
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
	log.CtxDebugw(ctx, "succeed to authenticate", "user", req.GetUserAccount(), "auth_type", req.GetAuthType(),
		"allow", allow, "error", err)
	return &gfspserver.GfSpAuthenticationResponse{
		Err:     gfsperrors.MakeGfSpError(err),
		Allowed: allow,
	}, nil
}

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (g *GfSpBaseApp) GetAuthNonce(ctx context.Context, req *gfspserver.GetAuthNonceRequest) (*gfspserver.GetAuthNonceResponse, error) {
	if req == nil {
		log.Error("failed to get auth nonce due to pointer dangling")
		return &gfspserver.GetAuthNonceResponse{
			Err: ErrAuthenticatorTaskDangling,
		}, nil
	}
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
	if req == nil {
		log.Error("failed to update user publicKey due to pointer dangling")
		return &gfspserver.UpdateUserPublicKeyResponse{
			Err: ErrAuthenticatorTaskDangling,
		}, nil
	}
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
	if req == nil {
		log.Error("failed to verify gnfd1EddsaSignature due to pointer dangling")
		return &gfspserver.VerifyGNFD1EddsaSignatureResponse{
			Err: ErrAuthenticatorTaskDangling,
		}, nil
	}
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

func (g *GfSpBaseApp) GetAuthKeyV2(ctx context.Context, req *gfspserver.GetAuthKeyV2Request) (*gfspserver.GetAuthKeyV2Response, error) {
	if req == nil {
		log.Error("failed to GetAuthKeyV2 to pointer dangling")
		return &gfspserver.GetAuthKeyV2Response{
			Err: ErrAuthenticatorTaskDangling,
		}, nil
	}
	log.CtxDebugw(ctx, "begin to GetAuthKeyV2", "user", req.GetAccountId(), "domain", req.GetDomain(), "public_key", req.GetUserPublicKey())
	resp, err := g.authenticator.GetAuthKeyV2(ctx, req.AccountId, req.Domain, req.UserPublicKey)
	log.CtxDebugw(ctx, "finish to GetAuthKeyV2", "user", req.GetAccountId(), "domain", req.GetDomain(), "public_key", req.GetUserPublicKey(), "error", err)
	if err != nil {
		return &gfspserver.GetAuthKeyV2Response{
			Err: gfsperrors.MakeGfSpError(err),
		}, nil
	}
	if resp == nil {
		return &gfspserver.GetAuthKeyV2Response{
			Err: gfsperrors.MakeGfSpError(err),
		}, nil
	}

	return &gfspserver.GetAuthKeyV2Response{
		Err:        gfsperrors.MakeGfSpError(err),
		PublicKey:  resp.PublicKey,
		ExpiryDate: resp.ExpiryDate.UnixMilli(),
	}, nil
}

func (g *GfSpBaseApp) UpdateUserPublicKeyV2(ctx context.Context, req *gfspserver.UpdateUserPublicKeyV2Request) (*gfspserver.UpdateUserPublicKeyV2Response, error) {
	if req == nil {
		log.Error("failed to UpdateUserPublicKeyV2 due to pointer dangling")
		return &gfspserver.UpdateUserPublicKeyV2Response{
			Err: ErrAuthenticatorTaskDangling,
		}, nil
	}
	log.CtxDebugw(ctx, "begin to UpdateUserPublicKeyV2", "user", req.GetAccountId(), "domain", req.GetDomain(), "public_key", req.UserPublicKey)
	resp, err := g.authenticator.UpdateUserPublicKeyV2(ctx, req.AccountId, req.Domain, req.UserPublicKey, req.ExpiryDate)
	log.CtxDebugw(ctx, "finish to UpdateUserPublicKeyV2", "user", req.GetAccountId(), "domain", req.GetDomain(), "public_key", req.GetUserPublicKey(), "error", err)

	return &gfspserver.UpdateUserPublicKeyV2Response{
		Err:    gfsperrors.MakeGfSpError(err),
		Result: resp,
	}, nil
}

func (g *GfSpBaseApp) VerifyGNFD2EddsaSignature(ctx context.Context, req *gfspserver.VerifyGNFD2EddsaSignatureRequest) (*gfspserver.VerifyGNFD2EddsaSignatureResponse, error) {
	if req == nil {
		log.Error("failed to VerifyGNFD2EddsaSignature due to pointer dangling")
		return &gfspserver.VerifyGNFD2EddsaSignatureResponse{
			Err: ErrAuthenticatorTaskDangling,
		}, nil
	}
	log.CtxDebugw(ctx, "begin to verify off-chain signature", "user", req.GetAccountId(), "domain", req.GetDomain(), "public_key", req.UserPublicKey,
		"off_chain_sig", req.OffChainSig, "real_msg_to_sign", req.RealMsgToSign)
	resp, err := g.authenticator.VerifyGNFD2EddsaSignature(ctx, req.AccountId, req.Domain, req.UserPublicKey, req.OffChainSig, req.RealMsgToSign)
	log.CtxDebugw(ctx, "finish to verify off-chain signature", "user", req.GetAccountId(), "domain", req.GetDomain(), "public_key", req.UserPublicKey,
		"error", err)
	return &gfspserver.VerifyGNFD2EddsaSignatureResponse{
		Err:    gfsperrors.MakeGfSpError(err),
		Result: resp,
	}, nil
}
