package gater

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspmdmgr"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultGatewayDomain    = "localhost:9133"
	DefaultMaxListReadQuota = 100
)

func init() {
	gfspmdmgr.RegisterModularInfo(GateModularName, GateModularDescription, NewGateModular)
}

func NewGateModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspconfig.Option) (
	coremodule.Modular, error) {
	if cfg.Gater != nil {
		app.SetGater(cfg.Gater)
		return cfg.Gater, nil
	}
	gater := &GateModular{baseApp: app}
	opts = append(opts, gater.DefaultGaterOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetGater(gater)
	return gater, nil
}

func (g *GateModular) DefaultGaterOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.GatewayDomain == "" {
		cfg.GatewayDomain = DefaultGatewayDomain
	}
	if cfg.GatewayHttpAddress == "" {
		cfg.GatewayHttpAddress = DefaultGatewayDomain
	}
	if cfg.MaxListReadQuota == 0 {
		cfg.MaxListReadQuota = DefaultMaxListReadQuota
	}
	g.domain = cfg.GatewayDomain
	g.httpAddress = cfg.GatewayHttpAddress
	g.maxListReadQuota = cfg.MaxListReadQuota
	return nil
}
