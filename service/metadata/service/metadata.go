package service

import (
	"context"
	"errors"
	"net"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/store"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Metadata struct {
	config  *metadata.MetadataConfig
	ctx     context.Context
	name    string
	running atomic.Bool

	store      store.IStore
	grpcServer *grpc.Server
}

// Name implement the lifecycle interface
func (metadata *Metadata) Name() string {
	return metadata.name
}

// Start implement the lifecycle interface
// to delete api/v1
func (metadata *Metadata) Start(ctx context.Context) error {
	metadata.ctx = ctx
	if metadata.running.Swap(true) {
		return errors.New("metadata has started")
	}
	errCh := make(chan error)
	go metadata.serve(errCh)
	err := <-errCh
	log.Debug("metadata service succeed to start")
	return err
}

// Serve starts grpc service.
func (metadata *Metadata) serve(errCh chan error) {
	lis, err := net.Listen("tcp", metadata.config.Address)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}

	grpcServer := grpc.NewServer()
	stypes.RegisterMetadataServiceServer(grpcServer, metadata)
	metadata.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

// Stop implement the lifecycle interface
func (metadata *Metadata) Stop(ctx context.Context) error {
	if !metadata.running.Swap(false) {
		return errors.New("uploader has stopped")
	}
	metadata.grpcServer.GracefulStop()
	return nil
}

func NewMetadataService(cfg *metadata.MetadataConfig) (metadata *Metadata, err error) {
	metadataStore, _ := store.NewStore(cfg.SpDBConfig)
	metadata = &Metadata{
		config: cfg,
		name:   model.MetadataService,
		store:  metadataStore,
	}
	return
}
