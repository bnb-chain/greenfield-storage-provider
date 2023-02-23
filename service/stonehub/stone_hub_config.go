package stonehub

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobsql"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metalevel"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
)

type StoneHubConfig struct {
	StorageProvider   string
	Address           string
	JobDBType         string
	JobSqlDBConfig    *config.SqlDBConfig
	MetaDBType        string
	MetaLevelDBConfig *config.LevelDBConfig
	MetaSqlDBConfig   *config.SqlDBConfig
}

var DefaultStoneHubConfig = &StoneHubConfig{
	StorageProvider:   model.StorageProvider,
	Address:           model.DefaultStoneHubAddress,
	JobDBType:         model.MemoryDB,
	JobSqlDBConfig:    jobsql.DefaultJobSqlDBConfig,
	MetaDBType:        model.LevelDB,
	MetaLevelDBConfig: metalevel.DefaultMetaLevelDBConfig,
	MetaSqlDBConfig:   metasql.DefaultMetaSqlDBConfig,
}

func overrideConfigFromEnv(config *StoneHubConfig) {
	config.StorageProvider = model.StorageProvider
}
