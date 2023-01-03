package gateway

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// getObjectOption is the getObject Option.
type getObjectOption struct {
	reqCtx *requestContext
	offset uint64
	length uint64
}

// downloaderClientInterface define interface to download object.
type downloaderClientInterface interface {
	getObject(name string, writer io.Writer, opt *getObjectOption) error
}

// debugDownloaderImpl is an implement of download for local debugging.
type debugDownloaderImpl struct {
	localDir string
}

// getObject is used to get object data from local directory file for debugging.
func (ddl *debugDownloaderImpl) getObject(name string, writer io.Writer, opt *getObjectOption) error {
	var (
		bucketDir      = ddl.localDir + "/" + opt.reqCtx.bucket
		objectDataFile = bucketDir + "/" + name + ".data"
		buf            = make([]byte, 65536)
		readN, writeN  int
		size           uint64
		innerErr       error
		msg            string
	)
	defer func() {
		if innerErr != nil {
			log.Warnw("get object failed", "err", innerErr, "msg", msg)
		}
	}()

	if s, innerErr := os.Stat(bucketDir); innerErr != nil || !s.IsDir() {
		msg = "failed to stat bucket dir"
		return errors.ErrInternalError
	}
	if _, innerErr = os.Stat(objectDataFile); innerErr != nil && os.IsNotExist(innerErr) {
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
type downloaderClientConfig struct {
	Mode     string
	DebugDir string
}

// downloaderClient is a wrapper of download object.
// todo: impl of call DownloaderService
type downloaderClient struct {
	impl downloaderClientInterface
}

// newDownloaderClient return a downloader wrapper.
func newDownloaderClient(c downloaderClientConfig) (*downloaderClient, error) {
	switch {
	case c.Mode == "DebugMode":
		return &downloaderClient{impl: &debugDownloaderImpl{localDir: c.DebugDir}}, nil
	default:
		return nil, fmt.Errorf("not support mode, %v", c.Mode)
	}
}

// getObject get object from downloader.
func (dc *downloaderClient) getObject(name string, writer io.Writer, opt *getObjectOption) error {
	return dc.impl.getObject(name, writer, opt)
}
