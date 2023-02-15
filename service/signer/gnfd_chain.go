package signer

import (
	"context"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield/app"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	clitx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
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

	greenfieldClients   []*GreenfieldClient
	greenfieldClientIdx int
	config              *GreenfieldChainConfig
	privKeys            map[SignType]*ethsecp256k1.PrivKey

	wg     sync.WaitGroup
	stopCh chan struct{}
}

func getPrivKey(priv string) (*ethsecp256k1.PrivKey, error) {
	privKey, err := HexToEthSecp256k1PrivKey(priv)
	if err != nil {
		log.Panic(err)
	}

	return privKey, nil
}

// NewGreenfieldChainClient return the GreenfieldChainClient instance
func NewGreenfieldChainClient(config *GreenfieldChainConfig) (*GreenfieldChainClient, error) {
	operatorPrivKey, err := getPrivKey(config.OperatorPrivateKey)
	if err != nil {
		return nil, err
	}
	fundingPrivKey, err := getPrivKey(config.FundingPrivateKey)
	if err != nil {
		return nil, err
	}
	sealPrivKey, err := getPrivKey(config.SealPrivateKey)
	if err != nil {
		return nil, err
	}
	approvalPrivKey, err := getPrivKey(config.ApprovalPrivateKey)
	if err != nil {
		return nil, err
	}
	privKeys := map[SignType]*ethsecp256k1.PrivKey{
		SignOperator: operatorPrivKey,
		SignFunding:  fundingPrivKey,
		SignSeal:     sealPrivKey,
		SignApproval: approvalPrivKey,
	}

	cli := &GreenfieldChainClient{
		config:              config,
		privKeys:            privKeys,
		greenfieldClientIdx: 0,
		greenfieldClients:   initGreenfieldClients(config.RPCAddrs, config.GRPCAddrs),

		wg:     sync.WaitGroup{},
		stopCh: make(chan struct{}),
	}
	return cli, nil
}

func (client *GreenfieldChainClient) Sign(scope SignType, msg []byte) ([]byte, error) {
	return client.privKeys[scope].Sign(msg)
}

// SealObject seal the object on the greenfield chain.
func (client *GreenfieldChainClient) SealObject(ctx context.Context, scope SignType, object *ptypes.ObjectInfo) ([]byte, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	encodingConfig := app.MakeEncodingConfig()
	txConfig := encodingConfig.TxConfig
	txBuilder := txConfig.NewTxBuilder()
	addr := client.privKeys[scope].PubKey().Address().String()

	acct, err := client.greenfieldClients[client.greenfieldClientIdx].GetAccount(addr)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get account: "+addr, "err", err)
		return nil, merrors.ErrGetAccount
	}

	var (
		secondarySPAccs       = make([]types.AccAddress, 0, len(object.SecondarySps))
		secondarySpSignatures = make([][]byte, 0, len(object.SecondarySps))
	)

	for _, sp := range object.SecondarySps {
		secondarySPAccs = append(secondarySPAccs, types.AccAddress(sp.SpId))
		secondarySpSignatures = append(secondarySpSignatures, sp.Signature)
	}

	msgSealObject := storagetypes.NewMsgSealObject(types.AccAddress(client.privKeys[scope].PubKey().Address().Bytes()),
		object.BucketName, object.ObjectName, secondarySPAccs, secondarySpSignatures)

	err = txBuilder.SetMsgs(msgSealObject)
	if err != nil {
		log.CtxErrorw(ctx, "failed to build tx", "err", err)
		return nil, merrors.ErrSealObjectTx
	}
	txBuilder.SetGasLimit(client.config.GasLimit)

	sig := signing.SignatureV2{
		PubKey: client.privKeys[scope].PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_EIP_712,
		},
		Sequence: acct.GetSequence(),
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		log.CtxErrorw(ctx, "failed to set sig", "err", err)
		return nil, merrors.ErrSealObjectTx
	}

	sig = signing.SignatureV2{}

	signerData := xauthsigning.SignerData{
		ChainID:       client.config.ChainIdString,
		AccountNumber: acct.GetAccountNumber(),
		Sequence:      acct.GetSequence(),
	}

	sig, err = clitx.SignWithPrivKey(signing.SignMode_SIGN_MODE_EIP_712,
		signerData,
		txBuilder,
		client.privKeys[scope],
		txConfig,
		acct.GetSequence(),
	)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sig with private key", "err", err)
		return nil, merrors.ErrSealObjectTx
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		log.CtxErrorw(ctx, "failed to set sig", "err", err)
		return nil, merrors.ErrSealObjectTx
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		log.CtxErrorw(ctx, "failed to encode tx", "err", err)
		return nil, merrors.ErrSealObjectTx
	}

	resp, err := client.greenfieldClients[client.greenfieldClientIdx].txClient.BroadcastTx(
		ctx,
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: txBytes,
		})
	if err != nil {
		log.CtxErrorw(ctx, "failed to broadcast tx", "err", err)
		return nil, merrors.ErrSealObjectOnChain
	}

	if resp.TxResponse.Code != 0 {
		log.CtxErrorf(ctx, "failed to broadcast tx, resp code: %d", resp.TxResponse.Code)
		return nil, merrors.ErrSealObjectOnChain
	}
	object.TxHash, err = resp.GetTxResponse().Marshal()
	if err != nil {
		log.CtxErrorw(ctx, "failed to marshal tx hash", "err", err)
		return nil, merrors.ErrSealObjectOnChain
	}

	return object.TxHash, nil
}

func (c *GreenfieldChainClient) getLatestBlockHeight(ctx context.Context, client rpcclient.Client) (uint64, error) {
	status, err := client.Status(ctx)
	if err != nil {
		return 0, err
	}
	return uint64(status.SyncInfo.LatestBlockHeight), nil
}

func (c *GreenfieldChainClient) getLatestBlockHeightWithRetry(client rpcclient.Client) (latestHeight uint64, err error) {
	return latestHeight, retry.Do(func() error {
		latestHeightQueryCtx, cancelLatestHeightQueryCtx := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelLatestHeightQueryCtx()
		var err error
		latestHeight, err = c.getLatestBlockHeight(latestHeightQueryCtx, client)
		return err
	}, RtyAttem,
		RtyDelay,
		RtyErr,
		retry.OnRetry(func(n uint, err error) {
			log.Infof("failed to query latest height, attempt: %d times, max_attempts: %d", n+1, RtyAttNum)
		}))
}

func (client *GreenfieldChainClient) updateClientLoop() {
	ticker := time.NewTicker(SleepIntervalForUpdateClient)

	defer func() {
		ticker.Stop()
		client.wg.Done()
	}()
	for {
		select {
		case <-client.stopCh:
			log.Info("stop to monitor greenfield data-seeds healthy")
			return
		case <-ticker.C:
			log.Infof("start to monitor greenfield data-seeds healthy")
			for _, greenfieldClient := range client.greenfieldClients {

				height, err := client.getLatestBlockHeightWithRetry(greenfieldClient.rpcClient)
				if err != nil {
					log.Errorf("get latest block height error, err=%s", err.Error())
					continue
				}
				greenfieldClient.Height = height
				greenfieldClient.UpdatedAt = time.Now()
			}
			highestHeight := uint64(0)
			highestIdx := 0
			for idx := 0; idx < len(client.greenfieldClients); idx++ {
				if client.greenfieldClients[idx].Height > highestHeight {
					highestHeight = client.greenfieldClients[idx].Height
					highestIdx = idx
				}
			}
			// current GreenfieldClient block sync is fall behind, switch to the GreenfieldClient with the highest block height
			if client.greenfieldClients[client.greenfieldClientIdx].Height+FallBehindThreshold < highestHeight {
				client.mu.Lock()
				client.greenfieldClientIdx = highestIdx
				client.mu.Unlock()
			}
		}
	}
}
