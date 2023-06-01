package command

import (
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
)

var ConfigDumpCmd = &cli.Command{
	Action:   dumpConfigAction,
	Name:     "config.dump",
	Usage:    "Dump default configuration to the './config.toml' file for editing",
	Category: "CONFIG COMMANDS",
	Description: `The config.dump command writes default configuration 
values to ./config.toml file for editing.`,
}

// dumpConfigAction is the dump.config command action.
func dumpConfigAction(ctx *cli.Context) error {
	bz, err := toml.Marshal(&gfspconfig.GfSpConfig{})
	if err != nil {
		return err
	}
	if err = os.WriteFile("./config.toml", bz, os.ModePerm); err != nil {
		return err
	}
	return nil
}
