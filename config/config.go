package config

import (
	"bufio"
	"os"

	"github.com/bnb-chain/inscription-storage-provider/service/stonehub"
	"github.com/bnb-chain/inscription-storage-provider/util"

	"github.com/naoina/toml"
)

type StorageProviderConfig struct {
	Service     []string
	StoneHubCfg *stonehub.StoneHubConfig
}

var DefaultStorageProviderConfig = &StorageProviderConfig{
	StoneHubCfg: stonehub.DefaultStoneHubConfig,
}

// LoadSPConfig loads the config file
func LoadSPConfig(file string) *StorageProviderConfig {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	cfg := StorageProviderConfig{}
	err = util.TomlSettings.NewDecoder(bufio.NewReader(f)).Decode(&cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		panic(err)
	}
	return &cfg
}

// Config toml config
type Config struct {
	PieceStore PieceStoreConfig
	Log        LogConfig
}

// PieceStoreConfig contains some parameters which are used to run PieceStore
type PieceStoreConfig struct {
	Shards int           // store the blocks into N buckets by hash of key
	Store  ObjectStorage // config of object storage
}

// ObjectStorage config
type ObjectStorage struct {
	Storage               string // backend storage type (e.g. s3, file, memory)
	BucketURL             string // the bucket URL of object storage to store data
	AccessKey             string // access key for object storage
	SecretKey             string // secret key for object storage
	SessionToken          string // temporary credential used to access backend storage
	NoSignRequest         bool   // whether access public bucket
	MaxRetries            int    // the number of max retries that will be performed
	MinRetryDelay         int64  // the minimum retry delay after which retry will be performed
	TlsInsecureSkipVerify bool   // whether skip the certificate verification of HTTPS requests
}

// LogConfig log config
type LogConfig struct {
	FilePath     string // log file path
	Level        string // log level
	MaxBytesSize int64  // max bytes size of log file
}

// LoadConfig is used to parse toml config file
func LoadConfig(file string) *Config {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	cfg := Config{}
	err = util.TomlSettings.NewDecoder(bufio.NewReader(f)).Decode(&cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		panic(err)
	}
	return &cfg
}
