package uploader

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// Uploader respond to putObject impl
type Uploader struct {
	config  *UploaderConfig
	name    string
	running atomic.Bool

	grpcServer *grpc.Server
	stoneHub   *client.StoneHubClient
	store      *client.StoreClient
	chain      *gnfd.Greenfield
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
	if u.stoneHub, err = client.NewStoneHubClient(cfg.StoneHubServiceAddress); err != nil {
		log.Warnw("failed to stone hub client", "err", err)
		return nil, err
	}
	if u.store, err = client.NewStoreClient(cfg.PieceStoreConfig); err != nil {
		log.Warnw("failed to piece store client", "err", err)
		return nil, err
	}
	if u.chain, err = gnfd.NewGreenfield(cfg.ChainConfig); err != nil {
		log.Warnw("failed to create chain client", "err", err)
		return nil, err
	}
	return u, err
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
	var errs []error
	if err := uploader.stoneHub.Close(); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
