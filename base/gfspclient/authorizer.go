package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func (s *GfSpClient) VerifyAuthorize(ctx context.Context,
	auth coremodule.AuthOpType, account, bucket, object string) (bool, error) {
	conn, err := s.Connection(ctx, s.authorizerEndpoint)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect authorizer", "error", err)
		return false, err
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
