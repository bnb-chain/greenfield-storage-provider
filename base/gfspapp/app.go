package gfspapp

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
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
	operateAddress string

	appCtx context.Context
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

	metrics module.Modular
	pprof   module.Modular

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

func (g *GfSpBaseApp) SetApprover(approver module.Approver) error {
	if g.approver != nil {
		log.Panic("repeated set approver to base app")
	}
	g.approver = approver
	return nil
}

func (g *GfSpBaseApp) SetAuthorizer(authorizer module.Authorizer) error {
	if g.authorizer != nil {
		log.Panic("repeated set authorizer to base app")
	}
	g.authorizer = authorizer
	return nil
}

func (g *GfSpBaseApp) SetDownloader(downloader module.Downloader) error {
	if g.downloader != nil {
		log.Panic("repeated set downloader to base app")
	}
	g.downloader = downloader
	return nil
}

func (g *GfSpBaseApp) SetTaskExecutor(executor module.TaskExecutor) error {
	if g.executor != nil {
		log.Panic("repeated set executor to base app")
	}
	g.executor = executor
	return nil
}

func (g *GfSpBaseApp) SetGater(gater module.Modular) error {
	if g.gater != nil {
		log.Panic("repeated set gater to base app")
	}
	g.gater = gater
	return nil
}

func (g *GfSpBaseApp) SetManager(manager module.Manager) error {
	if g.manager != nil {
		log.Panic("repeated set manager to base app")
	}
	g.manager = manager
	return nil
}

func (g *GfSpBaseApp) SetP2P(p2p module.P2P) error {
	if g.p2p != nil {
		log.Panic("repeated set p2p to base app")
	}
	g.p2p = p2p
	return nil
}

func (g *GfSpBaseApp) SetReceiver(receiver module.Receiver) error {
	if g.receiver != nil {
		log.Panic("repeated set receiver to base app")
	}
	g.receiver = receiver
	return nil
}

func (g *GfSpBaseApp) SetSigner(signer module.Signer) error {
	if g.signer != nil {
		log.Panic("repeated set signer to base app")
	}
	g.signer = signer
	return nil
}

func (g *GfSpBaseApp) SetUploader(uploader module.Uploader) error {
	if g.uploader != nil {
		log.Panic("repeated set uploader to base app")
	}
	g.uploader = uploader
	return nil
}
