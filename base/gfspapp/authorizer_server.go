package gfspapp

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var _ gfspserver.GfSpAuthorizationServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpVerifyAuthorize(
	ctx context.Context,
	req *gfspserver.GfSpAuthorizeRequest) (
	*gfspserver.GfSpAuthorizeResponse, error) {
	ctx = log.WithValue(ctx, log.CtxKeyBucketName, req.GetBucketName())
	ctx = log.WithValue(ctx, log.CtxKeyObjectName, req.GetObjectName())
	log.CtxDebugw(ctx, "begin to authorize", "user", req.GetUserAccount(), "auth_type", req.GetAuthType())
	allow, err := g.authorizer.VerifyAuthorize(ctx, coremodule.AuthOpType(req.GetAuthType()),
		req.GetUserAccount(), req.GetBucketName(), req.GetObjectName())
	log.CtxDebugw(ctx, "finish to authorize", "user", req.GetUserAccount(), "auth_type", req.GetAuthType(),
		"allow", allow, "error", err)
	return &gfspserver.GfSpAuthorizeResponse{
		Err:     gfsperrors.MakeGfSpError(err),
		Allowed: allow,
	}, nil
}
