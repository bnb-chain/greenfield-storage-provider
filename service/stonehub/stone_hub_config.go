package stonehub

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/bnb-chain/inscription-storage-provider/store/jobdb/jobsql"
)

var (
	MemoryDB string = "memory"
	MySqlDB  string = "MySql"
	LevelDB  string = "leveldb"
)

type StoneHubConfig struct {
	StorageProvider string
	Address         string
	JobDBType       string
	JobDB           *jobsql.DBOption
	MetaDBType      string
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
	JobDBType:       MemoryDB,
	JobDB: &jobsql.DBOption{
		User:     "root",
		Passwd:   "bfs-test",
		Address:  "127.0.0.1:3306",
		Database: "job_context",
	},
	MetaDBType: LevelDB,
}
