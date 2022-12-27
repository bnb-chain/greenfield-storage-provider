package syncer

import (
	"context"
	"io"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// UploadECPieceAPI mock
type UploadECPieceAPI interface {
	UploadECPiece(ctx context.Context, stream service.SyncerService_UploadECPieceServer) error
}

// syncerServer synchronizes ec data to piece store
type syncerServer struct {
}

// UploadECPiece uploads piece data encoded using the ec algorithm to secondary storage provider
func (s *syncerServer) UploadECPiece(ctx context.Context, stream service.SyncerService_UploadECPieceServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			if err := stream.SendAndClose(&service.SyncerServiceUploadECPieceResponse{
				//TraceId:         "TraceIDASDFGHJKL1234567890",
				SecondarySpInfo: nil,
				ErrMessage:      nil,
			}); err != nil {
				log.Errorw("UploadECPiece SendAndClose error", "err", err)
			}
			return nil
		}
		if err != nil {
			log.Errorw("UploadECPiece receive data error", "err", err)
		}
	}
}
