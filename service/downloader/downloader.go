package downloader

import (
	"context"
	"net"

	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// Downloader manage the payload data download
type Downloader struct {
	cfg        *DownloaderConfig
	name       string
	pieceStore *client.StoreClient
	chain      *gnfd.Greenfield
}

// NewDownloaderService return a downloader instance.
func NewDownloaderService(cfg *DownloaderConfig) (*Downloader, error) {
	var err error
	downloader := &Downloader{
		cfg:  cfg,
		name: model.DownloaderService,
	}
	if downloader.pieceStore, err = client.NewStoreClient(cfg.PieceStoreConfig); err != nil {
		log.Errorw("failed to create piece store client", "err", err)
		return nil, err
	}
	if downloader.chain, err = gnfd.NewGreenfield(cfg.ChainConfig); err != nil {
		log.Errorw("failed to create chain client", "err", err)
		return nil, err
	}

	return downloader, nil
}

// Name implement the lifecycle interface
func (downloader *Downloader) Name() string {
	return downloader.name
}

// Start implement the lifecycle interface
func (downloader *Downloader) Start(ctx context.Context) error {
	errCh := make(chan error)

	go func(errCh chan error) {
		lis, err := net.Listen("tcp", downloader.cfg.Address)
		errCh <- err
		if err != nil {
			log.Errorw("syncer listen failed", "error", err)
			return
		}
		grpcServer := grpc.NewServer()
		stypes.RegisterDownloaderServiceServer(grpcServer, downloader)
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
