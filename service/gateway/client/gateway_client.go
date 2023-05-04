package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

const (
	// maxIdleConns defines the max idle connections for HTTP server
	maxIdleConns = 20
	// idleConnTimout defines the idle time of connection for closing
	idleConnTimout = 60 * time.Second
)

// GatewayClient is a http client wrapper
type GatewayClient struct {
	address    string
	httpClient *http.Client
}

// NewGatewayClient return a gateway grpc client instance, and use http://ip:port or http://domain_name as address
func NewGatewayClient(address string) (*GatewayClient, error) {
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "https://" + address
	}
	client := &GatewayClient{
		address: address,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    maxIdleConns,
				IdleConnTimeout: idleConnTimout,
			}},
	}
	return client, nil
}

// ReplicateObjectPieceStream replicates object piece stream to the target storage-provider.
func (client *GatewayClient) ReplicateObjectPieceStream(objectID uint64, pieceSize uint32, contentLength int64, redundancyIdx uint32,
	approval *p2ptypes.GetApprovalResponse, objectDataReader io.Reader) (integrityHash []byte, signature []byte, err error) {
	req, err := http.NewRequest(http.MethodPut, client.address+model.ReplicateObjectPiecePath, objectDataReader)
	if err != nil {
		log.Errorw("failed to replicate piece stream due to new request error", "error", err)
		return nil, nil, err
	}
	req.ContentLength = contentLength
	marshalApproval, err := json.Marshal(approval)
	if err != nil {
		log.Errorw("failed to proto marshal approval", "error", err)
		return
	}
	req.Header.Add(model.GnfdObjectIDHeader, util.Uint64ToString(objectID))
	req.Header.Add(model.GnfdRedundancyIndexHeader, util.Uint32ToString(redundancyIdx))
	req.Header.Add(model.GnfdPieceSizeHeader, util.Uint32ToString(pieceSize))
	req.Header.Add(model.GnfdReplicateApproval, string(marshalApproval))
	req.Header.Add(model.ContentTypeHeader, model.OctetStream)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		log.Errorw("failed to replicate piece stream to other sp", "sp_endpoint", client.address, "error", err)
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorw("failed to replicate piece stream", "status_code", resp.StatusCode, "sp_endpoint", client.address)
		return nil, nil, fmt.Errorf("failed to replicate piece")
	}

	integrityHash, err = hex.DecodeString(resp.Header.Get(model.GnfdIntegrityHashHeader))
	if err != nil {
		log.Errorw("failed to parse integrity hash header",
			"integrity_hash", resp.Header.Get(model.GnfdIntegrityHashHeader),
			"sp_endpoint", client.address, "error", err)
		return nil, nil, err
	}
	signature, err = hex.DecodeString(resp.Header.Get(model.GnfdIntegrityHashSignatureHeader))
	if err != nil {
		log.Errorw("failed to parse integrity hash signature header",
			"integrity_hash_signature", resp.Header.Get(model.GnfdIntegrityHashSignatureHeader),
			"sp_endpoint", client.address, "error", err)
		return nil, nil, err
	}
	return integrityHash, signature, nil
}
