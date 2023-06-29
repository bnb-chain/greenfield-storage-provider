package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/database/mysql"
	"github.com/forbole/juno/v4/database/sqlclient"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/forbole/juno/v4/log"
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

func (db *DB) AutoMigrate(ctx context.Context, tables []schema.Tabler) error {
	q := db.Db.WithContext(ctx)
	m := db.Db.Migrator()
	for _, t := range tables {
		if t.TableName() == bsdb.PrefixTreeTableName || t.TableName() == bsdb.ObjectTableName {
			for i := 0; i < 64; i++ {
				shardTableName := fmt.Sprintf(t.TableName()+"_%02d", i)
				if err := q.Table(shardTableName).AutoMigrate(t); err != nil {
					log.Errorw("migrate table failed", "table", t.TableName(), "err", err)
					return err
				}
			}
		} else {
			if err := m.AutoMigrate(t); err != nil {
				log.Errorw("migrate table failed", "table", t.TableName(), "err", err)
				return err
			}
		}
	}
	return nil
}

func (db *DB) PrepareTables(ctx context.Context, tables []schema.Tabler) error {
	q := db.Db.WithContext(ctx)
	m := db.Db.Migrator()

	for _, t := range tables {
		if t.TableName() == bsdb.PrefixTreeTableName || t.TableName() == bsdb.ObjectTableName {
			for i := 0; i < 64; i++ {
				shardTableName := fmt.Sprintf(t.TableName()+"_%02d", i)
				if m.HasTable(shardTableName) {
					continue
				}
				if err := q.Table(shardTableName).AutoMigrate(t); err != nil {
					log.Errorw("migrate table failed", "table", shardTableName, "err", err)
					return err
				}
			}
		} else {
			if m.HasTable(t.TableName()) {
				continue
			}
			if err := q.Table(t.TableName()).AutoMigrate(t); err != nil {
				log.Errorw("migrate table failed", "table", t.TableName(), "err", err)
				return err
			}
		}

	}

	return nil
}
