package utils

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/urfave/cli/v2"
)

var ListServiceCmd = &cli.Command{
	Action:      listServiceAction,
	Name:        "list",
	Usage:       "List the services in greenfield storage provider",
	Category:    "MISCELLANEOUS COMMANDS",
	Description: `The list command output the services in greenfield storage provider.`,
}

func listServiceAction(ctx *cli.Context) error {
	fmt.Println(gfspapp.GetRegisterModulusDescription())
	return nil
}
