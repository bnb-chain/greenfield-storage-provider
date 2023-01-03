package stonenode

import (
	"context"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

func (s *StoneNode) AllocStoneJob(ctx context.Context) (
	*service.StoneHubServiceAllocStoneJobResponse, error) {
	resp, err := s.stoneHubClient.AllocStoneJob(ctx, &service.StoneHubServiceAllocStoneJobRequest{TraceId: "test"})
	if err != nil {
		log.Errorw("stone node invokes AllocStoneJob failed", "error", err, "traceID", resp.GetTraceId())
		return nil, err
	}
	if resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("AllocStoneJob response code is not success", "error", err, "traceID", resp.GetTraceId())
	}
	return resp, nil
}

func (s *StoneNode) DoneSecondaryPieceJob(ctx context.Context) (
	*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	resp, err := s.stoneHubClient.DoneSecondaryPieceJob(ctx, &service.StoneHubServiceDoneSecondaryPieceJobRequest{
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
