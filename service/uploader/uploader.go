package uploader

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	stoneclient "github.com/bnb-chain/greenfield-storage-provider/service/stonenode/client"
	types "github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
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
	stone      *stoneclient.StoneNodeClient
	grpcServer *grpc.Server
}

// NewUploaderService returns an instance of Uploader that implementation of
// the lifecycle.Service and UploaderService interface
func NewUploaderService(config *UploaderConfig) (*Uploader, error) {
	cache, _ := lru.New(model.LruCacheLimit)
	signer, err := signerclient.NewSignerClient(config.SignerGrpcAddress)
	if err != nil {
		return nil, err
	}
	stone, err := stoneclient.NewStoneNodeClient(config.StoneNodeGrpcAddress)
	pieceStore, err := psclient.NewStoreClient(config.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	spDB, err := sqldb.NewSQLStore(config.SPDBConfig)
	if err != nil {
		return nil, err
	}
	uploader := &Uploader{
		config:     config,
		cache:      cache,
		spDB:       spDB,
		stone:      stone,
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
	uploader.signer.Close()
	uploader.stone.Close()
	return nil
}

func (uploader *Uploader) serve(errCh chan error) {
	lis, err := net.Listen("tcp", uploader.config.GrpAddress)
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
