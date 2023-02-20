package stonehub

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store"
	"github.com/oleiade/lane"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/stone"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	// GCMemoryTimer define the period of GC memory stone.
	GCMemoryTimer = 60 * 60
	// GCDBTimer define the period of GC DB.
	GCDBTimer = 60 * 60
	// JobChannelSize define the size of receive stone job channel
	JobChannelSize = 100
	// GcChannelSize define the size of gc stone channel
	GcChannelSize = 100
	// WaitSealTimeoutHeight define timeout height of seal object
	WaitSealTimeoutHeight = 10
)

// Stone defines the interface that managed by stone hub.
type Stone interface {
	LastModifyTime() int64
	StoneKey() uint64
	GetStoneState() (string, error)
}

var _ lifecycle.Service = &StoneHub{}

// StoneHub manage all stones, the stone is an abstraction of job context and fsm.
type StoneHub struct {
	config    *StoneHubConfig
	jobDB     spdb.JobDB          // store the stones(include job and fsm) context
	metaDB    spdb.MetaDB         // store the storage provider meta
	stone     sync.Map            // hold all the running stones, goroutine safe
	jobQueue  *lane.Queue         // hold the stones that wait to be requested by stone node service
	jobCh     chan stone.StoneJob // stone receive channel
	gcCh      chan uint64         // notify stone hub delete the stone channel
	stopCh    chan struct{}
	running   atomic.Bool
	gcRunning atomic.Bool

	chain *gnfd.Greenfield
}

// NewStoneHubService return the StoneHub instance
func NewStoneHubService(hubCfg *StoneHubConfig) (*StoneHub, error) {
	hub := &StoneHub{
		config:   hubCfg,
		jobQueue: lane.NewQueue(),
		jobCh:    make(chan stone.StoneJob, JobChannelSize),
		gcCh:     make(chan uint64, GcChannelSize),
		stopCh:   make(chan struct{}),
	}
	chain, err := gnfd.NewGreenfield(hubCfg.ChainConfig)
	if err != nil {
		return nil, err
	}
	hub.chain = chain
	// init job and meta db
	if err = hub.initDB(); err != nil {
		return nil, err
	}
	if err = hub.loadStone(); err != nil {
		return nil, err
	}
	return hub, nil
}

// Name return the service name, implement the lifecycle interface.
func (hub *StoneHub) Name() string {
	return model.StoneHubService
}

// Start stone hub service, implement the lifecycle interface.
func (hub *StoneHub) Start(ctx context.Context) error {
	if hub.running.Swap(true) {
		return errors.New("stone hub has already started")
	}

	// start background task and rpc service
	go hub.eventLoop()
	go hub.serve()
	return nil
}

// Stop stone hub service, implement the lifecycle interface.
func (hub *StoneHub) Stop(ctx context.Context) error {
	if !hub.running.Swap(false) {
		return errors.New("stone hub has already stop")
	}

	close(hub.stopCh)
	close(hub.gcCh)

	var errs []error
	if err := hub.metaDB.Close(); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

// Serve starts grpc stone hub service.
func (hub *StoneHub) serve() {
	lis, err := net.Listen("tcp", hub.config.Address)
	if err != nil {
		log.Errorw("failed to listen", "address", hub.config.Address, "error", err)
		return
	}
	grpcServer := grpc.NewServer()
	stypes.RegisterStoneHubServiceServer(grpcServer, hub)
	// register reflection service
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorf("grpc serve error : %v", err)
		return
	}
}

// eventLoop background goroutine, responsible for GC, seal object, piece job receiving, etc.
func (hub *StoneHub) eventLoop() {
	gcMemTicker := time.NewTicker(time.Duration(GCMemoryTimer) * time.Second)
	gcDBTicker := time.NewTicker(time.Duration(GCDBTimer) * time.Second)
	for {
		select {
		case stoneJob := <-hub.jobCh:
			go hub.processStoneJob(stoneJob)
		case stoneKey := <-hub.gcCh:
			log.Infow("delete stone", "job_key", stoneKey)
			hub.stone.Delete(stoneKey)
		case <-gcMemTicker.C:
			go hub.gcMemoryStone()
		case <-gcDBTicker.C:
			go hub.gcDBStone()
			// TODO::retry the timeout stone
			//case <-retryTicker.C:
			//	hub.stoneRetry()
		case <-hub.stopCh:
			return
		}
	}
}

// processStoneJob according to the stone job types to process.
func (hub *StoneHub) processStoneJob(stoneJob stone.StoneJob) {
	switch job := stoneJob.(type) {
	case *stypes.PieceJob:
		hub.jobQueue.Enqueue(job)
		log.Infow("push secondary piece job to queue",
			"object_id", job.GetObjectId(), "object_size", job.GetPayloadSize(),
			"redundancy", job.GetRedundancyType(), "piece_idx", job.GetTargetIdx())
	case *stone.SealObjectJob:
		objectID := job.ObjectInfo.GetObjectId()
		st, ok := hub.stone.Load(objectID)
		if !ok {
			log.Warnw("stone has gone", "object_id", objectID)
			break
		}
		if _, ok := st.(*stone.UploadPayloadStone); !ok {
			log.Warnw("stone typecast to UploadPayloadStone error", "object_id", objectID)
			break
		}
		object := st.(*stone.UploadPayloadStone).GetObjectInfo()

		// TODO:: send request to signer
		go func(stoneJob *stone.UploadPayloadStone) {
			ctx := log.Context(context.Background(), object)
			defer hub.DeleteStone(stoneJob.StoneKey())
			_, err := hub.chain.ListenObjectSeal(context.Background(), object.BucketName, object.ObjectName, WaitSealTimeoutHeight)
			if err != nil {
				stoneJob.InterruptStone(ctx, err)
				log.CtxWarnw(ctx, "interrupt stone error", "error", err)
				return
			}
			err = stoneJob.ActionEvent(ctx, stone.SealObjectDoneEvent)
			if err != nil {
				log.CtxWarnw(ctx, "receive seal event, seal done fsm error", "error", err)
				return
			}
			log.CtxInfow(ctx, "seal object success")
		}(stoneJob.(*stone.UploadPayloadStone))
	default:
		log.Infow("unrecognized stone job type")
	}
}

// gcMemoryStone iterate the memory stone and garbage collect the abandoned stone.
func (hub *StoneHub) gcMemoryStone() {
	current := time.Now().Add(time.Second * -1 * time.Duration(GCMemoryTimer)).Unix()
	hub.stone.Range(func(key, value any) bool {
		val := value.(Stone)
		if val.LastModifyTime() <= current {
			log.Infow("gc memory stone", "object_id", key)
			hub.stone.Delete(key)
		}
		return true
	})
}

// gcDBStone iterate the db stone and garbage collect the zombie stone.
func (hub *StoneHub) gcDBStone() {
	if hub.gcRunning.Swap(true) {
		log.Errorw("gc stone db is running")
	}
	log.Infow("begin gc stones")
	defer hub.gcRunning.Swap(false)
	it := hub.jobDB.NewIterator(uint64(0))
	defer it.Release()
	for {
		if !it.IsValid() {
			if err := it.Error(); err != nil {
				log.Warnw("failed to gc, due to iterate stone", "error", err)
				return
			}
			log.Infow("succeed to gc stones")
			break
		}
		job := it.Value().(*ptypes.JobContext)
		if job.GetJobState() == ptypes.JobState_JOB_STATE_SEAL_OBJECT_DONE {
			err := hub.jobDB.DeleteJob(job.GetJobId())
			log.Infow("gc sealed job", "job", job, "error", err)
		}
		if len(job.GetJobErr()) != 0 {
			err := hub.jobDB.DeleteJob(job.GetJobId())
			log.Infow("gc failed job", "job", job, "error", err)
		}
		it.Next()
	}
}

// LoadStone read all stone form db and add stone hub.
func (hub *StoneHub) loadStone() error {
	if hub.gcRunning.Swap(true) {
		return errors.New("gc stone db is running")
	}
	log.Infow("begin load stones")
	defer hub.gcRunning.Swap(false)
	it := hub.jobDB.NewIterator(uint64(0))
	defer it.Release()
	for {
		if !it.IsValid() {
			if err := it.Error(); err != nil {
				log.Warnw("failed to load, due to iterate stone", "error", err)
				return err
			}
			log.Infow("succeed to load stones")
			break
		}
		job := it.Value().(*ptypes.JobContext)
		if len(job.GetJobErr()) != 0 {
			it.Next()
			continue
		}
		if job.GetJobState() == ptypes.JobState_JOB_STATE_SEAL_OBJECT_DONE {
			it.Next()
			continue
		}
		log.Infow("load unsealed job", "job", job)
		object, objErr := hub.jobDB.GetObjectInfoByJob(job.GetJobId())
		if objErr != nil {
			log.Errorw("load stone get object err", "job_id", job.GetJobId(), "error", objErr)
			it.Next()
			continue
		}
		st, stErr := stone.NewUploadPayloadStone(context.Background(), job, object,
			hub.jobDB, hub.metaDB, hub.jobCh, hub.gcCh)
		if stErr != nil {
			log.Errorw("load stone err", "job_id", job.GetJobId(),
				"object_id", object.GetObjectId(), "error", stErr)
			it.Next()
			continue
		}
		hub.SetStoneExclude(st)
		it.Next()
	}
	return nil
}

// initDB init job, meta, etc. db instance
func (hub *StoneHub) initDB() error {
	var (
		jobDB  spdb.JobDB
		metaDB spdb.MetaDB
		err    error
	)

	if jobDB, err = store.NewJobDB(hub.config.JobDBType, hub.config.JobSqlDBConfig); err != nil {
		log.Errorw("failed to init jobDB", "err", err)
		return err
	}
	if metaDB, err = store.NewMetaDB(hub.config.MetaDBType,
		hub.config.MetaLevelDBConfig, hub.config.MetaSqlDBConfig); err != nil {
		log.Errorw("failed to init metaDB", "err", err)
		return err
	}
	hub.jobDB = jobDB
	hub.metaDB = metaDB
	return nil
}

// ConsumeJob pop stone from remote stone queue.
// TODO::current only support sync piece data to secondary sp job.
func (hub *StoneHub) ConsumeJob() interface{} {
	return hub.jobQueue.Dequeue()
}

// HasStone return whether exist the stone corresponding to the stoneKey
func (hub *StoneHub) HasStone(stoneKey uint64) bool {
	_, ok := hub.stone.Load(stoneKey)
	return ok
}

// GetStone return the stone corresponding to the stoneKey
func (hub *StoneHub) GetStone(stoneKey uint64) Stone {
	st, ok := hub.stone.Load(stoneKey)
	if !ok {
		return nil
	}
	return st.(Stone)
}

// DeleteStone delete stone from memory.
func (hub *StoneHub) DeleteStone(stoneKey uint64) {
	hub.stone.Delete(stoneKey)
}

// SetStoneExclude set the stone, returns false if already exists
func (hub *StoneHub) SetStoneExclude(stone Stone) bool {
	_, exist := hub.stone.LoadOrStore(stone.StoneKey(), stone)
	return !exist
}
