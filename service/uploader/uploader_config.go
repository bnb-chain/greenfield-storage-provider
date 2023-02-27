package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/store"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type UploaderConfig struct {
	Address          string
	SignerAddress    string
	StoneNodeAddress string
	DbConfig         *store.SqlDBConfig
	PieceStoreConfig *storage.PieceStoreConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	Address:          "localhost:5311",
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
	DbConfig:         store.DefaultSqlDBConfig,
}
