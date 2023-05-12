package p2p

import (
	"context"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
)

const (
	PingProtocol = "/ping/0.0.1"
	PongProtocol = "/pong/0.0.1"
)

// OnPing handles remote node ping request
func (n *Node) onPing(s network.Stream) {
	var (
		peerID = s.Conn().RemotePeer()
		err    error
	)
	defer func() {
		if err != nil {
			n.peers.DeletePeer(peerID)
			log.Warnw("failed to response ping", "peer_id", peerID, "error", err)
			return
		}
		// log.Debugw("succeed to response ping", "peer_id", peerID)
	}()
	ping := &types.Ping{}
	buf, err := io.ReadAll(s)
	if err != nil {
		log.Errorw("failed to read ping msg from stream", "error", err)
		s.Reset()
		return
	}
	s.Close()
	err = proto.Unmarshal(buf, ping)
	if err != nil {
		log.Errorw("failed to unmarshal ping msg", "error", err)
		return
	}
	// log.Debugf("%s received ping request from %s. Message: %s", s.Conn().LocalPeer(), s.Conn().RemotePeer(), ping.String())

	err = types.VerifySignature(ping.GetSpOperatorAddress(), ping.GetSignBytes(), ping.GetSignature())
	if err != nil {
		log.Warnw("failed to verify ping msg signature", "local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	n.node.Peerstore().AddAddr(s.Conn().RemotePeer(), s.Conn().RemoteMultiaddr(), peerstore.PermanentAddrTTL)
	n.peers.AddPeer(peerID, ping.SpOperatorAddress, s.Conn().RemoteMultiaddr())

	pong := &types.Pong{}
	for _, pID := range n.node.Peerstore().PeersWithAddrs() {
		nodeInfo := &types.Node{
			NodeId: pID.String(),
		}
		addrs := n.node.Peerstore().Addrs(pID)
		for _, addr := range addrs {
			nodeInfo.MultiAddr = append(nodeInfo.MultiAddr, addr.String())
		}
		pong.Nodes = append(pong.Nodes, nodeInfo)
		// log.Debugw("send node to remote", "node_id", pID.String(), "remote_node", s.Conn().RemotePeer())
	}
	pong.SpOperatorAddress = n.SpOperatorAddress
	pong, err = n.signer.SignPongMsg(context.Background(), pong)
	if err != nil {
		log.Errorw("failed to sign pong msg", "local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	err = n.sendToPeer(s.Conn().RemotePeer(), PongProtocol, pong)
}

// onPong handles remote node pong response
func (n *Node) onPong(s network.Stream) {
	var (
		peerID = s.Conn().RemotePeer()
		err    error
	)
	defer func() {
		if err != nil {
			n.peers.DeletePeer(peerID)
			log.Warnw("failed to receive pong", "peer_id", peerID, "error", err)
			return
		}
		// log.Debugw("succeed to receive pong", "peer_id", peerID)
	}()
	pong := &types.Pong{}
	buf, err := io.ReadAll(s)
	if err != nil {
		log.Errorw("failed to read pong msg from stream", "error", err)
		s.Reset()
		return
	}
	s.Close()
	err = proto.Unmarshal(buf, pong)
	if err != nil {
		log.Errorw("failed to unmarshal ping msg", "error", err)
		return
	}
	// log.Debugf("%s received pong request from %s.", s.Conn().LocalPeer(), s.Conn().RemotePeer())

	err = types.VerifySignature(pong.GetSpOperatorAddress(), pong.GetSignBytes(), pong.GetSignature())
	if err != nil {
		log.Warnw("failed to verify pong msg signature", "local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	n.peers.AddPeer(peerID, pong.SpOperatorAddress, s.Conn().RemoteMultiaddr())

	for _, node := range pong.Nodes {
		pID, err := peer.Decode(node.NodeId)
		if err != nil {
			continue
		}
		var addrs []ma.Multiaddr
		for _, maddr := range node.MultiAddr {
			addr, err := ma.NewMultiaddr(maddr)
			if err != nil {
				continue
			}
			addrs = append(addrs, addr)
		}
		n.node.Peerstore().AddAddrs(pID, addrs, peerstore.PermanentAddrTTL)
		// log.Debugw("receive node from remote and permanent", "remote_node", s.Conn().RemotePeer(), "node_id", pID)
	}
}
