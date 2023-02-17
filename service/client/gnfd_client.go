package client

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

// GreenfieldChainSignClient the greenfield chain client
type GreenfieldChainSignClient struct {
	mu sync.Mutex

	gasLimit          uint64
	greenfieldClients map[SignType]*chain.GreenfieldClient
}

// NewGreenfieldChainSignClient return the GreenfieldChainSignClient instance
func NewGreenfieldChainSignClient(
	gRPCAddr string,
	chainID string,
	gasLimit uint64,
	operatorPrivateKey string,
	fundingPrivateKey string,
	sealPrivateKey string,
	approvalPrivateKey string) (*GreenfieldChainSignClient, error) {
	// init clients
	// TODO: Get private key from KMS(AWS, GCP, Azure, Aliyun)
	operatorKM, err := keys.NewPrivateKeyManager(operatorPrivateKey)
	if err != nil {
		return nil, err
	}
	operatorClient := chain.NewGreenfieldClientWithKeyManager(gRPCAddr, chainID, operatorKM)
	fundingKM, err := keys.NewPrivateKeyManager(fundingPrivateKey)
	if err != nil {
		return nil, err
	}
	fundingClient := chain.NewGreenfieldClientWithKeyManager(gRPCAddr, chainID, fundingKM)
	sealKM, err := keys.NewPrivateKeyManager(sealPrivateKey)
	if err != nil {
		return nil, err
	}
	sealClient := chain.NewGreenfieldClientWithKeyManager(gRPCAddr, chainID, sealKM)
	approvalKM, err := keys.NewPrivateKeyManager(approvalPrivateKey)
	if err != nil {
		return nil, err
	}
	approvalClient := chain.NewGreenfieldClientWithKeyManager(gRPCAddr, chainID, approvalKM)
	greenfieldClients := map[SignType]*chain.GreenfieldClient{
		SignOperator: &operatorClient,
		SignFunding:  &fundingClient,
		SignSeal:     &sealClient,
		SignApproval: &approvalClient,
	}

	return &GreenfieldChainSignClient{
		gasLimit:          gasLimit,
		greenfieldClients: greenfieldClients,
	}, nil
}

// Sign returns a msg signature signed by private key.
func (client *GreenfieldChainSignClient) Sign(scope SignType, msg []byte) ([]byte, error) {
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		return nil, err
	}
	return km.GetPrivKey().Sign(msg)
}

// SealObject seal the object on the greenfield chain.
func (client *GreenfieldChainSignClient) SealObject(ctx context.Context, scope SignType, object *ptypes.ObjectInfo) ([]byte, error) {
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
		GasLimit: client.gasLimit,
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
