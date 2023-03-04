package gateway

import (
	"context"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	syncertypes "github.com/bnb-chain/greenfield-storage-provider/service/syncer/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

func (g *Gateway) syncPieceHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		objectInfo     = types.ObjectInfo{}
		replicateIdx   uint32
		segmentSize    uint64
		size           int
		readN          int
		buf            = make([]byte, model.StreamBufSize)
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode == http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", syncPieceRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", syncPieceRouterName, reqContext.generateRequestDetail())
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

	stream, err := g.syncer.SyncObject(context.Background())
	if err != nil {
		log.Errorw("failed to sync piece", "error", err)
		errDescription = InternalError
		return
	}
	resp := &syncertypes.SyncObjectResponse{}
	for {
		readN, err = r.Body.Read(buf)
		if err != nil && err != io.EOF {
			log.Errorw("failed to sync piece due to reader error", "error", err)
			errDescription = InternalError
			return
		}
		if readN > 0 {
			if err = stream.Send(&syncertypes.SyncObjectRequest{
				ObjectInfo:    &objectInfo,
				ReplicateIdx:  replicateIdx,
				SegmentSize:   segmentSize,
				ReplicateData: buf[:readN],
			}); err != nil {
				log.Errorw("failed to send stream", "error", err)
				errDescription = InternalError
				return
			}
			size += readN
		}
		if err == io.EOF {
			if size == 0 {
				log.Errorw("failed to sync piece due to payload is empty")
				errDescription = InvalidPayload
				return
			}
			resp, err = stream.CloseAndRecv()
			if err != nil {
				log.Errorw("failed to sync piece due to stream close", "error", err)
				errDescription = InternalError
				return
			}
			// succeed to sync piece
			break
		}
	}

	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(resp.GetIntegrityHash()))
	w.Header().Set(model.GnfdIntegrityHashSignatureHeader, hex.EncodeToString(resp.GetSignature()))
}
