package p2p

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/node"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	SegmentsTimeoutHeight = 3
	UpdateSpInfoPeriod    = 2
	RouterMaxSize         = 10240
)

var _ lifecycle.Service = &P2PService{}

// P2PService communicate with peers by p2p
type P2PService struct {
	config   *P2PServiceConfig
	pnodeId  types.NodeID
	pnode    *node.P2PNode
	preactor *P2PReactor
	spInfoDB *spdb.MetaDB

	routerSize int
	router     map[string]chan *libs.Envelope
	mux        sync.Mutex
}

// NewP2PService return an P2PService instance.
func NewP2PService(config *P2PServiceConfig) (*P2PService, error) {
	nodeCfg, err := config.makeP2pConfig()
	if err != nil {
		return nil, err
	}
	pnode, preactor, err := node.NewDefault(context.Background(), nodeCfg, NewP2PReactor)
	if err != nil {
		return nil, err
	}
	return &P2PService{
		config:   config,
		pnode:    pnode.(*node.P2PNode),
		preactor: preactor.(*P2PReactor),
		pnodeId:  pnode.(*node.P2PNode).GetNodeId(),
		router:   make(map[string]chan *libs.Envelope),
	}, nil
}

// Name implement the lifecycle interface
func (service *P2PService) Name() string {
	return model.P2PService
}

// Start implement the lifecycle interface
func (service *P2PService) Start(ctx context.Context) error {
	go service.pnode.Start(ctx)
	go service.eventloop(ctx)
	return nil
}

// Stop implement the lifecycle interface
func (service *P2PService) Stop(ctx context.Context) error {
	service.pnode.Stop()
	return nil
}

// handler the p2p request from peers.
func (service *P2PService) eventloop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(UpdateSpInfoPeriod) * time.Second)
	for {
		select {
		case envelope := <-service.preactor.SubscribeRequest():
			switch msg := envelope.Message.(type) {
			case *ptypes.AskApprovalRequest:
				_ = util.ComputeSegmentCount(msg.GetCreateObjectMsg().GetPayloadSize()) * uint32(SegmentsTimeoutHeight)
				// TODO:: 1. fill timeout height to CreateObjectMsg and send signer to sign
				// TODO:: 2. add secondary refuse approval Strategy
				service.preactor.PublishRequest(&libs.Envelope{
					To:      envelope.From,
					Message: msg,
				})
			case *ptypes.AckApproval:
				routerKey := hash.HexStringHash(msg.GetCreateObjectMsg().GetBucketName(), msg.GetCreateObjectMsg().GetObjectName())
				service.mux.Lock()
				service.router[routerKey] <- envelope
				service.mux.Unlock()
			case *ptypes.RefuseApproval:
				log.Errorw("secondary sp refuse approval", "info", envelope)
			}
		case <-ticker.C:
			//TODO:: query sp info from chain and update db
		case <-ctx.Done():
			return
		}
	}
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

func (service *P2PService) publishInfo(envelope *libs.Envelope) {
	service.preactor.PublishRequest(envelope)
}

func (service *P2PService) notifyByRouter(routerKey string, envelope *libs.Envelope) {
	service.mux.Lock()
	defer service.mux.Unlock()
	service.router[routerKey] <- envelope
}

func (service *P2PService) addRouter(routerKey string) error {
	service.mux.Lock()
	defer service.mux.Unlock()
	if service.routerSize >= RouterMaxSize {
		return errors.New("p2p overload protection is activated")
	}
	service.routerSize++
	if _, ok := service.router[routerKey]; !ok {
		return errors.New("has already")
	}
	service.router[routerKey] = make(chan *libs.Envelope)
	return nil
}

func (service *P2PService) deleteRouter(routerKey string) {
	service.mux.Lock()
	defer service.mux.Unlock()

	if _, ok := service.router[routerKey]; !ok {
		return
	}
	service.routerSize--
	delete(service.router, routerKey)
	return
}

func (service *P2PService) getRouterCh(routerKey string) (chan *libs.Envelope, error) {
	service.mux.Lock()
	defer service.mux.Unlock()
	if _, ok := service.router[routerKey]; !ok {
		return nil, errors.New("not exist")
	}
	return service.router[routerKey], nil
}
