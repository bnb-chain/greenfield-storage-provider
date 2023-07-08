package signer

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield/sdk/types"
)

const (
	DefaultSealGasLimit                      = 1200 // fix gas limit for MsgSealObject is 1200
	DefaultSealFeeAmount                     = 6000000000000
	DefaultRejectSealGasLimit                = 12000 // fix gas limit for MsgRejectSealObject is 12000
	DefaultRejectSealFeeAmount               = 60000000000000
	DefaultDiscontinueBucketGasLimit         = 2400 // fix gas limit for MsgDiscontinueBucket is 2400
	DefaultDiscontinueBucketFeeAmount        = 12000000000000
	DefaultCreateGlobalVirtualGroupGasLimit  = 1200 // fix gas limit for MsgCreateGlobalVirtualGroup is 1200
	DefaultCreateGlobalVirtualGroupFeeAmount = 6000000000000
	DefaultCompleteMigrateBucketGasLimit     = 1200 // fix gas limit for MsgCompleteMigrateBucket is 1200
	DefaultCompleteMigrateBucketFeeAmount    = 6000000000000

	// SpOperatorPrivKey defines env variable name for sp operator private key
	SpOperatorPrivKey = "SIGNER_OPERATOR_PRIV_KEY"
	// SpFundingPrivKey defines env variable name for sp funding private key
	SpFundingPrivKey = "SIGNER_FUNDING_PRIV_KEY"
	// SpApprovalPrivKey defines env variable name for sp approval private key
	SpApprovalPrivKey = "SIGNER_APPROVAL_PRIV_KEY"
	// SpSealPrivKey defines env variable name for sp seal private key
	SpSealPrivKey = "SIGNER_SEAL_PRIV_KEY"
	// SpBlsPrivKey defines env variable name for sp bls private key
	SpBlsPrivKey = "SIGNER_BLS_PRIV_KEY"
	// SpGcPrivKey defines env variable name for sp gc private key
	SpGcPrivKey = "SIGNER_GC_PRIV_KEY"
)

func NewSignModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	signer := &SignModular{baseApp: app}
	if err := DefaultSignerOptions(signer, cfg); err != nil {
		return nil, err
	}
	return signer, nil
}

func DefaultSignerOptions(signer *SignModular, cfg *gfspconfig.GfSpConfig) error {
	if len(cfg.Chain.ChainAddress) == 0 {
		return fmt.Errorf("chain address missing")
	}
	if cfg.Chain.SealGasLimit == 0 {
		cfg.Chain.SealGasLimit = DefaultSealGasLimit
	}
	if cfg.Chain.SealFeeAmount == 0 {
		cfg.Chain.SealFeeAmount = DefaultSealFeeAmount
	}
	if cfg.Chain.RejectSealGasLimit == 0 {
		cfg.Chain.RejectSealGasLimit = DefaultRejectSealGasLimit
	}
	if cfg.Chain.RejectSealFeeAmount == 0 {
		cfg.Chain.RejectSealFeeAmount = DefaultRejectSealFeeAmount
	}
	if cfg.Chain.DiscontinueBucketGasLimit == 0 {
		cfg.Chain.DiscontinueBucketGasLimit = DefaultDiscontinueBucketGasLimit
	}
	if cfg.Chain.DiscontinueBucketFeeAmount == 0 {
		cfg.Chain.DiscontinueBucketFeeAmount = DefaultDiscontinueBucketFeeAmount
	}
	if cfg.Chain.CreateGlobalVirtualGroupGasLimit == 0 {
		cfg.Chain.CreateGlobalVirtualGroupGasLimit = DefaultCreateGlobalVirtualGroupGasLimit
	}
	if cfg.Chain.CreateGlobalVirtualGroupFeeAmount == 0 {
		cfg.Chain.CreateGlobalVirtualGroupFeeAmount = DefaultCreateGlobalVirtualGroupFeeAmount
	}
	if cfg.Chain.CompleteMigrateBucketGasLimit == 0 {
		cfg.Chain.CompleteMigrateBucketGasLimit = DefaultCompleteMigrateBucketGasLimit
	}
	if cfg.Chain.CompleteMigrateBucketFeeAmount == 0 {
		cfg.Chain.CompleteMigrateBucketFeeAmount = DefaultCompleteMigrateBucketFeeAmount
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
	if val, ok := os.LookupEnv(SpBlsPrivKey); ok {
		cfg.SpAccount.BlsPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpApprovalPrivKey); ok {
		cfg.SpAccount.ApprovalPrivateKey = val
	}
	if val, ok := os.LookupEnv(SpGcPrivKey); ok {
		cfg.SpAccount.GcPrivateKey = val
	}

	gasInfo := make(map[GasInfoType]GasInfo)
	gasInfo[Seal] = GasInfo{
		GasLimit:  cfg.Chain.SealGasLimit,
		FeeAmount: sdk.NewCoins(sdk.NewCoin(types.Denom, sdk.NewInt(int64(cfg.Chain.SealFeeAmount)))),
	}
	gasInfo[RejectSeal] = GasInfo{
		GasLimit:  cfg.Chain.RejectSealGasLimit,
		FeeAmount: sdk.NewCoins(sdk.NewCoin(types.Denom, sdk.NewInt(int64(cfg.Chain.RejectSealFeeAmount)))),
	}
	gasInfo[DiscontinueBucket] = GasInfo{
		GasLimit:  cfg.Chain.DiscontinueBucketGasLimit,
		FeeAmount: sdk.NewCoins(sdk.NewCoin(types.Denom, sdk.NewInt(int64(cfg.Chain.DiscontinueBucketFeeAmount)))),
	}
	gasInfo[CreateGlobalVirtualGroup] = GasInfo{
		GasLimit:  cfg.Chain.CreateGlobalVirtualGroupGasLimit,
		FeeAmount: sdk.NewCoins(sdk.NewCoin(types.Denom, sdk.NewInt(int64(cfg.Chain.CreateGlobalVirtualGroupFeeAmount)))),
	}

	client, err := NewGreenfieldChainSignClient(cfg.Chain.ChainAddress[0], cfg.Chain.ChainID,
		gasInfo, cfg.SpAccount.OperatorPrivateKey, cfg.SpAccount.FundingPrivateKey,
		cfg.SpAccount.SealPrivateKey, cfg.SpAccount.ApprovalPrivateKey, cfg.SpAccount.GcPrivateKey, cfg.SpAccount.BlsPrivateKey)
	if err != nil {
		return err
	}
	signer.client = client
	return nil
}
