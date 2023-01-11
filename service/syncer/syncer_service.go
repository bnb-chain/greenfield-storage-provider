package syncer

import (
	"context"
	"io"

	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
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
			sealInfo, err = s.handleUploadPiece(ctx, req)
			if err != nil {
				log.CtxErrorw(ctx, "handle upload piece error", "error", err)
				break
			}
			err = stream.SendAndClose(&service.SyncerServiceUploadECPieceResponse{
				TraceId:         req.GetTraceId(),
				SecondarySpInfo: sealInfo,
			})
			log.CtxInfow(ctx, "upload ec piece closed", "error", err)
			return
		}
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
	log.Infow("second", "req", req.GetSyncerInfo(), "traceID", req.GetTraceId())
	pieceChecksumList := make([][]byte, 0)
	for key, value := range req.GetPieceData() {
		_, _, pieceIndex, err = piecestore.DecodeECPieceKey(key)
		if err != nil {
			log.CtxErrorw(ctx, "decode piece key failed", "error", err)
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
