package receiver

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type ReceiverConfig struct {
	SpOperatorAddress string
	GRPCAddress       string
	SignerGRPCAddress string
	SpDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
}
