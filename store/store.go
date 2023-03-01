package store

import (
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobsql"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
)

var (
	// MetaDB environment constants
	MetaDBUser     = "META_DB_USER"
	MetaDBPassword = "META_DB_PASSWORD"

	// JobDB environment constants
	JobDBUser     = "JOB_DB_USER"
	JobDBPassword = "JOB_DB_PASSWORD"
)

// NewMetaDB return a meta-db instance
func NewMetaDB(dbType string, levelDBConfig *config.LevelDBConfig, sqlDBConfig *config.SqlDBConfig) (spdb.MetaDB, error) {
	var err error
	sqlDBConfig.User, sqlDBConfig.Passwd, err = getDBConfigFromEnv(MetaDBUser, MetaDBPassword)
	if err != nil {
		log.Error("load meta db config from env failed")
		return nil, err
	}
	return metasql.NewMetaDB(sqlDBConfig)

}

// NewJobDB return a job-db instance
func NewJobDB(dbType string, sqlDBConfig *config.SqlDBConfig) (spdb.JobDB, error) {
	var err error
	sqlDBConfig.User, sqlDBConfig.Passwd, err = getDBConfigFromEnv(JobDBUser, JobDBPassword)
	if err != nil {
		log.Error("load job db config from env failed")
		return nil, err
	}
	return jobsql.NewJobMetaImpl(sqlDBConfig)
}

func getDBConfigFromEnv(user, passwd string) (string, string, error) {
	userVal, ok := os.LookupEnv(user)
	if !ok {
		return "", "", fmt.Errorf("db %s config is not set in environment", user)
	}
	passwdVal, ok := os.LookupEnv(passwd)
	if !ok {
		return "", "", fmt.Errorf("db %s config is not set in environment", passwd)
	}
	return userVal, passwdVal, nil
}
