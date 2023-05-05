package tasknode

import (
	"context"
	"net"
	"time"

	managerclient "github.com/bnb-chain/greenfield-storage-provider/service/manager/client"
	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	managertypes "github.com/bnb-chain/greenfield-storage-provider/service/manager/types"
	p2pclient "github.com/bnb-chain/greenfield-storage-provider/service/p2p/client"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

const (
	fetchTaskPeriod = time.Second * 1
)

var _ lifecycle.Service = &TaskNode{}

// TaskNode as background min execution unit, execute storage provider's background tasks.
// implements the gRPC of TaskNodeService.
type TaskNode struct {
	config     *TaskNodeConfig
	cache      *lru.Cache
	signer     *signerclient.SignerClient
	p2p        *p2pclient.P2PClient
	manager    *managerclient.ManagerClient
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
	if taskNode.manager, err = managerclient.NewManagerClient(cfg.ManagerGrpcAddress); err != nil {
		log.Errorw("failed to create manager client", "error", err)
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
func (t *TaskNode) Name() string {
	return model.TaskNodeService
}

// Start the task node gRPC service and background tasks
func (t *TaskNode) Start(ctx context.Context) error {
	errCh := make(chan error)
	go t.serve(errCh)
	if err := <-errCh; err != nil {
		return err
	}
	go t.loopFetchTask()
	return nil
}

// Stop the task node gRPC service and recycle the resources
func (t *TaskNode) Stop(ctx context.Context) error {
	t.grpcServer.GracefulStop()
	t.signer.Close()
	t.p2p.Close()
	t.manager.Close()
	t.chain.Close()
	t.rcScope.Release()
	return nil
}

// serve start the task node gRPC service
func (t *TaskNode) serve(errCh chan error) {
	lis, err := net.Listen("tcp", t.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "error", err)
		return
	}

	options := utilgrpc.GetDefaultServerOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, utilgrpc.GetDefaultServerInterceptor()...)
	}
	t.grpcServer = grpc.NewServer(options...)
	types.RegisterTaskNodeServiceServer(t.grpcServer, t)
	reflection.Register(t.grpcServer)
	if err := t.grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "error", err)
		return
	}
}

// loopFetchTask fetches tasks from the manager.
func (t *TaskNode) loopFetchTask() {
	fetchTaskTicker := time.NewTicker(fetchTaskPeriod)
	for range fetchTaskTicker.C {
		// 1.fetch
		// 2.run
		// 3.report progress
		go func() {
			// TODO: refine limits
			resp, err := t.manager.AllocTask(context.Background(), rcmgr.InfiniteLimit())
			if err != nil {
				log.Errorw("failed to alloc task", "error", err)
				return
			}
			if resp.GetTask() == nil {
				log.Infow("alloc nil task")
				return
			}
			switch (resp.GetTask()).(type) {
			// TODO: refine it, split to seal task.
			case *managertypes.AllocTaskResponse_ReplicatePieceTask:
				log.Infow("alloc a replicate task", "response", resp)
				allocTask := (resp.GetTask()).(*managertypes.AllocTaskResponse_ReplicatePieceTask)
				objectInfo := allocTask.ReplicatePieceTask.GetObjectInfo()

				ctx := log.WithValue(context.Background(), "object_id", objectInfo.Id.String())
				task, err := newReplicateObjectTask(ctx, t, objectInfo)
				if err != nil {
					log.CtxErrorw(ctx, "failed to new replicate object task", "error", err)
					return
				}
				if err = task.init(); err != nil {
					log.CtxErrorw(ctx, "failed to init replicate object task", "error", err)
					return
				}
				waitCh := make(chan error)
				go task.execute(waitCh)
				if err = <-waitCh; err != nil {
					log.CtxErrorw(ctx, "failed to execute replicate object task", "error", err)
					return
				}
			case *managertypes.AllocTaskResponse_SealObjectTask:
				log.Infow("alloc a seal task", "response", resp)
				allocTask := (resp.GetTask()).(*managertypes.AllocTaskResponse_SealObjectTask)
				objectInfo := allocTask.SealObjectTask.GetObjectInfo()
				sealObjectInfo := allocTask.SealObjectTask.GetSealObject()

				ctx := log.WithValue(context.Background(), "object_id", objectInfo.Id.String())
				task, err := newSealObjectTask(ctx, t, objectInfo, sealObjectInfo)
				if err != nil {
					log.CtxErrorw(ctx, "failed to new seal object task", "error", err)
					return
				}
				if err = task.init(); err != nil {
					log.CtxErrorw(ctx, "failed to init seal object task", "error", err)
					return
				}
				waitCh := make(chan error)
				go task.execute(waitCh)
				if err = <-waitCh; err != nil {
					log.CtxErrorw(ctx, "failed to execute seal object task", "error", err)
					return
				}

			default:
				log.Infow("alloc unknown task", "response", resp)
				return
			}
		}()
	}
}
