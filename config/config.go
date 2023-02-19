package config

import (
	"bufio"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/service/p2p"
	"github.com/naoina/toml"

	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonehub"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

type StorageProviderConfig struct {
	Service       []string
	GatewayCfg    *gateway.GatewayConfig
	UploaderCfg   *uploader.UploaderConfig
	DownloaderCfg *downloader.DownloaderConfig
	ChallengeCfg  *challenge.ChallengeConfig
	StoneHubCfg   *stonehub.StoneHubConfig
	StoneNodeCfg  *stonenode.StoneNodeConfig
	SyncerCfg     *syncer.SyncerConfig
	P2PCfg        *p2p.P2PServiceConfig
}

var DefaultStorageProviderConfig = &StorageProviderConfig{
	GatewayCfg:    gateway.DefaultGatewayConfig,
	UploaderCfg:   uploader.DefaultUploaderConfig,
	DownloaderCfg: downloader.DefaultDownloaderConfig,
	ChallengeCfg:  challenge.DefaultChallengeConfig,
	StoneHubCfg:   stonehub.DefaultStoneHubConfig,
	StoneNodeCfg:  stonenode.DefaultStoneNodeConfig,
	SyncerCfg:     syncer.DefaultSyncerConfig,
	P2PCfg:        p2p.DefaultP2PServiceConfig,
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
