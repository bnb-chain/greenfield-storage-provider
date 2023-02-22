package p2p

import (
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/node"
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
	GrpcAddress:      "127.0.0.1:9733",
	P2PListenAddress: "127.0.0.1:21303",
	NodeKeyPath:      "node_key.json",
}
