package greenfield

type NodeConfig struct {
	GreenfieldAddresses []string
	TendermintAddresses []string
}

type GreenfieldChainConfig struct {
	ChainID  string
	NodeAddr []*NodeConfig
}
