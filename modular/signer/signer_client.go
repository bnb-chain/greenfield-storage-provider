package signer

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	ctypes "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/common"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// SignType is the type of msg signature
type SignType string

// GasInfoType is the type of gas info
type GasInfoType string

const (
	// SignOperator is the type of signature signed by the operator account
	SignOperator SignType = "operator"

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
	CompleteMigrateBucket    GasInfoType = "CompleteMigrateBucket"
	SwapOut                  GasInfoType = "SwapOut"
	CompleteSwapOut          GasInfoType = "CompleteSwapOut"
	SPExit                   GasInfoType = "SPExit"
	CompleteSPExit           GasInfoType = "CompleteSPExit"
	UpdateSPPrice            GasInfoType = "UpdateSPPrice"
)

type GasInfo struct {
	GasLimit  uint64
	FeeAmount sdk.Coins
}

// GreenfieldChainSignClient the greenfield chain client
type GreenfieldChainSignClient struct {
	signer *SignModular

	opLock   sync.Mutex
	sealLock sync.Mutex
	gcLock   sync.Mutex

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
	operatorAccNonce, err := operatorClient.GetNonce(context.Background())
	if err != nil {
		log.Errorw("failed to get operator nonce", "error", err)
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
	sealAccNonce, err := sealClient.GetNonce(context.Background())
	if err != nil {
		log.Errorw("failed to get seal nonce", "error", err)
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
		log.Errorw("failed to new gc private key manager", "error", err)
		return nil, err
	}
	gcClient, err := client.NewGreenfieldClient(rpcAddr, chainID, client.WithKeyManager(gcKM))
	if err != nil {
		log.Errorw("failed to new gc greenfield client", "error", err)
		return nil, err
	}
	gcAccNonce, err := gcClient.GetNonce(context.Background())
	if err != nil {
		log.Errorw("failed to get gc nonce", "error", err)
		return nil, err
	}

	greenfieldClients := map[SignType]*client.GreenfieldClient{
		SignOperator: operatorClient,
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
	sealObject *storagetypes.MsgSealObject) (string, error) {
	if sealObject == nil {
		log.CtxError(ctx, "failed to seal object due to pointer dangling")
		return "", ErrDanglingPointer
	}
	ctx = log.WithValue(ctx, log.CtxKeyBucketName, sealObject.GetBucketName())
	ctx = log.WithValue(ctx, log.CtxKeyObjectName, sealObject.GetObjectName())
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.sealLock.Lock()
	defer client.sealLock.Unlock()

	msgSealObject := storagetypes.NewMsgSealObject(km.GetAddr(),
		sealObject.GetBucketName(), sealObject.GetObjectName(), sealObject.GetGlobalVirtualGroupId(),
		sealObject.GetSecondarySpBlsAggSignatures())

	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash   string
		nonce    uint64
		nonceErr error
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
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get seal account nonce", "error", nonceErr)
				ErrSealObjectOnChain.SetError(fmt.Errorf("failed to get seal account nonce, error: %v", nonceErr))
				return "", ErrSealObjectOnChain
			}
			client.sealAccNonce = nonce
		}

		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast seal object tx", "retry_number", i, "error", err)
			continue
		}
		client.sealAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast seal object tx", "tx_hash", txHash, "seal_msg", msgSealObject)
		return txHash, nil
	}

	// failed to broadcast tx
	ErrSealObjectOnChain.SetError(fmt.Errorf("failed to broadcast seal object tx, error: %v", err))
	return "", ErrSealObjectOnChain
}

// RejectUnSealObject reject seal object on the greenfield chain.
func (client *GreenfieldChainSignClient) RejectUnSealObject(ctx context.Context, scope SignType,
	rejectObject *storagetypes.MsgRejectSealObject) (string, error) {
	if rejectObject == nil {
		log.CtxError(ctx, "failed to reject unseal object due to pointer dangling")
		return "", ErrDanglingPointer
	}
	ctx = log.WithValue(ctx, log.CtxKeyBucketName, rejectObject.GetBucketName())
	ctx = log.WithValue(ctx, log.CtxKeyObjectName, rejectObject.GetObjectName())
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.sealLock.Lock()
	defer client.sealLock.Unlock()

	msgRejectUnSealObject := storagetypes.NewMsgRejectUnsealedObject(km.GetAddr(), rejectObject.GetBucketName(), rejectObject.GetObjectName())
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash   string
		nonce    uint64
		nonceErr error
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
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get seal account nonce", "error", nonceErr)
				ErrRejectUnSealObjectOnChain.SetError(fmt.Errorf("failed to get seal account nonce, error: %v", nonceErr))
				return "", ErrRejectUnSealObjectOnChain
			}
			client.sealAccNonce = nonce
		}

		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast reject unseal object", "retry_number", i, "error", err)
			continue
		}

		client.sealAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast reject unseal object tx", "tx_hash", txHash)
		return txHash, nil
	}
	// failed to broadcast tx
	ErrRejectUnSealObjectOnChain.SetError(fmt.Errorf("failed to broadcast reject unseal object tx, error: %v", err))
	return "", ErrRejectUnSealObjectOnChain
}

// DiscontinueBucket stops serving the bucket on the greenfield chain.
func (client *GreenfieldChainSignClient) DiscontinueBucket(ctx context.Context, scope SignType, discontinueBucket *storagetypes.MsgDiscontinueBucket) (string, error) {
	log.Infow("signer start to discontinue bucket", "scope", scope)
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "err", err)
		return "", ErrSignMsg
	}

	client.gcLock.Lock()
	defer client.gcLock.Unlock()
	nonce := client.gcAccNonce

	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(km.GetAddr(),
		discontinueBucket.BucketName, discontinueBucket.Reason)
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &ctypes.TxOption{ // allow simulation here to save gas cost
		Mode:  &mode,
		Nonce: nonce,
	}

	txHash, err := client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgDiscontinueBucket}, txOpt)
	if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
		// if nonce mismatch, wait for next block, reset nonce by querying the nonce on chain
		nonce, nonceErr := client.getNonceOnChain(ctx, client.greenfieldClients[scope])
		if nonceErr != nil {
			log.CtxErrorw(ctx, "failed to get gc account nonce", "error", nonceErr)
			ErrDiscontinueBucketOnChain.SetError(fmt.Errorf("failed to get gc account nonce, error: %v", nonceErr))
			return "", ErrDiscontinueBucketOnChain
		}
		client.gcAccNonce = nonce
	}

	// failed to broadcast tx
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast discontinue bucket", "error", err, "discontinue_bucket", msgDiscontinueBucket.String())
		ErrDiscontinueBucketOnChain.SetError(fmt.Errorf("failed to broadcast discontinue bucket, error: %v", err))
		return "", ErrDiscontinueBucketOnChain
	}
	// update nonce when tx is successful submitted
	client.gcAccNonce = nonce + 1
	return txHash, nil
}

func (client *GreenfieldChainSignClient) CreateGlobalVirtualGroup(ctx context.Context, scope SignType,
	gvg *virtualgrouptypes.MsgCreateGlobalVirtualGroup) (string, error) {
	log.Infow("signer starts to create a new global virtual group", "scope", scope)
	if gvg == nil {
		log.CtxError(ctx, "failed to create virtual group due to pointer dangling")
		return "", ErrDanglingPointer
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.opLock.Lock()
	defer client.opLock.Unlock()

	msgCreateGlobalVirtualGroup := virtualgrouptypes.NewMsgCreateGlobalVirtualGroup(km.GetAddr(),
		gvg.FamilyId, gvg.GetSecondarySpIds(), gvg.GetDeposit())
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash   string
		nonce    uint64
		nonceErr error
	)
	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.operatorAccNonce
		txOpt := &ctypes.TxOption{
			Mode:      &mode,
			GasLimit:  client.gasInfo[CreateGlobalVirtualGroup].GasLimit,
			FeeAmount: client.gasInfo[CreateGlobalVirtualGroup].FeeAmount,
			Nonce:     nonce,
		}
		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgCreateGlobalVirtualGroup}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatches, waiting for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get operator account nonce", "error", nonceErr)
				ErrCreateGVGOnChain.SetError(fmt.Errorf("failed to get approval account nonce, error: %v", err))
				return "", ErrCreateGVGOnChain
			}
			client.operatorAccNonce = nonce
		}
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast global virtual group tx", "global_virtual_group",
				msgCreateGlobalVirtualGroup.String(), "retry_number", i, "error", err)
			continue
		}
		client.operatorAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast create virtual group tx", "tx_hash", txHash,
			"virtual_group_msg", msgCreateGlobalVirtualGroup)
		return txHash, nil

	}

	// failed to broadcast tx
	ErrCreateGVGOnChain.SetError(fmt.Errorf("failed to broadcast create virtual group tx, error: %v", err))
	return "", ErrCreateGVGOnChain
}

func (client *GreenfieldChainSignClient) CompleteMigrateBucket(ctx context.Context, scope SignType,
	migrateBucket *storagetypes.MsgCompleteMigrateBucket) (string, error) {
	log.Infow("signer starts to complete migrate bucket", "scope", scope)
	if migrateBucket == nil {
		log.CtxError(ctx, "complete migrate bucket msg pointer dangling")
		return "", ErrDanglingPointer
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.opLock.Lock()
	defer client.opLock.Unlock()

	msgCompleteMigrateBucket := storagetypes.NewMsgCompleteMigrateBucket(km.GetAddr(), migrateBucket.GetBucketName(),
		migrateBucket.GetGlobalVirtualGroupFamilyId(), migrateBucket.GetGvgMappings())

	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash   string
		nonce    uint64
		nonceErr error
	)
	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.operatorAccNonce
		txOpt := &ctypes.TxOption{
			Mode:      &mode,
			GasLimit:  client.gasInfo[CompleteMigrateBucket].GasLimit,
			FeeAmount: client.gasInfo[CompleteMigrateBucket].FeeAmount,
			Nonce:     nonce,
		}
		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgCompleteMigrateBucket}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatches, waiting for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get operator account nonce", "error", err)
				ErrCompleteMigrateBucketOnChain.SetError(fmt.Errorf("failed to get operator account nonce, error: %v", err))
				return "", ErrCompleteMigrateBucketOnChain
			}
			client.operatorAccNonce = nonce
		}
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast complete migrate bucket tx", "retry_number", i, "error", err)
			continue
		}
		client.operatorAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast complete migrate bucket tx", "tx_hash", txHash, "seal_msg", msgCompleteMigrateBucket)
		return txHash, nil
	}

	// failed to broadcast tx
	ErrCompleteMigrateBucketOnChain.SetError(fmt.Errorf("failed to broadcast complete migrate bucket, error: %v", err))
	return "", ErrCompleteMigrateBucketOnChain
}

func (client *GreenfieldChainSignClient) UpdateSPPrice(ctx context.Context, scope SignType,
	priceInfo *sptypes.MsgUpdateSpStoragePrice) (string, error) {
	log.Infow("signer starts to complete update SP price info", "scope", scope)
	if priceInfo == nil {
		log.CtxError(ctx, "complete migrate bucket msg pointer dangling")
		return "", ErrDanglingPointer
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.opLock.Lock()
	defer client.opLock.Unlock()
	nonce := client.operatorAccNonce

	msgUpdateStorageSPPrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     km.GetAddr().String(),
		ReadPrice:     priceInfo.ReadPrice,
		FreeReadQuota: priceInfo.FreeReadQuota,
		StorePrice:    priceInfo.StorePrice,
	}
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &ctypes.TxOption{
		Mode:      &mode,
		GasLimit:  client.gasInfo[UpdateSPPrice].GasLimit,
		FeeAmount: client.gasInfo[UpdateSPPrice].FeeAmount,
		Nonce:     nonce,
	}

	txHash, err := client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgUpdateStorageSPPrice}, txOpt)
	if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
		// if nonce mismatches, waiting for next block, reset nonce by querying the nonce on chain
		nonce, nonceErr := client.getNonceOnChain(ctx, client.greenfieldClients[scope])
		if nonceErr != nil {
			log.CtxErrorw(ctx, "failed to get approval account nonce", "error", err)
			ErrUpdateSPPriceOnChain.SetError(fmt.Errorf("failed to get approval account nonce, error: %v", err))
			return "", ErrUpdateSPPriceOnChain
		}
		client.operatorAccNonce = nonce
	}
	// failed to broadcast tx
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast update sp price msg", "error", err, "update_sp_price",
			msgUpdateStorageSPPrice.String())
		ErrUpdateSPPriceOnChain.SetError(fmt.Errorf("failed to broadcast msg to update sp price, error: %v", err))
		return "", ErrUpdateSPPriceOnChain
	}

	// update nonce when tx is successfully submitted
	client.operatorAccNonce = nonce + 1
	return txHash, nil
}

func (client *GreenfieldChainSignClient) SwapOut(ctx context.Context, scope SignType,
	swapOut *virtualgrouptypes.MsgSwapOut) (string, error) {
	log.Infow("signer starts to swap out", "scope", scope)
	if swapOut == nil {
		log.CtxError(ctx, "failed to swap out due to pointer dangling")
		return "", ErrDanglingPointer
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.opLock.Lock()
	defer client.opLock.Unlock()

	msgSwapOut := virtualgrouptypes.NewMsgSwapOut(km.GetAddr(), swapOut.GetGlobalVirtualGroupFamilyId(), swapOut.GetGlobalVirtualGroupIds(),
		swapOut.GetSuccessorSpId())
	msgSwapOut.SuccessorSpApproval = &common.Approval{
		ExpiredHeight: swapOut.SuccessorSpApproval.GetExpiredHeight(),
		Sig:           swapOut.SuccessorSpApproval.GetSig(),
	}
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash   string
		nonce    uint64
		nonceErr error
	)

	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.operatorAccNonce
		txOpt := &ctypes.TxOption{
			Mode:      &mode,
			GasLimit:  client.gasInfo[SwapOut].GasLimit,
			FeeAmount: client.gasInfo[SwapOut].FeeAmount,
			Nonce:     nonce,
		}
		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgSwapOut}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatches, waiting for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get operator account nonce", "error", nonceErr)
				ErrSwapOutOnChain.SetError(fmt.Errorf("failed to get operator account nonce, error: %v", err))
				return "", ErrSwapOutOnChain
			}
			client.operatorAccNonce = nonce
		}
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast swap out", "retry_number", i, "swap_out", msgSwapOut.String(), "error", err)
			continue
		}
		client.operatorAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast start swap out tx", "tx_hash", txHash, "swap_out", msgSwapOut.String())
		return txHash, nil

	}

	// failed to broadcast tx
	ErrSwapOutOnChain.SetError(fmt.Errorf("failed to broadcast swap out tx, error: %v", err))
	return "", ErrSwapOutOnChain
}

func (client *GreenfieldChainSignClient) CompleteSwapOut(ctx context.Context, scope SignType,
	completeSwapOut *virtualgrouptypes.MsgCompleteSwapOut) (string, error) {
	log.Infow("signer starts to complete swap out", "scope", scope)
	if completeSwapOut == nil {
		log.CtxError(ctx, "complete swap out msg pointer dangling")
		return "", ErrDanglingPointer
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.opLock.Lock()
	defer client.opLock.Unlock()

	msgCompleteSwapOut := virtualgrouptypes.NewMsgCompleteSwapOut(km.GetAddr(), completeSwapOut.GetGlobalVirtualGroupFamilyId(),
		completeSwapOut.GetGlobalVirtualGroupIds())
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash   string
		nonce    uint64
		nonceErr error
	)

	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.operatorAccNonce
		txOpt := &ctypes.TxOption{
			Mode:      &mode,
			GasLimit:  client.gasInfo[CompleteSwapOut].GasLimit,
			FeeAmount: client.gasInfo[CompleteSwapOut].FeeAmount,
			Nonce:     nonce,
		}
		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgCompleteSwapOut}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatches, waiting for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get operator account nonce", "error", nonceErr)
				ErrCompleteSwapOutOnChain.SetError(fmt.Errorf("failed to get operator account nonce, error: %v", nonceErr))
				return "", ErrCompleteSwapOutOnChain
			}
			client.operatorAccNonce = nonce
		}
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast complete swap out tx", "retry_number", i, "error", err)
			continue
		}
		client.operatorAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast complete swap out tx", "tx_hash", txHash, "seal_msg", msgCompleteSwapOut)
		return txHash, nil
	}

	ErrCompleteSwapOutOnChain.SetError(fmt.Errorf("failed to broadcast complete swap out, error: %v", err))
	return "", ErrCompleteSwapOutOnChain
}

func (client *GreenfieldChainSignClient) SPExit(ctx context.Context, scope SignType,
	spExit *virtualgrouptypes.MsgStorageProviderExit) (string, error) {
	log.Infow("signer starts to sp exit", "scope", scope)
	if spExit == nil {
		log.CtxError(ctx, "sp exit msg pointer dangling")
		return "", ErrDanglingPointer
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.opLock.Lock()
	defer client.opLock.Unlock()

	msgSPExit := virtualgrouptypes.NewMsgStorageProviderExit(km.GetAddr())

	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	var (
		txHash   string
		nonce    uint64
		nonceErr error
	)

	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.operatorAccNonce
		txOpt := &ctypes.TxOption{
			Mode:      &mode,
			GasLimit:  client.gasInfo[SPExit].GasLimit,
			FeeAmount: client.gasInfo[SPExit].FeeAmount,
			Nonce:     nonce,
		}
		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgSPExit}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatches, waiting for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get operator account nonce", "error", nonceErr)
				ErrSPExitOnChain.SetError(fmt.Errorf("failed to get operator account nonce, error: %v", nonceErr))
				return "", ErrSPExitOnChain
			}
			client.operatorAccNonce = nonce
		}
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast start sp exit tx", "retry_number", i, "error", err)
			continue
		}
		client.operatorAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast start sp exit tx", "tx_hash", txHash, "exit_msg", msgSPExit)
		return txHash, nil
	}

	ErrSPExitOnChain.SetError(fmt.Errorf("failed to broadcast start sp exit, error: %v", err))
	return "", ErrSPExitOnChain
}

func (client *GreenfieldChainSignClient) CompleteSPExit(ctx context.Context, scope SignType,
	completeSPExit *virtualgrouptypes.MsgCompleteStorageProviderExit) (string, error) {
	log.Infow("signer starts to complete sp exit", "scope", scope)
	if completeSPExit == nil {
		log.CtxError(ctx, "complete sp exit msg pointer dangling")
		return "", ErrDanglingPointer
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.opLock.Lock()
	defer client.opLock.Unlock()

	msgCompleteSPExit := virtualgrouptypes.NewMsgCompleteStorageProviderExit(km.GetAddr())

	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	var (
		txHash   string
		nonce    uint64
		nonceErr error
	)
	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.operatorAccNonce
		txOpt := &ctypes.TxOption{
			Mode:      &mode,
			GasLimit:  client.gasInfo[CompleteSPExit].GasLimit,
			FeeAmount: client.gasInfo[CompleteSPExit].FeeAmount,
			Nonce:     nonce,
		}
		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgCompleteSPExit}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatches, waiting for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get operator account nonce", "error", nonceErr)
				ErrCompleteSPExitOnChain.SetError(fmt.Errorf("failed to get operator account nonce, error: %v", nonceErr))
				return "", ErrCompleteSPExitOnChain
			}
			client.operatorAccNonce = nonce
		}
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast complete sp exit tx", "retry_number", i, "error", err)
			continue
		}
		client.operatorAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast complete sp exit tx", "tx_hash", txHash, "complete_sp_exit_msg", msgCompleteSPExit)
		return txHash, nil
	}

	ErrCompleteSPExitOnChain.SetError(fmt.Errorf("failed to broadcast complete sp exit, error: %v", err))
	return "", ErrCompleteSPExitOnChain
}

func (client *GreenfieldChainSignClient) RejectMigrateBucket(ctx context.Context, scope SignType,
	msg *storagetypes.MsgRejectMigrateBucket) (string, error) {
	log.Infow("signer starts to reject migrate bucket", "scope", scope)
	if msg == nil {
		log.CtxError(ctx, "reject migrate bucket msg pointer dangling")
		return "", ErrDanglingPointer
	}
	km, err := client.greenfieldClients[scope].GetKeyManager()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get private key", "error", err)
		return "", ErrSignMsg
	}

	client.opLock.Lock()
	defer client.opLock.Unlock()

	msgRejectMigrateBucket := storagetypes.NewMsgRejectMigrateBucket(km.GetAddr(), msg.GetBucketName())

	var (
		txHash   string
		nonce    uint64
		nonceErr error
	)
	for i := 0; i < BroadcastTxRetry; i++ {
		nonce = client.operatorAccNonce
		txOpt := &ctypes.TxOption{
			Nonce: nonce,
		}
		txHash, err = client.broadcastTx(ctx, client.greenfieldClients[scope], []sdk.Msg{msgRejectMigrateBucket}, txOpt)
		if errors.IsOf(err, sdkErrors.ErrWrongSequence) {
			// if nonce mismatches, waiting for next block, reset nonce by querying the nonce on chain
			nonce, nonceErr = client.getNonceOnChain(ctx, client.greenfieldClients[scope])
			if nonceErr != nil {
				log.CtxErrorw(ctx, "failed to get operator account nonce", "error", err)
				ErrRejectMigrateBucketOnChain.SetError(fmt.Errorf("failed to get operator account nonce, error: %v", err))
				return "", ErrRejectMigrateBucketOnChain
			}
			client.operatorAccNonce = nonce
		}
		if err != nil {
			log.CtxErrorw(ctx, "failed to broadcast reject migrate bucket tx", "retry_number", i, "error", err)
			continue
		}
		client.operatorAccNonce = nonce + 1
		log.CtxDebugw(ctx, "succeed to broadcast complete migrate bucket tx", "tx_hash", txHash, "reject_migrate_bucket_msg", msgRejectMigrateBucket)
		return txHash, nil
	}

	// failed to broadcast tx
	ErrRejectMigrateBucketOnChain.SetError(fmt.Errorf("failed to broadcast reject migrate bucket, error: %v", err))
	return "", ErrRejectMigrateBucketOnChain
}

func (client *GreenfieldChainSignClient) getNonceOnChain(ctx context.Context, gnfdClient *client.GreenfieldClient) (uint64, error) {
	err := client.signer.baseApp.Consensus().WaitForNextBlock(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to wait next block", "error", err)
		return 0, err
	}
	nonce, err := gnfdClient.GetNonce(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get nonce on chain", "error", err)
		return 0, err
	}
	return nonce, nil
}

func (client *GreenfieldChainSignClient) broadcastTx(ctx context.Context, gnfdClient *client.GreenfieldClient,
	msgs []sdk.Msg, txOpt *ctypes.TxOption, opts ...grpc.CallOption) (string, error) {
	resp, err := gnfdClient.BroadcastTx(ctx, msgs, txOpt, opts...)
	if err != nil {
		if strings.Contains(err.Error(), "account sequence mismatch") {
			return "", sdkErrors.ErrWrongSequence
		}
		return "", errors.Wrap(err, "failed to broadcast tx with greenfield client")
	}
	if resp.TxResponse.Code == sdkErrors.ErrWrongSequence.ABCICode() {
		return "", sdkErrors.ErrWrongSequence
	}
	if resp.TxResponse.Code != 0 {
		return "", fmt.Errorf("failed to broadcast tx, resp code: %d, code space: %s", resp.TxResponse.Code, resp.TxResponse.Codespace)
	}
	return resp.TxResponse.TxHash, nil
}
