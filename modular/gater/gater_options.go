package gater

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultGatewayDomain    = "localhost:9133"
	DefaultMaxListReadQuota = 100
)

func init() {
	gfspapp.RegisterModularInfo(GateModularName, GateModularDescription, NewGateModular)
}

func NewGateModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	if cfg.Customize.Gater != nil {
		app.SetGater(cfg.Customize.Gater)
		return cfg.Customize.Gater, nil
	}
	gater := &GateModular{baseApp: app}
	if err := DefaultGaterOptions(gater, cfg); err != nil {
		return nil, err
	}
	app.SetGater(gater)
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
	gater.domain = cfg.Gateway.Domain
	gater.httpAddress = cfg.Gateway.HttpAddress
	gater.maxListReadQuota = cfg.Bucket.MaxListReadQuotaNumber
	return nil
}
