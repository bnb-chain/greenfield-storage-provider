package p2p

import (
	"context"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	ds "github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
)

// Node defines the p2p protocol node, encapsulates the go-lib.p2p
// p2p protocol message interactive security based on storage provider
// operator private-public key pair signature, use leveldb as underlying DB.
type Node struct {
	config       *NodeConfig
	node         host.Host
	peers        *PeerProvider
	persistentDB ds.Batching
	stopCh       chan struct{}
}

// NewNode return an instance of Node
func NewNode(config *NodeConfig) (*Node, error) {
	if config.PingPeriod < PingPeriodMin {
		config.PingPeriod = PingPeriodMin
	}
	privKey, localAddr, bootstrapIDs, bootstrapAddrs, err := config.ParseConfing()
	if err != nil {
		return nil, err
	}
	// init store for storing peers addr
	ds, err := leveldb.NewDatastore(DefaultDataPath, nil)
	if err != nil {
		return nil, err
	}
	store, err := pstoreds.NewPeerstore(context.Background(), ds, pstoreds.DefaultOpts())
	if err != nil {
		return nil, err
	}
	host, err := libp2p.New(
		libp2p.ListenAddrs(localAddr),
		libp2p.Identity(privKey),
		libp2p.Peerstore(store),
		// Customized ping service, it implemented dynamic update of permanent nodes. As usual, permanent nodes
		// should cover as many storage providers as possible, which is more decentralized, and also meets sp
		// requirements, eg: get approval request needs at least 6 or more replies from different storage providers
		// but p2p node are offline and replacement, which is an inevitable problem, If permanent nodes belonging
		// to the same sp all fail and are replaced, then the sp will be unable to communicate, this requires
		// dynamic updates permanent nodes
		libp2p.Ping(false),
		libp2p.WithDialTimeout(time.Duration(DailTimeOut)*time.Second),
	)
	if err != nil {
		return nil, err
	}
	for i, addr := range bootstrapAddrs {
		host.Peerstore().AddAddr(bootstrapIDs[i], addr, peerstore.PermanentAddrTTL)
	}
	n := &Node{
		config:       config,
		node:         host,
		peers:        NewPeerProvider(store),
		persistentDB: ds,
		stopCh:       make(chan struct{}),
	}
	n.node.SetStreamHandler(PingProtocol, n.onPing)
	n.node.SetStreamHandler(PongProtocol, n.onPong)
	log.Infow("success to init p2p node", "node_id", n.node.ID())
	return n, nil
}

// Name return the p2p protocol node name
func (n *Node) Name() string {
	return P2PNode
}

// Start runs background task that trigger broadcast ping request
func (n *Node) Start(ctx context.Context) error {
	go n.eventloop()
	return nil
}

// Stop recycle the resources and termination background goroutine
func (n *Node) Stop(ctx context.Context) error {
	close(n.stopCh)
	n.persistentDB.Close()
	return nil
}

// eventloop run the background task
func (n *Node) eventloop() {
	ticker := time.NewTicker(time.Duration(n.config.PingPeriod) * time.Second)
	for {
		select {
		case <-n.stopCh:
			return
		case <-ticker.C:
			// TODO:: send to signer and back fill the signature field
			ping := &types.Ping{
				SpOperatorAddress: n.config.SpOperatorAddress,
			}
			log.Debugw("trigger broadcast ping")
			n.broadcast(PingProtocol, ping)
		}
	}
}

// broadcast sends request to all p2p nodes
func (n *Node) broadcast(pc protocol.ID, data proto.Message) {
	for _, peerID := range n.node.Peerstore().PeersWithAddrs() {
		log.Debugw("broadcast ping to peer", "peer_id", peerID)
		if strings.Compare(n.node.ID().String(), peerID.String()) == 0 {
			continue
		}
		n.sendToPeer(peerID, pc, data)
	}
}

// sendToPeer sends request to all special p2p node
func (n *Node) sendToPeer(peerID peer.ID, pc protocol.ID, data proto.Message) error {
	host := n.node
	s, err := host.NewStream(context.Background(), peerID, pc)
	if err != nil {
		log.Warnw("failed to init stream", "error", err)
		n.peers.DeletePeer(peerID)
		return err
	}
	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Infow("failed to send msg", "peer_id", peerID, "protocol", pc, "error", err)
		s.Close()
		n.peers.DeletePeer(peerID)
		return err
	}
	log.Debugw("success to send msg", "peer_id", peerID, "protocol", pc, "msg", data.String())
	s.Close()
	return err
}
