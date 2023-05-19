package p2pnode

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	ds "github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// Node defines the p2p protocol node, encapsulates the go-lib.p2p
// p2p protocol message interactive security based on storage provider
// operator private-public key pair signature, use leveldb as underlying DB.
type Node struct {
	baseApp      *gfspapp.GfSpBaseApp
	node         host.Host
	peers        *PeerProvider
	persistentDB ds.Batching
	approval     *ApprovalProtocol
	stopCh       chan struct{}

	p2pPrivateKey                  crypto.PrivKey
	p2pProtocolAddress             ma.Multiaddr
	p2pPingPeriod                  int
	secondaryApprovalExpiredHeight uint64
	p2pBootstrap                   []string
}

// NewNode return an instance of Node
func NewNode(baseApp *gfspapp.GfSpBaseApp, privateKey string, address string,
	bootstrap []string, pingPeriod int, secondaryApprovalExpiredHeight uint64) (*Node, error) {
	if pingPeriod < PingPeriodMin {
		pingPeriod = PingPeriodMin
	}
	if secondaryApprovalExpiredHeight <= MinSecondaryApprovalExpiredHeight {
		secondaryApprovalExpiredHeight = MinSecondaryApprovalExpiredHeight
	}
	var privKey crypto.PrivKey
	if len(privateKey) > 0 {
		priKeyBytes, err := hex.DecodeString(privateKey)
		if err != nil {
			log.Errorw("failed to hex decode private key",
				"priv_key", privateKey, "error", err)
			return nil, err
		}
		privKey, err = crypto.UnmarshalSecp256k1PrivateKey(priKeyBytes)
		if err != nil {
			log.Errorw("failed to unmarshal secp256k1 private key",
				"priv_key", privateKey, "error", err)
			return nil, err
		}
	} else {
		newPrivKey, _, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		if err != nil {
			log.Errorw("failed to generate secp256k1 key pair", "error", err)
			return nil, err
		}
		privKey = newPrivKey
	}
	hostAddr, err := MakeMultiaddr(address)
	if err != nil {
		log.Errorw("failed to parser p2p protocol address", "error", err)
		return nil, err
	}
	bootstrapIDs, bootstrapAddrs, err := MakeBootstrapMultiaddr(bootstrap)
	if err != nil {
		log.Errorw("failed to parser bootstrap address", "error", err)
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
		libp2p.ListenAddrs(hostAddr),
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
		baseApp:                        baseApp,
		node:                           host,
		peers:                          NewPeerProvider(store),
		persistentDB:                   ds,
		p2pPrivateKey:                  privKey,
		p2pProtocolAddress:             hostAddr,
		p2pPingPeriod:                  pingPeriod,
		p2pBootstrap:                   bootstrap,
		secondaryApprovalExpiredHeight: secondaryApprovalExpiredHeight,
		stopCh:                         make(chan struct{}),
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

func (n *Node) Bootstrap() []string {
	return n.p2pBootstrap
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

// GetSecondaryReplicatePieceApproval broadcast get approval request and blocking
// goroutine until timeout or collect expect accept approval response number
func (n *Node) GetSecondaryReplicatePieceApproval(
	ctx context.Context,
	task coretask.ApprovalReplicatePieceTask,
	expectedAccept int, timeout int64) (
	accept []coretask.ApprovalReplicatePieceTask,
	err error) {
	approvalCh, err := n.approval.hangApprovalRequest(task.GetObjectInfo().Id.Uint64())
	if err != nil {
		log.CtxErrorw(ctx, "failed to hang replicate piece approval request")
		return
	}
	defer n.approval.cancelApprovalRequest(task.GetObjectInfo().Id.Uint64())
	task.SetAskSpOperatorAddress(n.baseApp.OperateAddress())
	signature, err := n.baseApp.GfSpClient().SignReplicatePieceApproval(ctx, task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign replicate piece approval request")
		return
	}
	task.SetAskSignature(signature)
	n.broadcast(ctx, GetApprovalRequest, task.(*gfsptask.GfSpReplicatePieceApprovalTask))
	approvalCtx, _ := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	for {
		select {
		case approval := <-approvalCh:
			current, innerErr := n.baseApp.Consensus().CurrentHeight(approvalCtx)
			if innerErr == nil {
				if approval.GetExpiredHeight() < current {
					log.Warnw("discard expired approval", "sp", approval.GetApprovedSpApprovalAddress(),
						"object_id", approval.GetObjectInfo().Id.Uint64(), "current_height", current,
						"expire_height", approval.GetExpiredHeight())
					continue
				}
			} else {
				log.Warnw("failed to get current height", "error", err)
				accept = append(accept, approval)
			}
			if len(accept) >= expectedAccept {
				log.CtxErrorw(ctx, "succeed to get sufficient approvals",
					"expect", expectedAccept, "accepted", len(accept))
				return
			}
		case <-approvalCtx.Done():
			log.CtxWarnw(ctx, "failed to get sufficient approvals",
				"expect", expectedAccept, "accepted", len(accept))
			return
		}
	}
}

// eventLoop run the background task
func (n *Node) eventLoop() {
	ticker := time.NewTicker(time.Duration(n.p2pPingPeriod) * time.Second)
	for {
		select {
		case <-n.stopCh:
			return
		case <-ticker.C:
			ping := &gfspp2p.GfSpPing{
				SpOperatorAddress: n.baseApp.OperateAddress(),
			}
			ctx := context.Background()
			sinagture, err := n.baseApp.GfSpClient().SignP2PPingMsg(ctx, ping)
			if err != nil {
				log.Warnw("failed to sign ping msg", "error", err)
				continue
			}
			ping.Signature = sinagture
			//log.CtxDebugw(ctx, "trigger broadcast ping")
			n.broadcast(ctx, PingProtocol, ping)
		}
	}
}

// broadcast sends request to all p2p nodes
func (n *Node) broadcast(
	ctx context.Context,
	pc protocol.ID,
	data proto.Message) {
	for _, peerID := range n.node.Peerstore().PeersWithAddrs() {
		if strings.Compare(n.node.ID().String(), peerID.String()) == 0 {
			continue
		}
		//addrs := n.node.Peerstore().Addrs(peerID)
		//for _, addr := range addrs {
		//	log.CtxErrorw(ctx, "broadcast", "protocol", pc, "peer_addr", addr.String())
		//}
		n.sendToPeer(ctx, peerID, pc, data)
	}
}

// sendToPeer sends request to all special p2p node
func (n *Node) sendToPeer(
	ctx context.Context,
	peerID peer.ID,
	pc protocol.ID,
	data proto.Message) error {
	host := n.node
	addrs := n.node.Peerstore().Addrs(peerID)
	s, err := host.NewStream(ctx, peerID, pc)
	// current p2p only support approval, add log for debug
	if strings.EqualFold(string(pc), GetApprovalRequest) ||
		strings.EqualFold(string(pc), GetApprovalResponse) {
		log.CtxDebugw(ctx, "send approval protocol", "protocol", pc,
			"peer", peerID.String(), "addr", addrs)
	}
	if err != nil {
		//log.CtxErrorw(ctx, "failed to init stream", "protocol", pc,
		//	"peer_id", peerID, "addr", addrs, "error", err)
		n.peers.DeletePeer(peerID)
		return err
	}
	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send msg", "protocol", pc,
			"peer_id", peerID, "addr", addrs, "error", err)
		s.Close()
		n.peers.DeletePeer(peerID)
		return err
	}
	s.Close()
	return err
}
