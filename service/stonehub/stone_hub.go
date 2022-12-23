package stonehub

import (
	"github.com/bnb-chain/inscription-storage-provider/pkg/stone"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store"
	"github.com/oleiade/lane"
	"sync"
)

type Stone interface {
}

type Strategy interface {
}

type StoneHub struct {
	JobDB         store.JobDB
	MetaDB        store.MetaDB
	StoneJob      *sync.Map
	StoneJobQueue *lane.PQueue
	strategy      *Strategy
	stoneJobCh    chan stone.StoneJob
	stoneGC       chan uint64
	stopCH        chan struct{}
}

func (hub *StoneHub) GetStoneByTxHash(txHash []byte) (Stone, error) { return nil, nil }
func (hub *StoneHub) GetSegmentsSize() uint64                       { return 0 }
func (hub *StoneHub) GetInlineSize() uint64                         { return 0 }
func (hub *StoneHub) PopUploadSecondaryPieceJob() *service.PieceJob { return nil }
func (hub *StoneHub) GetStone(jobId uint64) Stone                   { return struct{}{} }
func (hub *StoneHub) SetStone(stone Stone) error                    { return nil }
func (hub *StoneHub) PopStoneJob() stone.StoneJob                   { return struct{}{} }

// EventLoop receive stone job, stone timeout, gc, etc.
func (hub *StoneHub) EventLoop() {}
