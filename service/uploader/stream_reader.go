package uploader

import (
	"io"
	"sync"

	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// streamReader is a wrapper of grpc stream request.
type streamReader struct {
	pr     *io.PipeReader
	pw     *io.PipeWriter
	txHash []byte
}

// newStreamReader is used to stream read UploaderService_UploadPayloadServer.
func newStreamReader(stream pbService.UploaderService_UploadPayloadServer, ch chan []byte) *streamReader {
	var sr = &streamReader{}
	sr.pr, sr.pw = io.Pipe()
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				if sr.txHash == nil {
					sr.txHash = res.TxHash
					ch <- res.TxHash
				}
				sr.pw.Close()
				return
			}
			if err != nil {
				if sr.txHash == nil {
					close(ch)
				}
				sr.pw.CloseWithError(err)
				return
			}
			if sr.txHash == nil {
				sr.txHash = res.TxHash
				ch <- res.TxHash
			}
			sr.pw.Write(res.PayloadData)
		}
	}()
	return sr
}

// Read is alike to ReadFull, if error is nil, b is full.
func (sr *streamReader) Read(b []byte) (int, error) {
	var (
		err       error
		totalSize int
		curSize   int
		size      int
	)

	totalSize = len(b)
	for {
		size, err = sr.pr.Read(b[curSize:])
		curSize = curSize + size
		if err != nil || curSize == totalSize {
			break
		}
	}
	return curSize, err
}

type SegmentContext struct {
	Index     uint32
	PieceData []byte
}

// splitSegment is used to split streamReader's data to []bytes by chan.
func (sr *streamReader) splitSegment(segmentSize uint32, ch chan *SegmentContext, wg *sync.WaitGroup) error {
	var (
		err   error
		readN int
		index uint32
		size  int
	)

	for {
		pieceData := make([]byte, segmentSize)
		readN, err = sr.Read(pieceData)
		if err != nil && err != io.EOF {
			log.Warnw("failed to stream read", "err", err)
			close(ch)
			return err
		}
		if err == io.EOF && readN == 0 {
			break
		}
		wg.Add(1)
		segmentCtx := &SegmentContext{
			Index:     index,
			PieceData: pieceData[:readN],
		}
		ch <- segmentCtx
		index = index + 1
		size = size + readN
	}

	log.Info("uploader total size:", size, " index:", index)
	close(ch)
	return nil
}
