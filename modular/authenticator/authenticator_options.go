package authenticator

import (
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspconfig"
	coremodule "github.com/zkMeLabs/mechain-storage-provider/core/module"
)

func NewAuthenticationModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	auth := &AuthenticationModular{baseApp: app}
	return auth, nil
}
