package gfspapp

import (
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	ErrSingTaskDangling = gfsperrors.Register(BaseCodeSpace, http.StatusInternalServerError, 991001, "OoooH... request lost")
)

var _ gfspserver.GfSpSignServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpSign(
	ctx context.Context,
	req *gfspserver.GfSpSignRequest) (
	*gfspserver.GfSpSignResponse, error) {
	if req.GetRequest() == nil {
		log.Error("failed to sign msg, msg pointer dangling")
		return &gfspserver.GfSpSignResponse{Err: ErrSingTaskDangling}, nil
	}
	var (
		signature []byte
		integrity []byte
		err       error
	)
	switch t := req.GetRequest().(type) {
	case *gfspserver.GfSpSignRequest_CreateBucketInfo:
		signature, err = g.signer.SignCreateBucketApproval(ctx, t.CreateBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign create bucket approval", "error", err)
		}
	case *gfspserver.GfSpSignRequest_CreateObjectInfo:
		signature, err = g.signer.SignCreateObjectApproval(ctx, t.CreateObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign create object approval", "error", err)
		}
	case *gfspserver.GfSpSignRequest_SealObjectInfo:
		err = g.signer.SealObject(ctx, t.SealObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to seal object", "error", err)
		}
	case *gfspserver.GfSpSignRequest_DiscontinueBucketInfo:
		err = g.signer.DiscontinueBucket(ctx, t.DiscontinueBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to discontinue bucket", "error", err)
		}
	case *gfspserver.GfSpSignRequest_SignIntegrity:
		signature, integrity, err = g.signer.SignIntegrityHash(ctx, t.SignIntegrity.ObjectId, t.SignIntegrity.Checksums)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign integrity hash", "error", err)
		}
	case *gfspserver.GfSpSignRequest_PingMsg:
		signature, err = g.signer.SignP2PPingMsg(ctx, t.PingMsg)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign p2p ping msg", "error", err)
		}
	case *gfspserver.GfSpSignRequest_PongMsg:
		signature, err = g.signer.SignP2PPongMsg(ctx, t.PongMsg)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign p2p pong msg", "error", err)
		}
	case *gfspserver.GfSpSignRequest_GfspReceivePieceTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspReceivePieceTask.Key().String())
		signature, err = g.signer.SignReceivePieceTask(ctx, t.GfspReceivePieceTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign receive piece task", "error", err)
		}
	case *gfspserver.GfSpSignRequest_GfspReplicatePieceApprovalTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspReplicatePieceApprovalTask.Key().String())
		signature, err = g.signer.SignReplicatePieceApproval(ctx, t.GfspReplicatePieceApprovalTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign replicate piece task", "error", err)
		}
	}
	return &gfspserver.GfSpSignResponse{
		Err:           gfsperrors.MakeGfSpError(err),
		Signature:     signature,
		IntegrityHash: integrity}, nil
}
