package downloader

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

// DownloaderConfig defines downloader service config
type DownloaderConfig struct {
	GRPCAddress      string
	SpDBConfig       *config.SQLDBConfig
	PieceStoreConfig *storage.PieceStoreConfig
}
