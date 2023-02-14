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

// GreenfieldChain the greenfield chain
type GreenfieldChain struct {
	mu sync.Mutex

	greenfieldClients   []*GreenfieldClient
	greenfieldClientIdx int
	config              *GreenfieldChainConfig
	privateKey          *ethsecp256k1.PrivKey

	wg     sync.WaitGroup
	stopCh chan struct{}
}

// NewGreenfieldChain return the GreenfieldChain instance
func NewGreenfieldChain(config *GreenfieldChainConfig) *GreenfieldChain {
	privKey, err := HexToEthSecp256k1PrivKey(config.PrivateKey)
	if err != nil {
		log.Panic(err)
	}
	cli := &GreenfieldChain{
		config:              config,
		privateKey:          privKey,
		greenfieldClientIdx: 0,
		greenfieldClients:   initGreenfieldClients(config.RPCAddrs, config.GRPCAddrs),

		wg:     sync.WaitGroup{},
		stopCh: make(chan struct{}),
	}
	return cli
}

func (cli *GreenfieldChain) Sign(msg []byte) ([]byte, error) {
	return cli.privateKey.Sign(msg)
}

// SealObject seal the object on the greenfield chain.
func (cli *GreenfieldChain) SealObject(ctx context.Context, object *ptypes.ObjectInfo) ([]byte, error) {
	cli.mu.Lock()
	defer cli.mu.Unlock()

	encodingConfig := app.MakeEncodingConfig()
	txConfig := encodingConfig.TxConfig
	txBuilder := txConfig.NewTxBuilder()
	addr := cli.privateKey.PubKey().Address().String()

	acct, err := cli.greenfieldClients[cli.greenfieldClientIdx].GetAccount(addr)
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

	msgSealObject := storagetypes.NewMsgSealObject(types.AccAddress(cli.privateKey.PubKey().Address().Bytes()),
		object.BucketName, object.ObjectName, secondarySPAccs, secondarySpSignatures)

	err = txBuilder.SetMsgs(msgSealObject)
	if err != nil {
		log.CtxErrorw(ctx, "failed to build tx", "err", err)
		return nil, merrors.ErrSealObjectTx
	}
	txBuilder.SetGasLimit(cli.config.GasLimit)

	sig := signing.SignatureV2{
		PubKey: cli.privateKey.PubKey(),
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
		ChainID:       cli.config.ChainIdString,
		AccountNumber: acct.GetAccountNumber(),
		Sequence:      acct.GetSequence(),
	}

	sig, err = clitx.SignWithPrivKey(signing.SignMode_SIGN_MODE_EIP_712,
		signerData,
		txBuilder,
		cli.privateKey,
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

	resp, err := cli.greenfieldClients[cli.greenfieldClientIdx].txClient.BroadcastTx(
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

func (cli *GreenfieldChain) getLatestBlockHeight(ctx context.Context, client rpcclient.Client) (uint64, error) {
	status, err := client.Status(ctx)
	if err != nil {
		return 0, err
	}
	return uint64(status.SyncInfo.LatestBlockHeight), nil
}

func (cli *GreenfieldChain) getLatestBlockHeightWithRetry(client rpcclient.Client) (latestHeight uint64, err error) {
	return latestHeight, retry.Do(func() error {
		latestHeightQueryCtx, cancelLatestHeightQueryCtx := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelLatestHeightQueryCtx()
		var err error
		latestHeight, err = cli.getLatestBlockHeight(latestHeightQueryCtx, client)
		return err
	}, RtyAttem,
		RtyDelay,
		RtyErr,
		retry.OnRetry(func(n uint, err error) {
			log.Infof("failed to query latest height, attempt: %d times, max_attempts: %d", n+1, RtyAttNum)
		}))
}

func (cli *GreenfieldChain) updateClientLoop() {
	ticker := time.NewTicker(SleepIntervalForUpdateClient)

	defer func() {
		ticker.Stop()
		cli.wg.Done()
	}()
	for {
		select {
		case <-cli.stopCh:
			log.Info("stop to monitor greenfield data-seeds healthy")
			return
		case <-ticker.C:
			log.Infof("start to monitor greenfield data-seeds healthy")
			for _, greenfieldClient := range cli.greenfieldClients {

				height, err := cli.getLatestBlockHeightWithRetry(greenfieldClient.rpcClient)
				if err != nil {
					log.Errorf("get latest block height error, err=%s", err.Error())
					continue
				}
				greenfieldClient.Height = height
				greenfieldClient.UpdatedAt = time.Now()
			}
			highestHeight := uint64(0)
			highestIdx := 0
			for idx := 0; idx < len(cli.greenfieldClients); idx++ {
				if cli.greenfieldClients[idx].Height > highestHeight {
					highestHeight = cli.greenfieldClients[idx].Height
					highestIdx = idx
				}
			}
			// current GreenfieldClient block sync is fall behind, switch to the GreenfieldClient with the highest block height
			if cli.greenfieldClients[cli.greenfieldClientIdx].Height+FallBehindThreshold < highestHeight {
				cli.mu.Lock()
				cli.greenfieldClientIdx = highestIdx
				cli.mu.Unlock()
			}
		}
	}
}
