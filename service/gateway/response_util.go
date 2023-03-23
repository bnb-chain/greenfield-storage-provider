package gateway

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

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
	InvalidKey         = &errorDescription{errorCode: "InvalidKey", errorMessage: "Object key is Illegal", statusCode: http.StatusBadRequest}
	InvalidPayload     = &errorDescription{errorCode: "InvalidPayload", errorMessage: "Payload is empty", statusCode: http.StatusBadRequest}
	InvalidObjectState = &errorDescription{errorCode: "InvalidObjectState", errorMessage: "Object state is invalid", statusCode: http.StatusBadRequest}
	InvalidRange       = &errorDescription{errorCode: "InvalidRange", errorMessage: "Range is invalid", statusCode: http.StatusBadRequest}
	InvalidAddress     = &errorDescription{errorCode: "InvalidAddress", errorMessage: "Address is Illegal", statusCode: http.StatusBadRequest}
	SignatureNotMatch  = &errorDescription{errorCode: "SignatureDoesNotMatch", errorMessage: "SignatureDoesNotMatch", statusCode: http.StatusForbidden}
	AccessDenied       = &errorDescription{errorCode: "AccessDenied", errorMessage: "Access Denied", statusCode: http.StatusForbidden}
	NoSuchKey          = &errorDescription{errorCode: "NoSuchKey", errorMessage: "The specified key does not exist.", statusCode: http.StatusNotFound}
	NoSuchBucket       = &errorDescription{errorCode: "NoSuchBucket", errorMessage: "The specified bucket does not exist.", statusCode: http.StatusNotFound}
	NoRouter           = &errorDescription{errorCode: "NoRouter", errorMessage: "The request can not route any handlers", statusCode: http.StatusNotFound}
	// 5xx
	InternalError          = &errorDescription{errorCode: "InternalError", errorMessage: "Internal Server Error", statusCode: http.StatusInternalServerError}
	NotImplementedError    = &errorDescription{errorCode: "NotImplementedError", errorMessage: "Not Implemented Error", statusCode: http.StatusNotImplemented}
	NotExistComponentError = &errorDescription{errorCode: "NotExistComponentError", errorMessage: "Not Existed Component Error", statusCode: http.StatusNotImplemented}
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
	case merrors.ErrNoPermission, merrors.ErrCheckPaymentAccountActive, merrors.ErrCheckQuotaEnough:
		return AccessDenied
	case merrors.ErrCheckObjectCreated, merrors.ErrCheckObjectSealed:
		return InvalidObjectState
	default:
		return &errorDescription{errorCode: "InternalError", errorMessage: err.Error(), statusCode: http.StatusInternalServerError}
	}
}
