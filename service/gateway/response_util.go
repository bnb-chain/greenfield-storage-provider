package gateway

import (
	"encoding/xml"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
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
	InvalidHeader         = &errorDescription{errorCode: "InvalidHeader", errorMessage: "The headers maybe is invalid.", statusCode: http.StatusBadRequest}
	InvalidBucketName     = &errorDescription{errorCode: "InvalidBucketName", errorMessage: "The specified bucket is not valid.", statusCode: http.StatusBadRequest}
	InvalidKey            = &errorDescription{errorCode: "InvalidKey", errorMessage: "Object key is Illegal", statusCode: http.StatusBadRequest}
	InvalidTxHash         = &errorDescription{errorCode: "InvalidTxHash", errorMessage: "transaction hash is Illegal", statusCode: http.StatusBadRequest}
	InvalidPayload        = &errorDescription{errorCode: "InvalidPayload", errorMessage: "payload is empty", statusCode: http.StatusBadRequest}
	InvalidRange          = &errorDescription{errorCode: "InvalidRange", errorMessage: "range is invalid", statusCode: http.StatusBadRequest}
	UnauthorizedAccess    = &errorDescription{errorCode: "UnauthorizedAccess", errorMessage: "UnauthorizedAccess", statusCode: http.StatusUnauthorized}
	AccessDenied          = &errorDescription{errorCode: "AccessDenied", errorMessage: "Access Denied", statusCode: http.StatusForbidden}
	SignatureDoesNotMatch = &errorDescription{errorCode: "SignatureDoesNotMatch", errorMessage: "SignatureDoesNotMatch", statusCode: http.StatusForbidden}
	NoSuchKey             = &errorDescription{errorCode: "NoSuchKey", errorMessage: "The specified key does not exist.", statusCode: http.StatusNotFound}
	ObjectTxNotFound      = &errorDescription{errorCode: "ObjectTxNotFound", errorMessage: "The specified object tx does not exist.", statusCode: http.StatusNotFound}
	BucketAlreadyExists   = &errorDescription{errorCode: "CreateBucketFailed", errorMessage: "Duplicate bucket name.", statusCode: http.StatusConflict}
	ObjectAlreadyExists   = &errorDescription{errorCode: "PutObjectFailed", errorMessage: "Duplicate object name.", statusCode: http.StatusConflict}
	// 5xx
	InternalError       = &errorDescription{errorCode: "InternalError", errorMessage: "Internal Server Error", statusCode: http.StatusInternalServerError}
	NotImplementedError = &errorDescription{errorCode: "NotImplementedError", errorMessage: "Not Implemented Error", statusCode: http.StatusNotImplemented}
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

func generateContentRangeHeader(w http.ResponseWriter, start int64, end int64) {
	if end < 0 {
		w.Header().Set(model.ContentLengthHeader, "bytes "+util.Uint64ToHeader(uint64(start))+"-")
	} else {
		w.Header().Set(model.ContentLengthHeader, "bytes "+util.Uint64ToHeader(uint64(start))+"-"+util.Uint64ToHeader(uint64(end)))
	}
}
