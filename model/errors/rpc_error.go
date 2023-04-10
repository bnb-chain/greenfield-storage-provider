package errors

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// common error
var (
	// ErrCacheMiss defines cache missing error
	ErrCacheMiss = errors.New("cache missing")
	// ErrSealTimeout defines seal object timeout error
	ErrSealTimeout = errors.New("seal object timeout")
	// ErrNoSuchObject defines not existed object error
	ErrNoSuchObject = errors.New("the specified key does not exist")
	// ErrNoSuchBucket defines not existed bucket error
	ErrNoSuchBucket = errors.New("the specified bucket does not exist")
	// ErrInvalidBucketName defines invalid bucket name
	ErrInvalidBucketName = errors.New("invalid bucket name")
	// ErrUnsupportedMethod defines unsupported method error
	ErrUnsupportedMethod = errors.New("unsupported method")
	// ErrIntegerOverflow defines integer overflow
	ErrIntegerOverflow = errors.New("integer overflow")
	// ErrDanglingPointer defines the nil pointer error
	ErrDanglingPointer = errors.New("pointer dangling")
	// ErrInvalidParams defines invalid params
	ErrInvalidParams = errors.New("invalid params")
)

// piece store errors
var (
	// ErrUnsupportedDelimiter defines invalid key with delimiter error
	ErrUnsupportedDelimiter = errors.New("unsupported delimiter")
	// ErrInvalidObjectKey defines invalid object key error
	ErrInvalidObjectKey = errors.New("invalid object key")
	// ErrNoPermissionAccessBucket defines deny access bucket error
	ErrNoPermissionAccessBucket = errors.New("deny access bucket")
)

// gateway errors
var (
	// ErrInternalError defines storage provider internal error
	ErrInternalError = errors.New("internal error")
	// ErrDuplicateBucket defines duplicate bucket error
	ErrDuplicateBucket = errors.New("duplicate bucket")
	// ErrDuplicateObject defines duplicate object error
	ErrDuplicateObject = errors.New("duplicate object")
	// ErrPayloadZero defines payload size is zero error
	ErrPayloadZero = errors.New("object payload is zero")
	// ErrAuthorizationFormat defines the invalid authorization format error
	ErrAuthorizationFormat = errors.New("authorization format error")
	// ErrRequestConsistent defines the invalid request checksum error
	ErrRequestConsistent = errors.New("request is tampered")
	// ErrSignatureConsistent defines the invalid signature error
	ErrSignatureConsistent = errors.New("signature is not consistent")
	// ErrUnsupportedSignType defines the unsupported signature type error
	ErrUnsupportedSignType = errors.New("unsupported signature type")
	// ErrEmptyReqHeader defines the empty header error
	ErrEmptyReqHeader = errors.New("request header is empty")
	// ErrInvalidHeader defines the invalid header error
	ErrInvalidHeader = errors.New("invalid request header")
	// ErrNoPermission defines the authorization error
	ErrNoPermission = errors.New("no permission")
	// ErrCheckObjectCreated defines the check object state error
	ErrCheckObjectCreated = errors.New("object is not created")
	// ErrCheckObjectSealed defines the check object state error
	ErrCheckObjectSealed = errors.New("object is not sealed")
	// ErrCheckPaymentAccountActive defines check payment account state is active
	ErrCheckPaymentAccountActive = errors.New("payment account is not active")
	// ErrCheckQuotaEnough defines check quota is enough
	ErrCheckQuotaEnough = errors.New("quota is not enough")
	// ErrSPMismatch defines the SP's operate address mismatch error
	ErrSPMismatch = errors.New("the operator address of SP is a mismatch")
	// ErrApprovalExpire defines the SP's operate address mismatch error
	ErrApprovalExpire = errors.New("approval expired")
	// ErrSignatureInvalid defines the replicate approval signature invalid
	ErrSignatureInvalid = errors.New("invalid replicate approval signature")
)

// signer service error
var (
	// ErrIPBlocked defines deny request by ip error
	ErrIPBlocked = errors.New("ip blocked")
	// ErrAPIKey defines invalid signer api key
	ErrAPIKey = errors.New("invalid api key")
	// ErrSignMsg defines sign msg error by private key
	ErrSignMsg = errors.New("sign message with private key failed")
	// ErrSealObjectOnChain defines send seal object tx to chain error
	ErrSealObjectOnChain = errors.New("send sealObject msg failed")
)

// metadata service error
var (
	// ErrInvalidAccountID defines invalid account id
	ErrInvalidAccountID = errors.New("invalid account id")
)

// task node service error
var (
	// ErrSPApprovalNumber defines failed to insufficient SPs' approvals from p2p server
	ErrSPApprovalNumber = errors.New("failed to get sufficient approvals of SPs from p2p server")
	// ErrSPNumber defines failed to get insufficient SPs from DB
	ErrSPNumber = errors.New("failed to get sufficient SPs from DB")
	// ErrExhaustedSP defines no backup SP to pick up error
	ErrExhaustedSP = errors.New("backup storage providers exhausted")
)

// uploader service error
var (
	// ErrMismatchIntegrityHash defines integrity hash mismatch error
	ErrMismatchIntegrityHash = errors.New("integrity hash mismatch")
	// ErrMismatchChecksumNum defines checksum number mismatch error
	ErrMismatchChecksumNum = errors.New("checksum number mismatch")
)

// InnerErrorToGRPCError convents inner error to grpc/status error
func InnerErrorToGRPCError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		errors.Is(err, ErrNoSuchObject) {
		return status.Errorf(codes.NotFound, "Object is not found")
	}
	if errors.Is(err, ErrCheckQuotaEnough) {
		return status.Errorf(codes.PermissionDenied, "Quota is not enough")
	}
	return err
}

// GRPCErrorToInnerError convents grpc/status error to inner error
func GRPCErrorToInnerError(err error) error {
	errStatus, _ := status.FromError(err)
	if codes.NotFound == errStatus.Code() {
		return ErrNoSuchObject
	}
	if codes.PermissionDenied == errStatus.Code() {
		return ErrCheckQuotaEnough
	}
	return err
}
