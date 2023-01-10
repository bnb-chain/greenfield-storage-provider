package challenge

import (
	"context"
	"errors"

	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// ChallengePiece implement challenge service server interface and handle the grpc request.
func (challenge *Challenge) ChallengePiece(ctx context.Context, req *service.ChallengeServiceChallengePieceRequest) (resp *service.ChallengeServiceChallengePieceResponse, err error) {
	ctx = log.Context(ctx, req)
	resp = &service.ChallengeServiceChallengePieceResponse{
		TraceId:  req.TraceId,
		ObjectId: req.ObjectId,
	}
	defer func() {
		if err != nil {
			resp.ErrMessage.ErrCode = service.ErrCode_ERR_CODE_ERROR
			resp.ErrMessage.ErrMsg = err.Error()
			log.CtxErrorw(ctx, "change failed", "error", err)
		}
		log.CtxInfow(ctx, "change success")
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
	resp.IntegrityHash = integrityMeta.IntegrityHash
	resp.PieceHash = integrityMeta.PieceHash
	resp.IsPrimary = integrityMeta.IsPrimary
	resp.RedundancyType = integrityMeta.RedundancyType
	resp.ChallengePieceKey = piecestore.EncodeECPieceKey(req.ObjectId, req.ChallengeIdx, integrityMeta.PieceIdx)
	var data []byte
	data, err = challenge.pieceStore.GetPiece(ctx, resp.ChallengePieceKey, 0, -1)
	if err != nil {
		return
	}
	resp.PieceData = data
	return resp, nil
}
