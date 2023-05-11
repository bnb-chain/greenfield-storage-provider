package gateway

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// APIError structure
type APIError struct {
	Code           string
	Description    string
	HTTPStatusCode int
}

// map error code to APIError structure, these fields carry respective
// descriptions and http status code for all the error responses.
var errCodeMap = map[int]APIError{
	merrors.InternalErrCode: {
		Code:           "InternalError",
		Description:    "We encountered an internal error, please try again.",
		HTTPStatusCode: http.StatusInternalServerError,
	},
	merrors.NoSuchBucketErrCode: {
		Code:           "NoSuchBucket",
		Description:    "The specified bucket does not exist.",
		HTTPStatusCode: http.StatusNotFound,
	},
	merrors.NoSuchObjectErrCode: {
		Code:           "NoSuchKey",
		Description:    "The specified key does not exist.",
		HTTPStatusCode: http.StatusNotFound,
	},
	merrors.RouterNotFoundErrCode: {
		Code:           "NoSuchRouter",
		Description:    "The request can not route any handlers.",
		HTTPStatusCode: http.StatusNotFound,
	},
	merrors.DBQuotaNotEnoughErrCode: {
		Code:           "AccessDenied",
		Description:    "Out of quota.",
		HTTPStatusCode: http.StatusForbidden,
	},
	merrors.InvalidHeaderErrCode: {
		Code:           "InvalidHeader",
		Description:    "The headers are invalid.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	merrors.StringToInt64ErrCode: {
		Code:           "InvalidRequest",
		Description:    "The query parameter in HTTP request is invalid.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	merrors.InvalidBucketNameErrCode: {
		Code:           "InvalidBucketName",
		Description:    "The specified bucket is not valid.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	merrors.InvalidObjectNameErrCode: {
		Code:           "InvalidKey",
		Description:    "Object key is illegal.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	merrors.ZeroPayloadErrCode: {
		Code:           "InvalidPayload",
		Description:    "Payload is empty.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	merrors.ObjectNotCreatedErrCode: {
		Code:           "InvalidObjectState",
		Description:    "Object is not created.",
		HTTPStatusCode: http.StatusForbidden,
	},
	merrors.ObjectNotSealedErrCode: {
		Code:           "InvalidObjectState",
		Description:    "Object is not sealed.",
		HTTPStatusCode: http.StatusForbidden,
	},
	merrors.InvalidRangeErrCode: {
		Code:           "InvalidRequest",
		Description:    "Range is invalid.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	merrors.InvalidAddressErrCode: {
		Code:           "InvalidAddress",
		Description:    "Address is illegal.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	merrors.AuthorizationFormatErrCode: {
		Code:           "InvalidRequest",
		Description:    "Authorization format is invalid.",
		HTTPStatusCode: http.StatusForbidden,
	},
	merrors.InconsistentCanonicalRequestErrCode: {
		Code:           "SignatureDoesNotMatch",
		Description:    "Inconsistent canonical request.",
		HTTPStatusCode: http.StatusForbidden,
	},
	merrors.SignatureConsistentErrCode: {
		Code:           "SignatureDoesNotMatch",
		Description:    "Authorization format is invalid.",
		HTTPStatusCode: http.StatusForbidden,
	},
	merrors.UnsupportedSignTypeErrCode: {
		Code:           "InvalidRequest",
		Description:    "Unsupported signature type.",
		HTTPStatusCode: http.StatusForbidden,
	},
	merrors.NoPermissionErrCode: {
		Code:           "InvalidRequest",
		Description:    "Unsupported signature type.",
		HTTPStatusCode: http.StatusForbidden,
	},
}

func makeXMLHTPPResponse(w http.ResponseWriter, errorCode int, requestID string) error {
	apiErr, ok := errCodeMap[errorCode]
	if !ok {
		apiErr = errCodeMap[merrors.InternalErrCode]
	}
	var xmlInfo = struct {
		XMLName   xml.Name `xml:"Error"`
		Code      string   `xml:"Code"`
		Message   string   `xml:"Message"`
		RequestID string   `xml:"RequestId"`
	}{
		Code:      apiErr.Code,
		Message:   fmt.Sprintf("%s ErrorCode: %d.", apiErr.Description, errorCode),
		RequestID: requestID,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		return err
	}
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeXMLHeaderValue)
	w.WriteHeader(apiErr.HTTPStatusCode)
	if _, err = w.Write(xmlBody); err != nil {
		return err
	}
	return nil
}
