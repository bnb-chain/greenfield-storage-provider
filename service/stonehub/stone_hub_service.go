package stonehub

import (
	"context"
	"errors"

	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/pkg/stone"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

/* stone_hub_service.go implement StoneHubServiceServer grpc interface.
 * CreateObject and SetObjectCreateHeight implement the first stage of uploading(abandoned).
 * BeginUploadPayload and DonePrimaryPieceJob use to interact with uploader service
 * aim to complete uploading primary storage provider.
 * AllocStoneJob and DoneSecondaryPieceJob use to interact with stone node service
 * aim to complete uploading secondary storage provider.
 */

var _ service.StoneHubServiceServer = &StoneHub{}
var _ Stone = &stone.UploadPayloadStone{}

// CreateObject create job and object info, store the DB table, if already exists will return error
func (hub *StoneHub) CreateObject(ctx context.Context,
	req *service.StoneHubServiceCreateObjectRequest) (
	*service.StoneHubServiceCreateObjectResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceCreateObjectResponse{
		TraceId: req.TraceId,
	}
	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrInterfaceAbandoned)
	log.CtxErrorw(ctx, "create object interface is abandoned")
	return rsp, nil
	//if len(req.TxHash) != hash.LengthHash {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrTxHash)
	//	log.CtxErrorw(ctx, "hash format error")
	//	return rsp, nil
	//}
	//req.ObjectInfo.TxHash = req.TxHash
	//if req.ObjectInfo.Size == 0 {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrObjectSize)
	//	log.CtxErrorw(ctx, "object size error")
	//	return rsp, nil
	//}
	//if req.ObjectInfo.Size <= model.InlineSize {
	//	log.CtxWarnw(ctx, "create object adjust to inline type", "object size", req.ObjectInfo.Size)
	//	req.ObjectInfo.RedundancyType = types.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE
	//}
	//_, err := hub.jobDB.CreateUploadPayloadJob(req.TxHash, req.ObjectInfo)
	//if err != nil {
	//	// maybe query retrieve service
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//	log.CtxErrorw(ctx, "create object error", "error", err)
	//	return rsp, nil
	//}
	//log.CtxInfow(ctx, "create object success")
	//return rsp, nil
}

// SetObjectCreateInfo set CreateObjectTX the height and object resource id on the inscription chain
func (hub *StoneHub) SetObjectCreateInfo(ctx context.Context,
	req *service.StoneHubServiceSetObjectCreateInfoRequest) (
	*service.StoneHubServiceSetObjectCreateInfoResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceSetObjectCreateInfoResponse{TraceId: req.TraceId}
	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrInterfaceAbandoned)
	log.CtxErrorw(ctx, "set object create info interface is abandoned")
	return rsp, nil
	//if len(req.TxHash) != hash.LengthHash {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrTxHash)
	//	log.CtxErrorw(ctx, "hash format error")
	//	return rsp, nil
	//}
	//if req.ObjectId == 0 {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrObjectID)
	//	log.CtxErrorw(ctx, "object id error", "ObjectId", req.ObjectId)
	//	return rsp, nil
	//}
	//if req.TxHeight == 0 {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrObjectCreateHeight)
	//	log.CtxErrorw(ctx, "create object height error", "Height", req.TxHeight)
	//	return rsp, nil
	//}
	//if err := hub.jobDB.SetObjectCreateHeightAndObjectID(req.TxHash, req.TxHeight, req.ObjectId); err != nil {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//	log.CtxErrorw(ctx, "set object height and object id error", "error", err)
	//	return rsp, nil
	//}
	//log.CtxInfow(ctx, "set object create height and object id success")
	//return rsp, nil
}

// BeginUploadPayload create upload payload stone and start the fsm to upload
// if the job context or object info is nil in local, will query from inscription chain
func (hub *StoneHub) BeginUploadPayload(ctx context.Context,
	req *service.StoneHubServiceBeginUploadPayloadRequest) (
	*service.StoneHubServiceBeginUploadPayloadResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceBeginUploadPayloadResponse{TraceId: req.TraceId}
	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrInterfaceAbandoned)
	log.CtxErrorw(ctx, "set object create info interface is abandoned")
	return rsp, nil
	//if len(req.TxHash) != hash.LengthHash {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrTxHash)
	//	log.CtxErrorw(ctx, "hash format error")
	//	return rsp, nil
	//}
	//// check the stone whether already running
	//if hub.HasStone(string(req.TxHash)) {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrUploadPayloadJobRunning)
	//	log.CtxErrorw(ctx, "upload payload stone is running")
	//	return rsp, nil
	//}
	//// load the stone context from db
	//object, err := hub.jobDB.GetObjectInfo(req.TxHash)
	//if err != nil {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//	log.CtxErrorw(ctx, "get object info error", "error", err)
	//	return rsp, nil
	//}
	//jobCtx, err := hub.jobDB.GetJobContext(object.JobId)
	//if err != nil {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//	log.CtxErrorw(ctx, "get job info error", "error", err)
	//	return rsp, nil
	//}
	//// the stone context is nil, query from inscription-chain
	//if jobCtx == nil {
	//	log.CtxWarnw(ctx, "query object info from inscription chain")
	//	objectInfo, err := hub.insCli.QueryObjectByTx(req.TxHash)
	//	if err != nil {
	//		rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//		log.CtxErrorw(ctx, "query inscription chain error", "error", err)
	//		return rsp, nil
	//	}
	//	if objectInfo == nil {
	//		rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrObjectInfoOnInscription)
	//		log.CtxErrorw(ctx, "object is not on inscription chain")
	//		return rsp, nil
	//	}
	//	// the temporary solution determine whether the seal object is successful
	//	// TBD :: inscription client will return the object info type to determine
	//	if len(objectInfo.SecondarySps) > 0 {
	//		rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrUploadPayloadJobDone)
	//		log.CtxWarnw(ctx, "payload has uploaded")
	//		return rsp, nil
	//	}
	//	_, err = hub.jobDB.CreateUploadPayloadJob(req.TxHash, objectInfo)
	//	if err != nil {
	//		rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//		log.CtxErrorw(ctx, "create upload payload job error", "error", err)
	//		return rsp, nil
	//	}
	//	object, err = hub.jobDB.GetObjectInfo(req.TxHash)
	//	if err != nil {
	//		rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//		log.CtxErrorw(ctx, "get object info error", "error", err)
	//		return rsp, nil
	//	}
	//	jobCtx, err = hub.jobDB.GetJobContext(object.JobId)
	//	if err != nil {
	//		rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//		log.CtxErrorw(ctx, "get job info error", "error", err)
	//		return rsp, nil
	//	}
	//}
	//// the stone context is ready
	//uploadStone, err := stone.NewUploadPayloadStone(ctx, jobCtx, object, hub.jobDB, hub.metaDB, hub.jobCh, hub.stoneGC)
	//if err != nil {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(err)
	//	log.CtxErrorw(ctx, "create upload payload stone error", "error", err)
	//	return rsp, nil
	//}
	//if uploadStone.PrimarySPJobDone() {
	//	log.CtxInfow(ctx, "upload primary storage provider has completed")
	//	rsp.PrimaryDone = true
	//}
	//if !hub.SetStoneExclude(uploadStone) {
	//	rsp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrUploadPayloadJobRunning)
	//	log.CtxErrorw(ctx, "add upload payload stone error", "error", err)
	//	return rsp, nil
	//}
	//rsp.PieceJob = uploadStone.PopPendingPrimarySPJob()
	//log.CtxInfow(ctx, "begin upload payload success")
	//return rsp, nil
}

// BeginUploadPayloadV2 merge CreateObject, SetObjectCreateInfo and BeginUploadPayload, special for heavy client use.
func (hub *StoneHub) BeginUploadPayloadV2(ctx context.Context,
	req *service.StoneHubServiceBeginUploadPayloadV2Request) (
	resp *service.StoneHubServiceBeginUploadPayloadV2Response, err error) {
	ctx = log.Context(ctx, req, req.GetObjectInfo())
	resp = &service.StoneHubServiceBeginUploadPayloadV2Response{
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
	var (
		jobCtx      *types.JobContext
		uploadStone *stone.UploadPayloadStone
	)
	// create upload stone
	if req.ObjectInfo.JobId, err = hub.jobDB.CreateUploadPayloadJob(
		req.GetObjectInfo().GetTxHash(), req.GetObjectInfo()); err != nil {
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
func (hub *StoneHub) DonePrimaryPieceJob(ctx context.Context,
	req *service.StoneHubServiceDonePrimaryPieceJobRequest) (
	*service.StoneHubServiceDonePrimaryPieceJobResponse, error) {
	ctx = log.Context(ctx, req, req.GetPieceJob())
	resp := &service.StoneHubServiceDonePrimaryPieceJobResponse{TraceId: req.TraceId, TxHash: req.TxHash}
	var (
		uploadStone  *stone.UploadPayloadStone
		job          Stone
		interruptErr error
		err          error
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
		log.CtxInfow(ctx, "done primary piece job completed", "error", err)
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
		service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		interruptErr = errors.New(resp.GetErrMessage().GetErrMsg())
		return resp, nil
	}

	if req.GetPieceJob().GetStorageProviderSealInfo() == nil {
		err = merrors.ErrSealInfoMissing
		return resp, nil
	}
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
		return resp, nil
	}
	return resp, nil
}

// AllocStoneJob pop the secondary piece job
func (hub *StoneHub) AllocStoneJob(ctx context.Context,
	req *service.StoneHubServiceAllocStoneJobRequest) (
	*service.StoneHubServiceAllocStoneJobResponse, error) {
	ctx = log.Context(ctx, req)
	resp := &service.StoneHubServiceAllocStoneJobResponse{TraceId: req.TraceId}
	stoneJob := hub.ConsumeJob()
	if stoneJob == nil {
		log.CtxDebugw(ctx, "no stone job to dispatch")
		return resp, nil
	}
	switch job := stoneJob.(type) {
	case *service.PieceJob:
		resp.PieceJob = job
	default:
		resp.ErrMessage = merrors.MakeErrMsgResponse(merrors.ErrStoneJobTypeUnrecognized)
		log.CtxErrorw(ctx, "unrecognized stone job type")
	}
	return resp, nil
}

// DoneSecondaryPieceJob set the secondary piece job completed state
func (hub *StoneHub) DoneSecondaryPieceJob(ctx context.Context,
	req *service.StoneHubServiceDoneSecondaryPieceJobRequest) (
	*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	ctx = log.Context(ctx, req, req.GetPieceJob())
	resp := &service.StoneHubServiceDoneSecondaryPieceJobResponse{TraceId: req.TraceId}
	var (
		uploadStone  *stone.UploadPayloadStone
		job          Stone
		interruptErr error
		err          error
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
		log.CtxInfow(ctx, "done secondary piece job completed", "error", err)
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
		service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		interruptErr = errors.New(resp.GetErrMessage().GetErrMsg())
		return resp, nil
	}

	if req.GetPieceJob().GetStorageProviderSealInfo() == nil {
		err = merrors.ErrSealInfoMissing
		return resp, nil
	}
	if len(req.GetPieceJob().GetStorageProviderSealInfo().GetPieceChecksum()) > 0 {
		err = merrors.ErrCheckSumCountMismatch
		return resp, nil
	}
	if len(req.GetPieceJob().GetStorageProviderSealInfo().GetStorageProviderId()) == 0 {
		err = merrors.ErrStorageProviderMissing
		return resp, nil
	}
	if interruptErr = uploadStone.ActionEvent(ctx, stone.UploadSecondaryPieceDoneEvent, req.PieceJob); interruptErr != nil {
		return resp, nil
	}
	return resp, nil
}

// QueryStone return the stone info, debug interface
func (hub *StoneHub) QueryStone(ctx context.Context, req *service.StoneHubServiceQueryStoneRequest) (*service.StoneHubServiceQueryStoneResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceQueryStoneResponse{}

	st := hub.GetStone(req.GetObjectId())
	uploadStone := st.(*stone.UploadPayloadStone)
	rsp.JobInfo = uploadStone.GetJobContext()
	rsp.PendingPrimaryJob = uploadStone.PopPendingPrimarySPJob()
	rsp.PendingSecondaryJob = uploadStone.PopPendingSecondarySPJob()
	rsp.ObjectInfo, _ = hub.jobDB.GetObjectInfo(uploadStone.GetObjectInfo().GetTxHash())
	return rsp, nil
}
