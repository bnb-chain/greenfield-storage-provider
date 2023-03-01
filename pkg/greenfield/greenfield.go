package greenfield

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client/chain"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	UpdateClientInternal     = 60
	ListenChainEventInternal = 2
)

// GreenfieldClient the greenfield chain client, only use to query.
type GreenfieldClient struct {
	// TODO: polish it by new sdk version
	gnfdCompositeClients *chain.GnfdCompositeClients
	gnfdCompositeClient  *chain.GnfdCompositeClient
	currentHeight        int64
	updatedAt            time.Time
	Provider             []string
}

func (client *GreenfieldClient) GnfdCompositeClient() *chain.GnfdCompositeClient {
	return client.gnfdCompositeClient
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
			Provider: config.GreenfieldAddrs,
			gnfdCompositeClients: chain.NewGnfdCompositClients(config.GreenfieldAddrs, config.TendermintAddrs, cfg.ChainID,
				chain.WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()))),
		}
		client.gnfdCompositeClient, _ = client.gnfdCompositeClients.GetClient()
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
				maxHeightClient = greenfield.getCurrentClient()
			)
			for _, client := range greenfield.backUpClients {
				gnfdCompositeClient, err := client.gnfdCompositeClients.GetClient()
				if err != nil {
					log.Errorw("get composite client failed ", "node_addr", client.Provider, "error", err)
					continue
				}
				status, err := gnfdCompositeClient.RpcClient.TmClient.Status(context.Background())
				if err != nil {
					log.Errorw("get status failed", "node_addr", client.Provider, "error", err)
					continue
				}
				currentHeight := status.SyncInfo.LatestBlockHeight
				if currentHeight > maxHeight {
					maxHeight = currentHeight
					maxHeightClient = client
					client.gnfdCompositeClient = gnfdCompositeClient
				}
				client.currentHeight = currentHeight
				client.updatedAt = time.Now()
				log.Debugw("chain info", "node_addr", client.Provider, "current_height", currentHeight)
			}
			if maxHeightClient != greenfield.getCurrentClient() {
				greenfield.setCurrentClient(maxHeightClient)
			}
		case <-greenfield.stopCh:
			return
		}
	}
}
