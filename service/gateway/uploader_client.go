package gateway

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"io"
	"os"
)

type debugUploaderImpl struct {
}

func (uc *debugUploaderImpl) putObjectTx(name string, opt *putObjectTxOption) (objectInfoTx *objectTxInfo, err error) {
	var (
		bucketDir    = opt.debugDir + "/" + opt.reqCtx.bucket
		objectTxFile = bucketDir + "/" + name + ".tx"
	)

	s, err := os.Stat(bucketDir)
	if err != nil || !s.IsDir() {
		log.Warnw("put object tx failed, due to stat bucket dir", "err", err)
		return nil, errors.ErrInternalError
	}
	_, err = os.Stat(objectTxFile)
	if err == nil {
		log.Warn("put object tx failed, due to has existed")
		return nil, errors.ErrDuplicateObject
	}

	var txInfo = struct {
		TxHash string `json:"TxHash"`
		Weight uint64 `json:"Weight"`
	}{
		TxHash: "mockhash-123",
		Weight: 2012,
	}

	txJson, err := json.Marshal(txInfo)
	if err != nil {
		log.Warnw("put object tx failed, due to json marshal", "err", err)
		return nil, errors.ErrInternalError
	}

	f, err := os.OpenFile(objectTxFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	defer f.Close()
	if err != nil {
		log.Warnw("put object tx failed, due to open file", "err", err)
		return nil, errors.ErrInternalError
	}
	n, err := f.Write(txJson)
	if err == nil && n < len(txJson) {
		log.Warnw("put object tx failed, due to write file", "err", err)
		return nil, errors.ErrInternalError
	}
	return &objectTxInfo{txHash: txInfo.TxHash, weight: txInfo.Weight}, nil
}

func (dui *debugUploaderImpl) putObject(name string, reader io.Reader, opt *putObjectOption) (*objectInfo, error) {
	var (
		bucketDir      = opt.debugDir + "/" + opt.reqCtx.bucket
		objectTxFile   = bucketDir + "/" + name + ".tx"
		objectDataFile = bucketDir + "/" + name + ".data"

		buf           = make([]byte, 65536)
		readN, writeN int
		size          uint64
		hashBuf       = make([]byte, 65536)
		md5Hash       = md5.New()
		md5Value      string
	)

	s, err := os.Stat(bucketDir)
	if err != nil || !s.IsDir() {
		log.Warnw("put object failed, due to stat bucket dir", "err", err)
		return nil, errors.ErrInternalError
	}
	_, err = os.Stat(objectTxFile)
	if err != nil && os.IsNotExist(err) {
		log.Warn("put object failed, due to tx is not existed")
		return nil, errors.ErrObjectTxNotExist
	}

	// todo: check tx-hash by json
	f, err := os.OpenFile(objectDataFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	defer f.Close()
	if err != nil {
		log.Warnw("put object failed, due to open file", "err", err)
		return nil, errors.ErrInternalError
	}
	writer := bufio.NewWriter(f)

	for {
		readN, err = reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Warnw("put object failed, due to reader", "err", err)
			return nil, errors.ErrInternalError
		}
		if readN > 0 {
			if writeN, err = writer.Write(buf[:readN]); err != nil {
				log.Warnw("put object failed, due to writer", "err", err)
				return nil, errors.ErrInternalError
			}
			writer.Flush()
			size += uint64(writeN)
			copy(hashBuf, buf[:readN])
			md5Hash.Write(hashBuf[:readN])
		}
		if err == io.EOF {
			err = nil
			break
		}
	}
	md5Value = hex.EncodeToString(md5Hash.Sum(nil))
	return &objectInfo{eTag: md5Value, size: size}, nil
}

// todo: impl of call UploaderService
type uploaderClient struct {
}

func newUploaderClient() *uploaderClient {
	return &uploaderClient{}
}

type putObjectTxOption struct {
	reqCtx   *requestContext
	debugDir string
}

type objectTxInfo struct {
	txHash string
	weight uint64
}

func (uc *uploaderClient) putObjectTx(name string, opt *putObjectTxOption) (objectInfoTx *objectTxInfo, err error) {
	if opt.debugDir != "" {
		dui := &debugUploaderImpl{}
		return dui.putObjectTx(name, opt)
	}
	return &objectTxInfo{}, nil
}

// todo: pick policy
type putObjectOption struct {
	reqCtx   *requestContext
	debugDir string
}

type objectInfo struct {
	size uint64
	eTag string
}

// todo: check md5
func (uc *uploaderClient) putObject(name string, reader io.Reader, opt *putObjectOption) (*objectInfo, error) {
	if opt.debugDir != "" {
		dui := &debugUploaderImpl{}
		return dui.putObject(name, reader, opt)
	}
	return &objectInfo{eTag: "2012"}, nil
}
