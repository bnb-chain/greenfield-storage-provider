package stonenode

import (
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

func setup(t *testing.T) *StoneNodeService {
	return &StoneNodeService{
		cfg: &StoneNodeConfig{
			Address:                "test1",
			StoneHubServiceAddress: "test2",
			StorageProvider:        "test",
			StoneJobLimit:          0,
		},
		name:       model.StoneNodeService,
		stoneLimit: 0,
	}
}

func mockAllocResp(objectID uint64, payloadSize uint64, redundancyType ptypes.RedundancyType) *stypes.StoneHubServiceAllocStoneJobResponse {
	return &stypes.StoneHubServiceAllocStoneJobResponse{
		TraceId: "123456",
		PieceJob: &stypes.PieceJob{
			ObjectId:       objectID,
			PayloadSize:    payloadSize,
			TargetIdx:      nil,
			RedundancyType: redundancyType,
		},
		ErrMessage: &stypes.ErrMessage{
			ErrCode: 0,
			ErrMsg:  "Success",
		},
	}
}

func dispatchECPiece() [][][]byte {
	ecList1 := [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5"), []byte("6")}
	ecList2 := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e"), []byte("f")}
	pSlice := make([][][]byte, 0)
	pSlice = append(pSlice, ecList1)
	pSlice = append(pSlice, ecList2)
	return pSlice
}

func dispatchSegmentPieceSlice() [][][]byte {
	segmentList1 := [][]byte{[]byte("10")}
	segmentList2 := [][]byte{[]byte("20")}
	segmentList3 := [][]byte{[]byte("30")}
	segSlice := make([][][]byte, 0)
	segSlice = append(segSlice, segmentList1)
	segSlice = append(segSlice, segmentList2)
	segSlice = append(segSlice, segmentList3)
	return segSlice
}

func dispatchInlinePieceSlice() [][][]byte {
	inlineList := [][]byte{[]byte("+")}
	inlineSlice := make([][][]byte, 0)
	inlineSlice = append(inlineSlice, inlineList)
	return inlineSlice
}
