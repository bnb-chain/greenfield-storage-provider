package gater

import (
	"encoding/xml"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	ErrUnsupportedSignType    = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50001, "unsupported sign type")
	ErrAuthorizationFormat    = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50002, "authorization format error")
	ErrRequestConsistent      = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50003, "request is tampered")
	ErrNoPermission           = gfsperrors.Register(GateModularName, http.StatusUnauthorized, 50004, "no permission")
	ErrDecodeMsg              = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50005, "gnfd msg encoding error")
	ErrValidateMsg            = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50006, "gnfd msg validate error")
	ErrRefuseApproval         = gfsperrors.Register(GateModularName, http.StatusOK, 50007, "approval request is refuse")
	ErrUnsupportedRequestType = gfsperrors.Register(GateModularName, http.StatusNotFound, 50008, "unsupported request type")
	ErrInvalidHeader          = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50009, "invalid request header")
	ErrInvalidQuery           = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50010, "invalid request header params for query")
	ErrEncodeResponse         = gfsperrors.Register(GateModularName, http.StatusInternalServerError, 50011, "server slipped away, try again later")
	ErrInvalidRange           = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50012, "invalid range params")
	ErrExceptionStream        = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50013, "stream exception")
	ErrMisMatchSp             = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50014, "mismatch sp")
	ErrSignature              = gfsperrors.Register(GateModularName, http.StatusBadRequest, 50015, "signature verification failed")

	ErrConsensus = gfsperrors.Register(GateModularName, http.StatusBadRequest, 55001, "server slipped away, try again later")
)

func MakeErrorResponse(w http.ResponseWriter, err error) {
	gfspErr := gfsperrors.MakeGfSpError(err)
	var xmlInfo = struct {
		XMLName xml.Name `xml:"Error"`
		Code    int32    `xml:"Code"`
		Message string   `xml:"Message"`
	}{
		Code:    gfspErr.GetInnerCode(),
		Message: gfspErr.GetDescription(),
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal error response", "error", gfspErr.String())
	}
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeXMLHeaderValue)
	w.WriteHeader(int(gfspErr.GetHttpStatusCode()))
	w.Write(xmlBody)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write error response", "error", gfspErr.String())
	}
}
