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
// return the piece's integrity hash, piece hash and piece data
func (challenge *Challenge) ChallengePiece(ctx context.Context, req *types.ChallengePieceRequest) (*types.ChallengePieceResponse, error) {
	var (
		pieceKey              string
		approximatedPieceSize int
		err                   error
		resp                  *types.ChallengePieceResponse
	)

	ctx = log.Context(ctx, req)
	params, err := challenge.spDB.GetStorageParams()
	if err != nil {
		return resp, err
	}
	if req.GetRedundancyIdx() < 0 {
		// useless iff it is a segment piece
		pieceKey = piecestore.EncodeSegmentPieceKey(req.GetObjectId(), req.GetSegmentIdx())
		approximatedPieceSize = int(params.GetMaxSegmentSize())
	} else {
		// as the ec piece index iff it is a ec piece
		pieceKey = piecestore.EncodeECPieceKey(req.GetObjectId(),
			req.GetSegmentIdx(), uint32(req.GetRedundancyIdx()))
		approximatedPieceSize = int(math.Ceil(float64(params.GetMaxSegmentSize()) / float64(params.GetRedundantDataChunkNum())))
	}

	// allocates memory from resource manager
	scope, err := challenge.rcScope.BeginSpan()
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin reserve resource", "error", err)
		return resp, err
	}
	stateFunc := func() string {
		var state string
		rcmgr.ResrcManager().ViewSystem(func(scope rcmgr.ResourceScope) error {
			state = scope.Stat().String()
			return nil
		})
		return state
	}
	err = scope.ReserveMemory(approximatedPieceSize, rcmgr.ReservationPriorityAlways)
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve memory from resource manager",
			"reserve_size", approximatedPieceSize, "resource_state", stateFunc(), "error", err)
		return resp, err
	}
	defer func() {
		scope.Done()
		log.CtxDebugw(ctx, "end challenge piece request", "resource_state", stateFunc())
	}()

	integrity, err := challenge.spDB.GetObjectIntegrity(req.GetObjectId())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get integrity hash from db", "error", err)
		err = merrors.InnerErrorToGRPCError(err)
		return resp, err
	}
	resp = &types.ChallengePieceResponse{}
	resp.IntegrityHash = integrity.IntegrityHash
	resp.PieceHash = integrity.Checksum
	resp.PieceData, err = challenge.pieceStore.GetPiece(ctx, pieceKey, 0, -1)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get piece data", "error", err)
		err = merrors.InnerErrorToGRPCError(err)
		return resp, err
	}
	log.CtxInfow(ctx, "succeed to challenge piece", "object_id", req.GetObjectId(),
		"piece_idx", req.GetSegmentIdx(), "redundancy_idx", req.GetRedundancyIdx(), "segment_count", len(integrity.Checksum))
	return resp, err
}
