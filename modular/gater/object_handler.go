package gater

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield/types/s3util"

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

	reqCtx, err = NewRequestContext(r, g)
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
	reqCtx, reqCtxErr = NewRequestContext(r, g)
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

	isRange, rangeStart, rangeEnd := parseRange(reqCtx.request.Header.Get(RangeHeader))
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
	w.Header().Set(ContentTypeHeader, objectInfo.GetContentType())
	if isRange {
		w.Header().Set(ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(low))+
			"-"+util.Uint64ToString(uint64(high)))
	} else {
		w.Header().Set(ContentLengthHeader, util.Uint64ToString(objectInfo.GetPayloadSize()))
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

	reqCtx, err = NewRequestContext(r, g)
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
	jobStateDescription := servicetypes.StateToDescription(servicetypes.JobState(jobState))

	var xmlInfo = struct {
		XMLName             xml.Name `xml:"QueryUploadProgress"`
		Version             string   `xml:"version,attr"`
		ProgressDescription string   `xml:"ProgressDescription"`
	}{
		Version:             GnfdResponseXMLVersion,
		ProgressDescription: jobStateDescription,
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
		err               error
		reqCtx            *RequestContext
		authorized        bool
		isRange           bool
		rangeStart        int64
		rangeEnd          int64
		redirectUrl       string
		params            *storagetypes.Params
		escapedObjectName string
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
		err = getBucketInfoErr
		return
	}

	bucketPrimarySpAddress := getBucketInfoRes.GetBucketInfo().GetPrimarySpAddress()

	// if bucket not in the current sp, 302 redirect to the sp that contains the bucket
	if bucketPrimarySpAddress != g.baseApp.OperateAddress() {
		log.Debugw("primary sp address not matched ",
			"bucketPrimarySpAddress", bucketPrimarySpAddress, "gateway.config.SpOperatorAddress", g.baseApp.OperateAddress(),
		)

		endpoint, getEndpointErr := g.baseApp.GfSpClient().GetEndpointBySpAddress(reqCtx.Context(), bucketPrimarySpAddress)

		if getEndpointErr != nil || endpoint == "" {
			log.Errorw("failed to get endpoint by address ", "sp_address", reqCtx.bucketName, "error", getEndpointErr)
			err = getEndpointErr
			return
		}

		redirectUrl = endpoint + r.RequestURI
		log.Debugw("getting redirect url:", "redirectUrl", redirectUrl)

		http.Redirect(w, r, r.URL.String(), 302)
		return
	}

	getObjectInfoRes, err := g.baseApp.GfSpClient().GetObjectMeta(reqCtx.Context(), escapedObjectName, reqCtx.bucketName, true)
	if err != nil || getObjectInfoRes == nil || getObjectInfoRes.GetObjectInfo() == nil {
		log.Errorw("failed to check object meta", "object_name", escapedObjectName, "error", err)
		return
	}

	if getObjectInfoRes.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.Errorw("object is not sealed",
			"status", getObjectInfoRes.GetObjectInfo().GetObjectStatus())
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
			signedContentTemplate := `Sign this message to access the file:
%s
This signature will not cost you any fees.
Expiration Time: %s`

			signedMsg := fmt.Sprintf(signedContentTemplate, "gnfd://"+getBucketInfoRes.GetBucketInfo().BucketName+"/"+getObjectInfoRes.GetObjectInfo().GetObjectName(), expiry)
			accAddress, verifySigErr := VerifyPersonalSignature(signedMsg, signature)
			if verifySigErr != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify signature", "error", verifySigErr)
				err = verifySigErr
				return
			}
			reqCtx.account = accAddress.String()

			// 2. check permission
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

		} else {
			// return a built-in dapp for users to make the signature
			var htmlConfigMap = map[string]string{
				"9000": "{\n  \"envType\": \"qa\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 9000,\n  \"chainName\": \"qa - greenfield\",\n  \"rpcUrls\": [\"https://gnfd.qa.bnbchain.world\"],\n  \"nativeCurrency\": { \"name\": \"BNB\", \"symbol\": \"BNB\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://greenfieldscan-qanet.fe.nodereal.cc/\"]\n}\n",
				"5600": "{\n  \"envType\": \"testnet\",\n  \"signedMsg\": \"Sign this message to access the file:\\n$1\\nThis signature will not cost you any fees.\\nExpiration Time: $2\",\n  \"chainId\": 5600,\n  \"chainName\": \"greenfield testnet\",\n  \"rpcUrls\": [\"https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org\"],\n  \"nativeCurrency\": { \"name\": \"BNB\", \"symbol\": \"BNB\", \"decimals\": 18 },\n  \"blockExplorerUrls\": [\"https://greenfieldscan.com/\"]\n}\n",
			}

			htmlConfig := htmlConfigMap[g.baseApp.ChainID()]
			if htmlConfig == "" {
				log.CtxErrorw(reqCtx.Context(), "chain id is not found", "chain id ", g.baseApp.ChainID())
				err = gfsperrors.MakeGfSpError(fmt.Errorf("chain id is not found"))
				return
			}
			html := "<!DOCTYPE html>\n<html lang=\"en\">\n\n<head>\n  <meta charset=\"UTF-8\" />\n  <title>BNB Greenfield</title>\n  <link rel=\"icon\" type=\"image/svg+xml\" sizes=\"14x16\" href=\"data:image/svg+xml,%3Csvg width='14' height='16' viewBox='0 0 14 16' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M2.66928 2.45073L6.91513 0L11.161 2.45073L9.6 3.3561L6.91513 1.81073L4.23026 3.3561L2.66928 2.45073ZM11.161 5.54146L9.6 4.6361L6.91513 6.18147L4.23026 4.6361L2.66928 5.54146V7.3522L5.35415 8.89753V11.9883L6.91513 12.8937L8.47613 11.9883V8.89753L11.161 7.3522V5.54146ZM11.161 10.4429V8.6322L9.6 9.53753V11.3483L11.161 10.4429ZM12.2693 11.0829L9.5844 12.6283V14.439L13.8303 11.9883V7.0868L12.2693 7.9922V11.0829ZM10.7083 3.9961L12.2693 4.90146V6.7122L13.8303 5.80683V3.9961L12.2693 3.09073L10.7083 3.9961ZM5.35415 13.2839V15.0947L6.91513 16L8.47613 15.0947V13.2839L6.91513 14.1893L5.35415 13.2839ZM2.66928 10.4429L4.23026 11.3483V9.53753L2.66928 8.6322V10.4429ZM5.35415 3.9961L6.91513 4.90146L8.47613 3.9961L6.91513 3.09073L5.35415 3.9961ZM1.56097 4.90146L3.12195 3.9961L1.56097 3.09073L0 3.9961V5.80683L1.56097 6.7122V4.90146ZM1.56097 7.9922L0 7.0868V11.9883L4.24585 14.439V12.6283L1.56097 11.0829V7.9922Z' fill='%23F0B90B' /%3E%3C/svg%3E\" />\n  <meta name=\"viewport\" content=\"width=device-width,initial-scale=1.0,maximum-scale=1.0,minimum-scale=1.0,user-scalable=no,viewport-fit=true\" />\n  <link rel=\"preconnect\" href=\"https://fonts.gstatic.com\" crossOrigin=\"anonymous\" />\n  <link href=\"https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=Poppins:wght@400;500;600&display=swap\" rel=\"stylesheet\" />\n  <script crossorigin src=\"https://unpkg.com/react@18/umd/react.production.min.js\"></script>\n  <script crossorigin src=\"https://unpkg.com/react-dom@18/umd/react-dom.production.min.js\"></script>\n  <script>\n    try {\n      window.envConfig = JSON.parse(`<% env %>`)\n    } catch (err) {\n      console.error('Parse config json error, error detail: ', err)\n    }\n  </script>\n  <script type=\"module\" crossorigin>\n(function(){const t=document.createElement(\"link\").relList;if(t&&t.supports&&t.supports(\"modulepreload\"))return;for(const l of document.querySelectorAll('link[rel=\"modulepreload\"]'))n(l);new MutationObserver(l=>{for(const o of l)if(o.type===\"childList\")for(const i of o.addedNodes)i.tagName===\"LINK\"&&i.rel===\"modulepreload\"&&n(i)}).observe(document,{childList:!0,subtree:!0});function a(l){const o={};return l.integrity&&(o.integrity=l.integrity),l.referrerPolicy&&(o.referrerPolicy=l.referrerPolicy),l.crossOrigin===\"use-credentials\"?o.credentials=\"include\":l.crossOrigin===\"anonymous\"?o.credentials=\"omit\":o.credentials=\"same-origin\",o}function n(l){if(l.ep)return;l.ep=!0;const o=a(l);fetch(l.href,o)}})();var j={exports:{}},Z={};const T=React;/**\n * @license React\n * react-jsx-runtime.production.min.js\n *\n * Copyright (c) Facebook, Inc. and its affiliates.\n *\n * This source code is licensed under the MIT license found in the\n * LICENSE file in the root directory of this source tree.\n */var W=T,U=Symbol.for(\"react.element\"),$=Symbol.for(\"react.fragment\"),G=Object.prototype.hasOwnProperty,q=W.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.ReactCurrentOwner,z={key:!0,ref:!0,__self:!0,__source:!0};function N(e,t,a){var n,l={},o=null,i=null;a!==void 0&&(o=\"\"+a),t.key!==void 0&&(o=\"\"+t.key),t.ref!==void 0&&(i=t.ref);for(n in t)G.call(t,n)&&!z.hasOwnProperty(n)&&(l[n]=t[n]);if(e&&e.defaultProps)for(n in t=e.defaultProps,t)l[n]===void 0&&(l[n]=t[n]);return{$$typeof:U,type:e,key:o,ref:i,props:l,_owner:q.current}}Z.Fragment=$;Z.jsx=N;Z.jsxs=N;j.exports=Z;var c=j.exports,B={};const K=ReactDOM;var k=K;B.createRoot=k.createRoot,B.hydrateRoot=k.hydrateRoot;const J=\"_root_1kufm_1\",X=\"_container_1kufm_13\",Y=\"_icon_1kufm_23\",Q=\"_description_1kufm_29\",v={root:J,container:X,icon:Y,description:Q},e1=e=>React.createElement(\"svg\",{width:20,height:20,viewBox:\"0 0 20 20\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"path\",{d:\"M18.3333 10.0001C18.3333 14.6025 14.6024 18.3334 9.99999 18.3334C5.39762 18.3334 1.66666 14.6025 1.66666 10.0001C1.66666 5.39771 5.39762 1.66675 9.99999 1.66675C14.6024 1.66675 18.3333 5.39771 18.3333 10.0001Z\",fill:\"#D9304E\"}),React.createElement(\"path\",{fillRule:\"evenodd\",clipRule:\"evenodd\",d:\"M13.5503 7.41394C13.8166 7.14767 13.8166 6.71597 13.5503 6.4497C13.284 6.18343 12.8523 6.18343 12.5861 6.4497L10 8.90372L7.41394 6.4497C7.14767 6.18343 6.71597 6.18343 6.4497 6.4497C6.18343 6.71597 6.18343 7.14767 6.4497 7.41394L8.90372 10L6.4497 12.5861C6.18343 12.8523 6.18343 13.284 6.4497 13.5503C6.71597 13.8166 7.14767 13.8166 7.41394 13.5503L10 11.0963L12.5861 13.5503C12.8523 13.8166 13.284 13.8166 13.5503 13.5503C13.8166 13.284 13.8166 12.8523 13.5503 12.5861L11.0963 10L13.5503 7.41394Z\",fill:\"white\"})),t1=e=>React.createElement(\"svg\",{width:20,height:20,viewBox:\"0 0 20 20\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"path\",{d:\"M18.3333 10.0001C18.3333 14.6025 14.6024 18.3334 9.99999 18.3334C5.39762 18.3334 1.66666 14.6025 1.66666 10.0001C1.66666 5.39771 5.39762 1.66675 9.99999 1.66675C14.6024 1.66675 18.3333 5.39771 18.3333 10.0001Z\",fill:\"#5F8BFF\"}),React.createElement(\"circle\",{cx:10,cy:5.41675,r:1.25,fill:\"white\"}),React.createElement(\"path\",{d:\"M10.579 7.55143L7.87065 7.8755C7.65964 7.89895 7.5 8.07731 7.5 8.28962V8.75002C7.5 8.98014 7.68931 9.15814 7.91392 9.20817C8.47131 9.33231 8.95833 9.73515 8.95833 10.4167V12.9167C8.95833 13.5982 8.47131 14.0011 7.91392 14.1252C7.68931 14.1752 7.5 14.3532 7.5 14.5834V15C7.5 15.2301 7.68655 15.4167 7.91667 15.4167H12.0833C12.3135 15.4167 12.5 15.2301 12.5 15V14.5834C12.5 14.3532 12.3107 14.1752 12.0861 14.1252C11.5287 14.0011 11.0417 13.5982 11.0417 12.9167V7.96525C11.0417 7.71691 10.8258 7.52401 10.579 7.55143Z\",fill:\"white\"})),x={listeners:[],toastList:[],autoIncreaseId:0,add(e){const t=this.autoIncreaseId++;return this.toastList.push({toastId:t,...e}),this.notify(),t},remove(e){const t=this.toastList.findIndex(a=>a.toastId===e);t>-1&&(this.toastList.splice(t,1),this.notify())},notify(){this.listeners.forEach(e=>{e([...this.toastList])})},subscribe(e){this.listeners.push(e)},unsubscribe(e){const t=this.listeners.findIndex(a=>a===e);t>-1&&this.listeners.splice(t,1)}},n1={info:c.jsx(t1,{}),error:c.jsx(e1,{})};function a1(e){const{variant:t=\"info\",description:a,duration:n,toastId:l}=e;return React.useEffect(()=>{const o=setTimeout(()=>{x.remove(l)},n);return()=>{clearTimeout(o)}},[n,l]),c.jsxs(\"div\",{className:v.container,children:[c.jsx(\"div\",{className:v.icon,children:n1[t]}),c.jsx(\"div\",{className:v.description,children:a})]})}const P=()=>{const[e,t]=React.useState([]);return React.useEffect(()=>{const a=n=>{t(n)};return x.subscribe(a),()=>{x.unsubscribe(a)}},[]),c.jsx(c.Fragment,{children:e.length>0&&ReactDOM.createPortal(c.jsx(\"div\",{className:v.root,children:e.map(a=>c.jsx(a1,{...a},a.toastId))}),document.body)})};P.displayName=\"ToastProvider\";const D=React.createContext(void 0);function c1(e,t){switch(t.type){case\"walletUnavailable\":return{chainId:null,account:null,status:\"unavailable\",reason:\"\"};case\"walletNotConnected\":return{chainId:t.payload.chainId,account:null,status:\"notConnected\",reason:\"\"};case\"walletConnected\":const a=t.payload.accounts;return{chainId:t.payload.chainId,account:a[0],status:\"connected\",reason:\"\"};case\"walletConnecting\":return e.status===\"initializing\"||e.status===\"unavailable\"?e:{...e,account:null,status:\"connecting\",reason:\"\"};case\"walletPermissionRejected\":return e.status===\"initializing\"||e.status===\"unavailable\"?e:{...e,account:null,status:\"notConnected\",reason:\"rejected\"};case\"walletAccountsChanged\":if(e.status!==\"connected\")return e;const n=t.payload;return n.length===0?{...e,account:null,status:\"notConnected\",reason:\"\"}:{...e,account:n[0],reason:\"\"};case\"walletChainChanged\":return e.status===\"initializing\"||e.status===\"unavailable\"?e:{...e,chainId:t.payload,reason:\"\"};case\"walletSwitch\":return{status:\"initializing\",account:null,chainId:null,reason:\"\"}}}const l1=typeof window<\"u\"?React.useLayoutEffect:React.useEffect;function o1(e){const t=React.useRef(!1);return l1(()=>(t.current=!0,()=>{t.current=!1}),[]),React.useCallback(n=>{t.current&&e(n)},[e])}async function s1(e,t){if(!e){t({type:\"walletUnavailable\"});return}const a=await e.request({method:\"eth_chainId\"}),n=await e.request({method:\"eth_accounts\"});n.length===0?t({type:\"walletNotConnected\",payload:{chainId:a}}):t({type:\"walletConnected\",payload:{accounts:n,chainId:a}})}function i1(e,t){const a=async n=>{if(n.length===0)return;const l=await e.request({method:\"eth_chainId\"});t({type:\"walletConnected\",payload:{accounts:n,chainId:l}})};return e==null||e.on(\"accountsChanged\",a),()=>{e==null||e.off(\"accountsChanged\",a)}}function r1(e,t){const a=n=>t({type:\"walletAccountsChanged\",payload:n});return e==null||e.on(\"accountsChanged\",a),()=>{e==null||e.off(\"accountsChanged\",a)}}function d1(e,t){const a=n=>t({type:\"walletChainChanged\",payload:n});return e.on(\"chainChanged\",a),()=>{e.off(\"chainChanged\",a)}}function L1(e,t){return t({type:\"walletConnecting\"}),new Promise((a,n)=>{e.request({method:\"eth_requestAccounts\"}).then(async l=>{const o=await e.request({method:\"eth_chainId\"});t({type:\"walletConnected\",payload:{accounts:l,chainId:o}}),a(l)}).catch(l=>{t({type:\"walletPermissionRejected\"}),n(l)})})}async function u1(e,t){await e.request({method:\"wallet_addEthereumChain\",params:[t]})}async function f1(e,t){await e.request({method:\"wallet_switchEthereumChain\",params:[{chainId:t}]})}const h1={status:\"initializing\",account:null,chainId:null,reason:\"\"};function m1(e){const[t,a]=React.useState(),[n,l]=React.useReducer(c1,h1),o=o1(l),{status:i}=n;React.useEffect(()=>{o({type:\"walletSwitch\"})},[o,t]);const r=t&&i===\"initializing\";React.useEffect(()=>{r&&s1(t,o)},[t,o,r]);const L=t&&i===\"connected\";React.useEffect(()=>L?r1(t,o):()=>{},[t,o,L]);const d=t&&i!==\"unavailable\"&&i!==\"initializing\";React.useEffect(()=>d?d1(t,o):()=>{},[t,o,d]);const m=t&&i===\"notConnected\";React.useEffect(()=>m?i1(t,o):()=>{},[t,o,m]);const u=React.useCallback(()=>d?L1(t,o):(console.warn(\"`enable` method has been called while Wallet is not available or synchronising. Nothing will be done in this case.\"),Promise.resolve([])),[t,o,d]),C=React.useCallback(f=>d?u1(t,f):(console.warn(\"`addChain` method has been called while Wallet is not available or synchronising. Nothing will be done in this case.\"),Promise.resolve()),[t,d]),s=React.useCallback(f=>d?f1(t,f):(console.warn(\"`switchChain` method has been called while Wallet is not available or synchronising. Nothing will be done in this case.\"),Promise.resolve()),[t,d]),E=React.useMemo(()=>({...n,connect:u,addChain:C,switchChain:s,setProvider:a,provider:t}),[n,u,C,s,t]);return c.jsx(D.Provider,{value:E,...e})}const C1={Second:1e3,Minute:60*1e3,Hour:60*60*1e3,Day:24*60*60*1e3},y={Initializing:\"initializing\",Unavailable:\"unavailable\",NotConnected:\"notConnected\",Connecting:\"connecting\",Connected:\"connected\"},b={Error:\"error\",Processing:\"processing\",Connected:\"connected\",Signed:\"signed\"},S={UseRejectStatus:4001,RequestPending:-32002,NoChain:4902},g1={PermissionRejected:\"rejected\"},g={Unknown:\"Oops, met some errors, please try again.\",Connect:\"Oops, connect wallet met error, please try again.\",Switch:\"Oops, switch network met error, please try again.\",NotInstall:\"Wallet not installed. Please install and reconnect.\",Pending:\"Oops, connect wallet action is pending, please confirm it in wallet extension that you are using.\"},p1=3e3,V=\"Trust Wallet\",H=\"MetaMask\";function w1(e){const t=React.useContext(D);if(!t)throw new Error(\"`useWallet` should be used within a `WalletProvider`\");return React.useEffect(()=>{let a=null;switch(e){case V:a=window.trustwallet;break;case H:a=R1(window.ethereum);break}t.setProvider(a)},[e]),t}function R1(e){if(!e)return null;if(Array.isArray(e.providers)){const t=e.providers.find(n=>n.isMetaMask&&!n.isBraveWallet);if(t)return t;const a=e.providers.find(n=>n.isMetaMask&&n.isBraveWallet);return a||null}return e.isMetaMask?(e.off||(e.off=e.removeListener),e):null}const _1=e=>React.createElement(\"svg\",{width:403,height:48,viewBox:\"0 0 403 48\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"g\",{clipPath:\"url(#clip0_4048_116094)\"},React.createElement(\"path\",{d:\"M8.02903 7.35219L20.8003 0L33.5715 7.35219L28.8762 10.0683L20.8003 5.4322L12.7243 10.0683L8.02903 7.35219ZM33.5715 16.6244L28.8762 13.9083L20.8003 18.5444L12.7243 13.9083L8.02903 16.6244V22.0566L16.1049 26.6927V35.9649L20.8003 38.681L25.4956 35.9649V26.6927L33.5715 22.0566V16.6244ZM33.5715 31.3288V25.8966L28.8762 28.6127V34.0449L33.5715 31.3288ZM36.9052 33.2488L28.8293 37.8849V43.3171L41.6005 35.9649V21.2605L36.9052 23.9766V33.2488ZM32.2099 11.9883L36.9052 14.7044V20.1366L41.6005 17.4205V11.9883L36.9052 9.2722L32.2099 11.9883ZM16.1049 39.8517V45.2839L20.8003 48L25.4956 45.2839V39.8517L20.8003 42.5678L16.1049 39.8517ZM8.02903 31.3288L12.7243 34.0449V28.6127L8.02903 25.8966V31.3288ZM16.1049 11.9883L20.8003 14.7044L25.4956 11.9883L20.8003 9.2722L16.1049 11.9883ZM4.69531 14.7044L9.39062 11.9883L4.69531 9.2722L0 11.9883V17.4205L4.69531 20.1366V14.7044ZM4.69531 23.9766L0 21.2605V35.9649L12.7712 43.3171V37.8849L4.69531 33.2488V23.9766Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M78.9043 30.6739V30.5976C78.9043 27.0115 76.9968 25.2185 73.9067 24.0358C75.8142 22.9676 77.4165 21.289 77.4165 18.2752V18.1989C77.4165 14.0024 74.0593 11.2937 68.6038 11.2937H56.2051V37.9987H68.909C74.9367 37.9987 78.9043 35.5571 78.9043 30.6739ZM71.5795 19.2671C71.5795 21.2509 69.9391 22.0902 67.3449 22.0902H61.9276V16.444H67.7264C70.2061 16.444 71.5795 17.4359 71.5795 19.1908V19.2671ZM73.0674 29.9491C73.0674 31.9329 71.5032 32.8485 68.909 32.8485H61.9276V26.9734H68.7183C71.7321 26.9734 73.0674 28.0797 73.0674 29.8728V29.9491Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M107.848 37.9987V11.2937H102.049V27.7364L89.536 11.2937H84.1187V37.9987H89.9175V21.022L102.85 37.9987H107.848Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M137.339 30.6739V30.5976C137.339 27.0115 135.432 25.2185 132.342 24.0358C134.249 22.9676 135.852 21.289 135.852 18.2752V18.1989C135.852 14.0024 132.494 11.2937 127.039 11.2937H114.64V37.9987H127.344C133.372 37.9987 137.339 35.5571 137.339 30.6739ZM130.015 19.2671C130.015 21.2509 128.374 22.0902 125.78 22.0902H120.363V16.444H126.161C128.641 16.444 130.015 17.4359 130.015 19.1908V19.2671ZM131.502 29.9491C131.502 31.9329 129.938 32.8485 127.344 32.8485H120.363V26.9734H127.153C130.167 26.9734 131.502 28.0797 131.502 29.8728V29.9491Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M173.448 34.2982V23.0821H162.079V27.1641H168.908V32.1618C167.267 33.3826 164.978 34.1837 162.461 34.1837C157.005 34.1837 153.381 30.1398 153.381 24.6081C153.381 19.4197 157.12 15.1469 162.041 15.1469C165.436 15.1469 167.458 16.2532 169.518 18.0081L172.494 14.4602C169.747 12.133 166.886 10.8359 162.232 10.8359C154.182 10.8359 148.46 17.1688 148.46 24.6844C148.46 32.5051 153.953 38.4565 162.308 38.4565C167.039 38.4565 170.815 36.549 173.448 34.2982Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M202.601 37.9987L195.391 27.889C199.13 26.8589 201.762 24.1884 201.762 19.763C201.762 14.5365 197.985 11.2937 191.843 11.2937H179.94V37.9987H184.633V28.6901H190.508L197.069 37.9987H202.601ZM196.993 20.0301C196.993 22.7769 194.857 24.5318 191.5 24.5318H184.633V15.5665H191.461C194.933 15.5665 196.993 17.0925 196.993 20.0301Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M227.691 11.2937H207.891V37.9987H227.882V33.8022H212.583V26.63H225.974V22.4335H212.583V15.4902H227.691V11.2937Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M253.63 11.2937H233.83V37.9987H253.821V33.8022H238.522V26.63H251.913V22.4335H238.522V15.4902H253.63V11.2937Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M278.424 29.7583L264.118 11.2937H259.769V37.9987H264.385V19L279.111 37.9987H283.041V11.2937H278.424V29.7583Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M310.167 11.2937H290.291V37.9987H294.983V27.126H308.45V22.8532H294.983V15.5665H310.167V11.2937Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M315.975 11.2937V37.9987H320.667V11.2937H315.975Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M347.997 11.2937H328.198V37.9987H348.188V33.8022H332.89V26.63H346.281V22.4335H332.89V15.4902H347.997V11.2937Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M354.137 37.9987H372.868V33.7259H358.829V11.2937H354.137V37.9987Z\",fill:\"var(--color-active)\"}),React.createElement(\"path\",{d:\"M402.287 24.6081C402.287 17.0544 396.488 11.2937 388.095 11.2937H378.138V37.9987H388.095C396.488 37.9987 402.287 32.1618 402.287 24.6081ZM397.366 24.6844C397.366 29.9491 393.704 33.7259 388.095 33.7259H382.831V15.5665H388.095C393.704 15.5665 397.366 19.4197 397.366 24.6844Z\",fill:\"var(--color-active)\"})),React.createElement(\"defs\",null,React.createElement(\"clipPath\",{id:\"clip0_4048_116094\"},React.createElement(\"rect\",{width:403,height:48,fill:\"white\"})))),E1=\"_container_1cbn6_1\",M1=\"_content_1cbn6_7\",v1=\"_footer_1cbn6_16\",M={container:E1,content:M1,footer:v1,\"txt-connect-tips\":\"_txt-connect-tips_1cbn6_29\"},b1=\"_button_fyqdl_1\",x1={button:b1};function Z1(e){const{...t}=e;return c.jsx(\"button\",{className:x1.button,...t})}const y1=\"_modal_1ct5u_1\",h={modal:y1,\"modal-mask\":\"_modal-mask_1ct5u_15\",\"modal-content\":\"_modal-content_1ct5u_23\",\"modal-header\":\"_modal-header_1ct5u_32\",\"modal-body\":\"_modal-body_1ct5u_39\",\"modal-footer\":\"_modal-footer_1ct5u_42\",\"modal-btn-close\":\"_modal-btn-close_1ct5u_51\"},B1=\"_link_44242_1\",V1={link:B1};function H1(e){const{children:t,isExternal:a=!1,...n}=e;return c.jsx(\"a\",{className:V1.link,target:a?\"_blank\":void 0,rel:a?\"noopener\":void 0,...n,children:t})}const k1=e=>React.createElement(\"svg\",{width:24,height:24,viewBox:\"0 0 24 24\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"path\",{fillRule:\"evenodd\",clipRule:\"evenodd\",d:\"M7.20711 5.70711C6.81658 5.31658 6.18342 5.31658 5.79289 5.70711C5.40237 6.09763 5.40237 6.7308 5.79289 7.12132L10.5858 11.9142L5.70711 16.7929C5.31658 17.1834 5.31658 17.8166 5.70711 18.2071C6.09763 18.5976 6.7308 18.5976 7.12132 18.2071L12 13.3284L16.753 18.0815C17.1436 18.472 17.7767 18.472 18.1673 18.0815C18.5578 17.691 18.5578 17.0578 18.1673 16.6673L13.4142 11.9142L18.0815 7.24695C18.472 6.85643 18.472 6.22326 18.0815 5.83274C17.6909 5.44221 17.0578 5.44221 16.6673 5.83274L12 10.5L7.20711 5.70711Z\",fill:\"#76808F\"})),S1=e=>React.createElement(\"svg\",{width:52,height:52,viewBox:\"0 0 52 52\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"g\",{clipPath:\"url(#clip0_4138_115346)\"},React.createElement(\"path\",{fillRule:\"evenodd\",clipRule:\"evenodd\",d:\"M25.1038 9.82664C25.6315 9.38591 26.399 9.38591 26.9267 9.82664C31.5058 13.6509 36.7265 13.6204 38.7219 13.6088C38.8221 13.6082 38.9142 13.6076 38.9976 13.6076C39.3784 13.6076 39.7433 13.7604 40.0106 14.0317C40.2778 14.303 40.4251 14.6702 40.4193 15.051C40.3251 21.2927 40.075 25.7357 39.5864 28.9851C39.0981 32.2328 38.3534 34.4269 37.1725 36.068C35.9843 37.7194 34.4557 38.662 32.7882 39.5835C32.5006 39.7425 32.2067 39.9019 31.9052 40.0654C30.4375 40.8615 28.7927 41.7536 26.844 43.1515C26.3486 43.5068 25.6818 43.5068 25.1865 43.1515C23.2397 41.755 21.597 40.8633 20.1315 40.0677C19.8287 39.9033 19.5334 39.743 19.2445 39.5832C17.5784 38.6615 16.0519 37.7187 14.866 36.0668C13.6877 34.4255 12.9456 32.2316 12.4591 28.9844C11.9723 25.7353 11.7232 21.2925 11.6291 15.051C11.6233 14.6702 11.7706 14.303 12.0378 14.0317C12.3051 13.7604 12.67 13.6076 13.0508 13.6076C13.1333 13.6076 13.2244 13.6082 13.3237 13.6088C15.303 13.6204 20.5243 13.6512 25.1038 9.82664ZM14.4977 16.4475C14.6068 21.8967 14.8504 25.7528 15.2714 28.563C15.7289 31.6165 16.3779 33.2965 17.1761 34.4084C17.9667 35.5096 19.0019 36.1991 20.6211 37.0949C20.8861 37.2415 21.165 37.3927 21.4573 37.5512C22.7316 38.2423 24.2616 39.0721 26.0153 40.2624C27.7713 39.0706 29.3036 38.2403 30.5799 37.5487C30.8709 37.391 31.1487 37.2405 31.4127 37.0946C33.0339 36.1986 34.0715 35.5088 34.8642 34.4071C35.6643 33.2952 36.3152 31.6153 36.7743 28.5623C37.1968 25.7524 37.4414 21.8966 37.5507 16.4477C35.0368 16.3893 30.3736 15.9228 26.0152 12.7234C21.6618 15.9193 17.0048 16.3882 14.4977 16.4475Z\",fill:\"#3375BB\"})),React.createElement(\"defs\",null,React.createElement(\"clipPath\",{id:\"clip0_4138_115346\"},React.createElement(\"rect\",{width:52,height:52,fill:\"white\"})))),I1=e=>React.createElement(\"svg\",{width:52,height:52,viewBox:\"0 0 52 52\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"g\",{clipPath:\"url(#clip0_4138_115356)\"},React.createElement(\"path\",{d:\"M40.6314 10.9688L27.95 20.3874L30.2951 14.8305L40.6314 10.9688Z\",fill:\"#E2761B\"}),React.createElement(\"path\",{d:\"M11.3559 10.9688L23.9353 20.4766L21.7049 14.8305L11.3559 10.9688Z\",fill:\"#E4761B\"}),React.createElement(\"path\",{d:\"M36.0687 32.8011L32.6912 37.9756L39.9177 39.9638L41.9951 32.9158L36.0687 32.8011Z\",fill:\"#E4761B\"}),React.createElement(\"path\",{d:\"M10.0176 32.9158L12.0823 39.9638L19.3088 37.9756L15.9314 32.8011L10.0176 32.9158Z\",fill:\"#E4761B\"}),React.createElement(\"path\",{d:\"M18.901 24.058L16.8873 27.104L24.0627 27.4227L23.8078 19.7119L18.901 24.058Z\",fill:\"#E4761B\"}),React.createElement(\"path\",{d:\"M33.0863 24.058L28.1157 19.6227L27.95 27.4227L35.1127 27.104L33.0863 24.058Z\",fill:\"#E4761B\"}),React.createElement(\"path\",{d:\"M19.3088 37.9756L23.6167 35.8727L19.8951 32.9668L19.3088 37.9756Z\",fill:\"#E4761B\"}),React.createElement(\"path\",{d:\"M28.3706 35.8727L32.6912 37.9756L32.0922 32.9668L28.3706 35.8727Z\",fill:\"#E4761B\"}),React.createElement(\"path\",{d:\"M32.6912 37.9756L28.3706 35.8727L28.7147 38.6893L28.6765 39.8746L32.6912 37.9756Z\",fill:\"#D7C1B3\"}),React.createElement(\"path\",{d:\"M19.3088 37.9756L23.3235 39.8746L23.298 38.6893L23.6167 35.8727L19.3088 37.9756Z\",fill:\"#D7C1B3\"}),React.createElement(\"path\",{d:\"M23.3873 31.106L19.7931 30.0482L22.3294 28.8884L23.3873 31.106Z\",fill:\"#233447\"}),React.createElement(\"path\",{d:\"M28.6 31.106L29.6578 28.8884L32.2068 30.0482L28.6 31.106Z\",fill:\"#233447\"}),React.createElement(\"path\",{d:\"M19.3088 37.9756L19.9206 32.8011L15.9314 32.9158L19.3088 37.9756Z\",fill:\"#CD6116\"}),React.createElement(\"path\",{d:\"M32.0794 32.8011L32.6912 37.9756L36.0686 32.9158L32.0794 32.8011Z\",fill:\"#CD6116\"}),React.createElement(\"path\",{d:\"M35.1127 27.104L27.95 27.4227L28.6128 31.106L29.6706 28.8884L32.2196 30.0482L35.1127 27.104Z\",fill:\"#CD6116\"}),React.createElement(\"path\",{d:\"M19.7931 30.0482L22.3422 28.8884L23.3873 31.106L24.0627 27.4227L16.8873 27.104L19.7931 30.0482Z\",fill:\"#CD6116\"}),React.createElement(\"path\",{d:\"M16.8873 27.104L19.8951 32.9668L19.7931 30.0482L16.8873 27.104Z\",fill:\"#E4751F\"}),React.createElement(\"path\",{d:\"M32.2196 30.0482L32.0922 32.9668L35.1127 27.104L32.2196 30.0482Z\",fill:\"#E4751F\"}),React.createElement(\"path\",{d:\"M24.0627 27.4227L23.3873 31.106L24.2284 35.4521L24.4196 29.7295L24.0627 27.4227Z\",fill:\"#E4751F\"}),React.createElement(\"path\",{d:\"M27.95 27.4227L27.6059 29.7168L27.7588 35.4521L28.6128 31.106L27.95 27.4227Z\",fill:\"#E4751F\"}),React.createElement(\"path\",{d:\"M28.6128 31.106L27.7588 35.4521L28.3706 35.8727L32.0922 32.9668L32.2196 30.0482L28.6128 31.106Z\",fill:\"#F6851B\"}),React.createElement(\"path\",{d:\"M19.7931 30.0482L19.8951 32.9668L23.6167 35.8727L24.2284 35.4521L23.3873 31.106L19.7931 30.0482Z\",fill:\"#F6851B\"}),React.createElement(\"path\",{d:\"M28.6765 39.8746L28.7147 38.6893L28.3961 38.4089H23.5912L23.298 38.6893L23.3235 39.8746L19.3088 37.9756L20.7108 39.1227L23.5529 41.0982H28.4343L31.2892 39.1227L32.6912 37.9756L28.6765 39.8746Z\",fill:\"#C0AD9E\"}),React.createElement(\"path\",{d:\"M28.3706 35.8727L27.7588 35.4521H24.2284L23.6167 35.8727L23.298 38.6893L23.5912 38.4089H28.3961L28.7147 38.6893L28.3706 35.8727Z\",fill:\"#161616\"}),React.createElement(\"path\",{d:\"M41.1667 20.9991L42.25 15.7991L40.6314 10.9688L28.3706 20.0688L33.0863 24.058L39.752 26.008L41.2304 24.2874L40.5931 23.8286L41.6127 22.8982L40.8225 22.2864L41.8422 21.5089L41.1667 20.9991Z\",fill:\"#763D16\"}),React.createElement(\"path\",{d:\"M9.75 15.7991L10.8333 20.9991L10.1451 21.5089L11.1647 22.2864L10.3873 22.8982L11.4069 23.8286L10.7696 24.2874L12.2353 26.008L18.901 24.058L23.6167 20.0688L11.3559 10.9688L9.75 15.7991Z\",fill:\"#763D16\"}),React.createElement(\"path\",{d:\"M39.752 26.008L33.0863 24.058L35.1127 27.104L32.0922 32.9668L36.0686 32.9158H41.9951L39.752 26.008Z\",fill:\"#F6851B\"}),React.createElement(\"path\",{d:\"M18.901 24.058L12.2353 26.008L10.0176 32.9158H15.9314L19.8951 32.9668L16.8873 27.104L18.901 24.058Z\",fill:\"#F6851B\"}),React.createElement(\"path\",{d:\"M27.95 27.4227L28.3706 20.0688L30.3078 14.8305L21.7049 14.8305L23.6167 20.0688L24.0627 27.4227L24.2157 29.7423L24.2284 35.4521H27.7588L27.7843 29.7423L27.95 27.4227Z\",fill:\"#F6851B\"}),React.createElement(\"path\",{fillRule:\"evenodd\",clipRule:\"evenodd\",d:\"M11.2682 10.7855C11.318 10.7617 11.3752 10.7592 11.4269 10.7784L21.7416 14.6274L30.2584 14.6274L40.5603 10.7785C40.6119 10.7592 40.6691 10.7617 40.7188 10.7854C40.7685 10.8091 40.8065 10.852 40.824 10.9042L42.4426 15.7346C42.454 15.7687 42.4562 15.8053 42.4489 15.8406L41.3918 20.9146L41.9645 21.3468C42.0152 21.3851 42.0451 21.4449 42.0453 21.5084C42.0454 21.572 42.0159 21.6319 41.9653 21.6705L41.1559 22.2876L41.7371 22.7375C41.7845 22.7743 41.8133 22.83 41.8157 22.89C41.8181 22.9499 41.794 23.0078 41.7497 23.0482L40.9149 23.8099L41.3491 24.1225C41.3953 24.1558 41.4254 24.2069 41.4321 24.2634C41.4388 24.3199 41.4215 24.3766 41.3845 24.4198L39.9804 26.0538L42.1883 32.8531C42.201 32.892 42.2015 32.9339 42.19 32.9732L40.1125 40.0213C40.0812 40.1275 39.9706 40.1891 39.8638 40.1597L32.7386 38.1993L31.4178 39.2799C31.4136 39.2833 31.4093 39.2866 31.4048 39.2897L28.5499 41.2652C28.5159 41.2887 28.4756 41.3013 28.4343 41.3013H23.5529C23.5115 41.3013 23.471 41.2886 23.437 41.265L20.5949 39.2895C20.5905 39.2864 20.5863 39.2832 20.5822 39.2799L19.2615 38.1993L12.1362 40.1597C12.0293 40.1891 11.9186 40.1274 11.8874 40.021L9.8227 32.9729C9.81128 32.9339 9.81182 32.8924 9.82423 32.8537L12.0073 26.0536L10.615 24.4191C10.5782 24.3759 10.5611 24.3193 10.568 24.2629C10.5748 24.2066 10.6049 24.1557 10.6509 24.1225L11.0851 23.8099L10.2503 23.0482C10.2063 23.008 10.1821 22.9504 10.1843 22.8908C10.1864 22.8311 10.2147 22.7754 10.2616 22.7385L10.833 22.2889L10.0219 21.6705C9.97115 21.6317 9.94153 21.5714 9.94198 21.5075C9.94242 21.4437 9.97287 21.3837 10.0242 21.3457L10.608 20.9133L9.55114 15.8406C9.54383 15.8055 9.54594 15.7691 9.55725 15.7351L11.1631 10.9047C11.1805 10.8523 11.2185 10.8093 11.2682 10.7855ZM12.3958 26.1727L10.2979 32.7072L15.9274 32.598C15.9974 32.5967 16.0632 32.6314 16.1015 32.6901L16.1128 32.7074L19.484 32.6105L16.7065 27.1968C16.673 27.1315 16.6774 27.0532 16.7178 26.992L18.4256 24.4087L12.3958 26.1727ZM19.0544 24.1943L17.2544 26.917L23.8463 27.2097L23.4385 20.4856L19.0544 24.1943ZM23.8209 27.6153L17.3956 27.3299L19.8379 29.8043L22.2449 28.7036C22.2744 28.6902 22.3054 28.6843 22.3357 28.6853C22.4143 28.6828 22.4904 28.7263 22.5259 28.8018L23.3015 30.4475L23.8209 27.6153ZM23.008 30.7826L22.2342 29.1606L20.3735 30.0072L23.008 30.7826ZM20.006 30.3225L20.0886 32.6869C20.1159 32.7271 20.1281 32.7761 20.1223 32.825L20.1157 32.8813L23.6243 35.6209L24.0037 35.3601L23.2113 31.266L20.006 30.3225ZM23.594 31.105L24.0205 33.3089L24.0126 29.7492L23.9676 29.0675L23.594 31.105ZM24.2915 35.6552L23.8081 35.9876L23.5568 38.2087C23.5681 38.2068 23.5796 38.2058 23.5912 38.2058H28.3961C28.4151 38.2058 28.4339 38.2085 28.452 38.2137L28.1801 35.9882L27.6957 35.6552H24.2915ZM27.9838 35.3603L28.363 35.6209L31.8833 32.8722L31.8777 32.825C31.8728 32.7833 31.8809 32.7416 31.9003 32.7054L32.0043 30.323L28.7883 31.2662L27.9838 35.3603ZM32.4192 30.1349L32.3355 32.051L34.3185 28.2021L32.4192 30.1349ZM34.7444 26.9171L32.9331 24.1944L28.5501 20.4867L28.1656 27.2098L34.7444 26.9171ZM33.1873 23.8759L39.6833 25.7762L40.9321 24.3229L40.4744 23.9934C40.4248 23.9576 40.3939 23.9014 40.3904 23.8403C40.3868 23.7792 40.411 23.7198 40.4562 23.6785L41.2974 22.9109L40.6982 22.447C40.6483 22.4084 40.6192 22.3488 40.6194 22.2857C40.6197 22.2225 40.6492 22.1631 40.6994 22.1249L41.5061 21.5098L41.0443 21.1613C40.9816 21.1139 40.9518 21.0347 40.9678 20.9577L42.0399 15.8115L40.5277 11.2988L28.8294 19.9873L33.1873 23.8759ZM39.3568 11.6618L30.4681 14.9827L28.798 19.4985L39.3568 11.6618ZM29.9889 15.0336L22.0036 15.0336L23.7726 19.5118C23.8109 19.5051 23.851 19.5093 23.8882 19.5253C23.9604 19.5564 24.0083 19.6265 24.0109 19.7052L24.0256 20.1524L24.1242 20.402C24.1577 20.4866 24.1307 20.5832 24.0582 20.6383C24.053 20.6423 24.0476 20.646 24.0421 20.6494L24.2654 27.4037L24.6203 29.6985C24.6223 29.711 24.623 29.7237 24.6226 29.7363L24.4385 35.249H27.5502L27.4028 29.7222C27.4025 29.7103 27.4032 29.6984 27.405 29.6867L27.7472 27.4054L27.8921 20.5821C27.8681 20.575 27.8452 20.5633 27.8246 20.5471C27.7525 20.4906 27.7273 20.3928 27.7629 20.3084L27.9051 19.9714L27.9126 19.6184C27.9143 19.5391 27.9619 19.4681 28.0346 19.4364C28.0676 19.422 28.1031 19.417 28.1375 19.4207L29.9889 15.0336ZM27.9715 33.3162L28.3994 31.1382C28.3958 31.1159 28.396 31.093 28.4 31.0705L28.0371 29.0539L27.9874 29.75L27.9715 33.3162ZM28.191 27.6153L28.6975 30.4298L29.4745 28.8009C29.5103 28.7259 29.5862 28.6828 29.6644 28.6853C29.6946 28.6843 29.7254 28.6901 29.7547 28.7035L32.1744 29.8044L34.606 27.3299L28.191 27.6153ZM31.6257 30.0069L29.7658 29.1606L28.9939 30.7788L31.6257 30.0069ZM21.5462 14.9881L12.6288 11.6605L23.1935 19.5017L21.5462 14.9881ZM11.4593 11.3015L9.96002 15.8113L11.0322 20.9577C11.0484 21.0353 11.0179 21.1152 10.9542 21.1624L10.4832 21.5113L11.2879 22.1249C11.3378 22.163 11.3673 22.222 11.3678 22.2848C11.3683 22.3477 11.3397 22.4072 11.2903 22.446L10.7012 22.9096L11.5438 23.6785C11.589 23.7198 11.6132 23.7792 11.6096 23.8403C11.6061 23.9014 11.5752 23.9576 11.5256 23.9934L11.0672 24.3234L12.3046 25.7761L18.8003 23.8758L23.0803 20.085L11.4593 11.3015ZM39.5919 26.1728L33.564 24.4094L35.2818 26.9915C35.3227 27.0529 35.3271 27.1316 35.2933 27.1971L32.5044 32.6101L35.8873 32.7074L35.8986 32.6901C35.9368 32.6315 36.0026 32.5967 36.0726 32.598L41.7138 32.7072L39.5919 26.1728ZM31.9333 33.3486L28.7521 35.8325L32.4453 37.63L31.9333 33.3486ZM32.2217 37.973L28.6175 36.2188L28.9164 38.6647C28.9176 38.6751 28.9181 38.6855 28.9178 38.6959L28.8902 39.5488L32.2217 37.973ZM31.2798 38.8679L28.7634 40.0583C28.6994 40.0885 28.6243 40.0833 28.5651 40.0445C28.5059 40.0057 28.4712 39.9388 28.4735 39.8681L28.5086 38.7786L28.3194 38.6121H23.6727L23.503 38.7743L23.5266 39.8703C23.5281 39.9406 23.4931 40.0068 23.434 40.0451C23.375 40.0834 23.3003 40.0884 23.2367 40.0583L20.7202 38.8679L20.8332 38.9604L23.6166 40.895H28.3709L31.1669 38.9603L31.2798 38.8679ZM19.7778 37.9727L23.1134 39.5505L23.095 38.6937C23.0948 38.6877 23.095 38.6816 23.0954 38.6756C23.0956 38.6726 23.0959 38.6695 23.0962 38.6665L23.3733 36.2175L19.7778 37.9727ZM23.2355 35.8327L19.5542 37.6298L20.0599 33.3531L23.2355 35.8327ZM19.6603 32.0642L17.6716 28.1879L19.5929 30.1345L19.6603 32.0642ZM10.2888 33.1189L12.2211 39.715L18.9837 37.8544L15.8228 33.1189H10.2888ZM19.6672 33.167L16.3852 33.1248L19.1723 37.3948L19.6672 33.167ZM33.0163 37.8544L39.7792 39.7151L41.7235 33.1189H36.1773L33.0163 37.8544ZM35.6148 33.1248L32.3272 33.1669L32.8272 37.3957L35.6148 33.1248Z\",fill:\"#F6851B\"})),React.createElement(\"defs\",null,React.createElement(\"clipPath\",{id:\"clip0_4138_115356\"},React.createElement(\"rect\",{width:52,height:52,rx:24,fill:\"white\"})))),j1=e=>\"0x\"+Number(e).toString(16);function N1(e){const t=new TextEncoder().encode(e);return Array.from(t,n=>n.toString(16).padStart(2,\"0\")).join(\"\")}const I=\"https://trustwallet.com/browser-extension\",P1=\"https://metamask.io/download/\",D1=`Sign this message to access the file:\n$1\nThis signature will not cost you any fees.\nExpiration Time: $2`,{envType:F1=\"qa\",chainId:O1=9e3,signedMsg:A1=D1,...T1}=window.envConfig||{},R={envType:F1,signedMsg:A1,greenfieldChain:{chainId:j1(O1),chainName:\"qa - greenfield\",rpcUrls:[\"https://gnfd.qa.bnbchain.world\"],nativeCurrency:{name:\"BNB\",symbol:\"BNB\",decimals:18},blockExplorerUrls:[\"https://greenfieldscan-qanet.fe.nodereal.cc/\"],...T1},trustWalletDownloadUrl:I,wallets:[{icon:c.jsx(S1,{}),name:V,downloadLink:I},{icon:c.jsx(I1,{}),name:H,downloadLink:P1}]},W1=e=>React.createElement(\"svg\",{width:26,height:26,viewBox:\"0 0 26 26\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"path\",{d:\"M23.5 13C23.5 14.3789 23.2284 15.7443 22.7007 17.0182C22.1731 18.2921 21.3996 19.4496 20.4246 20.4246C19.4496 21.3996 18.2921 22.1731 17.0182 22.7007C15.7443 23.2284 14.3789 23.5 13 23.5C11.6211 23.5 10.2557 23.2284 8.98182 22.7007C7.7079 22.1731 6.55039 21.3996 5.57538 20.4246C4.60036 19.4496 3.82694 18.2921 3.29926 17.0182C2.77159 15.7443 2.5 14.3789 2.5 13C2.5 11.6211 2.77159 10.2557 3.29927 8.98182C3.82694 7.7079 4.60037 6.55039 5.57538 5.57538C6.5504 4.60036 7.70791 3.82694 8.98183 3.29926C10.2557 2.77159 11.6211 2.5 13 2.5C14.3789 2.5 15.7443 2.77159 17.0182 3.29927C18.2921 3.82694 19.4496 4.60037 20.4246 5.57538C21.3996 6.5504 22.1731 7.70791 22.7007 8.98183C23.2284 10.2557 23.5 11.6211 23.5 13L23.5 13Z\",stroke:\"#F0B90B\",strokeOpacity:.1,strokeWidth:4,strokeLinecap:\"round\",strokeLinejoin:\"round\"}),React.createElement(\"path\",{d:\"M13 2.5C14.3789 2.5 15.7443 2.77159 17.0182 3.29927C18.2921 3.82694 19.4496 4.60036 20.4246 5.57538C21.3996 6.55039 22.1731 7.70791 22.7007 8.98183C23.2284 10.2557 23.5 11.6211 23.5 13\",stroke:\"#F0B90B\",strokeWidth:4,strokeLinecap:\"round\",strokeLinejoin:\"round\"})),U1=\"_loading_1slm0_1\",$1=\"_rotation_1slm0_1\",G1={loading:U1,rotation:$1};function q1(e){return c.jsx(W1,{className:G1.loading,...e})}const z1=\"_active_z15a5_22\",K1=\"_disabled_z15a5_25\",p={\"wallet-item\":\"_wallet-item_z15a5_1\",active:z1,disabled:K1,\"wallet-item-icon\":\"_wallet-item-icon_z15a5_29\",\"wallet-item-loading\":\"_wallet-item-loading_z15a5_34\"},_=e=>{const{variant:t=\"info\",duration:a=p1,...n}=e;return x.add({variant:t,duration:a,...n})};_.info=e=>_({variant:\"info\",...e});_.error=e=>_({variant:\"error\",...e});const J1=C1.Minute*5;function X1(e){const{reason:t,status:a,connect:n,account:l,chainId:o,provider:i,switchChain:r,addChain:L}=w1(e),[d,m]=React.useState(),u=(C,s)=>{m(b.Error),(s==null?void 0:s.code)===S.RequestPending?_.error({description:g.Pending}):_.error({description:C??g.Unknown}),s&&console.error(s)};return React.useEffect(()=>{if(!e)return;const C=async()=>{switch(!0){case a===y.Connected:if(o===R.greenfieldChain.chainId&&l)try{const s=Q1(),f=`gnfd://${window.location.pathname.replace(/^\\/(download|view)\\//,\"\")}`,F=R.signedMsg.replace(\"$1\",f).replace(\"$2\",s),O=await Y1(l,F,i);m(b.Signed);const A=`${window.location.origin}${window.location.pathname}?expiry=${s}&signature=${O}`;window.location.href=A}catch(s){u(s==null?void 0:s.message,s)}else try{await r(R.greenfieldChain.chainId)}catch(s){if(s.code===S.NoChain)try{await L(R.greenfieldChain)}catch(E){u(g.Switch,E)}else u(g.Switch,s)}break;case(a===y.NotConnected&&t!==g1.PermissionRejected):try{await n()}catch(s){u(g.Connect,s)}break;case a===y.Unavailable:u(g.NotInstall);break}};m(b.Processing),setTimeout(()=>{C()},0)},[a,o,i,e,l,r,L,n,t]),{actionStatus:d}}async function Y1(e,t,a){const n=`0x${N1(t)}`;return await a.request({method:\"personal_sign\",params:[n,e]})}function Q1(){const e=Date.now()+J1;return new Date(e).toISOString().replace(/(\\.\\d+)Z$/,\"Z\")}function e2(e){const[t,a]=React.useState(\"\"),[n,l]=React.useState(!1),{actionStatus:o}=X1(t);React.useEffect(()=>{o===b.Error&&(a(\"\"),l(!1))},[o]);const i=async r=>{if(!n){switch(r.name){case V:if(typeof window.trustwallet>\"u\"){window.open(r.downloadLink,\"_blank\",\"noopener,noreferrer\");return}break;case H:if(typeof window.ethereum>\"u\"){window.open(r.downloadLink,\"_blank\",\"noopener,noreferrer\");return}break}a(r.name),l(!0)}};return c.jsx(\"ul\",{className:p[\"wallet-list\"],...e,children:R.wallets.map(r=>{const L=t===r.name,d=[p[\"wallet-item\"],L?p.active:\"\",n?p.disabled:\"\"];return c.jsxs(\"li\",{className:d.join(\" \"),onClick:()=>i(r),children:[c.jsx(\"div\",{className:p[\"wallet-item-icon\"],children:r.icon}),r.name,c.jsx(\"div\",{className:p[\"wallet-item-loading\"],children:L&&c.jsx(q1,{})})]},r.name)})})}function t2(e){const{isOpen:t,onClose:a}=e,[n,l]=React.useState(null);React.useEffect(()=>{n&&t&&setTimeout(()=>{n.style.opacity=\"1\"},30)},[t,n]);const o=React.useCallback(()=>{n&&(n.style.opacity=\"0\"),setTimeout(()=>{l(null),a()},200)},[n,a]);return t?c.jsxs(\"section\",{ref:i=>l(i),className:h.modal,children:[c.jsx(\"div\",{className:h[\"modal-mask\"]}),c.jsxs(\"div\",{className:h[\"modal-content\"],children:[c.jsx(\"div\",{className:h[\"modal-header\"],children:\"Connect a Wallet\"}),c.jsx(\"div\",{className:h[\"modal-btn-close\"],onClick:o,children:c.jsx(k1,{})}),c.jsx(\"div\",{className:h[\"modal-body\"],children:c.jsx(e2,{onClose:o})}),c.jsxs(\"div\",{className:h[\"modal-footer\"],children:[\"Don’t have a wallet?\",c.jsx(H1,{style:{marginLeft:2},href:R.trustWalletDownloadUrl,isExternal:!0,children:\"Get one here!\"})]})]})]}):null}const n2=\"_container_1fm5k_1\",w={container:n2,\"cube-1\":\"_cube-1_1fm5k_16\",\"cube-2\":\"_cube-2_1fm5k_21\",\"cube-3\":\"_cube-3_1fm5k_26\",\"cube-4\":\"_cube-4_1fm5k_31\",\"cube-5\":\"_cube-5_1fm5k_36\"},a2=e=>React.createElement(\"svg\",{width:137,height:137,viewBox:\"0 0 137 137\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"g\",{filter:\"url(#filter0_f_4048_115770)\"},React.createElement(\"path\",{d:\"M127.711 49.032L68.0412 67.9426L23.5868 27.3898L83.0045 8.56337L127.711 49.032Z\",fill:\"#474D57\",stroke:\"#474D57\"}),React.createElement(\"path\",{d:\"M67.917 68.5064L54.7021 128.369L115.199 109.194L128.692 49.2456L67.917 68.5064Z\",fill:\"#2B2F36\"}),React.createElement(\"path\",{d:\"M8 87.4884L54.7023 128.369L67.9171 68.5064L22.6092 27.175L8 87.4884Z\",fill:\"#14151A\"})),React.createElement(\"defs\",null,React.createElement(\"filter\",{id:\"filter0_f_4048_115770\",x:0,y:0,width:136.692,height:136.369,filterUnits:\"userSpaceOnUse\",colorInterpolationFilters:\"sRGB\"},React.createElement(\"feFlood\",{floodOpacity:0,result:\"BackgroundImageFix\"}),React.createElement(\"feBlend\",{mode:\"normal\",in:\"SourceGraphic\",in2:\"BackgroundImageFix\",result:\"shape\"}),React.createElement(\"feGaussianBlur\",{stdDeviation:4,result:\"effect1_foregroundBlur_4048_115770\"})))),c2=e=>React.createElement(\"svg\",{width:110,height:106,viewBox:\"0 0 110 106\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"g\",{filter:\"url(#filter0_f_4048_115774)\"},React.createElement(\"path\",{d:\"M85.0552 17.3909L54.4888 53.3319L9.8625 45.204L40.2999 9.41347L85.0552 17.3909Z\",fill:\"#474D57\",stroke:\"#474D57\"}),React.createElement(\"path\",{d:\"M54.6831 53.8755L70.2756 97.6895L101.461 61.0198L86.0005 17.0515L54.6831 53.8755Z\",fill:\"#2B2F36\"}),React.createElement(\"path\",{d:\"M23.7846 90.2145L70.2754 97.6895L54.6829 53.8755L8.91992 45.5405L23.7846 90.2145Z\",fill:\"#14151A\"})),React.createElement(\"defs\",null,React.createElement(\"filter\",{id:\"filter0_f_4048_115774\",x:.919922,y:.87085,width:108.541,height:104.819,filterUnits:\"userSpaceOnUse\",colorInterpolationFilters:\"sRGB\"},React.createElement(\"feFlood\",{floodOpacity:0,result:\"BackgroundImageFix\"}),React.createElement(\"feBlend\",{mode:\"normal\",in:\"SourceGraphic\",in2:\"BackgroundImageFix\",result:\"shape\"}),React.createElement(\"feGaussianBlur\",{stdDeviation:4,result:\"effect1_foregroundBlur_4048_115774\"})))),l2=e=>React.createElement(\"svg\",{width:94,height:98,viewBox:\"0 0 94 98\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"g\",{filter:\"url(#filter0_f_4048_115778)\"},React.createElement(\"path\",{d:\"M66.968 3.27817L76.5958 78.2882L3.93671 45.5352L66.968 3.27817Z\",fill:\"#474D57\",stroke:\"#474D57\"}),React.createElement(\"path\",{d:\"M2.91064 45.6212L77.2056 79.1116L27.1556 95.7554L2.91064 45.6212Z\",fill:\"#14151A\"}),React.createElement(\"path\",{d:\"M91.6371 47.1031L77.3148 78.862L67.4702 2.16321L91.6371 47.1031Z\",fill:\"#2B2F36\"}),React.createElement(\"path\",{d:\"M17.3579 13.6127L3.03564 45.3716L67.4704 2.16321L17.3579 13.6127Z\",fill:\"#2B2F36\"})),React.createElement(\"defs\",null,React.createElement(\"filter\",{id:\"filter0_f_4048_115778\",x:.910645,y:.163208,width:92.7266,height:97.5922,filterUnits:\"userSpaceOnUse\",colorInterpolationFilters:\"sRGB\"},React.createElement(\"feFlood\",{floodOpacity:0,result:\"BackgroundImageFix\"}),React.createElement(\"feBlend\",{mode:\"normal\",in:\"SourceGraphic\",in2:\"BackgroundImageFix\",result:\"shape\"}),React.createElement(\"feGaussianBlur\",{stdDeviation:1,result:\"effect1_foregroundBlur_4048_115778\"})))),o2=e=>React.createElement(\"svg\",{width:57,height:52,viewBox:\"0 0 57 52\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"g\",{filter:\"url(#filter0_f_4048_115766)\"},React.createElement(\"path\",{d:\"M41.9213 4.18758L28.1691 25.7056L3.6603 24.5874L17.3649 3.1646L41.9213 4.18758Z\",fill:\"#474D57\",stroke:\"#474D57\"}),React.createElement(\"path\",{d:\"M28.4351 26.2181L40.2199 49.0111L54.5709 26.6167L42.811 3.72412L28.4351 26.2181Z\",fill:\"#2B2F36\"}),React.createElement(\"path\",{d:\"M14.2334 48.3884L40.2197 49.0112L28.4349 26.2183L2.77246 25.0475L14.2334 48.3884Z\",fill:\"#14151A\"})),React.createElement(\"defs\",null,React.createElement(\"filter\",{id:\"filter0_f_4048_115766\",x:.772461,y:.653076,width:55.7983,height:50.3582,filterUnits:\"userSpaceOnUse\",colorInterpolationFilters:\"sRGB\"},React.createElement(\"feFlood\",{floodOpacity:0,result:\"BackgroundImageFix\"}),React.createElement(\"feBlend\",{mode:\"normal\",in:\"SourceGraphic\",in2:\"BackgroundImageFix\",result:\"shape\"}),React.createElement(\"feGaussianBlur\",{stdDeviation:1,result:\"effect1_foregroundBlur_4048_115766\"})))),s2=e=>React.createElement(\"svg\",{width:64,height:58,viewBox:\"0 0 64 58\",fill:\"none\",xmlns:\"http://www.w3.org/2000/svg\",...e},React.createElement(\"g\",{filter:\"url(#filter0_f_4048_115783)\"},React.createElement(\"path\",{d:\"M46.0109 5.82816L31.8718 29.1815L5.64887 28.6489L19.7288 5.41388L46.0109 5.82816Z\",fill:\"#474D57\",stroke:\"#474D57\"}),React.createElement(\"path\",{d:\"M32.1499 29.6873L45.3137 53.6619L59.9916 29.4402L46.8897 5.34204L32.1499 29.6873Z\",fill:\"#2B2F36\"}),React.createElement(\"path\",{d:\"M17.5957 53.662H45.3139L32.1501 29.6874L4.77197 29.1313L17.5957 53.662Z\",fill:\"#14151A\"})),React.createElement(\"defs\",null,React.createElement(\"filter\",{id:\"filter0_f_4048_115783\",x:.771973,y:.909424,width:63.2197,height:56.7526,filterUnits:\"userSpaceOnUse\",colorInterpolationFilters:\"sRGB\"},React.createElement(\"feFlood\",{floodOpacity:0,result:\"BackgroundImageFix\"}),React.createElement(\"feBlend\",{mode:\"normal\",in:\"SourceGraphic\",in2:\"BackgroundImageFix\",result:\"shape\"}),React.createElement(\"feGaussianBlur\",{stdDeviation:2,result:\"effect1_foregroundBlur_4048_115783\"}))));function i2(e){return c.jsxs(\"div\",{className:w.container,...e,children:[c.jsx(a2,{className:w[\"cube-1\"]}),c.jsx(c2,{className:w[\"cube-2\"]}),c.jsx(l2,{className:w[\"cube-3\"]}),c.jsx(o2,{className:w[\"cube-4\"]}),c.jsx(s2,{className:w[\"cube-5\"]})]})}function r2(){const[e,t]=React.useState(!1),a=React.useCallback(()=>{t(!0)},[]),n=React.useCallback(()=>{t(!1)},[]);return c.jsxs(c.Fragment,{children:[c.jsx(i2,{}),c.jsxs(\"main\",{className:M.container,children:[c.jsxs(\"section\",{className:M.content,children:[c.jsx(_1,{}),c.jsx(\"p\",{className:M[\"txt-connect-tips\"],children:\"Please connect wallet to view private files.\"}),c.jsx(Z1,{onClick:a,children:\"Connect Wallet\"})]}),c.jsx(\"footer\",{className:M.footer,children:\"© 2023 BNB Greenfield. All rights reserved.\"})]}),c.jsx(t2,{isOpen:e,onClose:n})]})}function d2(){return c.jsxs(m1,{children:[c.jsx(P,{}),c.jsx(r2,{})]})}B.createRoot(document.getElementById(\"root\")).render(c.jsx(React.StrictMode,{children:c.jsx(d2,{})}));\n\n</script>\n  <style>\n._root_1kufm_1{position:fixed;top:16px;left:50%;transform:translate(-50%);display:flex;flex-direction:column;align-items:center;z-index:var(--index-toast);gap:16px}._container_1kufm_13{display:flex;align-items:center;padding:16px 20px 16px 16px;background:#ffffff;box-shadow:0 4px 24px #00000014;border-radius:8px;word-wrap:break-word}._icon_1kufm_23{display:flex;align-items:center;justify-content:center}._description_1kufm_29{margin-left:8px;font-weight:500;font-size:14px;line-height:1.4}._container_1cbn6_1{display:flex;flex-direction:column;min-height:100vh}._content_1cbn6_7{display:flex;flex-direction:column;flex:1;justify-content:center;align-items:center;margin-top:-16px}._footer_1cbn6_16{flex-shrink:0;height:48px;color:#76808f;font-weight:400;font-size:14px;line-height:17px;display:flex;align-items:center;justify-content:center;text-align:center}._txt-connect-tips_1cbn6_29{font-weight:600;font-size:28px;line-height:34px;color:#fff;padding:40px 0 48px;text-align:center}._button_fyqdl_1{height:48px;padding:0 48px;display:flex;align-items:center;justify-content:center;font-weight:600;font-size:18px;line-height:22px;border-radius:8px;cursor:pointer;transition:var(--transition-normal);background-color:#fff}._button_fyqdl_1:hover{background-color:var(--color-active)}._modal_1ct5u_1{position:fixed;left:0;right:0;top:0;bottom:0;z-index:var(--index-modal);backdrop-filter:blur(10px);transition:var(--transition-normal);align-items:center;justify-content:center;opacity:0;display:flex}._modal-mask_1ct5u_15{width:100%;height:100%;position:absolute;left:0;top:0;background-color:#0003}._modal-content_1ct5u_23{position:relative;width:484px;box-shadow:0 4px 20px #0000000a;border-radius:12px;padding:40px 24px 48px;background-color:#fff;z-index:var(--index-relative)}._modal-header_1ct5u_32{font-weight:600;font-size:24px;line-height:32px;font-family:var(--font-poppins);text-align:center}._modal-body_1ct5u_39{margin:34px 0 40px}._modal-footer_1ct5u_42{display:flex;align-items:center;justify-content:center;font-weight:400;font-size:14px;line-height:17px;color:#76808f}._modal-btn-close_1ct5u_51{top:16px;right:24px;position:absolute;cursor:pointer;transition:var(--transition-normal);width:24px;height:24px;border-radius:4px}._modal-btn-close_1ct5u_51:hover{background-color:#e6e8ea}._link_44242_1{color:#76808f;transition:var(--transition-normal);text-decoration:underline}._link_44242_1:hover{color:var(--color-active)}._loading_1slm0_1{animation:_rotation_1slm0_1 1s infinite linear}@keyframes _rotation_1slm0_1{0%{transform:rotate(0)}to{transform:rotate(360deg)}}._wallet-item_z15a5_1{display:flex;align-items:center;justify-content:center;height:68px;background:#f5f5f5;border-radius:8px;font-weight:600;font-size:18px;line-height:27px;font-family:var(--font-poppins);cursor:pointer;transition:var(--transition-normal);position:relative}._wallet-item_z15a5_1:hover{box-shadow:0 0 0 2px var(--color-active)}._wallet-item_z15a5_1:not(:last-child){margin-bottom:16px}._wallet-item_z15a5_1._active_z15a5_22{box-shadow:0 0 0 2px var(--color-active)!important}._wallet-item_z15a5_1._disabled_z15a5_25{cursor:not-allowed;box-shadow:none}._wallet-item-icon_z15a5_29{position:absolute;left:16px;top:8px}._wallet-item-loading_z15a5_34{position:absolute;top:22px;right:24px}._container_1fm5k_1{position:fixed;pointer-events:none;width:1220px;height:664px;left:50%;top:50%;transform:translate3d(calc(-50% + 47px),-50%,0);opacity:.6;filter:blur(4px)}._container_1fm5k_1>svg{position:absolute}._cube-1_1fm5k_16{left:0;top:0}._cube-2_1fm5k_21{right:178px;top:0}._cube-3_1fm5k_26{left:98px;bottom:51px}._cube-4_1fm5k_31{right:270px;bottom:0}._cube-5_1fm5k_36{right:0;top:236px}:root{--font-inter: Inter, -apple-system, system ui, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif;--font-poppins: Poppins, Inter, -apple-system, system-ui, BlinkMacSystemFont, Segoe UI, Roboto, Helvetica Neue, Arial, sans-serif, Apple Color Emoji, Segoe UI Emoji, Segoe UI Symbol, Noto Color Emoji;--color-active: #f0b90b;--transition-normal: all .2s;--index-modal: 1000;--index-toast: 1500;--index-relative: 1}*,*:before,*:after{border-width:0;border-style:solid;box-sizing:border-box}body,p,div,button,h1,h2,h3,h4,ul,li{padding:0;margin:0}html{line-height:1.5;font-size:14px;font-family:var(--font-inter);-webkit-text-size-adjust:100%;-webkit-font-smoothing:antialiased;-webkit-tap-highlight-color:transparent;text-rendering:optimizeLegibility;-moz-osx-font-smoothing:grayscale;touch-action:manipulation}body{background-color:#1f2026;color:#1e2026}\n\n</style>\n</head>\n\n<body>\n  <div id=\"root\"></div>\n  \n</body>\n\n</html>"
			hc, _ := json.Marshal(htmlConfig)
			html = strings.Replace(html, "<% env %>", string(hc), 1)

			fmt.Fprintf(w, "%s", html)
			return
		}

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

// downloadObjectByUniversalEndpointHandler handles the download object request sent by universal endpoint
func (g *GateModular) downloadObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
	g.getObjectByUniversalEndpointHandler(w, r, true)
}

// viewObjectByUniversalEndpointHandler handles the view object request sent by universal endpoint
func (g *GateModular) viewObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
	g.getObjectByUniversalEndpointHandler(w, r, false)
}
