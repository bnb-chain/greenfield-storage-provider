package stonenode

import (
	"testing"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/stretchr/testify/assert"
)

func Test_dispatchSecondarySP(t *testing.T) {
	spList := []string{"sp1", "sp2", "sp3", "sp4", "sp5", "sp6"}
	cases := []struct {
		name         string
		req1         [][][]byte
		req2         ptypes.RedundancyType
		req3         []string
		req4         []uint32
		wantedResult int
		wantedErr    error
	}{
		{
			name:         "ec type dispatch",
			req1:         dispatchECPiece(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:         spList,
			req4:         []uint32{0, 1, 2, 3, 4, 5},
			wantedResult: 6,
			wantedErr:    nil,
		},
		{
			name:         "replica type dispatch",
			req1:         dispatchSegmentPieceSlice(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			req3:         spList,
			req4:         []uint32{0, 1, 2},
			wantedResult: 3,
			wantedErr:    nil,
		},
		{
			name:         "inline type dispatch",
			req1:         dispatchInlinePieceSlice(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE,
			req3:         spList,
			req4:         []uint32{0},
			wantedResult: 1,
			wantedErr:    nil,
		},
		{
			name:         "ec type data retransmission",
			req1:         dispatchECPiece(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:         spList,
			req4:         []uint32{2, 3},
			wantedResult: 2,
			wantedErr:    nil,
		},
		{
			name:         "replica type data retransmission",
			req1:         dispatchSegmentPieceSlice(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			req3:         spList,
			req4:         []uint32{1, 2},
			wantedResult: 2,
			wantedErr:    nil,
		},
		{
			name:         "wrong secondary sp number",
			req1:         dispatchECPiece(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:         []string{},
			req4:         []uint32{0, 1, 2, 3, 4, 5},
			wantedResult: 0,
			wantedErr:    merrors.ErrSecondarySPNumber,
		},
		{
			name:         "wrong ec segment data length",
			req1:         dispatchSegmentPieceSlice(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:         spList,
			req4:         []uint32{0, 1, 2, 3, 4, 5},
			wantedResult: 0,
			wantedErr:    merrors.ErrInvalidECData,
		},
		{
			name:         "invalid piece data",
			req1:         nil,
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			req3:         spList,
			req4:         []uint32{0, 1, 2, 3, 4, 5},
			wantedResult: 0,
			wantedErr:    merrors.ErrInvalidPieceData,
		},
	}

	node := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := node.dispatchSecondarySP(tt.req1, tt.req2, tt.req3, tt.req4)
			assert.Equal(t, tt.wantedErr, err)
			assert.Equal(t, tt.wantedResult, len(result))
		})
	}
}
