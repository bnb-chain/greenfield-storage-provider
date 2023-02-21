package reactor

import (
	"strings"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/types"
	dao "github.com/bnb-chain/greenfield-storage-provider/store/spdb"
)

type ProviderQuerier interface {
	Check(nodeId types.NodeID) bool
}

type providerQuerier struct {
	persisted map[types.NodeID]struct{}
	db        dao.P2PNodeDB

	mtx     sync.Mutex
	dbCache map[types.NodeID]struct{}
}

func NewProviderQuerier(persistedPeers string, providerDao dao.P2PNodeDB) *providerQuerier {
	persistedPeers = strings.TrimSpace(persistedPeers)
	m := make(map[types.NodeID]struct{})
	var peers []string
	if len(persistedPeers) > 0 {
		peers = strings.Split(persistedPeers, ";")
	}
	for _, peer := range peers {
		splits := strings.Split(peer, "@")
		if len(splits) > 0 {
			m[types.NodeID(splits[1])] = struct{}{}
		}
	}

	//TODO: there could be a routine to refresh
	dbAll, _ := providerDao.FetchAll()
	dbCache := make(map[types.NodeID]struct{})
	for _, item := range dbAll {
		dbCache[types.NodeID(item.NodeId)] = struct{}{}
	}

	return &providerQuerier{
		persisted: m,
		db:        providerDao,
		dbCache:   dbCache,
	}
}

func (q *providerQuerier) Check(nodeId types.NodeID) bool {
	if _, ok := q.persisted[nodeId]; ok {
		return true
	}
	if _, ok := q.dbCache[nodeId]; ok {
		return true
	}
	if _, err := q.db.Get(string(nodeId)); err == nil {
		return true
	}
	return false
}
