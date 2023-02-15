package signer

type GreenfieldChainConfig struct {
	RPCAddrs           []string
	GRPCAddrs          []string
	ChainId            uint16
	GasLimit           uint64
	ChainIdString      string
	OperatorPrivateKey string
	FundingPrivateKey  string
	SealPrivateKey     string
	ApprovalPrivateKey string
}

var DefaultGreenfieldChainConfig = &GreenfieldChainConfig{
	GRPCAddrs:     []string{},
	ChainId:       9000,
	GasLimit:      210000,
	ChainIdString: "greenfield_9000-1741",
}

type SignerConfig struct {
	Address               string
	APIKey                string
	WhitelistCIDR         []string
	GreenfieldChainConfig *GreenfieldChainConfig
}

var DefaultSignerChainConfig = &SignerConfig{
	WhitelistCIDR: []string{
		"127.0.0.1/32",
	},
	GreenfieldChainConfig: DefaultGreenfieldChainConfig,
}
