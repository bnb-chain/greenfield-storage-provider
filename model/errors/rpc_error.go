package errors

import (
	"errors"
)

// common error
var (
	ErrCacheMiss   = errors.New("cache missing")
	ErrSealTimeout = errors.New("seal object timeout")
)

// piece store errors
var (
	NotSupportedMethod          = errors.New("unsupported method")
	NotSupportedDelimiter       = errors.New("unsupported delimiter")
	EmptyObjectKey              = errors.New("invalid object key")
	EmptyMemoryObject           = errors.New("memory object not exist")
	BucketNotExisted            = errors.New("bucket not exist")
	ErrNoPermissionAccessBucket = errors.New("deny access bucket")
)

// gateway errors
var (
	ErrInternalError    = errors.New("internal error")
	ErrDuplicateBucket  = errors.New("duplicate bucket")
	ErrDuplicateObject  = errors.New("duplicate object")
	ErrObjectTxNotExist = errors.New("object tx not exist")
	ErrObjectNotExist   = errors.New("object not exist")
	ErrObjectIsEmpty    = errors.New("object payload is empty")

	// signature error
	ErrAuthorizationFormat = errors.New("authorization format error")
	ErrRequestConsistent   = errors.New("failed to check request consistent")
	ErrSignatureConsistent = errors.New("failed to check signature consistent")
	ErrUnsupportedSignType = errors.New("unsupported signature type")
	ErrEmptyReqHeader      = errors.New("request header is empty")
	ErrReqHeader           = errors.New("invalid request header")
)

// signer service error
var (
	ErrIPBlocked         = errors.New("ip blocked")
	ErrAPIKey            = errors.New("invalid api key")
	ErrSignMsg           = errors.New("sign message with private key failed")
	ErrSealObjectOnChain = errors.New("send sealObject msg failed")
)

// block syncer service errors
var (
	ErrSyncerStopped = errors.New("syncer service has already stopped")
)
