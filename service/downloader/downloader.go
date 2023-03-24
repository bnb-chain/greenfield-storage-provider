package downloader

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	mwgrpc "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/grpc"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

var _ lifecycle.Service = &Downloader{}

// Downloader implements the gRPC of DownloaderService,
// responsible for downloading object payload data
type Downloader struct {
	config     *DownloaderConfig
	spDB       sqldb.SPDB
	pieceStore *psclient.StoreClient
	grpcServer *grpc.Server
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
	go downloader.serve(errCh)
	err := <-errCh
	return err
}

// Stop the downloader gRPC service and recycle the resources
func (downloader *Downloader) Stop(ctx context.Context) error {
	downloader.grpcServer.GracefulStop()
	return nil
}

func (downloader *Downloader) serve(errCh chan error) {
	lis, err := net.Listen("tcp", downloader.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "error", err)
		return
	}

	options := utilgrpc.GetDefaultServerOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, mwgrpc.GetDefaultServerInterceptor()...)
	}
	downloader.grpcServer = grpc.NewServer(options...)
	types.RegisterDownloaderServiceServer(downloader.grpcServer, downloader)
	reflection.Register(downloader.grpcServer)
	if err = downloader.grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to serve", "error", err)
		return
	}
}
