package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	maxIdleConns   = 20
	idleConnTimout = 60 * time.Second
)

// GatewayClient is a http client wrapper
type GatewayClient struct {
	address    string
	httpClient *http.Client
}

// NewGatewayClient return a gateway grpc client instance, and use http://ip:port or http://domain_name as address
func NewGatewayClient(address string) (*GatewayClient, error) {
	tr := &http.Transport{
		MaxIdleConns:    maxIdleConns,
		IdleConnTimeout: idleConnTimout,
	}
	client := &GatewayClient{
		address:    address,
		httpClient: &http.Client{Transport: tr},
	}
	return client, nil
}

func (gatewayClient *GatewayClient) SyncPieceData(objectInfo *types.ObjectInfo, replicateIdx uint32, segmentSize uint32,
	pieceData [][]byte) (integrityHash []byte, signature []byte, err error) {
	marshalObjectInfo := hex.EncodeToString(types.ModuleCdc.MustMarshalJSON(objectInfo))
	marshalPieceData, err := json.Marshal(pieceData)
	if err != nil {
		log.Errorw("failed to marshal piece data", "error", err)
		return nil, nil, err
	}

	req, err := http.NewRequest(http.MethodPut, gatewayClient.address+model.SyncerPath, bytes.NewReader(marshalPieceData))
	req.Header.Add(model.GnfdObjectInfoHeader, marshalObjectInfo)
	req.Header.Add(model.GnfdReplicateIdxHeader, util.Uint32ToString(replicateIdx))
	req.Header.Add(model.GnfdSegmentSizeHeader, util.Uint32ToString(segmentSize))
	req.Header.Add(model.ContentTypeHeader, model.OctetStream)

	resp, err := gatewayClient.httpClient.Do(req)
	if err != nil {
		log.Errorw("failed to sync piece data to other sp", "sp_endpoint", gatewayClient.address, "error", err)
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// TODO: get more error info from body
		log.Errorw("failed to sync piece data", "status_code", resp.StatusCode, "sp_endpoint", gatewayClient.address)
		return nil, nil, fmt.Errorf("failed to sync piece")
	}

	integrityHash, err = hex.DecodeString(resp.Header.Get(model.GnfdIntegrityHashHeader))
	if err != nil {
		log.Errorw("failed to parse integrity hash header",
			"integrity_hash", resp.Header.Get(model.GnfdIntegrityHashHeader),
			"sp_endpoint", gatewayClient.address, "error", err)
		return nil, nil, err
	}
	signature, err = hex.DecodeString(resp.Header.Get(model.GnfdIntegrityHashSignatureHeader))
	if err != nil {
		log.Errorw("failed to parse integrity hash signature header",
			"integrity_hash_signature", resp.Header.Get(model.GnfdIntegrityHashSignatureHeader),
			"sp_endpoint", gatewayClient.address, "error", err)
		return nil, nil, err
	}
	return integrityHash, signature, nil
}
