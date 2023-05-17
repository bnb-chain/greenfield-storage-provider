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

type Option func(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error
type NewModularFunc = func(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error)

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
		// TODO: add modular
	})
}

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

func GetRegisterModulus() []string {
	mdmgr.mux.RLock()
	defer mdmgr.mux.RUnlock()
	return mdmgr.modulus
}

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

func GetNewModularFunc(name string) NewModularFunc {
	mdmgr.mux.RLock()
	defer mdmgr.mux.RUnlock()
	if _, ok := mdmgr.newModularFunc[name]; !ok {
		log.Panicf("not register [%s] modular info", name)
	}
	return mdmgr.newModularFunc[name]
}
