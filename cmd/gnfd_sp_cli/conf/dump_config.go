package conf

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/bnb-chain/greenfield-storage-provider/config"
)

var DumpConfigCommand = cli.Command{
	Name:   "dump.config",
	Usage:  "dump default config",
	Action: DumpDefaultConfig,
}

func DumpDefaultConfig(c *cli.Context) {
	err := config.SaveConfig("./", config.DefaultStorageProviderConfig)
	if err != nil {
		fmt.Println(err)
	}
}
