package syncer

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
)

var _ lifecycle.Service = &Syncer{}

// Syncer implements the gRPC of SyncerService,
// responsible for receive replicate object payload data
type Syncer struct {
	config     *SyncerConfig
	cache      *lru.Cache
	signer     *signerclient.SignerClient
	pieceStore *psclient.StoreClient
	spDB       sqldb.SPDB
	grpcServer *grpc.Server
}

// NewSyncerService return a Syncer instance and init the resource
func NewSyncerService(config *SyncerConfig) (*Syncer, error) {
	cache, _ := lru.New(model.LruCacheLimit)
	pieceStore, err := psclient.NewStoreClient(config.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	signer, err := signerclient.NewSignerClient(config.SignerGRPCAddress)
	if err != nil {
		return nil, err
	}
	spDB, err := sqldb.NewSQLStore(config.SPDBConfig)
	if err != nil {
		return nil, err
	}
	s := &Syncer{
		config:     config,
		cache:      cache,
		pieceStore: pieceStore,
		spDB:       spDB,
		signer:     signer,
	}
	return s, nil
}

// Name return the syncer service name, for the lifecycle management
func (syncer *Syncer) Name() string {
	return model.SyncerService
}

// Start the syncer background goroutine
func (syncer *Syncer) Start(ctx context.Context) error {
	errCh := make(chan error)
	go syncer.serve(errCh)
	err := <-errCh
	return err
}

// Stop the syncer gRPC service and recycle the resources
func (syncer *Syncer) Stop(ctx context.Context) error {
	syncer.grpcServer.GracefulStop()
	return nil
}

func (syncer *Syncer) serve(errCh chan error) {
	lis, err := net.Listen("tcp", syncer.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}

	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(model.MaxCallMsgSize))
	types.RegisterSyncerServiceServer(grpcServer, syncer)
	syncer.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}
