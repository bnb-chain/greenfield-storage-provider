package errors

import (
	"errors"
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
)

// piece store errors
var (
	// ErrUnsupportedMethod defines unsupported method error
	ErrUnsupportedMethod = errors.New("unsupported method")
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
