package provider

import (
	"strings"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/types"
	dao "github.com/bnb-chain/greenfield-storage-provider/store/spdb"
)

type ProviderQuerier interface {
	Check(nodeId types.NodeID) bool
}

type providerQuerier struct {
	mtx     sync.Mutex
	db      dao.P2PNodeDB
	logger  log.Logger
	dbCache map[types.NodeID]struct{}
}

func NewProviderQuerier(persistedPeers string, logger log.Logger, providerDao dao.P2PNodeDB) *providerQuerier {
	persistedPeers = strings.TrimSpace(persistedPeers)
	m := make(map[types.NodeID]struct{})
	var peers []string
	if len(persistedPeers) > 0 {
		peers = strings.Split(persistedPeers, ",")
	}
	for _, peer := range peers {
		splits := strings.Split(peer, "@")
		if len(splits) > 0 {
			m[types.NodeID(splits[0])] = struct{}{}
			providerDao.Create(&dao.Provider{NodeId: splits[0]})
			logger.Info("add peer to db", "node_id", splits[0])
		}
	}

	//TODO: there could be a routine to refresh
	dbAll, _ := providerDao.FetchAll()
	dbCache := make(map[types.NodeID]struct{})
	for _, item := range dbAll {
		dbCache[types.NodeID(item.NodeId)] = struct{}{}
	}

	return &providerQuerier{
		db:      providerDao,
		logger:  logger,
		dbCache: dbCache,
	}
}

func (q *providerQuerier) Check(nodeId types.NodeID) bool {
	if _, ok := q.dbCache[nodeId]; ok {
		return true
	}
	q.dbCache[nodeId] = struct{}{}
	if _, err := q.db.Get(string(nodeId)); err == nil {
		return true
	}
	return false
}
