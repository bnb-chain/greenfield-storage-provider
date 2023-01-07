package stonenode

import (
	"context"
	"fmt"

	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// AllocStoneJob sends rpc request to stone hub to alloc stone job.
func (node *StoneNodeService) AllocStoneJob(ctx context.Context) (*service.StoneHubServiceAllocStoneJobResponse, error) {
	resp, err := node.stoneHub.AllocStoneJob(ctx)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "stone node invokes AllocStoneJob failed", "error", err)
		return nil, err
	}
	if resp.PieceJob == nil {
		log.CtxErrorw(ctx, "stone node invokes AllocStoneJob empty.")
		return nil, merrors.ErrEmptyJob
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "stone node invokes AllocStoneJob failed", "error", resp.GetErrMessage().GetErrMsg())
		return nil, fmt.Errorf(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

// DoneSecondaryPieceJob sends rpc request to stone hub to report upload secondary piece job.
func (node *StoneNodeService) DoneSecondaryPieceJob(ctx context.Context, traceID string, pieceJob *service.PieceJob, errMsg *service.ErrMessage) error {
	resp, err := node.stoneHub.DoneSecondaryPieceJob(ctx, &service.StoneHubServiceDoneSecondaryPieceJobRequest{
		TraceId:    traceID,
		TxHash:     pieceJob.GetTxHash(),
		PieceJob:   pieceJob,
		ErrMessage: errMsg,
	})
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "stone node invokes DoneSecondaryPieceJob failed", "error", err)
		return err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "done secondary piece job response code is not success", "error", resp.GetErrMessage().GetErrMsg())
		return fmt.Errorf(resp.GetErrMessage().GetErrMsg())
	}
	return nil
}
