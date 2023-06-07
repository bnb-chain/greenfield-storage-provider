package gnfd

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	chainClient "github.com/bnb-chain/greenfield/sdk/client"
	chttp "github.com/cometbft/cometbft/rpc/client/http"
)

const (
	GreenFieldChain = "GreenfieldChain"
	// UpdateClientInternal defines the period of updating the best chain client
	UpdateClientInternal = 60
	// ExpectedOutputBlockInternal defines the time of estimating output block time
	ExpectedOutputBlockInternal = 2
)

var (
	ErrNoSuchBucket = gfsperrors.Register(GreenFieldChain, http.StatusBadRequest, 500001, "no such bucket")
	ErrSealTimeout  = gfsperrors.Register(GreenFieldChain, http.StatusBadRequest, 500002, "seal failed")
)

// GreenfieldClient the greenfield chain client, only use to query.
type GreenfieldClient struct {
	chainClient   *chainClient.GreenfieldClient
	currentHeight int64
	updatedAt     time.Time
	Provider      string
}

// GnfdClient returns the greenfield chain client.
func (client *GreenfieldClient) GnfdClient() *chainClient.GreenfieldClient {
	return client.chainClient
}

var _ consensus.Consensus = &Gnfd{}

type GnfdChainConfig struct {
	ChainID      string
	ChainAddress []string
}

type Gnfd struct {
	client          *GreenfieldClient
	backUpClients   []*GreenfieldClient
	wsClient        *chttp.HTTP
	backUpWsClients []*chttp.HTTP
	stopCh          chan struct{}
	mutex           sync.RWMutex
}

// NewGnfd returns the Greenfield instance.
func NewGnfd(cfg *GnfdChainConfig) (*Gnfd, error) {
	if len(cfg.ChainAddress) == 0 {
		return nil, errors.New("greenfield nodes missing")
	}
	var clients []*GreenfieldClient
	var wsClients []*chttp.HTTP
	for _, address := range cfg.ChainAddress {
		cc, err := chainClient.NewGreenfieldClient(address, cfg.ChainID)
		if err != nil {
			return nil, err
		}
		client := &GreenfieldClient{
			Provider:    address,
			chainClient: cc,
		}
		clients = append(clients, client)
		wsClient, err := chttp.New(address, "/websocket")
		if err != nil {
			return nil, err
		}
		wsClients = append(wsClients, wsClient)
	}
	greenfield := &Gnfd{
		client:          clients[0],
		backUpClients:   clients,
		wsClient:        wsClients[0],
		backUpWsClients: wsClients,
		stopCh:          make(chan struct{}),
	}

	go greenfield.updateClient()
	return greenfield, nil
}

// Close the Greenfield instance.
func (g *Gnfd) Close() error {
	close(g.stopCh)
	return nil
}

// getCurrentClient returns the current client to use.
func (g *Gnfd) getCurrentClient() *GreenfieldClient {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.client
}

// setCurrentClient sets client to current client for using.
func (g *Gnfd) setCurrentClient(client *GreenfieldClient) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.client = client
}

// getCurrentWsClient returns the current websocket client to get last block height use.
func (g *Gnfd) getCurrentWsClient() *chttp.HTTP {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.wsClient
}

// setCurrentWsAddress sets client to current websocket client for get last block height using.
func (g *Gnfd) setCurrentWsAddress(client *chttp.HTTP) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	g.wsClient = client
}

// updateClient selects the client that block height is the largest and set to current client.
func (g *Gnfd) updateClient() {
	ticker := time.NewTicker(UpdateClientInternal * time.Second)
	for {
		select {
		case <-ticker.C:
			var (
				maxHeight, _      = g.CurrentHeight(context.Background())
				maxHeightClient   = g.getCurrentClient()
				maxHeightWsClient = g.getCurrentWsClient()
			)
			for idx, client := range g.backUpWsClients {
				info, err := client.ABCIInfo(context.Background())
				if err != nil {
					log.Errorw("failed to get latest block height",
						"node_addr", g.backUpClients[idx].Provider, "error", err)
					continue
				}
				currentHeight := uint64(info.Response.LastBlockHeight)
				if currentHeight > maxHeight {
					maxHeight = currentHeight
					maxHeightClient = g.backUpClients[idx]
					maxHeightWsClient = client
				}
				g.backUpClients[idx].currentHeight = (int64)(currentHeight)
				g.backUpClients[idx].updatedAt = time.Now()
			}
			if maxHeightClient != g.getCurrentClient() {
				g.setCurrentClient(maxHeightClient)
				g.setCurrentWsAddress(maxHeightWsClient)
			}
		case <-g.stopCh:
			return
		}
	}
}
