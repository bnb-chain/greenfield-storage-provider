package types

import (
	"time"
)

const (
	ArchivePoolAddress        = "0x0000000000000000000000000000000000000000" // the data settling address, a unified standard across WeaveVM archiving services
	WeaveVMMaxTransactionSize = 8_388_608
)

// Config...WeaveVM client configuration
type Config struct {
	Timeout  time.Duration `json:"timeout,omitempty"`
	ChainID  int64         `json:"chain_id,omitempty"`
	Endpoint string        `json:"endpoint,omitempty"`

	// Signer config (either private key or web3signer required)
	PrivateKeyHex           string `json:"private_key_hex,omitempty"`
	Web3SignerEndpoint      string `json:"web3_signer_endpoint,omitempty"`
	Web3SignerTLSCertFile   string `json:"web3_signer_tls_cert_file,omitempty"`
	Web3SignerTLSKeyFile    string `json:"web3_signer_tls_key_file,omitempty"`
	Web3SignerTLSCACertFile string `json:"web3_signer_tls_ca_cert_file,omitempty"`

	RetryAttempts int
	RetryDelay    time.Duration
}

type RetrieverResponse struct {
	ArweaveBlockHash   string `json:"arweave_block_hash"`
	Calldata           string `json:"calldata"`
	WarDecodedCalldata string `json:"war_decoded_calldata"`
	WvmBlockHash       string `json:"wvm_block_hash"`
}

type WvmGatewayData struct {
	ArweaveBlockHash string
	WvmBlockHash     string
	WvmTxHash        string
	WvmBlockNumber   uint64
	Blob             []byte
}

type PutObjectInput struct {
	Body     []byte
	Metadata map[string]*string
}
