package downloader

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	pscli "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// Downloader manage the payload data download
type Downloader struct {
	cfg        *DownloaderConfig
	chain      *gnfd.Greenfield
	pieceStore *pscli.StoreClient
}

// NewDownloaderService return a downloader instance.
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
	downloader := &Downloader{
		cfg:        cfg,
		chain:      chain,
		pieceStore: pieceStore,
	}
	return downloader, nil
}

// Name implement the lifecycle interface
func (downloader *Downloader) Name() string {
	return model.DownloaderService
}

// Start implement the lifecycle interface
func (downloader *Downloader) Start(ctx context.Context) error {
	errCh := make(chan error)

	go func(errCh chan error) {
		lis, err := net.Listen("tcp", downloader.cfg.GrpcAddress)
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

// Stop implement the lifecycle interface
func (downloader *Downloader) Stop(ctx context.Context) error {
	return nil
}
