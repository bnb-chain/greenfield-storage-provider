package conf

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"gopkg.in/urfave/cli.v1"
)

var ConfigDumpCmd = cli.Command{
	Action:   dumpConfigAction,
	Name:     "config.dump",
	Usage:    "Dump default configuration to file for editing",
	Category: "CONFIG COMMANDS",
	Description: `
The config.dump command writes default configuration 
values to ./config.toml file for editing.`,
}

// dumpConfigAction is the dump.config command.
func dumpConfigAction(ctx *cli.Context) error {
	return config.SaveConfig("./config.toml", config.DefaultStorageProviderConfig)
}

var ListEnvCmd = cli.Command{
	Action:   listEnvAction,
	Name:     "config.listenv",
	Usage:    "List environment variables that recommended to be set for security reasons",
	Category: "CONFIG COMMANDS",
	Description: `
The config.listenv command output the environment variables that can be set,
for security reasons, these variables are not recommended to be exposed in 
configuration files`,
}

func listEnvAction(ctx *cli.Context) error {
	fmt.Printf("SQL DB: \n\t%s \n\t%s\n", model.SpDBUser, model.SpDBPasswd)
	fmt.Printf("AWS S3: \n\t%s \n\t%s \n\t%s \n\t%s\n", model.AWSAccessKey,
		model.AWSSecretKey, model.AWSSessionToken, model.BucketURL)
	fmt.Printf("SP KEY: \n\t%s \n\t%s \n\t%s \n\t%s \n\t%s\n", model.SpOperatorAddress, model.SpOperatorPrivKey,
		model.SpFundingPrivKey, model.SpApprovalPrivKey, model.SpSealPrivKey)
	return nil
}
