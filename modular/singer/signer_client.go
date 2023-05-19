package singer

import (
	"context"
	"encoding/hex"
	"strings"
	"sync"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	ctypes "github.com/bnb-chain/greenfield/sdk/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// SignType is the type of msg signature
type SignType string

const (
	// SignOperator is the type of signature signed by the operator account
	SignOperator SignType = "operator"

	// SignFunding is the type of signature signed by the funding account
	SignFunding SignType = "funding"

	// SignSeal is the type of signature signed by the seal account
	SignSeal SignType = "seal"

	// SignApproval is the type of signature signed by the approval account
	SignApproval SignType = "approval"
)

// GreenfieldChainSignClient the greenfield chain client
type GreenfieldChainSignClient struct {
	mu sync.Mutex

	gasLimit          uint64
	greenfieldClients map[SignType]*client.GreenfieldClient
	sealAccNonce      uint64
}

// NewGreenfieldChainSignClient return the GreenfieldChainSignClient instance
func NewGreenfieldChainSignClient(rpcAddr, chainID string, gasLimit uint64, operatorPrivateKey, fundingPrivateKey,
	sealPrivateKey, approvalPrivateKey string) (*GreenfieldChainSignClient, error) {
	// init clients
	// TODO: Get private key from KMS(AWS, GCP, Azure, Aliyun)
	operatorKM, err := keys.NewPrivateKeyManager(operatorPrivateKey)
	if err != nil {
		return nil, err
	}

	operatorClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(operatorKM))
	if err != nil {
		return nil, err
	}
	fundingKM, err := keys.NewPrivateKeyManager(fundingPrivateKey)
	if err != nil {
		return nil, err
	}
	fundingClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(fundingKM))
	if err != nil {
		return nil, err
	}

	sealKM, err := keys.NewPrivateKeyManager(sealPrivateKey)
	if err != nil {
		return nil, err
	}
	sealClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(sealKM))
	if err != nil {
		return nil, err
	}
	sealAccNonce, err := sealClient.GetNonce()
	if err != nil {
		return nil, err
	}

	approvalKM, err := keys.NewPrivateKeyManager(approvalPrivateKey)
	if err != nil {
		return nil, err
	}
	approvalClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(approvalKM))
	if err != nil {
		return nil, err
	}

	greenfieldClients := map[SignType]*client.GreenfieldClient{
		SignOperator: operatorClient,
		SignFunding:  fundingClient,
		SignSeal:     sealClient,
		SignApproval: approvalClient,
	}

	return &GreenfieldChainSignClient{
		gasLimit:          gasLimit,
		greenfieldClients: greenfieldClients,
		sealAccNonce:      sealAccNonce - 1,
	}, nil
}

// GetAddr returns the public address of the private key.
func (client *GreenfieldChainSignClient) GetAddr(scope SignType) (sdk.AccAddress, error) {
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		return nil, err
	}
	return km.GetAddr(), nil
}

// Sign returns a msg signature signed by private key.
func (client *GreenfieldChainSignClient) Sign(scope SignType, msg []byte) ([]byte, error) {
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		return nil, err
	}
	return km.Sign(msg)
}

// VerifySignature verifies the signature.
func (client *GreenfieldChainSignClient) VerifySignature(scope SignType, msg, sig []byte) bool {
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		return false
	}

	return storagetypes.VerifySignature(km.GetAddr(), crypto.Keccak256(msg), sig) == nil
}

// SealObject seal the object on the greenfield chain.
func (client *GreenfieldChainSignClient) SealObject(ctx context.Context, scope SignType, sealObject *storagetypes.MsgSealObject) ([]byte, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "err", err)
		return nil, merrors.ErrSignMsg
	}

	var secondarySPAccs []sdk.AccAddress
	for _, sp := range sealObject.SecondarySpAddresses {
		opAddr, err := sdk.AccAddressFromHexUnsafe(sp) // should be 0x...
		if err != nil {
			log.CtxErrorw(ctx, "failed to parse address", "error", err, "address", opAddr)
			return nil, err
		}
		secondarySPAccs = append(secondarySPAccs, opAddr)
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	nonce := client.sealAccNonce + 1

	msgSealObject := storagetypes.NewMsgSealObject(km.GetAddr(),
		sealObject.BucketName, sealObject.ObjectName, secondarySPAccs, sealObject.SecondarySpSignatures)
	mode := tx.BroadcastMode_BROADCAST_MODE_ASYNC
	txOpt := &ctypes.TxOption{
		Mode:     &mode,
		GasLimit: client.gasLimit,
		Nonce:    nonce,
	}

	resp, err := client.greenfieldClients[scope].BroadcastTx(ctx, []sdk.Msg{msgSealObject}, txOpt)
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast tx", "err", err, "seal_info", msgSealObject.String())
		if strings.Contains(err.Error(), "account sequence mismatch") {
			// if nonce mismatch, reset nonce by querying the nonce on chain
			nonce, err = client.greenfieldClients[scope].GetNonce()
			if err != nil {
				log.CtxErrorw(ctx, "failed to get seal account nonce", "err", err, "seal_info", msgSealObject.String())
				return nil, merrors.ErrSealObjectOnChain
			}
			client.sealAccNonce = nonce - 1
		}
		return nil, merrors.ErrSealObjectOnChain

	}

	if resp.TxResponse.Code != 0 {
		log.CtxErrorf(ctx, "failed to broadcast tx, resp code: %d", resp.TxResponse.Code, "seal_info", msgSealObject.String())
		return nil, merrors.ErrSealObjectOnChain
	}
	txHash, err := hex.DecodeString(resp.TxResponse.TxHash)
	if err != nil {
		log.CtxErrorw(ctx, "failed to marshal tx hash", "err", err, "seal_info", msgSealObject.String())
		return nil, merrors.ErrSealObjectOnChain
	}
	client.sealAccNonce = nonce

	return txHash, nil
}
