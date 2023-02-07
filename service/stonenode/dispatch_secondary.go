package stonenode

import (
	"errors"
	"fmt"

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
		return nil, errors.New("invalid data length")
	}
	pieceDataBySecondary := make([][][]byte, 0)
	var err error
	switch redundancyType {
	case ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceDataBySecondary, err = dispatchReplicaOrInlineData(pieceDataBySegment, secondarySPs, targetIdx)
	default: // ec type
		pieceDataBySecondary, err = dispatchECData(pieceDataBySegment, secondarySPs, targetIdx)
	}
	if err != nil {
		log.Errorw("fill piece data by secondary error", "error", err)
		return nil, err
	}
	return pieceDataBySecondary, nil
}

// dispatchReplicaOrInlineData dispatches replica or inline data into different sp, each sp should store all segments data of an object
// if an object uses replica type, it's split into 10 segments and there are 6 sp, each sp should store 10 segments data
// if an object uses inline type, there is only one segment and there are 6 sp, each sp should store 1 segment data
func dispatchReplicaOrInlineData(pieceDataBySegment [][][]byte, secondarySPs []string, targetIdx []uint32) ([][][]byte, error) {
	targetIdxLength := len(targetIdx)
	if len(secondarySPs) < targetIdxLength {
		return nil, merrors.ErrSecondarySPNumber
	}
	segmentLength := len(pieceDataBySegment[0])
	//riPieceDataSlice := make([][]byte, segmentLength)
	//for i := 0; i < segmentLength; i++ {
	//	riPieceDataSlice[i] = make([]byte, 0)
	//	for j := 0; j < len(pieceDataBySegment); j++ {
	//		riPieceDataSlice[i] = append(riPieceDataSlice[i], pieceDataBySegment[j][i][0])
	//	}
	//}
	//pds := make([][][]byte, targetIdxLength)
	//for i := 0; i < targetIdxLength; i++ {
	//	pds[i] = riPieceDataSlice
	//}
	newData := make([][][]byte, segmentLength)
	for i := 0; i < segmentLength; i++ {
		newData[i] = make([][]byte, 0)
		for j := 0; j < len(pieceDataBySegment); j++ {
			newData[i] = append(newData[i], pieceDataBySegment[j][i])
		}
	}
	pds := make([][][]byte, len(targetIdx))
	for i := 0; i < len(targetIdx); i++ {
		pds[i] = newData[0]
	}

	log.Infow("length", "pds length", len(pds), "newData length", len(newData))
	//for i, j := range pds {
	//	log.Infow("print pds meta", "index", i, "inner array length", len(j))
	//}
	return pds, nil
}

func dRID(data [][][]byte, targetIdx []uint32) [][][]byte {
	segmentLength := len(data[0])
	fmt.Println(segmentLength)
	newData := make([][][]byte, segmentLength)

	for i := 0; i < segmentLength; i++ {
		newData[i] = make([][]byte, 0)
		for j := 0; j < len(data); j++ {
			newData[i] = append(newData[i], data[j][i])
		}
	}
	pds := make([][][]byte, len(targetIdx))
	for i := 0; i < len(targetIdx); i++ {
		pds[i] = newData[0]
	}

	return pds
}

// dispatchECData dispatched ec data into different sp
// one sp stores same ec column data: sp1 stores all ec1 data, sp2 stores all ec2 data, etc
func dispatchECData(pieceDataBySegment [][][]byte, secondarySPs []string, targetIdx []uint32) ([][][]byte, error) {
	segmentLength := len(pieceDataBySegment[0])
	ecPieceDataSlice := make([][][]byte, segmentLength)
	for i := 0; i < segmentLength; i++ {
		if i > len(secondarySPs) {
			return nil, merrors.ErrSecondarySPNumber
		}
		for _, idx := range targetIdx {
			if int(idx) == i {
				ecPieceDataSlice[i] = make([][]byte, 0)
				for j := 0; j < len(pieceDataBySegment); j++ {
					ecPieceDataSlice[i] = append(ecPieceDataSlice[i], pieceDataBySegment[j][i])
				}
			}
		}
	}
	return ecPieceDataSlice, nil
}
