package utils

import (
	"fmt"
	"strconv"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/maps"
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
	services := maps.SortKeys(model.SpServiceDesc)
	for _, name := range services {
		desc := model.SpServiceDesc[name]
		fmt.Printf("%-"+strconv.Itoa(20)+"s %s\n", name, desc)
	}
	return nil
}
