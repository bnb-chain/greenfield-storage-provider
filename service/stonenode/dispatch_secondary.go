package stonenode

import (
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// dispatchSecondarySP dispatch piece data to secondary storage provider.
// returned map key is spID, value map key is ec piece key or segment key, value map's value is piece data

// pieceDataBySegment key is segment key; if redundancyType is EC, value is [][]byte type,
// a two-dimensional array which contains ec data from ec1 []byte data to ec6 []byte data
// if redundancyType is replica or inline, value is [][]byte type, a two-dimensional array
// which only contains one []byte data

// pieceDataBySegment is a three-dimensional slice, first dimensional is segment index, second is [][]byte data
// dispatchSecondarySP convert pieceDataBySegment to ec dimensional slice, first dimensional is ec number such as ec1, contains [][]byte data
func (node *StoneNodeService) dispatchSecondarySP(pieceDataBySegment [][][]byte, redundancyType ptypesv1pb.RedundancyType, secondarySPs []string,
	targetIdx []uint32) ([][][]byte, error) {
	if len(pieceDataBySegment) == 0 {
		return nil, merrors.ErrInvalidPieceData
	}
	if len(secondarySPs) == 0 {
		return nil, merrors.ErrSecondarySPNumber
	}
	var pieceDataBySecondary [][][]byte
	var err error
	switch redundancyType {
	case ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceDataBySecondary, err = dispatchReplicaOrInlineData(pieceDataBySegment, secondarySPs, targetIdx)
	default: // ec type
		pieceDataBySecondary, err = dispatchECData(pieceDataBySegment, secondarySPs, targetIdx)
	}
	if err != nil {
		log.Errorw("dispatch piece data by secondary error", "error", err)
		return nil, err
	}
	return pieceDataBySecondary, nil
}

// dispatchReplicaOrInlineData dispatches replica or inline data into different sp, each sp should store all segments data of an object
// if an object uses replica type, it's split into 10 segments and there are 6 sp, each sp should store 10 segments data
// if an object uses inline type, there is only one segment and there are 6 sp, each sp should store 1 segment data
func dispatchReplicaOrInlineData(pieceDataBySegment [][][]byte, secondarySPs []string, targetIdx []uint32) ([][][]byte, error) {
	if len(secondarySPs) < len(targetIdx) {
		return nil, merrors.ErrSecondarySPNumber
	}

	segmentLength := len(pieceDataBySegment[0])
	if segmentLength != 1 {
		return nil, merrors.ErrInvalidSegmentData
	}

	dataSlice := make([][][]byte, segmentLength)
	for i := 0; i < segmentLength; i++ {
		dataSlice[i] = make([][]byte, 0)
		for j := 0; j < len(pieceDataBySegment); j++ {
			dataSlice[i] = append(dataSlice[i], pieceDataBySegment[j][i])
		}
	}
	segmentPieceSlice := make([][][]byte, len(targetIdx))
	for i := 0; i < len(targetIdx); i++ {
		segmentPieceSlice[i] = segmentPieceSlice[0]
	}
	log.Infow("segmentPieceSlice", "length", len(segmentPieceSlice), "content 0", segmentPieceSlice[0])
	return segmentPieceSlice, nil
}

// dispatchECData dispatched ec data into different sp
// one sp stores same ec column data: sp1 stores all ec1 data, sp2 stores all ec2 data, etc
func dispatchECData(pieceDataBySegment [][][]byte, secondarySPs []string, targetIdx []uint32) ([][][]byte, error) {
	segmentLength := len(pieceDataBySegment[0])
	if segmentLength < 6 {
		return nil, merrors.ErrInvalidECData
	}

	pieceSlice := make([][][]byte, segmentLength)
	for i := 0; i < segmentLength; i++ {
		if i > len(secondarySPs) {
			return nil, merrors.ErrSecondarySPNumber
		}
		pieceSlice[i] = make([][]byte, 0)
		for j := 0; j < len(pieceDataBySegment); j++ {
			pieceSlice[i] = append(pieceSlice[i], pieceDataBySegment[j][i])
		}
	}

	ecPieceSlice := make([][][]byte, len(targetIdx))
	for index, value := range targetIdx {
		ecPieceSlice[index] = pieceSlice[value]
	}
	return ecPieceSlice, nil
}
