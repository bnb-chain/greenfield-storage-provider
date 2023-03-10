package challenge

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge/types"
)

// ChallengePiece handles the piece challenge request
// return the replica's integrity hash, piece hash and piece data
func (challenge *Challenge) ChallengePiece(
	ctx context.Context,
	req *types.ChallengePieceRequest) (
	resp *types.ChallengePieceResponse, err error) {
	ctx = log.Context(ctx, req)
	resp = &types.ChallengePieceResponse{}

	integrity, err := challenge.spDB.GetObjectIntegrity(req.ObjectId.String())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get integrity hash from db", "error", err)
		return
	}
	resp.IntegrityHash = integrity.IntegrityHash
	resp.PieceHash = integrity.Checksum

	var key string
	if req.GetReplicaIdx() < 0 {
		key = piecestore.EncodeSegmentPieceKey(req.ObjectId.String(), req.GetSegmentIdx())
	} else {
		key = piecestore.EncodeECPieceKey(req.ObjectId.String(),
			uint32(req.GetReplicaIdx()), req.GetSegmentIdx())
	}
	resp.PieceData, err = challenge.pieceStore.GetSegment(ctx, key, 0, -1)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get payload", "error", err)
		return
	}
	log.CtxInfow(ctx, "success to challenge the payload")
	return
}
