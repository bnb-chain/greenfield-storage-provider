package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type UploaderConfig struct {
	GrpAddress           string
	SignerGrpcAddress    string
	StoneNodeGrpcAddress string
	SPDBConfig           *config.SQLDBConfig
	PieceStoreConfig     *storage.PieceStoreConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	GrpAddress:           model.UploaderGrpcAddress,
	SignerGrpcAddress:    model.SyncerGrpcAddress,
	StoneNodeGrpcAddress: model.StoneNodeGrpcAddress,
	SPDBConfig:           config.DefaultSQLDBConfig,
	PieceStoreConfig:     storage.DefaultPieceStoreConfig,
}
