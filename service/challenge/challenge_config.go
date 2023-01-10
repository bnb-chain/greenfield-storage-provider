package challenge

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb/leveldb"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
)

type ChallengeConfig struct {
	Address         string
	StorageProvider string
	MetaType        string
	MetaDB          *leveldb.MetaLevelDBConfig
	PieceConfig     *storage.PieceStoreConfig
}

var DefaultStorageProvider = "bnb-sp"

func DefaultStorageProviderID() string {
	hash := sha256.New()
	hash.Write([]byte(DefaultStorageProvider))
	return hex.EncodeToString(hash.Sum(nil))
}

var DefaultChallengeConfig = &ChallengeConfig{
	Address:         "127.0.0.1:5423",
	StorageProvider: DefaultStorageProviderID(),
	MetaType:        model.LevelDB,
	MetaDB:          leveldb.DefaultMetaLevelDBConfig,
	PieceConfig:     storage.DefaultPieceStoreConfig,
}
