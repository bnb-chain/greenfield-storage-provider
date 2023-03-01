package downloader

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type DownloaderConfig struct {
	GrpcAddress      string
	ChainConfig      *gnfd.GreenfieldChainConfig
	PieceStoreConfig *storage.PieceStoreConfig
}

var DefaultDownloaderConfig = &DownloaderConfig{
	GrpcAddress:      model.DowmloaderGrpcAddress,
	ChainConfig:      gnfd.DefaultGreenfieldChainConfig,
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
}
