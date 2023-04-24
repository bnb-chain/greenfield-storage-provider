package gateway

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strconv"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// errorDescription describe error info.
type errorDescription struct {
	errorCode    string
	errorMessage string
	statusCode   int
}

// refer: https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
var (
	// 4xx
	InvalidHeader      = &errorDescription{errorCode: "InvalidHeader", errorMessage: "The headers are invalid.", statusCode: http.StatusBadRequest}
	InvalidQuery       = &errorDescription{errorCode: "InvalidQuery", errorMessage: "The queries are invalid.", statusCode: http.StatusBadRequest}
	InvalidBucketName  = &errorDescription{errorCode: "InvalidBucketName", errorMessage: "The specified bucket is not valid.", statusCode: http.StatusBadRequest}
	InvalidKey         = &errorDescription{errorCode: "InvalidKey", errorMessage: "Object key is illegal", statusCode: http.StatusBadRequest}
	InvalidPayload     = &errorDescription{errorCode: "InvalidPayload", errorMessage: "Payload is empty", statusCode: http.StatusBadRequest}
	InvalidObjectState = &errorDescription{errorCode: "InvalidObjectState", errorMessage: "Object state is invalid", statusCode: http.StatusForbidden}
	InvalidRange       = &errorDescription{errorCode: "InvalidRange", errorMessage: "Range is invalid", statusCode: http.StatusBadRequest}
	InvalidAddress     = &errorDescription{errorCode: "InvalidAddress", errorMessage: "Address is illegal", statusCode: http.StatusBadRequest}
	InvalidMaxKeys     = &errorDescription{errorCode: "InvalidMaxKeys", errorMessage: "MaxKeys is illegal", statusCode: http.StatusBadRequest}
	InvalidStartAfter  = &errorDescription{errorCode: "InvalidStartAfter", errorMessage: "StartAfter is illegal", statusCode: http.StatusBadRequest}
	SignatureNotMatch  = &errorDescription{errorCode: "SignatureDoesNotMatch", errorMessage: "SignatureDoesNotMatch", statusCode: http.StatusForbidden}
	AccessDenied       = &errorDescription{errorCode: "AccessDenied", errorMessage: "Access Denied", statusCode: http.StatusForbidden}
	OutOfQuota         = &errorDescription{errorCode: "AccessDenied", errorMessage: "Out of Quota", statusCode: http.StatusForbidden}
	NoSuchKey          = &errorDescription{errorCode: "NoSuchKey", errorMessage: "The specified key does not exist.", statusCode: http.StatusNotFound}
	NoSuchBucket       = &errorDescription{errorCode: "NoSuchBucket", errorMessage: "The specified bucket does not exist.", statusCode: http.StatusNotFound}
	NoRouter           = &errorDescription{errorCode: "NoRouter", errorMessage: "The request can not route any handlers", statusCode: http.StatusNotFound}
	// 5xx
	InternalError          = &errorDescription{errorCode: "InternalError", errorMessage: "Internal Server Error", statusCode: http.StatusInternalServerError}
	NotImplementedError    = &errorDescription{errorCode: "NotImplementedError", errorMessage: "Not Implemented Error", statusCode: http.StatusNotImplemented}
	NotExistComponentError = &errorDescription{errorCode: "NotExistComponentError", errorMessage: "Not Existed Component Error", statusCode: http.StatusNotImplemented}
)

// off-chain-auth errors
var (
	// 4xx
	InvalidRegNonceHeader     = &errorDescription{errorCode: "InvalidRegNonceHeader", errorMessage: "The " + model.GnfdOffChainAuthAppRegNonceHeader + " header is incorrect.", statusCode: http.StatusBadRequest}
	SignedMsgNotMatchHeaders  = &errorDescription{errorCode: "SigMsgNotMatchHeaders", errorMessage: "The signed message in " + model.GnfdAuthorizationHeader + " does not match the content in headers.", statusCode: http.StatusBadRequest}
	SignedMsgNotMatchSPAddr   = &errorDescription{errorCode: "SignedMsgNotMatchSPAddr", errorMessage: "The signed message in " + model.GnfdAuthorizationHeader + " is not for the this SP.", statusCode: http.StatusBadRequest}
	SignedMsgNotMatchTemplate = &errorDescription{errorCode: "SignedMsgNotMatchTemplate", errorMessage: "The signed message in " + model.GnfdAuthorizationHeader + " does not match the template.", statusCode: http.StatusBadRequest}
	InvalidExpiryDateHeader   = &errorDescription{errorCode: "InvalidExpiryDateHeader",
		errorMessage: "The " + model.GnfdOffChainAuthAppRegExpiryDateHeader + " header is incorrect. " +
			"The expiry date is expected to be within " + strconv.Itoa(int(MaxExpiryAgeInSec)) + " seconds and formatted in YYYY-DD-MM HH:MM:SS 'GMT'Z, e.g. 2023-04-20 16:34:12 GMT+08:00 . ",
		statusCode: http.StatusBadRequest}
)

// errorResponse is used to error response xml.
func (desc *errorDescription) errorResponse(w http.ResponseWriter, reqCtx *requestContext) error {
	var (
		xmlBody []byte
		err     error
	)

	var xmlInfo = struct {
		XMLName   xml.Name `xml:"Error"`
		Code      string   `xml:"Code"`
		Message   string   `xml:"Message"`
		RequestId string   `xml:"RequestId"`
	}{
		Code:      desc.errorCode,
		Message:   desc.errorMessage,
		RequestId: reqCtx.requestID,
	}
	if xmlBody, err = xml.Marshal(&xmlInfo); err != nil {
		return err
	}
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeXMLHeaderValue)
	w.WriteHeader(desc.statusCode)
	if _, err = w.Write(xmlBody); err != nil {
		return err
	}
	return nil
}

// errorJSONResponse is used to error response JSON.
func (desc *errorDescription) errorJSONResponse(w http.ResponseWriter, reqCtx *requestContext) error {
	var (
		jsonBody []byte
		err      error
	)

	var jsonInfo = struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"requestID"`
	}{
		Code:      desc.errorCode,
		Message:   desc.errorMessage,
		RequestID: reqCtx.requestID,
	}
	if jsonBody, err = json.Marshal(&jsonInfo); err != nil {
		return err
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.WriteHeader(desc.statusCode)

	if _, err = w.Write(jsonBody); err != nil {
		return err
	}
	return nil
}

// makeContentRangeHeader make http response range header
func makeContentRangeHeader(w http.ResponseWriter, start int64, end int64) {
	if end < 0 {
		w.Header().Set(model.ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(start))+"-")
	} else {
		w.Header().Set(model.ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(start))+"-"+util.Uint64ToString(uint64(end)))
	}
}

// makeErrorDescription is used to convent err to errorDescription
func makeErrorDescription(err error) *errorDescription {
	switch err {
	case merrors.ErrNoSuchObject:
		return NoSuchKey
	case merrors.ErrNoSuchBucket:
		return NoSuchBucket
	case merrors.ErrAuthorizationFormat, merrors.ErrRequestConsistent, merrors.ErrSignatureConsistent, merrors.ErrUnsupportedSignType:
		return SignatureNotMatch
	case merrors.ErrNoPermission, merrors.ErrCheckPaymentAccountActive:
		return AccessDenied
	case merrors.ErrCheckQuotaEnough:
		return OutOfQuota
	case merrors.ErrCheckObjectCreated, merrors.ErrCheckObjectSealed:
		return InvalidObjectState
	default:
		return InternalError
	}
}
