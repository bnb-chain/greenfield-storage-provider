package downloader

import (
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type DownloaderConfig struct {
	GRPCAddress      string
	SpDBConfig       *config.SQLDBConfig
	PieceStoreConfig *storage.PieceStoreConfig
	ChainConfig      *gnfd.GreenfieldChainConfig
}
