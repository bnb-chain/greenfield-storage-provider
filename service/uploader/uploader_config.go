package uploader

import (
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
