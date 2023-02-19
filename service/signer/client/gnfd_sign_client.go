package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"sync"

	"cosmossdk.io/errors"
	"github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	ctypes "github.com/bnb-chain/greenfield/sdk/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
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
	operatorClient := client.NewGreenfieldClient(gRPCAddr, chainID, client.WithKeyManager(operatorKM),
		client.WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	fundingKM, err := keys.NewPrivateKeyManager(fundingPrivateKey)
	if err != nil {
		return nil, err
	}
	fundingClient := client.NewGreenfieldClient(gRPCAddr, chainID, client.WithKeyManager(fundingKM),
		client.WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	sealKM, err := keys.NewPrivateKeyManager(sealPrivateKey)
	if err != nil {
		return nil, err
	}
	sealClient := client.NewGreenfieldClient(gRPCAddr, chainID, client.WithKeyManager(sealKM),
		client.WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	approvalKM, err := keys.NewPrivateKeyManager(approvalPrivateKey)
	if err != nil {
		return nil, err
	}
	approvalClient := client.NewGreenfieldClient(gRPCAddr, chainID, client.WithKeyManager(approvalKM),
		client.WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	greenfieldClients := map[SignType]*client.GreenfieldClient{
		SignOperator: operatorClient,
		SignFunding:  fundingClient,
		SignSeal:     sealClient,
		SignApproval: approvalClient,
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

func (client *GreenfieldChainSignClient) VerifySignature(scope SignType, msg, sig []byte) bool {
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		return false
	}
	sigHash := crypto.Keccak256(msg)
	return VerifySignature(km.GetAddr(), sigHash, sig) == nil
	//return storagetypes.VerifySignature(km.GetAddr(), crypto.Keccak256(msg), sig) == nil
}

// TODO: bump to the latest version of greenfield, waiting for the fixed version
func VerifySignature(sigAccAddress sdk.AccAddress, sigHash []byte, sig []byte) error {
	if len(sig) != crypto.SignatureLength {
		return errors.Wrapf(sdkerrors.ErrorInvalidSigner, "signature length (actual: %d) doesn't match typical [R||S||V] signature 65 bytes", len(sig))
	}
	if sig[crypto.RecoveryIDOffset] == 27 || sig[crypto.RecoveryIDOffset] == 28 {
		sig[crypto.RecoveryIDOffset] -= 27
	}
	pubKeyBytes, err := secp256k1.RecoverPubkey(sigHash, sig)
	if err != nil {
		return errors.Wrap(err, "failed to recover delegated fee payer from sig")
	}

	ecPubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal recovered fee payer pubkey")
	}

	pubKeyAddr := crypto.PubkeyToAddress(*ecPubKey)
	if !bytes.Equal(pubKeyAddr.Bytes(), sigAccAddress.Bytes()) {
		return errors.Wrapf(sdkerrors.ErrInvalidPubKey, "feePayer pubkey %s is different from transaction pubkey %s", pubKeyAddr, sigAccAddress)
	}

	recoveredSignerAcc := sdk.AccAddress(pubKeyAddr.Bytes())

	if !recoveredSignerAcc.Equals(sigAccAddress) {
		return errors.Wrapf(sdkerrors.ErrorInvalidSigner, "failed to verify delegated fee payer %s signature", recoveredSignerAcc)
	}

	// VerifySignature of ethsecp256k1 accepts 64 byte signature [R||S]
	// WARNING! Under NO CIRCUMSTANCES try to use pubKey.VerifySignature there
	if !secp256k1.VerifySignature(pubKeyBytes, sigHash, sig[:len(sig)-1]) {
		return errors.Wrap(sdkerrors.ErrorInvalidSigner, "unable to verify signer signature of EIP712 typed data")
	}

	return nil
}

// SealObject seal the object on the greenfield chain.
func (client *GreenfieldChainSignClient) SealObject(ctx context.Context, scope SignType, object *ptypes.ObjectInfo) ([]byte, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	var (
		secondarySPAccs       = make([]sdk.AccAddress, 0, len(object.SecondarySps))
		secondarySPSignatures = make([][]byte, 0, len(object.SecondarySps))
	)

	for _, sp := range object.SecondarySps {
		secondarySPAccs = append(secondarySPAccs, sdk.AccAddress(sp.SpId))
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
		[]sdk.Msg{msgSealObject}, txOpt)
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
