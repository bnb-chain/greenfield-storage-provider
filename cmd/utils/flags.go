package utils

import (
	"errors"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"
)

const (
	ConfigCategory          = "SP CONFIG"
	LoggingCategory         = "LOGGING AND DEBUGGING"
	MetricsCategory         = "METRICS AND STATS"
	DatabaseCategory        = "DATABASE"
	ResourceManagerCategory = "RESOURCE MANAGER"
	PerfCategory            = "PERFORMANCE TUNING"
)

var (
	ConfigFileFlag = &cli.StringFlag{
		Name:     "config",
		Category: ConfigCategory,
		Aliases:  []string{"c"},
		Usage:    "Config file path for uploading to db",
		Value:    "./config.toml",
	}
	ServerFlag = &cli.StringFlag{
		Name:     "server",
		Category: ConfigCategory,
		Aliases:  []string{"service", "svc", "s"},
		Usage:    "Services to be started list, e.g. -server gateway, uploader, receiver...",
	}

	// resource manager flags
	DisableResourceManagerFlag = &cli.BoolFlag{
		Name:     "rcmgr.disable",
		Category: ResourceManagerCategory,
		Usage:    "Disable resource manager",
	}

	// log flags
	LogLevelFlag = &cli.StringFlag{
		Name:     "log.level",
		Category: LoggingCategory,
		Usage:    "Log level",
		Value:    "info",
	}
	LogPathFlag = &cli.StringFlag{
		Name:     "log.path",
		Category: LoggingCategory,
		Usage:    "Log output file path",
		Value:    "./",
	}
	LogStdOutputFlag = &cli.BoolFlag{
		Name:     "log.std",
		Category: LoggingCategory,
		Usage:    "Log output standard io",
	}

	// Metrics flags
	MetricsDisableFlag = &cli.BoolFlag{
		Name:     "metrics.disable",
		Category: MetricsCategory,
		Usage:    "Disable metrics collection and reporting",
	}
	MetricsHTTPFlag = &cli.StringFlag{
		Name:     "metrics.addr",
		Category: MetricsCategory,
		Usage:    "Specify stand-alone metrics HTTP server listening address",
	}

	// PProf flags
	PProfDisableFlag = &cli.BoolFlag{
		Name:     "pprof.disable",
		Category: PerfCategory,
		Usage:    "Disable the pprof HTTP server",
	}
	PProfHTTPFlag = &cli.StringFlag{
		Name:     "pprof.addr",
		Category: PerfCategory,
		Usage:    "Specify pprof HTTP server listening address",
	}
)

// MergeFlags merges the given flag slices.
func MergeFlags(groups ...[]cli.Flag) []cli.Flag {
	var ret []cli.Flag
	for _, group := range groups {
		ret = append(ret, group...)
	}
	return ret
}

// LoadConfig loads the configuration from file.
func LoadConfig(file string, cfg *gfspconfig.GfSpConfig) error {
	if cfg == nil {
		return errors.New("failed to load config file, the config param invalid")
	}
	bz, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return toml.Unmarshal(bz, cfg)
}
