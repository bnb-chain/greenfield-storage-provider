package stonenode

import (
	"context"
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/mock"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// allocStoneJob sends rpc request to stone hub alloc stone job
func (node *StoneNodeService) allocStoneJob(ctx context.Context) {
	resp, err := node.stoneHub.AllocStoneJob(ctx)
	ctx = log.Context(ctx, resp, resp.GetPieceJob())
	if err != nil {
		if errors.Is(err, merrors.ErrEmptyJob) {
			return
		}
		log.CtxErrorw(ctx, "alloc stone from stone hub failed", "error", err)
		return
	}
	// TBD:: stone node will support more types of stone job,
	// currently only support upload secondary piece job
	if err := node.loadAndSyncPieces(ctx, resp); err != nil {
		log.CtxErrorw(ctx, "upload secondary piece job failed", "error", err)
		return
	}
	log.CtxInfow(ctx, "upload secondary piece job success")
}

// loadAndSyncPieces load segment data from primary and sync to secondary
func (node *StoneNodeService) loadAndSyncPieces(ctx context.Context, allocResp *stypes.StoneHubServiceAllocStoneJobResponse) error {
	// TBD:: check secondarySPs count by redundancyType.
	// EC_TYPE need EC_M + EC_K + backup
	// REPLICA_TYPE and INLINE_TYPE need segments count + backup
	secondarySPs := mock.AllocUploadSecondarySP()

	// validate redundancyType and targetIdx
	redundancyType := allocResp.GetPieceJob().GetRedundancyType()
	if err := util.ValidateRedundancyType(redundancyType); err != nil {
		log.CtxErrorw(ctx, "invalid redundancy type", "redundancy type", redundancyType)
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}
	targetIdx := allocResp.GetPieceJob().GetTargetIdx()
	if len(targetIdx) == 0 {
		log.CtxError(ctx, "invalid target idx length")
		node.reportErrToStoneHub(ctx, allocResp, merrors.ErrEmptyTargetIdx)
		return merrors.ErrEmptyTargetIdx
	}

	// 1. load all segments data from primary piece store and encode
	pieceData, err := node.encodeSegmentsData(ctx, allocResp)
	if err != nil {
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}

	// 2. dispatch the piece data to different secondary sp
	secondaryPieceData, err := node.dispatchSecondarySP(pieceData, redundancyType, secondarySPs, targetIdx)
	if err != nil {
		log.CtxErrorw(ctx, "dispatch piece data to secondary sp error")
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}

	if len(secondaryPieceData) > len(node.syncer) {
		log.Errorw("syncer number is not enough")
		return merrors.ErrSyncerNumber
	}

	// 3. send piece data to the secondary
	node.doSyncToSecondarySP(ctx, allocResp, secondaryPieceData, secondarySPs)
	return nil
}

// reportErrToStoneHub send error message to stone hub
func (node *StoneNodeService) reportErrToStoneHub(ctx context.Context, resp *stypes.StoneHubServiceAllocStoneJobResponse,
	reportErr error) {
	if reportErr == nil {
		return
	}
	req := &stypes.StoneHubServiceDoneSecondaryPieceJobRequest{
		TraceId: resp.GetTraceId(),
		ErrMessage: &stypes.ErrMessage{
			ErrCode: stypes.ErrCode_ERR_CODE_ERROR,
			ErrMsg:  reportErr.Error(),
		},
	}
	if _, err := node.stoneHub.DoneSecondaryPieceJob(ctx, req); err != nil {
		log.CtxErrorw(ctx, "report stone hub err msg failed", "error", err)
		return
	}
	log.CtxInfow(ctx, "report stone hub err msg success")
}
