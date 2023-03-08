package main

import (
	"bufio"
	"context"
	"os"
	"syscall"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

func main() {
	log.SetLevel(log.InfoLevel)
	// parser config
	f, err := os.Open("./config.toml")
	if err != nil {
		log.Errorw("failed to open config file", "error", err)
		return
	}
	cfg := &p2p.NodeConfig{}
	err = util.TomlSettings.NewDecoder(bufio.NewReader(f)).Decode(&cfg)
	if err != nil {
		log.Errorw("failed to load config", "error", err)
		return
	}
	// init p2p node
	node, err := p2p.NewNode(cfg)
	if err != nil {
		log.Errorw("failed to init p2p node", "error", err)
	}
	// start p2p node
	slc := lifecycle.NewServiceLifecycle()
	slc.RegisterServices(node)
	slcCtx := context.Background()
	slc.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).StartServices(slcCtx).Wait(slcCtx)
}
