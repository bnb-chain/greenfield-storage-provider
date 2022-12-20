package gateway

import (
	"bufio"
	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"io"
	"os"
)

type debugDownloaderImpl struct {
}

func (ddl *debugDownloaderImpl) getObject(name string, writer io.Writer, opt *getObjectOption) error {
	var (
		bucketDir      = opt.debugDir + "/" + opt.reqCtx.bucket
		objectDataFile = bucketDir + "/" + name + ".data"
		buf            = make([]byte, 65536)
		readN, writeN  int
		size           uint64
	)

	s, err := os.Stat(bucketDir)
	if err != nil || !s.IsDir() {
		log.Warnw("get object failed, due to stat bucket dir", "err", err)
		return errors.ErrInternalError
	}
	_, err = os.Stat(objectDataFile)
	if err != nil && os.IsNotExist(err) {
		log.Warn("get object failed, due to not existed")
		return errors.ErrObjectNotExist
	}

	f, err := os.OpenFile(objectDataFile, os.O_RDONLY, 0777)
	defer f.Close()
	if err != nil {
		log.Warnw("get object failed, due to open file error", "err", err)
		return errors.ErrInternalError
	}
	reader := bufio.NewReader(f)

	for {
		readN, err = reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Warnw("get object failed, due to reader err", "err", err)
			return errors.ErrInternalError
		}
		if readN > 0 {
			if writeN, err = writer.Write(buf[:readN]); err != nil {
				log.Warnw("get object failed, due to writer err", "err", err)
				return errors.ErrInternalError
			}
			size += uint64(writeN)
		}
		if err == io.EOF {
			err = nil
			break
		}
	}
	return nil
}

// todo: impl of call DownloaderService
type downloaderClient struct {
}

func newDownloaderClient() *downloaderClient {
	return &downloaderClient{}
}

type getObjectOption struct {
	reqCtx   *requestContext
	debugDir string
	offset   uint64
	length   uint64
}

func (dc *downloaderClient) getObject(name string, writer io.Writer, opt *getObjectOption) error {
	if opt.debugDir != "" {
		dui := &debugDownloaderImpl{}
		return dui.getObject(name, writer, opt)
	}
	return nil
}
