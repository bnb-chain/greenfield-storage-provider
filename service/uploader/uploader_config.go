package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

// UploaderConfig defines Uploader service config
type UploaderConfig struct {
	GRPCAddress         string
	SignerGrpcAddress   string
	TaskNodeGrpcAddress string // TODO: will be deleted.
	ManagerGrpcAddress  string
	SpDBConfig          *config.SQLDBConfig
	PieceStoreConfig    *storage.PieceStoreConfig
}
