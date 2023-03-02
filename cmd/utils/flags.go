package utils

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

var (
	VersionFlag = cli.BoolFlag{
		Name:  "version",
		Usage: "Show the storage provider version information",
	}
	ConfigFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Config file path for uploading to db",
		Value: "./config.toml",
	}
	DBUserFlag = cli.StringFlag{
		Name:   "user",
		Usage:  "DB user name",
		EnvVar: model.SPDBUser,
	}
	DBPasswordFlag = cli.StringFlag{
		Name:   "password",
		Usage:  "DB password",
		EnvVar: model.SPDBPasswd,
	}
)
