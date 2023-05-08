package gfspapp

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

const (
	BaseCodeSpace = "gfsp-base-app"
)

type GfSpBaseApp struct {
	appID          string
	grpcAddress    string
	httpAddress    string
	domain         string // external domain name is used for virtual-hosted style url
	operateAddress string

	appCtx context.Context
	server *grpc.Server
	client *gfspclient.GfSpClient

	gfSpDB     spdb.SPDB
	pieceStore piecestore.PieceStore
	pieceKey   piecestore.PieceOp
	rcmgr      corercmgr.ResourceManager
	chain      consensus.Consensus

	approver   module.Approver
	downloader module.Downloader
	manager    module.Manager
	p2p        module.P2P
	receiver   module.Receiver
	taskNode   module.TaskExecutor
	uploader   module.Uploader
	signer     module.Signer
	authorizer module.Authorizer

	modulars map[string]module.Modular
	mux      sync.RWMutex

	uploadSpeed    int64
	downloadSpeed  int64
	replicateSpeed int64
	receiveSpeed   int64

	sealObjectTimeout int64
	gcObjectTimeout   int64
	gcZombieTimeout   int64
	gcMetaTimeout     int64

	sealObjectRetry     int64
	replicateRetry      int64
	receiveConfirmRetry int64
	gcObjectRetry       int64
	gcZombieRetry       int64
	gcMetaRetry         int64

	pprofEnable    bool
	metricsEnable  bool
	pprofAddress   string
	metricsAddress string
}

func (g *GfSpBaseApp) AppID() string {
	return g.appID
}

func (g *GfSpBaseApp) GfSpClient() *gfspclient.GfSpClient {
	return g.client
}

func (g *GfSpBaseApp) PieceStore() piecestore.PieceStore {
	return g.pieceStore
}

func (g *GfSpBaseApp) PieceOp() piecestore.PieceOp {
	return g.pieceKey
}

func (g *GfSpBaseApp) Consensus() consensus.Consensus {
	return g.chain
}

func (g *GfSpBaseApp) OperateAddress() string {
	return g.operateAddress
}

func (g *GfSpBaseApp) GfSpDB() spdb.SPDB {
	return g.gfSpDB
}

func (g *GfSpBaseApp) RegisterModular(m module.Modular) error {
	g.mux.Lock()
	defer g.mux.Unlock()
	if _, ok := g.modulars[m.Name()]; ok {
		return errors.New("module name repeated")
	}
	g.modulars[m.Name()] = m
	return nil
}

func (g *GfSpBaseApp) ServerForRegister() *grpc.Server {
	return g.server
}

func (g *GfSpBaseApp) ResourceManager() corercmgr.ResourceManager {
	return g.rcmgr
}
