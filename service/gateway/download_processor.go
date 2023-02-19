package gateway

import (
	"context"
	"fmt"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/client"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// getObjectOption is the getObject Option.
type getObjectOption struct {
	requestContext *requestContext
	// offset         uint64
	// length         uint64
}

// downloaderClient is a wrapper of download object.
type downloadProcessor struct {
	downloader *client.DownloaderClient
}

// newDownloaderClient return a downloader wrapper.
func newDownloadProcessor(addr string) (*downloadProcessor, error) {
	downloader, err := client.NewDownloaderClient(addr)
	if err != nil {
		return nil, err
	}
	return &downloadProcessor{downloader: downloader}, nil
}

// getObject get object from downloader.
func (dp *downloadProcessor) getObject(objectName string, writer io.Writer, option *getObjectOption) error {
	var (
		err           error
		readN, writeN int
		size          int
	)

	req := &stypes.DownloaderServiceDownloaderObjectRequest{
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

	stream, err := dp.downloader.DownloaderObject(ctx, req)
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
		if res.ErrMessage != nil && res.ErrMessage.ErrCode != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
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
