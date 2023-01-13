package gateway

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	pbPkg "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/service/client"
	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
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
	weight uint64
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

// uploaderClientInterface define interface to upload object. BFS upload process is divided into two stages:
// 1.putObjectTx: set object meta to blockchain;
// 2.putObject: write object data to BFS, and update object seal info to blockchain.
type uploaderClientInterface interface {
	putObjectTx(string, *putObjectTxOption) (*objectTxInfo, error)
	putObject(string, io.Reader, *putObjectOption) (*objectInfo, error)
}

// debugUploaderImpl is an implement of upload for local debugging.
type debugUploaderImpl struct {
	localDir string
}

// putObjectTx is used to put object tx to local directory file for debugging.
func (dui *debugUploaderImpl) putObjectTx(objectName string, option *putObjectTxOption) (*objectTxInfo, error) {
	var (
		innerErr     error
		bucketDir    = dui.localDir + "/" + option.requestContext.objectName
		objectTxFile = bucketDir + "/" + objectName + ".tx"
		txJson       []byte
	)
	defer func() {
		if innerErr != nil {
			log.Warnw("put object tx failed", "err", innerErr)
		}
	}()

	if s, innerErr := os.Stat(bucketDir); innerErr != nil || !s.IsDir() {
		return nil, errors.ErrInternalError
	}
	if _, innerErr = os.Stat(objectTxFile); innerErr == nil {
		return nil, errors.ErrDuplicateObject
	}
	// mock tx info
	var txInfo = struct {
		TxHash string `json:"TxHash"`
		Weight uint64 `json:"Weight"`
	}{
		TxHash: "debugmode-hash",
		Weight: 2012,
	}
	if txJson, innerErr = json.Marshal(txInfo); innerErr != nil {
		return nil, errors.ErrInternalError
	}
	if f, innerErr := os.OpenFile(objectTxFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777); innerErr != nil {
		return nil, errors.ErrInternalError
	} else {
		defer f.Close()
		if n, innerErr := f.Write(txJson); innerErr == nil && n < len(txJson) {
			return nil, errors.ErrInternalError
		}
		return &objectTxInfo{txHash: []byte(txInfo.TxHash), weight: txInfo.Weight}, nil
	}
}

// putObject is used to put object data to local directory file for debugging.
func (dui *debugUploaderImpl) putObject(objectName string, reader io.Reader, option *putObjectOption) (*objectInfo, error) {
	var (
		innerErr       error
		bucketDir      = dui.localDir + "/" + option.requestContext.bucketName
		objectTxFile   = bucketDir + "/" + objectName + ".tx"
		objectDataFile = bucketDir + "/" + objectName + ".data"

		buf           = make([]byte, 65536)
		readN, writeN int
		size          uint64
		hashBuf       = make([]byte, 65536)
		md5Hash       = md5.New()
		md5Value      string
	)
	defer func() {
		if innerErr != nil {
			log.Warnw("put object failed", "err", innerErr)
		}
	}()

	if s, innerErr := os.Stat(bucketDir); innerErr != nil || !s.IsDir() {
		return nil, errors.ErrInternalError
	}
	if _, innerErr = os.Stat(objectTxFile); innerErr != nil && os.IsNotExist(innerErr) {
		return nil, errors.ErrObjectTxNotExist
	}

	// todo: check tx-hash by json
	if f, innerErr := os.OpenFile(objectDataFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777); innerErr != nil {
		return nil, errors.ErrInternalError
	} else {
		defer f.Close()
		writer := bufio.NewWriter(f)
		for {
			if readN, innerErr = reader.Read(buf); innerErr != nil && innerErr != io.EOF {
				return nil, errors.ErrInternalError
			}
			if readN > 0 {
				if writeN, innerErr = writer.Write(buf[:readN]); innerErr != nil {
					return nil, errors.ErrInternalError
				}
				writer.Flush()
				size += uint64(writeN)
				copy(hashBuf, buf[:readN])
				md5Hash.Write(hashBuf[:readN])
			}
			if innerErr == io.EOF {
				innerErr = nil
				break
			}
		}
		md5Value = hex.EncodeToString(md5Hash.Sum(nil))
		return &objectInfo{eTag: md5Value, size: size}, nil
	}
}

// grpcUploaderImpl is an implement of call grpc uploader service.
type grpcUploaderImpl struct {
	uploader *client.UploaderClient
}

// putObjectTx is used to call uploaderService's CreateObject by grpc.
func (gui *grpcUploaderImpl) putObjectTx(objectName string, option *putObjectTxOption) (*objectTxInfo, error) {
	log.Infow("put object tx", "option", option)
	resp, err := gui.uploader.CreateObject(context.Background(), &pbService.UploaderServiceCreateObjectRequest{
		TraceId: option.requestContext.requestID,
		ObjectInfo: &pbPkg.ObjectInfo{
			BucketName:     option.requestContext.bucketName,
			ObjectName:     objectName,
			Size:           option.objectSize,
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
func (gui *grpcUploaderImpl) putObject(objectName string, reader io.Reader, option *putObjectOption) (*objectInfo, error) {
	var (
		buf      = make([]byte, 65536)
		readN    int
		size     uint64
		hashBuf  = make([]byte, 65536)
		md5Hash  = md5.New()
		md5Value string
	)

	stream, err := gui.uploader.UploadPayload(context.Background())
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

			req := &pbService.UploaderServiceUploadPayloadRequest{
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
			err = nil
			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Warnw("put object failed, due to stream close", "err", err)
				return nil, errors.ErrInternalError
			}
			if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != pbService.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
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

// getAuthentication is used to call uploaderService's getAuthentication by grpc.
func (gui *grpcUploaderImpl) getAuthentication(option *getAuthenticationOption) (*authenticationInfo, error) {
	resp, err := gui.uploader.GetAuthentication(context.Background(), &pbService.UploaderServiceGetAuthenticationRequest{
		TraceId: option.requestContext.requestID,
		Bucket:  option.requestContext.bucketName,
		Object:  option.requestContext.objectName,
		Action:  option.requestContext.actionName,
	})
	if err != nil {
		log.Warnw("failed to rpc to uploader", "err", err)
		return nil, errors.ErrInternalError
	}
	return &authenticationInfo{preSignature: resp.PreSignature}, nil
}

// putObjectV2 copy from putObject.
func (gui *grpcUploaderImpl) putObjectV2(objectName string, reader io.Reader, option *putObjectOption) (*objectInfo, error) {
	var (
		buf      = make([]byte, 65536)
		readN    int
		size     uint64
		hashBuf  = make([]byte, 65536)
		md5Hash  = md5.New()
		md5Value string
	)

	stream, err := gui.uploader.UploadPayloadV2(context.Background())
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
			req := &pbService.UploaderServiceUploadPayloadV2Request{
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
			err = nil
			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Warnw("put object failed, due to stream close", "err", err)
				return nil, errors.ErrInternalError
			}
			if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != pbService.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
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

// uploadProcessorConfig is the configuration information when creating uploaderClient.
// currently Mode support "DebugMode" and "GrpcMode".
type uploadProcessorConfig struct {
	Mode     string
	DebugDir string
	Address  string
}

var defaultUploadProcessorConfig = &uploadProcessorConfig{
	Mode:     "DebugMode",
	DebugDir: "./debug",
	Address:  "127.0.0.1:5311",
}

// uploadProcessor is a wrapper of uploader client.
type uploadProcessor struct {
	impl uploaderClientInterface
}

// newUploaderClient return a uploaderClient.
func newUploadProcessor(c *uploadProcessorConfig) (*uploadProcessor, error) {
	switch {
	case c.Mode == "DebugMode":
		if c.DebugDir == "" {
			return nil, fmt.Errorf("has no debug dir")
		}
		if err := os.Mkdir(c.DebugDir, 0777); err != nil && !os.IsExist(err) {
			log.Warnw("failed to make debug dir", "err", err)
			return nil, err
		}
		return &uploadProcessor{impl: &debugUploaderImpl{localDir: c.DebugDir}}, nil
	case c.Mode == "GrpcMode":
		u, err := client.NewUploaderClient(c.Address)
		if err != nil {
			return nil, err
		}
		return &uploadProcessor{impl: &grpcUploaderImpl{uploader: u}}, nil
	default:
		return nil, fmt.Errorf("not support mode, %v", c.Mode)
	}
}

// putObjectTx call uploaderClient putObjectTx interface.
func (up *uploadProcessor) putObjectTx(objectName string, option *putObjectTxOption) (objectInfoTx *objectTxInfo, err error) {
	return up.impl.putObjectTx(objectName, option)
}

// putObject call uploaderClient putObject interface.
func (up *uploadProcessor) putObject(objectName string, reader io.Reader, option *putObjectOption) (*objectInfo, error) {
	return up.impl.putObject(objectName, reader, option)
}

type getAuthenticationOption struct {
	requestContext *requestContext
}
type authenticationInfo struct {
	preSignature []byte
}

// getAuthentication call uploaderService getAuthentication interface.
func (up *uploadProcessor) getAuthentication(option *getAuthenticationOption) (*authenticationInfo, error) {
	if p, ok := up.impl.(*grpcUploaderImpl); ok {
		return p.getAuthentication(option)
	}
	return nil, fmt.Errorf("not supported")
}

// putObjectV2 call uploaderService putObjectV2 interface.
func (up *uploadProcessor) putObjectV2(objectName string, reader io.Reader, option *putObjectOption) (*objectInfo, error) {
	if p, ok := up.impl.(*grpcUploaderImpl); ok {
		return p.putObjectV2(objectName, reader, option)
	}
	return nil, fmt.Errorf("not supported")
}

// Close release uploadProcessor resource.
func (up *uploadProcessor) Close() error {
	if p, ok := up.impl.(*grpcUploaderImpl); ok {
		return p.uploader.Close()
	}
	return nil
}
