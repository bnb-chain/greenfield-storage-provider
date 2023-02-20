package challenge

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// ChallengePiece implement challenge service server interface and handle the grpc request.
func (challenge *Challenge) ChallengePiece(ctx context.Context, req *stypes.ChallengeServiceChallengePieceRequest) (
	resp *stypes.ChallengeServiceChallengePieceResponse, err error) {
	var (
		integrityMeta *spdb.IntegrityMeta
	)

	ctx = log.Context(ctx, req)
	resp = &stypes.ChallengeServiceChallengePieceResponse{}
	defer func(resp *stypes.ChallengeServiceChallengePieceResponse, err error) {
		if err != nil {
			resp.ErrMessage = &stypes.ErrMessage{
				ErrCode: stypes.ErrCode_ERR_CODE_ERROR,
				ErrMsg:  err.Error(),
			}
			log.CtxErrorw(ctx, "challenge failed", "error", err)
		} else {
			log.CtxInfow(ctx, "challenge success")
		}
	}(resp, err)

	if integrityMeta, err = challenge.metaDB.GetIntegrityMeta(req.GetObjectId()); err != nil {
		return
	}

	var pieceKey string
	if integrityMeta.IsPrimary {
		pieceKey = piecestore.EncodeSegmentPieceKey(req.GetObjectId(), req.GetSegmentIdx())
	} else {
		if integrityMeta.RedundancyType == ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED {
			// TODO: check integrityMeta.EcIdx == req.EcIdx
			pieceKey = piecestore.EncodeECPieceKey(req.GetObjectId(), req.GetSegmentIdx(), integrityMeta.EcIdx)
		} else {
			pieceKey = piecestore.EncodeSegmentPieceKey(req.GetObjectId(), req.GetSegmentIdx())
		}
	}
	if resp.PieceData, err = challenge.pieceStore.GetPiece(ctx, pieceKey, 0, -1); err != nil {
		return
	}
	resp.IntegrityHash = integrityMeta.IntegrityHash
	resp.PieceHash = integrityMeta.PieceHash
	return resp, nil
}
