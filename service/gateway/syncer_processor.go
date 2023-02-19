package gateway

import (
	"context"
	"fmt"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// syncPiece send rpc request to secondary storage provider to sync the piece data
func (g *Gateway) syncPiece(ctx context.Context, syncerInfo *stypes.SyncerInfo, pieceData [][]byte, traceID string) (
	*stypes.SyncerServiceSyncPieceResponse, error) {
	stream, err := g.syncer.SyncPiece(ctx)
	if err != nil {
		log.Errorw("sync secondary piece job error", "err", err)
		return nil, err
	}

	// send data one by one to avoid exceeding rpc max msg size
	for _, value := range pieceData {
		if err := stream.Send(&stypes.SyncerServiceSyncPieceRequest{
			TraceId:    traceID,
			SyncerInfo: syncerInfo,
			PieceData:  value,
		}); err != nil {
			log.Errorw("client send request error", "error", err)
			return nil, err
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorw("client close error", "error", err, "traceID", resp.GetTraceId())
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("sync piece sends to stone node response code is not success", "error", err, "traceID", resp.GetTraceId())
		return nil, fmt.Errorf(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}
