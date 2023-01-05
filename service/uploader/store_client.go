package uploader

import (
	"bytes"
	"context"

	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// storeClient is wrapper of pieceStore.
type storeClient struct {
	ps *piece.PieceStore
}

// newStoreClient return a piece store wrapper.
func newStoreClient(config *storage.PieceStoreConfig) (*storeClient, error) {
	ps, err := piece.NewPieceStore(config)
	if err != nil {
		log.Warnw("failed to new piece store", "err", err)
		return nil, err
	}
	return &storeClient{ps: ps}, nil
}

// putPiece is used to put KV in piece store.
func (sc *storeClient) putPiece(key string, value []byte) error {
	return sc.ps.Put(context.Background(), key, bytes.NewReader(value))
}
