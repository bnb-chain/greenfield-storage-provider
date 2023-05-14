package singer

import (
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspmdmgr"
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

func init() {
	gfspmdmgr.RegisterModularInfo(SignerModularName, SignerModularDescription, NewSingModular)
}

func NewSingModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspconfig.Option) (
	coremodule.Modular, error) {
	if cfg.Signer != nil {
		app.SetSigner(cfg.Signer)
		return cfg.Signer, nil
	}
	signer := &SingModular{baseApp: app}
	opts = append(opts, signer.DefaultSingerOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetSigner(signer)
	return signer, nil
}

func (s *SingModular) DefaultSingerOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if len(cfg.ChainAddress) == 0 {
		return fmt.Errorf("chain address missing")
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = DefaultGasLimit
	}
	if val, ok := os.LookupEnv(SpOperatorPrivKey); ok {
		cfg.OperatorPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpFundingPrivKey); ok {
		cfg.FundingPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpSealPrivKey); ok {
		cfg.SealPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpApprovalPrivKey); ok {
		cfg.ApprovalPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpGcPrivKey); ok {
		cfg.GcPrivateKey = val
	}
	client, err := NewGreenfieldChainSignClient(cfg.ChainAddress[0], cfg.ChainID, cfg.GasLimit,
		cfg.OperatorPrivateKey, cfg.FundingPrivateKey, cfg.SealPrivateKey, cfg.ApprovalPrivateKey)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}
