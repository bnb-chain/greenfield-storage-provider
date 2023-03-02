package signer

import (
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

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
	Address: model.SignerGRPCAddress,
	WhitelistCIDR: []string{
		"127.0.0.1/32",
	},
	GreenfieldChainConfig: DefaultGreenfieldChainConfig,
}

func overrideConfigFromEnv(config *SignerConfig) {
	if val, ok := os.LookupEnv(model.SPSignerAPIKey); ok {
		config.APIKey = val
	}
	if val, ok := os.LookupEnv(model.SPOperatorPrivKey); ok {
		config.GreenfieldChainConfig.OperatorPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SPFundingPrivKey); ok {
		config.GreenfieldChainConfig.FundingPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SPApprovalPrivKey); ok {
		config.GreenfieldChainConfig.ApprovalPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SPSealPrivKey); ok {
		config.GreenfieldChainConfig.SealPrivateKey = val
	}
}
