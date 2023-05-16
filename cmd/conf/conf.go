package conf

import "github.com/urfave/cli/v2"

var ConfigDumpCmd = &cli.Command{
	Action:   dumpConfigAction,
	Name:     "config.dump",
	Usage:    "Dump default configuration to the './config.toml' file for editing",
	Category: "CONFIG COMMANDS",
	Description: `The config.dump command writes default configuration 
values to ./config.toml file for editing.`,
}

// dumpConfigAction is the dump.config command.
func dumpConfigAction(ctx *cli.Context) error {
	// return config.SaveConfig("./config.toml", config.DefaultStorageProviderConfig)
	// TODO: add dump config
	return nil
}
