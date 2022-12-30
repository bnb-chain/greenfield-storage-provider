package stonenode

import (
	"context"
	"io"

	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type storeClient struct {
	ps *piece.PieceStore
}

func newStoreClient(configFile string) (*storeClient, error) {
	ps, err := piece.NewPieceStore(configFile)
	if err != nil {
		log.Errorw("SyncerService NewPieceStore failed", "error", err)
		return nil, err
	}
	return &storeClient{ps: ps}, nil
}

func (sc *storeClient) getPiece(key string, offset, limit int64) ([]byte, error) {
	rc, err := sc.ps.Get(context.Background(), key, offset, limit)
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
