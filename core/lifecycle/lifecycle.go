package lifecycle

import (
	"context"
	"os"
)

// Service provides abstract methods to control the lifecycle of a service
// Every service must implement Service interface.
type Service interface {
	// Name describe service name
	Name() string
	// Start a service, this method should be used in non-block form
	Start(ctx context.Context) error
	// Stop a service, this method should be used in non-block form
	Stop(ctx context.Context) error
}

// Lifecycle is an interface to describe how service is managed.
// The Lifecycle tracks the Service lifecycle, listens for signals from
// the process to ensure a graceful shutdown.
//
// All managed services must firstly call RegisterServices to register with Lifecycle.
type Lifecycle interface {
	// RegisterServices registers service to ServiceLifecycle for managing.
	RegisterServices(modular ...Service)
	// StartServices starts all registered services by calling Service.Start method.
	StartServices(ctx context.Context) Lifecycle
	// StopServices stops all registered services by calling Service.Stop method.
	StopServices(ctx context.Context)
	// Signals listens the system signals for gracefully stop the registered services.
	Signals(sigs ...os.Signal) Lifecycle
	// Wait waits the signal for stopping the ServiceLifecycle, before stopping
	// the ServiceLifecycle will call StopServices stops all registered services.
	Wait(ctx context.Context)
}
