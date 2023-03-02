package stonenode

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
)

var _ lifecycle.Service = &StoneNode{}

// StoneNode as background min execution unit, execute storage provider's background tasks
// implements the gRPC of StoneNodeService,
// TODO :: StoneNode support more task tpyes, such as gc etc.
type StoneNode struct {
	config     *StoneNodeConfig
	cache      *lru.Cache
	signer     *signerclient.SignerClient
	spDB       sqldb.SPDB
	chain      *greenfield.Greenfield
	pieceStore *psclient.StoreClient
	grpcServer *grpc.Server
}

// NewStoneNodeService return an instance of StoneNode and init resource
func NewStoneNodeService(config *StoneNodeConfig) (*StoneNode, error) {
	cache, _ := lru.New(model.LruCacheLimit)
	pieceStore, err := psclient.NewStoreClient(config.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	signer, err := signerclient.NewSignerClient(config.SignerGrpcAddress)
	if err != nil {
		return nil, err
	}
	chain, err := greenfield.NewGreenfield(config.ChainConfig)
	if err != nil {
		return nil, err
	}
	spDB, err := sqldb.NewSQLStore(config.SPDBConfig)
	if err != nil {
		return nil, err
	}
	node := &StoneNode{
		config:     config,
		cache:      cache,
		signer:     signer,
		spDB:       spDB,
		chain:      chain,
		pieceStore: pieceStore,
	}
	return node, nil
}

// Name return the stone node service name, for the lifecycle management
func (node *StoneNode) Name() string {
	return model.StoneNodeService
}

// Start the stone node gRPC service and background tasks
func (node *StoneNode) Start(ctx context.Context) error {
	errCh := make(chan error)
	go node.serve(errCh)
	err := <-errCh
	return err
}

// Stop the stone node gRPC service and recycle the resources
func (node *StoneNode) Stop(ctx context.Context) error {
	node.grpcServer.GracefulStop()
	node.signer.Close()
	node.chain.Close()
	return nil
}

func (node *StoneNode) serve(errCh chan error) {
	lis, err := net.Listen("tcp", node.config.GrpcAddress)
	errCh <- err
	if err != nil {
		log.Errorw("fail to listen", "err", err)
		return
	}

	grpcServer := grpc.NewServer()
	types.RegisterStoneNodeServiceServer(grpcServer, node)
	node.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

// EncodeReplicateSegments load segment data and encode according to redundancy type
func (node *StoneNode) EncodeReplicateSegments(
	ctx context.Context,
	objectId uint64,
	segments uint32,
	replicates int,
	rType storagetypes.RedundancyType) (
	data [][][]byte, err error) {
	params, err := node.spDB.GetStorageParams()
	if err != nil {
		return
	}
	for i := 0; i < replicates; i++ {
		data = append(data, make([][]byte, int(segments)))
	}

	var mux sync.Mutex
	var done int64
	errCh := make(chan error, 10)
	for segIdx := 0; segIdx < int(segments); segIdx++ {
		go func(segIdx int) {
			key := piecestore.EncodeSegmentPieceKey(objectId, uint32(segIdx))
			segmentData, err := node.pieceStore.GetSegment(ctx, key, 0, 0)
			if err != nil {
				errCh <- err
				return
			}
			if rType == storagetypes.REDUNDANCY_EC_TYPE {
				enodeData, err := redundancy.EncodeRawSegment(segmentData,
					int(params.GetRedundantDataChunkNum()),
					int(params.GetRedundantParityChunkNum()))
				if err != nil {
					errCh <- err
					return
				}
				mux.Lock()
				defer mux.Unlock()
				for idx, ec := range enodeData {
					data[idx][segIdx] = ec
				}
			} else {
				mux.Lock()
				defer mux.Unlock()
				for idx := 0; idx < replicates; idx++ {
					data[idx][segIdx] = segmentData
				}
			}
			if atomic.AddInt64(&done, 1) == int64(replicates) {
				errCh <- nil
				return
			}
		}(segIdx)
	}

	for {
		select {
		case err = <-errCh:
			return
		}
	}
}
