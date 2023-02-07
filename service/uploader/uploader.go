package uploader

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-storage-provider/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/mock"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// Uploader respond to putObjectTx/putObject impl.
type Uploader struct {
	config  *UploaderConfig
	name    string
	running atomic.Bool

	grpcServer  *grpc.Server
	stoneHub    *client.StoneHubClient
	signer      *mock.SignerServerMock
	eventWaiter *mock.InscriptionChainMock
	store       *client.StoreClient
	metaDB      metadb.MetaDB // store auth info
}

// NewUploaderService return the uploader instance
func NewUploaderService(cfg *UploaderConfig) (*Uploader, error) {
	var (
		err error
		u   *Uploader
	)
	u = &Uploader{
		config: cfg,
		name:   model.UploaderService,
	}
	stoneHub, err := client.NewStoneHubClient(cfg.StoneHubServiceAddress)
	if err != nil {
		return nil, err
	}
	store, err := client.NewStoreClient(cfg.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	u.stoneHub = stoneHub
	u.store = store
	u.eventWaiter = mock.GetInscriptionChainMockSingleton()
	u.signer = mock.NewSignerServerMock(u.eventWaiter)
	u.eventWaiter.Start()
	if err := u.initDB(); err != nil {
		return nil, err
	}
	return u, err
}

// initDB init a meta-db instance
func (uploader *Uploader) initDB() error {
	var (
		metaDB metadb.MetaDB
		err    error
	)

	metaDB, err = store.NewMetaDB(uploader.config.MetaDBType,
		uploader.config.MetaLevelDBConfig, uploader.config.MetaSqlDBConfig)
	if err != nil {
		log.Errorw("failed to init metaDB", "err", err)
		return err
	}
	uploader.metaDB = metaDB
	return nil
}

// Name implement the lifecycle interface
func (uploader *Uploader) Name() string {
	return uploader.name
}

// Start implement the lifecycle interface
func (uploader *Uploader) Start(ctx context.Context) error {
	if uploader.running.Swap(true) {
		return errors.New("uploader has started")
	}
	errCh := make(chan error)
	uploader.eventWaiter.Start()
	go uploader.serve(errCh)
	err := <-errCh
	return err
}

// Serve starts grpc service.
func (uploader *Uploader) serve(errCh chan error) {
	lis, err := net.Listen("tcp", uploader.config.Address)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}

	grpcServer := grpc.NewServer()
	stypes.RegisterUploaderServiceServer(grpcServer, uploader)
	uploader.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

// Stop implement the lifecycle interface
func (uploader *Uploader) Stop(ctx context.Context) error {
	if !uploader.running.Swap(false) {
		return errors.New("uploader has stopped")
	}
	uploader.grpcServer.GracefulStop()
	uploader.eventWaiter.Stop()
	var errs []error
	if err := uploader.stoneHub.Close(); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
