package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/avast/retry-go/v4"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	gateway "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage/weavevm/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage/weavevm/rpc"
	signer "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage/weavevm/signer"
	weaveVMtypes "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage/weavevm/types"
)

type WeaveVM interface {
	SendWeaveTransaction(ctx context.Context, to string, data []byte, tag string) (string, error)
	GetTransactionReceipt(ctx context.Context, txHash string) (*ethtypes.Receipt, error)
	GetWvmTransactionByTag(ctx context.Context, tag [2]string) (*ethtypes.Transaction, error)
}

type WeaveGateway interface {
	RetrieveFromGatewayByTag(ctx context.Context, tag string) (*weaveVMtypes.WvmGatewayData, error)
}

type weavevmStore struct {
	client  WeaveVM
	gateway WeaveGateway
}

func getAuthType(cfg ObjectStorageConfig) string {
	if cfg.WeavevmConfig.Web3SignerEndpoint != "" {
		if cfg.WeavevmConfig.Web3SignerTLSCertFile != "" {
			return "web3signer_tls"
		}
		return "web3signer"
	}
	return "private_key"
}

func newWeaveVMStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	var client WeaveVM
	var err error
	if cfg.WeavevmConfig.Web3SignerEndpoint != "" {
		// Initialize with web3signer
		web3signer, err := signer.NewWeb3SignerClient(&cfg.WeavevmConfig)
		if err != nil {
			log.Errorw("failed to initialize web3signer client", "error", err)
			return nil, fmt.Errorf("web3signer init error: %w", err)
		}

		client, err = rpc.NewWvmRPCClient(&cfg.WeavevmConfig, web3signer)
		if err != nil {
			log.Errorw("failed to initialize rpc client with web3signer", "error", err)
			return nil, fmt.Errorf("rpc client init error: %w", err)
		}

	} else if cfg.WeavevmConfig.PrivateKeyHex != "" {
		// Initialize with private key
		privateKeySigner := signer.NewPrivateKeySigner(cfg.WeavevmConfig.PrivateKeyHex, cfg.WeavevmConfig.ChainID)

		client, err = rpc.NewWvmRPCClient(&cfg.WeavevmConfig, privateKeySigner)
		if err != nil {
			log.Errorw("failed to initialize rpc client with private key", "error", err)
			return nil, fmt.Errorf("rpc client init error: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either web3signer endpoint or private key must be provided")
	}

	gateway := gateway.NewGatewayClient(&cfg.WeavevmConfig)

	store := &weavevmStore{
		client:  client,
		gateway: gateway,
	}

	log.Infow("initialized WeaveVM storage",
		"endpoint", cfg.WeavevmConfig.Endpoint,
		"chain_id", cfg.WeavevmConfig.ChainID,
		"auth_type", getAuthType(cfg),
	)

	return store, nil
}

// show address on weavevm?
func (s *weavevmStore) String() string {
	return "weavevm"
}

// TODO: 	return nil, ErrUnsupportedMethod ?
func (s *weavevmStore) CreateBucket(ctx context.Context) error {
	return ErrUnsupportedMethod
}

func (s *weavevmStore) GetObject(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
	if key == "" {
		return nil, ErrInvalidObjectKey
	}

	tag := [2]string{fmt.Sprintf("greenfield:%s", "addrss"), key}
	resp, err := s.client.GetWvmTransactionByTag(ctx, tag)
	if err != nil {
		resp, err := s.gateway.RetrieveFromGatewayByTag(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get object data from weavevm rpc: %w", err)
		}

		if isGatewayTransactionNotFoundErr(resp) {
			return nil, fmt.Errorf("failed to find transaction data in weavevm using gateway")
		}
		return processData(resp.Blob, offset, limit)

	}

	return processData(resp.Data(), offset, limit)
}

func processData(blob []byte, offset, limit int64) (io.ReadCloser, error) {
	var objData weaveVMtypes.PutObjectInput
	if err := json.Unmarshal(blob, &objData); err != nil {
		return nil, err
	}

	if offset == 0 && limit == -1 {
		if cs := objData.Metadata[ChecksumAlgo]; cs != nil {
			return verifyChecksum(io.NopCloser(bytes.NewReader(objData.Body)), aws.StringValue(cs)), nil
		}
	}

	if offset > int64(len(objData.Body)) {
		offset = int64(len(objData.Body))
	}
	data := objData.Body[offset:]
	if limit > 0 && limit < int64(len(data)) {
		data = data[:limit]
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// isGatewayTransactionNotFoundErr checks if the transaction is absent in the Gateway.
// TODO: Gateway indicates a missing transaction by setting WvmBlockHash to "0x".
// it will be fixed in the future
func isGatewayTransactionNotFoundErr(data *weaveVMtypes.WvmGatewayData) bool {
	return data.WvmBlockHash == "0x"
}

// hmmm, key incoming?
func (s *weavevmStore) PutObject(ctx context.Context, key string, reader io.Reader) error {
	var body io.ReadSeeker
	if b, ok := reader.(io.ReadSeeker); ok {
		body = b
	} else {
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	checksum := generateChecksum(body)
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) > weaveVMtypes.WeaveVMMaxTransactionSize {
		return fmt.Errorf("size bigger than maximum blob size: max n bytes: %d", weaveVMtypes.WeaveVMMaxTransactionSize)
	}

	params := &weaveVMtypes.PutObjectInput{
		Body:     data,
		Metadata: map[string]*string{ChecksumAlgo: aws.String(checksum)},
	}

	encodedData, err := json.Marshal(params)
	if err != nil {
		return err
	}

	txHash, err := s.client.SendWeaveTransaction(ctx, weaveVMtypes.ArchivePoolAddress, encodedData, key)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	_, err = s.waitForTxReceipt(ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to get tx receipt: %w", err)
	}

	return nil
}

const (
	WeaveVMReceiptSuccess = "weavevm_receipt_success"
	WeaveVMReceiptFailure = "weavevm_receipt_failure"
)

var (
	// Reuse the same retry configuration pattern as seen in executor.go
	WeavevmRtyAttNum = uint(3)
	WeavevmRtyAttem  = retry.Attempts(WeavevmRtyAttNum)
	WeavevmRtyDelay  = retry.Delay(time.Millisecond * 100)
	WeavevmRtyErr    = retry.LastErrorOnly(true)

	weavevmReceiptTimeout = 10 * time.Second
)

func (s *weavevmStore) waitForTxReceipt(ctx context.Context, txHash string) (*ethtypes.Receipt, error) {
	var (
		receipt *ethtypes.Receipt
		err     error
	)

	err = retry.Do(
		func() error {
			ctxWithTimeout, cancel := context.WithTimeout(ctx, weavevmReceiptTimeout)
			defer cancel()

			receipt, err = s.client.GetTransactionReceipt(ctxWithTimeout, txHash)
			if err != nil {
				return fmt.Errorf("get receipt failed: %w", err)
			}
			if receipt == nil {
				return fmt.Errorf("receipt not found")
			}
			if receipt.BlockNumber == nil || receipt.BlockNumber.Cmp(big.NewInt(0)) == 0 {
				return fmt.Errorf("no block number in receipt")
			}
			return nil
		},
		WeavevmRtyAttem,
		WeavevmRtyDelay,
		WeavevmRtyErr,
		retry.OnRetry(func(n uint, err error) {
			log.CtxDebugw(ctx, "waiting for receipt",
				"txHash", txHash,
				"attempt", n+1,
				"max_attempts", WeavevmRtyAttNum,
				"error", err)
		}),
		retry.Context(ctx),
	)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get receipt after retries",
			"txHash", txHash,
			"attempts", WeavevmRtyAttNum,
			"error", err)
		return nil, fmt.Errorf("failed to get receipt after %d attempts: %w",
			WeavevmRtyAttNum, err)
	}

	return receipt, nil
}

func (s *weavevmStore) DeleteObject(ctx context.Context, key string) error {
	log.Debugw("DeleteObject on WeaveVM store - data remains permanently available", "key", key)
	return nil
}

func (s *weavevmStore) DeleteObjectsByPrefix(ctx context.Context, key string) (uint64, error) {
	log.Debugw("DeleteObjectsByPrefix on WeaveVM store - data remains permanently available", "key", key)
	return 0, nil
}

func (s *weavevmStore) HeadBucket(ctx context.Context) error {
	return nil
}

func (s *weavevmStore) HeadObject(ctx context.Context, key string) (Object, error) {
	if key == "" {
		return nil, ErrInvalidObjectKey
	}

	resp, err := s.gateway.RetrieveFromGatewayByTag(ctx, key)
	if err != nil {
		return nil, err
	}

	var objData weaveVMtypes.PutObjectInput
	if err := json.Unmarshal(resp.Blob, &objData); err != nil {
		return nil, err
	}

	return &object{
		key:  key,
		size: int64(len(objData.Body)),
		// modTime: resp.Timestamp,
		isDir: strings.HasSuffix(key, "/"),
	}, nil
}

func (s *weavevmStore) ListObjects(ctx context.Context, prefix, marker, delimiter string, limit int64) ([]Object, error) {
	return nil, ErrUnsupportedMethod
}

func (s *weavevmStore) ListAllObjects(ctx context.Context, prefix, marker string) (<-chan Object, error) {
	return nil, ErrUnsupportedMethod
}
