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
		Key:   "test_api_limit",
		Names: []string{"test_api_limit"},
	}
	nameToLimit := mwhttp.MemoryLimiterConfig{
		Name:       "test_api_limit",
		RateLimit:  1,
		RatePeriod: "A",
	}
	apiList := make([]mwhttp.KeyToRateLimiterNameCell, 0)
	apiList = append(apiList, apiLimits)
	nameToLimitList := make([]mwhttp.MemoryLimiterConfig, 0)
	nameToLimitList = append(nameToLimitList, nameToLimit)
	cfg := &gfspconfig.GfSpConfig{
		APIRateLimiter: mwhttp.RateLimiterConfig{
			APILimits:   apiList,
			NameToLimit: nameToLimitList,
		},
	}
	result, err := NewGateModular(app, cfg)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func Test_makeAPIRateLimitCfg(t *testing.T) {
	pathPattern := mwhttp.KeyToRateLimiterNameCell{
		Key:   "test_path_pattern",
		Names: []string{"PathPattern"},
	}
	pathList := make([]mwhttp.KeyToRateLimiterNameCell, 0)
	pathList = append(pathList, pathPattern)

	hostPattern := mwhttp.KeyToRateLimiterNameCell{
		Key:   "test_path_pattern",
		Names: []string{"PathPattern"},
	}
	hostList := make([]mwhttp.KeyToRateLimiterNameCell, 0)
	hostList = append(hostList, hostPattern)

	apiLimits := mwhttp.KeyToRateLimiterNameCell{
		Key:   "test_api_limit",
		Names: []string{"ApiLimit"},
	}
	apiList := make([]mwhttp.KeyToRateLimiterNameCell, 0)
	apiList = append(apiList, apiLimits)

	nameToLimit := mwhttp.MemoryLimiterConfig{
		Name:       "test_api_limit",
		RateLimit:  1,
		RatePeriod: "H",
	}
	nameToLimitList := make([]mwhttp.MemoryLimiterConfig, 0)
	nameToLimitList = append(nameToLimitList, nameToLimit)
	cfg := mwhttp.RateLimiterConfig{
		PathPattern: pathList,
		HostPattern: hostList,
		APILimits:   apiList,
		NameToLimit: nameToLimitList,
	}
	result := makeAPIRateLimitCfg(cfg)
	assert.NotNil(t, result)
}
