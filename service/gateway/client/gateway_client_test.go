package client

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pieceDataReader(t *testing.T) {
	pieceData := make([][]byte, 3)
	pieceData[0] = []byte{'A', 'B'}
	pieceData[1] = []byte{'C', 'D', 'E'}
	pieceData[2] = []byte{'F'}

	{
		// read all data
		pieceDataReader, err := NewPieceDataReader(pieceData)
		require.NoError(t, err)
		readBuf := make([]byte, 10)
		readN, err := pieceDataReader.Read(readBuf)
		assert.Equal(t, 6, readN)
		assert.Equal(t, "ABCDEF", string(readBuf[0:readN]))

		require.NoError(t, err)
		readN, err = pieceDataReader.Read(readBuf)
		assert.Equal(t, 0, readN)
		assert.Equal(t, io.EOF, err)
	}
	{
		// read part data
		pieceDataReader, err := NewPieceDataReader(pieceData)
		require.NoError(t, err)
		readBufA := make([]byte, 1)
		readN, err := pieceDataReader.Read(readBufA)
		assert.Equal(t, 1, readN)
		require.NoError(t, err)
		assert.Equal(t, byte('A'), readBufA[0])

		readBufBC := make([]byte, 2)
		readN, err = pieceDataReader.Read(readBufBC)
		assert.Equal(t, 2, readN)
		require.NoError(t, err)
		assert.Equal(t, byte('B'), readBufBC[0])
		assert.Equal(t, byte('C'), readBufBC[1])

		readBufDEF := make([]byte, 3)
		readN, err = pieceDataReader.Read(readBufDEF)
		assert.Equal(t, 3, readN)
		require.NoError(t, err)
		assert.Equal(t, byte('D'), readBufDEF[0])
		assert.Equal(t, byte('E'), readBufDEF[1])
		assert.Equal(t, byte('F'), readBufDEF[2])

		readN, err = pieceDataReader.Read(readBufDEF)
		assert.Equal(t, 0, readN)
		assert.Equal(t, io.EOF, err)
	}
}
