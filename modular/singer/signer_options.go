package singer

import (
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	// DefaultGasLimit defines the default gas limit
	DefaultGasLimit = 210000
	// SpOperatorPrivKey defines env variable name for sp operator priv key
	SpOperatorPrivKey = "SIGNER_OPERATOR_PRIV_KEY"
	// SpFundingPrivKey defines env variable name for sp funding priv key
	SpFundingPrivKey = "SIGNER_FUNDING_PRIV_KEY"
	// SpApprovalPrivKey defines env variable name for sp approval priv key
	SpApprovalPrivKey = "SIGNER_APPROVAL_PRIV_KEY"
	// SpSealPrivKey defines env variable name for sp seal priv key
	SpSealPrivKey = "SIGNER_SEAL_PRIV_KEY"
	// SpGcPrivKey defines env variable name for sp gc priv key
	SpGcPrivKey = "SIGNER_GC_PRIV_KEY"
)

func NewSingModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	signer := &SingModular{baseApp: app}
	if err := DefaultSingerOptions(signer, cfg); err != nil {
		return nil, err
	}
	return signer, nil
}

func DefaultSingerOptions(signer *SingModular, cfg *gfspconfig.GfSpConfig) error {
	if len(cfg.Chain.ChainAddress) == 0 {
		return fmt.Errorf("chain address missing")
	}
	if cfg.Chain.GasLimit == 0 {
		cfg.Chain.GasLimit = DefaultGasLimit
	}
	if val, ok := os.LookupEnv(SpOperatorPrivKey); ok {
		cfg.SpAccount.OperatorPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpFundingPrivKey); ok {
		cfg.SpAccount.FundingPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpSealPrivKey); ok {
		cfg.SpAccount.SealPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpApprovalPrivKey); ok {
		cfg.SpAccount.ApprovalPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpGcPrivKey); ok {
		cfg.SpAccount.GcPrivateKey = val
	}
	client, err := NewGreenfieldChainSignClient(cfg.Chain.ChainAddress[0], cfg.Chain.ChainID,
		cfg.Chain.GasLimit, cfg.SpAccount.OperatorPrivateKey, cfg.SpAccount.FundingPrivateKey,
		cfg.SpAccount.SealPrivateKey, cfg.SpAccount.ApprovalPrivateKey)
	if err != nil {
		return err
	}
	signer.client = client
	return nil
}
