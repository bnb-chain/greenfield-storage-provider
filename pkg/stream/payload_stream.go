package stream

import (
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// PieceEntry is used to stream produce piece data.
type PieceEntry struct {
	objectID          uint64
	redundancyType    storagetypes.RedundancyType
	segmentPieceIndex uint32
	ecPieceIndex      uint32 // meaningful iff redundancyType == REDUNDANCY_EC_TYPE
	pieceData         []byte
	err               error
}

// ObjectID returns piece's object id
func (entry *PieceEntry) ObjectID() uint64 {
	return entry.objectID
}

// PieceKey returns piece key, may be ECPieceKey/SegmentPieceKey
func (entry *PieceEntry) PieceKey() string {
	if entry.redundancyType == storagetypes.REDUNDANCY_EC_TYPE {
		return piecestore.EncodeECPieceKey(entry.objectID, entry.segmentPieceIndex, entry.ecPieceIndex)
	}
	return piecestore.EncodeSegmentPieceKey(entry.objectID, entry.segmentPieceIndex)
}

// Data returns piece data
func (entry *PieceEntry) Data() []byte {
	return entry.pieceData
}

// Error returns stream read error
func (entry *PieceEntry) Error() error {
	return entry.err
}

// PayloadStream implements a one-way data flow, writes bytes of any size
// read the fixed data size with payload metadata
type PayloadStream struct {
	objectID       uint64
	redundancyType storagetypes.RedundancyType
	pieceSize      uint64 // pieceSize is used to split
	ecPieceIndex   uint32
	entryCh        chan *PieceEntry
	init           atomic.Value
	close          atomic.Value

	pRead  *io.PipeReader
	pWrite *io.PipeWriter
}

// NewAsyncPayloadStream returns an instance of PayloadStream, and start async read stream
// TODO:: implement the SyncPayloadStream in the future base on requirements
func NewAsyncPayloadStream() *PayloadStream {
	stream := &PayloadStream{
		entryCh: make(chan *PieceEntry, 10),
	}
	stream.pRead, stream.pWrite = io.Pipe()
	return stream
}

// InitAsyncPayloadStream only be called once, init the payload metadata
// must be called before write or read stream
func (stream *PayloadStream) InitAsyncPayloadStream(objectID uint64, redundancyType storagetypes.RedundancyType,
	pieceSize uint64, ecPieceIndex uint32) error {
	if stream.init.Load() == true {
		return nil
	}
	stream.init.Store(true)
	stream.objectID = objectID
	stream.redundancyType = redundancyType
	stream.pieceSize = pieceSize
	stream.ecPieceIndex = ecPieceIndex
	go stream.readStream()
	return nil
}

// StreamWrite writes data with the bytes of any size
func (stream *PayloadStream) StreamWrite(data []byte) (n int, err error) {
	if stream.init.Load() == nil {
		return 0, errors.New("payload stream uninitialized")
	}
	if stream.close.Load() == true {
		return 0, errors.New("payload stream has been closed")
	}
	return stream.pWrite.Write(data)
}

// StreamClose closes write stream without error
func (stream *PayloadStream) StreamClose() error {
	if stream.close.Load() == true {
		return nil
	}
	stream.close.Store(true)
	return stream.pWrite.Close()
}

// StreamCloseWithError closes write stream with error
func (stream *PayloadStream) StreamCloseWithError(err error) error {
	if stream.init.Load() == nil {
		return errors.New("payload stream is uninitialized")
	}
	if stream.close.Load() == true {
		return nil
	}
	stream.close.Store(true)
	return stream.pWrite.CloseWithError(err)
}

// AsyncStreamRead returns a channel which receives PieceEntry
func (stream *PayloadStream) AsyncStreamRead() <-chan *PieceEntry {
	return stream.entryCh
}

// Close writes and reads stream by the safe way
func (stream *PayloadStream) Close() {
	stream.StreamClose()
}

func (stream *PayloadStream) readStream() {
	var (
		segmentPieceIdx uint32
		totalReadSize   int
	)
	for {
		entry := &PieceEntry{
			objectID:          stream.objectID,
			redundancyType:    stream.redundancyType,
			segmentPieceIndex: segmentPieceIdx,
			ecPieceIndex:      stream.ecPieceIndex,
		}
		data := make([]byte, stream.pieceSize)
		n, err := stream.readN(data)
		if err != nil && err != io.EOF {
			log.Errorw("failed to read payload stream", "error", err)
			entry.err = err
			stream.entryCh <- entry
			return
		}
		data = data[0:n]
		if err == io.EOF && n == 0 {
			close(stream.entryCh)
			return
		}
		entry.pieceData = data
		stream.entryCh <- entry
		segmentPieceIdx++
		totalReadSize += n
		log.Debugw("succeed to produce stream piece", "total_read_size", totalReadSize, "object_id", stream.objectID,
			"segment_piece_index", segmentPieceIdx-1, "cur_read_size", n, "error", err)
		if err == io.EOF {
			close(stream.entryCh)
			return
		}
	}
}

// readN is used to read len(b) data or io.EOF.
func (stream *PayloadStream) readN(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, fmt.Errorf("failed to read due to invalid args")
	}

	var (
		totalReadLen int
		curReadLen   int
		err          error
	)

	for {
		curReadLen, err = stream.pRead.Read(b[totalReadLen:])
		totalReadLen += curReadLen
		if err != nil || totalReadLen == len(b) {
			return totalReadLen, err
		}
	}
}
