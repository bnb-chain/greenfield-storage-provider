package stonenode

import "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"

type StoneNodeConfig struct {
	Address                string
	StoneHubServiceAddress string
	SyncerServiceAddress   string
	StorageProvider        string
	PieceConfig            *storage.PieceStoreConfig
	StoneJobLimit          int64
}

var DefaultStoneNodeConfig = &StoneNodeConfig{
	Address:                "127.0.0.1:5325",
	StoneHubServiceAddress: "127.0.0.1:5323",
	SyncerServiceAddress:   "127.0.0.1:5324",
	StorageProvider:        "bnb-sp",
	PieceConfig:            storage.DefaultPieceStoreConfig,
	StoneJobLimit:          64,
}
