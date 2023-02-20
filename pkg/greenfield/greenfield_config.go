package greenfield

type NodeConfig struct {
	GreenfieldAddr string
	TendermintAddr string
}

type GreenfieldChainConfig struct {
	ChainID  string
	NodeAddr []*NodeConfig
}

var DefaultGreenfieldChainConfig = &GreenfieldChainConfig{
	ChainID: "greenfield_9000-121",
	NodeAddr: []*NodeConfig{&NodeConfig{
		GreenfieldAddr: "127.0.0.1:9090",
		TendermintAddr: "127.0.0.1:9091",
	},
	},
}
