package p2p

import (
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util/maps"
)

const (
	// PeerFailureMax defines the threshold of setting node to fail
	PeerFailureMax = 1
	// PrunePeersNumberMax defines the threshold to trigger the prune p2p node
	PrunePeersNumberMax = 10
	// PeerSpUnspecified defines default sp operator address
	PeerSpUnspecified = "PEER_SP_UNSPECIFIED"
)

// Peer defines the peer info in memory
type Peer struct {
	peerID  peer.ID
	sp      string
	addr    ma.Multiaddr
	last    int64
	failCnt int
}

// ID return the peer id
func (p *Peer) ID() peer.ID {
	return p.peerID
}

// SP return the storage provider operator address of peer
func (p *Peer) SP() string {
	return p.sp
}

// IncrFail increase peer's fail counter and update last
func (p *Peer) IncrFail() int {
	p.failCnt++
	p.last = time.Now().Unix()
	return p.failCnt
}

// Fail return the indicator whether in fail state
func (p *Peer) Fail() bool {
	return p.failCnt >= PeerFailureMax
}

// Reset peer's fail counter and update last
func (p *Peer) Reset() {
	p.failCnt = 0
	p.last = time.Now().Unix()
}

// PeerProvider implements the pruning function of permanent nodes. Ping service only add permanent node,
// for zombie nodes, PeerProvider's pruning strategy takes into account the information of the storage
// provider dimension, and uses a very conservative pruning strategy. Nodes are only pruned if there are
// enough backups and multiple failed interactions, can try to keep each storage provider with enough nodes
// to try to connect, so that each sp has an equal opportunity to receive requests
type PeerProvider struct {
	peerStore peerstore.Peerstore
	peers     map[peer.ID]*Peer
	spPeers   map[string][]*Peer
	mux       sync.RWMutex
}

// NewPeerProvider return an instance of PeerProvider
func NewPeerProvider(store peerstore.Peerstore) *PeerProvider {
	return &PeerProvider{
		peerStore: store,
		peers:     make(map[peer.ID]*Peer),
		spPeers:   make(map[string][]*Peer),
	}
}

// UpdateSp change the storage provider
func (pr *PeerProvider) UpdateSp(SPs []string) {
	pr.mux.Lock()
	defer pr.mux.Unlock()
	sp2Peers := make(map[string][]*Peer)
	sp2Peers[PeerSpUnspecified] = pr.spPeers[PeerSpUnspecified]
	for _, sp := range SPs {
		peers, ok := pr.spPeers[sp]
		if ok {
			sp2Peers[sp] = peers
		} else {
			sp2Peers[sp] = make([]*Peer, 0)
		}
	}
	pr.spPeers = sp2Peers
}

// checkSP checks the sp is valid
func (pr *PeerProvider) checkSP(sp string) bool {
	pr.mux.RLock()
	defer pr.mux.RUnlock()
	_, ok := pr.spPeers[sp]
	return ok
}

// DeletePeer increase the peer's fail counter and trigger prunePeers
func (pr *PeerProvider) DeletePeer(peerID peer.ID) {
	pr.mux.Lock()
	defer pr.mux.Unlock()
	peer, ok := pr.peers[peerID]
	if !ok {
		return
	}
	peer.IncrFail()
	pr.prunePeers()
}

// AddPeer add a new peer or reset an old peer
func (pr *PeerProvider) AddPeer(peerID peer.ID, sp string, addr ma.Multiaddr) {
	pr.mux.Lock()
	defer pr.mux.Unlock()
	if _, ok := pr.spPeers[sp]; !ok {
		sp = PeerSpUnspecified
	}
	node, ok := pr.peers[peerID]
	if !ok {
		node = &Peer{
			peerID: peerID,
			sp:     sp,
			addr:   addr,
		}
		pr.peers[node.peerID] = node
	}
	if _, ok = pr.spPeers[sp]; !ok {
		pr.spPeers[sp] = make([]*Peer, 0)
	}
	find := false
	for _, spPeer := range pr.spPeers[sp] {
		if node.peerID.String() == spPeer.peerID.String() {
			find = true
			break
		}
	}
	if !find {
		pr.spPeers[sp] = append(pr.spPeers[sp], node)
	}
	node.Reset()
	pr.prunePeers()
}

// prunePeers deletes the peers that in fail state and there are enough backups
// notice: no lock for prune peers, only be called by DeletePeer and AddPeer
func (pr *PeerProvider) prunePeers() {
	sp2Peers := pr.spPeers
	sps := maps.SortKeys(sp2Peers)
	for _, sp := range sps {
		peers := pr.spPeers[sp]
		if len(peers) <= PrunePeersNumberMax {
			continue
		}
		log.Infow("trigger prune sp nodes", "sp", sp)
		sort.SliceStable(peers, func(i, j int) bool {
			if peers[i].failCnt != peers[j].failCnt {
				return peers[i].failCnt < peers[j].failCnt
			}
			return peers[i].last > peers[j].last
		})
		var discard []peer.ID
		for i := PrunePeersNumberMax; i <= len(peers); i++ {
			discard = append(discard, peers[i].ID())
		}
		log.Infow("prune sp nodes", "prune_num", len(discard))
		pr.deletePeers(discard)
		sp2Peers[sp] = peers[0:PrunePeersNumberMax]
	}
	pr.spPeers = sp2Peers
}

// deletePeers deletes the peer from store
// notice: no lock for delete peers, only be called by prunePeers
func (pr *PeerProvider) deletePeers(peers []peer.ID) {
	for _, peerID := range peers {
		pr.peerStore.ClearAddrs(peerID)
		pr.peerStore.RemovePeer(peerID)
		log.Infow("delete node", "node_id", peerID.String())
	}
}
