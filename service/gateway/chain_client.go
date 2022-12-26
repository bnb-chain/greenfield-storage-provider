package gateway

import (
	"fmt"
	"os"

	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type createBucketOption struct {
	reqCtx *requestContext
}

// chainClientInterface define some interfaces to maintain metadata in the blockchain.
type chainClientInterface interface {
	createBucket(name string, opt *createBucketOption) error
}

// debugChainImpl is an implement of Chain interface for local debugging.
type debugChainImpl struct {
	localDir string
}

// createBucket is used to create bucket directory for local debugging.
func (dci *debugChainImpl) createBucket(name string, opt *createBucketOption) error {
	var (
		innerErr error
		msg      string
	)
	defer func() {
		if innerErr != nil {
			log.Warnw("create bucket failed", "err", innerErr, "msg", msg)
		}
	}()

	if s, innerErr := os.Stat(dci.localDir); innerErr != nil || !s.IsDir() {
		msg = "failed to stat"
		return errors.ErrInternalError
	}
	if innerErr = os.Mkdir(dci.localDir+"/"+name, 0777); os.IsExist(innerErr) {
		msg = "bucket has existed"
		return errors.ErrDuplicateBucket
	}
	if innerErr != nil {
		msg = "failed to mkdir bucket"
		return errors.ErrInternalError
	}
	return nil
}

// chainClientConfig is the configuration information when creating chainClient.
// currently Mode only support "DebugMode".
type chainClientConfig struct {
	Mode     string
	DebugDir string
}

// chainClient is a wrapper of maintaining metadata in the blockchain.
// todo: impl of call UpdateChainService, forward msg to blockchain.
type chainClient struct {
	impl chainClientInterface
}

func newChainClient(c chainClientConfig) (*chainClient, error) {
	switch {
	case c.Mode == "DebugMode":
		if c.DebugDir == "" {
			return nil, fmt.Errorf("has no debug dir")
		}
		if err := os.Mkdir(c.DebugDir, 0777); err != nil && !os.IsExist(err) {
			log.Warnw("failed to make debug dir", "err", err)
			return nil, err
		}
		return &chainClient{impl: &debugChainImpl{localDir: c.DebugDir}}, nil
	default:
		return nil, fmt.Errorf("not support mode, %v", c.Mode)
	}
}

func (cc *chainClient) createBucket(name string, opt *createBucketOption) error {
	return cc.impl.createBucket(name, opt)
}
