package config

import (
	"bufio"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/service/blocksyncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	"github.com/naoina/toml"
)

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

var DefaultStorageProviderConfig = &StorageProviderConfig{
	GatewayCfg:     gateway.DefaultGatewayConfig,
	UploaderCfg:    uploader.DefaultUploaderConfig,
	DownloaderCfg:  downloader.DefaultDownloaderConfig,
	ChallengeCfg:   challenge.DefaultChallengeConfig,
	StoneNodeCfg:   stonenode.DefaultStoneNodeConfig,
	SyncerCfg:      syncer.DefaultSyncerConfig,
	SignerCfg:      signer.DefaultSignerChainConfig,
	MetadataCfg:    metadata.DefaultMetadataConfig,
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
