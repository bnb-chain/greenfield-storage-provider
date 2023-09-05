package gater

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	mwhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
)

const (
	DefaultGatewayDomainName = "localhost:9133"
	DefaultMaxListReadQuota  = 100
	DefaultMaxPayloadSize    = 2 * 1024 * 1024 * 1024
)

func NewGateModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	gater := &GateModular{baseApp: app}
	if err := defaultGaterOptions(gater, cfg); err != nil {
		return nil, err
	}
	return gater, nil
}

func defaultGaterOptions(gater *GateModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Gateway.DomainName == "" {
		cfg.Gateway.DomainName = DefaultGatewayDomainName
	}
	if cfg.Gateway.HTTPAddress == "" {
		cfg.Gateway.HTTPAddress = DefaultGatewayDomainName
	}
	if cfg.Bucket.MaxListReadQuotaNumber == 0 {
		cfg.Bucket.MaxListReadQuotaNumber = DefaultMaxListReadQuota
	}
	if cfg.Bucket.MaxPayloadSize == 0 {
		cfg.Bucket.MaxPayloadSize = DefaultMaxPayloadSize
	}
	gater.maxPayloadSize = cfg.Bucket.MaxPayloadSize
	gater.env = cfg.Env
	gater.domain = cfg.Gateway.DomainName
	gater.httpAddress = cfg.Gateway.HTTPAddress
	gater.maxListReadQuota = cfg.Bucket.MaxListReadQuotaNumber
	rateCfg := makeAPIRateLimitCfg(cfg.APIRateLimiter)
	if err := mwhttp.NewAPILimiter(rateCfg); err != nil {
		log.Errorw("failed to new api limiter", "err", err)
		return err
	}
	return nil
}

func makeAPIRateLimitCfg(cfg mwhttp.RateLimiterConfig) *mwhttp.APILimiterConfig {
	defaultMap := make(map[string]mwhttp.MemoryLimiterConfig)
	for _, c := range cfg.PathPattern {
		defaultMap[c.Key] = mwhttp.MemoryLimiterConfig{
			RateLimit:  c.RateLimit,
			RatePeriod: c.RatePeriod,
		}
	}
	patternMap := make(map[string]mwhttp.MemoryLimiterConfig)
	for _, c := range cfg.HostPattern {
		patternMap[c.Key] = mwhttp.MemoryLimiterConfig{
			RateLimit:  c.RateLimit,
			RatePeriod: c.RatePeriod,
		}
	}
	apiLimitsMap := make(map[string]mwhttp.MemoryLimiterConfig)
	for _, c := range cfg.APILimits {
		apiLimitsMap[c.Key] = mwhttp.MemoryLimiterConfig{
			RateLimit:  c.RateLimit,
			RatePeriod: c.RatePeriod,
		}
	}
	return &mwhttp.APILimiterConfig{
		PathPattern: defaultMap,
		HostPattern: patternMap,
		APILimits:   apiLimitsMap,
		IPLimitCfg:  cfg.IPLimitCfg,
	}
}
