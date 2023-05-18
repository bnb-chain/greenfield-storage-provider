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
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
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

	gfSpDB       spdb.SPDB
	gfBsDB       bsdb.BSDB
	gfBsDBMaster bsdb.BSDB
	gfBsDBBackup bsdb.BSDB
	pieceStore   piecestore.PieceStore
	pieceOp      piecestore.PieceOp
	rcmgr        corercmgr.ResourceManager
	chain        consensus.Consensus

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

// AppID returns the GfSpBaseApp ID, the default value is prefix(gfsp) add
// started modules' name.
func (g *GfSpBaseApp) AppID() string {
	return g.appID
}

// GfSpClient returns the sp client that grpc and http protocol.
func (g *GfSpBaseApp) GfSpClient() *gfspclient.GfSpClient {
	return g.client
}

// PieceStore returns the piece store client.
func (g *GfSpBaseApp) PieceStore() piecestore.PieceStore {
	return g.pieceStore
}

// PieceOp returns piece helper struct instance.
func (g *GfSpBaseApp) PieceOp() piecestore.PieceOp {
	return g.pieceOp
}

// Consensus returns greenfield consensus query client.
func (g *GfSpBaseApp) Consensus() consensus.Consensus {
	return g.chain
}

// OperateAddress returns the sp operator address.
func (g *GfSpBaseApp) OperateAddress() string {
	return g.operateAddress
}

// GfSpDB returns the sp db client.
func (g *GfSpBaseApp) GfSpDB() spdb.SPDB {
	return g.gfSpDB
}

// GfBsDB returns the block syncer db client.
func (g *GfSpBaseApp) GfBsDB() bsdb.BSDB {
	return g.gfBsDB
}

// GfBsDBMaster returns the master block syncer db client.
func (g *GfSpBaseApp) GfBsDBMaster() bsdb.BSDB {
	return g.gfBsDBMaster
}

// GfBsDBBackup returns the backup block syncer db client.
func (g *GfSpBaseApp) GfBsDBBackup() bsdb.BSDB {
	return g.gfBsDBBackup
}

// SetGfBsDB set the block syncer db client.
func (g *GfSpBaseApp) SetGfBsDB(setDB bsdb.BSDB) bsdb.BSDB {
	g.gfBsDB = setDB
	return g.gfBsDB
}

// ServerForRegister returns the Grpc server for module register own service.
func (g *GfSpBaseApp) ServerForRegister() *grpc.Server {
	return g.server
}

// ResourceManager returns the resource manager for module to open own resource span.
func (g *GfSpBaseApp) ResourceManager() corercmgr.ResourceManager {
	return g.rcmgr
}

// Start the GfSpBaseApp and blocks the progress until signal.
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

// close recycles the GfSpBaseApp resource on the stop time.
func (g *GfSpBaseApp) close(ctx context.Context) error {
	g.StopRpcServer(ctx)
	g.GfSpClient().Close()
	g.rcmgr.Close()
	g.chain.Close()
	return nil
}

// EnableMetrics returns an indicator whether enable the metrics service.
func (g *GfSpBaseApp) EnableMetrics() bool {
	return g.metrics != nil
}
