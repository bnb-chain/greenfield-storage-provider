package signer

import (
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

type GreenfieldChainConfig struct {
	GRPCAddress        string
	ChainID            string
	GasLimit           uint64
	OperatorPrivateKey string
	FundingPrivateKey  string
	SealPrivateKey     string
	ApprovalPrivateKey string
}

var DefaultGreenfieldChainConfig = &GreenfieldChainConfig{
	GRPCAddress: "localhost:9090",
	ChainID:     "greenfield_9000-121",
	GasLimit:    210000,
}

type SignerConfig struct {
	GRPCAddress           string
	APIKey                string
	WhitelistCIDR         []string
	GreenfieldChainConfig *GreenfieldChainConfig
}

var DefaultSignerChainConfig = &SignerConfig{
	GRPCAddress: model.SignerGRPCAddress,
	WhitelistCIDR: []string{
		"127.0.0.1/32",
	},
	GreenfieldChainConfig: DefaultGreenfieldChainConfig,
}

func overrideConfigFromEnv(config *SignerConfig) {
	if val, ok := os.LookupEnv(model.SpSignerAPIKey); ok {
		config.APIKey = val
	}
	if val, ok := os.LookupEnv(model.SpOperatorPrivKey); ok {
		config.GreenfieldChainConfig.OperatorPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SpFundingPrivKey); ok {
		config.GreenfieldChainConfig.FundingPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SpApprovalPrivKey); ok {
		config.GreenfieldChainConfig.ApprovalPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SpSealPrivKey); ok {
		config.GreenfieldChainConfig.SealPrivateKey = val
	}
}
