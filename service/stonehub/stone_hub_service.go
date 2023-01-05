package stonehub

import (
	"context"
	"errors"

	"github.com/bnb-chain/inscription-storage-provider/model"
	errors2 "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/pkg/stone"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

/* stone_hub_service.go implement StoneHubServiceServer grpc interface.
 * CreateObject and SetObjectCreateHeight implement the first stage of uploading.
 * BeginUploadPayload and DonePrimaryPieceJob use to interact with uploader service
 * aim to complete uploading primary storage provider.
 * AllocStoneJob and DoneSecondaryPieceJob use to interact with stone node service
 * aim to complete uploading secondary storage provider.
 */

var _ service.StoneHubServiceServer = &StoneHub{}
var _ Stone = &stone.UploadPayloadStone{}

// CreateObject create job and object info, store the DB table, if already exists will return error
func (hub *StoneHub) CreateObject(ctx context.Context, req *service.StoneHubServiceCreateObjectRequest) (*service.StoneHubServiceCreateObjectResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceCreateObjectResponse{
		TraceId: req.TraceId,
		TxHash:  req.TxHash,
	}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrTxHash)
		log.CtxErrorw(ctx, "hash format error")
		return rsp, nil
	}
	req.ObjectInfo.TxHash = req.TxHash
	if req.ObjectInfo.Size == 0 {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrObjectSize)
		log.CtxErrorw(ctx, "object size error")
		return rsp, nil
	}
	if req.ObjectInfo.Size <= model.InlineSize {
		log.CtxWarnw(ctx, "create object adjust to inline type", "object size", req.ObjectInfo.Size)
		req.ObjectInfo.RedundancyType = types.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE
	}
	err := hub.jobDB.CreateUploadPayloadJob(req.TxHash, req.ObjectInfo)
	if err != nil {
		// maybe query retrieve service
		rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
		log.CtxErrorw(ctx, "create object error", "error", err)
		return rsp, nil
	}
	log.CtxInfow(ctx, "create object success")
	return rsp, nil
}

// SetObjectCreateInfo set CreateObjectTX the height and object resource id on the inscription chain
func (hub *StoneHub) SetObjectCreateInfo(ctx context.Context, req *service.StoneHubServiceSetObjectCreateInfoRequest) (*service.StoneHubServiceSetSetObjectCreateInfoResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceSetSetObjectCreateInfoResponse{TraceId: req.TraceId}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrTxHash)
		log.CtxErrorw(ctx, "hash format error")
		return rsp, nil
	}
	if req.ObjectId == 0 {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrObjectID)
		log.CtxErrorw(ctx, "object id error", "ObjectId", req.ObjectId)
		return rsp, nil
	}
	if req.TxHeight == 0 {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrObjectCreateHeight)
		log.CtxErrorw(ctx, "create object height error", "Height", req.TxHeight)
		return rsp, nil
	}
	if err := hub.jobDB.SetObjectCreateHeightAndObjectID(req.TxHash, req.TxHeight, req.ObjectId); err != nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
		log.CtxErrorw(ctx, "set object height and object id error", "error", err)
		return rsp, nil
	}
	log.CtxInfow(ctx, "set object create height and object id success")
	return rsp, nil
}

// BeginUploadPayload create upload payload stone and start the fsm to upload
// if the job context or object info is nil in local, will query from inscription chain
func (hub *StoneHub) BeginUploadPayload(ctx context.Context, req *service.StoneHubServiceBeginUploadPayloadRequest) (*service.StoneHubServiceBeginUploadPayloadResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceBeginUploadPayloadResponse{TraceId: req.TraceId}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrTxHash)
		log.CtxErrorw(ctx, "hash format error")
		return rsp, nil
	}
	// check the stone whether already running
	if hub.HasStone(string(req.TxHash)) {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrUploadPayloadJobRunning)
		log.CtxErrorw(ctx, "upload payload stone is running")
		return rsp, nil
	}
	// load the stone context from db
	object, err := hub.jobDB.GetObjectInfo(req.TxHash)
	if err != nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
		log.CtxErrorw(ctx, "get object info error", "error", err)
		return rsp, nil
	}
	jobCtx, err := hub.jobDB.GetJobContext(object.JobId)
	if err != nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
		log.CtxErrorw(ctx, "get job info error", "error", err)
		return rsp, nil
	}
	// the stone context is nil, query from inscription-chain
	if jobCtx == nil {
		log.CtxWarnw(ctx, "query object info from inscription chain")
		objectInfo, err := hub.insCli.QueryObjectByTx(req.TxHash)
		if err != nil {
			rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "query inscription chain error", "error", err)
			return rsp, nil
		}
		if objectInfo == nil {
			rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrObjectInfoOnInscription)
			log.CtxErrorw(ctx, "object is not on inscription chain")
			return rsp, nil
		}
		// the temporary solution determine whether the seal object is successful
		// TBD :: inscription client will return the object info type to determine
		if len(objectInfo.SecondarySps) > 0 {
			rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrUploadPayloadJobDone)
			log.CtxWarnw(ctx, "payload has uploaded")
			return rsp, nil
		}
		err = hub.jobDB.CreateUploadPayloadJob(req.TxHash, objectInfo)
		if err != nil {
			rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "create upload payload job error", "error", err)
			return rsp, nil
		}
		object, err = hub.jobDB.GetObjectInfo(req.TxHash)
		if err != nil {
			rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "get object info error", "error", err)
			return rsp, nil
		}
		jobCtx, err = hub.jobDB.GetJobContext(object.JobId)
		if err != nil {
			rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "get job info error", "error", err)
			return rsp, nil
		}
	}
	// the stone context is ready
	uploadStone, err := stone.NewUploadPayloadStone(ctx, jobCtx, object, hub.jobDB, hub.metaDB, hub.jobCh, hub.stoneGC)
	if err != nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
		log.CtxErrorw(ctx, "create upload payload stone error", "error", err)
		return rsp, nil
	}
	if uploadStone.PrimarySPJobDone() {
		log.CtxInfow(ctx, "upload primary storage provider has completed")
		rsp.PrimaryDone = true
	}
	if !hub.SetStoneExclude(uploadStone) {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrUploadPayloadJobRunning)
		log.CtxErrorw(ctx, "add upload payload stone error", "error", err)
		return rsp, nil
	}
	rsp.PieceJob = uploadStone.PopPendingPrimarySPJob()
	log.CtxInfow(ctx, "begin upload payload success")
	return rsp, nil
}

// DonePrimaryPieceJob set the primary piece job completed state
func (hub *StoneHub) DonePrimaryPieceJob(ctx context.Context, req *service.StoneHubServiceDonePrimaryPieceJobRequest) (*service.StoneHubServiceDonePrimaryPieceJobResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceDonePrimaryPieceJobResponse{TraceId: req.TraceId, TxHash: req.TxHash}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrTxHash)
		log.CtxErrorw(ctx, "hash format error")
		return rsp, nil
	}
	if req.PieceJob == nil || req.PieceJob.StorageProviderSealInfo == nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrParamMissing)
		log.CtxErrorw(ctx, "params missing error")
		return rsp, nil
	}
	if req.PieceJob.StorageProviderSealInfo.StorageProviderId != hub.config.StorageProvider {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrPrimaryStorageProvider)
		log.CtxErrorw(ctx, "primary storage provider mismatch")
		return rsp, nil
	}

	if len(req.PieceJob.StorageProviderSealInfo.PieceChecksum) != 1 {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrPrimaryPieceChecksum)
		log.CtxErrorw(ctx, "primary storage provider piece job checksum error")
		return rsp, nil
	}
	if req.PieceJob.StorageProviderSealInfo.StorageProviderId != hub.config.StorageProvider {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrPrimaryStorageProvider)
		log.Error("tx hash format error", "trace_id", req.TraceId, "hash", req.TxHash)
		return rsp, nil
	}

	if len(req.PieceJob.StorageProviderSealInfo.PieceChecksum) != 1 {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrPrimaryPieceChecksum)
		log.Error("tx hash format error", "trace_id", req.TraceId, "hash", req.TxHash)
		return rsp, nil
	}
	st := hub.GetStone(string(req.TxHash))
	if st == nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrUploadPayloadJobNotExist)
		log.CtxErrorw(ctx, "upload payload stone not exist")
		return rsp, nil
	}
	uploadStone := st.(*stone.UploadPayloadStone)
	if req.ErrMessage != nil && req.ErrMessage.ErrCode != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "done primary job error", "error", req.ErrMessage)
		if err := uploadStone.InterruptStone(ctx, errors.New(req.ErrMessage.ErrMsg)); err != nil {
			rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "interrupt stone error", "error", err)
		}
		return rsp, nil
	}
	if err := uploadStone.ActionEvent(ctx, stone.UploadPrimaryPieceDoneEvent, req.PieceJob); err != nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
		log.CtxErrorw(ctx, "action(UploadPrimaryPieceDoneEvent) stone fsm error", "error", err)
		return rsp, nil
	}
	log.CtxInfow(ctx, "done primary piece job success", "piece idx", req.PieceJob.StorageProviderSealInfo.PieceIdx)
	return rsp, nil
}

// AllocStoneJob pop the secondary piece job
func (hub *StoneHub) AllocStoneJob(ctx context.Context, req *service.StoneHubServiceAllocStoneJobRequest) (*service.StoneHubServiceAllocStoneJobResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceAllocStoneJobResponse{TraceId: req.TraceId}
	pieceJob := hub.PopUploadSecondaryPieceJob()
	if pieceJob != nil {
		rsp.TxHash = pieceJob.TxHash
		rsp.PieceJob = pieceJob
		log.CtxInfow(ctx, "dispatch stone job", "piece job info", pieceJob)
	} else {
		log.CtxDebugw(ctx, "no stone job to alloc")
	}
	return rsp, nil
}

// DoneSecondaryPieceJob set the secondary piece job completed state
func (hub *StoneHub) DoneSecondaryPieceJob(ctx context.Context, req *service.StoneHubServiceDoneSecondaryPieceJobRequest) (*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	rsp := &service.StoneHubServiceDoneSecondaryPieceJobResponse{TraceId: req.TraceId}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrTxHash)
		log.CtxErrorw(ctx, "hash format error")
		return rsp, nil
	}
	if req.PieceJob == nil || req.PieceJob.StorageProviderSealInfo == nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrParamMissing)
		log.CtxErrorw(ctx, "params missing error")
		return rsp, nil
	}
	st := hub.GetStone(string(req.TxHash))
	if st == nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrUploadPayloadJobNotExist)
		log.CtxErrorw(ctx, "upload payload stone not exist")
		return rsp, nil
	}
	uploadStone := st.(*stone.UploadPayloadStone)
	if req.ErrMessage != nil && req.ErrMessage.ErrCode != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.CtxErrorw(ctx, "done secondary job error", "error", req.ErrMessage)
		if err := uploadStone.InterruptStone(ctx, errors.New(req.ErrMessage.ErrMsg)); err != nil {
			rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "interrupt stone error", "error", err)
		}
		return rsp, nil
	}
	if err := uploadStone.ActionEvent(ctx, stone.UploadSecondaryPieceDoneEvent, req.PieceJob); err != nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(err)
		log.CtxErrorw(ctx, "action(UploadSecondaryPieceDoneEvent) stone fsm error", "error", err)
		return rsp, nil
	}
	log.CtxInfow(ctx, "done secondary piece job success", "piece idx", req.PieceJob.StorageProviderSealInfo.PieceIdx)
	return rsp, nil
}

// QueryStone return the stone info
func (hub *StoneHub) QueryStone(ctx context.Context, req *service.StoneHubServiceQueryStoneRequest) (*service.StoneHubServiceQueryStoneResponse, error) {
	ctx = log.Context(ctx, req)
	rsp := &service.StoneHubServiceQueryStoneResponse{}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrTxHash)
		log.CtxErrorw(ctx, "hash format error")
		return rsp, nil
	}
	rsp.ObjectInfo, _ = hub.jobDB.GetObjectInfo(req.TxHash)
	st := hub.GetStone(string(req.TxHash))
	if st == nil {
		rsp.ErrMessage = errors2.MakeErrMsgResponse(errors2.ErrUploadPayloadJobNotExist)
		log.CtxErrorw(ctx, "upload payload stone not exist")
		return rsp, nil
	}
	uploadStone := st.(*stone.UploadPayloadStone)
	jobInfo := uploadStone.GetJobContext()
	rsp.JobInfo = &jobInfo
	rsp.PendingPrimaryJob = uploadStone.PopPendingPrimarySPJob()
	rsp.PendingSecondaryJob = uploadStone.PopPendingSecondarySPJob()
	return rsp, nil
}
