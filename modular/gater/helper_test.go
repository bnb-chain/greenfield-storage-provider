package gater

import (
	"errors"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
)

const (
	testDomain        = "www.route-test.com"
	scheme            = "https://"
	mockBucketName    = "mock-bucket-name"
	mockObjectName    = "mock-object-name"
	invalidBucketName = "1"
	invalidObjectName = ".."
)

var mockErr = errors.New("mock error")

func setup(t *testing.T) *GateModular {
	return &GateModular{
		env:     gfspapp.EnvLocal,
		domain:  testDomain,
		baseApp: &gfspapp.GfSpBaseApp{},
	}
}
