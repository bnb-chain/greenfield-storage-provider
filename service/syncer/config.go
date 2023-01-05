package syncer

import "github.com/bnb-chain/inscription-storage-provider/store/piecestore/piece"

type SyncerConfig struct {
	Address     string
	PieceConfig *piece.PieceStoreConfig
}
