package gater

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
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
	if cfg.Gateway.HttpAddress == "" {
		cfg.Gateway.HttpAddress = DefaultGatewayDomain
	}
	if cfg.Bucket.MaxListReadQuotaNumber == 0 {
		cfg.Bucket.MaxListReadQuotaNumber = DefaultMaxListReadQuota
	}
	if cfg.Bucket.MaxPayloadSize == 0 {
		cfg.Bucket.MaxPayloadSize = DefaultMaxPayloadSize
	}
	gater.maxPayloadSize = cfg.Bucket.MaxPayloadSize
	gater.domain = cfg.Gateway.Domain
	gater.httpAddress = cfg.Gateway.HttpAddress
	gater.maxListReadQuota = cfg.Bucket.MaxListReadQuotaNumber
	return nil
}
