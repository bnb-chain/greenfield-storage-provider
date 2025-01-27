package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	weaveVMtypes "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage/weavevm/types"
)

type Gateway struct {
	config *weaveVMtypes.Config
}

func NewGatewayClient(config *weaveVMtypes.Config) *Gateway {
	return &Gateway{config}
}

const weaveVMGatewayURL = "https://gateway.wvm.dev/v1/calldata/%s"

// Modified get function with improved error handling
func (g *Gateway) RetrieveFromGatewayByTag(ctx context.Context, txHash string) (*weaveVMtypes.WvmGatewayData, error) {
	type WvmRetrieverResponse struct {
		ArweaveBlockHash   string `json:"arweave_block_hash"`
		Calldata           string `json:"calldata"`
		WarDecodedCalldata string `json:"war_decoded_calldata"`
		WvmBlockHash       string `json:"weavevm_block_hash"`
		WvmBlockNumber     uint64 `json:"wvm_block_id"`
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf(weaveVMGatewayURL,
			txHash), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	r.Header.Set("Accept", "application/json")
	client := &http.Client{
		Timeout: g.config.Timeout,
	}

	log.Debugw("sending request to WeaveVM data retriever",
		"url", r.URL.String(),
		"headers", r.Header)

	resp, err := client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to call weaveVM-data-retriever: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := validateResponse(resp, body); err != nil {
		log.Errorw("invalid response from WeaveVM data retriever",
			"status", resp.Status,
			"content_type", resp.Header.Get("Content-Type"),
			"body", string(body))
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	var weaveVMData WvmRetrieverResponse
	if err = json.Unmarshal(body, &weaveVMData); err != nil {
		log.Errorw("failed to unmarshal response",
			"error", err,
			"body", string(body),
			"content_type", resp.Header.Get("Content-Type"))
		return nil, fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
	}

	log.Infow("weaveVM backend: get data from weaveVM",
		"arweave_block_hash", weaveVMData.ArweaveBlockHash,
		"weavevm_block_hash", weaveVMData.WvmBlockHash,
		"calldata_length", len(weaveVMData.Calldata))

	blob, err := hexutil.Decode(weaveVMData.Calldata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode calldata: %w", err)
	}

	if len(blob) == 0 {
		return nil, fmt.Errorf("decoded blob has length zero")
	}

	return &weaveVMtypes.WvmGatewayData{ArweaveBlockHash: weaveVMData.ArweaveBlockHash, WvmBlockHash: weaveVMData.WvmBlockHash, WvmTxHash: txHash, Blob: blob}, nil
}

func validateResponse(resp *http.Response, body []byte) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("unexpected content type: %s, body: %s", contentType, string(body))
	}

	return nil
}
