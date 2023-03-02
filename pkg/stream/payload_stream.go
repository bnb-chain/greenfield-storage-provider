package stream

import (
	"errors"
	"io"
	"sync/atomic"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const StreamResultSize = 10

type SegmentEntry struct {
	objectId       uint64
	replicaIdx     uint32
	segmentIdx     uint32
	redundancyType storagetypes.RedundancyType
	segmentData    []byte
	err            error
}

func (entry SegmentEntry) ID() uint64 {
	return entry.objectId
}

func (entry SegmentEntry) Key() string {
	if entry.redundancyType == storagetypes.REDUNDANCY_EC_TYPE {
		return piecestore.EncodeECPieceKey(entry.objectId, entry.replicaIdx, entry.segmentIdx)
	}
	return piecestore.EncodeSegmentPieceKey(entry.objectId, entry.segmentIdx)
}

func (entry SegmentEntry) Data() []byte {
	return entry.segmentData
}

func (entry SegmentEntry) Error() error {
	return entry.err
}

// PayloadStream implement a one-way data flow, writes bytes of any size
// read the fixed data size with payload metadata
type PayloadStream struct {
	objectId       uint64
	replicaIdx     uint32
	segmentIdx     uint32
	segmentSize    uint64
	redundancyType storagetypes.RedundancyType
	entryCh        chan *SegmentEntry
	init           atomic.Bool
	close          atomic.Bool

	pRead  *io.PipeReader
	pWrite *io.PipeWriter
}

// NewAsyncPayloadStream return an instance of PayloadStream, and start async read stream
func NewAsyncPayloadStream() *PayloadStream {
	stream := &PayloadStream{
		entryCh: make(chan *SegmentEntry, StreamResultSize),
	}
	stream.pRead, stream.pWrite = io.Pipe()
	return stream
}

// InitAsyncPayloadStream only be called once, init the payload meta data
// must be called before write or read stream
func (stream *PayloadStream) InitAsyncPayloadStream(oId uint64, rIdx uint32, segSize uint64,
	redundancyType storagetypes.RedundancyType) error {
	if stream.init.Load() {
		return nil
	}
	stream.init.Store(true)
	stream.objectId = oId
	stream.replicaIdx = rIdx
	stream.segmentSize = segSize
	stream.redundancyType = redundancyType
	go stream.readStream()
	return nil
}

// StreamWrite writes data with the bytes of any size
func (stream *PayloadStream) StreamWrite(data []byte) (n int, err error) {
	if !stream.init.Load() {
		return 0, errors.New("payload stream uninitialized")
	}
	if stream.close.Load() {
		return 0, errors.New("payload stream has been closed")
	}
	return stream.pWrite.Write(data)
}

// StreamClose close write stream without error
func (stream *PayloadStream) StreamClose() error {
	if stream.close.Load() {
		return nil
	}
	stream.close.Store(true)
	return stream.pWrite.Close()
}

// StreamCloseWithError close write stream with error
func (stream *PayloadStream) StreamCloseWithError(err error) error {
	if !stream.init.Load() {
		return errors.New("payload stream is uninitialized")
	}
	if stream.close.Load() {
		return nil
	}
	stream.close.Store(true)
	return stream.pWrite.CloseWithError(err)
}

// AsyncStreamRead return a channel that receive the payload and it's metadata
func (stream *PayloadStream) AsyncStreamRead() <-chan *SegmentEntry {
	return stream.entryCh
}

// Close write and read stream by the safe way
func (stream *PayloadStream) Close() {
	close(stream.entryCh)
	stream.StreamClose()
}

func (stream *PayloadStream) readStream() {
	var (
		count    uint32
		readSize uint32
	)
	for {
		entry := &SegmentEntry{
			objectId:       stream.objectId,
			replicaIdx:     stream.replicaIdx,
			segmentIdx:     count,
			redundancyType: stream.redundancyType,
		}
		data := make([]byte, stream.segmentSize)
		n, err := stream.readN(data)
		if err != nil && err != io.EOF {
			log.Errorw("failed to read payload stream", "err", err)
			entry.err = err
			stream.entryCh <- entry
			return
		}
		data = data[0:n]
		if err == io.EOF && n == 0 {
			entry.err = err
			stream.entryCh <- entry
			log.Debugw("payload stream on closed", "object_id", stream.objectId)
			return
		}
		entry.segmentData = data
		stream.entryCh <- entry
		count++
		readSize = readSize + uint32(n)
		log.Debugw("payload stream has read", "read_total_size", readSize, "object_id", stream.objectId, "segment_count:", count-1)
	}
}

func (stream *PayloadStream) readN(b []byte) (int, error) {
	var (
		err       error
		totalSize int
		curSize   int
		size      int
	)

	totalSize = len(b)
	for {
		size, err = stream.pRead.Read(b[curSize:])
		curSize = curSize + size
		if err != nil || curSize == totalSize {
			break
		}
	}
	return curSize, err
}
