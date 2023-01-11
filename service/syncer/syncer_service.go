package syncer

import (
	"context"
	"io"

	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// UploadECPiece uploads piece data encoded using the ec algorithm to secondary storage provider
func (s *Syncer) UploadECPiece(stream service.SyncerService_UploadECPieceServer) (err error) {
	var (
		req      *service.SyncerServiceUploadECPieceRequest
		sealInfo *service.StorageProviderSealInfo
		ctx      = context.Background()
	)
	defer func(sealInfo *service.StorageProviderSealInfo, err error) {
		//if err != nil {
		//	err = stream.SendAndClose(&service.SyncerServiceUploadECPieceResponse{
		//		TraceId:         req.GetTraceId(),
		//		SecondarySpInfo: sealInfo,
		//	})
		//}
		resp := &service.SyncerServiceUploadECPieceResponse{
			TraceId:         req.GetTraceId(),
			SecondarySpInfo: sealInfo,
		}
		if err != nil {
			resp.ErrMessage.ErrCode = service.ErrCode_ERR_CODE_ERROR
			resp.ErrMessage.ErrMsg = err.Error()
		}
		err = stream.SendAndClose(resp)
		log.CtxInfow(ctx, "upload ec piece closed", "error", err, "storage_provider_id", sealInfo.GetStorageProviderId(),
			"piece_idx", sealInfo.GetPieceIdx(), "piece_checksum", sealInfo.GetPieceChecksum(), "integrity_hash", sealInfo.GetIntegrityHash())
	}(sealInfo, err)

	for {
		req, err = stream.Recv()
		log.Infow("first", "object_id", req.GetSyncerInfo().GetObjectId(), "tx_hash", req.GetSyncerInfo().GetTxHash(),
			"storage_provider_id", req.GetSyncerInfo().GetStorageProviderId(), "rType", req.GetSyncerInfo().GetRedundancyType(), "traceID", req.GetTraceId())
		log.Context(ctx, req)
		if err != nil && err != io.EOF {
			log.CtxErrorw(ctx, "upload piece receive data error", "error", err)
			break
		}
		if err == io.EOF {
			err = nil
			log.Info("50")
			if req.GetSyncerInfo() != nil {
				log.Info("51")
				sealInfo, err = s.handleUploadPiece(ctx, req)
				if err != nil {
					log.CtxErrorw(ctx, "handle upload piece error", "error", err)
					return
				}
			}
			//return
		}
		sealInfo, err = s.handleUploadPiece(ctx, req)
		if err != nil {
			log.CtxErrorw(ctx, "handle upload piece error", "error", err)
			break
		}
		return
	}
	return
}

// handleUploadPiece store piece data to piece store and compute integrity hash.
func (s *Syncer) handleUploadPiece(ctx context.Context, req *service.SyncerServiceUploadECPieceRequest) (
	*service.StorageProviderSealInfo, error) {
	var (
		pieceIndex uint32
		err        error
	)
	log.Infow("second", "req", req.GetSyncerInfo(), "rType", req.GetSyncerInfo().GetRedundancyType(),
		"traceID", req.GetTraceId())
	pieceChecksumList := make([][]byte, 0)
	for key, value := range req.GetPieceData() {
		// if redundancyType is ec, if check all pieceIndex is equal
		pieceIndex, err = parsePieceIndex(req.GetSyncerInfo().GetRedundancyType(), key)
		if err != nil {
			return nil, err
		}
		checksum := hash.GenerateChecksum(value)
		pieceChecksumList = append(pieceChecksumList, checksum)
		if err = s.store.PutPiece(key, value); err != nil {
			log.CtxErrorw(ctx, "put piece failed", "error", err)
			return nil, err
		}
	}

	spID := req.GetSyncerInfo().GetStorageProviderId()
	integrityHash := hash.GenerateIntegrityHash(pieceChecksumList, spID)
	log.CtxInfow(ctx, "handleUploadPiece", "spID", spID, "pieceIndex", pieceIndex,
		"pieceChecksum", pieceChecksumList, "IntegrityHash", integrityHash)
	resp := &service.StorageProviderSealInfo{
		StorageProviderId: spID,
		PieceIdx:          pieceIndex,
		PieceChecksum:     pieceChecksumList,
		IntegrityHash:     integrityHash,
		Signature:         nil, // TODO(mock)
	}
	return resp, nil
}

func parsePieceIndex(redundancyType ptypes.RedundancyType, key string) (uint32, error) {
	var (
		err        error
		pieceIndex uint32
	)
	log.Infow("parsePieceIndex", "rType", redundancyType, "key", key)
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		_, _, pieceIndex, err = piecestore.DecodeECPieceKey(key)
		if err != nil {
			log.Errorw("decode ec piece key failed", "error", err)
			return 0, err
		}
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		_, pieceIndex, err = piecestore.DecodeSegmentPieceKey(key)
		if err != nil {
			log.Errorw("decode segment piece key failed", "error", err)
			return 0, err
		}
	default:
		return 0, merrors.ErrRedundancyType
	}
	return pieceIndex, nil
}
