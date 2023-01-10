package uploader

import (
	"github.com/bnb-chain/inscription-storage-provider/store/metadb/leveldb"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
)

type UploaderConfig struct {
	StorageProvider        string
	Address                string
	StoneHubServiceAddress string
	PieceStoreConfig       storage.PieceStoreConfig
	MetaDBConfig           *leveldb.MetaLevelDBConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	StorageProvider:        "bnb-sp",
	Address:                "127.0.0.1:5311",
	StoneHubServiceAddress: "127.0.0.1:5323",
	PieceStoreConfig: struct {
		Shards int
		Store  *storage.ObjectStorageConfig
	}{Shards: 0, Store: &storage.ObjectStorageConfig{
		Storage:               "file",
		BucketURL:             "./debug",
		AccessKey:             "",
		SecretKey:             "",
		SessionToken:          "",
		NoSignRequest:         false,
		MaxRetries:            3,
		MinRetryDelay:         0,
		TlsInsecureSkipVerify: false,
	}},
	MetaDBConfig: leveldb.DefaultMetaLevelDBConfig,
}
