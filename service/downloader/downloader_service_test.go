package downloader

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
)

func Test_Spilt_Segment_info(t *testing.T) {
	pieces, err := DownloadPieceInfo(123, 50*1024*1024, 0, 100)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(pieces))
	assert.Equal(t, piecestore.EncodeSegmentPieceKey(123, uint32(0)), pieces[0].pieceKey)

	pieces, err = DownloadPieceInfo(123, 50*1024*1024, 16*1024*1024, 33*1024*1024)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(pieces))
	assert.Equal(t, piecestore.EncodeSegmentPieceKey(123, uint32(1)), pieces[0].pieceKey)
	assert.Equal(t, piecestore.EncodeSegmentPieceKey(123, uint32(2)), pieces[1].pieceKey)

	pieces, err = DownloadPieceInfo(123, 50*1024*1024, 16*1024*1024, 32*1024*1024-1)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(pieces))
	assert.Equal(t, piecestore.EncodeSegmentPieceKey(123, uint32(1)), pieces[0].pieceKey)

	pieces, err = DownloadPieceInfo(123, 50*1024*1024, 16*1024*1024-1, 32*1024*1024-1)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(pieces))
	assert.Equal(t, piecestore.EncodeSegmentPieceKey(123, uint32(0)), pieces[0].pieceKey)
	assert.Equal(t, 1, int(pieces[0].length))
	assert.Equal(t, piecestore.EncodeSegmentPieceKey(123, uint32(1)), pieces[1].pieceKey)
	assert.Equal(t, 16*1024*1024, int(pieces[1].length))

	pieces, err = DownloadPieceInfo(123, 50*1024*1024, 32*1024*1024+1, 50*1024*1024-1)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(pieces))
	assert.Equal(t, piecestore.EncodeSegmentPieceKey(123, uint32(2)), pieces[0].pieceKey)
	assert.Equal(t, 16*1024*1024-1, int(pieces[0].length))
	assert.Equal(t, piecestore.EncodeSegmentPieceKey(123, uint32(3)), pieces[1].pieceKey)
	assert.Equal(t, 2*1024*1024, int(pieces[1].length))
}
