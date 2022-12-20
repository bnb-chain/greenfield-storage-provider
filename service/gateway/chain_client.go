package gateway

import (
	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"os"
)

type debugChainImpl struct {
}

func (dci *debugChainImpl) createBucket(name string, opt *createBucketOption) error {
	s, err := os.Stat(opt.debugDir)
	if err != nil || !s.IsDir() {
		log.Warnw("create bucket failed, due to stat debug dir", "err", err)
		return errors.ErrInternalError
	}
	err = os.Mkdir(opt.debugDir+"/"+name, 0777)
	if os.IsExist(err) {
		log.Warn("create bucket failed, due to bucket has existed")
		return errors.ErrDuplicateBucket
	}
	if err != nil {
		log.Warnw("create bucket failed, due to mkdir", "err", err)
		return errors.ErrInternalError
	}
	return nil
}

// todo: impl of call UpdateChainService
type chainClient struct {
	// forward msg to blockchain
}

func newChainClient() *chainClient {
	return &chainClient{}
}

type createBucketOption struct {
	reqCtx   *requestContext
	debugDir string
}

func (cc *chainClient) createBucket(name string, opt *createBucketOption) error {
	if opt.debugDir != "" {
		dci := &debugChainImpl{}
		return dci.createBucket(name, opt)
	}
	return nil
}
