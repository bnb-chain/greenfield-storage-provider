package greenfield

type NodeConfig struct {
	GreenfieldAddrs []string
	TendermintAddrs []string
}

type GreenfieldChainConfig struct {
	ChainID  string
	NodeAddr []*NodeConfig
}

var DefaultGreenfieldChainConfig = &GreenfieldChainConfig{
	ChainID: "greenfield_9000-1741",
	NodeAddr: []*NodeConfig{&NodeConfig{
		GreenfieldAddrs: []string{"localhost:9090"},
		TendermintAddrs: []string{"http://0.0.0.0:26750"},
	},
	},
}
