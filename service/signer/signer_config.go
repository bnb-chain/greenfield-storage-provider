package signer

import "os"

type SecretKey struct {
	Driver string
}

type GreenfieldChainConfig struct {
	GRPCAddr           string
	ChainID            string
	GasLimit           uint64
	OperatorPrivateKey string
	FundingPrivateKey  string
	SealPrivateKey     string
	ApprovalPrivateKey string
}

var DefaultGreenfieldChainConfig = &GreenfieldChainConfig{
	GRPCAddr: "localhost:9090",
	ChainID:  "greenfield_9000-121",
	GasLimit: 210000,
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

func overrideConfigFromEnv(config *SignerConfig) {
	if val, ok := os.LookupEnv("SIGNER_API_KEY"); ok {
		config.APIKey = val
	}
	if val, ok := os.LookupEnv("SIGNER_OPERATOR_PRIV_KEY"); ok {
		config.GreenfieldChainConfig.OperatorPrivateKey = val
	}
	if val, ok := os.LookupEnv("SIGNER_FUNDING_PRIV_KEY"); ok {
		config.GreenfieldChainConfig.FundingPrivateKey = val
	}
	if val, ok := os.LookupEnv("SIGNER_APPROVAL_PRIV_KEY"); ok {
		config.GreenfieldChainConfig.ApprovalPrivateKey = val
	}
	if val, ok := os.LookupEnv("SIGNER_SEAL_PRIV_KEY"); ok {
		config.GreenfieldChainConfig.SealPrivateKey = val
	}
}
