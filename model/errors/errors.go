package errors

import (
	"errors"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
)

// piece store errors
var (
	NotSupportedMethod          = errors.New("not supported method")
	NotSupportedDelimiter       = errors.New("not supported delimiter")
	EmptyObjectKey              = errors.New("object key cannot be empty")
	EmptyMemoryObject           = errors.New("memory object is empty")
	BucketNotExisted            = errors.New("bucket not existed")
	ErrNoPermissionAccessBucket = errors.New("no permission to access the bucket")

	ErrInternalError    = errors.New("internal error")
	ErrDuplicateBucket  = errors.New("duplicate bucket")
	ErrDuplicateObject  = errors.New("duplicate object")
	ErrObjectTxNotExist = errors.New("object tx not exist")
	ErrObjectNotExist   = errors.New("object not exist")
)

// stone hub service errors
var (
	ErrTxHash                   = errors.New("tx hash format error")
	ErrObjectInfo               = errors.New("object info is empty")
	ErrObjectID                 = errors.New("object id is zero")
	ErrObjectSize               = errors.New("object size is zero")
	ErrObjectCreateHeight       = errors.New("object create height is zero")
	ErrParamMissing             = errors.New("params missing")
	ErrPrimaryStorageProvider   = errors.New("primary storage provider mismatch")
	ErrPrimaryPieceChecksum     = errors.New("primary storage provider piece checksum error")
	ErrUploadPayloadJobDone     = errors.New("upload payload job is already completed")
	ErrUploadPayloadJobRunning  = errors.New("upload payload job is running")
	ErrObjectInfoOnInscription  = errors.New("object info not on the inscription")
	ErrUploadPayloadJobNotExist = errors.New("upload payload job not exist")
)

// stone node service errors
var (
	ErrStoneNodeStarted   = errors.New("stone node resource is running")
	ErrStoneNodeStopped   = errors.New("stone node service has stopped")
	ErrIntegrityHash      = errors.New("secondary integrity hash check error")
	ErrRedundancyType     = errors.New("unknown redundancy type")
	ErrEmptyJob           = errors.New("job is empty")
	ErrSecondarySPNumber  = errors.New("secondary sp is not enough")
	ErrInvalidSegmentData = errors.New("invalid segment data, length is not equal to 1")
)

func MakeErrMsgResponse(err error) *service.ErrMessage {
	return &service.ErrMessage{
		ErrCode: service.ErrCode_ERR_CODE_ERROR,
		ErrMsg:  err.Error(),
	}
}
