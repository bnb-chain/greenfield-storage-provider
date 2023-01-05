package uploader

import (
	"context"
	"errors"
	"net"
	"sync/atomic"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"google.golang.org/grpc"
)

// Uploader respond to putObjectTx/putObject impl.
type Uploader struct {
	config  *UploaderConfig
	name    string
	running atomic.Bool

	grpcServer  *grpc.Server
	signer      *signerClient
	stoneHub    *stoneHubClient
	eventWaiter *eventClient
	store       *storeClient
}

// NewUploaderService return the uploader instance
func NewUploaderService(cfg *UploaderConfig) (*Uploader, error) {
	var (
		err error
		u   *Uploader
	)
	u = &Uploader{
		config: cfg,
		name:   "Uploader",
	}
	u.signer = newSignerClient()
	u.stoneHub = newStoneHubClient(&cfg.StoneHubConfig, cfg.StorageProvider)
	u.eventWaiter = newEventClient()
	if u.store, err = newStoreClient(&u.config.PieceStoreConfig); err != nil {
		log.Warnw("failed to new store", "err", err)
		return nil, err
	}

	level := log.DebugLevel
	switch u.config.LogConfig.Level {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	default:
		level = log.InfoLevel
	}
	log.Init(level, u.config.LogConfig.FilePath)
	log.Infow("uploader succeed to init")
	return u, err
}

// Name implement the lifecycle interface
func (u *Uploader) Name() string {
	return u.name
}

// Start implement the lifecycle interface
func (u *Uploader) Start(ctx context.Context) error {
	if u.running.Swap(true) {
		return errors.New("uploader has started")
	}
	go u.Serve()
	log.Info("uploader succeed to start")
	return nil
}

// Serve starts grpc service.
func (u *Uploader) Serve() {
	lis, err := net.Listen("tcp", u.config.Address)
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}
	server := grpc.NewServer()
	service.RegisterUploaderServiceServer(server, &uploaderImpl{uploader: u})
	u.grpcServer = server
	if err := server.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

// Stop implement the lifecycle interface
func (g *Uploader) Stop(ctx context.Context) error {
	if !g.running.Swap(false) {
		return errors.New("uploader has stopped")
	}
	g.grpcServer.GracefulStop()
	log.Info("uploader succeed to stop")
	return nil
}
