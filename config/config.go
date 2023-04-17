package config

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/naoina/toml"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/service/blocksyncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	storeconfig "github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// StorageProviderConfig defines the configuration of storage provider
type StorageProviderConfig struct {
	Service           []string
	SpOperatorAddress string
	Endpoint          map[string]string
	ListenAddress     map[string]string
	SpDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
	ChainConfig       *gnfd.GreenfieldChainConfig
	SignerCfg         *signer.SignerConfig
	BlockSyncerCfg    *blocksyncer.Config
	P2PCfg            *p2p.NodeConfig
	LogCfg            *LogConfig
	MetricsCfg        *metrics.MetricsConfig
}

// JSONMarshal marshal the StorageProviderConfig to json format
func (cfg *StorageProviderConfig) JSONMarshal() ([]byte, error) {
	return json.Marshal(cfg)
}

// JSONUnmarshal unmarshal bytes to StorageProviderConfig struct
func (cfg *StorageProviderConfig) JSONUnmarshal(jsonBytes []byte) error {
	return json.Unmarshal(jsonBytes, cfg)
}

// DefaultStorageProviderConfig defines the default configuration of storage provider services
var DefaultStorageProviderConfig = &StorageProviderConfig{
	Service: []string{
		model.GatewayService,
		model.UploaderService,
		model.DownloaderService,
		model.ChallengeService,
		model.TaskNodeService,
		model.ReceiverService,
		model.SignerService,
		model.MetadataService,
		model.ManagerService,
		model.P2PService,
	},
	ListenAddress: map[string]string{
		model.GatewayService:    model.GatewayHTTPAddress,
		model.UploaderService:   model.UploaderGRPCAddress,
		model.DownloaderService: model.DownloaderGRPCAddress,
		model.ChallengeService:  model.ChallengeGRPCAddress,
		model.ReceiverService:   model.ReceiverGRPCAddress,
		model.TaskNodeService:   model.TaskNodeGRPCAddress,
		model.SignerService:     model.SignerGRPCAddress,
		model.MetadataService:   model.MetadataGRPCAddress,
		model.P2PService:        model.P2PGRPCAddress,
	},
	Endpoint: map[string]string{
		model.GatewayService:    "gnfd.test-sp.com",
		model.UploaderService:   model.UploaderGRPCAddress,
		model.DownloaderService: model.DownloaderGRPCAddress,
		model.ChallengeService:  model.ChallengeGRPCAddress,
		model.ReceiverService:   model.ReceiverGRPCAddress,
		model.TaskNodeService:   model.TaskNodeGRPCAddress,
		model.SignerService:     model.SignerGRPCAddress,
		model.MetadataService:   model.MetadataGRPCAddress,
		model.P2PService:        model.P2PGRPCAddress,
	},
	SpOperatorAddress: hex.EncodeToString([]byte(model.SpOperatorAddress)),
	SpDBConfig:        DefaultSQLDBConfig,
	PieceStoreConfig:  DefaultPieceStoreConfig,
	ChainConfig:       DefaultGreenfieldChainConfig,
	SignerCfg:         signer.DefaultSignerChainConfig,
	BlockSyncerCfg:    DefaultBlockSyncerConfig,
	P2PCfg:            DefaultP2PConfig,
	LogCfg:            DefaultLogConfig,
	MetricsCfg:        DefaultMetricsConfig,
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
	Store: storage.ObjectStorageConfig{
		Storage:    "file",
		BucketURL:  "./data",
		MaxRetries: 5,
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

// DefaultBlockSyncerConfig defines the default configuration of BlockSyncer service
var DefaultBlockSyncerConfig = &blocksyncer.Config{
	Modules:        []string{"epoch", "bucket", "object", "payment"},
	Dsn:            "localhost:3308",
	RecreateTables: true,
}

// DefaultMetricsConfig defines the default config of Metrics service
var DefaultMetricsConfig = &metrics.MetricsConfig{
	Enabled:     false,
	HTTPAddress: model.MetricsHTTPAddress,
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

var DefaultP2PConfig = &p2p.NodeConfig{
	ListenAddress: model.P2PListenAddress,
	PingPeriod:    model.DefaultPingPeriod,
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
