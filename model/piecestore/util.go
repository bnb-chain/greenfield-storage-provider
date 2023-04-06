package piecestore

import (
	"math"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
)

// ComputeApproximatePieceSize return the size of piece data
func ComputeApproximatePieceSize(object *storagetypes.ObjectInfo, maxSegmentPieceSize uint64, dataChunkNum uint32,
	pieceType model.PieceType, idx uint32) (int, error) {
	if object.GetPayloadSize() == 0 {
		return 0, merrors.ErrPayloadZero
	}
	switch pieceType {
	case model.SegmentPieceType:
		segmentCnt := ComputeSegmentCount(object.GetPayloadSize(), maxSegmentPieceSize)
		if idx > segmentCnt-1 {
			return 0, merrors.ErrInvalidParams
		}
		if idx == segmentCnt-1 {
			return int(object.GetPayloadSize() - maxSegmentPieceSize*(uint64(idx))), nil
		}
		return int(maxSegmentPieceSize), nil
	case model.ECPieceType:
		segmentCnt := ComputeSegmentCount(object.GetPayloadSize(), maxSegmentPieceSize)
		if idx > segmentCnt-1 {
			return 0, merrors.ErrInvalidParams
		}
		segmentPieceSize := maxSegmentPieceSize
		if idx == segmentCnt-1 {
			// ignore the size of a small amount of data filled by ec encoding
			segmentPieceSize = object.GetPayloadSize() - maxSegmentPieceSize*(uint64(idx))
		}
		return int(math.Ceil(float64(segmentPieceSize) / float64(dataChunkNum))), nil
	default:
		return 0, merrors.ErrInvalidParams
	}
}
