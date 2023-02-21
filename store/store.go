package store

import (
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobmemory"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobsql"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metalevel"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// NewMetaDB return a meta-db instance
func NewMetaDB(dbType string, levelDBConfig *config.LevelDBConfig, sqlDBConfig *config.SqlDBConfig) (spdb.MetaDB, error) {
	var (
		metaDB spdb.MetaDB
		err    error
	)

	switch dbType {
	case model.MySqlDB:
		// load meta db config from env vars
		sqlDBConfig.User, sqlDBConfig.Passwd, err = getDBConfigFromEnv(model.MetaDBUser, model.MetaDBPassword)
		if err != nil {
			log.Error("load meta db config from env failed")
			return nil, err
		}
		metaDB, err = metasql.NewMetaDB(sqlDBConfig)
	case model.LevelDB:
		metaDB, err = metalevel.NewMetaDB(levelDBConfig)
	default:
		err = fmt.Errorf("meta db not support %s type", dbType)
	}
	return metaDB, err
}

// NewJobDB return a job-db instance
func NewJobDB(dbType string, sqlDBConfig *config.SqlDBConfig) (spdb.JobDB, error) {
	var (
		jobDB spdb.JobDB
		err   error
	)

	switch dbType {
	case model.MySqlDB:
		// load job db config from env vars
		sqlDBConfig.User, sqlDBConfig.Passwd, err = getDBConfigFromEnv(model.JobDBUser, model.JobDBPassword)
		if err != nil {
			log.Error("load job db config from env failed")
			return nil, err
		}
		jobDB, err = jobsql.NewJobMetaImpl(sqlDBConfig)
	case model.MemoryDB:
		jobDB = jobmemory.NewMemJobDB()
	default:
		err = fmt.Errorf("job db not support %s type", dbType)
	}
	return jobDB, err
}

func getDBConfigFromEnv(user, passwd string) (string, string, error) {
	userVal, ok := os.LookupEnv(user)
	if !ok {
		return "", "", fmt.Errorf("db %s config is not setted in environment", user)
	}
	passwdVal, ok := os.LookupEnv(passwd)
	if !ok {
		return "", "", fmt.Errorf("db %s config is not setted in environment", passwd)
	}
	return userVal, passwdVal, nil
}
