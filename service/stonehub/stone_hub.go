package stonehub

import (
	"context"
	"errors"
	"github.com/bnb-chain/inscription-storage-provider/pkg/stone"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"github.com/oleiade/lane"
	"sync"
	"sync/atomic"
	"time"
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

// InscriptionClient defines the inscription client interface
// TBD::temporary interface, need to wait for the final version
type InscriptionClient interface {
	QueryObjectByTx(txHash []byte) (*types.ObjectInfo, error)
}

// Signer defines the storage provider signer service interface
// TBD::temporary interface, need to wait for the final version
type Signer interface {
	BroadcastMessage(interface{}) []byte
}

// MonitorInscription defines the storage provider event monitor interface
// TBD::temporary interface, need to wait for the final version
type MonitorInscription interface {
	SubscribeEvent(interface{}) chan interface{}
}

// StoneHub manage all stones, the stone is an abstraction of job context and fsm.
type StoneHub struct {
	jobDB  jobdb.JobDB   // job context db
	metaDB metadb.MetaDB // storage provider meta db
	stone  sync.Map      // stone map(stoneKey->stone), goroutine safe
	// entering the seal object stage will transfer the sealStone map,
	// because the index is converted from CreateObjectTX hash to SealObjectTX hash
	sealStone         sync.Map
	secondaryJobQueue *lane.Queue         // store secondary piece job, waiting pop by remote service
	jobCh             chan stone.StoneJob // stone fsm send job by jobCh to secondaryJobQueue
	stoneGC           chan string         // use to notify StoneHub delete stone

	stopCH  chan struct{}
	running atomic.Bool

	// TBD::temporary interface, need to wait for the final version
	insCli InscriptionClient
	signer Signer
	events MonitorInscription
}

// Start implement the lifecycle interface
func (hub *StoneHub) Start() error {
	if hub.running.Swap(true) {
		return errors.New("stone hub has started")
	}
	go hub.eventLoop()
	go hub.listenInscription()
	return nil
}

// Stop implement the lifecycle interface
func (hub *StoneHub) Stop() error {
	if !hub.running.Swap(false) {
		return errors.New("stone hub has stopped")
	}
	close(hub.stoneGC)
	close(hub.stopCH)
	return nil
}

// HasStone return whether exist the stone corresponding to the stoneKey
func (hub *StoneHub) HasStone(stoneKey string) bool {
	_, ok := hub.stone.Load(stoneKey)
	return ok
}

// GetStone return the stone corresponding to the stoneKey
func (hub *StoneHub) GetStone(stoneKey string) Stone {
	st, _ := hub.stone.Load(stoneKey)
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

// eventLoop background goroutine, responsible for GC, seal object, piece job receiving, etc.
func (hub *StoneHub) eventLoop() {
	gcMemTicker := time.NewTicker(time.Duration(GCMemoryTimer) * time.Second)
	gcDBTicker := time.NewTicker(time.Duration(GCDBTimer) * time.Second)
	for {
		select {
		case stoneJob := <-hub.jobCh:
			switch job := stoneJob.(type) {
			case *service.PieceJob:
				log.Info("stone hub receive the piece job", "hash", job.TxHash, "bucket", job.BucketName, "object", job.ObjectName)
				hub.secondaryJobQueue.Enqueue(job)
			case *stone.SealObjectJob:
				log.Info("stone hub receive the seal object job", "hash", job.StoneKey, "bucket", job.BucketName, "object", job.ObjectName)
				txHash := job.StoneKey
				stone, ok := hub.stone.Load(txHash)
				if !ok {
					log.Error("stone has gone", "hash", job.StoneKey)
					break
				}
				sealHash := hub.signer.BroadcastMessage(job)
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
	ch := hub.events.SubscribeEvent(struct{}{})
	for {
		select {
		case sealHash := <-ch:
			st, ok := hub.sealStone.Load(sealHash)
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
