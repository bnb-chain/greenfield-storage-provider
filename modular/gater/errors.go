package gater

import (
	"net/http"
	"strconv"

	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
)

var (
	ErrUnsupportedSignType       = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50001, "unsupported sign type")
	ErrAuthorizationHeaderFormat = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50002, "authorization header format error")
	ErrRequestConsistent         = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50003, "request is tampered")
	ErrNoPermission              = gfsperrors.Register(module.GateModularName, http.StatusUnauthorized, 50004, "no permission")
	ErrDecodeMsg                 = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50005, "gnfd msg encoding error")
	ErrValidateMsg               = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50006, "gnfd msg validate error")
	ErrRefuseApproval            = gfsperrors.Register(module.GateModularName, http.StatusOK, 50007, "approval request is refused")
	ErrUnsupportedRequestType    = gfsperrors.Register(module.GateModularName, http.StatusNotFound, 50008, "unsupported request type")
	ErrInvalidHeader             = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50009, "invalid request header")
	ErrInvalidQuery              = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50010, "invalid request params for query")
	ErrInvalidRange              = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50012, "invalid range params")
	ErrExceptionStream           = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50013, "stream exception")
	ErrMismatchSp                = gfsperrors.Register(module.GateModularName, http.StatusNotAcceptable, 50014, "mismatch sp")
	ErrSignature                 = gfsperrors.Register(module.GateModularName, http.StatusNotAcceptable, 50015, "signature is invalid")
	ErrInvalidPayloadSize        = gfsperrors.Register(module.GateModularName, http.StatusForbidden, 50016, "invalid payload")
	ErrInvalidDomainHeader       = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50017, "The "+GnfdOffChainAuthAppDomainHeader+" header is incorrect.")
	ErrInvalidPublicKeyHeader    = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50018, "The "+GnfdOffChainAuthAppRegPublicKeyHeader+" header is incorrect.")
	ErrInvalidRegNonceHeader     = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50019, "The "+GnfdOffChainAuthAppRegNonceHeader+" header is incorrect.")
	ErrSignedMsgNotMatchHeaders  = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50020, "The signed message in "+GnfdAuthorizationHeader+" does not match the content in headers.")
	ErrSignedMsgNotMatchSPAddr   = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50021, "The SP addr in the signed message in "+GnfdAuthorizationHeader+" is not for the this SP.")
	ErrSignedMsgNotMatchSPNonce  = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50040, "The SP Nonce in the signed message in "+GnfdAuthorizationHeader+" is not for the this SP.")
	ErrSignedMsgNotMatchDomain   = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50037, "The domain in the signed message in "+GnfdAuthorizationHeader+" does not match this website.")
	ErrSignedMsgNotMatchExpiry   = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50038, "The expiry time in signed message in "+GnfdAuthorizationHeader+" does not match the expiry time in the header "+GnfdOffChainAuthAppRegExpiryDateHeader+".")
	ErrSignedMsgNotMatchPubKey   = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50039, "The public key in signed message in "+GnfdAuthorizationHeader+" does not match the expiry time in the header "+GnfdOffChainAuthAppRegPublicKeyHeader+".")
	ErrSignedMsgNotMatchTemplate = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50022, "The signed message in "+GnfdAuthorizationHeader+" does not match the template.")
	ErrInvalidExpiryDateHeader   = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50023, "The "+commonhttp.HTTPHeaderExpiryTimestamp+" header is incorrect. "+
		"The expiry date is expected to be within "+strconv.Itoa(int(MaxExpiryAgeInSec))+" seconds and formatted in YYYY-DD-MM HH:MM:SS 'GMT'Z, e.g. 2023-04-20 16:34:12 GMT+08:00 . ")
	ErrInvalidExpiryDateParam = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50024, "The "+commonhttp.HTTPHeaderExpiryTimestamp+" parameter is incorrect. "+
		"The expiry date is expected to be within "+strconv.Itoa(int(MaxExpiryAgeInSec))+" seconds and formatted in YYYY-DD-MM HH:MM:SS 'GMT'Z, e.g. 2023-04-20 16:34:12 GMT+08:00 . ")
	ErrNoSuchObject           = gfsperrors.Register(module.AuthenticationModularName, http.StatusNotFound, 50025, "no such object")
	ErrForbidden              = gfsperrors.Register(module.GateModularName, http.StatusForbidden, 50026, "Forbidden to access")
	ErrInvalidComplete        = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50027, "invalid complete")
	ErrInvalidOffset          = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50028, "invalid offset")
	ErrSPUnavailable          = gfsperrors.Register(module.GateModularName, http.StatusForbidden, 50029, "sp is not in service status")
	ErrRecoverySP             = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50030, "The SP is not the correct SP to recovery")
	ErrRecoveryRedundancyType = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50031, "The redundancy type of the recovering piece is not EC")
	ErrRecoveryTimeout        = gfsperrors.Register(module.GateModularName, http.StatusInternalServerError, 50032, "System busy, try to request later")
	ErrInvalidRedundancyIndex = gfsperrors.Register(module.GateModularName, http.StatusInternalServerError, 50035, "invalid redundancy index")
	ErrBucketUnavailable      = gfsperrors.Register(module.GateModularName, http.StatusForbidden, 50036, "bucket is not in service status")
	ErrReplyData              = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50037, "reply the downloaded data to client failed")
	ErrTaskMsgExpired         = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50038, "the update time of the task has exceed the expire time")
	ErrSecondaryMismatch      = gfsperrors.Register(module.GateModularName, http.StatusNotAcceptable, 50039, "secondary sp mismatch")
	ErrPrimaryMismatch        = gfsperrors.Register(module.GateModularName, http.StatusNotAcceptable, 50041, "primary sp mismatch")
	ErrNotCreatedState        = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50042, "object has not been created state")
	ErrNotSealedState         = gfsperrors.Register(module.GateModularName, http.StatusBadRequest, 50043, "object has not been sealed state")
)

func ErrEncodeResponseWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.GateModularName, http.StatusInternalServerError, 50011, detail)
}

func ErrMigrateApprovalWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.GateModularName, http.StatusInternalServerError, 50033, detail)
}

func ErrNotifySwapOutWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.GateModularName, http.StatusInternalServerError, 50034, detail)
}

func ErrConsensusWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.GateModularName, http.StatusInternalServerError, 55001, detail)
}
