package challenge

import (
	"context"
	"math"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge/types"
)

var _ types.ChallengeServiceServer = &Challenge{}

// ChallengePiece handles the piece challenge request
// return the replica's integrity hash, piece hash and piece data
func (challenge *Challenge) ChallengePiece(
	ctx context.Context,
	req *types.ChallengePieceRequest) (
	resp *types.ChallengePieceResponse, err error) {
	ctx = log.Context(ctx, req)
	resp = &types.ChallengePieceResponse{}

	integrity, err := challenge.spDB.GetObjectIntegrity(req.GetObjectId())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get integrity hash from db", "error", err)
		err = merrors.InnerErrorToGRPCError(err)
		return
	}
	resp.IntegrityHash = integrity.IntegrityHash
	resp.PieceHash = integrity.Checksum

	params, err := challenge.spDB.GetStorageParams()
	if err != nil {
		return
	}
	var key string
	var memSize int
	if req.GetReplicaIdx() < 0 {
		key = piecestore.EncodeSegmentPieceKey(req.GetObjectId(), req.GetSegmentIdx())
		memSize = int(params.GetMaxSegmentSize())
	} else {
		key = piecestore.EncodeECPieceKey(req.GetObjectId(),
			req.GetSegmentIdx(), uint32(req.GetReplicaIdx()))
		memSize = int(math.Ceil(float64(params.GetMaxSegmentSize()) / float64(params.GetRedundantDataChunkNum())))
	}
	scope, err := challenge.rcScope.BeginSpan()
	if err != nil {
		return
	}
	scope.ReserveMemory(memSize, rcmgr.ReservationPriorityAlways)
	defer scope.ReleaseMemory(memSize)
	resp.PieceData, err = challenge.pieceStore.GetSegment(ctx, key, 0, -1)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get payload", "error", err)
		err = merrors.InnerErrorToGRPCError(err)
		return
	}
	log.CtxInfow(ctx, "succeed to challenge the payload", "object_id", req.GetObjectId(),
		"piece_idx", req.GetSegmentIdx(), "replicate_idx", req.GetReplicaIdx(), "segment_count", len(integrity.Checksum))
	return
}
