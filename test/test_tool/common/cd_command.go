package common

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/bnb-chain/inscription-storage-provider/test/test_tool/context"
)

var CdCommand = cli.Command{
	Name:   "cd",
	Usage:  "enter service namespace",
	Action: cdService,
}

func cdService(c *cli.Context) {
	if len(c.Args()) == 0 {
		return
	}
	name := c.Args()[0]
	if len(name) == 0 {
		return
	}
	ctx := context.GetContext()
	if name == ".." {
		ctx.OutService()
		return
	}
	if err := ctx.EnterService(name); err != nil {
		fmt.Println(err)
		return
	}
	return
}
