package client

import (
	"bytes"
	"context"
	"io"
	"time"

	corepiecestore "github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
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

var _ corepiecestore.PieceStore = &StoreClient{}

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
		metrics.GetPieceTotalNumberCounter.WithLabelValues(client.name).Inc()
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
func (client *StoreClient) PutPiece(ctx context.Context, key string, value []byte) error {
	var (
		startTime = time.Now()
		err       error
	)
	defer func() {
		metrics.PutPieceTimeHistogram.WithLabelValues(client.name).Observe(
			time.Since(startTime).Seconds())
		metrics.PutPieceTotalNumberCounter.WithLabelValues(client.name).Inc()
		if err == nil {
			metrics.PieceUsageAmountGauge.WithLabelValues(client.name).Add(float64(len(value)))
		}
	}()
	err = client.ps.Put(ctx, key, bytes.NewReader(value))
	return err
}

// DeletePiece deletes piece from piece store.
func (client *StoreClient) DeletePiece(ctx context.Context, key string) error {
	var (
		startTime = time.Now()
		err       error
		valSize   int
	)
	defer func() {
		metrics.DeletePieceTimeHistogram.WithLabelValues(client.name).Observe(
			time.Since(startTime).Seconds())
		metrics.DeletePieceTotalNumberCounter.WithLabelValues(client.name).Inc()
		if err == nil {
			metrics.PieceUsageAmountGauge.WithLabelValues(client.name).Add(0 - float64(valSize))
		}
	}()
	val, err := client.GetPiece(ctx, key, 0, -1)
	if err != nil {
		return err
	}
	valSize = len(val)
	err = client.ps.Delete(ctx, key)
	return err
}
