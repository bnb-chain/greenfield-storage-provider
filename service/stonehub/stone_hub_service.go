package stonehub

import (
	"context"
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/stone"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

/* stone_hub_service.go implement StoneHubServiceServer grpc interface.
 * CreateObject and SetObjectCreateHeight implement the first stage of uploading(abandoned).
 * BeginUploadPayload and DonePrimaryPieceJob use to interact with uploader service
 * aim to complete uploading primary storage provider.
 * AllocStoneJob and DoneSecondaryPieceJob use to interact with stone node service
 * aim to complete uploading secondary storage provider.
 */

var _ stypes.StoneHubServiceServer = &StoneHub{}
var _ Stone = &stone.UploadPayloadStone{}

// CreateObject create job and object info, store the DB table, if already exists will return error
func (hub *StoneHub) CreateObject(ctx context.Context, req *stypes.StoneHubServiceCreateObjectRequest) (
	*stypes.StoneHubServiceCreateObjectResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &stypes.StoneHubServiceCreateObjectResponse{
		TraceId: req.TraceId,
	}
	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrInterfaceAbandoned)
	log.CtxErrorw(ctx, "create object interface is abandoned")
	return rsp, nil
}

// SetObjectCreateInfo set CreateObjectTX the height and object resource id on the inscription chain
func (hub *StoneHub) SetObjectCreateInfo(ctx context.Context, req *stypes.StoneHubServiceSetObjectCreateInfoRequest) (
	*stypes.StoneHubServiceSetObjectCreateInfoResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &stypes.StoneHubServiceSetObjectCreateInfoResponse{TraceId: req.TraceId}
	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrInterfaceAbandoned)
	log.CtxErrorw(ctx, "set object create info interface is abandoned")
	return rsp, nil
}

// BeginUploadPayload create upload payload stone and start the fsm to upload
// if the job context or object info is nil in local, will query from inscription chain
func (hub *StoneHub) BeginUploadPayload(ctx context.Context, req *stypes.StoneHubServiceBeginUploadPayloadRequest) (
	*stypes.StoneHubServiceBeginUploadPayloadResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &stypes.StoneHubServiceBeginUploadPayloadResponse{TraceId: req.TraceId}
	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrInterfaceAbandoned)
	log.CtxErrorw(ctx, "set object create info interface is abandoned")
	return rsp, nil
}

// BeginUploadPayloadV2 merge CreateObject, SetObjectCreateInfo and BeginUploadPayload, special for heavy client use.
func (hub *StoneHub) BeginUploadPayloadV2(ctx context.Context, req *stypes.StoneHubServiceBeginUploadPayloadV2Request) (
	resp *stypes.StoneHubServiceBeginUploadPayloadV2Response, err error) {
	ctx = log.Context(ctx, req, req.GetObjectInfo())
	resp = &stypes.StoneHubServiceBeginUploadPayloadV2Response{
		TraceId: req.TraceId,
	}
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
		}
		log.CtxInfow(ctx, "begin upload payload stone completed", "error", err)
	}()
	// 1. set object info to db
	if req.GetObjectInfo() == nil {
		err = merrors.ErrObjectInfoNil
		return
	}
	if req.GetObjectInfo().GetSize() == 0 {
		err = merrors.ErrObjectSizeZero
		return
	}
	if req.GetObjectInfo().GetObjectId() == 0 {
		err = merrors.ErrObjectIdZero
		return
	}
	if req.GetObjectInfo().GetHeight() == 0 {
		err = merrors.ErrObjectHeightZero
		return
	}
	if req.GetObjectInfo().GetPrimarySp().GetSpId() != hub.config.StorageProvider {
		err = merrors.ErrPrimarySPMismatch
	}
	if hub.HasStone(req.GetObjectInfo().GetObjectId()) {
		err = merrors.ErrUploadPayloadJobRunning
		return
	}

	// TODO:: inline type check and change move to gate
	if req.GetObjectInfo().GetSize() <= model.InlineSize {
		req.GetObjectInfo().RedundancyType = ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE
	}

	var (
		jobCtx      *ptypes.JobContext
		uploadStone *stone.UploadPayloadStone
	)
	// create upload stone
	if req.ObjectInfo.JobId, err = hub.jobDB.CreateUploadPayloadJob(req.GetObjectInfo()); err != nil {
		return
	}
	// TODO::CreateUploadPayloadJob return jobContext
	if jobCtx, err = hub.jobDB.GetJobContext(req.GetObjectInfo().GetJobId()); err != nil {
		return
	}
	if uploadStone, err = stone.NewUploadPayloadStone(ctx, jobCtx, req.GetObjectInfo(),
		hub.jobDB, hub.metaDB, hub.jobCh, hub.gcCh); err != nil {
		return
	}
	if uploadStone.PrimarySPJobDone() {
		log.CtxInfow(ctx, "upload primary already completed")
		resp.PrimaryDone = true
	}
	if !hub.SetStoneExclude(uploadStone) {
		err = merrors.ErrUploadPayloadJobRunning
		return
	}
	resp.PieceJob = uploadStone.PopPendingPrimarySPJob()
	return resp, nil
}

// DonePrimaryPieceJob set the primary piece job completed state
func (hub *StoneHub) DonePrimaryPieceJob(ctx context.Context, req *stypes.StoneHubServiceDonePrimaryPieceJobRequest) (
	*stypes.StoneHubServiceDonePrimaryPieceJobResponse, error) {
	ctx = log.Context(ctx, req, req.GetPieceJob())
	resp := &stypes.StoneHubServiceDonePrimaryPieceJobResponse{TraceId: req.TraceId}
	var (
		uploadStone  *stone.UploadPayloadStone
		job          Stone
		interruptErr error
		err          error
		pieceIdx     = -1
		typeCast     bool
	)
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
		}
		if interruptErr != nil && uploadStone != nil {
			err = uploadStone.InterruptStone(ctx, interruptErr)
			log.CtxWarnw(ctx, "interrupt stone error", "error", err)
		}
		log.CtxInfow(ctx, "done primary piece job completed", "piece_idx", pieceIdx, "error", err)
	}()
	if req.GetPieceJob() == nil || req.GetPieceJob().GetObjectId() == 0 {
		err = merrors.ErrObjectIdZero
		return resp, nil
	}
	if job = hub.GetStone(req.GetPieceJob().GetObjectId()); job == nil {
		err = merrors.ErrUploadPayloadJobNotExist
		return resp, nil
	}
	if uploadStone, typeCast = job.(*stone.UploadPayloadStone); !typeCast {
		err = merrors.ErrUploadPayloadJobNotExist
		return resp, nil
	}
	if req.GetErrMessage() != nil && req.GetErrMessage().GetErrCode() ==
		stypes.ErrCode_ERR_CODE_ERROR {
		interruptErr = errors.New(resp.GetErrMessage().GetErrMsg())
		return resp, nil
	}

	if req.GetPieceJob().GetStorageProviderSealInfo() == nil {
		err = merrors.ErrSealInfoMissing
		return resp, nil
	}
	pieceIdx = int(req.GetPieceJob().GetStorageProviderSealInfo().GetPieceIdx())
	if len(req.GetPieceJob().GetStorageProviderSealInfo().GetPieceChecksum()) != 1 {
		err = merrors.ErrCheckSumCountMismatch
		return resp, nil
	}
	if req.GetPieceJob().GetStorageProviderSealInfo().GetStorageProviderId() != hub.config.StorageProvider {
		err = merrors.ErrPrimarySPMismatch
		return resp, nil
	}
	interruptErr = uploadStone.ActionEvent(ctx, stone.UploadPrimaryPieceDoneEvent, req.PieceJob)
	if interruptErr != nil {
		hub.DeleteStone(uploadStone.StoneKey())
		return resp, nil
	}
	return resp, nil
}

// AllocStoneJob pop the secondary piece job
func (hub *StoneHub) AllocStoneJob(ctx context.Context, req *stypes.StoneHubServiceAllocStoneJobRequest) (
	*stypes.StoneHubServiceAllocStoneJobResponse, error) {
	ctx = log.Context(ctx, req)
	resp := &stypes.StoneHubServiceAllocStoneJobResponse{TraceId: req.TraceId}
	stoneJob := hub.ConsumeJob()
	if stoneJob == nil {
		log.CtxDebugw(ctx, "no stone job to dispatch")
		return resp, nil
	}
	switch job := stoneJob.(type) {
	case *stypes.PieceJob:
		objectInfo, _ := hub.jobDB.GetObjectInfo(job.GetObjectId())
		resp.BucketName = objectInfo.BucketName
		resp.ObjectName = objectInfo.ObjectName
		resp.PieceJob = job
	default:
		resp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrStoneJobTypeUnrecognized)
		log.CtxErrorw(ctx, "unrecognized stone job type")
	}
	return resp, nil
}

// DoneSecondaryPieceJob set the secondary piece job completed state
func (hub *StoneHub) DoneSecondaryPieceJob(ctx context.Context, req *stypes.StoneHubServiceDoneSecondaryPieceJobRequest) (
	*stypes.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	ctx = log.Context(ctx, req, req.GetPieceJob())
	resp := &stypes.StoneHubServiceDoneSecondaryPieceJobResponse{TraceId: req.TraceId}
	var (
		uploadStone  *stone.UploadPayloadStone
		job          Stone
		interruptErr error
		err          error
		typeCast     bool
		pieceIdx     = -1
	)
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
		}
		if interruptErr != nil && uploadStone != nil {
			log.CtxErrorw(ctx, "interrupt stone", "error", interruptErr)
			uploadStone.InterruptStone(ctx, interruptErr)
		}
		log.CtxInfow(ctx, "done secondary piece job completed", "piece_idx", pieceIdx, "error", err)
	}()
	if req.GetErrMessage() != nil && req.GetErrMessage().GetErrCode() ==
		stypes.ErrCode_ERR_CODE_ERROR {
		interruptErr = errors.New(resp.GetErrMessage().GetErrMsg())
		return resp, nil
	}
	if req.GetPieceJob() == nil || req.GetPieceJob().GetObjectId() == 0 {
		err = merrors.ErrObjectIdZero
		return resp, nil
	}
	if job = hub.GetStone(req.GetPieceJob().GetObjectId()); job == nil {
		err = merrors.ErrUploadPayloadJobNotExist
		return resp, nil
	}
	if uploadStone, typeCast = job.(*stone.UploadPayloadStone); !typeCast {
		err = merrors.ErrUploadPayloadJobNotExist
		return resp, nil
	}
	if req.GetPieceJob().GetStorageProviderSealInfo() == nil {
		err = merrors.ErrSealInfoMissing
		return resp, nil
	}
	if len(req.GetPieceJob().GetStorageProviderSealInfo().GetPieceChecksum()) == 0 {
		err = merrors.ErrCheckSumCountMismatch
		return resp, nil
	}
	pieceIdx = int(req.GetPieceJob().GetStorageProviderSealInfo().GetPieceIdx())
	if len(req.GetPieceJob().GetStorageProviderSealInfo().GetStorageProviderId()) == 0 {
		err = merrors.ErrStorageProviderMissing
		return resp, nil
	}
	if interruptErr = uploadStone.ActionEvent(ctx, stone.UploadSecondaryPieceDoneEvent, req.PieceJob); interruptErr != nil {
		hub.DeleteStone(uploadStone.StoneKey())
		return resp, nil
	}
	return resp, nil
}

// QueryStone return the stone info, debug interface
func (hub *StoneHub) QueryStone(ctx context.Context, req *stypes.StoneHubServiceQueryStoneRequest) (*stypes.StoneHubServiceQueryStoneResponse, error) {
	// ctx = log.Context(ctx, req)
	rsp := &stypes.StoneHubServiceQueryStoneResponse{}

	st := hub.GetStone(req.GetObjectId())
	uploadStone := st.(*stone.UploadPayloadStone)
	rsp.JobInfo = uploadStone.GetJobContext()
	rsp.PendingPrimaryJob = uploadStone.PopPendingPrimarySPJob()
	rsp.PendingSecondaryJob = uploadStone.PopPendingSecondarySPJob()
	rsp.ObjectInfo, _ = hub.jobDB.GetObjectInfo(uploadStone.GetObjectInfo().GetObjectId())
	return rsp, nil
}
