package gater

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield/types/s3util"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// putObjectHandler handles the upload object request.
func (g *GateModular) putObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		objectInfo    *storagetypes.ObjectInfo
		params        *storagetypes.Params
	)

	uploadPrimaryStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
		metrics.PerfUploadTimeHistogram.WithLabelValues("uploader_primary_total_time").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	if reqCtx.NeedVerifyAuthentication() {
		startAuthenticationTime := time.Now()
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
			coremodule.AuthOpTypePutObject, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		metrics.PerfUploadTimeHistogram.WithLabelValues("uploader_authorizer").Observe(time.Since(startAuthenticationTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "no permission to operate")
			err = ErrNoPermission
			return
		}
	}

	startGetObjectInfoTime := time.Now()
	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfUploadTimeHistogram.WithLabelValues("uploader_get_object_info").Observe(time.Since(startGetObjectInfoTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensus
		return
	}
	if objectInfo.GetPayloadSize() == 0 || objectInfo.GetPayloadSize() > g.maxPayloadSize {
		log.CtxErrorw(reqCtx.Context(), "failed to put object payload size is zero")
		err = ErrInvalidPayloadSize
		return
	}
	startGetStorageParamTime := time.Now()
	params, err = g.baseApp.Consensus().QueryStorageParamsByTimestamp(reqCtx.Context(), objectInfo.GetCreateAt())
	metrics.PerfUploadTimeHistogram.WithLabelValues("uploader_get_storage_param").Observe(time.Since(startGetStorageParamTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensus
		return
	}
	task := &gfsptask.GfSpUploadObjectTask{}
	task.InitUploadObjectTask(objectInfo, params, g.baseApp.TaskTimeout(task, objectInfo.GetPayloadSize()))
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, task.Key().String())
	g.baseApp.GfSpDB().InsertUploadEvent(task.GetObjectInfo().Id.Uint64(), corespdb.GatewayBeginReceiveUpload, task.Key().String())
	err = g.baseApp.GfSpClient().UploadObject(ctx, task, r.Body)
	if err != nil {
		g.baseApp.GfSpDB().InsertUploadEvent(task.GetObjectInfo().Id.Uint64(), corespdb.GatewayEndReceiveUpload, task.Key().String()+":"+err.Error())
		log.CtxErrorw(ctx, "failed to upload payload data", "error", err)
	}
	g.baseApp.GfSpDB().InsertUploadEvent(task.GetObjectInfo().Id.Uint64(), corespdb.GatewayEndReceiveUpload, task.Key().String())
	log.CtxDebugw(ctx, "succeed to upload payload data")
}

func parseRange(rangeStr string) (bool, int64, int64) {
	if rangeStr == "" {
		return false, -1, -1
	}
	rangeStr = strings.ToLower(rangeStr)
	rangeStr = strings.ReplaceAll(rangeStr, " ", "")
	if !strings.HasPrefix(rangeStr, "bytes=") {
		return false, -1, -1
	}
	rangeStr = rangeStr[len("bytes="):]
	if strings.HasSuffix(rangeStr, "-") {
		rangeStr = rangeStr[:len(rangeStr)-1]
		rangeStart, err := util.StringToUint64(rangeStr)
		if err != nil {
			return false, -1, -1
		}
		return true, int64(rangeStart), -1
	}
	pair := strings.Split(rangeStr, "-")
	if len(pair) == 2 {
		rangeStart, err := util.StringToUint64(pair[0])
		if err != nil {
			return false, -1, -1
		}
		rangeEnd, err := util.StringToUint64(pair[1])
		if err != nil {
			return false, -1, -1
		}
		return true, int64(rangeStart), int64(rangeEnd)
	}
	return false, -1, -1
}

// resumablePutObjectHandler handles the resumable put object
func (g *GateModular) resumablePutObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		objectInfo    *storagetypes.ObjectInfo
		params        *storagetypes.Params
	)

	uploadPrimaryStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
		metrics.PerfUploadTimeHistogram.WithLabelValues("uploader_primary_total_time").Observe(time.Since(uploadPrimaryStartTime).Seconds())

	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	if reqCtx.NeedVerifyAuthentication() {
		startAuthirzerTime := time.Now()
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
			coremodule.AuthOpTypePutObject, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		metrics.PerfUploadTimeHistogram.WithLabelValues("uploader_authorizer").Observe(time.Since(startAuthirzerTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "no permission to operate")
			err = ErrNoPermission
			return
		}
	}

	startGetObjectInfoTime := time.Now()
	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfUploadTimeHistogram.WithLabelValues("uploader_get_object_info").Observe(time.Since(startGetObjectInfoTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensus
		return
	}
	if objectInfo.GetPayloadSize() == 0 || objectInfo.GetPayloadSize() > g.maxPayloadSize {
		log.CtxErrorw(reqCtx.Context(), "failed to put object payload size is zero")
		err = ErrInvalidPayloadSize
		return
	}
	startGetStorageParamTime := time.Now()
	params, err = g.baseApp.Consensus().QueryStorageParams(reqCtx.Context())
	metrics.PerfUploadTimeHistogram.WithLabelValues("uploader_get_storage_param").Observe(time.Since(startGetStorageParamTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensus
		return
	}

	var (
		complete        bool
		offset          uint64
		requestComplete string
		requestOffset   string
	)
	queryParams := reqCtx.request.URL.Query()
	requestComplete = queryParams.Get(ResumableUploadComplete)
	requestOffset = queryParams.Get(ResumableUploadOffset)
	if requestComplete != "" {
		complete, err = util.StringToBool(requestComplete)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to parse complete from url", "error", err)
			err = ErrInvalidComplete
			return
		}
	} else {
		err = ErrInvalidComplete
		return
	}

	if requestOffset != "" {
		offset, err = util.StringToUint64(requestOffset)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to parse complete from url", "error", err)
			err = ErrInvalidOffset
			return
		}
	} else {
		err = ErrInvalidOffset
		return
	}

	task := &gfsptask.GfSpResumableUploadObjectTask{}
	task.InitResumableUploadObjectTask(objectInfo, params, g.baseApp.TaskTimeout(task, objectInfo.GetPayloadSize()), complete, offset)
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, task.Key().String())
	err = g.baseApp.GfSpClient().ResumableUploadObject(ctx, task, r.Body)
	if err != nil {
		log.CtxErrorw(ctx, "failed to upload payload data", "error", err)
	}
	log.CtxDebugw(ctx, "succeed to upload payload data")

}

// queryResumeOffsetHandler handles the resumable put object
func (g *GateModular) queryResumeOffsetHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		objectInfo    *storagetypes.ObjectInfo
		segmentCount  uint32
		offset        uint64
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	if reqCtx.NeedVerifyAuthentication() {
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
			coremodule.AuthOpTypeGetUploadingState, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "no permission to operate")
			err = ErrNoPermission
			return
		}
	}

	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensus
		return
	}

	params, err := g.baseApp.Consensus().QueryStorageParamsByTimestamp(
		reqCtx.Context(), objectInfo.GetCreateAt())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensus
		return
	}

	segmentCount, err = g.baseApp.GfSpClient().GetUploadObjectSegment(reqCtx.Context(), objectInfo.Id.Uint64())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get uploading job state", "error", err)
		return
	}

	offset = uint64(segmentCount) * params.GetMaxSegmentSize()

	var xmlInfo = struct {
		XMLName xml.Name `xml:"QueryResumeOffset"`
		Version string   `xml:"version,attr"`
		Offset  uint64   `xml:"Offset"`
	}{
		Version: GnfdResponseXMLVersion,
		Offset:  offset,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponse
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponse
		return
	}
	log.Debugw("succeed to query resumable offset ", "xml_info", xmlInfo)
}

// getObjectHandler handles the download object request.
func (g *GateModular) getObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtxErr     error
		reqCtx        *RequestContext
		authenticated bool
		objectInfo    *storagetypes.ObjectInfo
		bucketInfo    *storagetypes.BucketInfo
		params        *storagetypes.Params
		lowOffset     int64
		highOffset    int64
		pieceInfos    []*downloader.SegmentPieceInfo
	)
	getObjectStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
		metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_total_time").Observe(time.Since(getObjectStartTime).Seconds())
	}()
	reqCtx, reqCtxErr = NewRequestContext(r, g)
	// check the object permission whether allow public read.
	verifyObjectPermissionTime := time.Now()
	if authenticated, err = g.baseApp.Consensus().VerifyGetObjectPermission(reqCtx.Context(), sdk.AccAddress{}.String(),
		reqCtx.bucketName, reqCtx.objectName); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authentication for getting public object", "error", err)
		err = ErrConsensus
		return
	}
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_verify_object_permission_time").Observe(time.Since(verifyObjectPermissionTime).Seconds())

	if !authenticated {
		if reqCtxErr != nil {
			err = reqCtxErr
			log.CtxErrorw(reqCtx.Context(), "no permission to operate, object is not public", "error", err)
			return
		}
		if reqCtx.NeedVerifyAuthentication() {
			authTime := time.Now()
			if authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
				coremodule.AuthOpTypeGetObject, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName); err != nil {
				metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_auth_time").Observe(time.Since(authTime).Seconds())
				log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
				return
			}
			metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_auth_time").Observe(time.Since(authTime).Seconds())
			if !authenticated {
				log.CtxErrorw(reqCtx.Context(), "no permission to operate")
				err = ErrNoPermission
				return
			}
		}
	} // else anonymous users can get public object.

	getObjectTime := time.Now()
	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_get_object_info_time").Observe(time.Since(getObjectTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensus
		return
	}

	getBucketTime := time.Now()
	bucketInfo, err = g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_get_bucket_info_time").Observe(time.Since(getBucketTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensus
		return
	}

	getParamTime := time.Now()
	params, err = g.baseApp.Consensus().QueryStorageParamsByTimestamp(reqCtx.Context(), objectInfo.GetCreateAt())
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_get_storage_param_time").Observe(time.Since(getParamTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensus
		return
	}

	isRange, rangeStart, rangeEnd := parseRange(reqCtx.request.Header.Get(RangeHeader))
	if isRange && (rangeEnd < 0 || rangeEnd >= int64(objectInfo.GetPayloadSize())) {
		rangeEnd = int64(objectInfo.GetPayloadSize()) - 1
	}
	if isRange && (rangeStart < 0 || rangeEnd < 0 || rangeStart > rangeEnd) {
		err = ErrInvalidRange
		return
	}

	if isRange {
		lowOffset = rangeStart
		highOffset = rangeEnd
	} else {
		lowOffset = 0
		highOffset = int64(objectInfo.GetPayloadSize()) - 1
	}

	task := &gfsptask.GfSpDownloadObjectTask{}
	task.InitDownloadObjectTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(task), reqCtx.Account(),
		lowOffset, highOffset, g.baseApp.TaskTimeout(task, uint64(highOffset-lowOffset+1)), g.baseApp.TaskMaxRetry(task))
	if pieceInfos, err = downloader.SplitToSegmentPieceInfos(task, g.baseApp.PieceOp()); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to download object", "error", err)
		return
	}
	w.Header().Set(ContentTypeHeader, objectInfo.GetContentType())
	if isRange {
		w.Header().Set(ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(lowOffset))+
			"-"+util.Uint64ToString(uint64(highOffset)))
	} else {
		w.Header().Set(ContentLengthHeader, util.Uint64ToString(objectInfo.GetPayloadSize()))
	}

	getDataTime := time.Now()
	for idx, pInfo := range pieceInfos {
		enableCheck := false
		if idx == 0 { // only check in first piece
			enableCheck = true
		}
		pieceTask := &gfsptask.GfSpDownloadPieceTask{}
		pieceTask.InitDownloadPieceTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(task),
			enableCheck, reqCtx.Account(), uint64(highOffset-lowOffset+1), pInfo.SegmentPieceKey, pInfo.Offset,
			pInfo.Length, g.baseApp.TaskTimeout(task, uint64(pieceTask.GetSize())), g.baseApp.TaskMaxRetry(task))
		getSegmentTime := time.Now()
		pieceData, err := g.baseApp.GfSpClient().GetPiece(reqCtx.Context(), pieceTask)
		metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_segment_data_time").Observe(time.Since(getSegmentTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to download piecem", "error", err)
			// TODO pieceStore should return exact error to indicate if the piece data lost
			// for now, if get piece fail, it is suspected that the piece has been lost
			// the recovery task will recovery the total segment (ignoring the offset and length of piece info)
			segmentIndex, parseErr := g.baseApp.PieceOp().ParseSegmentIdx(pInfo.SegmentPieceKey)
			if parseErr != nil {
				// no need to return recovery error to user
				log.CtxErrorw(reqCtx.Context(), "fail to parse recovery segment index", "error", err)
				return
			}
			segSize := g.baseApp.PieceOp().SegmentPieceSize(objectInfo.PayloadSize, segmentIndex, params.GetMaxSegmentSize())
			recoveryTask := &gfsptask.GfSpRecoveryPieceTask{}
			recoveryTask.InitRecoveryPieceTask(task.GetObjectInfo(), task.GetStorageParams(),
				g.baseApp.TaskPriority(recoveryTask),
				segmentIndex,
				int32(-1),
				uint64(segSize),
				g.baseApp.TaskTimeout(recoveryTask, task.GetStorageParams().GetMaxSegmentSize()),
				g.baseApp.TaskMaxRetry(recoveryTask))

			taskErr := g.baseApp.GfSpClient().ReportTask(reqCtx.Context(), recoveryTask)
			if taskErr != nil {
				// no need to return recovery error to user
				log.CtxErrorw(reqCtx.Context(), "fail to generate recovery task", "error", err)
				return
			}
			log.CtxDebugw(reqCtx.Context(), "recovery task run successfully", "recovery object", objectInfo.ObjectName, "segment index:", idx, "error", err)

			pieceData, err = g.tryDownloadAfterRecovery(reqCtx.Context(), pieceTask)
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "fail to get piece after recovery task submitted", "error", err)
				return
			}
		}

		writeTime := time.Now()
		w.Write(pieceData)
		metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_write_time").Observe(time.Since(writeTime).Seconds())
	}
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_get_data_time").Observe(time.Since(getDataTime).Seconds())
}

// tryDownloadAfterRecovery try to get piece data after data recoverying
func (g *GateModular) tryDownloadAfterRecovery(ctx context.Context, pieceTask *gfsptask.GfSpDownloadPieceTask) ([]byte, error) {
	timeout := time.After(RecoveryTimeOutSeconds * time.Second)
	ticker := time.NewTicker(RecoveryCheckInterval * time.Second)
	defer ticker.Stop()

	// sleep 500ms at least to waiting for recovering finish
	time.Sleep(RecoveryMinMilliseconds * time.Millisecond)
	startTime := time.Now()
	for {
		select {
		case <-timeout:
			// if recovering has not finished after time out , return err to client
			return nil, ErrRecoveryTimeout
		case <-ticker.C:
			pieceData, err := g.baseApp.GfSpClient().GetPiece(ctx, pieceTask)

			if err == nil {
				log.CtxDebugw(ctx, "get piece successfully after recovering", "recovery object", pieceTask.GetPieceKey(),
					"cost time:", time.Since(startTime))
				return pieceData, nil
			}
		}
	}

}

// queryUploadProgressHandler handles the query uploaded object progress request.
func (g *GateModular) queryUploadProgressHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		objectInfo    *storagetypes.ObjectInfo
		taskState     int32
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	if reqCtx.NeedVerifyAuthentication() {
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
			coremodule.AuthOpTypeGetUploadingState, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "no permission to operate")
			err = ErrNoPermission
			return
		}
	}

	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensus
		return
	}

	taskState, err = g.baseApp.GfSpClient().GetUploadObjectState(reqCtx.Context(), objectInfo.Id.Uint64())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get uploading job state", "error", err)
		return
	}
	taskStateDescription := servicetypes.StateToDescription(servicetypes.TaskState(taskState))

	var xmlInfo = struct {
		XMLName             xml.Name `xml:"QueryUploadProgress"`
		Version             string   `xml:"version,attr"`
		ProgressDescription string   `xml:"ProgressDescription"`
	}{
		Version:             GnfdResponseXMLVersion,
		ProgressDescription: taskStateDescription,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponse
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponse
		return
	}
	log.Debugw("succeed to query upload progress", "xml_info", xmlInfo)
}

// getObjectByUniversalEndpointHandler handles the get object request sent by universal endpoint
func (g *GateModular) getObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request, isDownload bool) {
	var (
		err                  error
		reqCtx               *RequestContext
		authenticated        bool
		isRange              bool
		rangeStart           int64
		rangeEnd             int64
		redirectURL          string
		params               *storagetypes.Params
		escapedObjectName    string
		isRequestFromBrowser bool
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			if isRequestFromBrowser {
				reqCtx.SetHttpCode(http.StatusOK)
				errorCodeForPage := "INTERNAL_ERROR" // default errorCode in built-in error page
				switch err {
				case downloader.ErrExceedBucketQuota:
					errorCodeForPage = "NO_ENOUGH_QUOTA"
				case ErrNoSuchObject:
					errorCodeForPage = "FILE_NOT_FOUND"
				case ErrForbidden:
					errorCodeForPage = "NO_PERMISSION"
				}
				html := strings.Replace(GnfdBuiltInUniversalEndpointDappErrorPage, "<% errorCode %>", errorCodeForPage, 1)

				fmt.Fprintf(w, "%s", html)
				return
			} else {
				reqCtx.SetError(gfsperrors.MakeGfSpError(err))
				reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
				MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			}

		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	userAgent := r.Header.Get("User-Agent")
	isRequestFromBrowser = checkIfRequestFromBrowser(userAgent)

	// ignore the error, because the universal endpoint does not need signature
	reqCtx, _ = NewRequestContext(r, g)

	escapedObjectName, err = url.PathUnescape(reqCtx.objectName)
	if err != nil {
		log.Errorw("failed to unescape object name ", "object_name", reqCtx.objectName, "error", err)
		return
	}

	if err = s3util.CheckValidBucketName(reqCtx.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqCtx.bucketName, "error", err)
		return
	}
	if err = s3util.CheckValidObjectName(escapedObjectName); err != nil {
		log.Errorw("failed to check object name", "object_name", escapedObjectName, "error", err)
		return
	}

	getBucketInfoRes, getBucketInfoErr := g.baseApp.GfSpClient().GetBucketByBucketName(reqCtx.Context(), reqCtx.bucketName, true)
	if getBucketInfoErr != nil || getBucketInfoRes == nil || getBucketInfoRes.GetBucketInfo() == nil {
		log.Errorw("failed to check bucket info", "bucket_name", reqCtx.bucketName, "error", getBucketInfoErr)
		err = ErrNoSuchObject
		return
	}

	bucketPrimarySpAddress := getBucketInfoRes.GetBucketInfo().GetPrimarySpAddress()
	// if bucket not in the current sp, 302 redirect to the sp that contains the bucket
	if !strings.EqualFold(bucketPrimarySpAddress, g.baseApp.OperatorAddress()) {
		log.Debugw("primary sp address not matched ",
			"bucketPrimarySpAddress", bucketPrimarySpAddress, "gateway.config.SpOperatorAddress", g.baseApp.OperatorAddress(),
		)

		spEndpoint, getEndpointErr := g.baseApp.GfSpClient().GetEndpointBySpAddress(reqCtx.Context(), bucketPrimarySpAddress)

		if getEndpointErr != nil || spEndpoint == "" {
			log.Errorw("failed to get endpoint by address ", "sp_address", reqCtx.bucketName, "error", getEndpointErr)
			err = getEndpointErr
			return
		}

		redirectURL = spEndpoint + r.RequestURI
		log.Debugw("getting redirect url:", "redirectURL", redirectURL)

		http.Redirect(w, r, redirectURL, 302)
		return
	}

	getObjectInfoRes, err := g.baseApp.GfSpClient().GetObjectMeta(reqCtx.Context(), escapedObjectName, reqCtx.bucketName, true)
	if err != nil || getObjectInfoRes == nil || getObjectInfoRes.GetObjectInfo() == nil {
		log.Errorw("failed to check object meta", "object_name", escapedObjectName, "error", err)
		err = ErrNoSuchObject
		return
	}

	if getObjectInfoRes.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.Errorw("object is not sealed",
			"status", getObjectInfoRes.GetObjectInfo().GetObjectStatus())
		err = ErrNoSuchObject
		return
	}

	if isPrivateObject(getBucketInfoRes.GetBucketInfo(), getObjectInfoRes.GetObjectInfo()) {
		// for private files, we return a built-in dapp and help users provide a signature for verification

		var (
			expiry    string
			signature string
		)
		queryParams := r.URL.Query()
		if queryParams["expiry"] != nil {
			expiry = queryParams["expiry"][0]
		}
		if queryParams["signature"] != nil {
			signature = queryParams["signature"][0]
		}
		if expiry != "" && signature != "" {
			// check if expiry set to far or expiry is past
			expiryDate, dateParseErr := time.Parse(ExpiryDateFormat, expiry)
			if dateParseErr != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to parse expiry date due to invalid format", "expiry", expiry)
				err = ErrInvalidExpiryDate
				return
			}
			log.Infof("%s", time.Until(expiryDate).Seconds())
			log.Infof("%s", MaxExpiryAgeInSec)
			expiryAge := int32(time.Until(expiryDate).Seconds())
			if MaxExpiryAgeInSec < expiryAge || expiryAge < 0 {
				err = ErrInvalidExpiryDate
				log.CtxErrorw(reqCtx.Context(), "failed to parse expiry date due to invalid expiry value", "expiry", expiry)
				return
			}

			// check permission

			// 1. solve the account
			signedMsg := fmt.Sprintf(GnfdBuiltInDappSignedContentTemplate, "gnfd://"+getBucketInfoRes.GetBucketInfo().BucketName+"/"+getObjectInfoRes.GetObjectInfo().GetObjectName(), expiry)
			accAddress, verifySigErr := VerifyPersonalSignature(signedMsg, signature)
			if verifySigErr != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify signature", "error", verifySigErr)
				err = verifySigErr
				return
			}
			reqCtx.account = accAddress.String()

			// 2. check permission
			authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
				coremodule.AuthOpTypeGetObject, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
				err = ErrForbidden
				return
			}
			if !authenticated {
				log.CtxErrorw(reqCtx.Context(), "no permission to operate")
				err = ErrForbidden
				return
			}

		} else {
			if !isRequestFromBrowser {
				err = ErrForbidden
				return
			}
			// if the request comes from browser, we will return a built-in dapp for users to make the signature
			var htmlConfigMap = map[string]string{
				"greenfield_7971-1":    "{\n  \"envType\": \"dev\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 7971,\n  \"chainName\": \"dev - greenfield\",\n  \"rpcUrls\": [\"https://gnfd-dev.qa.bnbchain.world\"],\n  \"nativeCurrency\": { \"name\": \"BNB\", \"symbol\": \"BNB\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://greenfieldscan-qanet.fe.nodereal.cc/\"]\n}\n",
				"greenfield_9000-1741": "{\n  \"envType\": \"qa\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 9000,\n  \"chainName\": \"qa - greenfield\",\n  \"rpcUrls\": [\"https://gnfd.qa.bnbchain.world\"],\n  \"nativeCurrency\": { \"name\": \"BNB\", \"symbol\": \"BNB\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://greenfieldscan-qanet.fe.nodereal.cc/\"]\n}\n",
				"greenfield_5600-1":    "{\n  \"envType\": \"testnet\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 5600,\n  \"chainName\": \"greenfield testnet\",\n  \"rpcUrls\": [\"https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org\"],\n  \"nativeCurrency\": { \"name\": \"BNB\", \"symbol\": \"BNB\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://greenfieldscan.com/\"]\n}\n",
			}

			htmlConfig := htmlConfigMap[g.baseApp.ChainID()]
			if htmlConfig == "" {
				log.CtxErrorw(reqCtx.Context(), "chain id is not found", "chain id ", g.baseApp.ChainID())
				err = gfsperrors.MakeGfSpError(fmt.Errorf("chain id is not found"))
				return
			}
			hc, _ := json.Marshal(htmlConfig)
			html := strings.Replace(GnfdBuiltInUniversalEndpointDappHtml, "<% env %>", string(hc), 1)

			fmt.Fprintf(w, "%s", html)
			return
		}

	}

	params, err = g.baseApp.Consensus().QueryStorageParamsByTimestamp(
		reqCtx.Context(), getObjectInfoRes.GetObjectInfo().GetCreateAt())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensus
		return
	}

	var low int64
	var high int64
	if isRange {
		low = rangeStart
		high = rangeEnd
	} else {
		low = 0
		high = int64(getObjectInfoRes.GetObjectInfo().GetPayloadSize()) - 1
	}

	task := &gfsptask.GfSpDownloadObjectTask{}
	task.InitDownloadObjectTask(getObjectInfoRes.GetObjectInfo(), getBucketInfoRes.GetBucketInfo(), params, g.baseApp.TaskPriority(task), reqCtx.Account(),
		low, high, g.baseApp.TaskTimeout(task, uint64(high-low+1)), g.baseApp.TaskMaxRetry(task))
	data, getObjectErr := g.baseApp.GfSpClient().GetObject(reqCtx.Context(), task)
	if getObjectErr != nil {
		err = getObjectErr
		log.CtxErrorw(reqCtx.Context(), "failed to download object", "error", err)
		return
	}

	if isDownload {
		w.Header().Set(ContentDispositionHeader, ContentDispositionAttachmentValue+"; filename=\""+escapedObjectName+"\"")
	} else {
		w.Header().Set(ContentDispositionHeader, ContentDispositionInlineValue)
	}
	w.Header().Set(ContentTypeHeader, getObjectInfoRes.GetObjectInfo().GetContentType())
	if isRange {
		w.Header().Set(ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(low))+
			"-"+util.Uint64ToString(uint64(high)))
	} else {
		w.Header().Set(ContentLengthHeader, util.Uint64ToString(getObjectInfoRes.GetObjectInfo().GetPayloadSize()))
	}
	w.Write(data)
	log.CtxDebugw(reqCtx.Context(), "succeed to download object for universal endpoint")
}

func isPrivateObject(bucket *storagetypes.BucketInfo, object *storagetypes.ObjectInfo) bool {
	return object.GetVisibility() == storagetypes.VISIBILITY_TYPE_PRIVATE ||
		(object.GetVisibility() == storagetypes.VISIBILITY_TYPE_INHERIT &&
			bucket.GetVisibility() == storagetypes.VISIBILITY_TYPE_PRIVATE)
}

func checkIfRequestFromBrowser(userAgent string) bool {
	// List of common user agent substrings for mainstream browsers
	mainstreamBrowsers := []string{"Chrome", "Firefox", "Safari", "Opera", "Edge"}
	// Check if the User-Agent header contains any of the mainstream browser substrings
	for _, browser := range mainstreamBrowsers {
		if strings.Contains(userAgent, browser) {
			return true
		}
	}
	return false
}

// downloadObjectByUniversalEndpointHandler handles the download object request sent by universal endpoint
func (g *GateModular) downloadObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
	g.getObjectByUniversalEndpointHandler(w, r, true)
}

// viewObjectByUniversalEndpointHandler handles the view object request sent by universal endpoint
func (g *GateModular) viewObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
	g.getObjectByUniversalEndpointHandler(w, r, false)
}
