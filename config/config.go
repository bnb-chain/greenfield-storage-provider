package config

import (
	"bufio"
	"os"
	"path/filepath"

	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	"github.com/naoina/toml"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/blocksyncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/service/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonehub"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

type StorageProviderConfig struct {
	Service        []string
	GatewayCfg     *gateway.GatewayConfig
	UploaderCfg    *uploader.UploaderConfig
	DownloaderCfg  *downloader.DownloaderConfig
	ChallengeCfg   *challenge.ChallengeConfig
	StoneHubCfg    *stonehub.StoneHubConfig
	StoneNodeCfg   *stonenode.StoneNodeConfig
	SyncerCfg      *syncer.SyncerConfig
	SignerCfg      *signer.SignerConfig
	MetadataCfg    *metadata.MetadataConfig
	P2PCfg         *p2p.P2PServiceConfig
	BlockSyncerCfg *tomlconfig.TomlConfig
}

var DefaultStorageProviderConfig = &StorageProviderConfig{
	Service: []string{
		model.GatewayService,
		model.UploaderService,
		model.DownloaderService,
		model.ChallengeService,
		model.SyncerService,
		model.StoneHubService,
		model.StoneNodeService,
		model.SignerService,
	},
	GatewayCfg:     gateway.DefaultGatewayConfig,
	UploaderCfg:    uploader.DefaultUploaderConfig,
	DownloaderCfg:  downloader.DefaultDownloaderConfig,
	ChallengeCfg:   challenge.DefaultChallengeConfig,
	StoneHubCfg:    stonehub.DefaultStoneHubConfig,
	StoneNodeCfg:   stonenode.DefaultStoneNodeConfig,
	SyncerCfg:      syncer.DefaultSyncerConfig,
	SignerCfg:      signer.DefaultSignerChainConfig,
	MetadataCfg:    metadata.DefaultMetadataConfig,
	P2PCfg:         p2p.DefaultP2PServiceConfig,
	BlockSyncerCfg: blocksyncer.DefaultBlockSyncerConfig,
}

// LoadConfig loads the config file
func LoadConfig(file string) *StorageProviderConfig {
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

// SaveConfig write the config to disk
func SaveConfig(file string, cfg *StorageProviderConfig) error {
	path := filepath.Join(file, "config.toml")
	f, err := os.Create(path)
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
