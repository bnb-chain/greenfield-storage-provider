package signer

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/bnb-chain/greenfield-go-sdk/client/chain"
	"github.com/bnb-chain/greenfield-go-sdk/keys"
	ctypes "github.com/bnb-chain/greenfield-go-sdk/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

type SignType string

const (
	SignOperator SignType = "operator"
	SignFunding  SignType = "funding"
	SignSeal     SignType = "seal"
	SignApproval SignType = "approval"
)

// GreenfieldChainClient the greenfield chain client
type GreenfieldChainClient struct {
	mu sync.Mutex

	config            *GreenfieldChainConfig
	greenfieldClients map[SignType]*chain.GreenfieldClient
}

// NewGreenfieldChainClient return the GreenfieldChainClient instance
func NewGreenfieldChainClient(config *GreenfieldChainConfig) (*GreenfieldChainClient, error) {
	// init clients
	// TODO: Get private key from KMS(AWS, GCP, Azure, Aliyun)
	operatorKM, err := keys.NewPrivateKeyManager(config.OperatorPrivateKey)
	if err != nil {
		return nil, err
	}
	operatorClient := chain.NewGreenfieldClientWithKeyManager(config.GRPCAddr, config.ChainIdString, operatorKM)
	fundingKM, err := keys.NewPrivateKeyManager(config.FundingPrivateKey)
	if err != nil {
		return nil, err
	}
	fundingClient := chain.NewGreenfieldClientWithKeyManager(config.GRPCAddr, config.ChainIdString, fundingKM)
	sealKM, err := keys.NewPrivateKeyManager(config.SealPrivateKey)
	if err != nil {
		return nil, err
	}
	sealClient := chain.NewGreenfieldClientWithKeyManager(config.GRPCAddr, config.ChainIdString, sealKM)
	approvalKM, err := keys.NewPrivateKeyManager(config.ApprovalPrivateKey)
	if err != nil {
		return nil, err
	}
	approvalClient := chain.NewGreenfieldClientWithKeyManager(config.GRPCAddr, config.ChainIdString, approvalKM)
	greenfieldClients := map[SignType]*chain.GreenfieldClient{
		SignOperator: &operatorClient,
		SignFunding:  &fundingClient,
		SignSeal:     &sealClient,
		SignApproval: &approvalClient,
	}

	cli := &GreenfieldChainClient{
		config:            config,
		greenfieldClients: greenfieldClients,
	}
	return cli, nil
}

// Sign returns a msg signature signed by private key.
func (client *GreenfieldChainClient) Sign(scope SignType, msg []byte) ([]byte, error) {
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		return nil, err
	}
	return km.GetPrivKey().Sign(msg)
}

// SealObject seal the object on the greenfield chain.
func (client *GreenfieldChainClient) SealObject(ctx context.Context, scope SignType, object *ptypes.ObjectInfo) ([]byte, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	var (
		secondarySPAccs       = make([]types.AccAddress, 0, len(object.SecondarySps))
		secondarySPSignatures = make([][]byte, 0, len(object.SecondarySps))
	)

	for _, sp := range object.SecondarySps {
		secondarySPAccs = append(secondarySPAccs, types.AccAddress(sp.SpId))
		secondarySPSignatures = append(secondarySPSignatures, sp.Signature)
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "err", err)
		return nil, merrors.ErrSignMsg
	}

	msgSealObject := storagetypes.NewMsgSealObject(km.GetAddr(),
		object.BucketName, object.ObjectName, secondarySPAccs, secondarySPSignatures)
	mode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	txOpt := &ctypes.TxOption{
		Mode:     &mode,
		GasLimit: client.config.GasLimit,
	}

	resp, err := client.greenfieldClients[scope].BroadcastTx(
		[]types.Msg{msgSealObject}, txOpt)
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast tx", "err", err)
		return nil, merrors.ErrSealObjectOnChain
	}

	if resp.TxResponse.Code != 0 {
		log.CtxErrorf(ctx, "failed to broadcast tx, resp code: %d", resp.TxResponse.Code)
		return nil, merrors.ErrSealObjectOnChain
	}
	object.TxHash, err = hex.DecodeString(resp.TxResponse.TxHash)
	if err != nil {
		log.CtxErrorw(ctx, "failed to marshal tx hash", "err", err)
		return nil, merrors.ErrSealObjectOnChain
	}

	return object.TxHash, nil
}
