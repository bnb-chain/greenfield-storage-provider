package downloader

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/mock"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	stypesv1pb "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// Downloader manage the payload data download
type Downloader struct {
	cfg        *DownloaderConfig
	name       string
	pieceStore *client.StoreClient
	mockChain  *mock.InscriptionChainMock
}

// NewDownloaderService return a downloader instance.
func NewDownloaderService(cfg *DownloaderConfig) (*Downloader, error) {
	downloader := &Downloader{
		cfg:  cfg,
		name: model.DownloaderService,
	}
	pieceStore, err := client.NewStoreClient(cfg.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	downloader.pieceStore = pieceStore
	downloader.mockChain = mock.GetInscriptionChainMockSingleton()
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
		stypesv1pb.RegisterDownloaderServiceServer(grpcServer, downloader)
		reflection.Register(grpcServer)
		if err = grpcServer.Serve(lis); err != nil {
			log.Errorw("syncer serve failed", "error", err)
			return
		}
		return
	}(errCh)

	err := <-errCh
	return err
}

// Stop implement the lifecycle interface
func (downloader *Downloader) Stop(ctx context.Context) error {
	return nil
}
