package gateway

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	errorstypes "github.com/bnb-chain/greenfield-storage-provider/pkg/errors/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	receivertypes "github.com/bnb-chain/greenfield-storage-provider/service/receiver/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// syncPieceHandler handle sync piece data request
func (gateway *Gateway) replicatePieceHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                    error
		errDescription         *errorDescription
		reqContext             *requestContext
		objectInfo             *types.ObjectInfo
		pieceSize              uint64
		redundancyIdx          int32
		replicateApproval      = &p2ptypes.GetApprovalResponse{}
		size                   int
		readN                  int
		buf                    = make([]byte, model.DefaultStreamBufSize)
		integrityHash          []byte
		integrityHashSignature []byte
		ctx, cancel            = context.WithCancel(context.Background())
	)

	reqContext = newRequestContext(r)
	defer func() {
		cancel()
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", replicateObjectPieceRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", replicateObjectPieceRouterName, reqContext.generateRequestDetail())
		}
	}()

	if gateway.receiver == nil {
		log.Error("failed to replicate piece due to not config receiver")
		_ = makeXMLHTPPResponse(w, merrors.NotExistedComponentErrCode, reqContext.requestID)
		return
	}

	// check object info by querying chain
	objectID := reqContext.request.Header.Get(model.GnfdObjectIDHeader)
	if objectInfo, err = gateway.chain.QueryObjectInfoByID(context.Background(), objectID); err != nil {
		log.Errorw("failed to query object info  on chain", "object_id", objectID, "error", err)
		_ = makeXMLHTPPResponse(w, errorstypes.Code(err), reqContext.requestID)
		return
	}
	if redundancyIdx, err = util.StringToInt32(r.Header.Get(model.GnfdRedundancyIndexHeader)); err != nil {
		log.Errorw("failed to parse redundancy_idx header", "redundancy_idx", r.Header.Get(model.GnfdRedundancyIndexHeader))
		_ = makeXMLHTPPResponse(w, merrors.ParseStringToNumberErrCode, reqContext.requestID)
		return
	}
	if pieceSize, err = util.StringToUint64(r.Header.Get(model.GnfdPieceSizeHeader)); err != nil {
		log.Errorw("failed to parse piece_size header", "piece_size", r.Header.Get(model.GnfdPieceSizeHeader))
		_ = makeXMLHTPPResponse(w, merrors.ParseStringToNumberErrCode, reqContext.requestID)
		return
	}
	if err = json.Unmarshal([]byte(r.Header.Get(model.GnfdReplicateApproval)), replicateApproval); err != nil {
		log.Errorw("failed to parse replicate_approval header", "replicate_approval", r.Header.Get(model.GnfdReplicateApproval))
		_ = makeXMLHTPPResponse(w, merrors.UnmarshalReplicateApprovalErrCode, reqContext.requestID)
		return
	}
	if err = gateway.verifyReplicateApproval(replicateApproval); err != nil {
		log.Errorw("failed to verify replicate_approval header", "replicate_approval", replicateApproval)
		_ = makeXMLHTPPResponse(w, errorstypes.Code(err), reqContext.requestID)
		return
	}

	stream, err := gateway.receiver.ReceiveObjectPiece(ctx)
	if err != nil {
		log.Errorw("failed to replicate piece", "error", err)
		_ = makeXMLHTPPResponse(w, merrors.InternalErrCode, reqContext.requestID)
		return
	}
	for {
		readN, err = r.Body.Read(buf)
		if err != nil && err != io.EOF {
			log.Errorw("failed to replicate piece due to reader error", "error", err)
			_ = makeXMLHTPPResponse(w, merrors.InternalErrCode, reqContext.requestID)
			return
		}
		if readN > 0 {
			if err = stream.Send(&receivertypes.ReceiveObjectPieceRequest{
				ObjectInfo:      objectInfo,
				PieceSize:       pieceSize,
				RedundancyIdx:   redundancyIdx,
				PieceStreamData: buf[:readN],
			}); err != nil {
				log.Errorw("failed to send stream", "error", err)
				_ = makeXMLHTPPResponse(w, merrors.InternalErrCode, reqContext.requestID)
				return
			}
			size += readN
		}
		if err == io.EOF {
			if size == 0 {
				log.Errorw("failed to replicate piece due to payload is empty")
				_ = makeXMLHTPPResponse(w, merrors.ZeroPayloadErrCode, reqContext.requestID)
				return
			}
			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Errorw("failed to replicate piece due to stream close", "error", err)
				_ = makeXMLHTPPResponse(w, merrors.InternalErrCode, reqContext.requestID)
				return
			}
			integrityHash = resp.GetIntegrityHash()
			integrityHashSignature = resp.GetSignature()
			// succeed to sync piece
			break
		}
	}

	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(integrityHash))
	w.Header().Set(model.GnfdIntegrityHashSignatureHeader, hex.EncodeToString(integrityHashSignature))
}
