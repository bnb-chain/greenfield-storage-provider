package reactor

import (
	"context"
	"time"

	tmlog "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/service"
)

//TODO: this file is just for demo purpose

// StorageRequest stands for the new storage events happened on blockchain.
type StorageRequest struct {
	ObjectId   int64
	ObjectSize int64
	TxId       string
}

// SpJoinLeft stands for the new storage provider joins or leaves the blockchain.
type SpJoinLeft struct {
	Address string
	IsJoin  bool
	TxId    string
}

// ProviderUpdater will get updates for reactor to handle.
type ProviderUpdater interface {

	//SubscribeStorageRequest supports subscribers to consume events.
	SubscribeStorageRequest() <-chan StorageRequest

	//SubscribeSpJoinLeft supports subscribers to consume events.
	SubscribeSpJoinLeft() <-chan SpJoinLeft
}

type fakeProviderUpdater struct {
	service.BaseService

	logger  tmlog.Logger
	srChan  chan StorageRequest
	sjlChan chan SpJoinLeft
}

func (m *fakeProviderUpdater) OnStart(ctx context.Context) error {
	go m.mock()
	return nil
}

func (m *fakeProviderUpdater) OnStop() {
	if m.srChan != nil {
		close(m.srChan)
	}
	if m.sjlChan != nil {
		close(m.sjlChan)
	}
}

func (m *fakeProviderUpdater) mock() {
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case t := <-ticker:
			m.srChan <- StorageRequest{
				ObjectId:   t.Unix(),
				ObjectSize: t.Unix(),
				TxId:       "faketxid",
			}
			m.sjlChan <- SpJoinLeft{
				Address: "2af9963539a8d4945a4590b85c4331d9d3a8f5a4",
				IsJoin:  true,
				TxId:    "faketxid",
			}
		}
	}
}

func (m *fakeProviderUpdater) SubscribeStorageRequest() <-chan StorageRequest {
	return m.srChan
}

func (m *fakeProviderUpdater) SubscribeSpJoinLeft() <-chan SpJoinLeft {
	return m.sjlChan
}

func NewProviderUpdater(logger tmlog.Logger) *fakeProviderUpdater {
	monitor := &fakeProviderUpdater{
		logger:  logger,
		srChan:  make(chan StorageRequest),
		sjlChan: make(chan SpJoinLeft),
	}
	monitor.BaseService = *service.NewBaseService(logger, "providerUpdater", monitor)
	return monitor
}
