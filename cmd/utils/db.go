package utils

import (
	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"github.com/urfave/cli/v2"
)

// MakeSPDB return sp db instance from db flags
func MakeSPDB(ctx *cli.Context) (*sqldb.SpDBImpl, error) {
	spDBCfg := config.DefaultSQLDBConfig
	if ctx.IsSet(ctx.String(DBUserFlag.Name)) {
		spDBCfg.User = ctx.String(DBUserFlag.Name)
	}
	if ctx.IsSet(ctx.String(DBPasswordFlag.Name)) {
		spDBCfg.Passwd = ctx.String(DBPasswordFlag.Name)
	}
	if ctx.IsSet(ctx.String(DBAddressFlag.Name)) {
		spDBCfg.Address = ctx.String(DBAddressFlag.Name)
	}
	if ctx.IsSet(ctx.String(DBDataBaseFlag.Name)) {
		spDBCfg.Database = ctx.String(DBDataBaseFlag.Name)
	}
	return sqldb.NewSpDB(spDBCfg)
}
