package stonehub

import (
	"context"
	"errors"

	"github.com/bnb-chain/inscription-storage-provider/model"
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
	rsp := &service.StoneHubServiceCreateObjectResponse{
		TraceId: req.TraceId,
		TxHash:  req.TxHash,
	}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrTxHash)
		log.Error("create object error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	if req.ObjectInfo.Size <= model.InlineSize {
		log.Warn("create object adjust to inline type", "trace_id", req.TraceId, "hash", req.TxHash)
		req.ObjectInfo.RedundancyType = types.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE
	}
	err := hub.jobDB.CreateUploadPayloadJob(req.TxHash, req.ObjectInfo)
	if err != nil {
		// maybe query retrieve service
		rsp.ErrMessage = model.MakeErrMsgResponse(err)
		log.Error("create object error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	log.Info("create object success", "trace_id", req.TraceId, "hash", req.TxHash)
	return rsp, nil
}

// SetObjectCreateInfo set CreateObjectTX the height and object resource id on the inscription chain
func (hub *StoneHub) SetObjectCreateInfo(ctx context.Context, req *service.StoneHubServiceSetObjectCreateInfoRequest) (*service.StoneHubServiceSetSetObjectCreateInfoResponse, error) {
	rsp := &service.StoneHubServiceSetSetObjectCreateInfoResponse{TraceId: req.TraceId}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrTxHash)
		log.Error("set object height error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	if err := hub.jobDB.SetObjectCreateHeightAndObjectID(req.TxHash, req.TxHeight, req.ObjectId); err != nil {
		rsp.ErrMessage = model.MakeErrMsgResponse(err)
		log.Error("set object height error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	log.Info("set object height success", "trace_id", req.TraceId, "hash", req.TxHash)
	return rsp, nil
}

// BeginUploadPayload create upload payload stone and start the fsm to upload
// if the job context or object info is nil in local, will query from inscription chain
func (hub *StoneHub) BeginUploadPayload(ctx context.Context, req *service.StoneHubServiceBeginUploadPayloadRequest) (*service.StoneHubServiceBeginUploadPayloadResponse, error) {
	rsp := &service.StoneHubServiceBeginUploadPayloadResponse{TraceId: req.TraceId}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrTxHash)
		log.Error("tx hash format error", "trace_id", req.TraceId, "hash", req.TxHash)
		return rsp, nil
	}
	// check the stone whether already running
	if hub.HasStone(string(req.TxHash)) {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrUploadPayloadJobRunning)
		log.Error("upload payload stone is running", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	// load the stone context from db
	object, err := hub.jobDB.GetObjectInfo(req.TxHash)
	if err != nil {
		rsp.ErrMessage = model.MakeErrMsgResponse(err)
		log.Error("get object info error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	jobCtx, err := hub.jobDB.GetJobContext(object.JobId)
	if err != nil {
		rsp.ErrMessage = model.MakeErrMsgResponse(err)
		log.Error("get job info error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	// the stone context is nil, query from inscription-chain
	if jobCtx == nil {
		log.Debug("query object info from inscription chain", "trace_id", req.TraceId, "hash", req.TxHash)
		objectInfo, err := hub.insCli.QueryObjectByTx(req.TxHash)
		if err != nil {
			rsp.ErrMessage = model.MakeErrMsgResponse(err)
			log.Error("query inscription chain error", "trace_id", req.TraceId, "hash", req.TxHash, "error", err)
			return rsp, nil
		}
		if objectInfo == nil {
			rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrObjectInfoOnInscription)
			log.Error("object is not on inscription chain", "trace_id", req.TraceId, "hash", req.TxHash)
			return rsp, nil
		}
		// the temporary solution determine whether the seal object is successful
		// TBD :: inscription client will return the object info type to determine
		if len(objectInfo.SecondarySps) > 0 {
			rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrUploadPayloadJobDone)
			log.Error("payload has uploaded", "trace_id", req.TraceId, "hash", req.TxHash)
			return rsp, nil
		}
		err = hub.jobDB.CreateUploadPayloadJob(req.TxHash, objectInfo)
		if err != nil {
			rsp.ErrMessage = model.MakeErrMsgResponse(err)
			log.Error("create upload payload job error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
			return rsp, nil
		}
		object, err = hub.jobDB.GetObjectInfo(req.TxHash)
		if err != nil {
			rsp.ErrMessage = model.MakeErrMsgResponse(err)
			log.Error("get object info error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
			return rsp, nil
		}
		jobCtx, err = hub.jobDB.GetJobContext(object.JobId)
		if err != nil {
			rsp.ErrMessage = model.MakeErrMsgResponse(err)
			log.Error("get job info error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
			return rsp, nil
		}
	}
	// the stone context is ready
	uploadStone, err := stone.NewUploadPayloadStone(jobCtx, object, hub.jobDB, hub.metaDB, hub.jobCh, hub.stoneGC)
	if err != nil {
		rsp.ErrMessage = model.MakeErrMsgResponse(err)
		log.Error("create upload payload stone error", "trace_id", req.TraceId, "hash", req.TxHash, "error", err)
		return rsp, nil
	}
	if uploadStone.PrimarySPJobDone() {
		log.Info("primary is done", "trace_id", req.TraceId, "hash", req.TxHash)
		rsp.PrimaryDone = true
	}
	if !hub.SetStoneExclude(uploadStone) {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrUploadPayloadJobRunning)
		log.Error("set upload payload stone error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	rsp.PieceJob = uploadStone.PopPendingPrimarySPJob()
	log.Info("begin upload payload success", "trace_id", req.TraceId, "hash", req.TxHash)
	return rsp, nil
}

// DonePrimaryPieceJob set the primary piece job completed state
func (hub *StoneHub) DonePrimaryPieceJob(ctx context.Context, req *service.StoneHubServiceDonePrimaryPieceJobRequest) (*service.StoneHubServiceDonePrimaryPieceJobResponse, error) {
	rsp := &service.StoneHubServiceDonePrimaryPieceJobResponse{TraceId: req.TraceId, TxHash: req.TxHash}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrTxHash)
		log.Error("tx hash format error", "trace_id", req.TraceId, "hash", req.TxHash)
		return rsp, nil
	}
	st := hub.GetStone(string(req.TxHash))
	if st == nil {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrUploadPayloadJobNotExist)
		log.Error("get stone error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	uploadStone := st.(*stone.UploadPayloadStone)
	if req.ErrMessage.ErrCode != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Error("done primary job error", "trace_id", req.TraceId, "hash", req.TxHash, "error", req.ErrMessage)
		if err := uploadStone.InterruptStone(ctx, errors.New(req.ErrMessage.ErrMsg)); err != nil {
			rsp.ErrMessage = model.MakeErrMsgResponse(err)
			log.Error("interrupt stone error", "trace_id", req.TraceId, "hash", req.TxHash, "error", req.ErrMessage)
		}
		return rsp, nil
	}
	if err := uploadStone.ActionEvent(ctx, stone.UploadPrimaryPieceDoneEvent, req.PieceJob); err != nil {
		rsp.ErrMessage = model.MakeErrMsgResponse(err)
		log.Error("action stone fsm error", "trace_id", req.TraceId, "hash", req.TxHash, "error", req.ErrMessage)
	} else {
		log.Info("done primary piece job success", "trace_id", req.TraceId, "hash", req.TxHash)
	}
	return rsp, nil
}

// AllocStoneJob pop the secondary piece job
func (hub *StoneHub) AllocStoneJob(ctx context.Context, req *service.StoneHubServiceAllocStoneJobRequest) (*service.StoneHubServiceAllocStoneJobResponse, error) {
	rsp := &service.StoneHubServiceAllocStoneJobResponse{TraceId: req.TraceId}
	pieceJob := hub.PopUploadSecondaryPieceJob()
	if pieceJob != nil {
		rsp.TxHash = pieceJob.TxHash
		rsp.PieceJob = pieceJob
		log.Debug("dispatch stone job", "piece job info", pieceJob)
	} else {
		log.Error("no stone job to alloc", "trace_id", req.TraceId)
	}
	return rsp, nil
}

// DoneSecondaryPieceJob set the secondary piece job completed state
func (hub *StoneHub) DoneSecondaryPieceJob(ctx context.Context, req *service.StoneHubServiceDoneSecondaryPieceJobRequest) (*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	rsp := &service.StoneHubServiceDoneSecondaryPieceJobResponse{TraceId: req.TraceId}
	if len(req.TxHash) != hash.LengthHash {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrTxHash)
		log.Error("tx hash format error", "trace_id", req.TraceId, "hash", req.TxHash)
		return rsp, nil
	}
	st := hub.GetStone(string(req.TxHash))
	if st == nil {
		rsp.ErrMessage = model.MakeErrMsgResponse(model.ErrUploadPayloadJobNotExist)
		log.Error("get stone error", "trace_id", req.TraceId, "hash", req.TxHash, "error", rsp.ErrMessage)
		return rsp, nil
	}
	uploadStone := st.(*stone.UploadPayloadStone)
	if req.ErrMessage.ErrCode != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Error("done primary job error", "trace_id", req.TraceId, "hash", req.TxHash, "error", req.ErrMessage)
		if err := uploadStone.InterruptStone(ctx, errors.New(req.ErrMessage.ErrMsg)); err != nil {
			rsp.ErrMessage = model.MakeErrMsgResponse(err)
			log.Error("interrupt stone error", "trace_id", req.TraceId, "hash", req.TxHash, "error", req.ErrMessage)
		}
		return rsp, nil
	}
	if err := uploadStone.ActionEvent(ctx, stone.UploadSecondaryPieceDoneEvent, req.PieceJob); err != nil {
		rsp.ErrMessage = model.MakeErrMsgResponse(err)
		log.Error("action stone fsm error", "trace_id", req.TraceId, "hash", req.TxHash, "error", req.ErrMessage)
	} else {
		log.Info("done primary piece job success", "trace_id", req.TraceId, "hash", req.TxHash)
	}
	return rsp, nil
}
