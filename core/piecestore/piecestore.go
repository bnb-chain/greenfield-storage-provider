package piecestore

import (
	"context"
)

const (
	PrimarySPRedundancyIndex = -1
)

// PieceOp is a helper interface for piece key operator and piece size calculate.
//
//go:generate mockgen -source=./piecestore.go -destination=./piecestore_mock.go -package=piecestore
type PieceOp interface {
	// SegmentPieceKey returns the segment piece key used as the key of store piece store.
	SegmentPieceKey(objectID uint64, segmentIdx uint32) string
	// ECPieceKey returns the ec piece key used as the key of store piece store.
	ECPieceKey(objectID uint64, segmentIdx, redundancyIdx uint32) string
	// ChallengePieceKey returns the  piece key used as the key of challenge piece key.
	// if replicateIdx < 0 , returns the SegmentPieceKey, otherwise returns the ECPieceKey.
	ChallengePieceKey(objectID uint64, segmentIdx uint32, redundancyIdx int32) string
	// MaxSegmentPieceSize returns the object max segment piece size by object payload size and
	// max segment size that comes from storage params.
	MaxSegmentPieceSize(payloadSize uint64, maxSegmentSize uint64) int64
	// SegmentPieceCount returns the segment piece count of object payload by object payload size
	// and max segment size that comes from storage params.
	SegmentPieceCount(payloadSize uint64, maxSegmentSize uint64) uint32
	// SegmentPieceSize returns the segment piece size of segment index by object payload size and
	// max segment size that comes from storage params.
	SegmentPieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64) int64
	// ECPieceSize returns the ec piece size of ec index, by object payload size, max segment
	// size and chunk number that ths last two params comes from storage params.
	ECPieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64, chunkNum uint32) int64
	// ParseSegmentIdx returns the segment index according to the segment piece key
	ParseSegmentIdx(segmentKey string) (uint32, error)
	// ParseChallengeIdx returns the segment index and EC piece index  according to the challenge piece key
	ParseChallengeIdx(challengeKey string) (uint32, int32, error)
}

// PieceStore is an abstract interface to piece store that store the object payload data.
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
