package gfspapp

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type Option func(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error
type NewModularFunc = func(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig, opts ...Option) (coremodule.Modular, error)

type ModularManager struct {
	modulus        []string
	descriptions   map[string]string
	newModularFunc map[string]NewModularFunc
	instances      map[string]coremodule.Modular
	mux            sync.RWMutex
}

var mdmgr *ModularManager
var once sync.Once

func init() {
	once.Do(func() {
		mdmgr = &ModularManager{
			descriptions:   make(map[string]string),
			instances:      make(map[string]coremodule.Modular),
			newModularFunc: make(map[string]NewModularFunc),
		}
	})
}

func RegisterModularInfo(name string, description string, newFunc NewModularFunc) {
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

func RegisterModularInstance(modular coremodule.Modular) {
	mdmgr.mux.RLock()
	defer mdmgr.mux.RUnlock()
	if _, ok := mdmgr.instances[modular.Name()]; ok {
		log.Panicf("[%s] modular repeated", modular.Name())
	}
	mdmgr.instances[modular.Name()] = modular
	return
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
	for name, description := range mdmgr.descriptions {
		descriptions = descriptions + fmt.Sprintln("%-"+strconv.Itoa(20)+"s %s\n", name, description)
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

func GetRegisterModulusInstances() []coremodule.Modular {
	mdmgr.mux.RLock()
	defer mdmgr.mux.RUnlock()
	var modulus []coremodule.Modular
	for _, modular := range mdmgr.instances {
		modulus = append(modulus, modular)
	}
	return modulus
}
