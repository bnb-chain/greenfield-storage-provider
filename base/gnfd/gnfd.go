package gnfd

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	chainClient "github.com/bnb-chain/greenfield/sdk/client"
)

const (
	GreenFieldChain = "GreenfieldChain"
	// UpdateClientInternal defines the period of updating the best chain client
	UpdateClientInternal = 60
	// ExpectedOutputBlockInternal defines the time of estimating output block time
	ExpectedOutputBlockInternal = 2
)

var (
	ErrNoSuchBucket = gfsperrors.Register(GreenFieldChain, http.StatusInternalServerError, 500001, "no such bucket")
	ErrSealTimeout  = gfsperrors.Register(GreenFieldChain, http.StatusInternalServerError, 500002, "seal failed")
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
	client        *GreenfieldClient
	backUpClients []*GreenfieldClient
	stopCh        chan struct{}
	mutex         sync.RWMutex
}

// NewGnfd returns the Greenfield instance.
func NewGnfd(cfg *GnfdChainConfig) (*Gnfd, error) {
	if len(cfg.ChainAddress) == 0 {
		return nil, errors.New("greenfield nodes missing")
	}
	var clients []*GreenfieldClient
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
	}
	greenfield := &Gnfd{
		client:        clients[0],
		backUpClients: clients,
		stopCh:        make(chan struct{}),
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

// updateClient selects the client that block height is the largest and set to current client.
func (g *Gnfd) updateClient() {
	ticker := time.NewTicker(UpdateClientInternal * time.Second)
	for {
		select {
		case <-ticker.C:
			var (
				maxHeight, _    = g.CurrentHeight(context.Background())
				maxHeightClient = g.getCurrentClient()
			)
			for _, client := range g.backUpClients {
				resp, err := client.GnfdClient().TmClient.GetLatestBlock(
					context.Background(),
					&tmservice.GetLatestBlockRequest{})
				if err != nil {
					log.Errorw("failed to get latest block height", "node_addr", client.Provider, "error", err)
					continue
				}
				currentHeight := (uint64)(resp.SdkBlock.Header.Height)
				if currentHeight > maxHeight {
					maxHeight = currentHeight
					maxHeightClient = client
				}
				client.currentHeight = (int64)(currentHeight)
				client.updatedAt = time.Now()
			}
			if maxHeightClient != g.getCurrentClient() {
				g.setCurrentClient(maxHeightClient)
			}
		case <-g.stopCh:
			return
		}
	}
}
