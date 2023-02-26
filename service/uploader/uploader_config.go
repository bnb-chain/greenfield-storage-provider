package uploader

import (
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type UploaderConfig struct {
	StorageProvider        string
	Address                string
	StoneHubServiceAddress string
	PieceStoreConfig       *storage.PieceStoreConfig
	ChainConfig            *gnfd.GreenfieldChainConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	StorageProvider:        "bnb-sp",
	Address:                "127.0.0.1:5311",
	StoneHubServiceAddress: "127.0.0.1:5323",
	PieceStoreConfig:       storage.DefaultPieceStoreConfig,
	ChainConfig:            gnfd.DefaultGreenfieldChainConfig,
}
