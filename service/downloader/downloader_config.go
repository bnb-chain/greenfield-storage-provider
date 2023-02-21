package downloader

import (
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type DownloaderConfig struct {
	Address          string
	PieceStoreConfig *storage.PieceStoreConfig
	ChainConfig      *gnfd.GreenfieldChainConfig
}

var DefaultDownloaderConfig = &DownloaderConfig{
	Address:          "127.0.0.1:9233",
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
	ChainConfig:      gnfd.DefaultGreenfieldChainConfig,
}
