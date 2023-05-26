package command

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
)

var ListModularCmd = &cli.Command{
	Action:      listModularAction,
	Name:        "list",
	Usage:       "List the modules in greenfield storage provider",
	Category:    "MISCELLANEOUS COMMANDS",
	Description: `The list command output the services in greenfield storage provider.`,
}

func listModularAction(ctx *cli.Context) error {
	fmt.Print(gfspapp.GetRegisterModulusDescription())
	return nil
}
