package stonenode

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

func newStoneHubClient(address string) (service.StoneHubServiceClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), grpcTimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke stoneHub service grpc.DialContext failed", "error", err)
		return nil, err
	}
	defer conn.Close()
	return service.NewStoneHubServiceClient(conn), nil
}

func (s *StoneNodeService) AllocStoneJob(ctx context.Context) (
	*service.StoneHubServiceAllocStoneJobResponse, error) {
	resp, err := s.stoneHub.AllocStoneJob(ctx, &service.StoneHubServiceAllocStoneJobRequest{TraceId: "test"})
	if err != nil {
		log.Errorw("stone node invokes AllocStoneJob failed", "error", err, "traceID", resp.GetTraceId())
		return nil, err
	}
	if resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("AllocStoneJob response code is not success", "error", err, "traceID", resp.GetTraceId())
	}
	return resp, nil
}

func (s *StoneNodeService) DoneSecondaryPieceJob(ctx context.Context) (
	*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	resp, err := s.stoneHub.DoneSecondaryPieceJob(ctx, &service.StoneHubServiceDoneSecondaryPieceJobRequest{
		TraceId:    "test",
		TxHash:     nil,
		PieceJob:   nil,
		ErrMessage: nil,
	})
	if err != nil {
		log.Errorw("stone node invokes DoneSecondaryPieceJob failed", "error", err, "traceID", resp.GetTraceId())
		return nil, err
	}
	if resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("DoneSecondaryPieceJob response code is not success", "error", err, "traceID", resp.GetTraceId())
	}
	return resp, nil
}
