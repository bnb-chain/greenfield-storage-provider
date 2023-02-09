package stonenode

import "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"

type StoneNodeConfig struct {
	Address                string
	StoneHubServiceAddress string
	SyncerServiceAddress   []string
	StorageProvider        string
	PieceConfig            *storage.PieceStoreConfig
	StoneJobLimit          int64
}

var DefaultStoneNodeConfig = &StoneNodeConfig{
	Address:                "127.0.0.1:9433",
	StoneHubServiceAddress: "127.0.0.1:9333",
	SyncerServiceAddress:   []string{"127.0.0.1:9533", "127.0.0.1:9543", "127.0.0.1:9553", "127.0.0.1:9563", "127.0.0.1:9573", "127.0.0.1:9583"},
	StorageProvider:        "bnb-sp",
	PieceConfig:            storage.DefaultPieceStoreConfig,
	StoneJobLimit:          64,
}
