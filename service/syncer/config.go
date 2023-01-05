package syncer

import "github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"

type SyncerConfig struct {
	Address         string
	StorageProvider string
	PieceConfig     *storage.PieceStoreConfig
}

var DefaultSyncerConfig = &SyncerConfig{
	Address:         "127.0.0.1:5324",
	StorageProvider: "bnb-sp",
	PieceConfig:     storage.DefaultPieceStoreConfig,
}
