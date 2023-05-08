package lifecycle

import (
	"context"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/core/module"
)

// Lifecycle is the interface to the service life cycle management subsystem.
// The ServiceLifecycle tracks the Service life cycle, listens to the signal of the
// process for graceful exit.
//
// All managed services must first call RegisterServices to register with ServiceLifecycle.
type Lifecycle interface {
	// RegisterModule registers service to ServiceLifecycle for managing.
	RegisterModule(modular ...module.Modular)
	// StartModules starts all registered services by calling Service.Start method.
	StartModules(ctx context.Context) *Lifecycle
	// StopModules stops all registered services by calling Service.Stop method.
	StopModules(ctx context.Context)
	// Signals listens the system signals for gracefully stop the registered services.
	Signals(sigs ...os.Signal) *Lifecycle
	// Wait waits the signal for stopping the ServiceLifecycle, before stopping the
	// ServiceLifecycle will call StopServices stops all registered services.
	Wait(ctx context.Context)
}
