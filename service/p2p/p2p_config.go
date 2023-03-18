package p2p

import (
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

type P2PConfig struct {
	SpOperatorAddress string
	GRPCAddress       string
	SignerGRPCAddress string
	SpDBConfig        *config.SQLDBConfig
	P2PConfig         *p2p.NodeConfig
}
