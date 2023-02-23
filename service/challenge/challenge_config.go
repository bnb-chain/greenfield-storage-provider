package challenge

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metalevel"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type ChallengeConfig struct {
	Address           string
	StorageProvider   string
	MetaDBType        string
	MetaLevelDBConfig *config.LevelDBConfig
	MetaSqlDBConfig   *config.SqlDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
}

var DefaultStorageProvider = "bnb-sp"

func DefaultStorageProviderID() string {
	hash := sha256.New()
	hash.Write([]byte(DefaultStorageProvider))
	return hex.EncodeToString(hash.Sum(nil))
}

var DefaultChallengeConfig = &ChallengeConfig{
	Address:           model.DefaultChallengeAddress,
	StorageProvider:   model.StorageProvider,
	MetaDBType:        model.LevelDB,
	MetaLevelDBConfig: metalevel.DefaultMetaLevelDBConfig,
	MetaSqlDBConfig:   metasql.DefaultMetaSqlDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
}

func overrideConfigFromEnv(config *ChallengeConfig) {
	config.StorageProvider = model.StorageProvider
}
