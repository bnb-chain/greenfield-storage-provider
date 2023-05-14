package authorizer

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspmdmgr"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

func init() {
	gfspmdmgr.RegisterModularInfo(AuthorizationModularName, AuthorizationModularDescription, NewAuthorizeModular)
}

func NewAuthorizeModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspconfig.Option) (
	coremodule.Modular, error) {
	if cfg.Authorizer != nil {
		app.SetAuthorizer(cfg.Authorizer)
		return cfg.Authorizer, nil
	}
	authorize := &AuthorizeModular{baseApp: app}
	for _, opt := range opts {
		err := opt(app, cfg)
		if err != nil {
			return nil, err
		}
	}
	app.SetAuthorizer(authorize)
	return authorize, nil
}
