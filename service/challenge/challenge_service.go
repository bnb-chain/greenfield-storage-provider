package challenge

import (
	"context"
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypesv1pb "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// ChallengePiece implement challenge service server interface and handle the grpc request.
func (challenge *Challenge) ChallengePiece(ctx context.Context, req *stypesv1pb.ChallengeServiceChallengePieceRequest) (
	resp *stypesv1pb.ChallengeServiceChallengePieceResponse, err error) {
	ctx = log.Context(ctx, req)
	resp = &stypesv1pb.ChallengeServiceChallengePieceResponse{
		TraceId:  req.TraceId,
		ObjectId: req.ObjectId,
	}
	defer func() {
		if err != nil {
			resp.ErrMessage.ErrCode = stypesv1pb.ErrCode_ERR_CODE_ERROR
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
	var integrityMeta *metadb.IntegrityMeta
	integrityMeta, err = challenge.metaDB.GetIntegrityMeta(req.ObjectId)
	if err != nil {
		return
	}
	if req.GetChallengePrimaryPiece() && !integrityMeta.IsPrimary {
		err = errors.New("storage provider type mismatch")
		return
	}
	if req.GetRedundancyType() != integrityMeta.RedundancyType {
		err = errors.New("redundancy type mismatch")
		return
	}
	if req.GetRedundancyType() == ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED &&
		req.GetEcIdx() != integrityMeta.PieceIdx {
		err = errors.New("ec idx mismatch")
		return
	}
	var pieceKey string
	if req.GetRedundancyType() == ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED {
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
