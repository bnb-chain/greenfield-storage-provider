package bs_data_migration

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	fixv101 "github.com/bnb-chain/greenfield-storage-provider/cmd/command/bs_data_migration/fix-v1.0.0"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

var JobNameFlag = &cli.StringFlag{
	Name:     "job",
	Usage:    "Specify job name",
	Aliases:  []string{"j"},
	Required: true,
}

var BsDataMigrationCmd = &cli.Command{
	Action: bsDataMigrationAction,
	Name:   "migration",
	Usage:  "fix error data",

	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		JobNameFlag,
	},

	Category:    "blocksyncer data migration COMMANDS",
	Description: `blocksyncer data migration`,
}

func getDBConfigFromEnv(usernameKey, passwordKey string) (username, password string, err error) {
	var ok bool
	username, ok = os.LookupEnv(usernameKey)
	if !ok {
		return "", "", errors.New("dsn config is not set in environment")
	}
	password, ok = os.LookupEnv(passwordKey)
	if !ok {
		return "", "", errors.New("dsn config is not set in environment")
	}
	return
}

func bsDataMigrationAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	if len(cfg.Chain.ChainAddress) == 0 {
		return fmt.Errorf("config error")
	}

	username, password, envErr := getDBConfigFromEnv(bsdb.BsDBUser, bsdb.BsDBPasswd)
	if envErr != nil {
		log.Infof("failed to get username and password err:%v", envErr)
		username = cfg.BsDB.User
		password = cfg.BsDB.Passwd
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true&loc=Local&interpolateParams=true", username, password, cfg.BsDB.Address, cfg.BsDB.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Errorw("failed to connect db", "error", err)
		return err
	}

	if !ctx.IsSet(JobNameFlag.Name) {
		return fmt.Errorf("command param error miss -job")
	}
	jobName := ctx.String(JobNameFlag.Name)
	log.Infof(jobName)
	if jobName == "fix-payment" {
		return fixv101.FixPayment(cfg.Chain.ChainAddress[0], db)
	}

	return nil
}
