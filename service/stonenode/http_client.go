package stonenode

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	storagetypespb "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// sendRequest send piece data to gateway through HTTP protocol
func sendRequest(pieceData [][]byte, httpEndpoint string, syncerInfo *stypes.SyncerInfo, traceID string) (
	*stypes.StorageProviderSealInfo, error) {
	//TODO, use io.Copy to avoid using big memory
	body, err := json.Marshal(pieceData)
	if err != nil {
		log.Errorw("marshal piece data failed", "error", err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s%s", httpEndpoint, model.SyncerPath), bytes.NewReader(body))
	if err != nil {
		log.Errorw("http NewRequest failed", "error", err)
		return nil, err
	}

	req = addReqHeader(req, syncerInfo, traceID)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorw("client Do failed", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	// if http.StatusCode isn't 200, return error
	if resp.StatusCode != http.StatusOK {
		log.Error("HTTP status code is not 200: ", resp.StatusCode)
		if err := parseBody(resp.Body); err != nil {
			log.Errorw("parse body error")
			return nil, err
		}
	}
	sealInfo, err := generateSealInfo(resp)
	if err != nil {
		log.Errorw("generate seal info", "error", err)
		return nil, err
	}
	return sealInfo, nil
}

func addReqHeader(req *http.Request, syncerInfo *stypes.SyncerInfo, traceID string) *http.Request {
	req.Header.Add(model.ContentTypeHeader, model.OctetStream)
	req.Header.Add(model.GnfdTraceIDHeader, traceID)
	req.Header.Add(model.GnfdObjectIDHeader, strconv.FormatUint(syncerInfo.GetObjectId(), 10))
	req.Header.Add(model.GnfdSPIDHeader, syncerInfo.GetStorageProviderId())
	req.Header.Add(model.GnfdPieceCountHeader, strconv.FormatUint(uint64(syncerInfo.GetPieceCount()), 10))
	req.Header.Add(model.GnfdPieceIndexHeader, strconv.FormatUint(uint64(syncerInfo.GetPieceIndex()), 10))
	req.Header.Add(model.GnfdRedundancyTypeHeader, syncerInfo.GetRedundancyType().String())
	return req
}

// TODO, perfect handling error message later
func parseBody(body io.ReadCloser) error {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, body)
	if err != nil {
		log.Errorw("copy request body failed", "error", err)
		return err
	}
	// parse error message in response body
	//if err := xml.Unmarshal(buf.Bytes(), nil); err != nil {
	//	log.Errorw("unmarshal xml response body failed", "error", err)
	//	return err
	//}
	return fmt.Errorf("HTTP status code is not 200")
}

func generateSealInfo(resp *http.Response) (*stypes.StorageProviderSealInfo, error) {
	spSealInfo := &stypes.StorageProviderSealInfo{}
	// get storage provider ID
	spID := resp.Header.Get(model.GnfdSPIDHeader)
	if spID == "" {
		log.Error("resp header sp id is empty")
		return nil, merrors.ErrEmptyRespHeader
	}
	spSealInfo.StorageProviderId = spID

	// get piece index
	pieceIndex := resp.Header.Get(model.GnfdPieceIndexHeader)
	if pieceIndex == "" {
		log.Error("resp header piece index is empty")
		return nil, merrors.ErrEmptyRespHeader
	}
	idx, err := strconv.ParseUint(pieceIndex, 10, 32)
	if err != nil {
		log.Errorw("parse piece index failed", "error", err)
		return nil, merrors.ErrRespHeader
	}
	spSealInfo.PieceIdx = uint32(idx)

	// get piece checksum
	pieceChecksum := resp.Header.Get(model.GnfdPieceChecksumHeader)
	if pieceChecksum == "" {
		log.Error("resp header piece checksum is empty")
		return nil, merrors.ErrEmptyRespHeader
	}
	checksum, err := handlePieceChecksumHeader(pieceChecksum)
	if err != nil {
		return nil, err
	}
	spSealInfo.PieceChecksum = checksum

	// get integrity hash
	integrityHash := resp.Header.Get(model.GnfdIntegrityHashHeader)
	if integrityHash == "" {
		log.Error("resp header integrity hash is empty")
		return nil, merrors.ErrEmptyRespHeader
	}
	iHash, err := hex.DecodeString(integrityHash)
	if err != nil {
		log.Errorw("decode integrity hash failed", "error", err)
		return nil, merrors.ErrRespHeader
	}
	spSealInfo.IntegrityHash = iHash

	// get signature
	signature := resp.Header.Get(model.GnfdSealSignatureHeader)
	if signature == "" {
		log.Error("resp header seal signature is empty")
		return nil, merrors.ErrEmptyRespHeader
	}
	sig, err := hex.DecodeString(signature)
	if err != nil {
		log.Errorw("decode seal signature failed", "error", err)
		return nil, merrors.ErrRespHeader
	}
	spSealInfo.Signature = sig
	return spSealInfo, nil
}

func handlePieceChecksumHeader(pieceChecksum string) ([][]byte, error) {
	list := strings.Split(pieceChecksum, ",")
	checksum := make([][]byte, len(list))
	for i, val := range list {
		data, err := hex.DecodeString(val)
		if err != nil {
			log.Errorw("decode piece checksum failed", "error", err)
			return nil, merrors.ErrRespHeader
		}
		checksum[i] = data
	}
	return checksum, nil
}

func getApprovalSigNature() {
	storagetypespb.NewMsgCreateObject(nil, "", "", 0, false, nil, "", nil, nil)
}
