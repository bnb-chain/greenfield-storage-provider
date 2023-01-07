package stonenode

import (
	"context"
	"errors"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// UploadECPiece send rpc request to secondary storage provider to sync the piece data.
func (node *StoneNodeService) UploadECPiece(ctx context.Context, segmentCount int, sInfo *service.SyncerInfo,
	pieceData map[string][]byte, traceID string) (*service.SyncerServiceUploadECPieceResponse, error) {
	stream, err := node.syncer.UploadECPiece(ctx)
	if err != nil {
		log.Errorw("stone node upload secondary job piece job error", "err", err)
		return nil, err
	}
	for i := 0; i < segmentCount; i++ {
		if err := stream.Send(&service.SyncerServiceUploadECPieceRequest{
			TraceId:    traceID,
			SyncerInfo: sInfo,
			PieceData:  pieceData,
		}); err != nil {
			log.Errorw("client send request error", "error", err)
			return nil, err
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorw("client close error", "error", err, "traceID", resp.GetTraceId())
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("alloc stone from stone hub response code is not success", "error", err, "traceID", resp.GetTraceId())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	log.Infof("traceID: %s", resp.GetTraceId())
	return resp, nil
}
