package stonenode

import (
	"context"
	"time"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// stone node 的数据都是从Primary SP获取
// stone node 无状态，与stone hub、syncer服务交互，居中调度
// RPC接口如何做e2e测试，或者不用e2e测试？
func invokeSyncerService(client service.SyncerServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := client.UploadECPiece(ctx)
	if err != nil {
		log.Errorw("client invoke UploadECPiece error", "err", err)
		return err
	}
	for i := 0; i <= 10; i++ {
		if err := stream.Send(&service.SyncerServiceUploadECPieceRequest{
			TraceId:      "",
			PrimaryJobId: 0,
			PieceJobs:    nil,
			PieceData:    nil,
		}); err != nil {
			log.Errorw("client send request error", "err", err)
			// continue if error, use continue to send next message  or return error
			return err
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorw("client CloseAndRecv error", "err", err)
	}
	log.Infof("traceID: %s", resp.GetTraceId())
	return nil
}
