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
	coreprober "github.com/bnb-chain/greenfield-storage-provider/core/prober"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

const (
	BaseCodeSpace = "gfsp-base-app"
)

type GfSpBaseApp struct {
	appID           string
	grpcAddress     string
	operatorAddress string
	chainID         string

	server *grpc.Server
	client gfspclient.GfSpClientAPI

	gfSpDB       spdb.SPDB
	gfBsDB       bsdb.BSDB
	gfBsDBMaster bsdb.BSDB

	pieceStore piecestore.PieceStore
	pieceOp    piecestore.PieceOp
	rcmgr      corercmgr.ResourceManager
	chain      consensus.Consensus
	httpProbe  coreprober.Prober

	approver      module.Approver
	authenticator module.Authenticator
	downloader    module.Downloader
	executor      module.TaskExecutor
	gater         module.Modular
	manager       module.Manager
	p2p           module.P2P
	receiver      module.Receiver
	signer        module.Signer
	uploader      module.Uploader
	metrics       module.Modular
	pprof         module.Modular
	probeSvr      module.Modular

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
	migrateGVGTimeout int64

	sealObjectRetry     int64
	replicateRetry      int64
	receiveConfirmRetry int64
	gcObjectRetry       int64
	gcZombieRetry       int64
	gcMetaRetry         int64
	recoveryRetry       int64
	migrateGVGRetry     int64
}

// AppID returns the GfSpBaseApp ID, the default value is prefix(gfsp) add
// started modules' name.
func (g *GfSpBaseApp) AppID() string {
	return g.appID
}

// SetAppID sets appID
func (g *GfSpBaseApp) SetAppID(appID string) {
	g.appID = appID
}

// GfSpClient returns the sp client that includes inner grpc and outer http protocol.
func (g *GfSpBaseApp) GfSpClient() gfspclient.GfSpClientAPI {
	return g.client
}

// SetGfSpClient sets gfsp client
func (g *GfSpBaseApp) SetGfSpClient(clientAPI gfspclient.GfSpClientAPI) {
	g.client = clientAPI
}

// PieceStore returns the piece store client.
func (g *GfSpBaseApp) PieceStore() piecestore.PieceStore {
	return g.pieceStore
}

// SetPieceStore sets piece store
func (g *GfSpBaseApp) SetPieceStore(ps piecestore.PieceStore) {
	g.pieceStore = ps
}

// PieceOp returns piece helper struct instance.
func (g *GfSpBaseApp) PieceOp() piecestore.PieceOp {
	return g.pieceOp
}

// SetPieceOp sets piece op
func (g *GfSpBaseApp) SetPieceOp(pieceOp piecestore.PieceOp) {
	g.pieceOp = pieceOp
}

// Consensus returns greenfield consensus query client.
func (g *GfSpBaseApp) Consensus() consensus.Consensus {
	return g.chain
}

// SetConsensus sets greenfield consensus query client.
func (g *GfSpBaseApp) SetConsensus(chain consensus.Consensus) {
	g.chain = chain
}

// OperatorAddress returns the sp operator address.
func (g *GfSpBaseApp) OperatorAddress() string {
	return g.operatorAddress
}

// SetOperatorAddress sets operator address
func (g *GfSpBaseApp) SetOperatorAddress(operatorAddress string) {
	g.operatorAddress = operatorAddress
}

// ChainID returns the chainID used by this sp instance
func (g *GfSpBaseApp) ChainID() string {
	return g.chainID
}

// SetChainID sets chainID
func (g *GfSpBaseApp) SetChainID(chainID string) {
	g.chainID = chainID
}

// GfSpDB returns the sp db client.
func (g *GfSpBaseApp) GfSpDB() spdb.SPDB {
	return g.gfSpDB
}

// SetGfSpDB sets spdb
func (g *GfSpBaseApp) SetGfSpDB(db spdb.SPDB) {
	g.gfSpDB = db
}

// GfBsDB returns the block syncer db client.
func (g *GfSpBaseApp) GfBsDB() bsdb.BSDB {
	return g.gfBsDB
}

// GfBsDBMaster returns the master block syncer db client.
func (g *GfSpBaseApp) GfBsDBMaster() bsdb.BSDB {
	return g.gfBsDBMaster
}

// SetGfBsDB sets the block syncer db client.
func (g *GfSpBaseApp) SetGfBsDB(setDB bsdb.BSDB) {
	g.gfBsDB = setDB
}

// ServerForRegister returns the Grpc server for module register own service.
func (g *GfSpBaseApp) ServerForRegister() *grpc.Server {
	return g.server
}

// ResourceManager returns the resource manager for module to open own resource span.
func (g *GfSpBaseApp) ResourceManager() corercmgr.ResourceManager {
	return g.rcmgr
}

// SetResourceManager sets the resource manager for module to open own resource span.
func (g *GfSpBaseApp) SetResourceManager(rcmgr corercmgr.ResourceManager) {
	g.rcmgr = rcmgr
}

// GetProbe returns the http probe.
func (g *GfSpBaseApp) GetProbe() coreprober.Prober {
	return g.httpProbe
}

// SetProbe sets the http probe.
func (g *GfSpBaseApp) SetProbe(prober coreprober.Prober) {
	g.httpProbe = prober
}

// Start the GfSpBaseApp and blocks the progress until signal.
func (g *GfSpBaseApp) Start(ctx context.Context) error {
	g.httpProbe.Healthy()
	err := g.StartRPCServer(ctx)
	if err != nil {
		log.Errorw("failed to start rpc server", "error", err)
		g.httpProbe.Unhealthy(err)
		return err
	}
	g.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).StartServices(ctx).Wait(ctx)
	_ = g.close(ctx)
	return nil
}

// close recycles the GfSpBaseApp resource on the stop time.
func (g *GfSpBaseApp) close(ctx context.Context) error {
	_ = g.StopRPCServer(ctx)
	_ = g.GfSpClient().Close()
	_ = g.rcmgr.Close()
	_ = g.chain.Close()
	return nil
}

// EnableMetrics returns an indicator whether enable the metrics service.
func (g *GfSpBaseApp) EnableMetrics() bool {
	return g.metrics != nil
}
