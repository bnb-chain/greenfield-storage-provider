package stonehub

import (
	"crypto/sha256"
	"encoding/hex"
)

var (
	MemoryDB string = "memory"
	MySqlDB  string = "MySql"
	LevelDB  string = "leveldb"
)

type StoneHubConfig struct {
	StorageProvider string
	Address         string
	JobDB           string
	MetaDB          string
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
	JobDB:           MemoryDB,
	MetaDB:          LevelDB,
}
