package greenfield

type NodeConfig struct {
	GreenfieldAddr string
	TendermintAddr string
}

type GreenfieldChainConfig struct {
	ChainID  string
	NodeAddr []*NodeConfig
}
