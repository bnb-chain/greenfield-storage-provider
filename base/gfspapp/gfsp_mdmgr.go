package gfspapp

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util/maps"
)

// Option defines the GfSpBaseApp and module init options func type.
type Option func(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error

// NewModularFunc defines the module new instance func type.
type NewModularFunc = func(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error)

// ModularManager manages the modulus, record the module info, module info include:
// module name, module description and new module func. Module name is an indexer for
// starting, the start module name comes from config file or '--service' command flag.
// Module description uses for 'list' command that shows the SP supports modules info.
// New module func is help module manager to init the module instance.
type ModularManager struct {
	modulus        []string
	descriptions   map[string]string
	newModularFunc map[string]NewModularFunc
	mux            sync.RWMutex
}

var mdmgr *ModularManager
var once sync.Once

func init() {
	once.Do(func() {
		mdmgr = &ModularManager{
			descriptions:   make(map[string]string),
			newModularFunc: make(map[string]NewModularFunc),
		}
	})
}

// RegisterModular registers the module info to the global ModularManager
func RegisterModular(name string, description string, newFunc NewModularFunc) {
	mdmgr.mux.Lock()
	defer mdmgr.mux.Unlock()
	if name == "" {
		log.Panic("modular name cannot be blank")
	}
	if _, ok := mdmgr.newModularFunc[name]; ok {
		log.Panicf("[%s] modular repeated", name)
	}
	mdmgr.modulus = append(mdmgr.modulus, name)
	if len(description) != 0 {
		mdmgr.descriptions[name] = description
	}
	mdmgr.newModularFunc[name] = newFunc
}

// GetRegisterModulus returns the list registered modules.
func GetRegisterModulus() []string {
	mdmgr.mux.RLock()
	defer mdmgr.mux.RUnlock()
	return mdmgr.modulus
}

// GetRegisterModulusDescription returns the list registered modules' description.
func GetRegisterModulusDescription() string {
	mdmgr.mux.RLock()
	defer mdmgr.mux.RUnlock()
	var descriptions string
	names := maps.SortKeys(mdmgr.newModularFunc)
	for _, name := range names {
		descriptions = descriptions + fmt.Sprintf("%-"+strconv.Itoa(20)+"s %s\n",
			name, mdmgr.descriptions[name])
	}
	return descriptions
}

// GetNewModularFunc returns the list registered module's new instances func.
func GetNewModularFunc(name string) NewModularFunc {
	mdmgr.mux.RLock()
	defer mdmgr.mux.RUnlock()
	if _, ok := mdmgr.newModularFunc[name]; !ok {
		log.Panicf("not register [%s] modular info", name)
	}
	return mdmgr.newModularFunc[name]
}
