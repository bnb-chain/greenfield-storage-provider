package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/database/mysql"
	"github.com/forbole/juno/v4/database/sqlclient"
	"gorm.io/gorm"
)

var _ database.Database = &DB{}

// DB represents a SQL database with expanded features.
// so that it can properly store custom BigDipper-related data.
type DB struct {
	*mysql.Database
}

// BlockSyncerDBBuilder allows to create a new DB instance implementing the db.Builder type
func BlockSyncerDBBuilder(ctx *database.Context) (database.Database, error) {
	db, err := sqlclient.New(&ctx.Cfg)
	if err != nil {
		return nil, err
	}
	return &DB{
		Database: &mysql.Database{
			Impl: database.Impl{
				Db:             db,
				EncodingConfig: ctx.EncodingConfig,
			},
		},
	}, nil
}

// Cast allows to cast the given db to a DB instance
func Cast(db database.Database) *DB {
	bdDatabase, ok := db.(*DB)
	if !ok {
		panic(fmt.Errorf("given database instance is not a DB"))
	}
	return bdDatabase
}

// errIsNotFound check if the error is not found
func errIsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, gorm.ErrRecordNotFound)
}
