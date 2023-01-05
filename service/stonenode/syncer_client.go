package stonenode

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

func newSyncerClient(address string) (service.SyncerServiceClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), grpcTimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke syncer service grpc.DialContext failed", "error", err)
		return nil, err
	}
	defer conn.Close()
	return service.NewSyncerServiceClient(conn), nil
}

// stone node 的数据都是从Primary SP获取
// stone node 无状态，与stone hub、syncer服务交互，居中调度
// RPC接口如何做e2e测试，或者不用e2e测试？
func (s *StoneNodeService) UploadECPiece(ctx context.Context, segmentNumber uint64, sInfo *service.SyncerInfo,
	pieceData map[string][]byte, traceID string) (*service.SyncerServiceUploadECPieceResponse, error) {
	stream, err := s.syncer.UploadECPiece(ctx)
	if err != nil {
		log.Errorw("stone node invokes UploadECPiece error", "err", err)
		return nil, err
	}

	for i := 0; i <= int(segmentNumber); i++ {
		if err := stream.Send(&service.SyncerServiceUploadECPieceRequest{
			TraceId: traceID,
			SyncerInfo: &service.SyncerInfo{
				ObjectId:          sInfo.GetObjectId(),
				StorageProviderId: sInfo.GetStorageProviderId(),
				RedundancyType:    sInfo.GetRedundancyType(),
			},
			PieceData: pieceData,
		}); err != nil {
			log.Errorw("client send request error", "error", err)
			return nil, err
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorw("client CloseAndRecv error", "error", err, "traceID", resp.GetTraceId())
	}
	if resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("AllocStoneJob response code is not success", "error", err, "traceID", resp.GetTraceId())
	}
	log.Infof("traceID: %s", resp.GetTraceId())
	return resp, nil
}
