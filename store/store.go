package store

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobmemory"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobsql"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metalevel"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
)

// NewMetaDB return a meta-db instance
func NewMetaDB(dbType string, levelDBConfig *config.LevelDBConfig, sqlDBConfig *config.SqlDBConfig) (metadb.MetaDB, error) {
	var (
		metaDB metadb.MetaDB
		err    error
	)

	switch dbType {
	case model.MySqlDB:
		metaDB, err = metasql.NewMetaDB(sqlDBConfig)
	case model.LevelDB:
		metaDB, err = metalevel.NewMetaDB(levelDBConfig)
	default:
		err = fmt.Errorf("meta db not support %s type", dbType)
	}
	return metaDB, err
}

// NewJobDB return a job-db instance
func NewJobDB(dbType string, sqlDBConfig *config.SqlDBConfig) (jobdb.JobDBV2, error) {
	var (
		jobDB jobdb.JobDBV2
		err   error
	)

	switch dbType {
	case model.MySqlDB:
		jobDB, err = jobsql.NewJobMetaImpl(sqlDBConfig)
	case model.MemoryDB:
		jobDB = jobmemory.NewMemJobDB()
	default:
		err = fmt.Errorf("job db not support %s type", dbType)
	}
	return jobDB, err
}
