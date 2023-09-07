package gfsppieceop

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
)

var _ piecestore.PieceOp = &GfSpPieceOp{}

type GfSpPieceOp struct{}

func (p *GfSpPieceOp) SegmentPieceKey(objectID uint64, segmentIdx uint32) string {
	return fmt.Sprintf("s%d_s%d", objectID, segmentIdx)
}

func (p *GfSpPieceOp) ECPieceKey(objectID uint64, segmentIdx, redundancyIdx uint32) string {
	return fmt.Sprintf("e%d_s%d_p%d", objectID, segmentIdx, redundancyIdx)
}

func (p *GfSpPieceOp) ChallengePieceKey(objectID uint64, segmentIdx uint32, redundancyIdx int32) string {
	if redundancyIdx < 0 {
		return p.SegmentPieceKey(objectID, segmentIdx)
	}
	return p.ECPieceKey(objectID, segmentIdx, uint32(redundancyIdx))
}

func (p *GfSpPieceOp) MaxSegmentPieceSize(payloadSize uint64, maxSegmentSize uint64) int64 {
	if payloadSize > maxSegmentSize {
		return int64(maxSegmentSize)
	}
	return int64(payloadSize)
}

func (p *GfSpPieceOp) SegmentPieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64) int64 {
	segmentCount := p.SegmentPieceCount(payloadSize, maxSegmentSize)
	if segmentCount == 1 {
		return int64(payloadSize)
	} else if segmentIdx == segmentCount-1 {
		return int64(payloadSize) - (int64(segmentCount)-1)*int64(maxSegmentSize)
	} else {
		return int64(maxSegmentSize)
	}
}

func (p *GfSpPieceOp) ECPieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64, dataChunkNum uint32) int64 {
	segmentSize := p.SegmentPieceSize(payloadSize, segmentIdx, maxSegmentSize)

	ECPieceSize := segmentSize / int64(dataChunkNum)
	// EC padding will cause the EC pieces to have one extra byte if it cannot be evenly divided.
	// for example, the segment size is 15, the ec piece size should be 15/4 + 1 = 4
	// TODO add test case
	if segmentSize > 0 && segmentSize%int64(dataChunkNum) != 0 {
		ECPieceSize++
	}

	return ECPieceSize
}

func (p *GfSpPieceOp) SegmentPieceCount(payloadSize uint64, maxSegmentSize uint64) uint32 {
	count := payloadSize / maxSegmentSize
	if payloadSize%maxSegmentSize > 0 {
		count++
	}
	return uint32(count)
}

func (p *GfSpPieceOp) ParseSegmentIdx(segmentKey string) (uint32, error) {
	keyParts := strings.Split(segmentKey, "_")
	if len(keyParts) != 2 {
		return 0, fmt.Errorf("invalid segmentKey format")
	}

	segPrefix := "s"
	segmentIdx, err := strconv.ParseUint(keyParts[1][len(segPrefix):], 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(segmentIdx), nil
}

func (p *GfSpPieceOp) ParseECPieceKeyIdx(ecPieceKey string) (uint32, int32, error) {
	keyParts := strings.Split(ecPieceKey, "_")
	if len(keyParts) != 3 {
		return 0, 0, fmt.Errorf("invalid EC piece key: %s", ecPieceKey)
	}
	segPrefix := "s"
	redundancyPrefix := "p"

	segmentIdx, err := strconv.ParseUint(keyParts[1][len(segPrefix):], 10, 32)
	if err != nil {
		return 0, 0, err
	}
	redundancyIndex, err := strconv.ParseUint(keyParts[2][len(redundancyPrefix):], 10, 32)
	if err != nil {
		return 0, 0, err
	}
	return uint32(segmentIdx), int32(redundancyIndex), nil
}

func (p *GfSpPieceOp) ParseChallengeIdx(challengeKey string) (uint32, int32, error) {
	keyParts := strings.Split(challengeKey, "_")

	if len(keyParts) == 2 {
		segmentIdx, err := p.ParseSegmentIdx(challengeKey)
		if err != nil {
			return 0, 0, err
		}
		return segmentIdx, -1, nil
	} else if len(keyParts) == 3 {
		return p.ParseECPieceKeyIdx(challengeKey)
	}

	return 0, 0, fmt.Errorf("invalid challenge key: %s", challengeKey)
}
