package service

import (
	"context"
	"net"
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
	config                *metadata.MetadataConfig
	name                  string
	bsDB                  bsdb.BSDB
	bsDBBlockSyncer       bsdb.BSDB
	bsDBBlockSyncerBackUp bsdb.BSDB
	grpcServer            *grpc.Server
	dbSwitchTicker        *time.Ticker
}

// NewMetadataService returns an instance of Metadata that
// supply query service for Inscription network
func NewMetadataService(config *metadata.MetadataConfig) (metadata *Metadata, err error) {
	bsDBBlockSyncer, err := bsdb.NewBsDB(config, false)
	if err != nil {
		return nil, err
	}

	bsDBBlockSyncerBackUp, err := bsdb.NewBsDB(config, true)
	if err != nil {
		return nil, err
	}

	metadata = &Metadata{
		config:                config,
		name:                  model.MetadataService,
		bsDB:                  bsDBBlockSyncer,
		bsDBBlockSyncer:       bsDBBlockSyncer,
		bsDBBlockSyncerBackUp: bsDBBlockSyncerBackUp,
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
	metadata.startDBSwitchListener(time.Second * time.Duration(metadata.config.BsDBSwitchCheckIntervalSec))
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
	grpcServer := grpc.NewServer()
	metatypes.RegisterMetadataServiceServer(grpcServer, metadata)
	metadata.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

// startDBSwitchListener sets up a ticker to periodically check for a database switch signal.
// If a signal is detected, it triggers the switchDB() method to switch to the new database.
// The ticker is stopped when the Metadata gRPC service is stopped, ensuring that
// resources are properly managed and released.
func (metadata *Metadata) startDBSwitchListener(switchInterval time.Duration) {
	// create a ticker to periodically check for a new database name
	metadata.dbSwitchTicker = time.NewTicker(switchInterval)

	// launch a goroutine to handle the ticker events
	go func() {
		// loop until the context is canceled (e.g., when the Metadata service is stopped)
		for range metadata.dbSwitchTicker.C {
			// check if there is a signal from block syncer database to switch the database
			signal, err := metadata.bsDBBlockSyncer.GetSwitchDBSignal()
			if err != nil || signal == nil {
				log.Errorw("failed to get switch db signal", "err", err)
			}
			log.Debugf("switchDB check: signal: %t and BsDBFlag: %t", signal.IsMaster, metadata.config.BsDBFlag)
			// if a signal db is not equal to current metadata db, attempt to switch the database
			if signal.IsMaster != metadata.config.BsDBFlag {
				metadata.switchDB(signal.IsMaster)
			}
		}
	}()
}

// switchDB is responsible for switching between the primary and backup Block Syncer databases.
// Depending on the current value of the BsDBFlag in the Metadata configuration, it switches
// the active Block Syncer database to the backup or the primary database.
// After switching, it toggles the value of the BsDBFlag to indicate the active database.
func (metadata *Metadata) switchDB(flag bool) {
	if flag {
		metadata.bsDB = metadata.bsDBBlockSyncer
	} else {
		metadata.bsDB = metadata.bsDBBlockSyncerBackUp
	}
	metadata.config.BsDBFlag = flag
	log.Infow("db switched successfully")
}
