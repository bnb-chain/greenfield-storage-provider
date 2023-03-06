package downloader

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	pscli "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
)

// Downloader implements the gRPC of DownloaderService,
// responsible for downloading object payload data
type Downloader struct {
	cfg        *DownloaderConfig
	spDB       sqldb.SPDB
	chain      *gnfd.Greenfield
	pieceStore *pscli.StoreClient
}

// NewDownloaderService returns an instance of Downloader that implementation of
// the lifecycle.Service and DownloaderService interface
func NewDownloaderService(cfg *DownloaderConfig) (*Downloader, error) {
	pieceStore, err := pscli.NewStoreClient(cfg.PieceStoreConfig)
	if err != nil {
		log.Errorw("failed to create piece store client", "err", err)
		return nil, err
	}
	chain, err := gnfd.NewGreenfield(cfg.ChainConfig)
	if err != nil {
		log.Errorw("failed to create chain client", "err", err)
		return nil, err
	}
	spDB, err := sqldb.NewSpDB(cfg.SpDBConfig)
	if err != nil {
		return nil, err
	}
	downloader := &Downloader{
		cfg:        cfg,
		spDB:       spDB,
		chain:      chain,
		pieceStore: pieceStore,
	}
	return downloader, nil
}

// Name return the downloader service name, for the lifecycle management
func (downloader *Downloader) Name() string {
	return model.DownloaderService
}

// Start the downloader gRPC service
func (downloader *Downloader) Start(ctx context.Context) error {
	errCh := make(chan error)

	go func(errCh chan error) {
		lis, err := net.Listen("tcp", downloader.cfg.GRPCAddress)
		errCh <- err
		if err != nil {
			log.Errorw("syncer listen failed", "error", err)
			return
		}
		grpcServer := grpc.NewServer()
		types.RegisterDownloaderServiceServer(grpcServer, downloader)
		reflection.Register(grpcServer)
		if err = grpcServer.Serve(lis); err != nil {
			log.Errorw("syncer serve failed", "error", err)
			return
		}
	}(errCh)

	err := <-errCh
	return err
}

// Stop the downloader gRPC service and recycle the resources
func (downloader *Downloader) Stop(ctx context.Context) error {
	return nil
}
