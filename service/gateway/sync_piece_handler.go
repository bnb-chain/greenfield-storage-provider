package gateway

//
//import (
//	"bytes"
//	"context"
//	"encoding/hex"
//	"encoding/json"
//	"io"
//	"net/http"
//	"strconv"
//
//	"github.com/bnb-chain/greenfield-storage-provider/model"
//	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
//	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
//	"github.com/bnb-chain/greenfield-storage-provider/util"
//	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
//)
//
//func (g *Gateway) syncPieceHandler(w http.ResponseWriter, r *http.Request) {
//	var (
//		err            error
//		errDescription *errorDescription
//		reqContext     *requestContext
//	)
//
//	reqContext = newRequestContext(r)
//	defer func() {
//		if errDescription != nil {
//			_ = errDescription.errorResponse(w, reqContext)
//		}
//		if errDescription != nil && errDescription.statusCode == http.StatusOK {
//			log.Errorf("action(%v) statusCode(%v) %v", syncPieceRouterName, errDescription.statusCode,
//				reqContext.generateRequestDetail())
//		} else {
//			log.Infof("action(%v) statusCode(200) %v", syncPieceRouterName,
//				reqContext.generateRequestDetail())
//		}
//	}()
//
//	log.Infow("sync piece handler receive data", "request", reqContext.generateRequestDetail())
//	syncerInfo, err := getReqHeader(r.Header)
//	if err != nil {
//		log.Errorw("get request header failed", "error", err)
//		return
//	}
//	// get trace id
//	traceID := r.Header.Get(model.GnfdRequestIDHeader)
//	if traceID == "" {
//		traceID = reqContext.requestID
//	}
//	pieceData, err := parseBody(r.Body)
//	if err != nil {
//		// TODO, add more error
//		log.Errorw("parse request body failed", "error", err)
//		return
//	}
//	resp, err := g.syncPiece(context.Background(), syncerInfo, pieceData, traceID)
//	if err != nil {
//		errDescription = InternalError
//	}
//	addRespHeader(resp, w)
//	log.Infow("sync piece handler reply response to stone node", "response header", w.Header())
//}
//
//func parseBody(body io.ReadCloser) ([][]byte, error) {
//	buf := &bytes.Buffer{}
//	_, err := io.Copy(buf, body)
//	if err != nil {
//		log.Errorw("copy request body failed", "error", err)
//		return nil, merrors.ErrInternalError
//	}
//	pieceData := make([][]byte, 0)
//	if err := json.Unmarshal(buf.Bytes(), &pieceData); err != nil {
//		log.Errorw("unmarshal body failed", "error", err)
//		return nil, merrors.ErrInternalError
//	}
//	return pieceData, nil
//}
//
//func getReqHeader(header http.Header) (*stypes.SyncerInfo, error) {
//	syncerInfo := &stypes.SyncerInfo{}
//	// get object id
//	objectID := header.Get(model.GnfdObjectIDHeader)
//	if objectID == "" {
//		log.Error("req header object id is empty")
//		return nil, merrors.ErrEmptyReqHeader
//	}
//	id, err := util.StringToUin64(objectID)
//	if err != nil {
//		log.Errorw("parse object id failed", "error", err)
//		return nil, merrors.ErrReqHeader
//	}
//	syncerInfo.ObjectId = id
//
//	// get storage provider id
//	spID := header.Get(model.GnfdSPIDHeader)
//	if spID == "" {
//		log.Error("req header sp id is empty")
//		return nil, merrors.ErrEmptyReqHeader
//	}
//	syncerInfo.StorageProviderId = spID
//
//	// get piece count
//	pieceCount := header.Get(model.GnfdPieceCountHeader)
//	if pieceCount == "" {
//		log.Error("req header piece count is empty")
//		return nil, merrors.ErrEmptyReqHeader
//	}
//	pCount, err := util.StringToUint32(pieceCount)
//	if err != nil {
//		log.Errorw("parse piece count failed", "error", err)
//		return nil, merrors.ErrReqHeader
//	}
//	syncerInfo.PieceCount = pCount
//
//	// get piece index
//	pieceIndex := header.Get(model.GnfdPieceIndexHeader)
//	if pieceIndex == "" {
//		log.Error("req header piece index is empty")
//		return nil, merrors.ErrEmptyReqHeader
//	}
//	pIdx, err := util.StringToUint32(pieceIndex)
//	if err != nil {
//		log.Errorw("parse piece index failed", "error", err)
//		return nil, merrors.ErrReqHeader
//	}
//	syncerInfo.PieceIndex = pIdx
//
//	// get redundancy type
//	redundancyType := header.Get(model.GnfdRedundancyTypeHeader)
//	if redundancyType == "" {
//		log.Error("req header redundancy type is empty")
//		return nil, merrors.ErrEmptyReqHeader
//	}
//	rType, err := util.TransferRedundancyType(redundancyType)
//	if err != nil {
//		log.Errorw("transfer redundancy type failed", "error", err)
//		return nil, err
//	}
//	syncerInfo.RedundancyType = rType
//	return syncerInfo, nil
//}
//
//func addRespHeader(resp *stypes.SyncerServiceSyncPieceResponse, w http.ResponseWriter) {
//	w.Header().Set(model.GnfdRequestIDHeader, resp.GetTraceId())
//	w.Header().Set(model.GnfdSPIDHeader, resp.GetSecondarySpInfo().GetStorageProviderId())
//	w.Header().Set(model.GnfdPieceIndexHeader, strconv.Itoa(int(resp.GetSecondarySpInfo().GetPieceIdx())))
//
//	checksum := util.BytesSliceToString(resp.GetSecondarySpInfo().GetPieceChecksum())
//	w.Header().Set(model.GnfdPieceChecksumHeader, checksum)
//
//	integrityHash := hex.EncodeToString(resp.GetSecondarySpInfo().GetIntegrityHash())
//	w.Header().Set(model.GnfdIntegrityHashHeader, integrityHash)
//
//	integrityHashSignature := hex.EncodeToString(resp.GetSecondarySpInfo().GetSignature())
//	w.Header().Set(model.GnfdIntegrityHashSignatureHeader, integrityHashSignature)
//}
