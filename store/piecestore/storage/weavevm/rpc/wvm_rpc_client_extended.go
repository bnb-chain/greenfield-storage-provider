package rpc

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// WvmTransactionRequest represents the payload for the custom method "eth_sendWvmTransaction".
// The "tx" field must be a signed transaction (in hex string format),
// and "tags" is an optional array of string pairs for indexing.
type WvmTransactionRequest struct {
	Tx   string      `json:"tx"`             // raw signed transaction hex
	Tags [][2]string `json:"tags,omitempty"` // e.g. [["Weavm", "v1.0"]]
}

// GetWvmTransactionByTagRequest represents the payload for "eth_getWvmTransactionByTag".
// The "tag" field is required and is expected to be a 2-element string array.
type GetWvmTransactionByTagRequest struct {
	Tag [2]string `json:"tag"`
}

// SendWvmTransaction prepares and sends a WeaveVM transaction using the custom RPC
// method "eth_sendWvmTransaction". It uses your existing logic to estimate gas,
// create a raw (signed) transaction, then wraps it in a WvmTransactionRequest with the provided tags.
// The RPC call returns the transaction hash as a string.
func (rpc *RPCClient) SendWvmTransaction(ctx context.Context, to string, data []byte, tags [][2]string) (string, error) {
	gas, err := rpc.estimateGas(ctx, to, data)
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas: %w", err)
	}

	rawTx, err := rpc.createRawTransaction(ctx, to, string(data), gas)
	if err != nil {
		return "", fmt.Errorf("failed to create raw transaction: %w", err)
	}

	req := &WvmTransactionRequest{
		Tx:   rawTx,
		Tags: tags,
	}

	var txHash string
	// Use CallContext to invoke the custom RPC method "eth_sendWvmTransaction"
	// The call expects the request payload and returns a transaction hash.
	if err := rpc.clientExtended.CallContext(ctx, &txHash, "eth_sendWvmTransaction", req); err != nil {
		return "", fmt.Errorf("failed to send weave transaction: %w", err)
	}

	log.Infow("weaveVM: successfully sent custom transaction", "txHash", txHash)
	return txHash, nil
}

// GetWvmTransactionByTag retrieves a transaction by tag and returns it as a typed Transaction
func (rpc *RPCClient) GetWvmTransactionByTag(ctx context.Context, tag [2]string) (*ethtypes.Transaction, error) {
	rawTx, err := rpc.GetWvmTransactionByTagRaw(ctx, tag)
	if err != nil {
		return nil, err
	}

	if len(rawTx) == 0 {
		return nil, nil
	}

	var tx ethtypes.Transaction
	if err := rlp.DecodeBytes(rawTx, &tx); err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	return &tx, nil
}

// GetWvmTransactionByTag retrieves a transaction by tag using the custom RPC method "eth_getWvmTransactionByTag".
func (rpc *RPCClient) GetWvmTransactionByTagRaw(ctx context.Context, tag [2]string) ([]byte, error) {
	req := &GetWvmTransactionByTagRequest{Tag: tag}

	var result hexutil.Bytes
	if err := rpc.clientExtended.CallContext(ctx, &result, "eth_getWvmTransactionByTagRaw", req); err != nil {
		return nil, fmt.Errorf("failed to get transaction by tag: %w", err)
	}

	if len(result) == 0 {
		log.Infow("weaveVM: no transaction found for tag", "tag", tag)
	}

	return result, nil
}
