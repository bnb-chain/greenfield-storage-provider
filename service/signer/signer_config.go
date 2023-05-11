package signer

import (
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type SignerConfig struct {
	GRPCAddress        string
	APIKey             string
	WhitelistCIDR      []string
	GasLimit           uint64
	OperatorPrivateKey string
	FundingPrivateKey  string
	SealPrivateKey     string
	ApprovalPrivateKey string
	GcPrivateKey       string
}

var DefaultSignerChainConfig = &SignerConfig{
	GRPCAddress:   model.SignerGRPCAddress,
	WhitelistCIDR: []string{model.WhiteListCIDR},
	GasLimit:      210000,
}

func overrideConfigFromEnv(config *SignerConfig) {
	if val, ok := os.LookupEnv(model.SpSignerAPIKey); ok {
		config.APIKey = val
	}
	if val, ok := os.LookupEnv(model.SpOperatorPrivKey); ok {
		log.Infow("signer", "SpOperatorPrivKey", val)
		config.OperatorPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SpFundingPrivKey); ok {
		log.Infow("signer", "SpFundingPrivKey", val)
		config.FundingPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SpSealPrivKey); ok {
		log.Infow("signer", "SpSealPrivKey", val)
		config.SealPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SpApprovalPrivKey); ok {
		log.Infow("signer", "SpApprovalPrivKey", val)
		config.ApprovalPrivateKey = val
	}
	if val, ok := os.LookupEnv(model.SpGcPrivKey); ok {
		log.Infow("signer", "SpGcPrivKey", val)
		config.GcPrivateKey = val
	}
}
