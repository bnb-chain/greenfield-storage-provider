package syncer

import (
	"bytes"
	"context"

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

func (sc *storeClient) putPiece(key string, value []byte) error {
	return sc.ps.Put(context.Background(), key, bytes.NewReader(value))
}
