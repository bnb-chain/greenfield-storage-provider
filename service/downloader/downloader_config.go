package downloader

import "github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"

type DownloaderConfig struct {
	Address     string
	PieceConfig *storage.PieceStoreConfig
}

var DefaultDownloaderConfig = &DownloaderConfig{
	Address:     "127.0.0.1:5523",
	PieceConfig: storage.DefaultPieceStoreConfig,
}
