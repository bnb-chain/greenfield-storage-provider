package client

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

// PieceStoreAPI provides an interface to enable mocking the
// StoreClient's API operation. This makes unit test to test your code easier.
//
//go:generate mockgen -source=./piece_store_client.go -destination=./mock/piece_store_mock.go -package=mock
type PieceStoreAPI interface {
	GetPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error)
	PutPiece(key string, value []byte) error
}

type StoreClient struct {
	ps *piece.PieceStore
}

const (
	getPieceMethodName = "getPiece"
	putPieceMethodName = "putPiece"
)

func NewStoreClient(pieceConfig *storage.PieceStoreConfig) (*StoreClient, error) {
	ps, err := piece.NewPieceStore(pieceConfig)
	if err != nil {
		return nil, err
	}
	return &StoreClient{ps: ps}, nil
}

// GetPiece gets piece data from piece store
func (client *StoreClient) GetPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error) {
	startTime := time.Now()
	defer func() {
		observer := metrics.PieceStoreTimeHistogram.WithLabelValues(getPieceMethodName)
		observer.Observe(time.Since(startTime).Seconds())
		metrics.PieceStoreRequestTotal.WithLabelValues(getPieceMethodName)
	}()

	rc, err := client.ps.Get(ctx, key, offset, limit)
	if err != nil {
		log.Errorw("failed to get piece data from piece store", "error", err)
		return nil, err
	}
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, rc)
	if err != nil {
		log.Errorw("failed to copy data", "error", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

// PutPiece puts piece to piece store
func (client *StoreClient) PutPiece(key string, value []byte) error {
	startTime := time.Now()
	defer func() {
		observer := metrics.PieceStoreTimeHistogram.WithLabelValues(putPieceMethodName)
		observer.Observe(time.Since(startTime).Seconds())
		metrics.PieceStoreRequestTotal.WithLabelValues(putPieceMethodName)
	}()

	return client.ps.Put(context.Background(), key, bytes.NewReader(value))
}
