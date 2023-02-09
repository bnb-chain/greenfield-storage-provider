package challenge

import (
	"context"
	"errors"

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
		integrityMeta  *spdb.IntegrityMeta
		queryCondition *spdb.IntegrityMeta
	)

	ctx = log.Context(ctx, req)
	resp = &stypes.ChallengeServiceChallengePieceResponse{
		TraceId:  req.TraceId,
		ObjectId: req.ObjectId,
	}
	defer func() {
		if err != nil {
			resp.ErrMessage.ErrCode = stypes.ErrCode_ERR_CODE_ERROR
			resp.ErrMessage.ErrMsg = err.Error()
			log.CtxErrorw(ctx, "challenge failed", "error", err)
		} else {
			log.CtxInfow(ctx, "challenge success")
		}
	}()
	if req.GetStorageProviderId() != challenge.config.StorageProvider {
		err = errors.New("storage provider id mismatch")
		return
	}
	queryCondition = &spdb.IntegrityMeta{
		ObjectID:       req.ObjectId,
		IsPrimary:      req.ChallengePrimaryPiece,
		RedundancyType: req.RedundancyType,
		EcIdx:          req.EcIdx,
	}
	integrityMeta, err = challenge.metaDB.GetIntegrityMeta(queryCondition)
	if err != nil {
		return
	}

	var pieceKey string
	if req.GetRedundancyType() == ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED {
		pieceKey = piecestore.EncodeECPieceKey(req.GetObjectId(), req.GetSegmentIdx(), req.GetEcIdx())
	} else {
		pieceKey = piecestore.EncodeSegmentPieceKey(req.GetObjectId(), req.GetSegmentIdx())
	}
	resp.PieceData, err = challenge.pieceStore.GetPiece(ctx, pieceKey, 0, -1)
	if err != nil {
		return
	}
	resp.IntegrityHash = integrityMeta.IntegrityHash
	resp.PieceHash = integrityMeta.PieceHash
	return resp, nil
}
