package uploader

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb/leveldb"
	"google.golang.org/grpc"

	"github.com/bnb-chain/inscription-storage-provider/service/client"

	"github.com/bnb-chain/inscription-storage-provider/mock"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	ServiceNameUploader string = "UploaderService"
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
		name:   ServiceNameUploader,
	}
	stoneHub, err := client.NewStoneHubClient(cfg.StoneHubServiceAddress)
	if err != nil {
		return nil, err
	}
	store, err := client.NewStoreClient(&cfg.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	u.stoneHub = stoneHub
	u.store = store
	u.eventWaiter = mock.NewInscriptionChainMock()
	u.signer = mock.NewSignerServerMock(u.eventWaiter)
	u.eventWaiter.Start()
	if err := u.initDB(cfg.MetaDBConfig); err != nil {
		return nil, err
	}
	return u, err
}

func (u *Uploader) initDB(config *leveldb.MetaLevelDBConfig) (err error) {
	u.metaDB, err = leveldb.NewMetaDB(config)
	if err != nil {
		log.Errorw("failed to init metaDB", "err", err)
		return err
	}
	return nil
}

// Name implement the lifecycle interface
func (u *Uploader) Name() string {
	return u.name
}

// Start implement the lifecycle interface
func (u *Uploader) Start(ctx context.Context) error {
	if u.running.Swap(true) {
		return errors.New("uploader has started")
	}
	errCh := make(chan error)
	u.eventWaiter.Start()
	go u.serve(errCh)
	err := <-errCh
	return err
}

// Serve starts grpc service.
func (u *Uploader) serve(errCh chan error) {
	lis, err := net.Listen("tcp", u.config.Address)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}
	server := grpc.NewServer()
	service.RegisterUploaderServiceServer(server, &uploaderImpl{uploader: u})
	u.grpcServer = server
	if err := server.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

// Stop implement the lifecycle interface
func (u *Uploader) Stop(ctx context.Context) error {
	if !u.running.Swap(false) {
		return errors.New("uploader has stopped")
	}
	u.grpcServer.GracefulStop()
	u.eventWaiter.Stop()
	var errs []error
	if err := u.stoneHub.Close(); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
