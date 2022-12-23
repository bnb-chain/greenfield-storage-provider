package stonehub

import (
	"context"
	"github.com/bnb-chain/inscription-storage-provider/pkg/stone"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
)

var _ service.StoneHubServiceServer = &StoneHub{}

func (hub *StoneHub) CreateObject(ctx context.Context, req *service.StoneHubServiceCreateObjectRequest) (*service.StoneHubServiceCreateObjectResponse, error) {
	if req == nil {
		// return error in response
		return nil, nil
	}
	if len(req.TxHash) != hash.LengthHash {
		// return error in response
		return nil, nil
	}
	if req.ObjectInfo.Size <= hub.GetInlineSize() {
		req.ObjectInfo.RedundancyType = types.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE
	}
	_, err := hub.JobDB.CreateUploadPayloadJob(req.TxHash, req.ObjectInfo)
	if err != nil {
		// return error in response
		return nil, nil
	}
	rsp := &service.StoneHubServiceCreateObjectResponse{
		TraceId: req.TraceId,
		TxHash:  []byte{},
	}
	return rsp, nil
}

func (hub *StoneHub) SetObjectCreateHeight(ctx context.Context, req *service.StoneHubServiceSetObjectCreateHeightRequest) (*service.StoneHubServiceSetObjectCreateHeightResponse, error) {
	if req == nil {
		// return error in response
		return nil, nil
	}
	if len(req.TxHash) != hash.LengthHash {
		// return error in response
		return nil, nil
	}
	if err := hub.JobDB.SetObjectCreateHeight(req.TxHash, req.TxHeight); err != nil {
		// return error in response
		return nil, nil
	}
	rsp := &service.StoneHubServiceSetObjectCreateHeightResponse{
		TraceId: req.TraceId,
	}
	return rsp, nil
}

func (hub *StoneHub) BeginUploadPayload(ctx context.Context, req *service.StoneHubServiceBeginUploadPayloadRequest) (*service.StoneHubServiceBeginUploadPayloadResponse, error) {
	if req == nil {
		// return error in response
		return nil, nil
	}
	if len(req.TxHash) != hash.LengthHash {
		// return error in response
		return nil, nil
	}
	stoneInstance, err := hub.GetStoneByTxHash(req.TxHash)
	if err != nil {
		// return error in response
		return nil, nil
	}
	if stoneInstance != nil {
		// return in response
		return nil, nil
	}
	jobCtx, err := hub.JobDB.GetJobContext(req.TxHash)
	if err != nil {
		// return error in response
		return nil, nil
	}
	uploadStone, err := stone.NewUploadPayloadStone(jobCtx, hub.JobDB, hub.MetaDB, hub.stoneJobCh, hub.stoneGC)
	if err != nil {
		// return error in response
		return nil, nil
	}
	current, err := uploadStone.GetJobState()
	if err != nil {
		// return error in response
		return nil, nil
	}
	rsp := &service.StoneHubServiceBeginUploadPayloadResponse{}
	if types.JobState_value[current] >= types.JobState_value[types.JOB_STATE_UPLOAD_PRIMARY_DONE] {
		rsp.PrimaryDone = true
	}
	if current != types.JOB_STATE_SEAL_OBJECT_DONE {
		if err := hub.SetStone(uploadStone); err != nil {
			// return error in response
			return nil, nil
		}
	}
	rsp.TraceId = req.TraceId
	rsp.TxHash = jobCtx.ObjectInfo.TxHash
	rsp.PayloadSize = jobCtx.ObjectInfo.Size
	rsp.SegmentSize = hub.GetSegmentsSize()
	rsp.RedundancyType = jobCtx.ObjectInfo.RedundancyType
	return rsp, nil
}

func (hub *StoneHub) DonePrimaryPieceJob(ctx context.Context, req *service.StoneHubServiceDonePrimaryPieceJobRequest) (*service.StoneHubServiceDonePrimaryPieceJobResponse, error) {
	if req == nil {
		// return error in response
		return nil, nil
	}
	if len(req.TxHash) != hash.LengthHash {
		// return error in response
		return nil, nil
	}
	stoneInstance, err := hub.GetStoneByTxHash(req.TxHash)
	if err != nil {
		// return error in response
		return nil, nil
	}
	uploadStone := stoneInstance.(stone.UploadPayloadStone)
	if req.ErrMessage.ErrCode != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		// set error
		// stone.SetJobErr
		if err := uploadStone.ActionEvent(ctx, stone.InterruptEvent); err != nil {
			// return error in response
			return nil, nil
		}
	}
	if err := uploadStone.ActionEvent(ctx, stone.UploadPrimaryPieceDoneEvent, req.PieceJob); err != nil {
		// return error in response
		return nil, nil
	}
	rsp := &service.StoneHubServiceDonePrimaryPieceJobResponse{
		TraceId: req.TraceId,
		TxHash:  req.TxHash,
	}
	return rsp, nil
}
func (hub *StoneHub) AllocStoneJob(ctx context.Context, req *service.StoneHubServiceAllocStoneJobRequest) (*service.StoneHubServiceAllocStoneJobResponse, error) {
	if req == nil {
		// return error in response
		return nil, nil
	}
	rsp := &service.StoneHubServiceAllocStoneJobResponse{}
	pieceJob := hub.PopUploadSecondaryPieceJob()
	if pieceJob != nil {
		rsp.TraceId = req.TraceId
		rsp.TxHash = pieceJob.TxHash
		rsp.PieceJob = pieceJob
	} else {
		rsp.TraceId = req.TraceId
	}
	return rsp, nil
}
func (hub *StoneHub) DoneSecondaryPieceJob(ctx context.Context, req *service.StoneHubServiceDoneSecondaryPieceJobRequest) (*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	if req == nil {
		// return error in response
		return nil, nil
	}
	stoneInstance, err := hub.GetStoneByTxHash(req.TxHash)
	if err != nil {
		// return error in response
		return nil, nil
	}
	uploadStone := stoneInstance.(stone.UploadPayloadStone)
	if req.ErrMessage.ErrCode != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		// set error
		// stone.SetJobErr
		if err := uploadStone.ActionEvent(ctx, stone.InterruptEvent); err != nil {
			// return error in response
			return nil, nil
		}
	}
	if err := uploadStone.ActionEvent(ctx, stone.UploadSecondaryPieceDoneEvent, req.PieceJob); err != nil {
		// return error in response
		return nil, nil
	}
	rsp := &service.StoneHubServiceDoneSecondaryPieceJobResponse{
		TraceId: req.TraceId,
	}
	return rsp, nil
}
