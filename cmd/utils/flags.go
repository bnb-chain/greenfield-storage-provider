package utils

import (
	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/model"
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
	ConfigRemoteFlag = &cli.BoolFlag{
		Name:     "config.remote",
		Category: ConfigCategory,
		Usage: "Flag load config from remote db,if 'config.remote' be set, the db.user, " +
			"db.password and db.address flags are needed, otherwise use the default value",
	}
	ServerFlag = &cli.StringFlag{
		Name:     "server",
		Category: ConfigCategory,
		Aliases:  []string{"service", "svc"},
		Usage:    "Services to be started list, e.g. -server gateway, uploader, receiver...",
	}

	// database flags
	DBUserFlag = &cli.StringFlag{
		Name:     "db.user",
		Category: DatabaseCategory,
		Usage:    "DB user name",
		EnvVars:  []string{model.SpDBUser},
	}
	DBPasswordFlag = &cli.StringFlag{
		Name:     "db.password",
		Category: DatabaseCategory,
		Usage:    "DB user password",
		EnvVars:  []string{model.SpDBPasswd},
	}
	DBAddressFlag = &cli.StringFlag{
		Name:     "db.address",
		Category: DatabaseCategory,
		Usage:    "DB listen address",
		EnvVars:  []string{model.SpDBAddress},
		Value:    config.DefaultSQLDBConfig.Address,
	}
	DBDatabaseFlag = &cli.StringFlag{
		Name:     "db.database",
		Category: DatabaseCategory,
		Usage:    "DB database name",
		EnvVars:  []string{model.SpDBDataBase},
		Value:    config.DefaultSQLDBConfig.Database,
	}

	// resource manager flags
	DisableResourceManagerFlag = &cli.BoolFlag{
		Name:     "rcmgr.disable",
		Category: ResourceManagerCategory,
		Usage:    "Disable resource manager",
		Value:    true,
	}
	ResourceManagerConfigFlag = &cli.StringFlag{
		Name:     "rcmgr.config",
		Category: ResourceManagerCategory,
		Usage:    "Resource manager config file path",
		Value:    "",
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
		Value:    config.DefaultLogConfig.Path,
	}
	LogStdOutputFlag = &cli.BoolFlag{
		Name:     "log.std",
		Category: LoggingCategory,
		Usage:    "Log output standard io",
	}

	// Metrics flags
	MetricsEnabledFlag = &cli.BoolFlag{
		Name:     "metrics",
		Category: MetricsCategory,
		Usage:    "Enable metrics collection and reporting",
		Value:    config.DefaultMetricsConfig.Enabled,
	}
	MetricsHTTPFlag = &cli.StringFlag{
		Name:     "metrics.addr",
		Category: MetricsCategory,
		Usage:    "Specify stand-alone metrics HTTP server listening address",
		Value:    config.DefaultMetricsConfig.HTTPAddress,
	}

	// PProf flags
	PProfEnabledFlag = &cli.BoolFlag{
		Name:     "pprof",
		Category: PerfCategory,
		Usage:    "Enable the pprof HTTP server",
		Value:    config.DefaultPProfConfig.Enabled,
	}
	PProfHTTPFlag = &cli.StringFlag{
		Name:     "pprof.addr",
		Category: PerfCategory,
		Usage:    "Specify pprof HTTP server listening address",
		Value:    config.DefaultPProfConfig.HTTPAddress,
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
