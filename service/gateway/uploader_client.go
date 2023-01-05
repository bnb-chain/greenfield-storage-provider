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

	"google.golang.org/grpc"

	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	pbPkg "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// putObjectTxOption is the putObjectTx Option.
type putObjectTxOption struct {
	reqCtx      *requestContext
	size        uint64
	contentType string
	checksum    []byte
	isPrivate   bool
}

// objectTxInfo is the return of putObjectTx.
type objectTxInfo struct {
	txHash string
	weight uint64
}

// putObjectOption is the putObject Option.
type putObjectOption struct {
	reqCtx *requestContext
	txHash []byte
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
func (dui *debugUploaderImpl) putObjectTx(name string, opt *putObjectTxOption) (*objectTxInfo, error) {
	var (
		innerErr     error
		bucketDir    = dui.localDir + "/" + opt.reqCtx.bucket
		objectTxFile = bucketDir + "/" + name + ".tx"
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
		return &objectTxInfo{txHash: txInfo.TxHash, weight: txInfo.Weight}, nil
	}
}

// putObject is used to put object data to local directory file for debugging.
func (dui *debugUploaderImpl) putObject(name string, reader io.Reader, opt *putObjectOption) (*objectInfo, error) {
	var (
		innerErr       error
		bucketDir      = dui.localDir + "/" + opt.reqCtx.bucket
		objectTxFile   = bucketDir + "/" + name + ".tx"
		objectDataFile = bucketDir + "/" + name + ".data"

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
	Address string
}

// putObjectTx is used to call uploaderService's CreateObject by grpc.
func (gui *grpcUploaderImpl) putObjectTx(name string, opt *putObjectTxOption) (*objectTxInfo, error) {
	conn, err := grpc.Dial(gui.Address, grpc.WithInsecure())
	if err != nil {
		log.Warnw("failed to dail to uploader", "err", err)
		return nil, errors.ErrInternalError
	}
	defer conn.Close()
	client := pbService.NewUploaderServiceClient(conn)
	// todo: fill more info
	resp, err := client.CreateObject(context.Background(), &pbService.UploaderServiceCreateObjectRequest{
		ObjectInfo: &pbPkg.ObjectInfo{
			BucketName:  opt.reqCtx.bucket,
			ObjectName:  name,
			Size:        opt.size,
			ContentType: opt.contentType,
			Checksum:    opt.checksum,
			IsPrivate:   opt.isPrivate,
		},
	})
	if err != nil {
		log.Warnw("failed to rpc to uploader", "err", err)
		return nil, errors.ErrInternalError
	}
	if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != pbService.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Warnw("failed to grpc", "err", resp.ErrMessage)
		return nil, fmt.Errorf(resp.ErrMessage.ErrMsg)
	}
	return &objectTxInfo{txHash: string(resp.TxHash)}, nil
}

// putObject is used to call uploaderService's UploadPayload by grpc.
func (gui *grpcUploaderImpl) putObject(name string, reader io.Reader, opt *putObjectOption) (*objectInfo, error) {
	var (
		buf      = make([]byte, 65536)
		readN    int
		size     uint64
		hashBuf  = make([]byte, 65536)
		md5Hash  = md5.New()
		md5Value string
	)

	conn, err := grpc.Dial(gui.Address, grpc.WithInsecure())
	if err != nil {
		log.Warnw("failed to dail to uploader", "err", err)
		return nil, errors.ErrInternalError
	}
	defer conn.Close()
	client := pbService.NewUploaderServiceClient(conn)
	stream, err := client.UploadPayload(context.Background())
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
			var req pbService.UploaderServiceUploadPayloadRequest
			req.TxHash = opt.txHash
			req.PayloadData = buf[:readN]
			// todo: fill job_id??
			if err := stream.Send(&req); err != nil {
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

// uploaderClientConfig is the configuration information when creating uploaderClient.
// currently Mode support "DebugMode" and "GrpcMode".
type uploaderClientConfig struct {
	Mode     string
	DebugDir string
	Address  string
}

// uploaderClient is a wrapper of uploader.
type uploaderClient struct {
	impl uploaderClientInterface
}

// newUploaderClient return a uploaderClient.
func newUploaderClient(c uploaderClientConfig) (*uploaderClient, error) {
	switch {
	case c.Mode == "DebugMode":
		if c.DebugDir == "" {
			return nil, fmt.Errorf("has no debug dir")
		}
		if err := os.Mkdir(c.DebugDir, 0777); err != nil && !os.IsExist(err) {
			log.Warnw("failed to make debug dir", "err", err)
			return nil, err
		}
		return &uploaderClient{impl: &debugUploaderImpl{localDir: c.DebugDir}}, nil
	case c.Mode == "GrpcMode":
		return &uploaderClient{impl: &grpcUploaderImpl{Address: c.Address}}, nil
	default:
		return nil, fmt.Errorf("not support mode, %v", c.Mode)
	}
}

// putObjectTx call uploader's putObjectTx interface.
func (uc *uploaderClient) putObjectTx(name string, opt *putObjectTxOption) (objectInfoTx *objectTxInfo, err error) {
	return uc.impl.putObjectTx(name, opt)
}

// putObject call uploader's putObject interface.
func (uc *uploaderClient) putObject(name string, reader io.Reader, opt *putObjectOption) (*objectInfo, error) {
	return uc.impl.putObject(name, reader, opt)
}
