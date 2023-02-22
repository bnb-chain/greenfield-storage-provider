package node

import (
	"context"
	"fmt"
	"strings"

	p2p "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs"
	tmlog "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/service"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/pex"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/node/provider"
	storeconf "github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/netdb"
)

const (
	P2PNodeDBPath = "/p2p"
	P2PNodeDBName = "node"
)

type P2PNode struct {
	service.BaseService
	logger tmlog.Logger

	// config
	config *NodeConfig

	// network
	peerManager *p2p.PeerManager
	router      *p2p.Router
	nodeKey     types.NodeKey // our node privkey
	isListening bool

	// services
	pexReactor      service.Service // for exchanging peer addresses
	customerReactor service.Service // for communication between node providers
	shutdownOps     closer
}

type MakeReactor func(peerManager *p2p.PeerManager, chCreator p2p.ChannelCreator,
	peerEvents p2p.PeerEventSubscriber, peersVerifier provider.ProviderQuerier) service.Service

// NewDefault constructs a node service for use in go
// process that host their own process-local node.
func NewDefault(ctx context.Context, cfg *NodeConfig, newReactor MakeReactor) (service.Service, service.Service, error) {
	cfg.EnsureRoot()
	nodeKey, err := types.LoadNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load or gen node key %s: %w", cfg.NodeKeyFile(), err)
	}
	logger, err := tmlog.NewDefaultLogger(tmlog.LogFormatPlain, tmlog.LogLevelInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init log, error: %w", err)
	}
	return makeNode(ctx, cfg, DefaultDBProvider, newReactor, logger, nodeKey)
}

// makeNode returns a new seed node, containing only p2p, pex reactor
func makeNode(ctx context.Context, cfg *NodeConfig, dbProvider DBProvider, newReactor MakeReactor,
	logger tmlog.Logger, nodeKey types.NodeKey) (service.Service, service.Service, error) {
	nodeInfo, err := makeNodeInfo(cfg, nodeKey)
	if err != nil {
		return nil, nil, err
	}
	// Setup Transport and Switch.
	p2pMetrics := p2p.PrometheusMetrics("sp", "p2p", "main")

	peerManager, closer, err := createPeerManager(cfg, dbProvider, nodeKey.ID, p2pMetrics)
	if err != nil {
		return nil, nil, combineCloseError(
			fmt.Errorf("failed to create peer manager: %w", err),
			closer)
	}

	router, err := createRouter(logger, p2pMetrics, func() *types.NodeInfo { return &nodeInfo }, nodeKey, peerManager, cfg)
	if err != nil {
		return nil, nil, combineCloseError(
			fmt.Errorf("failed to create router: %w", err),
			closer)
	}

	node := &P2PNode{
		config: cfg,
		logger: logger,

		nodeKey:     nodeKey,
		peerManager: peerManager,
		router:      router,

		shutdownOps: closer,
	}

	if cfg.P2P.PexReactor {
		node.pexReactor = pex.NewReactor(logger, peerManager, router.OpenChannel, peerManager.Subscribe)
	}

	db, err := netdb.NewNetDB(&storeconf.LevelDBConfig{Path: cfg.DBDir() + P2PNodeDBPath, NameSpace: P2PNodeDBName})
	if err != nil {
		return nil, nil, err
	}
	providerQuerier := provider.NewProviderQuerier(cfg.P2P.PersistentPeers, logger, db)
	node.customerReactor = newReactor(peerManager, router.OpenChannel, peerManager.Subscribe, providerQuerier)
	node.BaseService = *service.NewBaseService(logger, "node", node)

	return node, node.customerReactor, nil
}

func (n *P2PNode) GetNodeId() types.NodeID {
	return n.nodeKey.ID
}

// OnStart starts the Seed Node. It implements service.Service.
func (n *P2PNode) OnStart(ctx context.Context) error {

	// Start the transport.
	if err := n.router.Start(ctx); err != nil {
		return err
	}
	n.isListening = true

	if n.config.P2P.PexReactor {
		if err := n.pexReactor.Start(ctx); err != nil {
			return err
		}
	}
	if err := n.customerReactor.Start(ctx); err != nil {
		return err
	}

	return nil
}

// OnStop stops the Seed Node. It implements service.Service.
func (n *P2PNode) OnStop() {
	n.logger.Info("Stopping Node")

	n.customerReactor.Wait()
	n.pexReactor.Wait()
	n.router.Wait()
	n.isListening = false

	if err := n.shutdownOps(); err != nil {
		if strings.TrimSpace(err.Error()) != "" {
			n.logger.Error("problem shutting down additional services", "err", err)
		}
	}
}
