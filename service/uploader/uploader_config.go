package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type UploaderConfig struct {
	GRPCAddress          string
	SignerGrpcAddress    string
	StoneNodeGrpcAddress string
	SpDBConfig           *config.SQLDBConfig
	PieceStoreConfig     *storage.PieceStoreConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	GRPCAddress:          model.UploaderGRPCAddress,
	SignerGrpcAddress:    model.SyncerGRPCAddress,
	StoneNodeGrpcAddress: model.StoneNodeGRPCAddress,
	SpDBConfig:           config.DefaultSQLDBConfig,
	PieceStoreConfig:     storage.DefaultPieceStoreConfig,
}
