package redundancy

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestSegmentPieceEnode(t *testing.T) {
	segmentSize := 16*1024*1024 - 2
	// generate encode source data
	segmentData := initSegmentData(segmentSize)
	const ObjectID string = "testabc"
	log.Printf("origin data size :%d", len(segmentData))
	segment := NewSegment(int64(segmentSize), segmentData, 1, ObjectID)

	piecesObjects, err := EncodeSegment(segment)
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
	decodeSegment, err := DecodeSegment(shardsToReocver, int64(segmentSize))
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

	decodeSegment, err = DecodeSegment(shardsToReocver, int64(segmentSize))
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

	decodeSegment, err = DecodeSegment(shardsToReocver, int64(segmentSize))
	if err == nil {
		t.Errorf("segment decode should fail")
	}
}

func TestRawSegmentEnode(t *testing.T) {
	segmentSize := 16*1024*1024 - 2
	segmentData := initSegmentData(segmentSize)

	piecesShards, err := EncodeRawSegment(segmentData)
	if err != nil {
		t.Errorf("segment encode fail")
	}

	// set 2 dataBlock of origin as empty block
	shardsToRecover := make([][]byte, 6)
	shardsToRecover[0] = piecesShards[0]
	shardsToRecover[1] = piecesShards[1]
	shardsToRecover[2] = []byte("")
	shardsToRecover[3] = []byte("")
	shardsToRecover[4] = piecesShards[4]
	shardsToRecover[5] = piecesShards[5]

	deCodeBytes, err := DecodeRawSegment(shardsToRecover, int64(segmentSize))
	if err != nil {
		t.Errorf("decode fail")
	} else {
		log.Println("decode succ")
	}

	// compare decode data with original data
	if !bytes.Equal(deCodeBytes, segmentData) {
		t.Errorf("decode data error")
	}

	// set 2 data block and 1 priority block as empty, decode should fail
	shardsToRecover[2] = []byte("")
	shardsToRecover[3] = []byte("")
	shardsToRecover[4] = []byte("")

	deCodeBytes, err = DecodeRawSegment(shardsToRecover, int64(segmentSize))
	if err == nil {
		t.Errorf("segment decode should fail")
	}
}

func initSegmentData(segmentSize int) []byte {
	// generate encode source data
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	segmentData := make([]byte, segmentSize)
	for i := range segmentData {
		segmentData[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return segmentData
}
