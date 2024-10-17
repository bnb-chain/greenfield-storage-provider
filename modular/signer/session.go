package signer

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evmos/evmos/v12/x/evm/precompiles/storage"
	"github.com/evmos/evmos/v12/x/evm/precompiles/storageprovider"
	"github.com/evmos/evmos/v12/x/evm/precompiles/virtualgroup"
)

func CreateTxOpts(ctx context.Context, client *ethclient.Client, hexPrivateKey string, chain *big.Int, gasLimit uint64, nonce uint64) (*bind.TransactOpts, error) {
	// create private key
	privateKey, err := crypto.HexToECDSA(hexPrivateKey)
	if err != nil {
		return nil, err
	}

	// Build transact tx opts with private key
	txOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, chain)
	if err != nil {
		return nil, err
	}

	// set gas limit and gas price
	txOpts.GasLimit = gasLimit
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	txOpts.GasPrice = gasPrice

	txOpts.Nonce = big.NewInt(int64(nonce))

	return txOpts, nil
}

func CreateStorageSession(client *ethclient.Client, txOpts bind.TransactOpts, contractAddress string) (*storage.IStorageSession, error) {
	storageContract, err := storage.NewIStorage(common.HexToAddress(contractAddress), client)
	if err != nil {
		return nil, err
	}
	session := &storage.IStorageSession{
		Contract: storageContract,
		CallOpts: bind.CallOpts{
			Pending: false,
		},
		TransactOpts: txOpts,
	}
	return session, nil
}

func CreateVirtualGroupSession(client *ethclient.Client, txOpts bind.TransactOpts, contractAddress string) (*virtualgroup.IVirtualGroupSession, error) {
	virtualgroupContract, err := virtualgroup.NewIVirtualGroup(common.HexToAddress(contractAddress), client)
	if err != nil {
		return nil, err
	}
	session := &virtualgroup.IVirtualGroupSession{
		Contract: virtualgroupContract,
		CallOpts: bind.CallOpts{
			Pending: false,
		},
		TransactOpts: txOpts,
	}
	return session, nil
}

func CreateStorageProviderSession(client *ethclient.Client, txOpts bind.TransactOpts, contractAddress string) (*storageprovider.IStorageProviderSession, error) {
	storageproviderContract, err := storageprovider.NewIStorageProvider(common.HexToAddress(contractAddress), client)
	if err != nil {
		return nil, err
	}
	session := &storageprovider.IStorageProviderSession{
		Contract: storageproviderContract,
		CallOpts: bind.CallOpts{
			Pending: false,
		},
		TransactOpts: txOpts,
	}
	return session, nil
}
