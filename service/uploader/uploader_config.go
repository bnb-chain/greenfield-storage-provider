package uploader

import "github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"

type UploaderConfig struct {
	StorageProvider string
	Address         string
	LogConfig       struct {
		FilePath string
		Level    string
	}
	StoneHubConfig   stoneHubClientConfig
	PieceStoreConfig storage.PieceStoreConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	StorageProvider: "bnb-sp",
	Address:         "127.0.0.1:5311",
	LogConfig: struct {
		FilePath string
		Level    string
	}{FilePath: "./debug/uploader.log", Level: "info"},
	StoneHubConfig: struct {
		Address    string
		TimeoutSec uint32
	}{Address: "127.0.0.1:5323", TimeoutSec: 5},
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
}
