package challenge

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model"
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
	if req.GetObjectInfo() == nil {
		return nil, merrors.ErrDanglingPointer
	}
	var (
		scope                rcmgr.ResourceScopeSpan
		resp                 *types.ChallengePieceResponse
		err                  error
		pieceType            model.PieceType
		pieceKey             string
		approximatePieceSize int
		objectInfo           = req.GetObjectInfo()
	)
	ctx = log.WithValue(ctx, "object_id", objectInfo.Id.String())

	defer func() {
		if scope != nil {
			scope.Done()
		}
		log.CtxInfow(ctx, "finished to challenge piece request", "resource_state", rcmgr.GetServiceState(model.ChallengeService), "error", err)
	}()
	scope, err = challenge.rcScope.BeginSpan()
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin reserve resource", "error", err)
		return resp, err
	}

	params, err := challenge.spDB.GetStorageParams()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get storage params", "error", err)
		return resp, err
	}
	if req.GetRedundancyIdx() < 0 {
		pieceType = model.SegmentPieceType
		// useless iff it is a segment piece
		pieceKey = piecestore.EncodeSegmentPieceKey(objectInfo.Id.Uint64(), req.GetSegmentIdx())
	} else {
		pieceType = model.ECPieceType
		// as the ec piece index iff it is an ec piece
		pieceKey = piecestore.EncodeECPieceKey(objectInfo.Id.Uint64(),
			req.GetSegmentIdx(), uint32(req.GetRedundancyIdx()))
	}
	approximatePieceSize, err = piecestore.ComputeApproximatePieceSize(objectInfo,
		params.VersionedParams.GetMaxSegmentSize(), params.VersionedParams.GetRedundantDataChunkNum(), pieceType, req.GetSegmentIdx())
	if err != nil {
		log.CtxErrorw(ctx, "failed to compute Approximate piece size",
			"reserve_size", approximatePieceSize, "error", err)
		return resp, err
	}
	err = scope.ReserveMemory(approximatePieceSize, rcmgr.ReservationPriorityAlways)
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve memory from resource manager",
			"reserve_size", approximatePieceSize, "error", err)
		return resp, err
	}

	integrity, err := challenge.spDB.GetObjectIntegrity(objectInfo.Id.Uint64())
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
	log.CtxInfow(ctx, "succeed to challenge piece", "piece_idx", req.GetSegmentIdx(),
		"redundancy_idx", req.GetRedundancyIdx(), "segment_count", len(integrity.Checksum))
	return resp, err
}
