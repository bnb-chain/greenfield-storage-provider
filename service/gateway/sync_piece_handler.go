package gateway

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	syncertypes "github.com/bnb-chain/greenfield-storage-provider/service/syncer/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (g *Gateway) syncPieceHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		objectInfo     = types.ObjectInfo{}
		replicateIdx   uint32
		segmentSize    uint64
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode == http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", syncPieceRouterName, errDescription.statusCode,
				reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", syncPieceRouterName,
				reqContext.generateRequestDetail())
		}
	}()

	if g.syncer == nil {
		log.Errorw("failed to sync data due to not config syncer")
		errDescription = NotExistComponentError
		return
	}

	objectInfoMsg, err := hex.DecodeString(r.Header.Get(model.GnfdObjectInfoHeader))
	if err != nil {
		log.Errorw("failed to parse object info header", "object_info", r.Header.Get(model.GnfdObjectInfoHeader))
		errDescription = InvalidHeader
		return
	}
	if types.ModuleCdc.UnmarshalJSON(objectInfoMsg, &objectInfo) != nil {
		log.Errorw("failed to unmarshal object info header", "object_info", r.Header.Get(model.GnfdObjectInfoHeader))
		errDescription = InvalidHeader
		return
	}
	if replicateIdx, err = util.StringToUint32(r.Header.Get(model.GnfdReplicateIdxHeader)); err != nil {
		log.Errorw("failed to parse replicate_idx header", "replicate_idx", r.Header.Get(model.GnfdReplicateIdxHeader))
		errDescription = InvalidHeader
		return
	}
	if segmentSize, err = util.StringToUint64(r.Header.Get(model.GnfdSegmentSizeHeader)); err != nil {
		log.Errorw("failed to parse segment_size header", "segment_size", r.Header.Get(model.GnfdSegmentSizeHeader))
		errDescription = InvalidHeader
		return
	}

	pieceData, err := parseBody(r.Body)
	if err != nil {
		// TODO: add more error
		log.Errorw("failed to parse request body", "error", err)
		errDescription = InternalError
		return
	}

	stream, err := g.syncer.SyncObject(context.Background())
	if err != nil {
		log.Errorw("failed to sync piece", "err", err)
		errDescription = InternalError
		return
	}

	// send data one by one to avoid exceeding rpc max msg size
	for _, value := range pieceData {
		if err = stream.Send(&syncertypes.SyncObjectRequest{
			ObjectInfo:    &objectInfo,
			ReplicateIdx:  replicateIdx,
			SegmentSize:   segmentSize,
			ReplicateData: value,
		}); err != nil {
			log.Errorw("failed to send stream", "error", err)
			errDescription = InternalError
			return
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorw("failed to close stream", "error", err)
		errDescription = InternalError
		return
	}
	// TODO: check resp error code
	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(resp.GetIntegrityHash()))
	w.Header().Set(model.GnfdIntegrityHashSignatureHeader, hex.EncodeToString(resp.GetSignature()))

}

func parseBody(body io.ReadCloser) ([][]byte, error) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, body)
	if err != nil {
		log.Errorw("failed to copy request body", "error", err)
		return nil, merrors.ErrInternalError
	}
	pieceData := make([][]byte, 0)
	if err := json.Unmarshal(buf.Bytes(), &pieceData); err != nil {
		log.Errorw("failed to unmarshal body", "error", err)
		return nil, merrors.ErrInternalError
	}
	return pieceData, nil
}
