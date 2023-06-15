package authenticator

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

func NewAuthenticationModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	authorize := &AuthenticationModular{baseApp: app}
	return authorize, nil
}
