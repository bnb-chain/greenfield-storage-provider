package greenfield

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
		GreenfieldAddresses: []string{"localhost:9090"},
		TendermintAddresses: []string{"http://localhost:26750"},
	}},
}
