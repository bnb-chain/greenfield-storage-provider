package signer

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	ctypes "github.com/bnb-chain/greenfield/sdk/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// SignType is the type of msg signature
type SignType string

// GasInfoType is the type of gas info
type GasInfoType string

const (
	// SignOperator is the type of signature signed by the operator account
	SignOperator SignType = "operator"

	// SignFunding is the type of signature signed by the funding account
	SignFunding SignType = "funding"

	// SignSeal is the type of signature signed by the seal account
	SignSeal SignType = "seal"

	// SignApproval is the type of signature signed by the approval account
	SignApproval SignType = "approval"

	// SignGc is the type of signature signed by the gc account
	SignGc SignType = "gc"

	// BroadcastTxRetry defines the max retry for broadcasting tx on-chain
	BroadcastTxRetry = 3

	Seal              GasInfoType = "Seal"
	RejectSeal        GasInfoType = "RejectSeal"
	DiscontinueBucket GasInfoType = "DiscontinueBucket"
)

type GasInfo struct {
	GasLimit  uint64
	FeeAmount sdk.Coins
}

// GreenfieldChainSignClient the greenfield chain client
type GreenfieldChainSignClient struct {
	sealLock sync.Mutex
	gcLock   sync.Mutex

	gasInfo           map[GasInfoType]GasInfo
	greenfieldClients map[SignType]*client.GreenfieldClient
	sealAccNonce      uint64
	gcAccNonce        uint64
}

// NewGreenfieldChainSignClient return the GreenfieldChainSignClient instance
func NewGreenfieldChainSignClient(rpcAddr, chainID string, gasInfo map[GasInfoType]GasInfo, operatorPrivateKey, fundingPrivateKey,
	sealPrivateKey, approvalPrivateKey, gcPrivateKey string) (*GreenfieldChainSignClient, error) {
	// init clients
	// TODO: Get private key from KMS(AWS, GCP, Azure, Aliyun)
	operatorKM, err := keys.NewPrivateKeyManager(operatorPrivateKey)
	if err != nil {
		log.Errorw("failed to new operator private key manager", "error", err)
		return nil, err
	}

	operatorClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(operatorKM))
	if err != nil {
		log.Errorw("failed to new operator greenfield client", "error", err)
		return nil, err
	}
	fundingKM, err := keys.NewPrivateKeyManager(fundingPrivateKey)
	if err != nil {
		log.Errorw("failed to new funding private key manager", "error", err)
		return nil, err
	}
	fundingClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(fundingKM))
	if err != nil {
		log.Errorw("failed to new funding greenfield client", "error", err)
		return nil, err
	}

	sealKM, err := keys.NewPrivateKeyManager(sealPrivateKey)
	if err != nil {
		log.Errorw("failed to new seal private key manager", "error", err)
		return nil, err
	}
	sealClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(sealKM))
	if err != nil {
		log.Errorw("failed to new seal greenfield client", "error", err)
		return nil, err
	}
	sealAccNonce, err := sealClient.GetNonce()
	if err != nil {
		log.Errorw("failed to get nonce", "error", err)
		return nil, err
	}

	approvalKM, err := keys.NewPrivateKeyManager(approvalPrivateKey)
	if err != nil {
		log.Errorw("failed to new approval private key manager", "error", err)
		return nil, err
	}
	approvalClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(approvalKM))
	if err != nil {
		log.Errorw("failed to new approval greenfield client", "error", err)
		return nil, err
	}

	gcKM, err := keys.NewPrivateKeyManager(gcPrivateKey)
	if err != nil {
		return nil, err
	}
	gcClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(gcKM))
	if err != nil {
		return nil, err
	}
	gcAccNonce, err := gcClient.GetNonce()
	if err != nil {
		return nil, err
	}

	greenfieldClients := map[SignType]*client.GreenfieldClient{
		SignOperator: operatorClient,
		SignFunding:  fundingClient,
		SignSeal:     sealClient,
		SignApproval: approvalClient,
		SignGc:       gcClient,
	}

	return &GreenfieldChainSignClient{
		gasInfo:           gasInfo,
		greenfieldClients: greenfieldClients,
		sealAccNonce:      sealAccNonce,
		gcAccNonce:        gcAccNonce,
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
func (client *GreenfieldChainSignClient) SealObject(
	ctx context.Context,
	scope SignType,
	sealObject *storagetypes.MsgSealObject) (
	[]byte, error) {
	if sealObject == nil {
		log.CtxErrorw(ctx, "seal object msg pointer dangling")
		return nil, ErrDanglingPointer
	}
	ctx = log.WithValue(ctx, log.CtxKeyBucketName, sealObject.GetBucketName())
	ctx = log.WithValue(ctx, log.CtxKeyObjectName, sealObject.GetObjectName())
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return nil, ErrSignMsg
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

	client.sealLock.Lock()
	defer client.sealLock.Unlock()

	msgSealObject := storagetypes.NewMsgSealObject(km.GetAddr(),
		sealObject.BucketName, sealObject.ObjectName, secondarySPAccs, sealObject.SecondarySpSignatures)
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash []byte
		nonce  uint64
	)
	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.sealAccNonce
		txOpt := &ctypes.TxOption{
			NoSimulate: true,
			Mode:       &mode,
			GasLimit:   client.gasInfo[Seal].GasLimit,
			FeeAmount:  client.gasInfo[Seal].FeeAmount,
			Nonce:      nonce,
		}

		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgSealObject}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatch, wait for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr := client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get seal account nonce", "error", nonceErr)
				ErrSealObjectOnChain.SetError(fmt.Errorf("failed to get seal account nonce, error: %v", nonceErr))
				return nil, ErrSealObjectOnChain
			}
			client.sealAccNonce = nonce
		}

		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast seal object tx", "error", err, "retry", i)
			continue
		}

		// succeed to broadcast tx
		client.sealAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast seal object tx", "tx_hash", txHash)
		return txHash, nil
	}

	// failed to broadcast tx
	ErrSealObjectOnChain.SetError(fmt.Errorf("failed to broadcast seal object tx, error: %v", err))
	return nil, ErrSealObjectOnChain
}

// RejectUnSealObject reject seal object on the greenfield chain.
func (client *GreenfieldChainSignClient) RejectUnSealObject(
	ctx context.Context,
	scope SignType,
	rejectObject *storagetypes.MsgRejectSealObject) (
	[]byte, error) {
	if rejectObject == nil {
		log.CtxErrorw(ctx, "reject unseal object msg pointer dangling")
		return nil, ErrDanglingPointer
	}
	ctx = log.WithValue(ctx, log.CtxKeyBucketName, rejectObject.GetBucketName())
	ctx = log.WithValue(ctx, log.CtxKeyObjectName, rejectObject.GetObjectName())
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return nil, ErrSignMsg
	}

	client.sealLock.Lock()
	defer client.sealLock.Unlock()

	msgRejectUnSealObject := storagetypes.NewMsgRejectUnsealedObject(km.GetAddr(), rejectObject.GetBucketName(), rejectObject.GetObjectName())
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash []byte
		nonce  uint64
	)

	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.sealAccNonce
		txOpt := &ctypes.TxOption{
			NoSimulate: true,
			Mode:       &mode,
			GasLimit:   client.gasInfo[RejectSeal].GasLimit,
			FeeAmount:  client.gasInfo[RejectSeal].FeeAmount,
			Nonce:      nonce,
		}
		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgRejectUnSealObject}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatch, wait for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr := client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get seal account nonce", "error", nonceErr)
				ErrRejectUnSealObjectOnChain.SetError(fmt.Errorf("failed to get seal account nonce, error: %v", nonceErr))
				return nil, ErrRejectUnSealObjectOnChain
			}
			client.sealAccNonce = nonce
		}

		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast reject unseal object", "error", err, "retry", i)
			continue
		}

		// succeed to broadcast tx
		client.sealAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast reject unseal object", "tx_hash", txHash)
		return txHash, nil
	}
	// failed to broadcast tx
	ErrRejectUnSealObjectOnChain.SetError(fmt.Errorf("failed to broadcast reject unseal object tx, error: %v", err))
	return nil, ErrRejectUnSealObjectOnChain
}

// DiscontinueBucket stops serving the bucket on the greenfield chain.
func (client *GreenfieldChainSignClient) DiscontinueBucket(ctx context.Context, scope SignType, discontinueBucket *storagetypes.MsgDiscontinueBucket) ([]byte, error) {
	log.Infow("signer start to discontinue bucket", "scope", scope)
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "err", err)
		return nil, ErrSignMsg
	}

	client.gcLock.Lock()
	defer client.gcLock.Unlock()
	nonce := client.gcAccNonce

	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(km.GetAddr(),
		discontinueBucket.BucketName, discontinueBucket.Reason)
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &ctypes.TxOption{
		NoSimulate: true,
		Mode:       &mode,
		GasLimit:   client.gasInfo[DiscontinueBucket].GasLimit,
		FeeAmount:  client.gasInfo[DiscontinueBucket].FeeAmount,
		Nonce:      nonce,
	}

	txHash, err := client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgDiscontinueBucket}, txOpt)
	if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
		// if nonce mismatch, wait for next block, reset nonce by querying the nonce on chain
		nonce, nonceErr := client.getNonceOnChain(ctx, client.greenfieldClients[scope])
		if nonceErr != nil {
			log.CtxErrorw(ctx, "failed to get seal account nonce", "error", nonceErr)
			ErrDiscontinueBucketOnChain.SetError(fmt.Errorf("failed to get seal account nonce, error: %v", nonceErr))
			return nil, ErrDiscontinueBucketOnChain
		}
		client.gcAccNonce = nonce
	}

	// failed to broadcast tx
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast discontinue bucket", "error", err, "discontinue_bucket", msgDiscontinueBucket.String())
		ErrDiscontinueBucketOnChain.SetError(fmt.Errorf("failed to broadcast discontinue bucket, error: %v", err))
		return nil, ErrDiscontinueBucketOnChain
	}

	// update nonce when tx is successful submitted
	client.gcAccNonce = nonce + 1
	return txHash, nil
}

func (client *GreenfieldChainSignClient) getNonceOnChain(ctx context.Context, gnfdClient *client.GreenfieldClient) (uint64, error) {
	err := waitForNextBlock(ctx, gnfdClient)
	if err != nil {
		log.CtxErrorw(ctx, "failed to wait next block", "error", err)
		return 0, err
	}
	nonce, err := gnfdClient.GetNonce()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get seal account nonce on chain", "error", err)
		return 0, err
	}
	return nonce, nil
}

func (client *GreenfieldChainSignClient) broadcastTx(
	ctx context.Context, gnfdClient *client.GreenfieldClient,
	msgs []sdk.Msg, txOpt *ctypes.TxOption, opts ...grpc.CallOption,
) ([]byte, error) {
	resp, err := gnfdClient.BroadcastTx(ctx, msgs, txOpt, opts...)
	if err != nil {
		if strings.Contains(err.Error(), "account sequence mismatch") {
			return nil, sdkErrors.ErrWrongSequence
		}
		return nil, errors.Wrap(err, "failed to broadcast tx with greenfield client")
	}
	if resp.TxResponse.Code == sdkErrors.ErrWrongSequence.ABCICode() {
		return nil, sdkErrors.ErrWrongSequence
	}
	if resp.TxResponse.Code != 0 {
		return nil, fmt.Errorf("failed to broadcast tx, resp code: %d", resp.TxResponse.Code)
	}
	txHash, err := hex.DecodeString(resp.TxResponse.TxHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal tx hash")
	}
	return txHash, nil
}

func waitForNextBlock(ctx context.Context, client *client.GreenfieldClient) error {
	height, err := latestBlockHeight(ctx, client)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		latestBlockHeight, err := latestBlockHeight(ctx, client)
		if err != nil {
			return err
		}
		if latestBlockHeight >= height+1 {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout exceeded waiting for block")
		case <-ticker.C:
		}
	}
}

func latestBlockHeight(ctx context.Context, client *client.GreenfieldClient) (int64, error) {
	block, err := client.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return 0, err
	}
	return block.SdkBlock.Header.Height, nil
}
