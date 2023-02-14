package signer

type GreenfieldChainConfig struct {
	RPCAddrs      []string
	GRPCAddrs     []string
	ChainId       uint16
	GasLimit      uint64
	ChainIdString string
	PrivateKey    string
}

var DefaultGreenfieldChainConfig = &GreenfieldChainConfig{
	GRPCAddrs:     []string{},
	ChainId:       9000,
	GasLimit:      210000,
	ChainIdString: "greenfield_9000-1741",
	PrivateKey:    "",
}

type SignerConfig struct {
	Address               string
	GreenfieldChainConfig *GreenfieldChainConfig
}

var DefaultSignerChainConfig = &SignerConfig{
	GreenfieldChainConfig: DefaultGreenfieldChainConfig,
}
