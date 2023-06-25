package gater

import (
	"encoding/xml"
	"net/http"
	"strconv"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	ErrUnsupportedSignType       = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50001, "unsupported sign type")
	ErrAuthorizationHeaderFormat = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50002, "authorization header format error")
	ErrRequestConsistent         = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50003, "request is tampered")
	ErrNoPermission              = gfsperrors.Register(module.GateModularName, http.StatusUnauthorized, 50004, "no permission")
	ErrDecodeMsg                 = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50005, "gnfd msg encoding error")
	ErrValidateMsg               = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50006, "gnfd msg validate error")
	ErrRefuseApproval            = gfsperrors.Register(module.GateModularName, http.StatusOK, 50007, "approval request is refuse")
	ErrUnsupportedRequestType    = gfsperrors.Register(module.GateModularName, http.StatusNotFound, 50008, "unsupported request type")
	ErrInvalidHeader             = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50009, "invalid request header")
	ErrInvalidQuery              = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50010, "invalid request params for query")
	ErrEncodeResponse            = gfsperrors.Register(module.GateModularName, http.StatusInternalServerError, 50011, "server slipped away, try again later")
	ErrInvalidRange              = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50012, "invalid range params")
	ErrExceptionStream           = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50013, "stream exception")
	ErrMismatchSp                = gfsperrors.Register(module.GateModularName, http.StatusNotAcceptable, 50014, "mismatch sp")
	ErrSignature                 = gfsperrors.Register(module.GateModularName, http.StatusNotAcceptable, 50015, "signature verification failed")
	ErrInvalidPayloadSize        = gfsperrors.Register(module.GateModularName, http.StatusForbidden, 50016, "invalid payload")
	ErrInvalidDomainHeader       = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50017, "The "+GnfdOffChainAuthAppRegNonceHeader+" header is incorrect.")
	ErrInvalidPublicKeyHeader    = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50018, "The "+GnfdOffChainAuthAppRegNonceHeader+" header is incorrect.")
	ErrInvalidRegNonceHeader     = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50019, "The "+GnfdOffChainAuthAppRegNonceHeader+" header is incorrect.")
	ErrSignedMsgNotMatchHeaders  = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50020, "The signed message in "+GnfdAuthorizationHeader+" does not match the content in headers.")
	ErrSignedMsgNotMatchSPAddr   = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50021, "The signed message in "+GnfdAuthorizationHeader+" is not for the this SP.")
	ErrSignedMsgNotMatchTemplate = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50022, "The signed message in "+GnfdAuthorizationHeader+" does not match the template.")
	ErrInvalidExpiryDateHeader   = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50023, "The "+GnfdOffChainAuthAppRegExpiryDateHeader+" header is incorrect. "+
		"The expiry date is expected to be within "+strconv.Itoa(int(MaxExpiryAgeInSec))+" seconds and formatted in YYYY-DD-MM HH:MM:SS 'GMT'Z, e.g. 2023-04-20 16:34:12 GMT+08:00 . ")
	ErrInvalidExpiryDate = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50024, "The expiry parameter is incorrect. "+
		"The expiry date is expected to be within "+strconv.Itoa(int(MaxExpiryAgeInSec))+" seconds and formatted in YYYY-DD-MM HH:MM:SS 'GMT'Z, e.g. 2023-04-20 16:34:12 GMT+08:00 . ")
	ErrNoSuchObject = gfsperrors.Register(module.AuthenticationModularName, http.StatusNotFound, 50025, "no such object")
	ErrForbidden    = gfsperrors.Register(module.GateModularName, http.StatusForbidden, 50026, "Forbidden to access")

	ErrConsensus = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 55001, "server slipped away, try again later")

	ErrApprovalExpired = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 550015, "approval expired")
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
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	w.WriteHeader(int(gfspErr.GetHttpStatusCode()))
	w.Write(xmlBody)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write error response", "error", gfspErr.String())
	}
}
