package p2p

import (
	"context"
	"net"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/p2p/node"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	UpdateSpInfoTimer = 5
)

var _ lifecycle.Service = &P2PService{}

type P2PService struct {
	config   *P2PServiceConfig
	node     *node.P2PNode
	spInfoDB *spdb.MetaDB
	stopCh   chan struct{}
}

func NewP2PService(cfg *P2PServiceConfig) (*P2PService, error) {
	service, err := node.NewDefault(context.Background(), cfg.NodeConfig())
	if err != nil {
		return nil, err
	}
	//TODO:: init spInfoDB

	p2p := &P2PService{
		config:   cfg,
		node:     service.(*node.P2PNode),
		spInfoDB: nil,
		stopCh:   make(chan struct{}),
	}
	return p2p, nil
}

// Name return the service name, implement the lifecycle interface.
func (service *P2PService) Name() string {
	return model.P2PService
}

// Start p2p service, implement the lifecycle interface.
func (service *P2PService) Start(ctx context.Context) error {
	// start background task and rpc service
	if err := service.node.Start(ctx); err != nil {
		return err
	}
	go service.serve()
	go service.eventLoop()
	return nil
}

// Stop p2p service, implement the lifecycle interface.
func (service *P2PService) Stop(ctx context.Context) error {
	close(service.stopCh)
	service.node.Stop()
	return nil
}

// Serve starts grpc stone hub service.
func (service *P2PService) serve() {
	lis, err := net.Listen("tcp", service.config.GrpcAddress)
	if err != nil {
		log.Errorw("failed to listen", "address", service.config.GrpcAddress, "error", err)
		return
	}
	grpcServer := grpc.NewServer()
	stypes.RegisterP2PServiceServer(grpcServer, service)
	// register reflection service
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorf("grpc serve error : %v", err)
		return
	}
}

// eventLoop background goroutine, responsible for GC, seal object, piece job receiving, etc.
func (service *P2PService) eventLoop() {
	updateSpInfoTicker := time.NewTicker(time.Duration(UpdateSpInfoTimer) * time.Second)
	for {
		select {
		case <-updateSpInfoTicker.C:
		//TODO:: query sp info from chain and update to db
		case <-service.stopCh:
			return
		}
	}
}
