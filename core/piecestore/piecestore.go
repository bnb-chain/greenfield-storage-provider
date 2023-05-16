package piecestore

import (
	"context"
)

// PieceOp is the helper interface for piece key operator and piece size calculate.
type PieceOp interface {
	// SegmentPieceKey returns the segment piece key used as the key of store piece store.
	SegmentPieceKey(objectID uint64, segmentIdx uint32) string
	// ECPieceKey returns the ec piece key used as the key of store piece store.
	ECPieceKey(objectID uint64, segmentIdx uint32, replicateIdx uint32) string
	// ChallengePieceKey returns the  piece key used as the key of challenge piece key.
	// if replicateIdx < 0 , returns the SegmentPieceKey, otherwise returns the ECPieceKey.
	ChallengePieceKey(objectID uint64, segmentIdx uint32, replicateIdx int32) string
	// MaxSegmentSize returns the object max segment size by object payload size and
	// max segment size that comes from storage params.
	MaxSegmentSize(payloadSize uint64, maxSegmentSize uint64) int64
	// SegmentCount returns the segment count of object payload by object payload size
	// and max segment size that comes from storage params.
	SegmentCount(payloadSize uint64, maxSegmentSize uint64) uint32
	// SegmentSize returns the segment size of segment index by object payload size and
	// max segment size that comes from storage params.
	SegmentSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64) int64
	// PieceSize returns the ec piece size of ec index, by object payload size, max segment
	//size and chunk number that ths last two params comes from storage params.
	PieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64, chunkNum uint32) int64
}

// PieceStore is the interface to piece store that store the object payload data.
type PieceStore interface {
	// GetPiece returns the piece data from piece store by piece key.
	// the piece can segment or ec piece key.
	GetPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error)
	// PutPiece puts the piece data to piece store, it can put segment
	// or ec piece data.
	PutPiece(ctx context.Context, key string, value []byte) error
	// DeletePiece deletes the piece data from piece store, it can delete
	// segment or ec piece data.
	DeletePiece(ctx context.Context, key string) error
}
