package utils

import (
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

var (
	ConfigFileFlag = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "Config file path for uploading to db",
		Value:   "./config.toml",
	}
	ConfigRemoteFlag = &cli.BoolFlag{
		Name: "config.remote",
		Usage: "Flag load config from remote db,if 'config.remote' be set, the db.user, " +
			"db.password and db.address flags are needed, otherwise use default value",
	}
	ServerFlag = &cli.StringFlag{
		Name:    "server",
		Aliases: []string{"service", "s"},
		Usage:   "Services to be started list, eg -server gateway,uploader,receiver... ",
	}
	DBUserFlag = &cli.StringFlag{
		Name:    "db.user",
		Usage:   "DB user name",
		EnvVars: []string{model.SpDBUser},
	}
	DBPasswordFlag = &cli.StringFlag{
		Name:    "db.password",
		Usage:   "DB user password",
		EnvVars: []string{model.SpDBPasswd},
	}
	DBAddressFlag = &cli.StringFlag{
		Name:    "db.address",
		Usage:   "DB listen address",
		EnvVars: []string{model.SpDBAddress},
		Value:   "localhost:3306",
	}
	DBDataBaseFlag = &cli.StringFlag{
		Name:    "db.database",
		Usage:   "DB database name",
		EnvVars: []string{model.SpDBDataBase},
		Value:   "localhost:3306",
	}
	LogLevelFlag = &cli.StringFlag{
		Name:  "log.level",
		Usage: "log level",
		Value: "info",
	}
	LogPathFlag = &cli.StringFlag{
		Name:  "log.path",
		Usage: "log output file path",
		Value: "./gnfd-sp.log",
	}
	LogStdOutputFlag = &cli.BoolFlag{
		Name:  "log.std",
		Usage: "log output standard io",
	}
)
