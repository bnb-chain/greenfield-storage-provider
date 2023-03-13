package gateway

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
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
	InvalidHeader      = &errorDescription{errorCode: "InvalidHeader", errorMessage: "The headers maybe is invalid.", statusCode: http.StatusBadRequest}
	InvalidBucketName  = &errorDescription{errorCode: "InvalidBucketName", errorMessage: "The specified bucket is not valid.", statusCode: http.StatusBadRequest}
	InvalidKey         = &errorDescription{errorCode: "InvalidKey", errorMessage: "Object key is Illegal", statusCode: http.StatusBadRequest}
	InvalidPayload     = &errorDescription{errorCode: "InvalidPayload", errorMessage: "payload is empty", statusCode: http.StatusBadRequest}
	InvalidObjectState = &errorDescription{errorCode: "InvalidObjectState", errorMessage: "object state is invalid", statusCode: http.StatusBadRequest}
	InvalidRange       = &errorDescription{errorCode: "InvalidRange", errorMessage: "range is invalid", statusCode: http.StatusBadRequest}
	SignatureNotMatch  = &errorDescription{errorCode: "SignatureDoesNotMatch", errorMessage: "SignatureDoesNotMatch", statusCode: http.StatusForbidden}
	AccessDenied       = &errorDescription{errorCode: "AccessDenied", errorMessage: "Access Denied", statusCode: http.StatusForbidden}
	NoSuchKey          = &errorDescription{errorCode: "NoSuchKey", errorMessage: "The specified key does not exist.", statusCode: http.StatusNotFound}
	NoSuchBucket       = &errorDescription{errorCode: "NoSuchBucket", errorMessage: "The specified bucket does not exist.", statusCode: http.StatusNotFound}
	// 5xx
	InternalError          = &errorDescription{errorCode: "InternalError", errorMessage: "Internal Server Error", statusCode: http.StatusInternalServerError}
	NotImplementedError    = &errorDescription{errorCode: "NotImplementedError", errorMessage: "Not Implemented Error", statusCode: http.StatusNotImplemented}
	NotExistComponentError = &errorDescription{errorCode: "NotExistComponentError", errorMessage: "Not Exist Component Error", statusCode: http.StatusNotImplemented}
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

	w.WriteHeader(desc.statusCode)
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeXMLHeaderValue)
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
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestId string `json:"RequestId"`
	}{
		Code:      desc.errorCode,
		Message:   desc.errorMessage,
		RequestId: reqCtx.requestID,
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

func generateContentRangeHeader(w http.ResponseWriter, start int64, end int64) {
	if end < 0 {
		w.Header().Set(model.ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(start))+"-")
	} else {
		w.Header().Set(model.ContentRangeHeader, "bytes "+util.Uint64ToString(uint64(start))+"-"+util.Uint64ToString(uint64(end)))
	}
}

func makeErrorDescription(err error) *errorDescription {
	switch err {
	case errors.ErrNoSuchObject:
		return NoSuchKey
	case errors.ErrNoSuchBucket:
		return NoSuchBucket
	case errors.ErrAuthorizationFormat, errors.ErrRequestConsistent, errors.ErrSignatureConsistent, errors.ErrUnsupportedSignType:
		return SignatureNotMatch
	case errors.ErrNoPermission, errors.ErrCheckPaymentAccountActive, errors.ErrCheckQuotaEnough:
		return AccessDenied
	case errors.ErrCheckObjectCreated, errors.ErrCheckObjectSealed:
		return InvalidObjectState
	default:
		return &errorDescription{errorCode: "InternalError", errorMessage: err.Error(), statusCode: http.StatusInternalServerError}
	}
}
