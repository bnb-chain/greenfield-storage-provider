package conf

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/config"
)

var ConfigDumpCmd = cli.Command{
	Action:   dumpConfigAction,
	Name:     "config.dump",
	Usage:    "Dump default configuration to file for editing",
	Category: "CONFIG COMMANDS",
	Description: `
The config.dump command writes default configuration 
values to ./config.toml file for editing.`,
}

// dumpConfigAction is the dump.config command.
func dumpConfigAction(ctx *cli.Context) error {
	return config.SaveConfig("./config.toml", config.DefaultStorageProviderConfig)
}

var ConfigUploadCmd = cli.Command{
	Action: configUploadAction,
	Name:   "config.upload",
	Usage:  "Upload the config file to db",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		utils.DBUserFlag,
		utils.DBPasswordFlag,
	},
	Category: "CONFIG COMMANDS",
	Description: `
The config.upload command upload the file to db for sp to load`,
}

// configUploadAction is the config.upload command.
func configUploadAction(ctx *cli.Context) error {
	//var configFile string
	//if ctx.GlobalIsSet(utils.ConfigFileFlag.Name) {
	//	configFile = ctx.GlobalString(utils.ConfigFileFlag.Name)
	//} else {
	//	configFile = "./config.toml"
	//}
	//cfg := config.LoadConfig(configFile)

	// TODO:: new sp db, and upload
	// replace by env vars
	return nil
}
