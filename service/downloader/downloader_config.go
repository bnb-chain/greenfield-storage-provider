package downloader

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type DownloaderConfig struct {
	GrpcAddress      string
	SPDBConfig       *config.SQLDBConfig
	PieceStoreConfig *storage.PieceStoreConfig
	ChainConfig      *gnfd.GreenfieldChainConfig
}

var DefaultDownloaderConfig = &DownloaderConfig{
	GrpcAddress:      model.DownloaderGrpcAddress,
	SPDBConfig:       config.DefaultSQLDBConfig,
	ChainConfig:      gnfd.DefaultGreenfieldChainConfig,
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
}
