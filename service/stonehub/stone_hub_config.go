package stonehub

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobsql"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metalevel"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
)

type StoneHubConfig struct {
	StorageProvider      string
	Address              string
	SignerServiceAddress string
	ChainConfig          *gnfd.GreenfieldChainConfig
	JobDBType            string
	JobSqlDBConfig       *config.SqlDBConfig
	MetaDBType           string
	MetaLevelDBConfig    *config.LevelDBConfig
	MetaSqlDBConfig      *config.SqlDBConfig
}

var DefaultStorageProvider = "bnb-sp"

func DefaultStorageProviderID() string {
	hash := sha256.New()
	hash.Write([]byte(DefaultStorageProvider))
	return hex.EncodeToString(hash.Sum(nil))
}

var DefaultStoneHubConfig = &StoneHubConfig{
	StorageProvider:      DefaultStorageProviderID(),
	Address:              "127.0.0.1:9333",
	SignerServiceAddress: "127.0.0.1:9633",
	ChainConfig:          gnfd.DefaultGreenfieldChainConfig,
	JobDBType:            model.MemoryDB,
	JobSqlDBConfig:       jobsql.DefaultJobSqlDBConfig,
	MetaDBType:           model.LevelDB,
	MetaLevelDBConfig:    metalevel.DefaultMetaLevelDBConfig,
	MetaSqlDBConfig:      metasql.DefaultMetaSqlDBConfig,
}
