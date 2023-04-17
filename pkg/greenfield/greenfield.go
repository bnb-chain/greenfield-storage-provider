package greenfield

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	chainClient "github.com/bnb-chain/greenfield/sdk/client"
)

const (
	// UpdateClientInternal defines the period of updating the best chain client
	UpdateClientInternal = 60
	// ExpectedOutputBlockInternal defines the time of estimating output block time
	ExpectedOutputBlockInternal = 2
)

// GreenfieldClient the greenfield chain client, only use to query.
type GreenfieldClient struct {
	chainClient   *chainClient.GreenfieldClient
	currentHeight int64
	updatedAt     time.Time
	Provider      []string
}

// GnfdClient return the greenfield chain client
func (client *GreenfieldClient) GnfdClient() *chainClient.GreenfieldClient {
	return client.chainClient
}

// Greenfield is an encapsulation of greenfield chain go sdk which supports for more query request
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
		cc, err := chainClient.NewGreenfieldClient(config.TendermintAddresses[0], cfg.ChainID)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		client := &GreenfieldClient{
			Provider:    config.GreenfieldAddresses,
			chainClient: cc,
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

// getCurrentClient return the current client to use.
func (greenfield *Greenfield) getCurrentClient() *GreenfieldClient {
	greenfield.mutex.RLock()
	defer greenfield.mutex.RUnlock()
	return greenfield.client
}

// setCurrentClient set client to current client for using.
func (greenfield *Greenfield) setCurrentClient(client *GreenfieldClient) {
	greenfield.mutex.Lock()
	defer greenfield.mutex.Unlock()
	greenfield.client = client
}

// updateClient select the client that block height is the largest and set to current client.
func (greenfield *Greenfield) updateClient() {
	ticker := time.NewTicker(UpdateClientInternal * time.Second)
	for {
		select {
		case <-ticker.C:
			var (
				maxHeight       uint64
				maxHeightClient = greenfield.getCurrentClient()
			)
			for _, client := range greenfield.backUpClients {
				currentHeight, err := greenfield.GetCurrentHeight(context.Background())
				if err != nil {
					log.Errorw("get latest block height failed", "node_addr", client.Provider, "error", err)
					continue
				}
				if currentHeight > maxHeight {
					maxHeight = currentHeight
					maxHeightClient = client
				}
				client.currentHeight = (int64)(currentHeight)
				client.updatedAt = time.Now()
			}
			if maxHeightClient != greenfield.getCurrentClient() {
				greenfield.setCurrentClient(maxHeightClient)
			}
		case <-greenfield.stopCh:
			return
		}
	}
}
