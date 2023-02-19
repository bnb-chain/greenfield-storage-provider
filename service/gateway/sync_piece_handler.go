package gateway

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

func (g *Gateway) syncPieceHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
	)

	defer func() {
		statusCode := 200
		if errDescription != nil {
			statusCode = errDescription.statusCode
			_ = errDescription.errorResponse(w, reqContext)
		}
		if statusCode == 200 {
			log.Debugf("action(%v) statusCode(%v) %v", "syncPiece", statusCode,
				reqContext.generateRequestDetail())
		} else {
			log.Errorf("action(%v) statusCode(%v) %v", "syncPiece", statusCode,
				reqContext.generateRequestDetail())
		}
	}()

	reqContext = newRequestContext(r)
	syncerInfo, err := getReqHeader(r.Header)
	if err != nil {
		log.Errorw("get request header failed", "error", err)
		return
	}
	// get trace id
	traceID := r.Header.Get(model.GnfdTraceIDHeader)
	if traceID == "" {
		log.Errorw("traceID header is empty")
		return
	}
	pieceData, err := parseBody(r.Body)
	if err != nil {
		// TODO, add more error
		log.Errorw("parse request body failed", "error", err)
		return
	}
	resp, err := g.syncPiece(context.Background(), syncerInfo, pieceData, traceID)
	if err != nil {
		errDescription = InternalError
	}
	addRespHeader(resp, w)
}

func parseBody(body io.ReadCloser) ([][]byte, error) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, body)
	if err != nil {
		log.Errorw("copy request body failed", "error", err)
		return nil, merrors.ErrInternalError
	}
	pieceData := make([][]byte, 0)
	if err := json.Unmarshal(buf.Bytes(), &pieceData); err != nil {
		log.Errorw("unmarshal body failed", "error", err)
		return nil, merrors.ErrInternalError
	}
	return pieceData, nil
}

func getReqHeader(header http.Header) (*stypes.SyncerInfo, error) {
	syncerInfo := &stypes.SyncerInfo{}
	// get object id
	objectID := header.Get(model.GnfdObjectIDHeader)
	if objectID == "" {
		log.Error("req header object id is empty")
		return nil, merrors.ErrEmptyReqHeader
	}
	id, err := strconv.ParseUint(objectID, 10, 64)
	if err != nil {
		log.Errorw("parse object id failed", "error", err)
		return nil, merrors.ErrReqHeader
	}
	syncerInfo.ObjectId = id

	// get storage provider id
	spID := header.Get(model.GnfdSPIDHeader)
	if spID == "" {
		log.Error("req header sp id is empty")
		return nil, merrors.ErrEmptyReqHeader
	}
	syncerInfo.StorageProviderId = spID

	// get piece count
	pieceCount := header.Get(model.GnfdPieceCountHeader)
	if pieceCount == "" {
		log.Error("req header piece count is empty")
		return nil, merrors.ErrEmptyReqHeader
	}
	pCount, err := strconv.ParseUint(pieceCount, 10, 32)
	if err != nil {
		log.Errorw("parse piece count failed", "error", err)
		return nil, merrors.ErrReqHeader
	}
	syncerInfo.PieceCount = uint32(pCount)

	// get piece index
	pieceIndex := header.Get(model.GnfdPieceIndexHeader)
	if pieceIndex == "" {
		log.Error("req header piece index is empty")
		return nil, merrors.ErrEmptyReqHeader
	}
	pIdx, err := strconv.ParseUint(pieceIndex, 10, 32)
	if err != nil {
		log.Errorw("parse piece index failed", "error", err)
		return nil, merrors.ErrReqHeader
	}
	syncerInfo.PieceIndex = uint32(pIdx)

	// get redundancy type
	redundancyType := header.Get(model.GnfdRedundancyTypeHeader)
	if redundancyType == "" {
		log.Error("req header redundancy type is empty")
		return nil, merrors.ErrEmptyReqHeader
	}
	rType, err := transferRedundancyType(redundancyType)
	if err != nil {
		log.Errorw("transfer redundancy type failed", "error", err)
		return nil, err
	}
	syncerInfo.RedundancyType = rType
	return syncerInfo, nil
}

func addRespHeader(resp *stypes.SyncerServiceSyncPieceResponse, w http.ResponseWriter) {
	w.Header().Set(model.GnfdTraceIDHeader, resp.GetTraceId())
	w.Header().Set(model.GnfdSPIDHeader, resp.GetSecondarySpInfo().GetStorageProviderId())
	w.Header().Set(model.GnfdPieceIndexHeader, strconv.Itoa(int(resp.GetSecondarySpInfo().GetPieceIdx())))

	checksum := handlePieceChecksum(resp.GetSecondarySpInfo().GetPieceChecksum())
	log.Infow("gateway piece checksum", "cc", checksum)
	w.Header().Set(model.GnfdPieceChecksumHeader, checksum)

	integrityHash := hex.EncodeToString(resp.GetSecondarySpInfo().GetIntegrityHash())
	w.Header().Set(model.GnfdIntegrityHashHeader, integrityHash)

	sig := hex.EncodeToString([]byte("test_signature"))
	w.Header().Set(model.GnfdSealSignatureHeader, sig)
}

func handlePieceChecksum(pieceChecksum [][]byte) string {
	list := make([]string, len(pieceChecksum))
	for index, val := range pieceChecksum {
		list[index] = hex.EncodeToString(val)
	}
	return strings.Join(list, ",")
}

func transferRedundancyType(redundancyType string) (ptypes.RedundancyType, error) {
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED.String():
		return ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED, nil
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE.String():
		return ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, nil
	case ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE.String():
		return ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE, nil
	default:
		return -1, merrors.ErrRedundancyType
	}
}
