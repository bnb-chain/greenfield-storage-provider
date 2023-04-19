package gateway

import (
	"encoding/xml"
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
		Code:           "",
		Description:    "",
		HTTPStatusCode: http.StatusInternalServerError,
	},
	merrors.NoSuchBucketErrCode: {
		Code:           "",
		Description:    "The specified bucket does not exist.",
		HTTPStatusCode: http.StatusNotFound,
	},
	merrors.NoSuchObjectErrCode: {
		Code:           "",
		Description:    "The specified key does not exist.",
		HTTPStatusCode: http.StatusNotFound,
	},
	merrors.RouterNotFoundErrCode: {
		Code:           "",
		Description:    "The request can not route any handlers.",
		HTTPStatusCode: http.StatusNotFound,
	},
	merrors.DBQuotaNotEnoughErrCode: {
		Code:           "",
		Description:    "Out of quota.",
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
		Message:   apiErr.Description,
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
