package common

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/bnb-chain/greenfield-storage-provider/test/test_tool/context"
)

var ListCommand = cli.Command{
	Name:   "ls",
	Usage:  "list service of storage provider",
	Action: listService,
}

func listService(c *cli.Context) {
	for service, usage := range context.ServiceUsage {
		fmt.Printf("%-3s%-14s  -  %-50s\n", "   ", service, usage)
	}
}
