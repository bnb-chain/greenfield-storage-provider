package gater

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"

	"github.com/bnb-chain/greenfield/types/s3util"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/job"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// putObjectHandler handles the upload object request.
func (g *GateModular) putObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		reqCtx     *RequestContext
		authorized bool
		objectInfo *storagetypes.ObjectInfo
		params     *storagetypes.Params
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

	reqCtx, err = g.NewRequestContext(r)
	if err != nil {
		return
	}
	if reqCtx.NeedVerifyAuthorizer() {
		authorized, err = g.baseApp.GfSpClient().VerifyAuthorize(reqCtx.Context(),
			coremodule.AuthOpTypePutObject, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
			return
		}
		if !authorized {
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
	if objectInfo.GetPayloadSize() == 0 || objectInfo.GetPayloadSize() > g.maxPayloadSize {
		log.CtxErrorw(reqCtx.Context(), "failed to put object payload size is zero")
		err = ErrInvalidPayloadSize
		return
	}
	params, err = g.baseApp.Consensus().QueryStorageParams(reqCtx.Context())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensus
		return
	}
	task := &gfsptask.GfSpUploadObjectTask{}
	task.InitUploadObjectTask(objectInfo, params, g.baseApp.TaskTimeout(task, objectInfo.GetPayloadSize()))
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, task.Key().String())
	err = g.baseApp.GfSpClient().UploadObject(ctx, task, r.Body)
	if err != nil {
		log.CtxErrorw(ctx, "failed to upload payload data", "error", err)
	}
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

// getObjectHandler handles the download object request.
func (g *GateModular) getObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		reqCtxErr  error
		reqCtx     *RequestContext
		authorized bool
		objectInfo *storagetypes.ObjectInfo
		bucketInfo *storagetypes.BucketInfo
		params     *storagetypes.Params
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
	reqCtx, reqCtxErr = g.NewRequestContext(r)
	// check the object's visibility type whether equal to public read
	authorized, err = g.baseApp.Consensus().VerifyGetObjectPermission(reqCtx.Context(), "",
		reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authorize for getting public object", "error", err)
		err = ErrConsensus
		return
	}
	if !authorized {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate, object is not public")
	} else {
		reqCtxErr = nil
	}
	if reqCtxErr != nil {
		err = reqCtxErr
		return
	}

	if !authorized && reqCtx.NeedVerifyAuthorizer() {
		authorized, err = g.baseApp.GfSpClient().VerifyAuthorize(reqCtx.Context(),
			coremodule.AuthOpTypeGetObject, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
			return
		}
		if !authorized {
			log.CtxErrorw(reqCtx.Context(), "no permission to operate")
			return
		}
	}

	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensus
		return
	}
	bucketInfo, err = g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensus
		return
	}
	params, err = g.baseApp.Consensus().QueryStorageParams(reqCtx.Context())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensus
		return
	}

	isRange, rangeStart, rangeEnd := parseRange(reqCtx.request.Header.Get(model.RangeHeader))
	if isRange && (rangeEnd < 0 || rangeEnd >= int64(objectInfo.GetPayloadSize())) {
		rangeEnd = int64(objectInfo.GetPayloadSize()) - 1
	}
	if isRange && (rangeStart < 0 || rangeEnd < 0 || rangeStart > rangeEnd) {
		err = ErrInvalidRange
		return
	}
	var low int64
	var high int64
	if isRange {
		low = rangeStart
		high = rangeEnd
	} else {
		low = 0
		high = int64(objectInfo.GetPayloadSize()) - 1
	}

	task := &gfsptask.GfSpDownloadObjectTask{}
	task.InitDownloadObjectTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(task), reqCtx.Account(),
		low, high, g.baseApp.TaskTimeout(task, uint64(high-low+1)), g.baseApp.TaskMaxRetry(task))
	data, err := g.baseApp.GfSpClient().GetObject(reqCtx.Context(), task)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to download object", "error", err)
		return
	}
	w.Header().Set(model.ContentTypeHeader, objectInfo.GetContentType())
	if isRange {
		w.Header().Set(model.ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(low))+
			"-"+util.Uint64ToString(uint64(high)))
	} else {
		w.Header().Set(model.ContentLengthHeader, util.Uint64ToString(objectInfo.GetPayloadSize()))
	}
	w.Write(data)
	log.CtxDebugw(reqCtx.Context(), "succeed to download object")
}

// queryUploadProgressHandler handles the query uploaded object progress request.
func (g *GateModular) queryUploadProgressHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		reqCtx     *RequestContext
		authorized bool
		objectInfo *storagetypes.ObjectInfo
		jobState   int32
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

	reqCtx, err = g.NewRequestContext(r)
	if err != nil {
		return
	}
	if reqCtx.NeedVerifyAuthorizer() {
		authorized, err = g.baseApp.GfSpClient().VerifyAuthorize(reqCtx.Context(),
			coremodule.AuthOpTypeGetUploadingState, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
			return
		}
		if !authorized {
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

	jobState, err = g.baseApp.GfSpClient().GetUploadObjectState(reqCtx.Context(), objectInfo.Id.Uint64())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get uploading job state", "error", err)
		return
	}
	jobStateDescription := job.StateToDescription(servicetypes.JobState(jobState))

	var xmlInfo = struct {
		XMLName             xml.Name `xml:"QueryUploadProgress"`
		Version             string   `xml:"version,attr"`
		ProgressDescription string   `xml:"ProgressDescription"`
	}{
		Version:             model.GnfdResponseXMLVersion,
		ProgressDescription: jobStateDescription,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponse
		return
	}
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeXMLHeaderValue)
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
		err         error
		reqCtx      *RequestContext
		isRange     bool
		rangeStart  int64
		rangeEnd    int64
		redirectUrl string
		params      *storagetypes.Params
	)

	// ignore the error, because the universal endpoint does not need signature
	reqCtx, _ = g.NewRequestContext(r)

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

	escapedObjectName, err := url.PathUnescape(reqCtx.objectName)
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

	getBucketInfoRes, err := g.baseApp.GfSpClient().GetBucketByBucketName(reqCtx.Context(), reqCtx.bucketName, true)
	if err != nil || getBucketInfoRes == nil || getBucketInfoRes.GetBucketInfo() == nil {
		log.Errorw("failed to check bucket info", "bucket_name", reqCtx.bucketName, "error", err)
		return
	}

	bucketPrimarySpAddress := getBucketInfoRes.GetBucketInfo().GetPrimarySpAddress()

	// if bucket not in the current sp, 302 redirect to the sp that contains the bucket
	if bucketPrimarySpAddress != g.baseApp.OperateAddress() {
		log.Debugw("primary sp address not matched ",
			"bucketPrimarySpAddress", bucketPrimarySpAddress, "gateway.config.SpOperatorAddress", g.baseApp.OperateAddress(),
		)

		endpoint, err := g.baseApp.GfSpClient().GetEndpointBySpAddress(reqCtx.Context(), bucketPrimarySpAddress)

		if err != nil || endpoint == "" {
			log.Errorw("failed to get endpoint by address ", "sp_address", reqCtx.bucketName, "error", err)
			return
		}

		if isDownload {
			redirectUrl = endpoint + "/download/" + reqCtx.bucketName + "/" + reqCtx.objectName
		} else {
			redirectUrl = endpoint + "/view/" + reqCtx.bucketName + "/" + reqCtx.objectName
		}

		log.Debugw("getting redirect url:", "redirectUrl", redirectUrl)

		http.Redirect(w, r, redirectUrl, 302)
		return
	}

	// In first phase, do not provide universal endpoint for private object
	getObjectInfoRes, err := g.baseApp.GfSpClient().GetObjectMeta(reqCtx.Context(), escapedObjectName, reqCtx.bucketName, false)
	if err != nil || getObjectInfoRes == nil || getObjectInfoRes.GetObjectInfo() == nil {
		log.Errorw("failed to check object meta", "object_name", escapedObjectName, "error", err)
		return
	}

	if getObjectInfoRes.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.Errorw("object is not sealed",
			"status", getObjectInfoRes.GetObjectInfo().GetObjectStatus())
		return
	}

	params, err = g.baseApp.Consensus().QueryStorageParams(reqCtx.Context())
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
	data, err := g.baseApp.GfSpClient().GetObject(reqCtx.Context(), task)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to download object", "error", err)
		return
	}

	if isDownload {
		w.Header().Set(model.ContentDispositionHeader, model.ContentDispositionAttachmentValue+"; filename=\""+escapedObjectName+"\"")
	} else {
		w.Header().Set(model.ContentDispositionHeader, model.ContentDispositionInlineValue)
	}
	w.Header().Set(model.ContentTypeHeader, getObjectInfoRes.GetObjectInfo().GetContentType())
	if isRange {
		w.Header().Set(model.ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(low))+
			"-"+util.Uint64ToString(uint64(high)))
	} else {
		w.Header().Set(model.ContentLengthHeader, util.Uint64ToString(getObjectInfoRes.GetObjectInfo().GetPayloadSize()))
	}
	w.Write(data)
	log.CtxDebugw(reqCtx.Context(), "succeed to download object for universal endpoint")
}

// downloadObjectByUniversalEndpointHandler handles the download object request sent by universal endpoint
func (g *GateModular) downloadObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
	g.getObjectByUniversalEndpointHandler(w, r, true)
}

// viewObjectByUniversalEndpointHandler handles the view object request sent by universal endpoint
func (g *GateModular) viewObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
	g.getObjectByUniversalEndpointHandler(w, r, false)
}
