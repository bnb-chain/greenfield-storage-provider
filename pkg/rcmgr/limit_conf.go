package rcmgr

import (
	"bufio"
	"fmt"
	"os"

	"github.com/naoina/toml"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

var _ Limiter = &LimitConfig{}

type LimitConfig struct {
	SystemLimit *BaseLimit
	Service     map[string]*BaseLimit
}

var DefaultLimitConfig = &LimitConfig{
	SystemLimit: DynamicLimits(),
	Service:     make(map[string]*BaseLimit),
}

func NewLimitConfigFromToml(file string) (*LimitConfig, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := LimitConfig{}
	err = util.TomlSettings.NewDecoder(bufio.NewReader(f)).Decode(&cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		log.Errorw("failed to parser resource manager limit config file", "error", err)
		return nil, err
	}
	return &cfg, nil
}

func (cfg *LimitConfig) GetSystemLimits() Limit {
	return cfg.SystemLimit
}

func (cfg *LimitConfig) GetTransientLimits() Limit {
	// TODO:: differentiate between SystemLimit and TransientLimits
	return cfg.SystemLimit
}

func (cfg *LimitConfig) GetServiceLimits(svc string) Limit {
	if _, ok := cfg.Service[svc]; ok {
		return cfg.Service[svc]
	}
	return nil
}

func (cfg *LimitConfig) String() string {
	output := fmt.Sprintf("system limits [%s]", cfg.SystemLimit.String())
	for svc, limit := range cfg.Service {
		svcOutput := fmt.Sprintf("%s service limits [%s]", svc, limit.String())
		output = output + "," + svcOutput
	}
	return output
}
