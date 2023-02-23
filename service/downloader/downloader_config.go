package downloader

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type DownloaderConfig struct {
	Address          string
	PieceStoreConfig *storage.PieceStoreConfig
}

var DefaultDownloaderConfig = &DownloaderConfig{
	Address:          model.DefaultDownloaderAddress,
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
}
