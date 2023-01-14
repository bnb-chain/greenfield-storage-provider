package gateway

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// getObjectOption is the getObject Option.
type getObjectOption struct {
	requestContext *requestContext
	offset         uint64
	length         uint64
}

// downloaderClientInterface define interface to download object.
type downloaderClientInterface interface {
	getObject(name string, writer io.Writer, option *getObjectOption) error
}

// debugDownloaderImpl is an implement of download for local debugging.
type debugDownloaderImpl struct {
	localDir string
}

// getObject is used to get object data from local directory file for debugging.
func (ddl *debugDownloaderImpl) getObject(objectName string, writer io.Writer, option *getObjectOption) error {
	var (
		bucketDir      = ddl.localDir + "/" + option.requestContext.bucketName
		objectDataFile = bucketDir + "/" + objectName + ".data"
		buf            = make([]byte, 65536)
		readN, writeN  int
		size           uint64
		err            error
		msg            string
	)
	defer func() {
		if err != nil {
			log.Warnw("get object failed", "err", err, "msg", msg)
		}
	}()

	if s, err := os.Stat(bucketDir); err != nil || !s.IsDir() {
		msg = "bucket dir is not found"
		return errors.ErrInternalError
	}
	if _, err = os.Stat(objectDataFile); err != nil && os.IsNotExist(err) {
		msg = "object is not found"
		return errors.ErrObjectNotExist
	}

	if f, innerErr := os.OpenFile(objectDataFile, os.O_RDONLY, 0777); innerErr != nil {
		msg = "failed to open object file"
		return errors.ErrInternalError
	} else {
		defer f.Close()
		reader := bufio.NewReader(f)

		for {
			if readN, innerErr = reader.Read(buf); innerErr != nil && innerErr != io.EOF {
				msg = "reader is failed to read"
				return errors.ErrInternalError
			}
			if readN > 0 {
				if writeN, innerErr = writer.Write(buf[:readN]); innerErr != nil {
					msg = "writer is failed to write"
					return errors.ErrInternalError
				}
				size += uint64(writeN)
			}
			if innerErr == io.EOF {
				innerErr = nil
				break
			}
		}
		return nil
	}

}

// downloaderClientConfig is the configuration information when creating downloaderClient.
// currently Mode only support "DebugMode".
type downloadProcessorConfig struct {
	Mode     string
	DebugDir string
	Address  string
}

var defaultDownloadProcessorConfig = &downloadProcessorConfig{
	Mode:     "DebugMode",
	DebugDir: "./debug",
	Address:  "127.0.0.1:5523",
}

// downloaderClient is a wrapper of download object.
type downloadProcessor struct {
	impl downloaderClientInterface
}

// newDownloaderClient return a downloader wrapper.
func newDownloadProcessor(c *downloadProcessorConfig) (*downloadProcessor, error) {
	switch {
	case c.Mode == "DebugMode":
		return &downloadProcessor{impl: &debugDownloaderImpl{localDir: c.DebugDir}}, nil
	case c.Mode == "GrpcMode":
		downloader, err := client.NewDownloaderClient(c.Address)
		if err != nil {
			return nil, err
		}
		return &downloadProcessor{impl: &grpcDownloaderImpl{downloader: downloader}}, nil
	default:
		return nil, fmt.Errorf("not support mode, %v", c.Mode)
	}
}

// getObject get object from downloader.
func (dc *downloadProcessor) getObject(objectName string, writer io.Writer, option *getObjectOption) error {
	return dc.impl.getObject(objectName, writer, option)
}

// grpcDownloaderImpl is an implement of call grpc downloader service.
type grpcDownloaderImpl struct {
	downloader *client.DownloaderClient
}

// getObject get object from downloader.
func (d *grpcDownloaderImpl) getObject(objectName string, writer io.Writer, option *getObjectOption) error {
	var (
		err           error
		readN, writeN int
		size          int
	)

	req := &service.DownloaderServiceDownloaderObjectRequest{
		TraceId:    option.requestContext.requestID,
		BucketName: option.requestContext.bucketName,
		ObjectName: objectName,
	}
	ctx := log.Context(context.Background(), req)
	defer func() {
		if err != nil {
			log.Warnw("get object failed", "err", err)
		}
		log.CtxInfow(ctx, "get object", "receiveSize", size)
	}()

	stream, err := d.downloader.DownloaderObject(ctx, req)
	if err != nil {
		return errors.ErrInternalError
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warnw("failed to read stream", "error", err)
			return errors.ErrInternalError
		}
		if res.ErrMessage != nil && res.ErrMessage.ErrCode != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
			err = fmt.Errorf(res.ErrMessage.ErrMsg)
			log.Warnw("failed to read stream", "error", err)
			return errors.ErrInternalError
		}
		if readN = len(res.Data); readN == 0 {
			log.Warnw("download return empty data", "response", res)
			continue
		}
		if writeN, err = writer.Write(res.Data); err != nil {
			err = fmt.Errorf("failed to http body")
			log.Warnw("failed to read stream", "error", err)
			return errors.ErrInternalError
		}
		if readN != writeN {
			err = fmt.Errorf("part failed write http body")
			log.Warnw("failed to read stream", "error", err)
			return errors.ErrInternalError
		}
		size = size + writeN
	}
	// todo: check total length
	return nil
}
