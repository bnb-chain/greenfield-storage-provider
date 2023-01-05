package syncer

import (
	"io"

	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// UploadECPieceAPI used to mock
type UploadECPieceAPI interface {
	UploadECPiece(stream service.SyncerService_UploadECPieceServer) error
}

// UploadECPiece uploads piece data encoded using the ec algorithm to secondary storage provider
func (s *Syncer) UploadECPiece(stream service.SyncerService_UploadECPieceServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Errorw("UploadECPiece receive data error", "error", err, "traceID", req.GetTraceId())
			return err
		}
		if err == io.EOF {
			log.Info("UploadECPiece client closed")
			spInfo, err := handleRequest(req, s.store)
			if err != nil {
				log.Errorw("UploadECPiece handleRequest failed", "error", err)
				return err
			}
			if err := stream.SendAndClose(&service.SyncerServiceUploadECPieceResponse{
				TraceId:         req.GetTraceId(),
				SecondarySpInfo: spInfo,
				ErrMessage: &service.ErrMessage{
					ErrCode: service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED,
					ErrMsg:  "Successful",
				},
			}); err != nil {
				log.Errorw("UploadECPiece SendAndClose error", "traceID", req.GetTraceId(), "err", err)
				return err
			}
		}
	}
}

func handleRequest(req *service.SyncerServiceUploadECPieceRequest, store *storeClient) (
	*service.StorageProviderSealInfo, error) {
	var (
		pieceIndex int
		err        error
	)
	pieceChecksumList := make([][]byte, 0)
	for key, value := range req.GetPieceData() {
		_, _, pieceIndex, err = piecestore.DecodeECPieceKey(key)
		if err != nil {
			log.Errorw("UploadECPiece DecodeECPieceKey failed", "error", err)
			return nil, err
		}
		checksum := generateChecksum(value)
		pieceChecksumList = append(pieceChecksumList, checksum)
		if err := store.putPiece(key, value); err != nil {
			log.Errorw("UploadECPiece put piece failed", "error", err)
			return nil, err
		}
	}

	spID := req.GetSyncerInfo().GetStorageProviderId()
	integrityHash := generateIntegrityHash(pieceChecksumList, spID)
	resp := &service.StorageProviderSealInfo{
		StorageProviderId: spID,
		PieceIdx:          uint32(pieceIndex),
		PieceChecksum:     pieceChecksumList,
		IntegrityHash:     integrityHash,
		Signature:         nil, // TODO(mock)
	}
	return resp, nil
}
