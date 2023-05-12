package p2p

import (
	"context"
	"strings"
	"time"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	ds "github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/security/noise"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
)

// Node defines the p2p protocol node, encapsulates the go-lib.p2p
// p2p protocol message interactive security based on storage provider
// operator private-public key pair signature, use leveldb as underlying DB.
type Node struct {
	config            *NodeConfig
	SpOperatorAddress string
	node              host.Host
	signer            *signerclient.SignerClient
	peers             *PeerProvider
	persistentDB      ds.Batching
	approval          *ApprovalProtocol
	stopCh            chan struct{}
}

// NewNode return an instance of Node.
func NewNode(config *NodeConfig, SPAddr string, signer *signerclient.SignerClient) (*Node, error) {
	if config.PingPeriod < PingPeriodMin {
		config.PingPeriod = PingPeriodMin
	}
	privKey, localAddr, bootstrapIDs, bootstrapAddrs, err := config.ParseConfig()
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
		libp2p.NATPortMap(),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		libp2p.WithDialTimeout(time.Duration(DailTimeout)*time.Second),
	)
	if err != nil {
		return nil, err
	}
	for i, addr := range bootstrapAddrs {
		host.Peerstore().AddAddr(bootstrapIDs[i], addr, peerstore.PermanentAddrTTL)
	}
	n := &Node{
		config:            config,
		SpOperatorAddress: SPAddr,
		node:              host,
		signer:            signer,
		peers:             NewPeerProvider(store),
		persistentDB:      ds,
		stopCh:            make(chan struct{}),
	}
	n.initProtocol()
	log.Infow("succeed to init p2p node", "node_id", n.node.ID())
	return n, nil
}

func (n *Node) initProtocol() {
	// inline protocol
	n.node.SetStreamHandler(PingProtocol, n.onPing)
	n.node.SetStreamHandler(PongProtocol, n.onPong)
	// approval protocol
	n.approval = NewApprovalProtocol(n)
}

// Name return the p2p protocol node name
func (n *Node) Name() string {
	return P2PNode
}

// Start runs background task that trigger broadcast ping request
func (n *Node) Start(ctx context.Context) error {
	go n.eventLoop()
	return nil
}

// Stop recycle the resources and termination background goroutine
func (n *Node) Stop(ctx context.Context) error {
	close(n.stopCh)
	n.persistentDB.Close()
	return nil
}

// PeersProvider returns the p2p peers provider
func (n *Node) PeersProvider() *PeerProvider {
	return n.peers
}

// GetApproval broadcast get approval request and blocking goroutine until timeout or
// collect expect accept approval response number
func (n *Node) GetApproval(object *storagetypes.ObjectInfo, expectedAccept int, timeout int64) (
	accept map[string]*types.GetApprovalResponse, refuse map[string]*types.GetApprovalResponse, err error) {
	approvalCh, err := n.approval.hangApprovalRequest(object.Id.Uint64())
	if err != nil {
		return
	}
	defer n.approval.cancelApprovalRequest(object.Id.Uint64())
	accept = make(map[string]*types.GetApprovalResponse)
	refuse = make(map[string]*types.GetApprovalResponse)
	getApprovalReq := &types.GetApprovalRequest{
		ObjectInfo:        object,
		SpOperatorAddress: n.SpOperatorAddress,
	}
	getApprovalReq, err = n.signer.SignReplicateApprovalReqMsg(context.Background(), getApprovalReq)
	if err != nil {
		log.Errorw("failed to sign the get approval request", "object_id", object.Id.Uint64())
		return
	}
	n.broadcast(GetApprovalRequest, getApprovalReq)
	ticker := time.NewTicker(time.Duration(timeout) * time.Second)
	for {
		select {
		case approval := <-approvalCh:
			if approval.GetExpiredTime() <= time.Now().Unix() {
				log.Warnw("discard expired approval", "sp", approval.GetSpOperatorAddress(),
					"object_id", approval.GetObjectInfo().Id.Uint64(), "expire_time", approval.GetExpiredTime())
				continue
			}
			if len(approval.GetRefusedReason()) != 0 {
				if _, ok := refuse[approval.GetSpOperatorAddress()]; !ok {
					refuse[approval.GetSpOperatorAddress()] = approval
					delete(accept, approval.GetSpOperatorAddress())
				}
			} else {
				delete(refuse, approval.GetSpOperatorAddress())
				if _, ok := accept[approval.GetSpOperatorAddress()]; !ok {
					accept[approval.GetSpOperatorAddress()] = approval
				}
				if len(accept) >= expectedAccept {
					return
				}
			}
		case <-ticker.C:
			return
		}
	}
}

// eventLoop run the background task
func (n *Node) eventLoop() {
	ticker := time.NewTicker(time.Duration(n.config.PingPeriod) * time.Second)
	for {
		select {
		case <-n.stopCh:
			return
		case <-ticker.C:
			// TODO:: send to signer and back fill the signature field
			ping := &types.Ping{
				SpOperatorAddress: n.SpOperatorAddress,
			}
			ping, err := n.signer.SignPingMsg(context.Background(), ping)
			if err != nil {
				log.Warnw("failed to sign ping msg", "error", err)
				continue
			}
			n.broadcast(PingProtocol, ping)
		}
	}
}

// broadcast sends request to all p2p nodes
func (n *Node) broadcast(pc protocol.ID, data proto.Message) {
	for _, peerID := range n.node.Peerstore().PeersWithAddrs() {
		// log.Debugw("broadcast msg to peer", "peer_id", peerID, "protocol", pc)
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
		log.Debugw("failed to init stream", "peer_id", peerID, "protocol", pc, "error", err)
		n.peers.DeletePeer(peerID)
		return err
	}
	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Errorw("failed to send msg", "peer_id", peerID, "protocol", pc, "error", err)
		s.Close()
		n.peers.DeletePeer(peerID)
		return err
	}
	// log.Debugw("succeed to send msg", "peer_id", peerID, "protocol", pc, "msg", data.String())
	s.Close()
	return err
}
