package node

import (
	"fmt"
	"time"

	tmlog "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/log"
	"github.com/tendermint/tendermint/version"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/conn"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/pex"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/types"
	tmstrings "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/node/strings"
)

type closer func() error

func combineCloseError(err error, cl closer) error {
	if err == nil {
		return cl()
	}

	clerr := cl()
	if clerr == nil {
		return err
	}

	return fmt.Errorf("error=%q closerError=%q", err.Error(), clerr.Error())
}

func makeNodeInfo(cfg *NodeConfig, nodeKey types.NodeKey) (types.NodeInfo, error) {
	nodeInfo := types.NodeInfo{
		ProtocolVersion: types.ProtocolVersion{
			P2P: version.P2PProtocol, // global
		},
		NodeID:  nodeKey.ID,
		Version: version.TMVersionDefault,
		Channels: []byte{
			pex.PexChannel,
		},
		Moniker: cfg.Moniker,
		Other: types.NodeInfoOther{
			TxIndex: "off",
		},
	}

	nodeInfo.ListenAddr = cfg.P2P.ExternalAddress
	if nodeInfo.ListenAddr == "" {
		nodeInfo.ListenAddr = cfg.P2P.ListenAddress
	}

	return nodeInfo, nodeInfo.Validate()
}

func createPeerManager(cfg *NodeConfig, dbProvider DBProvider, nodeID types.NodeID, metrics *p2p.Metrics) (*p2p.PeerManager, closer, error) {
	selfAddr, err := p2p.ParseNodeAddress(nodeID.AddressString(cfg.P2P.ExternalAddress))
	if err != nil {
		return nil, func() error { return nil }, fmt.Errorf("couldn't parse ExternalAddress %q: %w", cfg.P2P.ExternalAddress, err)
	}

	privatePeerIDs := make(map[types.NodeID]struct{})
	for _, id := range tmstrings.SplitAndTrimEmpty(cfg.P2P.PrivatePeerIDs, ",", " ") {
		privatePeerIDs[types.NodeID(id)] = struct{}{}
	}

	var maxConns uint16

	switch {
	case cfg.P2P.MaxConnections > 0:
		maxConns = cfg.P2P.MaxConnections
	default:
		maxConns = 64
	}

	var maxOutgoingConns uint16
	switch {
	case cfg.P2P.MaxOutgoingConnections > 0:
		maxOutgoingConns = cfg.P2P.MaxOutgoingConnections
	default:
		maxOutgoingConns = maxConns / 2
	}

	maxUpgradeConns := uint16(4)

	options := p2p.PeerManagerOptions{
		SelfAddress:              selfAddr,
		MaxConnected:             maxConns,
		MaxOutgoingConnections:   maxOutgoingConns,
		MaxConnectedUpgrade:      maxUpgradeConns,
		DisconnectCooldownPeriod: 2 * time.Second,
		MaxPeers:                 maxUpgradeConns + 4*maxConns,
		MinRetryTime:             250 * time.Millisecond,
		MaxRetryTime:             30 * time.Minute,
		MaxRetryTimePersistent:   5 * time.Minute,
		RetryTimeJitter:          5 * time.Second,
		PrivatePeers:             privatePeerIDs,
		Metrics:                  metrics,
	}

	peers := []p2p.NodeAddress{}
	for _, p := range tmstrings.SplitAndTrimEmpty(cfg.P2P.PersistentPeers, ",", " ") {
		address, err := p2p.ParseNodeAddress(p)
		if err != nil {
			return nil, func() error { return nil }, fmt.Errorf("invalid peer address %q: %w", p, err)
		}

		peers = append(peers, address)
		options.PersistentPeers = append(options.PersistentPeers, address.NodeID)
	}

	for _, p := range tmstrings.SplitAndTrimEmpty(cfg.P2P.BootstrapPeers, ",", " ") {
		address, err := p2p.ParseNodeAddress(p)
		if err != nil {
			return nil, func() error { return nil }, fmt.Errorf("invalid peer address %q: %w", p, err)
		}
		peers = append(peers, address)
	}

	peerDB, err := dbProvider(&DBContext{ID: "peerstore", Config: cfg})
	if err != nil {
		return nil, func() error { return nil }, fmt.Errorf("unable to initialize peer store: %w", err)
	}

	peerManager, err := p2p.NewPeerManager(nodeID, peerDB, options)
	if err != nil {
		return nil, peerDB.Close, fmt.Errorf("failed to create peer manager: %w", err)
	}

	for _, peer := range peers {
		if _, err = peerManager.Add(peer); err != nil {
			return nil, peerDB.Close, fmt.Errorf("failed to add peer %q: %w", peer, err)
		}
	}

	return peerManager, peerDB.Close, nil
}

func createRouter(logger tmlog.Logger, p2pMetrics *p2p.Metrics,
	nodeInfoProducer func() *types.NodeInfo, nodeKey types.NodeKey,
	peerManager *p2p.PeerManager, cfg *NodeConfig) (*p2p.Router, error) {

	p2pLogger := logger.With("module", "p2p")

	transportConf := conn.DefaultMConnConfig()
	transportConf.FlushThrottle = cfg.P2P.FlushThrottleTimeout
	transportConf.SendRate = cfg.P2P.SendRate
	transportConf.RecvRate = cfg.P2P.RecvRate
	transportConf.MaxPacketMsgPayloadSize = cfg.P2P.MaxPacketMsgPayloadSize
	transport := p2p.NewMConnTransport(
		p2pLogger, transportConf, []*p2p.ChannelDescriptor{},
		p2p.MConnTransportOptions{
			MaxAcceptedConnections: uint32(cfg.P2P.MaxConnections),
		},
	)

	ep, err := p2p.NewEndpoint(nodeKey.ID.AddressString(cfg.P2P.ListenAddress))
	if err != nil {
		return nil, err
	}

	return p2p.NewRouter(
		p2pLogger,
		p2pMetrics,
		nodeKey.PrivKey,
		peerManager,
		nodeInfoProducer,
		transport,
		ep,
		getRouterConfig(cfg),
	)
}

func getRouterConfig(conf *NodeConfig) p2p.RouterOptions {
	opts := p2p.RouterOptions{
		QueueType:        conf.P2P.QueueType,
		HandshakeTimeout: conf.P2P.HandshakeTimeout,
		DialTimeout:      conf.P2P.DialTimeout,
	}
	return opts
}
