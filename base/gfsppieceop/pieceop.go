package gfsppieceop

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
)

var _ piecestore.PieceOp = &GfSpPieceOp{}

type GfSpPieceOp struct {
}

func (p *GfSpPieceOp) SegmentPieceKey(objectID uint64, segmentIdx uint32) string {
	return fmt.Sprintf("%d_s%d", objectID, segmentIdx)
}

func (p *GfSpPieceOp) ECPieceKey(objectID uint64, segmentIdx uint32, replicateIdx uint32) string {
	return fmt.Sprintf("%d_s%d_p%d", objectID, segmentIdx, replicateIdx)
}

func (p *GfSpPieceOp) ChallengePieceKey(objectID uint64, segmentIdx uint32, replicateIdx int32) string {
	if replicateIdx < 0 {
		return p.SegmentPieceKey(objectID, segmentIdx)
	}
	return p.ECPieceKey(objectID, segmentIdx, uint32(replicateIdx))
}

func (p *GfSpPieceOp) MaxSegmentSize(payloadSize uint64, maxSegmentSize uint64) int64 {
	if payloadSize > maxSegmentSize {
		return int64(maxSegmentSize)
	}
	return int64(payloadSize)
}

func (p *GfSpPieceOp) SegmentSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64) int64 {
	segmentCount := p.SegmentCount(payloadSize, maxSegmentSize)
	if segmentCount == 1 {
		return int64(payloadSize)
	} else if segmentIdx == segmentCount-1 {
		return int64(payloadSize) - (int64(segmentCount)-1)*int64(maxSegmentSize)
	} else {
		return int64(maxSegmentSize)
	}
}

func (p *GfSpPieceOp) PieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64, dataChunkNum uint32) int64 {
	segmentCount := p.SegmentCount(payloadSize, maxSegmentSize)
	if segmentCount == 1 {
		return int64(payloadSize) / int64(dataChunkNum)
	} else if segmentIdx == segmentCount-1 {
		return int64(payloadSize) - (int64(segmentCount)-1)*int64(maxSegmentSize)/int64(dataChunkNum)
	} else {
		return int64(maxSegmentSize) / int64(dataChunkNum)
	}
}

func (p *GfSpPieceOp) SegmentCount(payloadSize uint64, maxSegmentSize uint64) uint32 {
	count := payloadSize / maxSegmentSize
	if payloadSize%maxSegmentSize > 0 {
		count++
	}
	return uint32(count)
}
