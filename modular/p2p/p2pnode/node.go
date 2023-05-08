package p2pnode

import (
	"context"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	ds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/libp2p/go-libp2p/core/crypto"
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

	dnPath       string
	p2pBootstrap []string
}

// NewNode return an instance of Node
func NewNode() (*Node, error) {

	return nil, nil
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
			if innerErr != nil {
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
				log.CtxErrorw(ctx, "succeed to get sufficient approvals", "expect", expectedAccept)
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
			// TODO:: send to signer and back fill the signature field
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
	s, err := host.NewStream(ctx, peerID, pc)
	if err != nil {
		log.CtxErrorw(ctx, "failed to init stream", "peer_id", peerID, "protocol", pc, "error", err)
		n.peers.DeletePeer(peerID)
		return err
	}
	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send msg", "peer_id", peerID, "protocol", pc, "error", err)
		s.Close()
		n.peers.DeletePeer(peerID)
		return err
	}
	s.Close()
	return err
}
