package client

import (
	"bytes"
	"context"
	"io"

	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type StoreClient struct {
	ps *piece.PieceStore
}

func NewStoreClient(pieceConfig *storage.PieceStoreConfig) (*StoreClient, error) {
	ps, err := piece.NewPieceStore(pieceConfig)
	if err != nil {
		return nil, err
	}
	return &StoreClient{ps: ps}, nil
}

func (client *StoreClient) GetPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error) {
	rc, err := client.ps.Get(ctx, key, offset, limit)
	if err != nil {
		log.Errorw("stone node service invoke PieceStore Get failed", "error", err)
		return nil, err
	}
	data, err := io.ReadAll(rc)
	if err != nil {
		log.Errorw("stone node service invoke io.ReadAll failed", "error", err)
		return nil, err
	}
	return data, nil
}

func (client *StoreClient) PutPiece(key string, value []byte) error {
	return client.ps.Put(context.Background(), key, bytes.NewReader(value))
}
