package service

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Metadata implements the gRPC of MetadataService,
// responsible for interact with SP for complex query service.
type Metadata struct {
	config     *metadata.MetadataConfig
	name       string
	store      bsdb.IStore
	grpcServer *grpc.Server
}

// NewMetadataService returns an instance of Metadata that
// supply query service for Inscription network
func NewMetadataService(cfg *metadata.MetadataConfig) (metadata *Metadata, err error) {
	metadataStore, _ := bsdb.NewStore(cfg.SpDBConfig)
	metadata = &Metadata{
		config: cfg,
		name:   model.MetadataService,
		store:  metadataStore,
	}
	return
}

// Name return the metadata service name, for the lifecycle management
func (metadata *Metadata) Name() string {
	return metadata.name
}

// Start the metadata gRPC service
func (metadata *Metadata) Start(ctx context.Context) error {
	errCh := make(chan error)
	go metadata.serve(errCh)
	err := <-errCh
	return err
}

// Stop the metadata gRPC service and recycle the resources
func (metadata *Metadata) Stop(ctx context.Context) error {
	metadata.grpcServer.GracefulStop()
	return nil
}

// Serve starts grpc service.
func (metadata *Metadata) serve(errCh chan error) {
	lis, err := net.Listen("tcp", metadata.config.GRPCAddress)
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
