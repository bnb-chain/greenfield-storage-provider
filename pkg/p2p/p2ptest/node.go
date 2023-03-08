package main

import (
	"bufio"
	"context"
	"os"
	"syscall"
	"time"

	sdkmath "cosmossdk.io/math"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

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
	node.PeersProvider().UpdateSp([]string{"test_sp_operator_address_1", "test_sp_operator_address_2"})
	// start p2p node
	slc := lifecycle.NewServiceLifecycle()
	slc.RegisterServices(node)
	slcCtx := context.Background()
	// background send get approval request period
	go func() {
		object := &storagetypes.ObjectInfo{}
		sleepTimer := time.NewTimer(5 * time.Second)
		approvalTicker := time.NewTicker(1 * time.Second)
		begin := false
		for {
			select {
			case <-sleepTimer.C:
				begin = true
			case <-approvalTicker.C:
				if !begin {
					continue
				}
				object.Id = sdkmath.NewUint(uint64(time.Now().Unix()))
				accept, refuse, err := node.GetApproval(object, 9, 10)
				if err != nil {
					log.Errorw("failed to get approval from background", "error", err)
					continue
				}
				log.Infow("success collect get approval response", "accept", len(accept), "refuse", len(refuse))
				for _, response := range accept {
					log.Infow("receive get approval accept", "remote_sp", response.GetSpOperatorAddress(), "object_id", response.GetObjectInfo().Id.Uint64())
				}
				for _, response := range refuse {
					log.Infow("receive get approval refuse", "remote_sp", response.GetSpOperatorAddress(), "object_id", response.GetObjectInfo().Id.Uint64())
				}
			}
		}
	}()
	slc.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).StartServices(slcCtx).Wait(slcCtx)
}
