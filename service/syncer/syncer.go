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

type SyncerImpl struct {
	syncer *SyncerService
}

// UploadECPiece uploads piece data encoded using the ec algorithm to secondary storage provider
func (s *SyncerImpl) UploadECPiece(stream service.SyncerService_UploadECPieceServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Errorw("UploadECPiece receive data error", "error", err, "traceID", req.GetTraceId())
			return err
		}
		if err == io.EOF {
			log.Info("UploadECPiece client closed")
			spInfo, err := handleRequest(req, s.syncer.store)
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

// syncer服务生成signature与checksum，stone node校验signature与完整性hash
// Piece写入哪个sp，按照数组里的顺序写入即可？
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
			log.Errorw("UploadECPiece put piece error", "error", err)
			return nil, err
		}
	}
	resp := &service.StorageProviderSealInfo{
		StorageProviderId: req.GetStorageProviderId(),
		PieceIdx:          uint32(pieceIndex),
		PieceChecksum:     pieceChecksumList,
		IntegrityHash:     nil,
		Signature:         nil,
	}
	return resp, nil
}
