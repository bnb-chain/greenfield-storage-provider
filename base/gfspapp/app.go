package gfspapp

import (
	"context"
	"syscall"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	corelifecycle "github.com/bnb-chain/greenfield-storage-provider/core/lifecycle"
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
	operateAddress string

	server *grpc.Server
	client *gfspclient.GfSpClient

	gfSpDB     spdb.SPDB
	pieceStore piecestore.PieceStore
	pieceOp    piecestore.PieceOp
	rcmgr      corercmgr.ResourceManager
	chain      consensus.Consensus

	approver   module.Approver
	authorizer module.Authorizer
	downloader module.Downloader
	executor   module.TaskExecutor
	gater      module.Modular
	manager    module.Manager
	p2p        module.P2P
	receiver   module.Receiver
	signer     module.Signer
	uploader   module.Uploader
	metrics    module.Modular
	pprof      module.Modular

	appCtx    context.Context
	appCancel context.CancelFunc
	services  []corelifecycle.Service

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
	return g.pieceOp
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

func (g *GfSpBaseApp) ServerForRegister() *grpc.Server {
	return g.server
}

func (g *GfSpBaseApp) ResourceManager() corercmgr.ResourceManager {
	return g.rcmgr
}

func (g *GfSpBaseApp) Start(ctx context.Context) error {
	err := g.StartRpcServer(ctx)
	if err != nil {
		return err
	}
	g.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).
		StartServices(ctx).
		Wait(ctx)
	g.close(ctx)
	return nil
}

func (g *GfSpBaseApp) close(ctx context.Context) error {
	g.StopRpcServer(ctx)
	g.GfSpClient().Close()
	g.rcmgr.Close()
	g.chain.Close()
	return nil
}

func (g *GfSpBaseApp) EnableMetrics() bool {
	return g.metrics != nil
}
