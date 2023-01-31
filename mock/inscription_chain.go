package mock

import (
	"errors"
	"sync"

	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
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
	objectByHash map[string]*ptypesv1pb.ObjectInfo
	objectByName map[string]*ptypesv1pb.ObjectInfo
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
		objectByHash: make(map[string]*ptypesv1pb.ObjectInfo),
		objectByName: make(map[string]*ptypesv1pb.ObjectInfo),
		events:       make(map[string][]chan interface{}),
		notifyCh:     make(chan *ChainEvent, 10),
		stopCh:       make(chan struct{}),
	}
	return cli
}

var singleMockChain *InscriptionChainMock
var once sync.Once

// GetInscriptionChainMockSingleton return single InscriptionChainMock instance
func GetInscriptionChainMockSingleton() *InscriptionChainMock {
	once.Do(func() {
		singleMockChain = NewInscriptionChainMock()
	})
	return singleMockChain
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
		}
	}
}

// QueryObjectByTx return the object info by create object tx hash.
func (cli *InscriptionChainMock) QueryObjectByTx(txHash []byte) (*ptypesv1pb.ObjectInfo, error) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	obj, ok := cli.objectByHash[string(txHash)]
	if !ok {
		return nil, errors.New("object is not exist")
	}
	return obj, nil
}

// QueryObjectByName return the object info by create object bucketName/objectName.
func (cli *InscriptionChainMock) QueryObjectByName(name string) (*ptypesv1pb.ObjectInfo, error) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	obj, ok := cli.objectByName[name]
	if !ok {
		return nil, errors.New("object is not exist")
	}
	return obj, nil
}

// CreateObjectByTxHash create the object info on the mock inscription chain.
func (cli *InscriptionChainMock) CreateObjectByTxHash(txHash []byte, object *ptypesv1pb.ObjectInfo) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.objectID++
	cli.objectHeight++

	object.TxHash = txHash
	object.ObjectId = cli.objectID
	object.Height = cli.objectHeight
	cli.objectByHash[string(txHash)] = object
	cli.objectByName[object.BucketName+"/"+object.ObjectName] = object
	cli.notifyCh <- &ChainEvent{
		EventType: CreateObject,
		Event:     object,
	}
}

// CreateObjectByName create the object info on the mock inscription chain.
func (cli *InscriptionChainMock) CreateObjectByName(txHash []byte, object *ptypesv1pb.ObjectInfo) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.objectID++
	cli.objectHeight++

	object.TxHash = txHash
	object.ObjectId = cli.objectID
	object.Height = cli.objectHeight
	cli.objectByName[object.BucketName+"/"+object.ObjectName] = object
}

// SealObjectByTxHash seal the object on the mock inscription chain.
func (cli *InscriptionChainMock) SealObjectByTxHash(txHash []byte, object *ptypesv1pb.ObjectInfo) {
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
