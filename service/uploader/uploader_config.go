package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type UploaderConfig struct {
	GrpAddress           string
	SignerGrpcAddress    string
	StoneNodeGrpcAddress string
	DbConfig             *store.SqlDBConfig
	PieceStoreConfig     *storage.PieceStoreConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	GrpAddress:           model.UploaderGrpcAddress,
	SignerGrpcAddress:    model.SyncerGrpcAddress,
	StoneNodeGrpcAddress: model.StoneNodeGrpcAddress,
	DbConfig:             store.DefaultSqlDBConfig,
	PieceStoreConfig:     storage.DefaultPieceStoreConfig,
}
