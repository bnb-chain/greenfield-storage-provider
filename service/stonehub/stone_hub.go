package stonehub

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oleiade/lane"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/inscription-storage-provider/model"

	"github.com/bnb-chain/inscription-storage-provider/store/metadb/leveldb"

	"github.com/bnb-chain/inscription-storage-provider/store/jobdb/jobsql"

	"github.com/bnb-chain/inscription-storage-provider/mock"
	"github.com/bnb-chain/inscription-storage-provider/pkg/stone"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb/jobmemory"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var (
	// GCMemoryTimer define the cycle of GC memory stone
	GCMemoryTimer = 60 * 60

	// GCDBTimer define the cycle of GC DB
	GCDBTimer = 24 * 60 * 60
)

// Stone defines the interface that the stone managed by Stonehub needs to implement
type Stone interface {
	LastModifyTime() int64
	StoneKey() string
	GetStoneState() (string, error)
}

// StoneHub manage all stones, the stone is an abstraction of job context and fsm.
type StoneHub struct {
	config *StoneHubConfig
	name   string
	jobDB  jobdb.JobDB   // job context db
	metaDB metadb.MetaDB // storage provider meta db
	stone  sync.Map      // stone map(stoneKey->stone), goroutine safe
	// entering the seal object stage will transfer the stone to sealStone map,
	// because the index is converted from CreateObjectTX hash to SealObjectTX hash
	sealStone         sync.Map
	secondaryJobQueue *lane.Queue         // store secondary piece job, waiting pop by remote service
	jobCh             chan stone.StoneJob // stone fsm send job by jobCh to secondaryJobQueue
	stoneGC           chan string         // use to notify StoneHub delete stone

	stopCH  chan struct{}
	running atomic.Bool

	// TBD::temporary mock interface, need to wait for the final version
	insCli *mock.InscriptionChainMock
	signer *mock.SignerServerMock
	events *mock.InscriptionChainMock
}

// NewStoneHubService return the StoneHub instance
func NewStoneHubService(hubCfg *StoneHubConfig) (*StoneHub, error) {
	hub := &StoneHub{
		config:            hubCfg,
		name:              "StoneHub",
		secondaryJobQueue: lane.NewQueue(),
		jobCh:             make(chan stone.StoneJob, 100),
		stoneGC:           make(chan string, 10),
		stopCH:            make(chan struct{}),
	}
	// mock inscription chain related resource
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

// initDB init job, meta, etc. db instance
func (hub *StoneHub) initDB() (err error) {
	switch hub.config.JobDBType {
	case model.MemoryDB:
		hub.jobDB = jobmemory.NewMemJobDB()
	case model.MySqlDB:
		// TODO:: add mysql db
		if hub.config.JobDB == nil {
			hub.config.JobDB = DefaultStoneHubConfig.JobDB
		}
		hub.jobDB, err = jobsql.NewJobMetaImpl(hub.config.JobDB)
		if err != nil {
			return
		}
	default:
		return errors.New(fmt.Sprintf("job db not support type %s", hub.config.JobDBType))
	}

	switch hub.config.MetaDBType {
	case model.LevelDB:
		// TODO:: add leveldb, temporarily replace with memory job db
		if hub.config.MetaDB == nil {
			hub.config.MetaDB = DefaultStoneHubConfig.MetaDB
		}
		hub.metaDB, err = leveldb.NewMetaDB(hub.config.MetaDB)
		if err != nil {
			return
		}
	default:
		return errors.New(fmt.Sprintf("job db not support type %s", hub.config.MetaDBType))
	}
	return nil
}

// Name implement the lifecycle interface
func (hub *StoneHub) Name() string {
	return hub.name
}

// Start implement the lifecycle interface
func (hub *StoneHub) Start(ctx context.Context) error {
	if hub.running.Swap(true) {
		return errors.New("stone hub has started")
	}
	hub.insCli.Start()
	go hub.eventLoop()
	go hub.listenInscription()
	go hub.serve()
	return nil
}

// Stop implement the lifecycle interface
func (hub *StoneHub) Stop(ctx context.Context) error {
	if !hub.running.Swap(false) {
		return errors.New("stone hub has already stop")
	}
	hub.insCli.Stop()
	close(hub.stopCH)
	close(hub.stoneGC)
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
		log.Errorf("stone hub service failed to listen: %v", err)
		return
	}
	grpcServer := grpc.NewServer()
	service.RegisterStoneHubServiceServer(grpcServer, hub)
	// register reflection service
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorf("stone hub service failed to listen: %v", err)
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
			switch job := stoneJob.(type) {
			case *service.PieceJob:
				log.Infow("stone hub receive the piece job", "job info", job)
				hub.secondaryJobQueue.Enqueue(job)
			case *stone.SealObjectJob:
				txHash := job.StoneKey
				stone, ok := hub.stone.Load(txHash)
				if !ok {
					log.Warnw("stone has gone", "hash", job.StoneKey)
					break
				}
				object := &types.ObjectInfo{
					BucketName: job.BucketName,
					ObjectName: job.ObjectName,
				}
				sealHash := hub.signer.BroadcastSealObjectMessage(object)
				hub.sealStone.Store(string(sealHash), stone)
				hub.stone.Delete(txHash)
			default:
			}
		case stoneKey := <-hub.stoneGC:
			log.Info("stone hub receive gc stone", "hash", stoneKey)
			hub.stone.Delete(stoneKey)
			hub.sealStone.Delete(stoneKey)
		case <-gcMemTicker.C:
			log.Info("stone hub begin gc stone")
			current := time.Now().Add(time.Second * -1 * time.Duration(GCMemoryTimer)).Unix()
			hub.stone.Range(func(key, value any) bool {
				val := value.(Stone)
				state, _ := val.GetStoneState()
				if val.LastModifyTime() <= current || state == types.JOB_STATE_ERROR {
					stoneKey := key.(string)
					hub.stoneGC <- stoneKey
				}
				return true
			})
			// TBO::add another seal timeout ticker and retry
			hub.sealStone.Range(func(key, value any) bool {
				val := value.(Stone)
				if val.LastModifyTime() <= current {
					stoneKey := key.(string)
					hub.stoneGC <- stoneKey
				}
				return true
			})
			log.Info("stone hub end gc stone")
		case <-gcDBTicker.C:
			// TBO::scan db, gc the task
		case <-hub.stopCH:
			return
		}
	}
}

// listenInscription listen to the concerned events of inscription chain
// TBD::temporarily use the interface mock.
func (hub *StoneHub) listenInscription() {
	ch := hub.events.SubscribeEvent(mock.SealObject)
	for {
		select {
		case sealHash := <-ch:
			object := sealHash.(*types.ObjectInfo)
			st, ok := hub.sealStone.Load(string(object.TxHash))
			if !ok {
				break
			}
			uploadStone := st.(*stone.UploadPayloadStone)
			err := uploadStone.ActionEvent(context.Background(), stone.SealObjectDoneEvent)
			if err != nil {
				break
			}
			hub.sealStone.Delete(sealHash)
		case <-hub.stopCH:
			return
		}
	}
}

// HasStone return whether exist the stone corresponding to the stoneKey
func (hub *StoneHub) HasStone(stoneKey string) bool {
	_, ok := hub.stone.Load(stoneKey)
	return ok
}

// GetStone return the stone corresponding to the stoneKey
func (hub *StoneHub) GetStone(stoneKey string) Stone {
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

// PopUploadSecondaryPieceJob return secondary piece job from secondaryJobQueue
func (hub *StoneHub) PopUploadSecondaryPieceJob() *service.PieceJob {
	stoneJob := hub.secondaryJobQueue.Dequeue()
	pieceJob := stoneJob.(*service.PieceJob)
	return pieceJob
}
