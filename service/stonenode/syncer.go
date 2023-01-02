package stonenode

import (
	"context"

	pkg "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	segmentSize = 16 * 1024 * 1024
)

type StoneNode struct {
	syncerClient   service.SyncerServiceClient
	stoneHubClient service.StoneHubServiceClient
}

// stone node 的数据都是从Primary SP获取
// stone node 无状态，与stone hub、syncer服务交互，居中调度
// RPC接口如何做e2e测试，或者不用e2e测试？
func (s *StoneNode) UploadECPiece(ctx context.Context) error {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	stream, err := s.syncerClient.UploadECPiece(ctx)
	if err != nil {
		log.Errorw("client invoke UploadECPiece error", "err", err)
		return err
	}

	allocResp, err := s.AllocStoneJob(ctx)
	if err != nil {
		log.Errorw("UploadECPiece call AllocStoneJob failed", "error", err, "traceID", allocResp.GetTraceId())
		return err
	}
	//segmentKey := allocResp.GetPieceJob().GetBucketName() + allocResp.GetPieceJob().GetObjectName()
	payloadSize := allocResp.GetPieceJob().GetPayloadSize()
	var segmentNumber uint64
	if payloadSize%segmentSize == 0 {
		segmentNumber = payloadSize / segmentSize
	} else {
		segmentNumber = payloadSize/segmentSize + 1
	}
	if isEC := checkEC(allocResp.GetPieceJob().GetRedundancyType()); isEC {

	}

	for i := 0; i <= int(segmentNumber); i++ {
		if err := stream.Send(&service.SyncerServiceUploadECPieceRequest{
			TraceId:           "",
			StorageProviderId: "",
			PieceData:         nil,
		}); err != nil {
			log.Errorw("client send request error", "error", err)
			return err
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
	return nil
}

func doEC() {

}

func checkEC(rType pkg.RedundancyType) bool {
	if rType == pkg.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED {
		return true
	}
	return false
}
