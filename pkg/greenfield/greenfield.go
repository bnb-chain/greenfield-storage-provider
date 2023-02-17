package greenfield

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client/chain"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

const (
	UpdateClientInternal     = 60
	ListenChainEventInternal = 2
)

// GreenfieldClient the greenfield chain client, only use to query.
type GreenfieldClient struct {
	greenfieldClient chain.GreenfieldClient
	tendermintClient chain.TendermintClient
	currentHeight    int64
	updatedAt        time.Time
	Provider         string
}

func (client *GreenfieldClient) Greenfield() chain.GreenfieldClient {
	return client.greenfieldClient
}

func (client *GreenfieldClient) Tendermint() chain.TendermintClient {
	return client.tendermintClient
}

// Greenfield the greenfield chain service.
type Greenfield struct {
	config        *GreenfieldChainConfig
	client        *GreenfieldClient
	backUpClients []*GreenfieldClient
	stopCh        chan struct{}
	mutex         sync.RWMutex
}

// NewGreenfield return the Greenfield instance.
func NewGreenfield(cfg *GreenfieldChainConfig) (*Greenfield, error) {
	if len(cfg.NodeAddr) == 0 {
		return nil, errors.New("greenfield nodes missing")
	}
	var clients []*GreenfieldClient
	for _, config := range cfg.NodeAddr {
		client := &GreenfieldClient{
			Provider:         config.GreenfieldAddr,
			greenfieldClient: chain.NewGreenfieldClient(config.GreenfieldAddr, cfg.ChainID),
			tendermintClient: chain.NewTendermintClient(config.TendermintAddr),
		}
		clients = append(clients, client)
	}
	greenfield := &Greenfield{
		config:        cfg,
		client:        clients[0],
		backUpClients: clients,
		stopCh:        make(chan struct{}),
	}

	go greenfield.updateClient()
	return greenfield, nil
}

// Close the Greenfield instance.
func (greenfield *Greenfield) Close() error {
	close(greenfield.stopCh)
	return nil
}

// GetGreenfieldClient return one greenfield client.
func (greenfield *Greenfield) GetGreenfieldClient() *GreenfieldClient {
	greenfield.mutex.RLock()
	defer greenfield.mutex.RUnlock()
	return greenfield.client
}

// getCurrentClient return current use client.
func (greenfield *Greenfield) getCurrentClient() *GreenfieldClient {
	greenfield.mutex.RLock()
	defer greenfield.mutex.RUnlock()
	return greenfield.client
}

// setCurrentClient set current use client.
func (greenfield *Greenfield) setCurrentClient(client *GreenfieldClient) {
	greenfield.mutex.Lock()
	defer greenfield.mutex.Unlock()
	greenfield.client = client
}

// updateClient select block height is the largest from all clients and update to current client.
func (greenfield *Greenfield) updateClient() {
	ticker := time.NewTicker(UpdateClientInternal * time.Second)
	for {
		select {
		case <-ticker.C:
			var (
				maxHeight       int64
				maxHeightCLient = greenfield.getCurrentClient()
			)
			for _, client := range greenfield.backUpClients {
				chainInfo, err := client.tendermintClient.TmClient.Status(context.Background())
				if err != nil {
					log.Errorw("get chain info error", "node_addr", client.Provider, "error", err)
					continue
				}
				if chainInfo == nil {
					log.Errorw("get chain info nil", "node_addr", client.Provider)
					continue
				}
				currentHeight := chainInfo.SyncInfo.LatestBlockHeight
				if currentHeight > maxHeight {
					maxHeight = currentHeight
					maxHeightCLient = client
				}
				client.currentHeight = currentHeight
				client.updatedAt = time.Now()
				log.Debugw("chain info", "node_addr", client.Provider, "current_height", currentHeight)
			}
			if maxHeightCLient != greenfield.getCurrentClient() {
				greenfield.setCurrentClient(maxHeightCLient)
			}
		case <-greenfield.stopCh:
			return
		}
	}
}
