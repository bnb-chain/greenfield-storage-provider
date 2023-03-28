package tasknode

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-common/go/redundancy"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	mdgrpc "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/grpc"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	p2pclient "github.com/bnb-chain/greenfield-storage-provider/service/p2p/client"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

var _ lifecycle.Service = &TaskNode{}

// TaskNode as background min execution unit, execute storage provider's background tasks
// implements the gRPC of TaskNodeService,
// TODO :: TaskNode support more task types, such as gc etc.
type TaskNode struct {
	config     *TaskNodeConfig
	cache      *lru.Cache
	signer     *signerclient.SignerClient
	p2p        *p2pclient.P2PClient
	spDB       sqldb.SPDB
	chain      *greenfield.Greenfield
	rcScope    rcmgr.ResourceScope
	pieceStore *psclient.StoreClient
	grpcServer *grpc.Server
}

// NewTaskNodeService return an instance of TaskNode and init resource
func NewTaskNodeService(cfg *TaskNodeConfig) (*TaskNode, error) {
	var (
		taskNode *TaskNode
		err      error
	)

	taskNode = &TaskNode{
		config: cfg,
	}
	if taskNode.cache, err = lru.New(model.LruCacheLimit); err != nil {
		log.Errorw("failed to create lru cache", "error", err)
		return nil, err
	}
	if taskNode.pieceStore, err = psclient.NewStoreClient(cfg.PieceStoreConfig); err != nil {
		log.Errorw("failed to create piece store client", "error", err)
		return nil, err
	}
	if taskNode.signer, err = signerclient.NewSignerClient(cfg.SignerGrpcAddress); err != nil {
		log.Errorw("failed to create signer client", "error", err)
		return nil, err
	}
	if taskNode.p2p, err = p2pclient.NewP2PClient(cfg.P2PGrpcAddress); err != nil {
		log.Errorw("failed to create p2p server client", "error", err)
		return nil, err
	}
	if taskNode.chain, err = greenfield.NewGreenfield(cfg.ChainConfig); err != nil {
		log.Errorw("failed to create chain client", "error", err)
		return nil, err
	}
	if taskNode.spDB, err = sqldb.NewSpDB(cfg.SpDBConfig); err != nil {
		log.Errorw("failed to create sp db client", "error", err)
		return nil, err
	}
	if taskNode.rcScope, err = rcmgr.ResrcManager().OpenService(model.TaskNodeService); err != nil {
		log.Errorw("failed to open task node resource scope", "error", err)
		return nil, err
	}

	return taskNode, nil
}

// Name return the task node service name, for the lifecycle management
func (taskNode *TaskNode) Name() string {
	return model.TaskNodeService
}

// Start the task node gRPC service and background tasks
func (taskNode *TaskNode) Start(ctx context.Context) error {
	errCh := make(chan error)
	go taskNode.serve(errCh)
	err := <-errCh
	return err
}

// Stop the task node gRPC service and recycle the resources
func (taskNode *TaskNode) Stop(ctx context.Context) error {
	taskNode.grpcServer.GracefulStop()
	taskNode.signer.Close()
	taskNode.p2p.Close()
	taskNode.chain.Close()
	taskNode.rcScope.Release()
	return nil
}

// serve start the task node gRPC service
func (taskNode *TaskNode) serve(errCh chan error) {
	lis, err := net.Listen("tcp", taskNode.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "error", err)
		return
	}

	options := utilgrpc.GetDefaultServerOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, mdgrpc.GetDefaultServerInterceptor()...)
	}
	taskNode.grpcServer = grpc.NewServer(options...)
	types.RegisterTaskNodeServiceServer(taskNode.grpcServer, taskNode)
	reflection.Register(taskNode.grpcServer)
	if err := taskNode.grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "error", err)
		return
	}
}

// EncodeReplicateSegments load segment data and encode according to redundancy type
func (taskNode *TaskNode) EncodeReplicateSegments(ctx context.Context, objectID uint64, segments uint32, replicates int,
	rType storagetypes.RedundancyType) (data [][][]byte, err error) {
	params, err := taskNode.spDB.GetStorageParams()
	if err != nil {
		return
	}
	for i := 0; i < int(segments); i++ {
		data = append(data, make([][]byte, replicates))
	}
	log.Debugw("start to encode payload", "object_id", objectID, "segment_count", segments,
		"replicas", replicates, "redundancy_type", rType)

	var done int64
	errCh := make(chan error, 10)
	for segIdx := 0; segIdx < int(segments); segIdx++ {
		go func(segIdx int) {
			key := piecestore.EncodeSegmentPieceKey(objectID, uint32(segIdx))
			segmentData, err := taskNode.pieceStore.GetSegment(ctx, key, 0, 0)
			if err != nil {
				errCh <- err
				return
			}
			if rType == storagetypes.REDUNDANCY_EC_TYPE {
				encodeData, err := redundancy.EncodeRawSegment(segmentData,
					int(params.GetRedundantDataChunkNum()),
					int(params.GetRedundantParityChunkNum()))
				if err != nil {
					errCh <- err
					return
				}
				copy(data[segIdx], encodeData)
			} else {
				for rIdx := 0; rIdx < replicates; rIdx++ {
					data[segIdx][rIdx] = segmentData
				}
			}
			log.Debugw("finish to encode payload", "object_id", objectID, "segment_idx", segIdx)
			if atomic.AddInt64(&done, 1) == int64(segments) {
				errCh <- nil
				return
			}
		}(segIdx)
	}
	for err = range errCh {
		return
	}
	return
}
