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
)

// piece store errors
var (
	// ErrUnsupportMethod defines unsupported method error
	ErrUnsupportMethod = errors.New("unsupported method")
	// ErrUnsupportDelimiter defines invalid key with delimiter error
	ErrUnsupportDelimiter = errors.New("unsupported delimiter")
	// ErrInvalidObjectKey defines invalid object key error
	ErrInvalidObjectKey = errors.New("invalid object key")
	// ErrNotExitObject defines not exist object in memory error
	ErrNotExitObject = errors.New("object not exist")
	//ErrNotExistBucket defines not exist bucket error
	ErrNotExistBucket = errors.New("bucket not exist")
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
	ErrRequestConsistent = errors.New("failed to check request consistent")
	// ErrSignatureConsistent defines the invalid signature error
	ErrSignatureConsistent = errors.New("failed to check signature consistent")
	// ErrUnsupportSignType defines the unsupported signature type error
	ErrUnsupportSignType = errors.New("unsupported signature type")
	// ErrEmptyReqHeader defines the empty header error
	ErrEmptyReqHeader = errors.New("request header is empty")
	// ErrInvalidHeader defines the invalid header error
	ErrInvalidHeader = errors.New("invalid request header")
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
