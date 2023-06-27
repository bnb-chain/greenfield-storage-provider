package signer

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	ctypes "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// SignType is the type of msg signature
type SignType string

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

	Seal                     GasInfoType = "Seal"
	RejectSeal               GasInfoType = "RejectSeal"
	DiscontinueBucket        GasInfoType = "DiscontinueBucket"
	CreateGlobalVirtualGroup GasInfoType = "CreateGlobalVirtualGroup"
	CompleteMigration        GasInfoType = "CompleteMigration"
)

type GasInfo struct {
	GasLimit  uint64
	FeeAmount sdk.Coins
}

// GreenfieldChainSignClient the greenfield chain client
type GreenfieldChainSignClient struct {
	mu sync.Mutex

	gasInfo           map[GasInfoType]GasInfo
	greenfieldClients map[SignType]*client.GreenfieldClient
	operatorAccNonce  uint64
	sealAccNonce      uint64
	gcAccNonce        uint64
	blsKm             keys.KeyManager
}

// NewGreenfieldChainSignClient return the GreenfieldChainSignClient instance
func NewGreenfieldChainSignClient(rpcAddr, chainID string, gasInfo map[GasInfoType]GasInfo, operatorPrivateKey, fundingPrivateKey,
	sealPrivateKey, approvalPrivateKey, gcPrivateKey string, blsPrivKey string) (*GreenfieldChainSignClient, error) {
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
	operatorAccNonce, err := operatorClient.GetNonce()
	if err != nil {
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

	blsKM, err := keys.NewBlsPrivateKeyManager(blsPrivKey)
	if err != nil {
		log.Errorw("failed to new bls private key manager", "error", err)
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
		operatorAccNonce:  operatorAccNonce,
		blsKm:             blsKM,
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

	return types.VerifySignature(km.GetAddr(), crypto.Keccak256(msg), sig) == nil
}

// SealObject seal the object on the greenfield chain.
func (client *GreenfieldChainSignClient) SealObject(ctx context.Context, scope SignType,
	sealObject *storagetypes.MsgSealObject) ([]byte, error) {
	if sealObject == nil {
		log.CtxErrorw(ctx, "failed to seal object due to pointer dangling")
		return nil, ErrDanglingPointer
	}
	ctx = log.WithValue(ctx, log.CtxKeyBucketName, sealObject.GetBucketName())
	ctx = log.WithValue(ctx, log.CtxKeyObjectName, sealObject.GetObjectName())
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return nil, ErrSignMsg
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	msgSealObject := storagetypes.NewMsgSealObject(km.GetAddr(),
		sealObject.GetBucketName(), sealObject.GetObjectName(), sealObject.GetGlobalVirtualGroupId(),
		sealObject.GetSecondarySpBlsAggSignatures())

	mode := tx.BroadcastMode_BROADCAST_MODE_ASYNC

	var (
		resp   *tx.BroadcastTxResponse
		txHash []byte
		nonce  uint64
	)
	nonce = client.sealAccNonce
	for i := 0; i < BroadcastTxRetry; i++ {
		txOpt := &ctypes.TxOption{
			NoSimulate: true,
			Mode:       &mode,
			GasLimit:   client.gasInfo[Seal].GasLimit,
			FeeAmount:  client.gasInfo[Seal].FeeAmount,
			Nonce:      nonce,
		}
		resp, err = client.greenfieldClients[scope].BroadcastTx(ctx, []sdk.Msg{msgSealObject}, txOpt)
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast seal object tx", "error", err)
			if strings.Contains(err.Error(), "account sequence mismatch") {
				// if nonce mismatch, reset nonce by querying the nonce on chain
				nonce, err = client.greenfieldClients[scope].GetNonce()
				if err != nil {
					log.CtxErrorw(ctx, "failed to get seal account nonce", "error", err)
					ErrSealObjectOnChain.SetError(fmt.Errorf("failed to get seal account nonce, error: %v", err))
					return nil, ErrSealObjectOnChain
				}
				client.sealAccNonce = nonce
			}
			continue
		}
		if resp.TxResponse.Code != 0 {
			log.CtxErrorf(ctx, "failed to broadcast tx, resp code: %d", resp.TxResponse.Code)
			ErrSealObjectOnChain.SetError(fmt.Errorf("failed to broadcast  seal object tx, resp_code: %d", resp.TxResponse.Code))
			err = ErrSealObjectOnChain
			continue
		}
		txHash, err = hex.DecodeString(resp.TxResponse.TxHash)
		if err != nil {
			log.CtxErrorw(ctx, "failed to marshal tx hash", "error", err)
			ErrSealObjectOnChain.SetError(fmt.Errorf("failed to decode seal object tx hash, error: %v", err))
			err = ErrSealObjectOnChain
			continue
		}
		client.sealAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast seal object tx", "tx_hash", txHash, "seal_msg", msgSealObject)
		return txHash, nil
	}
	return nil, err
}

// RejectUnSealObject reject seal object on the greenfield chain.
func (client *GreenfieldChainSignClient) RejectUnSealObject(ctx context.Context, scope SignType,
	rejectObject *storagetypes.MsgRejectSealObject) ([]byte, error) {
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

	client.mu.Lock()
	defer client.mu.Unlock()

	msgRejectUnSealObject := storagetypes.NewMsgRejectUnsealedObject(km.GetAddr(), rejectObject.GetBucketName(), rejectObject.GetObjectName())
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		resp   *tx.BroadcastTxResponse
		txHash []byte
		nonce  uint64
	)
	nonce = client.sealAccNonce
	for i := 0; i < BroadcastTxRetry; i++ {
		txOpt := &ctypes.TxOption{
			NoSimulate: true,
			Mode:       &mode,
			GasLimit:   client.gasInfo[RejectSeal].GasLimit,
			FeeAmount:  client.gasInfo[RejectSeal].FeeAmount,
			Nonce:      nonce,
		}
		resp, err = client.greenfieldClients[scope].BroadcastTx(ctx, []sdk.Msg{msgRejectUnSealObject}, txOpt)
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast reject unseal object tx", "error", err)
			if strings.Contains(err.Error(), "account sequence mismatch") {
				// if nonce mismatch, reset nonce by querying the nonce on chain
				nonce, err = client.greenfieldClients[scope].GetNonce()
				if err != nil {
					log.CtxErrorw(ctx, "failed to get seal account nonce", "error", err)
					ErrRejectUnSealObjectOnChain.SetError(fmt.Errorf("failed to get seal account nonce, error: %v", err))
					return nil, ErrRejectUnSealObjectOnChain
				}
				client.sealAccNonce = nonce
			}
			continue
		}

		if resp.TxResponse.Code != 0 {
			log.CtxErrorf(ctx, "failed to broadcast tx, resp code: %d", resp.TxResponse.Code)
			ErrSealObjectOnChain.SetError(fmt.Errorf("failed to broadcast reject unseal object tx, resp_code: %d", resp.TxResponse.Code))
			err = ErrSealObjectOnChain
			continue
		}
		txHash, err = hex.DecodeString(resp.TxResponse.TxHash)
		if err != nil {
			log.CtxErrorw(ctx, "failed to marshal tx hash", "error", err)
			ErrRejectUnSealObjectOnChain.SetError(fmt.Errorf("failed to decode reject unseal object tx hash, error: %v", err))
			err = ErrRejectUnSealObjectOnChain
			continue
		}

		client.sealAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast reject unseal object tx", "tx_hash", txHash)
		return txHash, nil
	}
	return nil, err
}

// DiscontinueBucket stops serving the bucket on the greenfield chain.
func (client *GreenfieldChainSignClient) DiscontinueBucket(ctx context.Context, scope SignType, discontinueBucket *storagetypes.MsgDiscontinueBucket) ([]byte, error) {
	log.Infow("signer start to discontinue bucket", "scope", scope)
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "err", err)
		return nil, ErrSignMsg
	}

	client.mu.Lock()
	defer client.mu.Unlock()
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

	resp, err := client.greenfieldClients[scope].BroadcastTx(ctx, []sdk.Msg{msgDiscontinueBucket}, txOpt)
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast tx", "err", err, "discontinue_bucket", msgDiscontinueBucket.String())
		if strings.Contains(err.Error(), "account sequence mismatch") {
			// if nonce mismatch, reset nonce by querying the nonce on chain
			nonce, err := client.greenfieldClients[scope].GetNonce()
			if err != nil {
				log.CtxErrorw(ctx, "failed to get gc account nonce", "err", err)
				return nil, ErrDiscontinueBucketOnChain
			}
			client.gcAccNonce = nonce
		}
		return nil, ErrDiscontinueBucketOnChain
	}

	if resp.TxResponse.Code != 0 {
		log.CtxErrorf(ctx, "failed to broadcast tx, resp code: %d", resp.TxResponse.Code, "discontinue_bucket", msgDiscontinueBucket.String())
		return nil, ErrDiscontinueBucketOnChain
	}
	txHash, err := hex.DecodeString(resp.TxResponse.TxHash)
	if err != nil {
		log.CtxErrorw(ctx, "failed to marshal tx hash", "err", err, "discontinue_bucket", msgDiscontinueBucket.String())
		return nil, ErrDiscontinueBucketOnChain
	}

	// update nonce when tx is successful submitted
	client.gcAccNonce = nonce + 1
	return txHash, nil
}

func (client *GreenfieldChainSignClient) CreateGlobalVirtualGroup(ctx context.Context, scope SignType, gvg *virtualgrouptypes.MsgCreateGlobalVirtualGroup) ([]byte, error) {
	log.Infow("signer start to create a new global virtual group", "scope", scope)
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return nil, ErrSignMsg
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	nonce := client.operatorAccNonce

	msgCreateGlobalVirtualGroup := virtualgrouptypes.NewMsgCreateGlobalVirtualGroup(km.GetAddr(),
		gvg.FamilyId, gvg.GetSecondarySpIds(), gvg.GetDeposit())
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &ctypes.TxOption{
		Mode:      &mode,
		GasLimit:  client.gasInfo[CreateGlobalVirtualGroup].GasLimit,
		FeeAmount: client.gasInfo[CreateGlobalVirtualGroup].FeeAmount,
		Nonce:     nonce,
	}

	resp, err := client.greenfieldClients[scope].BroadcastTx(ctx, []sdk.Msg{msgCreateGlobalVirtualGroup}, txOpt)
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast tx", "err", err, "global_virtual_group", msgCreateGlobalVirtualGroup.String())
		if strings.Contains(err.Error(), "account sequence mismatch") {
			// if nonce mismatch, reset nonce by querying the nonce on chain
			nonce, err := client.greenfieldClients[scope].GetNonce()
			if err != nil {
				log.CtxErrorw(ctx, "failed to get approval account nonce", "err", err)
				return nil, ErrCreateGVGOnChain
			}
			client.operatorAccNonce = nonce
		}
		return nil, ErrCreateGVGOnChain
	}

	if resp.TxResponse.Code != 0 {
		log.CtxErrorw(ctx, "failed to broadcast tx", "code", resp.TxResponse.Code,
			"msg", resp.TxResponse.String(),
			"global_virtual_group", msgCreateGlobalVirtualGroup.String())
		return nil, ErrCreateGVGOnChain
	}
	txHash, err := hex.DecodeString(resp.TxResponse.TxHash)
	if err != nil {
		log.CtxErrorw(ctx, "failed to marshal tx hash", "err", err, "global_virtual_group", msgCreateGlobalVirtualGroup.String())
		return nil, ErrCreateGVGOnChain
	}

	// update nonce when tx is successful submitted
	client.operatorAccNonce = nonce + 1
	return txHash, nil
}

func (client *GreenfieldChainSignClient) CompleteMigrateBucket(ctx context.Context, scope SignType, completeMigration *storagetypes.MsgCompleteMigrateBucket) ([]byte, error) {
	log.Infow("signer start to Complete migrate bucket", "scope", scope)
	if completeMigration == nil {
		log.CtxErrorw(ctx, "complete migration msg pointer dangling")
		return nil, ErrDanglingPointer
	}

	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return nil, ErrSignMsg
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	nonce := client.operatorAccNonce

	msgCompleteMigrateBucket := &storagetypes.MsgCompleteMigrateBucket{
		Operator:                   km.GetAddr().String(),
		BucketName:                 completeMigration.BucketName,
		GlobalVirtualGroupFamilyId: completeMigration.GlobalVirtualGroupFamilyId,
		GvgMappings:                completeMigration.GvgMappings,
	}
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &ctypes.TxOption{
		Mode:      &mode,
		GasLimit:  client.gasInfo[CompleteMigration].GasLimit,
		FeeAmount: client.gasInfo[CompleteMigration].FeeAmount,
		Nonce:     nonce,
	}

	resp, err := client.greenfieldClients[scope].BroadcastTx(ctx, []sdk.Msg{msgCompleteMigrateBucket}, txOpt)
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast tx", "err", err, "complete_migrate_bucket", msgCompleteMigrateBucket.String())
		if strings.Contains(err.Error(), "account sequence mismatch") {
			// if nonce mismatch, reset nonce by querying the nonce on chain
			nonce, err := client.greenfieldClients[scope].GetNonce()
			if err != nil {
				log.CtxErrorw(ctx, "failed to get approval account nonce", "err", err)
				return nil, ErrCompleteMigrateBucketOnChain
			}
			client.operatorAccNonce = nonce
		}
		return nil, ErrCompleteMigrateBucketOnChain
	}

	if resp.TxResponse.Code != 0 {
		log.CtxErrorw(ctx, "failed to broadcast tx", "code", resp.TxResponse.Code,
			"msg", resp.TxResponse.String(),
			"complete_migrate_bucket", msgCompleteMigrateBucket.String())
		return nil, ErrCompleteMigrateBucketOnChain
	}
	txHash, err := hex.DecodeString(resp.TxResponse.TxHash)
	if err != nil {
		log.CtxErrorw(ctx, "failed to marshal tx hash", "err", err, "complete_migrate_bucket", msgCompleteMigrateBucket.String())
		return nil, ErrCompleteMigrateBucketOnChain
	}

	// update nonce when tx is successful submitted
	client.operatorAccNonce = nonce + 1
	return txHash, nil
}
