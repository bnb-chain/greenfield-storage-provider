package main

import (
	"os"

	"github.com/bnb-chain/inscription-storage-provider/config"
	"github.com/bnb-chain/inscription-storage-provider/service/stonehub"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var newMap = map[string]any{
	"StoneHub": sb,
}

func sb(cfg *config.StorageProviderConfig) any {
	if cfg.StoneHubCfg == nil {
		cfg.StoneHubCfg = config.DefaultStorageProviderConfig.StoneHubCfg
	}
	server, err := stonehub.NewStoneHubService(cfg.StoneHubCfg)
	if err != nil {
		log.Errorw("stone hub init failed", "error", err)
		os.Exit(1)
	}
	return server
}
