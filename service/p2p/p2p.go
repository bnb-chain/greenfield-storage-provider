package p2p

import (
	"bytes"
	"context"
	"net"
	"sort"
	"time"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/service/p2p/types"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

const UpdateSPDuration = 2

var _ lifecycle.Service = &P2PServer{}

// P2PServer p2p service
type P2PServer struct {
	config     *P2PConfig
	node       *p2p.Node
	signer     *signerclient.SignerClient
	spDB       sqldb.SPDB
	grpcServer *grpc.Server
	stopCh     chan struct{}
}

// NewP2PServer return an instance of P2PServer
func NewP2PServer(config *P2PConfig) (*P2PServer, error) {
	signer, err := signerclient.NewSignerClient(config.SignerGrpcAddress)
	if err != nil {
		return nil, err
	}
	node, err := p2p.NewNode(config.P2PConfig, config.SpOperatorAddress, signer)
	if err != nil {
		return nil, err
	}
	spDB, err := sqldb.NewSpDB(config.SpDBConfig)
	if err != nil {
		return nil, err
	}
	p := &P2PServer{
		config: config,
		node:   node,
		signer: signer,
		spDB:   spDB,
		stopCh: make(chan struct{}),
	}
	return p, nil
}

// Name return the p2p server name, for the lifecycle management
func (p *P2PServer) Name() string {
	return model.P2PService
}

// Start the p2p server background goroutine
func (p *P2PServer) Start(ctx context.Context) error {
	errCh := make(chan error)
	go p.serve(errCh)
	go p.eventLoop()
	err := p.node.Start(ctx)
	if err != nil {
		return err
	}
	err = <-errCh
	return err
}

// Stop the p2p server gRPC service and recycle the resources
func (p *P2PServer) Stop(ctx context.Context) error {
	close(p.stopCh)
	p.grpcServer.GracefulStop()
	p.node.Stop(ctx)
	return nil
}

func (p *P2PServer) serve(errCh chan error) {
	lis, err := net.Listen("tcp", p.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}

	options := []grpc.ServerOption{}
	if metrics.GetMetrics().Enabled() {
		options = append(options, utilgrpc.GetDefaultServerInterceptor()...)
	}
	p.grpcServer = grpc.NewServer(options...)
	p2ptypes.RegisterP2PServiceServer(p.grpcServer, p)
	reflection.Register(p.grpcServer)
	if err := p.grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

func (p *P2PServer) eventLoop() {
	ticker := time.NewTicker(time.Duration(UpdateSPDuration) * time.Second)
	var integrity []byte
	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			spList, err := p.spDB.FetchAllSp()
			if err != nil {
				log.Warnw("failed to fetch all SPs", "error", err)
				continue
			}
			var spOpAddrs []string
			for _, sp := range spList {
				spOpAddrs = append(spOpAddrs, sp.GetOperatorAddress())
			}
			sort.Strings(spOpAddrs)
			var spOpByte [][]byte
			for _, addr := range spOpAddrs {
				spOpByte = append(spOpByte, []byte(addr))
			}
			currentIntegrity := hash.GenerateIntegrityHash(spOpByte)
			if bytes.Equal(currentIntegrity, integrity) {
				continue
			}
			integrity = currentIntegrity[:]
			p.node.PeersProvider().UpdateSp(spOpAddrs)
		}
	}
}
