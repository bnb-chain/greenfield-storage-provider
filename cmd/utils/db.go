package utils

import (
	"github.com/urfave/cli/v2"

	storeconfig "github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

// MakeSPDB return sp db instance from db flags
func MakeSPDB(ctx *cli.Context, spDBCfg *storeconfig.SQLDBConfig) (*sqldb.SpDBImpl, error) {
	if ctx.IsSet(ctx.String(DBUserFlag.Name)) {
		spDBCfg.User = ctx.String(DBUserFlag.Name)
	}
	if ctx.IsSet(ctx.String(DBPasswordFlag.Name)) {
		spDBCfg.Passwd = ctx.String(DBPasswordFlag.Name)
	}
	if ctx.IsSet(ctx.String(DBAddressFlag.Name)) {
		spDBCfg.Address = ctx.String(DBAddressFlag.Name)
	}
	if ctx.IsSet(ctx.String(DBDatabaseFlag.Name)) {
		spDBCfg.Database = ctx.String(DBDatabaseFlag.Name)
	}
	return sqldb.NewSpDB(spDBCfg)
}
