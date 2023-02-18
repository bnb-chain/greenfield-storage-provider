package gateway

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// putObjectTxOption is the putObjectTx Option.
type putObjectTxOption struct {
	requestContext *requestContext
	objectSize     uint64
	contentType    string
	checksum       []byte
	isPrivate      bool
	redundancyType string
}

// objectTxInfo is the return of putObjectTx.
type objectTxInfo struct {
	txHash []byte
	// weight uint64
}

// putObjectOption is the putObject Option.
type putObjectOption struct {
	requestContext *requestContext
	txHash         []byte
	size           uint64
	redundancyType string
}

// objectInfo is the return of putObject.
type objectInfo struct {
	size uint64
	eTag string
}

// getApprovalOption is the getApproval Option.
type getApprovalOption struct {
	requestContext *requestContext
}

// approvalInfo is the return of getApproval
type approvalInfo struct {
	preSignature []byte
}

// uploadProcessor is a wrapper of uploader client.
type uploadProcessor struct {
	uploader *client.UploaderClient
}

// newUploaderClient return a uploaderClient.
func newUploadProcessor(addr string) (*uploadProcessor, error) {
	u, err := client.NewUploaderClient(addr)
	if err != nil {
		return nil, err
	}
	return &uploadProcessor{uploader: u}, nil
}

// putObjectTx is used to call uploaderService's CreateObject by grpc.
func (up *uploadProcessor) putObjectTx(objectName string, option *putObjectTxOption) (*objectTxInfo, error) {
	log.Infow("put object tx", "option", option)
	resp, err := up.uploader.CreateObject(context.Background(), &stypes.UploaderServiceCreateObjectRequest{
		TraceId: option.requestContext.requestID,
		ObjectInfo: &ptypes.ObjectInfo{
			BucketName:     option.requestContext.bucketName,
			ObjectName:     objectName,
			Size_:          option.objectSize,
			ContentType:    option.contentType,
			Checksum:       option.checksum,
			IsPrivate:      option.isPrivate,
			RedundancyType: redundancyTypeToEnum(option.redundancyType),
		},
	})
	if err != nil {
		log.Warnw("failed to rpc to uploader", "err", err)
		return nil, errors.ErrInternalError
	}
	return &objectTxInfo{txHash: resp.TxHash}, nil
}

// putObject is used to call uploaderService's UploadPayload by grpc.
func (up *uploadProcessor) putObject(objectName string, reader io.Reader, option *putObjectOption) (*objectInfo, error) {
	var (
		buf      = make([]byte, 65536)
		readN    int
		size     uint64
		hashBuf  = make([]byte, 65536)
		md5Hash  = md5.New()
		md5Value string
	)

	stream, err := up.uploader.UploadPayload(context.Background())
	if err != nil {
		log.Warnw("failed to dail to uploader", "err", err)
		return nil, errors.ErrInternalError
	}
	for {
		readN, err = reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Warnw("put object failed, due to reader", "err", err)
			return nil, errors.ErrInternalError
		}
		if readN > 0 {

			req := &stypes.UploaderServiceUploadPayloadRequest{
				TraceId:     option.requestContext.requestID,
				TxHash:      option.txHash,
				PayloadData: buf[:readN],
			}
			if err := stream.Send(req); err != nil {
				log.Warnw("put object failed, due to stream send", "err", err)
				return nil, errors.ErrInternalError
			}
			size += uint64(readN)
			copy(hashBuf, buf[:readN])
			md5Hash.Write(hashBuf[:readN])
		}
		if err == io.EOF {
			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Warnw("put object failed, due to stream close", "err", err)
				return nil, errors.ErrInternalError
			}
			if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
				log.Warnw("failed to grpc", "err", resp.ErrMessage)
				return nil, fmt.Errorf(resp.ErrMessage.ErrMsg)
			}
			break
		}
	}
	md5Value = hex.EncodeToString(md5Hash.Sum(nil))
	log.Info("gateway total size:", size)
	return &objectInfo{eTag: md5Value, size: size}, nil
}

// getApproval is used to call uploaderService's getApproval by grpc.
func (up *uploadProcessor) getApproval(option *getApprovalOption) (*approvalInfo, error) {
	resp, err := up.uploader.GetApproval(context.Background(), &stypes.UploaderServiceGetApprovalRequest{
		TraceId: option.requestContext.requestID,
		Bucket:  option.requestContext.bucketName,
		Object:  option.requestContext.objectName,
		Action:  option.requestContext.actionName,
	})
	if err != nil {
		log.Warnw("failed to rpc to uploader", "err", err)
		return nil, errors.ErrInternalError
	}
	return &approvalInfo{preSignature: resp.PreSignature}, nil
}

// putObjectV2 copy from putObject.
func (up *uploadProcessor) putObjectV2(objectName string, reader io.Reader, option *putObjectOption) (*objectInfo, error) {
	var (
		buf      = make([]byte, 65536)
		readN    int
		size     uint64
		hashBuf  = make([]byte, 65536)
		md5Hash  = md5.New()
		md5Value string
	)

	stream, err := up.uploader.UploadPayloadV2(context.Background())
	if err != nil {
		log.Warnw("failed to dail to uploader", "err", err)
		return nil, errors.ErrInternalError
	}
	for {
		readN, err = reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Warnw("put object failed, due to reader", "err", err)
			return nil, errors.ErrInternalError
		}
		if readN > 0 {
			req := &stypes.UploaderServiceUploadPayloadV2Request{
				TraceId:        option.requestContext.requestID,
				TxHash:         option.txHash,
				PayloadData:    buf[:readN],
				BucketName:     option.requestContext.bucketName,
				ObjectName:     objectName,
				ObjectSize:     option.size,
				RedundancyType: redundancyTypeToEnum(option.redundancyType),
			}
			if err := stream.Send(req); err != nil {
				log.Warnw("put object failed, due to stream send", "err", err)
				return nil, errors.ErrInternalError
			}
			size += uint64(readN)
			copy(hashBuf, buf[:readN])
			md5Hash.Write(hashBuf[:readN])
		}
		if err == io.EOF {
			if size == 0 {
				log.Warnw("put object failed, due to payload is empty")
				return nil, errors.ErrObjectIsEmpty
			}
			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Warnw("put object failed, due to stream close", "err", err)
				return nil, errors.ErrInternalError
			}
			if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
				log.Warnw("failed to grpc", "err", resp.ErrMessage)
				return nil, fmt.Errorf(resp.ErrMessage.ErrMsg)
			}
			break
		}
	}
	md5Value = hex.EncodeToString(md5Hash.Sum(nil))
	log.Info("gateway total size:", size)
	return &objectInfo{eTag: md5Value, size: size}, nil
}
