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

const (
	// PieceStoreSuccessPut defines the metrics label of successfully put piece data
	PieceStoreSuccessPut = "put_piece_store_success"
	// PieceStoreFailurePut defines the metrics label of unsuccessfully put piece data
	PieceStoreFailurePut = "put_piece_store_failure"
	// PieceStoreSuccessGet defines the metrics label of successfully get piece data
	PieceStoreSuccessGet = "get_piece_store_success"
	// PieceStoreFailureGet defines the metrics label of unsuccessfully get piece data
	PieceStoreFailureGet = "get_piece_store_failure"
	// PieceStoreSuccessDel defines the metrics label of successfully delete piece data
	PieceStoreSuccessDel = "del_piece_store_success"
	// PieceStoreFailureDel defines the metrics label of unsuccessfully delete piece data
	PieceStoreFailureDel = "del_piece_store_failure"
)

// PieceStoreAPI provides an interface to enable mocking the
// StoreClient's API operation. This makes unit test to test your code easier.
//
//go:generate mockgen -source=./piece_store_client.go -destination=./piece_store_client_mock.go -package=client
type PieceStoreAPI interface {
	GetPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error)
	PutPiece(key string, value []byte) error
	DeletePiece(ctx context.Context, key string) error
}

var _ corepiecestore.PieceStore = &StoreClient{}

type StoreClient struct {
	name string
	ps   piece.PieceAPI
}

func NewStoreClient(pieceConfig *storage.PieceStoreConfig) (*StoreClient, error) {
	ps, err := piece.NewPieceStore(pieceConfig)
	if err != nil {
		log.Errorw("failed to new piece store", "error", err)
		return nil, err
	}
	return &StoreClient{ps: ps, name: pieceConfig.Store.Storage}, nil
}

// GetPiece gets piece data from piece store.
func (client *StoreClient) GetPiece(ctx context.Context, key string, offset, limit int64) (data []byte, err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.PieceStoreCounter.WithLabelValues(PieceStoreFailureGet).Inc()
			metrics.PieceStoreTime.WithLabelValues(PieceStoreFailureGet).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.PieceStoreCounter.WithLabelValues(PieceStoreSuccessGet).Inc()
		metrics.PieceStoreTime.WithLabelValues(PieceStoreSuccessGet).Observe(
			time.Since(startTime).Seconds())
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
		if err != nil {
			metrics.PieceStoreCounter.WithLabelValues(PieceStoreFailurePut).Inc()
			metrics.PieceStoreTime.WithLabelValues(PieceStoreFailurePut).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.PieceStoreCounter.WithLabelValues(PieceStoreSuccessPut).Inc()
		metrics.PieceStoreTime.WithLabelValues(PieceStoreSuccessPut).Observe(
			time.Since(startTime).Seconds())
		metrics.PieceStoreUsageAmountGauge.WithLabelValues(PieceStoreSuccessPut).Add(float64(len(value)))
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
		if err != nil {
			metrics.PieceStoreCounter.WithLabelValues(PieceStoreFailureDel).Inc()
			metrics.PieceStoreTime.WithLabelValues(PieceStoreFailureDel).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.PieceStoreCounter.WithLabelValues(PieceStoreSuccessDel).Inc()
		metrics.PieceStoreTime.WithLabelValues(PieceStoreSuccessDel).Observe(
			time.Since(startTime).Seconds())
		metrics.PieceStoreUsageAmountGauge.WithLabelValues(PieceStoreSuccessDel).Add(0 - float64(valSize))
	}()
	val, err := client.GetPiece(ctx, key, 0, -1)
	if err != nil {
		return err
	}
	valSize = len(val)
	err = client.ps.Delete(ctx, key)
	return err
}
