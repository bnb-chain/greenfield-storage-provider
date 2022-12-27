package redundancy

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestSegmentPieceEnode(t *testing.T) {
	segmentSize := 16*1024*1024 - 2
	// generate encode source data
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ctx := context.Background()
	segmentData := make([]byte, segmentSize)
	for i := range segmentData {
		segmentData[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	const ObjectID string = "testabc"
	log.Printf("origin data size :%d", len(segmentData))
	segment := NewSegment(int64(segmentSize), segmentData, 1, ObjectID)

	piecesObjects, err := EncodeSegment(ctx, segment)
	if err != nil {
		t.Errorf("segment encode fail")
	}

	log.Print("encode result:")
	for _, piece := range piecesObjects {
		log.Printf("piece Object name: %s, index:%d \n", piece.Key, piece.ECIndex)
	}

	// set 2 dataBlocks as empty, decode should success
	shardsToReocver := make([]*PieceObject, 6)

	shardsToReocver[0] = piecesObjects[0]
	shardsToReocver[1] = piecesObjects[1]
	shardsToReocver[2] = &PieceObject{}
	shardsToReocver[3] = &PieceObject{}
	shardsToReocver[4] = piecesObjects[4] // priority block
	shardsToReocver[5] = piecesObjects[5] // priority block

	start := time.Now()
	decodeSegment, err := DecodeSegment(ctx, shardsToReocver, int64(segmentSize))
	if err != nil {
		t.Errorf("segment Reconstruct fail")
	}

	fmt.Printf("decode cost time: %d", time.Since(start).Milliseconds())
	if !bytes.Equal(decodeSegment.Data, segmentData) {
		t.Errorf("compare segment data fail")
	}
	if decodeSegment.SegmentID != segment.SegmentID {
		t.Errorf("compare segment id fail ")
	}
	if decodeSegment.SegName != segment.SegName {
		t.Errorf("compare segment name fail ")
	}

	// set 1 data block and 1 priority block as empty, decode should success
	shardsToReocver[0] = piecesObjects[0]
	shardsToReocver[1] = &PieceObject{}
	shardsToReocver[2] = piecesObjects[2]
	shardsToReocver[3] = piecesObjects[3]
	shardsToReocver[4] = &PieceObject{}   // priority block
	shardsToReocver[5] = piecesObjects[5] // priority block

	decodeSegment, err = DecodeSegment(ctx, shardsToReocver, int64(segmentSize))
	if err != nil {
		t.Errorf("segment Reconstruct fail")
	}
	if !bytes.Equal(decodeSegment.Data, segmentData) {
		t.Errorf("compare fail")
	}
	if decodeSegment.SegmentID != segment.SegmentID {
		t.Errorf("compare segment id fail ")
	}
	if decodeSegment.SegName != segment.SegName {
		t.Errorf("compare segment name fail ")
	}

	// set 2 data block and 1 priority block as empty, decode should fail
	shardsToReocver[0] = piecesObjects[0]
	shardsToReocver[1] = &PieceObject{}
	shardsToReocver[2] = &PieceObject{}
	shardsToReocver[3] = piecesObjects[3]
	shardsToReocver[4] = &PieceObject{}   // priority block
	shardsToReocver[5] = piecesObjects[5] // priority block

	decodeSegment, err = DecodeSegment(ctx, shardsToReocver, int64(segmentSize))
	if err == nil {
		t.Errorf("segment decode should fail")
	}
}
