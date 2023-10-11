package gater

import (
	"strings"

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
	pathPatternMap := make(map[string][]mwhttp.MemoryLimiterConfig)
	pathSequence := make([]string, len(cfg.PathPattern))
	// for each name inside all PathPattern, try match with NameToLimit and get its pairing RateLimit and RatePeriod
	for i, c := range cfg.PathPattern {
		for _, name := range c.Names {
			for _, v := range cfg.NameToLimit {
				if strings.EqualFold(v.Name, name) {
					pathSequence[i] = c.Key
					pathPatternMap[c.Key] = append(pathPatternMap[c.Key], v)
				}
			}
		}
	}
	patternMap := make(map[string][]mwhttp.MemoryLimiterConfig)
	hostSequence := make([]string, len(cfg.HostPattern))
	// for each name inside all HostPattern, try match with NameToLimit and get its pairing RateLimit and RatePeriod
	for i, c := range cfg.HostPattern {
		for _, name := range c.Names {
			for _, v := range cfg.NameToLimit {
				if strings.EqualFold(v.Name, name) {
					hostSequence[i] = c.Key
					patternMap[c.Key] = append(patternMap[c.Key], v)
				}
			}
		}
	}
	apiLimitsMap := make(map[string][]mwhttp.MemoryLimiterConfig)
	// for each name inside all APILimits, try match with NameToLimit and get its pairing RateLimit and RatePeriod
	for _, c := range cfg.APILimits {
		for _, name := range c.Names {
			for _, v := range cfg.NameToLimit {
				if strings.EqualFold(v.Name, name) {
					apiLimitsMap[c.Key] = append(apiLimitsMap[c.Key], v)
				}
			}
		}
	}
	return &mwhttp.APILimiterConfig{
		PathPattern:  pathPatternMap,
		PathSequence: pathSequence,
		HostPattern:  patternMap,
		HostSequence: hostSequence,
		APILimits:    apiLimitsMap,
		IPLimitCfg:   cfg.IPLimitCfg,
	}
}
