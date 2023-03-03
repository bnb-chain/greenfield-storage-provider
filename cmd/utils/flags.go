package utils

import (
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

var (
	VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Show the storage provider version information",
	}
	ConfigFileFlag = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "Config file path for uploading to db",
		Value:   "./config.toml",
	}
	ConfigRemoteFlag = &cli.BoolFlag{
		Name:  "config.remote",
		Usage: "Flag load config from remote db",
	}
	ServerFlag = &cli.StringFlag{
		Name:    "server",
		Aliases: []string{"service", "s"},
		Usage:   "Services to be started list, eg -server gateway,uploader,syncer... ",
	}
	DBUserFlag = &cli.StringFlag{
		Name:    "db.user",
		Usage:   "DB user name",
		EnvVars: []string{model.SpDBUser},
	}
	DBPasswordFlag = &cli.StringFlag{
		Name:    "db.password",
		Usage:   "DB password",
		EnvVars: []string{model.SpDBPasswd},
	}
	DBAddressFlag = &cli.StringFlag{
		Name:  "db.address",
		Usage: "DB address",
		Value: "localhost:3306",
	}
	LogLevelFlag = &cli.StringFlag{
		Name:  "log.level",
		Usage: "log level",
		Value: "info",
	}
	LogPathFlag = &cli.StringFlag{
		Name:  "log.path",
		Usage: "log path",
		Value: "./gnfd.log",
	}
)
