package mock

import (
	"errors"
	"sync"

	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
)

var (
	CreateObject = "CreateObjectChainEvent"
	SealObject   = "SealObjectChainEvent"
)

// ChainEvent defines the inscription chain event
type ChainEvent struct {
	EventType string
	Event     interface{}
}

// InscriptionChainMock mock the inscription chain
type InscriptionChainMock struct {
	object       map[string]*types.ObjectInfo
	events       map[string][]chan interface{}
	notifyCh     chan *ChainEvent
	stopCh       chan struct{}
	objectID     uint64
	objectHeight uint64
	mu           sync.Mutex
}

// NewInscriptionChainMock return the InscriptionChainMock instance
func NewInscriptionChainMock() *InscriptionChainMock {
	cli := &InscriptionChainMock{
		object:   make(map[string]*types.ObjectInfo),
		events:   make(map[string][]chan interface{}),
		notifyCh: make(chan *ChainEvent, 10),
		stopCh:   make(chan struct{}),
	}
	return cli
}

// Start the background thread to publish the chain event.
func (cli *InscriptionChainMock) Start() {
	go cli.eventLoop()
}

// Stop the mock inscription chain
func (cli *InscriptionChainMock) Stop() {
	close(cli.stopCh)
	close(cli.notifyCh)
	for _, eventsCh := range cli.events {
		for _, ch := range eventsCh {
			close(ch)
		}
	}
}

// eventLoop start the background thread to publish the chain event.
func (cli *InscriptionChainMock) eventLoop() {
	for {
		select {
		case <-cli.stopCh:
			return
		case chainEvent := <-cli.notifyCh:
			if _, ok := cli.events[chainEvent.EventType]; !ok {
				continue
			}
			for _, ch := range cli.events[chainEvent.EventType] {
				ch <- chainEvent.Event
			}
			cli.events[chainEvent.EventType] = make([]chan any, 0)
		}
	}
}

// QueryObjectByTx return the object info by create object tx hash.
func (cli *InscriptionChainMock) QueryObjectByTx(txHash []byte) (*types.ObjectInfo, error) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	obj, ok := cli.object[string(txHash)]
	if !ok {
		return nil, errors.New("object is not exist.")
	}
	return obj, nil
}

// CreateObjectByTxHash create the object info on the mock inscription chain.
func (cli *InscriptionChainMock) CreateObjectByTxHash(txHash []byte, object *types.ObjectInfo) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.objectID++
	cli.objectHeight++

	object.TxHash = txHash
	object.ObjectId = cli.objectID
	object.Height = cli.objectHeight
	cli.object[string(txHash)] = object
	cli.notifyCh <- &ChainEvent{
		EventType: CreateObject,
		Event:     object,
	}
}

// SealObjectByTxHash seal the object on the mock inscription chain.
func (cli *InscriptionChainMock) SealObjectByTxHash(txHash []byte, object *types.ObjectInfo) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	object.TxHash = txHash
	cli.notifyCh <- &ChainEvent{
		EventType: SealObject,
		Event:     object,
	}
}

// SubscribeEvent subscribes the chain event.
func (cli *InscriptionChainMock) SubscribeEvent(event string) chan interface{} {
	eventCh := make(chan interface{})
	cli.events[event] = append(cli.events[event], eventCh)
	return eventCh
}
