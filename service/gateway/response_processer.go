package gateway

import (
	"encoding/xml"
	"net/http"
)

type errorDescription struct {
	errorCode    string
	errorMessage string
	statusCode   int
}

// refer: https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
var (
	// 4xx
	InvalidBucketName     = &errorDescription{errorCode: "InvalidBucketName", errorMessage: "The specified bucket is not valid.", statusCode: http.StatusBadRequest}
	InvalidKey            = &errorDescription{errorCode: "InvalidKey", errorMessage: "Object key is Illegal", statusCode: http.StatusBadRequest}
	UnauthorizedAccess    = &errorDescription{errorCode: "UnauthorizedAccess", errorMessage: "UnauthorizedAccess", statusCode: http.StatusUnauthorized}
	AccessDenied          = &errorDescription{errorCode: "AccessDenied", errorMessage: "Access Denied", statusCode: http.StatusForbidden}
	SignatureDoesNotMatch = &errorDescription{errorCode: "SignatureDoesNotMatch", errorMessage: "SignatureDoesNotMatch", statusCode: http.StatusForbidden}
	NoSuchKey             = &errorDescription{errorCode: "NoSuchKey", errorMessage: "The specified key does not exist.", statusCode: http.StatusNotFound}
	ObjectTxNotFound      = &errorDescription{errorCode: "ObjectTxNotFound", errorMessage: "The specified object tx does not exist.", statusCode: http.StatusNotFound}
	BucketAlreadyExists   = &errorDescription{errorCode: "CreateBucketFailed", errorMessage: "Duplicate bucket name.", statusCode: http.StatusConflict}
	ObjectAlreadyExists   = &errorDescription{errorCode: "PutObjectFailed", errorMessage: "Duplicate object name.", statusCode: http.StatusConflict}
	// 5xx
	InternalError = &errorDescription{errorCode: "InternalError", errorMessage: "Internal Server Error", statusCode: http.StatusInternalServerError}
)

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
	w.Header().Set(ContextTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		return err
	}
	return nil
}
