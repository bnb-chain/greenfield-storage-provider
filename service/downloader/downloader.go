package downloader

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
)

var _ lifecycle.Service = &Downloader{}

// Downloader implements the gRPC of DownloaderService,
// responsible for downloading object payload data
type Downloader struct {
	config     *DownloaderConfig
	spDB       sqldb.SPDB
	pieceStore *psclient.StoreClient
}

// NewDownloaderService returns an instance of Downloader that implementation of
// the lifecycle.Service and DownloaderService interface
func NewDownloaderService(cfg *DownloaderConfig) (*Downloader, error) {
	var (
		downloader *Downloader
		err        error
	)
	downloader = &Downloader{
		config: cfg,
	}

	if downloader.spDB, err = sqldb.NewSpDB(cfg.SpDBConfig); err != nil {
		log.Errorw("failed to create sp db client", "error", err)
		return nil, err
	}
	if downloader.pieceStore, err = psclient.NewStoreClient(cfg.PieceStoreConfig); err != nil {
		log.Errorw("failed to create piece store client", "error", err)
		return nil, err
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
		lis, err := net.Listen("tcp", downloader.config.GRPCAddress)
		errCh <- err
		if err != nil {
			log.Errorw("failed to listen", "error", err)
			return
		}
		grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(model.MaxCallMsgSize), grpc.MaxSendMsgSize(model.MaxCallMsgSize))
		types.RegisterDownloaderServiceServer(grpcServer, downloader)
		reflection.Register(grpcServer)
		if err = grpcServer.Serve(lis); err != nil {
			log.Errorw("failed to serve", "error", err)
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
