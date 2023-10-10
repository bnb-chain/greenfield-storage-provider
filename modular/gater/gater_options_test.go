package gater

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	mwhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
)

func TestNewGateModularSuccess(t *testing.T) {
	app := &gfspapp.GfSpBaseApp{}
	cfg := &gfspconfig.GfSpConfig{}
	result, err := NewGateModular(app, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestNewGateModularFailure(t *testing.T) {
	app := &gfspapp.GfSpBaseApp{}
	apiLimits := mwhttp.KeyToRateLimiterNameCell{
		Key:  "test_api_limit",
		Name: "ApiLimit",
	}
	apiList := make([]mwhttp.KeyToRateLimiterNameCell, 0)
	apiList = append(apiList, apiLimits)
	cfg := &gfspconfig.GfSpConfig{
		APIRateLimiter: mwhttp.RateLimiterConfig{
			APILimits: apiList,
		},
	}
	result, err := NewGateModular(app, cfg)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func Test_makeAPIRateLimitCfg(t *testing.T) {
	pathPattern := mwhttp.KeyToRateLimiterNameCell{
		Key:  "test_path_pattern",
		Name: "PathPattern",
	}
	pathList := make([]mwhttp.KeyToRateLimiterNameCell, 0)
	pathList = append(pathList, pathPattern)

	hostPattern := mwhttp.KeyToRateLimiterNameCell{
		Key:  "test_path_pattern",
		Name: "PathPattern",
	}
	hostList := make([]mwhttp.KeyToRateLimiterNameCell, 0)
	hostList = append(hostList, hostPattern)

	apiLimits := mwhttp.KeyToRateLimiterNameCell{
		Key:  "test_api_limit",
		Name: "ApiLimit",
	}
	apiList := make([]mwhttp.KeyToRateLimiterNameCell, 0)
	apiList = append(apiList, apiLimits)
	cfg := mwhttp.RateLimiterConfig{
		PathPattern: pathList,
		HostPattern: hostList,
		APILimits:   apiList,
	}
	result := makeAPIRateLimitCfg(cfg)
	assert.NotNil(t, result)
}
