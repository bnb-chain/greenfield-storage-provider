package gater

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	localhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
)

const (
	DefaultGatewayDomain    = "localhost:9133"
	DefaultMaxListReadQuota = 100
	DefaultMaxPayloadSize   = 2 * 1024 * 1024 * 1024
)

func NewGateModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	gater := &GateModular{baseApp: app}
	if err := DefaultGaterOptions(gater, cfg); err != nil {
		return nil, err
	}
	return gater, nil
}

func DefaultGaterOptions(gater *GateModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Gateway.Domain == "" {
		cfg.Gateway.Domain = DefaultGatewayDomain
	}
	if cfg.Gateway.HTTPAddress == "" {
		cfg.Gateway.HTTPAddress = DefaultGatewayDomain
	}
	if cfg.Bucket.MaxListReadQuotaNumber == 0 {
		cfg.Bucket.MaxListReadQuotaNumber = DefaultMaxListReadQuota
	}
	if cfg.Bucket.MaxPayloadSize == 0 {
		cfg.Bucket.MaxPayloadSize = DefaultMaxPayloadSize
	}
	gater.maxPayloadSize = cfg.Bucket.MaxPayloadSize
	gater.domain = cfg.Gateway.Domain
	gater.httpAddress = cfg.Gateway.HTTPAddress
	gater.maxListReadQuota = cfg.Bucket.MaxListReadQuotaNumber
	rateCfg := makeAPIRateLimitCfg(cfg.APIRateLimiter)
	if err := localhttp.NewAPILimiter(rateCfg); err != nil {
		log.Errorw("failed to new api limiter", "err", err)
		return err
	}
	return nil
}

func makeAPIRateLimitCfg(cfg localhttp.RateLimiterConfig) *localhttp.APILimiterConfig {
	defaultMap := make(map[string]localhttp.MemoryLimiterConfig)
	for _, c := range cfg.PathPattern {
		defaultMap[c.Key] = localhttp.MemoryLimiterConfig{
			RateLimit:  c.RateLimit,
			RatePeriod: c.RatePeriod,
		}
	}
	patternMap := make(map[string]localhttp.MemoryLimiterConfig)
	for _, c := range cfg.HostPattern {
		patternMap[c.Key] = localhttp.MemoryLimiterConfig{
			RateLimit:  c.RateLimit,
			RatePeriod: c.RatePeriod,
		}
	}
	apiLimitsMap := make(map[string]localhttp.MemoryLimiterConfig)
	for _, c := range cfg.APILimits {
		apiLimitsMap[c.Key] = localhttp.MemoryLimiterConfig{
			RateLimit:  c.RateLimit,
			RatePeriod: c.RatePeriod,
		}
	}
	return &localhttp.APILimiterConfig{
		PathPattern: defaultMap,
		HostPattern: patternMap,
		APILimits:   apiLimitsMap,
		IPLimitCfg:  cfg.IPLimitCfg,
	}
}
