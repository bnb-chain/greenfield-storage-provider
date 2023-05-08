package utils

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/urfave/cli/v2"
)

var ListModularCmd = &cli.Command{
	Action:      listModularAction,
	Name:        "list",
	Usage:       "List the modular in greenfield storage provider",
	Category:    "MISCELLANEOUS COMMANDS",
	Description: `The list command output the services in greenfield storage provider.`,
}

func listModularAction(ctx *cli.Context) error {
	fmt.Printf(gfspapp.GetRegisterModulusDescription())
	return nil
}
