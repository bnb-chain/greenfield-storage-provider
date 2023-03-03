package conf

import (
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/config"
	storeconf "github.com/bnb-chain/greenfield-storage-provider/store/config"
)

var ConfigDumpCmd = &cli.Command{
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

var ConfigUploadCmd = &cli.Command{
	Action: configUploadAction,
	Name:   "config.upload",
	Usage:  "Upload the config file to db",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		utils.DBUserFlag,
		utils.DBPasswordFlag,
		utils.DBAddressFlag,
	},
	Category: "CONFIG COMMANDS",
	Description: `
The config.upload command upload the file to db for sp to load
db.user and db.password flag support come from ENV Var
SP_DB_USER and SP_DB_PASSWORD`,
}

// configUploadAction is the config.upload command.
func configUploadAction(ctx *cli.Context) error {
	_ = config.LoadConfig(ctx.String(utils.ConfigFileFlag.Name))
	_ = &storeconf.SQLDBConfig{
		User:     ctx.String(utils.DBUserFlag.Name),
		Passwd:   ctx.String(utils.DBPasswordFlag.Name),
		Address:  ctx.String(utils.DBAddressFlag.Name),
		Database: "job_db",
	}
	// TODO:: new sp db and upload config
	return nil
}
