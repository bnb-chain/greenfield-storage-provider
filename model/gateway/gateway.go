package gateway

import (
	"encoding/xml"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// ErrorResponse is used in gateway error response
type ErrorResponse struct {
	XMLName xml.Name `xml:"Error"`
	Code    int32    `xml:"Code"`
	Message string   `xml:"Message"`
}

// MakeErrorResponse writes error response to http body
func MakeErrorResponse(w http.ResponseWriter, err error) {
	gfspErr := gfsperrors.MakeGfSpError(err)
	xmlInfo := ErrorResponse{
		Code:    gfspErr.GetInnerCode(),
		Message: gfspErr.GetDescription(),
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal error response", "gfsp_error", gfspErr.String(), "error", err)
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(int(gfspErr.GetHttpStatusCode()))
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to marshal error response", "gfsp_error", gfspErr.String(), "error", err)
	}
}
