package piecestore

import (
	"context"
)

type PieceOp interface {
	SegmentPieceKey(objectID uint64, segmentIdx uint32) string
	ECPieceKey(objectID uint64, segmentIdx uint32, replicateIdx uint32) string
	ChallengePieceKey(objectID uint64, segmentIdx uint32, replicateIdx int32) string
	MaxSegmentSize(payloadSize uint64, maxSegmentSize uint64) int64
	SegmentCount(payloadSize uint64, maxSegmentSize uint64) uint32
	SegmentSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64) int64
	PieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64, chunkNum uint32) int64
}

type PieceStore interface {
	GetPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error)
	PutPiece(ctx context.Context, key string, value []byte) error
	DeletePiece(ctx context.Context, key string) error
}
