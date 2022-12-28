package redundancy

import (
	"context"
	"strconv"
	"strings"

	"github.com/bnb-chain/inscription-storage-provider/pkg/redundancy/erasure"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// PieceObject - details of the erasure encoded piece
type PieceObject struct {
	Key       string
	ECdata    []byte
	ECIndex   int
	PieceSize int
}

// Segment - detail of segment split from objects
type Segment struct {
	SegName     string
	SegmentSize int64
	SegmentID   int
	Data        []byte
}

const DATABLOCKS int = 4
const PARITYBLOCKS int = 2

// NewSegment creates a new Segment object
func NewSegment(size int64, content []byte, segmentId int, objectId string) *Segment {
	return &Segment{
		SegName:     objectId + "_s" + strconv.Itoa(segmentId),
		SegmentSize: size,
		SegmentID:   segmentId,
		Data:        content,
	}
}

// EncodeSegment encode to segment, return the piece content and the meta of pieces
func EncodeSegment(ctx context.Context, s *Segment) ([]*PieceObject, error) {
	erasure, err := erasure.NewRSEncoder(DATABLOCKS, PARITYBLOCKS, s.SegmentSize)
	if err != nil {
		log.Error("new RSEncoder fail", err.Error())
		return nil, err
	}
	shards, err := erasure.EncodeData(ctx, s.Data)
	if err != nil {
		log.Errorf("encode data fail %s, segment name: %s ", err, s.SegName)
		return nil, err
	}

	pieceObjectList := make([]*PieceObject, DATABLOCKS+PARITYBLOCKS)
	for index, shard := range shards {
		piece := &PieceObject{
			Key:       s.SegName + "_p" + strconv.Itoa(index),
			ECdata:    shard,
			ECIndex:   index,
			PieceSize: len(shard),
		}
		pieceObjectList[index] = piece
	}

	return pieceObjectList, nil
}

// DecodeSegment decode with the pieceObjects and reconstruct the original segment
func DecodeSegment(ctx context.Context, pieces []*PieceObject, segmentSize int64) (*Segment, error) {
	encoder, err := erasure.NewRSEncoder(DATABLOCKS, PARITYBLOCKS, segmentSize)
	if err != nil {
		log.Error("new RSEncoder fail", err.Error())
		return nil, err
	}

	pieceObjectsData := make([][]byte, DATABLOCKS+PARITYBLOCKS)
	for i := 0; i < DATABLOCKS+PARITYBLOCKS; i++ {
		pieceObjectsData[i] = pieces[i].ECdata
	}

	deCodeBytes, err := encoder.GetOriginalData(ctx, pieceObjectsData, segmentSize)
	if err != nil {
		log.Errorf("reconstruct segment content fail %s", err)
		return nil, err
	}

	// construct the segmentId and segmentName from piece key
	pieceName := pieces[0].Key
	segIndex := strings.Index(pieceName, "_s")
	ecIndex := strings.Index(pieceName, "_p")

	segIdStr := pieceName[segIndex+2 : ecIndex]
	segId, err := strconv.Atoi(segIdStr)
	if err != nil {
		log.Errorf("fetch segment ID fail, invalid number: %s", segIdStr)
		return nil, err
	}

	return &Segment{
		SegmentSize: segmentSize,
		SegName:     pieceName[:ecIndex],
		SegmentID:   segId,
		Data:        deCodeBytes,
	}, nil

}
