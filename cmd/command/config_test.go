package command

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestConfigDumpCmd(t *testing.T) {
	err := ConfigDumpCmd.Action(&cli.Context{})
	assert.Equal(t, nil, err)
	_, err = os.Stat(DefaultConfigFile)
	assert.Equal(t, nil, err)
	os.Remove(DefaultConfigFile)
}
