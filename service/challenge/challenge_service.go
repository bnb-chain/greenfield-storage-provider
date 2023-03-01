package challenge

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// ChallengePiece implement challenge service server interface and handle the grpc request.
func (challenge *Challenge) ChallengePiece(
	ctx context.Context,
	req *types.ChallengePieceRequest) (
	resp *types.ChallengePieceResponse, err error) {
	ctx = log.Context(ctx, req)
	resp = &types.ChallengePieceResponse{}

	integrity, err := challenge.spDb.GetObjectIntegrity(req.GetObjectId())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get integrity hash from db", "error", err)
		return
	}
	resp.IntegrityHash = integrity.IntegrityHash
	resp.PieceHash = integrity.Checksum

	var key string
	if req.GetReplicateIdx() < 0 {
		key = piecestore.EncodeSegmentPieceKey(req.GetObjectId(), req.GetSegmentIdx())
	} else {
		key = piecestore.EncodeECPieceKey(req.GetObjectId(),
			uint32(req.GetReplicateIdx()), req.GetSegmentIdx())
	}
	resp.PieceData, err = challenge.pieceStore.GetSegment(ctx, key, 0, -1)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get payload", "error", err)
		return
	}
	log.CtxInfow(ctx, "success to challenge the payload")
	return
}
