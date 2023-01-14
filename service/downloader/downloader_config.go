package downloader

import "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"

type DownloaderConfig struct {
	Address          string
	PieceStoreConfig *storage.PieceStoreConfig
}

var DefaultDownloaderConfig = &DownloaderConfig{
	Address:          "127.0.0.1:5523",
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
}
