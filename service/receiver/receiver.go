package receiver

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
	"github.com/bnb-chain/greenfield-storage-provider/service/receiver/types"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
)

var _ lifecycle.Service = &Receiver{}

// Receiver implements the gRPC of ReceiverService,
// responsible for receive replicate object payload data
type Receiver struct {
	config     *ReceiverConfig
	cache      *lru.Cache
	signer     *signerclient.SignerClient
	pieceStore *psclient.StoreClient
	spDB       sqldb.SPDB
	grpcServer *grpc.Server
}

// NewReceiverService return a Receiver instance and init the resource
func NewReceiverService(config *ReceiverConfig) (*Receiver, error) {
	cache, _ := lru.New(model.LruCacheLimit)
	pieceStore, err := psclient.NewStoreClient(config.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	signer, err := signerclient.NewSignerClient(config.SignerGRPCAddress)
	if err != nil {
		return nil, err
	}
	spDB, err := sqldb.NewSpDB(config.SpDBConfig)
	if err != nil {
		return nil, err
	}
	s := &Receiver{
		config:     config,
		cache:      cache,
		pieceStore: pieceStore,
		spDB:       spDB,
		signer:     signer,
	}
	return s, nil
}

// Name return the receiver service name, for the lifecycle management
func (receiver *Receiver) Name() string {
	return model.ReceiverService
}

// Start the receiver background goroutine
func (receiver *Receiver) Start(ctx context.Context) error {
	errCh := make(chan error)
	go receiver.serve(errCh)
	err := <-errCh
	return err
}

// Stop the receiver gRPC service and recycle the resources
func (receiver *Receiver) Stop(ctx context.Context) error {
	receiver.grpcServer.GracefulStop()
	return nil
}

// serve
func (receiver *Receiver) serve(errCh chan error) {
	lis, err := net.Listen("tcp", receiver.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}

	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(model.MaxCallMsgSize), grpc.MaxSendMsgSize(model.MaxCallMsgSize))
	types.RegisterReceiverServiceServer(grpcServer, receiver)
	receiver.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}
