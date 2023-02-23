package p2p

import (
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/node"
)

var (
	DefaultNodeKeyPath = "node_key.json"
)

type P2PServiceConfig struct {
	GrpcAddress      string
	P2PListenAddress string
	Whitelist        string
	NodeKeyPath      string
}

func (config P2PServiceConfig) makeP2pConfig() (*node.NodeConfig, error) {
	nodeCfg := node.DefaultNodeConfig()
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	nodeCfg.RootDir = pwd
	nodeCfg.NodeKey = config.NodeKeyPath
	nodeCfg.P2P.ListenAddress = config.P2PListenAddress
	nodeCfg.P2P.PersistentPeers = config.Whitelist
	return &nodeCfg, nil
}

var DefaultP2PServiceConfig = &P2PServiceConfig{
	GrpcAddress:      model.DefaultP2PServiceAddress,
	P2PListenAddress: model.DefaultP2PListenAddress,
	NodeKeyPath:      DefaultNodeKeyPath,
}
