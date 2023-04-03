package uploader

import (
	"context"
	"net"

	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	mwgrpc "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/grpc"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	tasknodeclient "github.com/bnb-chain/greenfield-storage-provider/service/tasknode/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

var _ lifecycle.Service = &Uploader{}

// Uploader implements the gRPC of UploaderService,
// responsible for uploading object payload data.
type Uploader struct {
	config     *UploaderConfig
	cache      *lru.Cache
	spDB       sqldb.SPDB
	pieceStore *psclient.StoreClient
	signer     *signerclient.SignerClient
	taskNode   *tasknodeclient.TaskNodeClient
	grpcServer *grpc.Server
}

// NewUploaderService returns an instance of Uploader that implementation of
// the lifecycle.Service and UploaderService interface
func NewUploaderService(cfg *UploaderConfig) (*Uploader, error) {
	var (
		uploader *Uploader
		err      error
	)

	uploader = &Uploader{
		config: cfg,
	}

	if uploader.cache, err = lru.New(model.LruCacheLimit); err != nil {
		log.Errorw("failed to create lru cache", "error", err)
		return nil, err
	}
	if uploader.signer, err = signerclient.NewSignerClient(cfg.SignerGrpcAddress); err != nil {
		log.Errorw("failed to create signer client", "error", err)
		return nil, err
	}
	if uploader.taskNode, err = tasknodeclient.NewTaskNodeClient(cfg.TaskNodeGrpcAddress); err != nil {
		log.Errorw("failed to create task node client", "error", err)
		return nil, err
	}
	if uploader.pieceStore, err = psclient.NewStoreClient(cfg.PieceStoreConfig); err != nil {
		log.Errorw("failed to create piece store client", "error", err)
		return nil, err
	}
	if uploader.spDB, err = sqldb.NewSpDB(cfg.SpDBConfig); err != nil {
		log.Errorw("failed to create sp db client", "error", err)
		return nil, err
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
	uploader.signer.Close()
	uploader.taskNode.Close()
	return nil
}

// serve start the uploader gRPC service
func (uploader *Uploader) serve(errCh chan error) {
	lis, err := net.Listen("tcp", uploader.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "error", err)
		return
	}

	options := utilgrpc.GetDefaultServerOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, mwgrpc.GetDefaultServerInterceptor()...)
	}
	uploader.grpcServer = grpc.NewServer(options...)
	types.RegisterUploaderServiceServer(uploader.grpcServer, uploader)
	reflection.Register(uploader.grpcServer)
	if err := uploader.grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "error", err)
		return
	}
}
