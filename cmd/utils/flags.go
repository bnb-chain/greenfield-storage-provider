package utils

import (
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

const (
	LoggingCategory  = "LOGGING AND DEBUGGING"
	MetricsCategory  = "METRICS AND STATS"
	DatabaseCategory = "DATABASE"
)

var (
	ConfigFileFlag = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "Config file path for uploading to db",
		Value:   "./config.toml",
	}
	ConfigRemoteFlag = &cli.BoolFlag{
		Name: "configremote",
		Usage: "Flag load config from remote db,if 'config.remote' be set, the db.user, " +
			"db.password and db.address flags are needed, otherwise use default value",
	}
	ServerFlag = &cli.StringFlag{
		Name:    "server",
		Aliases: []string{"service", "s"},
		Usage:   "Services to be started list, eg -server gateway,uploader,syncer... ",
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
		Usage:    "DB password",
		EnvVars:  []string{model.SpDBPasswd},
	}
	DBAddressFlag = &cli.StringFlag{
		Name:     "db.address",
		Category: DatabaseCategory,
		Usage:    "DB address",
		EnvVars:  []string{model.SpDBAddress},
		Value:    "localhost:3306",
	}
	DBDataBaseFlag = &cli.StringFlag{
		Name:     "db.database",
		Category: DatabaseCategory,
		Usage:    "DB database",
		EnvVars:  []string{model.SpDBDataBase},
		Value:    "localhost:3306",
	}
	// log flags
	LogLevelFlag = &cli.StringFlag{
		Name:     "log.level",
		Category: LoggingCategory,
		Usage:    "log level",
		Value:    "info",
	}
	LogPathFlag = &cli.StringFlag{
		Name:     "log.path",
		Category: LoggingCategory,
		Usage:    "log path",
		Value:    "./gnfd.log",
	}
	LogStdOutputFlag = &cli.BoolFlag{
		Name:     "log.std",
		Category: LoggingCategory,
		Usage:    "log standard output",
	}
	// Metrics flags
	MetricsEnabledFlag = &cli.BoolFlag{
		Name:     "metrics",
		Usage:    "Enable metrics collection and reporting",
		Category: MetricsCategory,
	}
	// MetricsHTTPFlag defines the endpoint for a stand-alone metrics HTTP endpoint.
	MetricsHTTPFlag = &cli.StringFlag{
		Name:     "metrics.addr",
		Usage:    `Enable stand-alone metrics HTTP server listening interface`,
		Category: MetricsCategory,
	}
)
