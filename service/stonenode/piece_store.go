package stonenode

import (
	"context"
	"io"

	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type storeClient struct {
	ps *piece.PieceStore
}

func newStoreClient(pieceConfig *storage.PieceStoreConfig) (*storeClient, error) {
	ps, err := piece.NewPieceStore(pieceConfig)
	if err != nil {
		return nil, err
	}
	return &storeClient{ps: ps}, nil
}

func (sc *storeClient) getPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error) {
	rc, err := sc.ps.Get(ctx, key, offset, limit)
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
