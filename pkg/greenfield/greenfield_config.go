package greenfield

type NodeConfig struct {
	GreenfieldAddr []string
	TendermintAddr []string
}

type GreenfieldChainConfig struct {
	ChainID  string
	NodeAddr []*NodeConfig
}

var DefaultGreenfieldChainConfig = &GreenfieldChainConfig{
	ChainID: "greenfield_9000-1741",
	NodeAddr: []*NodeConfig{&NodeConfig{
		GreenfieldAddr: []string{"localhost:9090"},
		TendermintAddr: []string{"http://0.0.0.0:26750"},
	},
	},
}
