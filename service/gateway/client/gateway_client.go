package client

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/x/storage/types"
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
		address = "http://" + address
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

// PieceDataReader defines [][]pieceData Reader.
type PieceDataReader struct {
	pieceData [][]byte
	outerIdx  int
	innerIdx  int
}

// NewPieceDataReader return a PieceDataReader instance
func NewPieceDataReader(pieceData [][]byte) (reader *PieceDataReader, err error) {
	if len(pieceData) == 0 {
		return nil, fmt.Errorf("failed to new due to invalid args")
	}
	return &PieceDataReader{
		pieceData: pieceData,
		outerIdx:  0,
		innerIdx:  0,
	}, nil
}

// Read populates the given byte slice with data and returns the number of bytes populated and an error value.
// It returns an io.EOF error when the stream ends.
func (p *PieceDataReader) Read(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("failed to read due to invalid args")
	}

	readLen := 0
	for p.outerIdx < len(p.pieceData) {
		curReadLen := copy(buf[readLen:], p.pieceData[p.outerIdx][p.innerIdx:])
		p.innerIdx += curReadLen
		if p.innerIdx == len(p.pieceData[p.outerIdx]) {
			p.outerIdx += 1
			p.innerIdx = 0
		}
		readLen = readLen + curReadLen
		if readLen == len(buf) {
			break
		}
	}
	if readLen != 0 {
		return readLen, nil
	}
	return 0, io.EOF
}

// SyncPieceData sync piece data to the target storage-provider.
func (gatewayClient *GatewayClient) SyncPieceData(
	objectInfo *types.ObjectInfo,
	replicaIdx uint32,
	segmentSize uint32,
	pieceData [][]byte) (integrityHash []byte, signature []byte, err error) {
	pieceDataReader, err := NewPieceDataReader(pieceData)
	if err != nil {
		log.Errorw("failed to sync piece data due to new piece data reader error", "error", err)
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPut, gatewayClient.address+model.SyncerPath, pieceDataReader)
	if err != nil {
		log.Errorw("failed to sync piece data due to new request error", "error", err)
		return nil, nil, err
	}
	marshalObjectInfo := hex.EncodeToString(types.ModuleCdc.MustMarshalJSON(objectInfo))
	req.Header.Add(model.GnfdObjectInfoHeader, marshalObjectInfo)
	req.Header.Add(model.GnfdReplicaIdxHeader, util.Uint32ToString(replicaIdx))
	req.Header.Add(model.GnfdSegmentSizeHeader, util.Uint32ToString(segmentSize))
	req.Header.Add(model.ContentTypeHeader, model.OctetStream)

	resp, err := gatewayClient.httpClient.Do(req)
	if err != nil {
		log.Errorw("failed to sync piece data to other sp", "sp_endpoint", gatewayClient.address, "error", err)
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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
