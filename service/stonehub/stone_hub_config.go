package stonehub

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb/jobsql"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb/leveldb"
)

type StoneHubConfig struct {
	StorageProvider string
	Address         string
	JobDBType       string
	JobDB           *jobsql.DBOption
	MetaDBType      string
	MetaDB          *leveldb.MetaLevelDBConfig
}

var DefaultStorageProvider = "bnb-sp"

func DefaultStorageProviderID() string {
	hash := sha256.New()
	hash.Write([]byte(DefaultStorageProvider))
	return hex.EncodeToString(hash.Sum(nil))
}

var DefaultStoneHubConfig = &StoneHubConfig{
	StorageProvider: DefaultStorageProviderID(),
	Address:         "127.0.0.1:5323",
	JobDBType:       model.MemoryDB,
	JobDB: &jobsql.DBOption{
		User:     "root",
		Passwd:   "bfs-test",
		Address:  "127.0.0.1:3306",
		Database: "job_context",
	},
	MetaDBType: model.LevelDB,
	MetaDB:     leveldb.DefaultMetaLevelDBConfig,
}
