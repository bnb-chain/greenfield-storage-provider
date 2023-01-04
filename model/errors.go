package model

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
)

// stone hub service errors
var (
	ErrTxHash                   = errors.New("tx hash format error")
	ErrUploadPayloadJobDone     = errors.New("upload payload job is already completed")
	ErrUploadPayloadJobRunning  = errors.New("upload payload job is running")
	ErrObjectInfoOnInscription  = errors.New("object info not on the inscription")
	ErrUploadPayloadJobNotExist = errors.New("upload payload job not exist")
)

func MakeErrMsgResponse(err error) *service.ErrMessage {
	return &service.ErrMessage{
		ErrCode: service.ErrCode_ERR_CODE_ERROR,
		ErrMsg:  err.Error(),
	}
}
