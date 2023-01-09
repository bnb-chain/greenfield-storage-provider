package config

import (
	"bufio"
	"os"

	"github.com/naoina/toml"

	"github.com/bnb-chain/inscription-storage-provider/service/gateway"
	"github.com/bnb-chain/inscription-storage-provider/service/stonehub"
	"github.com/bnb-chain/inscription-storage-provider/service/stonenode"
	"github.com/bnb-chain/inscription-storage-provider/service/syncer"
	"github.com/bnb-chain/inscription-storage-provider/service/uploader"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/inscription-storage-provider/util"
)

type StorageProviderConfig struct {
	Service          []string
	StoneHubCfg      *stonehub.StoneHubConfig
	PieceStoreConfig *storage.PieceStoreConfig
	GatewayCfg       *gateway.GatewayConfig
	UploaderCfg      *uploader.UploaderConfig
	StoneNodeCfg     *stonenode.StoneNodeConfig
	SyncerCfg        *syncer.SyncerConfig
}

var DefaultStorageProviderConfig = &StorageProviderConfig{
	StoneHubCfg:      stonehub.DefaultStoneHubConfig,
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
	GatewayCfg:       gateway.DefaultGatewayConfig,
	UploaderCfg:      uploader.DefaultUploaderConfig,
	StoneNodeCfg:     stonenode.DefaultStoneNodeConfig,
	SyncerCfg:        syncer.DefaultSyncerConfig,
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
