package uploader

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/store"
	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	signercli "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	types "github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	pscli "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var _ lifecycle.Service = &Uploader{}

// Uploader implements the gRPC of UploaderService,
// responsible for upload object payload data.
type Uploader struct {
	config *UploaderConfig
	cache  *lru.Cache

	store      store.SPDB
	pieceStore *pscli.StoreClient

	signer *signercli.SignerClient
	//stoneNode *signer.SignerClient

	grpcServer *grpc.Server
}

// NewUploaderService returns an instance of Uploader that implementation of
// the lifecycle.Service and UploaderService interface
func NewUploaderService(config *UploaderConfig) (*Uploader, error) {
	cache, _ := lru.New(model.LruCacheLimit)
	pieceStore, err := pscli.NewStoreClient(config.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	signer, err := signercli.NewSignerClient(config.SignerAddress)
	if err != nil {
		return nil, err
	}
	//TODO:: new spdb
	uploader := &Uploader{
		config:     config,
		cache:      cache,
		pieceStore: pieceStore,
		signer:     signer,
	}
	return uploader, nil
}

// Name return the uploader service name, for the lifecycle management
func (uploader *Uploader) Name() string {
	return model.UploaderService
}

// Start the uploader gRPC service
func (uploader *Uploader) Start(ctx context.Context) error {
	errCh := make(chan error)
	go uploader.serve(errCh)
	err := <-errCh
	return err
}

// Stop the uploader gRPC service and recycle the resources
func (uploader *Uploader) Stop(ctx context.Context) error {
	uploader.grpcServer.GracefulStop()
	return nil
}

func (uploader *Uploader) serve(errCh chan error) {
	lis, err := net.Listen("tcp", uploader.config.Address)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}

	grpcServer := grpc.NewServer()
	types.RegisterUploaderServiceServer(grpcServer, uploader)
	uploader.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}
