package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	weaveVMtypes "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage/weavevm/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type WeaveVM interface {
	SendTransaction(ctx context.Context, to string, data []byte) (string, error)
	SendWeaveTransaction(ctx context.Context, to string, data []byte, tag string) (string, error)
	GetTransactionReceipt(ctx context.Context, txHash string) (*ethtypes.Receipt, error)
	GetTransactionByHash(ctx context.Context, txHash string) (*ethtypes.Transaction, bool, error)
}

type WeaveGateway interface {
	RetrieveFromGateway(ctx context.Context, txHash string) (interface{}, error)
	RetrieveFromGatewayByTag(ctx context.Context, tag string) (*weaveVMtypes.WvmGatewayData, error)
}

type weavevmStore struct {
	client  WeaveVM
	gateway WeaveGateway
}

func newWeaveVMmStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	return &weavevmStore{}, nil
}

// show address on weavevm?
func (s *weavevmStore) String() string {
	return fmt.Sprintf("weavevm")
}

// TODO: 	return nil, ErrUnsupportedMethod ?
func (s *weavevmStore) CreateBucket(ctx context.Context) error {
	return ErrUnsupportedMethod
}

func (s *weavevmStore) GetObject(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
	if key == "" {
		return nil, ErrInvalidObjectKey
	}

	resp, err := s.gateway.RetrieveFromGatewayByTag(ctx, key)
	if err != nil {
		log.Errorw("weavevm failed to get object", "error", err)
		return nil, err
	}

	var objData weaveVMtypes.PutObjectInput
	if err := json.Unmarshal(resp.Blob, &objData); err != nil {
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

func (s *weavevmStore) waitForTxReceipt(ctx context.Context, txHash string) (*ethtypes.Receipt, error) {
	var receipt *ethtypes.Receipt
	// err := retry.Do(
	// 	func() error {
	// 		var err error
	// 		receipt, err = c.client.GetTransactionReceipt(ctx, txHash)
	// 		if err != nil {
	// 			// Mark network/temporary errors as retryable
	// 			return fmt.Errorf("get receipt failed: %w", err)
	// 		}
	// 		if receipt == nil {
	// 			// Receipt not found yet - this is retryable
	// 			return fmt.Errorf("receipt not found")
	// 		}
	// 		if receipt.BlockNumber == nil || receipt.BlockNumber.Cmp(big.NewInt(0)) == 0 {
	// 			return fmt.Errorf("no block number in receipt")
	// 		}
	// 		return nil
	// 	},
	// 	retry.Context(ctx),
	// 	retry.Attempts(uint(*c.config.RetryAttempts)),
	// 	retry.Delay(c.config.RetryDelay),
	// 	retry.DelayType(retry.FixedDelay), // Force fixed delay between attempts
	// 	retry.LastErrorOnly(true),         // Only log the last error
	// 	retry.OnRetry(func(n uint, err error) {
	// 		c.logger.Debug("waiting for receipt",
	// 			"txHash", txHash,
	// 			"attempt", n,
	// 			"error", err)
	// 	}),
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get receipt after %d attempts: %w",
	// 		*c.config.RetryAttempts, err)
	// }

	// return receipt, nil
	return receipt, nil
}

// TODO: 	return nil, ErrUnsupportedMethod ?
func (s *weavevmStore) DeleteObject(ctx context.Context, key string) error {
	log.Debugw("DeleteObject on WeaveVM store - data remains permanently available", "key", key)
	return nil
}

// TODO: 	return nil, ErrUnsupportedMethod ?
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
