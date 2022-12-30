package syncer

import (
	"io"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// UploadECPieceAPI used to mock
type UploadECPieceAPI interface {
	UploadECPiece(stream service.SyncerService_UploadECPieceServer) error
}

type Syncer struct{}

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
			if err := stream.SendAndClose(&service.SyncerServiceUploadECPieceResponse{
				TraceId:         req.GetTraceId(),
				SecondarySpInfo: nil,
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
func handleRequest(req *service.SyncerServiceUploadECPieceRequest) error {
	if len(req.GetPieceData()) != len(req.GetPieceJobs()) {
		log.Errorw("the number of PieceData is not equal to PieceJobs", "PieceData length", len(req.GetPieceData()),
			"PieceJobs length", len(req.GetPieceJobs()))
	}
	for _, i := range req.GetPieceJobs() {
		for _, j := range i.GetStorageProviderSealInfo() {
			spID := j.GetStorageProviderId()
			log.Infof("spID: %s", spID)
		}
	}
	return nil
}
