package downloader

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfsppieceop"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestSplitToSegmentPieceInfos(t *testing.T) {
	var (
		task          = &gfsptask.GfSpDownloadObjectTask{}
		objectInfo    = &storagetypes.ObjectInfo{}
		storageParams = &storagetypes.Params{}
	)
	storageParams.VersionedParams.MaxSegmentSize = 16 * 1024 * 1024

	testCases := []struct {
		name             string
		objectID         uint64
		objectSize       uint64
		startOffset      uint64
		endOffset        uint64
		isErr            bool
		wantedPieceCount int
	}{
		{
			"invalid params",
			1,
			1,
			1,
			0,
			true,
			0,
		},
		{
			"1 byte object full read",
			1,
			1,
			0,
			0,
			false,
			1,
		},
		{
			"16MB byte object full read",
			1,
			16 * 1024 * 1024,
			0,
			16*1024*1024 - 1,
			false,
			1,
		},
		{
			"16MB byte object range read",
			1,
			16 * 1024 * 1024,
			0,
			11,
			false,
			1,
		},
		{
			"17MB byte object range read, in first piece",
			1,
			17 * 1024 * 1024,
			0,
			11,
			false,
			1,
		},
		{
			"17MB byte object range read, in second piece",
			1,
			17 * 1024 * 1024,
			16 * 1024 * 1024,
			16*1024*1024 + 11,
			false,
			1,
		},
		{
			"17MB byte object range read, in two piece",
			1,
			17 * 1024 * 1024,
			16*1024*1024 - 11,
			16*1024*1024 + 11,
			false,
			2,
		},
		{
			"17MB byte object full read",
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
			objectInfo.PayloadSize = testCase.objectSize
			objectInfo.Id = sdkmath.NewUint(testCase.objectID)
			task.InitDownloadObjectTask(
				objectInfo, nil, storageParams, 1,
				"", int64(testCase.startOffset), int64(testCase.endOffset), 0, 0)
			pieceInfos, err := SplitToSegmentPieceInfos(task, &gfsppieceop.GfSpPieceOp{})
			if !testCase.isErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}
			assert.Equal(t, testCase.wantedPieceCount, len(pieceInfos))
			realLength := uint64(0)
			for _, p := range pieceInfos {
				realLength += p.Length
			}
			assert.Equal(t, testCase.endOffset-testCase.startOffset+1, realLength)
		})
	}
}
