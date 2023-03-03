package greenfield

import "github.com/bnb-chain/greenfield-storage-provider/model"

type NodeConfig struct {
	GreenfieldAddresses []string
	TendermintAddresses []string
}

type GreenfieldChainConfig struct {
	ChainID  string
	NodeAddr []*NodeConfig
}

var DefaultGreenfieldChainConfig = &GreenfieldChainConfig{
	ChainID: "greenfield_9000-1741",
	NodeAddr: []*NodeConfig{{
		GreenfieldAddresses: []string{model.GreenfieldAddress},
		TendermintAddresses: []string{model.TendermintAddress},
	}},
}
