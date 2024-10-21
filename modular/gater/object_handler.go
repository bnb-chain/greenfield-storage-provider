package gater

import (
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	commonhash "github.com/zkMeLabs/mechain-common/go/hash"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v12/types/s3util"
	permissiontypes "github.com/evmos/evmos/v12/x/permission/types"
	storagetypes "github.com/evmos/evmos/v12/x/storage/types"
	commonhttp "github.com/zkMeLabs/mechain-common/go/http"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsperrors"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsptask"
	coremodule "github.com/zkMeLabs/mechain-storage-provider/core/module"
	modelgateway "github.com/zkMeLabs/mechain-storage-provider/model/gateway"
	"github.com/zkMeLabs/mechain-storage-provider/modular/downloader"
	"github.com/zkMeLabs/mechain-storage-provider/modular/metadata"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/metrics"
	"github.com/zkMeLabs/mechain-storage-provider/store/sqldb"
	servicetypes "github.com/zkMeLabs/mechain-storage-provider/store/types"
	"github.com/zkMeLabs/mechain-storage-provider/util"
)

const ContentDefault = "application/octet-stream"

// putObjectHandler handles the upload object request.
func (g *GateModular) putObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		bucketInfo    *storagetypes.BucketInfo
		objectInfo    *storagetypes.ObjectInfo
		params        *storagetypes.Params
	)

	uploadPrimaryStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewayFailurePutObject).Inc()
			metrics.ReqTime.WithLabelValues(GatewayFailurePutObject).Observe(time.Since(uploadPrimaryStartTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqPieceSize.WithLabelValues(GatewayPutObjectSize).Observe(float64(objectInfo.GetPayloadSize()))
			metrics.ReqTime.WithLabelValues(GatewaySuccessPutObject).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewaySuccessPutObject).Inc()
			metrics.ReqPieceSize.WithLabelValues(GatewaySuccessPutObject).Observe(float64(objectInfo.GetPayloadSize()))
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}

	if err = g.checkSPAndBucketStatus(reqCtx.Context(), reqCtx.bucketName, reqCtx.account); err != nil {
		log.CtxErrorw(reqCtx.Context(), "put object failed to check sp and bucket status", "error", err)
		return
	}
	startAuthenticationTime := time.Now()
	authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(), coremodule.AuthOpTypePutObject,
		reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_put_object_authorizer").Observe(time.Since(startAuthenticationTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
		return
	}
	if !authenticated {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate")
		err = ErrNoPermission
		return
	}

	startGetObjectInfoTime := time.Now()
	bucketInfo, objectInfo, err = g.baseApp.Consensus().QueryBucketInfoAndObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_put_object_query_object_cost").Observe(time.Since(startGetObjectInfoTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_put_object_query_object_end").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}
	err = g.checkAndAssignShadowObjectInfo(reqCtx, objectInfo)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ",  error: " + err.Error())
		return
	}
	if objectInfo.GetPayloadSize() == 0 || objectInfo.GetPayloadSize() > g.maxPayloadSize {
		log.CtxErrorw(reqCtx.Context(), "failed to put object payload size is zero")
		err = ErrInvalidPayloadSize
		return
	}
	startGetStorageParamTime := time.Now()
	params, err = g.baseApp.Consensus().QueryStorageParamsByTimestamp(reqCtx.Context(), objectInfo.GetLatestUpdatedTime())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_put_object_query_params_cost").Observe(time.Since(startGetStorageParamTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_put_object_query_params_end").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get storage params from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}
	task := &gfsptask.GfSpUploadObjectTask{}
	task.InitUploadObjectTask(bucketInfo.GetGlobalVirtualGroupFamilyId(), objectInfo, params, g.baseApp.TaskTimeout(task, objectInfo.GetPayloadSize()), false)
	task.SetCreateTime(uploadPrimaryStartTime.Unix())
	task.AppendLog(fmt.Sprintf("gateway-prepare-upload-task-cost:%d", time.Now().UnixMilli()-uploadPrimaryStartTime.UnixMilli()))
	task.AppendLog("gateway-create-upload-task")
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, task.Key().String())
	uploadDataTime := time.Now()
	err = g.baseApp.GfSpClient().UploadObject(ctx, task, r.Body)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_put_object_data_cost").Observe(time.Since(uploadDataTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_put_object_data_end").Observe(time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to upload payload data", "error", err)
		return
	}
	log.CtxDebug(ctx, "succeed to upload payload data")
}

func (g *GateModular) checkAndAssignShadowObjectInfo(reqCtx *RequestContext, objectInfo *storagetypes.ObjectInfo) error {
	if !objectInfo.IsUpdating {
		return nil
	}
	shadowObject, err := g.baseApp.Consensus().QueryShadowObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		return err
	}
	// the shadowObjectInfo will be injected into the objectInfo and passed to related Tasks.
	// e.g. UploadObjectTask, ReceivePieceTask, SealObjetTask
	objectInfo.PayloadSize = shadowObject.PayloadSize
	objectInfo.Version = shadowObject.Version
	objectInfo.Checksums = shadowObject.Checksums
	objectInfo.UpdatedAt = shadowObject.UpdatedAt
	return nil
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
		rangeStart, err := util.StringToInt64(rangeStr)
		if err != nil {
			return false, -1, -1
		}
		return true, rangeStart, -1
	}
	pair := strings.Split(rangeStr, "-")
	if len(pair) == 2 {
		rangeStart, err := util.StringToInt64(pair[0])
		if err != nil {
			return false, -1, -1
		}
		rangeEnd, err := util.StringToInt64(pair[1])
		if err != nil {
			return false, -1, -1
		}
		return true, rangeStart, rangeEnd
	}
	return false, -1, -1
}

// resumablePutObjectHandler handles the resumable put object
func (g *GateModular) resumablePutObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		bucketInfo    *storagetypes.BucketInfo
		objectInfo    *storagetypes.ObjectInfo
		params        *storagetypes.Params
	)

	uploadPrimaryStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(uploadPrimaryStartTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(uploadPrimaryStartTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}

	if err = g.checkSPAndBucketStatus(reqCtx.Context(), reqCtx.bucketName, reqCtx.account); err != nil {
		log.CtxErrorw(reqCtx.Context(), "resumable put object failed to check sp and bucket status", "error", err)
		return
	}
	authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
		coremodule.AuthOpTypePutObject, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
		return
	}
	if !authenticated {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate")
		err = ErrNoPermission
		return
	}

	startGetObjectInfoTime := time.Now()
	bucketInfo, objectInfo, err = g.baseApp.Consensus().QueryBucketInfoAndObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_query_object_cost").Observe(time.Since(startGetObjectInfoTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_query_object_end").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ",error: " + err.Error())
		return
	}

	startGetStorageParamTime := time.Now()
	params, err = g.baseApp.Consensus().QueryStorageParams(reqCtx.Context())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_query_params_cost").Observe(time.Since(startGetStorageParamTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_query_params_end").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get storage params from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}
	err = g.checkAndAssignShadowObjectInfo(reqCtx, objectInfo)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ",error: " + err.Error())
		return
	}
	// the resumable upload utilizes the on-chain MaxPayloadSize as the maximum file size
	if objectInfo.GetPayloadSize() == 0 || objectInfo.GetPayloadSize() > params.GetMaxPayloadSize() {
		log.CtxErrorw(reqCtx.Context(), "failed to put object payload size is zero")
		err = ErrInvalidPayloadSize
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
			log.CtxErrorw(reqCtx.Context(), "failed to parse offset from url", "error", err)
			err = ErrInvalidOffset
			return
		}
	} else {
		err = ErrInvalidOffset
		return
	}

	task := &gfsptask.GfSpResumableUploadObjectTask{}
	task.InitResumableUploadObjectTask(bucketInfo.GetGlobalVirtualGroupFamilyId(), objectInfo, params, g.baseApp.TaskTimeout(task, objectInfo.GetPayloadSize()), complete, offset, false)
	task.SetCreateTime(uploadPrimaryStartTime.Unix())
	task.AppendLog(fmt.Sprintf("gateway-prepare-resumable-upload-task-cost:%d", time.Now().UnixMilli()-uploadPrimaryStartTime.UnixMilli()))
	task.AppendLog("gateway-create-resumable-upload-task")
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, task.Key().String())
	uploadDataTime := time.Now()
	err = g.baseApp.GfSpClient().ResumableUploadObject(ctx, task, r.Body)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_data_cost").Observe(time.Since(uploadDataTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_data_end").Observe(time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to resumable upload payload data", "error", err)
		return
	}
	log.CtxDebug(ctx, "succeed to resumable upload payload data")
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

	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
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

	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + "bucket_name: " + reqCtx.bucketName + " ,error: " + err.Error())
		return
	}

	params, err := g.baseApp.Consensus().QueryStorageParamsByTimestamp(reqCtx.Context(), objectInfo.GetLatestUpdatedTime())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get storage params from consensus, object_name: " + reqCtx.objectName + "bucket_name: " + reqCtx.bucketName + " ,error: " + err.Error())
		return
	}

	segmentCount, err = g.baseApp.GfSpClient().GetUploadObjectSegment(reqCtx.Context(), objectInfo.Id.Uint64())
	if err != nil && err.Error() != metadata.ErrNoRecord.String() {
		// ignore metadata.ErrNoRecord error
		log.CtxErrorw(reqCtx.Context(), "failed to get uploading object segment", "error", err)
		return
	}

	offset = uint64(segmentCount) * params.GetMaxSegmentSize()

	xmlInfo := struct {
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
		err = ErrEncodeResponseWithDetail("failed to marshal xml, object_name: " + reqCtx.objectName + "bucket_name: " + reqCtx.bucketName + " ,error: " + err.Error())
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponseWithDetail("failed to write body, object_name: " + reqCtx.objectName + "bucket_name: " + reqCtx.bucketName + " , error: " + err.Error())
		return
	}
	log.Debugw("succeed to query resumable offset", "xml_info", xmlInfo)
}

// getObjectHandler handles the download object request.
func (g *GateModular) getObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtxErr     error
		reqCtx        *RequestContext
		authenticated bool
	)
	getObjectStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(getObjectStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewayFailureGetObject).Inc()
			metrics.ReqTime.WithLabelValues(GatewayFailureGetObject).Observe(time.Since(getObjectStartTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(getObjectStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewaySuccessGetObject).Inc()
			metrics.ReqTime.WithLabelValues(GatewaySuccessGetObject).Observe(time.Since(getObjectStartTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	// GNFD1-ECDSA or GNFD1-EDDSA authentication, by checking the headers.
	reqCtx, reqCtxErr = NewRequestContext(r, g)

	if err = s3util.CheckValidBucketName(reqCtx.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqCtx.bucketName, "error", err)
		err = ErrInvalidQuery
		return
	}
	if err = s3util.CheckValidObjectName(reqCtx.objectName); err != nil {
		log.Errorw("failed to check object name", "object_name", reqCtx.objectName, "error", err)
		err = ErrInvalidQuery
		return
	}

	// check the object permission whether allow public read.
	verifyObjectPermissionTime := time.Now()
	var permission *permissiontypes.Effect
	if permission, err = g.baseApp.GfSpClient().VerifyPermission(reqCtx.Context(), sdk.AccAddress{}.String(),
		reqCtx.bucketName, reqCtx.objectName, permissiontypes.ACTION_GET_OBJECT); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authentication for getting public object", "error", err)
		return
	}
	if *permission == permissiontypes.EFFECT_ALLOW {
		authenticated = true
	}
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_verify_object_permission_time").Observe(time.Since(verifyObjectPermissionTime).Seconds())

	if !authenticated {
		// if not passed, then check authentication parameters
		if reqCtxErr != nil {
			queryParams := r.URL.Query()
			gnfdUserParam := queryParams.Get(GnfdUserAddressHeader)
			gnfdOffChainAuthAppDomainParam := queryParams.Get(GnfdOffChainAuthAppDomainHeader)
			gnfdOffChainAuthAppExpiryTimestampParam := queryParams.Get(commonhttp.HTTPHeaderExpiryTimestamp)
			gnfdOffChainAuthPublicKeyParam := queryParams.Get(GnfdOffChainAuthAppRegPublicKeyHeader)
			gnfdAuthorizationParam := queryParams.Get(GnfdAuthorizationHeader)

			// GNFD1-EDDSA
			gnfd1EddsaSignaturePrefix := commonhttp.Gnfd1Eddsa + ","
			// GNFD2-EDDSA
			gnfd2EddsaSignaturePrefix := commonhttp.Gnfd2Eddsa + ","
			if !strings.HasPrefix(gnfdAuthorizationParam, gnfd1EddsaSignaturePrefix) && !strings.HasPrefix(gnfdAuthorizationParam, gnfd2EddsaSignaturePrefix) {
				err = ErrUnsupportedSignType
				return
			}

			// if all required off-chain auth headers are passed in as query params, we fill corresponding headers
			if gnfdUserParam != "" && gnfdOffChainAuthAppDomainParam != "" && gnfdAuthorizationParam != "" && gnfdOffChainAuthAppExpiryTimestampParam != "" {
				var preSignedURLErr error
				var account sdk.AccAddress
				if strings.HasPrefix(gnfdAuthorizationParam, gnfd1EddsaSignaturePrefix) {
					account, preSignedURLErr = reqCtx.verifyGNFD1EddsaSignatureFromPreSignedURL(gnfdAuthorizationParam[len(gnfd1EddsaSignaturePrefix):], gnfdUserParam, gnfdOffChainAuthAppDomainParam)
				} else if strings.HasPrefix(gnfdAuthorizationParam, gnfd2EddsaSignaturePrefix) && gnfdOffChainAuthPublicKeyParam != "" {
					account, preSignedURLErr = reqCtx.verifyGNFD2EddsaSignatureFromPreSignedURL(gnfdAuthorizationParam[len(gnfd1EddsaSignaturePrefix):], gnfdUserParam, gnfdOffChainAuthAppDomainParam, gnfdOffChainAuthPublicKeyParam)
				}
				if preSignedURLErr != nil {
					reqCtxErr = preSignedURLErr
				}

				if account != nil {
					reqCtx.account = account.String()
					reqCtxErr = nil
					// default set content-disposition to download, if specified in query param as view, then set to view
					w.Header().Set(ContentDispositionHeader, ContentDispositionAttachmentValue+"; filename=\""+url.QueryEscape(reqCtx.objectName)+"\"")
					offChainAuthViewParam := queryParams.Get(OffChainAuthViewQuery)
					isView, _ := strconv.ParseBool(offChainAuthViewParam)
					if isView {
						w.Header().Set(ContentDispositionHeader, ContentDispositionInlineValue)
					}
				}
			}
		}

		if reqCtxErr != nil {
			err = reqCtxErr
			log.CtxErrorw(reqCtx.Context(), "no permission to operate, object is not public", "error", err)
			return
		}
		// check permission
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

	} // else anonymous users can get public object.

	// do the actual download
	err = g.downloadObject(w, reqCtx)
	if err != nil {
		return
	}
}

// downloadObject this is common method, which does the actual download action.
// It is called by both getObjectHandler and getObjectByUniversalEndpointHandler after passing the authentication and authorization.
func (g *GateModular) downloadObject(w http.ResponseWriter, reqCtx *RequestContext) error {
	var (
		err                       error
		params                    *storagetypes.Params
		bucketInfo                *storagetypes.BucketInfo
		objectInfo                *storagetypes.ObjectInfo
		isRange                   bool
		rangeStart, rangeEnd      int64
		lowOffset                 int64
		highOffset                int64
		pieceInfos                []*downloader.SegmentPieceInfo
		pieceData                 []byte
		extraQuota, consumedQuota uint64
		replyDataSize             int
		dbUpdateTimeStamp         int64
	)
	defer func() {
		if err != nil {
			// if the bucket exists extra quota when download object, recoup the quota to user
			if extraQuota > 0 {
				quotaUpdateErr := g.baseApp.GfSpClient().RecoupQuota(reqCtx.Context(), bucketInfo.Id.Uint64(), extraQuota, sqldb.TimestampYearMonth(dbUpdateTimeStamp))
				// no need to return the db error to user
				if quotaUpdateErr != nil {
					log.CtxErrorw(reqCtx.Context(), "failed to recoup extra quota to user", "error", err)
				}
				log.CtxDebugw(reqCtx.Context(), "//"+
					"success to recoup extra quota to user", "extra quota:", extraQuota)
			}
			w.Header().Del(ContentLengthHeader)
			w.Header().Del(ContentRangeHeader)
			w.Header().Del(ContentTypeHeader)
			w.Header().Del(ContentDispositionHeader)
		}
	}()

	getObjectTime := time.Now()
	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_get_object_info_time").Observe(time.Since(getObjectTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error:" + err.Error())
		return err
	}

	getBucketTime := time.Now()
	bucketInfo, err = g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_get_bucket_info_time").Observe(time.Since(getBucketTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get bucket info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return err
	}

	getParamTime := time.Now()
	params, err = g.baseApp.Consensus().QueryStorageParamsByTimestamp(reqCtx.Context(), objectInfo.GetLatestUpdatedTime())
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_get_storage_param_time").Observe(time.Since(getParamTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get storage params from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return err
	}

	isRange, rangeStart, rangeEnd = parseRange(reqCtx.request.Header.Get(RangeHeader))
	if isRange && (rangeEnd < 0 || rangeEnd >= int64(objectInfo.GetPayloadSize())) {
		rangeEnd = int64(objectInfo.GetPayloadSize()) - 1
	}
	if isRange && (rangeStart < 0 || rangeEnd < 0 || rangeStart > rangeEnd) {
		err = ErrInvalidRange
		return err
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
		return err
	}
	w.Header().Set(ContentTypeHeader, objectInfo.GetContentType())
	if isRange {
		w.Header().Set(ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(lowOffset))+
			"-"+util.Uint64ToString(uint64(highOffset)))
	} else {
		w.Header().Set(ContentLengthHeader, util.Uint64ToString(objectInfo.GetPayloadSize()))
	}

	getDataTime := time.Now()
	consumedQuota = 0
	extraQuota = 0
	downloadSize := uint64(highOffset - lowOffset + 1)
	for idx, pInfo := range pieceInfos {
		enableCheck := false
		if idx == 0 { // only check in first piece
			enableCheck = true
			dbUpdateTimeStamp = sqldb.GetCurrentTimestampUs()
		}
		pieceTask := &gfsptask.GfSpDownloadPieceTask{}
		pieceTask.InitDownloadPieceTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(task),
			enableCheck, reqCtx.Account(), downloadSize, pInfo.SegmentPieceKey, pInfo.Offset,
			pInfo.Length, g.baseApp.TaskTimeout(task, uint64(pieceTask.GetSize())), g.baseApp.TaskMaxRetry(task))
		getSegmentTime := time.Now()
		pieceData, err = g.baseApp.GfSpClient().GetPiece(reqCtx.Context(), pieceTask)

		metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_segment_data_time").Observe(time.Since(getSegmentTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to download piece", "error", err)
			downloaderErr := gfsperrors.MakeGfSpError(err)
			// if it is the first piece and the quota db is not updated, no extra data need to updated
			if idx >= 1 || (idx == 0 && downloaderErr.GetInnerCode() == 85101) {
				extraQuota = downloadSize - consumedQuota
			}
			return err
		}

		writeTime := time.Now()
		replyDataSize, err = w.Write(pieceData)
		// if the connection of client has been disconnected, the response will fail
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to write the data to connection", "objectName", objectInfo.ObjectName, "error", err)
			extraQuota = downloadSize - consumedQuota
			err = ErrReplyData
			return err
		}
		// the quota value should be computed by the reply content length
		consumedQuota += uint64(replyDataSize)
		metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_write_time").Observe(time.Since(writeTime).Seconds())
	}

	metrics.ReqPieceSize.WithLabelValues(GatewayGetObjectSize).Observe(float64(highOffset - lowOffset + 1))
	metrics.PerfGetObjectTimeHistogram.WithLabelValues("get_object_get_data_time").Observe(time.Since(getDataTime).Seconds())
	return nil
}

// queryUploadProgressHandler handles the query uploaded object progress request.
func (g *GateModular) queryUploadProgressHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                  error
		reqCtx               *RequestContext
		authenticated        bool
		objectInfo           *storagetypes.ObjectInfo
		errDescription       string
		taskStateDescription string
		taskState            int32
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
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

	objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + " ,error: " + err.Error())
		return
	}
	if objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_CREATED {
		taskState, errDescription, err = g.baseApp.GfSpClient().GetUploadObjectState(reqCtx.Context(), objectInfo.Id.Uint64())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to get uploading job state", "error", err)
			if !strings.Contains(err.Error(), "no uploading record") {
				return
			}
			taskState = int32(servicetypes.TaskState_TASK_STATE_INIT_UNSPECIFIED)
			taskStateDescription = servicetypes.StateToDescription(servicetypes.TaskState(taskState))
			err = nil
		} else {
			taskStateDescription = servicetypes.StateToDescription(servicetypes.TaskState(taskState))
		}
	} else if objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_SEALED && !objectInfo.GetIsUpdating() {
		taskState = int32(servicetypes.TaskState_TASK_STATE_SEAL_OBJECT_DONE)
		taskStateDescription = servicetypes.StateToDescription(servicetypes.TaskState(taskState))
	} else if objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_DISCONTINUED {
		taskState = int32(servicetypes.TaskState_TASK_STATE_OBJECT_DISCONTINUED)
		taskStateDescription = servicetypes.StateToDescription(servicetypes.TaskState(taskState))
	}

	xmlInfo := struct {
		XMLName             xml.Name `xml:"QueryUploadProgress"`
		Version             string   `xml:"version,attr"`
		ProgressDescription string   `xml:"ProgressDescription"`
		ErrorDescription    string   `xml:"ErrorDescription"`
	}{
		Version:             GnfdResponseXMLVersion,
		ProgressDescription: taskStateDescription,
		ErrorDescription:    errDescription,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponseWithDetail("failed to marshal xml for query upload progress, object_name: " + reqCtx.objectName + "bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponseWithDetail("failed to write body, object_name: " + reqCtx.objectName + "bucket_name: " + reqCtx.bucketName + ",error: " + err.Error())
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
		redirectURL          string
		bucketInfo           *storagetypes.BucketInfo
		objectInfo           *storagetypes.ObjectInfo
		isRequestFromBrowser bool
		spEndpoint           string
		getEndpointErr       error
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			spErrCode := gfsperrors.MakeGfSpError(err).GetInnerCode()
			if isRequestFromBrowser {
				reqCtx.SetHTTPCode(http.StatusOK)
				errorCodeForPage := "INTERNAL_ERROR" // default errorCode in built-in error page
				switch spErrCode {
				case downloader.ErrExceedBucketQuota.GetInnerCode():
					errorCodeForPage = "NO_ENOUGH_QUOTA"
				case ErrNoSuchObject.GetInnerCode():
					errorCodeForPage = "FILE_NOT_FOUND"
				case ErrForbidden.GetInnerCode():
					errorCodeForPage = "NO_PERMISSION"
				}
				html := strings.Replace(GnfdBuiltInUniversalEndpointDappErrorPage, "<% errorCode %>", errorCodeForPage, 1)

				_, _ = fmt.Fprintf(w, "%s", html)
				metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
				metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
				return
			} else {
				reqCtx.SetError(gfsperrors.MakeGfSpError(err))
				reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
				modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
				metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
				metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
			}

		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	userAgent := r.Header.Get("User-Agent")
	isRequestFromBrowser = checkIfRequestFromBrowser(userAgent)

	// ignore the error, because the universal endpoint does not need signature
	reqCtx, _ = NewRequestContext(r, g)

	if err = s3util.CheckValidBucketName(reqCtx.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqCtx.bucketName, "error", err)
		err = ErrInvalidQuery
		return
	}
	if err = s3util.CheckValidObjectName(reqCtx.objectName); err != nil {
		log.Errorw("failed to check object name", "object_name", reqCtx.objectName, "error", err)
		err = ErrInvalidQuery
		return
	}

	getBucketInfoRes, getBucketInfoErr := g.baseApp.GfSpClient().GetBucketByBucketName(reqCtx.Context(), reqCtx.bucketName, true)
	if getBucketInfoErr != nil || getBucketInfoRes == nil || getBucketInfoRes.GetBucketInfo() == nil {
		log.Errorw("failed to check bucket info", "bucket_name", reqCtx.bucketName, "error", getBucketInfoErr)
		err = ErrNoSuchObject
		return
	}

	bucketInfo = getBucketInfoRes.BucketInfo
	// if bucket not in the current sp, 302 redirect to the sp that contains the bucket
	// TODO get from config
	spID, err := g.getSPID()
	if err != nil {
		err = ErrConsensusWithDetail("failed to getSPID, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + " ,error: " + err.Error())
		return
	}
	bucketSPID, err := util.GetBucketPrimarySPID(reqCtx.Context(), g.baseApp.Consensus(), getBucketInfoRes.GetBucketInfo())
	if err != nil {
		err = ErrConsensusWithDetail("failed to GetBucketPrimarySPID, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ",error: " + err.Error())
		return
	}
	if spID != bucketSPID {
		log.Debugw("primary sp id not matched ", "bucket_sp_id", bucketSPID, "self_sp_id", spID)

		// get the endpoint where the bucket actually is in
		spEndpoint, getEndpointErr = g.baseApp.GfSpClient().GetEndpointBySpID(reqCtx.Context(), bucketSPID)
		if getEndpointErr != nil || spEndpoint == "" {
			log.Errorw("failed to get endpoint by id", "sp_id", bucketSPID, "error", getEndpointErr)
			err = getEndpointErr
			return
		}

		redirectURL = spEndpoint + r.URL.RequestURI()
		log.Debugw("getting redirect url:", "redirectURL", redirectURL)

		http.Redirect(w, r, redirectURL, 302)
		return
	}
	getObjectInfoRes, err := g.baseApp.GfSpClient().GetObjectMeta(reqCtx.Context(), reqCtx.objectName, reqCtx.bucketName, true)
	if err != nil || getObjectInfoRes == nil || getObjectInfoRes.GetObjectInfo() == nil {
		log.Errorw("failed to check object meta", "object_name", reqCtx.objectName, "error", err)
		err = ErrNoSuchObject
		return
	}
	objectInfo = getObjectInfoRes.ObjectInfo
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.Errorw("object is not sealed",
			"status", getObjectInfoRes.GetObjectInfo().GetObjectStatus())
		err = ErrNoSuchObject
		return
	}
	if isPrivateObject(bucketInfo, objectInfo) {
		// for private files, we return a built-in dapp and help users provide a signature for verification
		var (
			expiry    string
			signature string
		)
		requestURI := r.URL.RequestURI()

		splitPeriod := strings.Split(requestURI, ".")
		splitSuffix := splitPeriod[len(splitPeriod)-1]
		if !strings.Contains(requestURI, objectSpecialSuffixUrlReplacement) &&
			(strings.EqualFold(splitSuffix, ObjectPdfSuffix) || strings.EqualFold(splitSuffix, ObjectXmlSuffix)) {
			objectPathRequestURL := "/" + strings.Replace(requestURI[1:], "/", objectSpecialSuffixUrlReplacement, 1)
			redirectURL = spEndpoint + objectPathRequestURL
			log.Debugw("getting redirect url for private object:", "redirectURL", redirectURL)
			http.Redirect(w, r, redirectURL, 302)
			return
		}

		queryParams := r.URL.Query()
		if queryParams[commonhttp.HTTPHeaderExpiryTimestamp] != nil {
			expiry = queryParams[commonhttp.HTTPHeaderExpiryTimestamp][0]
		}
		if queryParams["signature"] != nil {
			signature = queryParams["signature"][0]
		}
		if expiry != "" && signature != "" {
			// check if expiry set to far or expiry is past
			expiryDate, dateParseErr := time.Parse(ExpiryDateFormat, expiry)
			if dateParseErr != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to parse expiry date due to invalid format", "expiry", expiry)
				err = ErrInvalidExpiryDateParam
				return
			}
			expiryAge := int32(time.Until(expiryDate).Seconds())
			if MaxExpiryAgeInSec < expiryAge || expiryAge < 0 {
				err = ErrInvalidExpiryDateParam
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
			htmlConfigMap := map[string]string{
				"mechain_7971-1":    "{\n  \"envType\": \"dev\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 7971,\n  \"chainName\": \"dev - mechain\",\n  \"rpcUrls\": [\"https://gnfd-dev.qa.mechain.world\"],\n  \"nativeCurrency\": { \"name\": \"azkme\", \"symbol\": \"azkme\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://mechainscan-qanet.fe.nodereal.cc/\"]\n}\n",
				"mechain_9000-1741": "{\n  \"envType\": \"qa\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 9000,\n  \"chainName\": \"qa - mechain\",\n  \"rpcUrls\": [\"https://gnfd.qa.mechain.world\"],\n  \"nativeCurrency\": { \"name\": \"azkme\", \"symbol\": \"azkme\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://mechainscan-qanet.fe.nodereal.cc/\"]\n}\n",
				"mechain_5151-1":    "{\n  \"envType\": \"testnet\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 5600,\n  \"chainName\": \"mechain testnet\",\n  \"rpcUrls\": [\"https://gnfd-testnet-fullnode-tendermint-us.mechain.org\"],\n  \"nativeCurrency\": { \"name\": \"azkme\", \"symbol\": \"azkme\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://mechainscan.com/\"]\n}\n",
				"mechain_920-1":     "{\n  \"envType\": \"pre-mainnet\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 920,\n  \"chainName\": \"mechain pre-main-net\",\n  \"rpcUrls\": [\"https://zk.me\"],\n  \"nativeCurrency\": { \"name\": \"azkme\", \"symbol\": \"azkme\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://mechainscan.com/\"]\n}\n",
				"mechain_5252-1":    "{\n  \"envType\": \"mainnet\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 1017,\n  \"chainName\": \"mechain main-net\",\n  \"rpcUrls\": [\"https://zk.me\"],\n  \"nativeCurrency\": { \"name\": \"azkme\", \"symbol\": \"azkme\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://mechainscan.com/\"]\n}\n",
			}

			htmlConfig := htmlConfigMap[g.baseApp.ChainID()]
			if htmlConfig == "" {
				htmlConfig = htmlConfigMap["mechain_5151-1"] // use testnet by default, actually, we only need the metamask sign function, regardless which mechain network the metamask is going to connect.
			}
			hc, _ := json.Marshal(htmlConfig)
			html := strings.Replace(GnfdBuiltInUniversalEndpointDappHtml, "<% env %>", string(hc), 1)

			_, _ = fmt.Fprintf(w, "%s", html)
			return
		}

	}
	if isDownload {
		w.Header().Set(ContentDispositionHeader, ContentDispositionAttachmentValue+"; filename=\""+url.QueryEscape(reqCtx.objectName)+"\"")
	} else {
		w.Header().Set(ContentDispositionHeader, ContentDispositionInlineValue)
	}

	// do the actual download
	err = g.downloadObject(w, reqCtx)
	if err != nil {
		return
	}
	log.CtxDebugw(reqCtx.Context(), "succeed to download object for universal endpoint")
}

// putObjectHandler handles the upload object request.
func (g *GateModular) delegatePutObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		bucketInfo    *storagetypes.BucketInfo
		objectInfo    *storagetypes.ObjectInfo
		params        *storagetypes.Params
		approvalMsg   []byte
		fingerprint   []byte
		payloadSize   uint64
		contentType   string
		visibility    storagetypes.VisibilityType
		txHash        string
		isUpdate      bool
	)

	uploadPrimaryStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewayFailurePutObject).Inc()
			metrics.ReqTime.WithLabelValues(GatewayFailurePutObject).Observe(time.Since(uploadPrimaryStartTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqPieceSize.WithLabelValues(GatewayPutObjectSize).Observe(float64(objectInfo.GetPayloadSize()))
			metrics.ReqTime.WithLabelValues(GatewaySuccessPutObject).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewaySuccessPutObject).Inc()
			metrics.ReqPieceSize.WithLabelValues(GatewaySuccessPutObject).Observe(float64(objectInfo.GetPayloadSize()))
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	queryParams := reqCtx.request.URL.Query()
	isUpdateStr := queryParams.Get("is_update")
	isUpdate, err = strconv.ParseBool(isUpdateStr)
	if err != nil {
		log.CtxErrorw(reqCtx.ctx, "failed to parse is_update", "error", err)
		return
	}

	err = s3util.CheckValidBucketName(reqCtx.bucketName)
	if err != nil {
		return
	}

	err = s3util.CheckValidObjectName(reqCtx.objectName)
	if err != nil {
		return
	}

	startAuthenticationTime := time.Now()
	if isUpdate {
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(), coremodule.AuthOpTypeAgentUpdateObject,
			reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		metrics.PerfPutObjectTime.WithLabelValues("gateway_agent_put_object_authorizer").Observe(time.Since(startAuthenticationTime).Seconds())
	} else {
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(), coremodule.AuthOpTypeAgentPutObject,
			reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
		metrics.PerfPutObjectTime.WithLabelValues("gateway_agent_put_object_authorizer").Observe(time.Since(startAuthenticationTime).Seconds())
	}
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
		return
	}
	if !authenticated {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate")
		err = ErrNoPermission
		return
	}
	contentType = r.Header.Get(ContentTypeHeader)
	if contentType == "" {
		contentType = ContentDefault
	}
	if strings.Contains(reqCtx.objectName, "..") ||
		reqCtx.objectName == "/" ||
		strings.Contains(reqCtx.objectName, "\\") ||
		util.IsSQLInjection(reqCtx.objectName) {
		log.Errorw("failed to check object name", "object_name", reqCtx.objectName)
		err = ErrInvalidObjectName
		return
	}

	startGetStorageParamTime := time.Now()
	params, err = g.baseApp.Consensus().QueryStorageParamsByTimestamp(reqCtx.Context(), startGetStorageParamTime.Unix())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_agent_put_object_query_params_cost").Observe(time.Since(startGetStorageParamTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_agent_put_object_query_params_end").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get storage params from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error" + err.Error())
		return
	}

	payloadSizeStr := queryParams.Get("payload_size")
	payloadSize, err = strconv.ParseUint(payloadSizeStr, 10, 64)
	if err != nil {
		log.CtxErrorw(reqCtx.ctx, "failed to parse payload_size", "error", err)
		return
	}
	if payloadSize == 0 || payloadSize > params.GetMaxPayloadSize() {
		log.CtxErrorw(reqCtx.Context(), "failed to put object payload size error", "size", payloadSize, "size limit", params.GetMaxPayloadSize())
		err = ErrInvalidPayloadSize
		return
	}

	if err = g.checkSPAndBucketStatus(reqCtx.Context(), reqCtx.bucketName, reqCtx.account); err != nil {
		log.CtxErrorw(reqCtx.Context(), "put object failed to check sp and bucket status", "error", err)
		return
	}

	approvalMsg, err = hex.DecodeString(r.Header.Get(GnfdUnsignedApprovalMsgHeader))
	if err != nil {
		log.Errorw("failed to parse approval header",
			"approval", r.Header.Get(GnfdUnsignedApprovalMsgHeader))
		err = ErrDecodeMsg
		return
	}
	fingerprint = commonhash.GenerateChecksum(approvalMsg)

	startTime := time.Now()
	if isUpdate {
		msg := &storagetypes.MsgDelegateUpdateObjectContent{
			Updater:     reqCtx.account,
			BucketName:  reqCtx.bucketName,
			ObjectName:  reqCtx.objectName,
			PayloadSize: payloadSize,
			ContentType: contentType,
		}
		txHash, err = g.baseApp.GfSpClient().DelegateUpdateObjectContent(reqCtx.ctx, msg)
		if err != nil {
			log.CtxErrorw(reqCtx.ctx, "failed to delegate update object", "error", err)
			return
		}
	} else {
		// if object has been created, we can skip the creation process
		objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.ctx, reqCtx.bucketName, reqCtx.objectName)
		if err != nil && !strings.Contains(err.Error(), "No such object") {
			log.CtxErrorw(reqCtx.ctx, "failed to QueryObjectInfo", "error", err)
			return
		}
		if objectInfo != nil && (objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_CREATED || (objectInfo.Creator != reqCtx.account && objectInfo.Owner != reqCtx.account) || objectInfo.PayloadSize != payloadSize) {
			err = ErrInvalidQuery
			return
		}
		if objectInfo == nil {
			var visibilityInt int64
			visibilityStr := queryParams.Get("visibility")
			visibilityInt, err = strconv.ParseInt(visibilityStr, 10, 32)
			if err != nil {
				return
			}
			visibility = storagetypes.VisibilityType(visibilityInt)
			if visibility == storagetypes.VISIBILITY_TYPE_UNSPECIFIED {
				visibility = storagetypes.VISIBILITY_TYPE_INHERIT // set default visibility type
			}
			task := &gfsptask.GfSpDelegateCreateObjectApprovalTask{}
			task.InitApprovalDelegateCreateObjectTask(reqCtx.Account(), &storagetypes.MsgDelegateCreateObject{
				Operator:       g.baseApp.OperatorAddress(),
				Creator:        reqCtx.account,
				BucketName:     reqCtx.bucketName,
				ObjectName:     reqCtx.objectName,
				PayloadSize:    payloadSize,
				ContentType:    contentType,
				Visibility:     visibility,
				RedundancyType: storagetypes.REDUNDANCY_EC_TYPE,
			}, fingerprint, g.baseApp.TaskPriority(task))
			startAskCreateObjectApproval := time.Now()
			authenticated, _, err = g.baseApp.GfSpClient().AskDelegateCreateObjectApproval(reqCtx.Context(), task)
			metrics.PerfApprovalTime.WithLabelValues("gateway_delegate_create_object_ask_approval_cost").Observe(time.Since(startAskCreateObjectApproval).Seconds())
			metrics.PerfApprovalTime.WithLabelValues("gateway_delegate_create_object_ask_approval_end").Observe(time.Since(startTime).Seconds())
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to ask object approval", "error", err)
				return
			}
			if !authenticated {
				log.CtxErrorw(reqCtx.Context(), "refuse the ask create object approval")
				return
			}
			startDelegateCreateObject := time.Now()
			txHash, err = g.baseApp.GfSpClient().DelegateCreateObject(reqCtx.ctx, task.GetDelegateCreateObject())
			metrics.PerfApprovalTime.WithLabelValues("delegate_create_object_cost").Observe(time.Since(startDelegateCreateObject).Seconds())
			metrics.PerfApprovalTime.WithLabelValues("delegate_create_object_end").Observe(time.Since(startDelegateCreateObject).Seconds())
			if err != nil {
				log.CtxErrorw(reqCtx.ctx, "failed to delegate create object", "error", err)
				return
			}
		}
	}

	if txHash != "" {
		_, err = g.baseApp.Consensus().ConfirmTransaction(reqCtx.ctx, txHash)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ConfirmTransaction", "error", err)
			return
		}
	}

	startGetObjectInfoTime := time.Now()
	bucketInfo, objectInfo, err = g.baseApp.Consensus().QueryBucketInfoAndObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_agent_put_object_query_object_cost").Observe(time.Since(startGetObjectInfoTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_agent_put_object_query_object_end").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}
	err = g.checkAndAssignShadowObjectInfo(reqCtx, objectInfo)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}

	uploadTask := &gfsptask.GfSpUploadObjectTask{}
	uploadTask.InitUploadObjectTask(bucketInfo.GetGlobalVirtualGroupFamilyId(), objectInfo, params, g.baseApp.TaskTimeout(uploadTask, objectInfo.GetPayloadSize()), true)
	uploadTask.SetCreateTime(uploadPrimaryStartTime.Unix())
	uploadTask.AppendLog(fmt.Sprintf("gateway-prepare-upload-task-cost:%d", time.Now().UnixMilli()-uploadPrimaryStartTime.UnixMilli()))
	uploadTask.AppendLog("gateway-create-upload-task")
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, uploadTask.Key().String())
	uploadDataTime := time.Now()
	err = g.baseApp.GfSpClient().UploadObject(ctx, uploadTask, r.Body)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_agent_put_object_data_cost").Observe(time.Since(uploadDataTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_agent_put_object_data_end").Observe(time.Since(time.Unix(uploadTask.GetCreateTime(), 0)).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to upload payload data", "error", err)
		return
	}
	log.CtxDebug(ctx, "succeed to upload payload data")
}

func (g *GateModular) delegateResumablePutObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		bucketInfo    *storagetypes.BucketInfo
		objectInfo    *storagetypes.ObjectInfo
		params        *storagetypes.Params
		txHash        string
		approvalMsg   []byte
		fingerprint   []byte
		payloadSize   uint64
		contentType   string
		visibility    storagetypes.VisibilityType
		isUpdate      bool
	)

	uploadPrimaryStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(uploadPrimaryStartTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(uploadPrimaryStartTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	queryParams := reqCtx.request.URL.Query()
	isUpdateStr := queryParams.Get("is_update")
	isUpdate, err = strconv.ParseBool(isUpdateStr)
	if err != nil {
		log.CtxErrorw(reqCtx.ctx, "failed to parse is_update", "error", err)
		return
	}

	err = s3util.CheckValidBucketName(reqCtx.bucketName)
	if err != nil {
		return
	}

	err = s3util.CheckValidObjectName(reqCtx.objectName)
	if err != nil {
		return
	}
	if strings.Contains(reqCtx.objectName, "..") ||
		reqCtx.objectName == "/" ||
		strings.Contains(reqCtx.objectName, "\\") ||
		util.IsSQLInjection(reqCtx.objectName) {
		log.Errorw("failed to check object name", "object_name", reqCtx.objectName)
		err = ErrInvalidObjectName
		return
	}
	if err = g.checkSPAndBucketStatus(reqCtx.Context(), reqCtx.bucketName, reqCtx.account); err != nil {
		log.CtxErrorw(reqCtx.Context(), "resumable put object failed to check sp and bucket status", "error", err)
		return
	}

	payloadSizeStr := queryParams.Get("payload_size")
	payloadSize, err = strconv.ParseUint(payloadSizeStr, 10, 64)
	if err != nil {
		return
	}
	// the resumable upload utilizes the on-chain MaxPayloadSize as the maximum file size
	startGetStorageParamTime := time.Now()
	params, err = g.baseApp.Consensus().QueryStorageParams(reqCtx.Context())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_query_params_cost").Observe(time.Since(startGetStorageParamTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_query_params_end").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get storage params from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}
	if payloadSize == 0 || payloadSize > params.GetMaxPayloadSize() {
		log.CtxErrorw(reqCtx.Context(), "failed to put object payload size is too large", "size", payloadSize, "size limit", params.GetMaxPayloadSize())
		err = ErrInvalidPayloadSize
		return
	}

	if isUpdate {
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(), coremodule.AuthOpTypeAgentUpdateObject,
			reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
	} else {
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(), coremodule.AuthOpTypeAgentPutObject,
			reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
	}
	if !authenticated {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate")
		err = ErrNoPermission
		return
	}

	approvalMsg, err = hex.DecodeString(r.Header.Get(GnfdUnsignedApprovalMsgHeader))
	if err != nil {
		log.Errorw("failed to parse approval header",
			"approval", r.Header.Get(GnfdUnsignedApprovalMsgHeader))
		err = ErrDecodeMsg
		return
	}
	fingerprint = commonhash.GenerateChecksum(approvalMsg)
	contentType = r.Header.Get(ContentTypeHeader)
	if contentType == "" {
		contentType = ContentDefault
	}
	startTime := time.Now()

	if isUpdate {
		objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.ctx, reqCtx.bucketName, reqCtx.objectName)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
			err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
			return
		}

		if !objectInfo.IsUpdating {
			msg := &storagetypes.MsgDelegateUpdateObjectContent{
				Updater:     reqCtx.account,
				BucketName:  reqCtx.bucketName,
				ObjectName:  reqCtx.objectName,
				PayloadSize: payloadSize,
				ContentType: contentType,
			}
			txHash, err = g.baseApp.GfSpClient().DelegateUpdateObjectContent(reqCtx.ctx, msg)
			if err != nil {
				log.CtxErrorw(reqCtx.ctx, "failed to delegate update object", "error", err)
				return
			}
		}

	} else {
		var visibilityInt int64
		visibilityStr := queryParams.Get("visibility")
		visibilityInt, err = strconv.ParseInt(visibilityStr, 10, 32)
		if err != nil {
			return
		}
		visibility = storagetypes.VisibilityType(visibilityInt)
		if visibility == storagetypes.VISIBILITY_TYPE_UNSPECIFIED {
			visibility = storagetypes.VISIBILITY_TYPE_INHERIT // set default visibility type
		}
		task := &gfsptask.GfSpDelegateCreateObjectApprovalTask{}
		task.InitApprovalDelegateCreateObjectTask(reqCtx.Account(), &storagetypes.MsgDelegateCreateObject{
			Operator:       g.baseApp.OperatorAddress(),
			Creator:        reqCtx.account,
			BucketName:     reqCtx.bucketName,
			ObjectName:     reqCtx.objectName,
			PayloadSize:    payloadSize,
			ContentType:    contentType,
			Visibility:     visibility,
			RedundancyType: storagetypes.REDUNDANCY_EC_TYPE,
		}, fingerprint, g.baseApp.TaskPriority(task))

		objectInfo, err = g.baseApp.Consensus().QueryObjectInfo(reqCtx.ctx, reqCtx.bucketName, reqCtx.objectName)
		if err != nil && strings.Contains(err.Error(), "No such object") {
			startAskCreateObjectApproval := time.Now()
			authenticated, _, err = g.baseApp.GfSpClient().AskDelegateCreateObjectApproval(reqCtx.Context(), task)
			metrics.PerfApprovalTime.WithLabelValues("gateway_delegate_create_object_cost").Observe(time.Since(startAskCreateObjectApproval).Seconds())
			metrics.PerfApprovalTime.WithLabelValues("gateway_delegate_create_object_cost").Observe(time.Since(startTime).Seconds())
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to ask object approval", "error", err)
				return
			}
			if !authenticated {
				log.CtxErrorw(reqCtx.Context(), "refuse the ask create object approval")
				err = ErrRefuseApproval
				return
			}
			startDelegateCreateObject := time.Now()
			txHash, err = g.baseApp.GfSpClient().DelegateCreateObject(reqCtx.ctx, task.GetDelegateCreateObject())
			metrics.PerfApprovalTime.WithLabelValues("approval_object_sign_create_object_cost").Observe(time.Since(startDelegateCreateObject).Seconds())
			metrics.PerfApprovalTime.WithLabelValues("approval_object_sign_create_object_end").Observe(time.Since(startDelegateCreateObject).Seconds())
			if err != nil {
				log.CtxErrorw(reqCtx.ctx, "failed to delegate create object", "error", err)
				return
			}
		} else if err != nil {
			log.CtxErrorw(reqCtx.ctx, "failed to QueryObjectInfo", "error", err)
			return
		} else if objectInfo.ObjectStatus != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(reqCtx.ctx, "object has been sealed", "status", objectInfo.ObjectStatus)
			return
		}
	}
	if txHash != "" {
		_, err = g.baseApp.Consensus().ConfirmTransaction(reqCtx.ctx, txHash)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to WaitForNextBlock", "error", err)
			return
		}
	}

	startGetObjectInfoTime := time.Now()
	bucketInfo, objectInfo, err = g.baseApp.Consensus().QueryBucketInfoAndObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_query_object_cost").Observe(time.Since(startGetObjectInfoTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_query_object_end").Observe(time.Since(uploadPrimaryStartTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}

	err = g.checkAndAssignShadowObjectInfo(reqCtx, objectInfo)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get object info from consensus, object_name: " + reqCtx.objectName + ", bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}

	var (
		complete        bool
		offset          uint64
		requestComplete string
		requestOffset   string
	)
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
			log.CtxErrorw(reqCtx.Context(), "failed to parse offset from url", "error", err)
			err = ErrInvalidOffset
			return
		}
	} else {
		err = ErrInvalidOffset
		return
	}

	uploadTask := &gfsptask.GfSpResumableUploadObjectTask{}
	uploadTask.InitResumableUploadObjectTask(bucketInfo.GetGlobalVirtualGroupFamilyId(), objectInfo, params, g.baseApp.TaskTimeout(uploadTask, objectInfo.GetPayloadSize()), complete, offset, true)
	uploadTask.SetCreateTime(uploadPrimaryStartTime.Unix())
	uploadTask.AppendLog(fmt.Sprintf("gateway-prepare-resumable-upload-task-cost:%d", time.Now().UnixMilli()-uploadPrimaryStartTime.UnixMilli()))
	uploadTask.AppendLog("gateway-create-resumable-upload-task")
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, uploadTask.Key().String())
	uploadDataTime := time.Now()
	err = g.baseApp.GfSpClient().ResumableUploadObject(ctx, uploadTask, r.Body)
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_data_cost").Observe(time.Since(uploadDataTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("gateway_resumable_put_object_data_end").Observe(time.Since(time.Unix(uploadTask.GetCreateTime(), 0)).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to resumable upload payload data", "error", err)
		return
	}
	log.CtxDebug(ctx, "succeed to resumable upload payload data")
}

// delegateCreateFolderHandler handles the delegate create folder request.
func (g *GateModular) delegateCreateFolderHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		objectInfo    *storagetypes.ObjectInfo
		approvalMsg   []byte
		fingerprint   []byte
		contentType   string
		visibility    storagetypes.VisibilityType
		txHash        string
	)

	uploadPrimaryStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewayFailurePutObject).Inc()
			metrics.ReqTime.WithLabelValues(GatewayFailurePutObject).Observe(time.Since(uploadPrimaryStartTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqPieceSize.WithLabelValues(GatewayPutObjectSize).Observe(float64(objectInfo.GetPayloadSize()))
			metrics.ReqTime.WithLabelValues(GatewaySuccessPutObject).Observe(time.Since(uploadPrimaryStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewaySuccessPutObject).Inc()
			metrics.ReqPieceSize.WithLabelValues(GatewaySuccessPutObject).Observe(float64(objectInfo.GetPayloadSize()))
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	queryParams := reqCtx.request.URL.Query()

	err = s3util.CheckValidBucketName(reqCtx.bucketName)
	if err != nil {
		return
	}

	err = s3util.CheckValidObjectName(reqCtx.objectName)
	if err != nil {
		return
	}

	authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(), coremodule.AuthOpTypeAgentPutObject,
		reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
		return
	}

	if !authenticated {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate")
		err = ErrNoPermission
		return
	}
	contentType = r.Header.Get(ContentTypeHeader)
	if contentType == "" {
		contentType = ContentDefault
	}
	if strings.Contains(reqCtx.objectName, "..") ||
		reqCtx.objectName == "/" ||
		strings.Contains(reqCtx.objectName, "\\") ||
		util.IsSQLInjection(reqCtx.objectName) {
		log.Errorw("failed to check object name", "object_name", reqCtx.objectName)
		err = ErrInvalidObjectName
		return
	}

	if err = g.checkSPAndBucketStatus(reqCtx.Context(), reqCtx.bucketName, reqCtx.account); err != nil {
		log.CtxErrorw(reqCtx.Context(), "put object failed to check sp and bucket status", "error", err)
		return
	}

	approvalMsg, err = hex.DecodeString(r.Header.Get(GnfdUnsignedApprovalMsgHeader))
	if err != nil {
		log.Errorw("failed to parse approval header",
			"approval", r.Header.Get(GnfdUnsignedApprovalMsgHeader))
		err = ErrDecodeMsg
		return
	}
	fingerprint = commonhash.GenerateChecksum(approvalMsg)

	startTime := time.Now()
	// if object has been created, we can skip the creation process
	objectInfo, _ = g.baseApp.Consensus().QueryObjectInfo(reqCtx.ctx, reqCtx.bucketName, reqCtx.objectName)
	if objectInfo != nil {
		err = ErrInvalidQuery
		return
	} else {
		var visibilityInt int64
		visibilityStr := queryParams.Get("visibility")
		visibilityInt, err = strconv.ParseInt(visibilityStr, 10, 32)
		if err != nil {
			return
		}
		visibility = storagetypes.VisibilityType(visibilityInt)
		if visibility == storagetypes.VISIBILITY_TYPE_UNSPECIFIED {
			visibility = storagetypes.VISIBILITY_TYPE_INHERIT // set default visibility type
		}
		task := &gfsptask.GfSpDelegateCreateObjectApprovalTask{}
		task.InitApprovalDelegateCreateObjectTask(reqCtx.Account(), &storagetypes.MsgDelegateCreateObject{
			Operator:       g.baseApp.OperatorAddress(),
			Creator:        reqCtx.account,
			BucketName:     reqCtx.bucketName,
			ObjectName:     reqCtx.objectName,
			PayloadSize:    0,
			ContentType:    contentType,
			Visibility:     visibility,
			RedundancyType: storagetypes.REDUNDANCY_EC_TYPE,
		}, fingerprint, g.baseApp.TaskPriority(task))
		startAskCreateObjectApproval := time.Now()
		authenticated, _, err = g.baseApp.GfSpClient().AskDelegateCreateObjectApproval(reqCtx.Context(), task)
		metrics.PerfApprovalTime.WithLabelValues("gateway_delegate_create_object_ask_approval_cost").Observe(time.Since(startAskCreateObjectApproval).Seconds())
		metrics.PerfApprovalTime.WithLabelValues("gateway_delegate_create_object_ask_approval_end").Observe(time.Since(startTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ask object approval", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "refuse the ask create object approval")
			return
		}
		startDelegateCreateObject := time.Now()
		txHash, err = g.baseApp.GfSpClient().DelegateCreateObject(reqCtx.ctx, task.GetDelegateCreateObject())
		metrics.PerfApprovalTime.WithLabelValues("delegate_create_object_cost").Observe(time.Since(startDelegateCreateObject).Seconds())
		metrics.PerfApprovalTime.WithLabelValues("delegate_create_object_end").Observe(time.Since(startDelegateCreateObject).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.ctx, "failed to delegate create object", "error", err)
			return
		}
	}

	if txHash != "" {
		_, err = g.baseApp.Consensus().ConfirmTransaction(reqCtx.ctx, txHash)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ConfirmTransaction", "error", err)
			return
		}
	}

	type CreateHash struct {
		TxHash string `xml:"TxHash"`
	}
	respBytes, err := xml.Marshal(CreateHash{TxHash: txHash})
	if err != nil {
		log.Errorf("failed to Marshal response", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)

	if _, err = w.Write(respBytes); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponseWithDetail("failed to write body, object_name: " + reqCtx.objectName + "bucket_name: " + reqCtx.bucketName + ", error: " + err.Error())
		return
	}
	log.Debugw("succeed to delegate create folder", "xml_info", respBytes)
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
