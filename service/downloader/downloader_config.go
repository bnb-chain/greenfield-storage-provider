package downloader

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
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

var DefaultDownloaderConfig = &DownloaderConfig{
	GRPCAddress:      model.DownloaderGRPCAddress,
	SpDBConfig:       config.DefaultSQLDBConfig,
	ChainConfig:      gnfd.DefaultGreenfieldChainConfig,
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
}
