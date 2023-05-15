package authorizer

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

func init() {
	gfspapp.RegisterModularInfo(AuthorizationModularName, AuthorizationModularDescription, NewAuthorizeModular)
}

func NewAuthorizeModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	if cfg.Customize.Authorizer != nil {
		app.SetAuthorizer(cfg.Customize.Authorizer)
		return cfg.Customize.Authorizer, nil
	}
	authorize := &AuthorizeModular{baseApp: app}
	app.SetAuthorizer(authorize)
	return authorize, nil
}
