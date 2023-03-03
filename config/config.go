package config

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/service/blocksyncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	"github.com/naoina/toml"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	storeconfig "github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// StorageProviderConfig defines the configuration of storage provider
type StorageProviderConfig struct {
	Service           []string
	SpOperatorAddress string
	Domain            string
	HTTPAddress       map[string]string
	GRPCAddress       map[string]string
	SpDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
	ChainConfig       *gnfd.GreenfieldChainConfig
	SignerCfg         *signer.SignerConfig
	BlockSyncerCfg    *tomlconfig.TomlConfig
	LogCfg            *LogConfig
}

// JsonMarshal marshal the StorageProviderConfig to json format
func (cfg *StorageProviderConfig) JsonMarshal() ([]byte, error) {
	return json.Marshal(cfg)
}

// JsonUnMarshal unmarshal bytes to StorageProviderConfig struct
func (cfg *StorageProviderConfig) JsonUnMarshal(jsonBytes []byte) error {
	return json.Unmarshal(jsonBytes, cfg)
}

// DefaultStorageProviderConfig defines the default configuration of storage provider services
var DefaultStorageProviderConfig = &StorageProviderConfig{
	Service: []string{
		model.GatewayService,
		model.UploaderService,
		model.DownloaderService,
		model.ChallengeService,
		model.StoneNodeService,
		model.SyncerService,
		model.SignerService,
	},
	GRPCAddress: map[string]string{
		model.UploaderService:   model.UploaderGRPCAddress,
		model.DownloaderService: model.DownloaderGRPCAddress,
		model.ChallengeService:  model.ChallengeGRPCAddress,
		model.SyncerService:     model.SyncerGRPCAddress,
		model.StoneNodeService:  model.StoneNodeGRPCAddress,
		model.SignerService:     model.SignerGRPCAddress,
	},
	HTTPAddress: map[string]string{
		model.GatewayService:  model.GatewayHTTPAddress,
		model.MetadataService: model.MetaDataServiceHTTPAddress,
	},
	SpOperatorAddress: hex.EncodeToString([]byte("greenfield-storage-provider")),
	Domain:            "gnfd.nodereal.com",
	SpDBConfig:        DefaultSQLDBConfig,
	PieceStoreConfig:  DefaultPieceStoreConfig,
	ChainConfig:       DefaultGreenfieldChainConfig,
	SignerCfg:         signer.DefaultSignerChainConfig,
	BlockSyncerCfg:    blocksyncer.DefaultBlockSyncerConfig,
	LogCfg:            DefaultLogConfig,
}

// DefaultSQLDBConfig defines the default configuration of SQL DB
var DefaultSQLDBConfig = &storeconfig.SQLDBConfig{
	User:     "root",
	Passwd:   "test_pwd",
	Address:  "localhost:3306",
	Database: "storage_provider_db",
}

// DefaultPieceStoreConfig defines the default configuration of piece store
var DefaultPieceStoreConfig = &storage.PieceStoreConfig{
	Shards: 0,
	Store: &storage.ObjectStorageConfig{
		Storage:               "file",
		BucketURL:             "./data",
		NoSignRequest:         false,
		MaxRetries:            5,
		MinRetryDelay:         0,
		TlsInsecureSkipVerify: false,
	},
}

// DefaultGreenfieldChainConfig defines the default configuration of greenfield chain
var DefaultGreenfieldChainConfig = &gnfd.GreenfieldChainConfig{
	ChainID: model.GreenfieldChainID,
	NodeAddr: []*gnfd.NodeConfig{{
		GreenfieldAddresses: []string{model.GreenfieldAddress},
		TendermintAddresses: []string{model.TendermintAddress},
	}},
}

type LogConfig struct {
	Level string
	Path  string
}

// DefaultLogConfig defines the default configuration of log
var DefaultLogConfig = &LogConfig{
	// TODO:: change to info level after releasing
	Level: "debug",
	Path:  "./gnfd-sp.log",
}

// LoadConfig loads the config file from path
func LoadConfig(path string) *StorageProviderConfig {
	f, err := os.Open(path)
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

// SaveConfig write the config to disk
func SaveConfig(file string, cfg *StorageProviderConfig) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	encode := util.TomlSettings.NewEncoder(f)
	if err = encode.Encode(cfg); err != nil {
		return err
	}
	return nil
}
