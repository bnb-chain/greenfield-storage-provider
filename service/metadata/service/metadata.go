package service

import (
	"context"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// Metadata implements the gRPC of MetadataService,
// responsible for interact with SP for complex query service.
type Metadata struct {
	config         *metadata.MetadataConfig
	name           string
	bsDB           bsdb.BSDB
	grpcServer     *grpc.Server
	dbSwitchTicker *time.Ticker
	dbMutex        sync.RWMutex
}

// NewMetadataService returns an instance of Metadata that
// supply query service for Inscription network
func NewMetadataService(config *metadata.MetadataConfig) (metadata *Metadata, err error) {
	bsDB, err := bsdb.NewBsDB(config)
	if err != nil {
		return nil, err
	}
	metadata = &Metadata{
		config: config,
		name:   model.MetadataService,
		bsDB:   bsDB,
	}
	return
}

// Name return the metadata service name, for the lifecycle management
func (metadata *Metadata) Name() string {
	return metadata.name
}

// Start the metadata gRPC service
func (metadata *Metadata) Start(ctx context.Context) error {
	// Start the timed listener to switch the database
	metadata.startDBSwitchListener(time.Second * 5)
	errCh := make(chan error)
	go metadata.serve(errCh)
	err := <-errCh
	return err
}

// Stop the metadata gRPC service and recycle the resources
func (metadata *Metadata) Stop(ctx context.Context) error {
	metadata.grpcServer.GracefulStop()
	metadata.dbSwitchTicker.Stop()
	return nil
}

// Serve starts grpc service.
func (metadata *Metadata) serve(errCh chan error) {
	lis, err := net.Listen("tcp", metadata.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(metadata.dbLockInterceptor))
	metatypes.RegisterMetadataServiceServer(grpcServer, metadata)
	metadata.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

// startDBSwitchListener sets up a ticker to periodically check for a new database name
// and, if found, triggers the switchDB() method to switch to the new database.
// The ticker is stopped when the Metadata gRPC service is stopped, ensuring that
// resources are properly managed and released.
func (metadata *Metadata) startDBSwitchListener(switchInterval time.Duration) {
	// Create a ticker to periodically check for a new database name
	metadata.dbSwitchTicker = time.NewTicker(switchInterval)

	// Launch a goroutine to handle the ticker events
	go func() {
		// Loop until the context is canceled (e.g., when the Metadata service is stopped)
		for range metadata.dbSwitchTicker.C {
			// Check if there is a signal to switch the database
			signal, err := metadata.bsDB.GetSwitchDBSignal()
			// TODO REMOVE BARRY
			log.Debugf("switchDB check: signal: %t", signal)
			// TODO REMOVE BARRY

			// If a signal is detected, attempt to switch the database
			if signal == true {
				err = metadata.switchDB()
				if err != nil {
					log.Errorw("failed to switch db", "err", err)
				} else {
					log.Infow("db switched successfully")
				}
			}
		}
	}()
}

// dbLockInterceptor is a gRPC middleware that locks the dbMutex as a read lock
// before the handler is called, ensuring that multiple read operations can be
// executed concurrently without blocking each other. However, write operations
// (e.g., switchDB) will still block new read and write requests.
func (metadata *Metadata) dbLockInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Acquire the read lock
	metadata.dbMutex.RLock()

	// Defer the release of the read lock, so it will be unlocked after the handler has finished
	defer metadata.dbMutex.RUnlock()

	// Call the actual handler of the gRPC request
	return handler(ctx, req)
}

// switchDB is a method that switches the Metadata service's underlying database
// to a new one specified by the dbName parameter. This method acquires a write
// lock on the dbMutex, ensuring that new read and write requests are blocked
// until the database switch is complete. This provides consistency and prevents
// race conditions during the database switch operation.
func (metadata *Metadata) switchDB() error {
	metadata.dbMutex.Lock()
	defer metadata.dbMutex.Unlock()
	// update metadata.bsDB with the new database instance
	metadata.config.BSDBFlag = !metadata.config.BSDBFlag

	bsDB, err := bsdb.NewBsDB(metadata.config)
	if err != nil {
		log.Errorw("failed to switch db", "err", err)
		return err
	}
	metadata.bsDB = bsDB
	// TODO REMOVE BARRY
	signal, err := metadata.bsDB.GetSwitchDBSignal()
	log.Debugf("switchDB successfully, signal: %t", signal)
	// TODO REMOVE BARRY
	return nil
}
