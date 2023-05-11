package downloader

import (
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitToSegmentPieceInfos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	//nolint:all
	mockDB := sqldb.NewMockSPDB(ctrl) //nolint:all
	d := &Downloader{
		spDB: mockDB,
	}
	mockDB.EXPECT().GetStorageParams().Return(&storagetypes.Params{
		VersionedParams: storagetypes.VersionedParams{
			MaxSegmentSize: 16 * 1024 * 1024,
		},
	}, nil).MaxTimes(100)

	testCases := []struct {
		name             string
		d                *Downloader
		objectID         uint64
		objectSize       uint64
		startOffset      uint64
		endOffset        uint64
		isErr            bool
		wantedPieceCount int
	}{
		{
			"invalid params",
			d,
			1,
			1,
			1,
			0,
			true,
			0,
		},
		{
			"1 byte object full read",
			d,
			1,
			1,
			0,
			0,
			false,
			1,
		},
		{
			"16MB byte object full read",
			d,
			1,
			16 * 1024 * 1024,
			0,
			16*1024*1024 - 1,
			false,
			1,
		},
		{
			"16MB byte object range read",
			d,
			1,
			16 * 1024 * 1024,
			0,
			11,
			false,
			1,
		},
		{
			"17MB byte object range read, in first piece",
			d,
			1,
			17 * 1024 * 1024,
			0,
			11,
			false,
			1,
		},
		{
			"17MB byte object range read, in second piece",
			d,
			1,
			17 * 1024 * 1024,
			16 * 1024 * 1024,
			16*1024*1024 + 11,
			false,
			1,
		},
		{
			"17MB byte object range read, in two piece",
			d,
			1,
			17 * 1024 * 1024,
			16*1024*1024 - 11,
			16*1024*1024 + 11,
			false,
			2,
		},
		{
			"17MB byte object full read",
			d,
			1,
			17 * 1024 * 1024,
			0,
			17*1024*1024 - 1,
			false,
			2,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			pieceInfos, err := testCase.d.SplitToSegmentPieceInfos(testCase.objectID, testCase.objectSize,
				testCase.startOffset, testCase.endOffset)
			if !testCase.isErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}
			assert.Equal(t, testCase.wantedPieceCount, len(pieceInfos))
			realLength := uint64(0)
			for _, p := range pieceInfos {
				realLength += p.length
			}
			assert.Equal(t, testCase.endOffset-testCase.startOffset+1, realLength)
		})
	}
}
