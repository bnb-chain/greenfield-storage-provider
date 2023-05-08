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
	name string
	ps   *piece.PieceStore
}

func NewStoreClient(pieceConfig *storage.PieceStoreConfig) (*StoreClient, error) {
	ps, err := piece.NewPieceStore(pieceConfig)
	if err != nil {
		return nil, err
	}
	return &StoreClient{ps: ps, name: pieceConfig.Store.Storage}, nil
}

// GetPiece gets piece data from piece store.
func (client *StoreClient) GetPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error) {
	startTime := time.Now()
	defer func() {
		metrics.GetPieceTimeHistogram.WithLabelValues(client.name).Observe(
			time.Since(startTime).Seconds())
		metrics.GetPieceTimeCounter.WithLabelValues(client.name).Inc()
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

// PutPiece puts piece to piece store.
func (client *StoreClient) PutPiece(key string, value []byte) error {
	startTime := time.Now()
	defer func() {
		metrics.PutPieceTimeHistogram.WithLabelValues(client.name).Observe(
			time.Since(startTime).Seconds())
		metrics.PutPieceTimeCounter.WithLabelValues(client.name).Inc()
		metrics.PieceWriteSizeGauge.WithLabelValues(client.name).Add(float64(len(value)))
	}()
	return client.ps.Put(context.Background(), key, bytes.NewReader(value))
}

// DeletePiece deletes piece from piece store.
func (client *StoreClient) DeletePiece(key string) error {
	startTime := time.Now()
	defer func() {
		metrics.DeletePieceTimeHistogram.WithLabelValues(client.name).Observe(
			time.Since(startTime).Seconds())
		metrics.DeletePieceTimeCounter.WithLabelValues(client.name).Inc()
	}()
	return client.ps.Delete(context.Background(), key)
}
