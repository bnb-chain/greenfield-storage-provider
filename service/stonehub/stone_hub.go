package stonehub

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/store"
	"github.com/oleiade/lane"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/mock"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/stone"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	// GCMemoryTimer define the period of GC memory stone.
	GCMemoryTimer = 60 * 60
	// GCDBTimer define the period of GC DB.
	GCDBTimer = 24 * 60 * 60
	// JobChannelSize define the size of receive stone job channel
	JobChannelSize = 100
	// GcChannelSize define the size of gc stone channel
	GcChannelSize = 100
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
	config   *StoneHubConfig
	jobDB    jobdb.JobDBV2       // store the stones(include job and fsm) context
	metaDB   metadb.MetaDB       // store the storage provider meta
	stone    sync.Map            // hold all the running stones, goroutine safe
	jobQueue *lane.Queue         // hold the stones that wait to be requested by stone node service
	jobCh    chan stone.StoneJob // stone receive channel
	gcCh     chan uint64         // notify stone hub delete the stone channel
	stopCh   chan struct{}
	running  atomic.Bool

	// TODO::temporary mock interface, need to wait for the final version.
	insCli *mock.InscriptionChainMock
	signer *mock.SignerServerMock
	events *mock.InscriptionChainMock
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
	// TODO:: replace the mock green field chain related resource by official version
	{
		hub.insCli = mock.NewInscriptionChainMock()
		hub.signer = mock.NewSignerServerMock(hub.insCli)
		hub.events = hub.insCli
	}
	// init job and meta db
	if err := hub.initDB(); err != nil {
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

	// TODO:: use green field chain client replace the mock client
	{
		hub.insCli.Start()
		go hub.listenChain()
	}

	// TODO:: scan db load the unfinished stone
	// if err := hub.LoadDB(); err != nil {
	// 		return err
	// }

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
	// TODO:: use green field chain client replace the mock client
	{
		hub.insCli.Stop()
	}

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
			// TODO::gc the abandoned task by scan db
			// go hub.gcDBStone()
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
		hub.signer.BroadcastSealObjectMessage(object)
	default:
		log.Infow("unrecognized stone job type")
	}
}

// gcMemoryStone scan the memory stone and garbage collect the abandoned stone.
func (hub *StoneHub) gcMemoryStone() {
	current := time.Now().Add(time.Second * -1 * time.Duration(GCMemoryTimer)).Unix()
	hub.stone.Range(func(key, value any) bool {
		val := value.(Stone)
		state, err := val.GetStoneState()
		if err != nil {
			return true // skip err stone
		}
		if val.LastModifyTime() <= current || state == ptypes.JOB_STATE_ERROR {
			log.Infow("gc memory stone", "object_id", key)
			hub.stone.Delete(key)
		}
		return true
	})
}

// listenInscription listen to the subscribe events of green field chain.
// TODO::temporarily use the mock green field chain.
func (hub *StoneHub) listenChain() {
	ch := hub.events.SubscribeEvent(mock.SealObject)
	for {
		select {
		case event := <-ch:
			if event == nil {
				continue
			}
			object := event.(*ptypes.ObjectInfo)
			st, ok := hub.stone.Load(object.GetObjectId())
			if !ok {
				log.Infow("receive seal event, stone has gone")
				break
			}
			uploadStone, ok := st.(*stone.UploadPayloadStone)
			if !ok {
				log.Infow("receive seal event, stone typecast to UploadPayloadStone error", "object_id", object.GetObjectId())
				break
			}
			err := uploadStone.ActionEvent(context.Background(), stone.SealObjectDoneEvent)
			if err != nil {
				log.Infow("receive seal event, seal done fsm error", "object_id", object.GetObjectId())
				break
			}
			hub.stone.Delete(object.GetObjectId())
			//TODO::delete secondary integrity hash in metadb
		case <-hub.stopCh:
			return
		}
	}
}

// initDB init job, meta, etc. db instance
func (hub *StoneHub) initDB() error {
	var (
		jobDB  jobdb.JobDBV2
		metaDB metadb.MetaDB
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

// SetStoneExclude set the stone, returns false if already exists
func (hub *StoneHub) SetStoneExclude(stone Stone) bool {
	_, exist := hub.stone.LoadOrStore(stone.StoneKey(), stone)
	return !exist
}
