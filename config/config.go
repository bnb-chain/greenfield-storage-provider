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
