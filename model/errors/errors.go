package errors

import (
	"errors"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
)

// piece store errors
var (
	NotSupportedMethod    = errors.New("Not supported method")
	NotSupportedDelimiter = errors.New("Not supported delimiter")
	EmptyObjectKey        = errors.New("Object key cannot be empty")
	EmptyMemoryObject     = errors.New("Memory object is empty")
	BucketNotExisted      = errors.New("Bucket not existed")

	ErrInternalError    = errors.New("internal error")
	ErrDuplicateBucket  = errors.New("duplicate bucket")
	ErrDuplicateObject  = errors.New("duplicate object")
	ErrObjectTxNotExist = errors.New("object tx not exist")
	ErrObjectNotExist   = errors.New("object not exist")
)

// stone hub service errors
var (
	ErrTxHash                   = errors.New("tx hash format error")
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
	ErrIntegrityHash  = errors.New("integrity hash of secondary sp is not equal to integrity hash of primary sp")
	ErrRedundancyType = errors.New("unknown redundancy type")
)

func MakeErrMsgResponse(err error) *service.ErrMessage {
	return &service.ErrMessage{
		ErrCode: service.ErrCode_ERR_CODE_ERROR,
		ErrMsg:  err.Error(),
	}
}
