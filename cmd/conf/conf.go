package conf

import (
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/config"
)

var ConfigDumpCmd = &cli.Command{
	Action:   dumpConfigAction,
	Name:     "config.dump",
	Usage:    "Dump default configuration to the './config.toml' file for editing",
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
		utils.DBDatabaseFlag,
	},
	Category: "CONFIG COMMANDS",
	Description: `
The config.upload command uploads the file to db. If db.user,db.password,db.address 
and db.database flags are not set, uses the 'SpDBConfig' in config file, it will load
SP_DB_USER, SP_DB_PASSWORD, SP_DB_ADDRESS and SP_DB_DATABASE from ENV.`,
}

// configUploadAction is the config.upload command.
func configUploadAction(ctx *cli.Context) error {
	cfg := config.LoadConfig(ctx.String(utils.ConfigFileFlag.Name))
	cfgBytes, err := cfg.JSONMarshal()
	if err != nil {
		return err
	}
	spDB, err := utils.MakeSPDB(ctx, cfg.SpDBConfig)
	if err != nil {
		return err
	}
	return spDB.SetAllServiceConfigs("default", string(cfgBytes))
}
