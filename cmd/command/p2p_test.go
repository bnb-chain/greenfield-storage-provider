package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCreateKeys(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		P2PCreateKeysCmd,
	}
	err := app.Run([]string{"./gnfd-sp", "p2p.create.key", "-n", "10"})
	assert.Nil(t, err)
}
