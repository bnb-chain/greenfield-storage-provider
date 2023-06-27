package gfspapp

import (
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	ErrSingTaskDangling = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 991001, "OoooH... request lost")
)

var _ gfspserver.GfSpSignServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpSign(ctx context.Context, req *gfspserver.GfSpSignRequest) (*gfspserver.GfSpSignResponse, error) {
	if req.GetRequest() == nil {
		log.Error("failed to sign msg due to pointer dangling")
		return &gfspserver.GfSpSignResponse{Err: ErrSingTaskDangling}, nil
	}
	var (
		signature []byte
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
	case *gfspserver.GfSpSignRequest_RejectObjectInfo:
		err = g.signer.RejectUnSealObject(ctx, t.RejectObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to reject unseal object", "error", err)
		}
	case *gfspserver.GfSpSignRequest_DiscontinueBucketInfo:
		err = g.signer.DiscontinueBucket(ctx, t.DiscontinueBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to discontinue bucket", "error", err)
		}
	case *gfspserver.GfSpSignRequest_SignSecondaryBls:
		signature, err = g.signer.SignSecondarySealBls(ctx, t.SignSecondaryBls.ObjectId,
			t.SignSecondaryBls.GlobalVirtualGroupId, t.SignSecondaryBls.Checksums)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign secondary bls", "error", err)
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
	case *gfspserver.GfSpSignRequest_CreateGlobalVirtualGroup:
		err = g.signer.CreateGlobalVirtualGroup(ctx, &virtualgrouptypes.MsgCreateGlobalVirtualGroup{
			PrimarySpAddress: g.operatorAddress,
			FamilyId:         t.CreateGlobalVirtualGroup.GetVirtualGroupFamilyId(),
			SecondarySpIds:   t.CreateGlobalVirtualGroup.GetSecondarySpIds(),
			Deposit:          *t.CreateGlobalVirtualGroup.GetDeposit(),
		})
		if err != nil {
			log.CtxErrorw(ctx, "failed to create global virtual group", "error", err)
		}
	}
	return &gfspserver.GfSpSignResponse{
		Err:       gfsperrors.MakeGfSpError(err),
		Signature: signature}, nil
}
