package config

import (
	"bufio"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	"github.com/naoina/toml"

	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// StorageProviderConfig defines the configuration of storage provider
type StorageProviderConfig struct {
	Service        []string
	GatewayCfg     *gateway.GatewayConfig
	UploaderCfg    *uploader.UploaderConfig
	DownloaderCfg  *downloader.DownloaderConfig
	ChallengeCfg   *challenge.ChallengeConfig
	StoneNodeCfg   *stonenode.StoneNodeConfig
	SyncerCfg      *syncer.SyncerConfig
	SignerCfg      *signer.SignerConfig
	MetadataCfg    *metadata.MetadataConfig
	BlockSyncerCfg *tomlconfig.TomlConfig
}

// DefaultStorageProviderConfig defines the default configuration of storage provider services
var DefaultStorageProviderConfig = &StorageProviderConfig{
	GatewayCfg:    gateway.DefaultGatewayConfig,
	UploaderCfg:   uploader.DefaultUploaderConfig,
	DownloaderCfg: downloader.DefaultDownloaderConfig,
	ChallengeCfg:  challenge.DefaultChallengeConfig,
	StoneNodeCfg:  stonenode.DefaultStoneNodeConfig,
	SyncerCfg:     syncer.DefaultSyncerConfig,
	SignerCfg:     signer.DefaultSignerChainConfig,
	Service: []string{
		model.GatewayService,
		model.UploaderService,
		model.DownloaderService,
		model.ChallengeService,
		model.StoneNodeService,
		model.SyncerService,
		model.SignerService,
	},
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
