package gfspapp

import (
	"context"
	"errors"
	"os"
	"os/signal"

	corelifecycle "github.com/bnb-chain/greenfield-storage-provider/core/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	DefaultStopTime = 30
)

var _ corelifecycle.Lifecycle = &GfSpBaseApp{}

// RegisterServices register services of an application
func (g *GfSpBaseApp) RegisterServices(services ...corelifecycle.Service) {
	registerModulus := GetRegisterModulusInstances()
	for _, modulus := range registerModulus {
		services = append(services, modulus)
	}
	g.services = append(g.services, services...)
}

// StartServices starts running services
func (g *GfSpBaseApp) StartServices(ctx context.Context) corelifecycle.Lifecycle {
	g.appCtx, g.appCancel = context.WithCancel(ctx)
	g.startServices(ctx)
	return g
}

func (g *GfSpBaseApp) startServices(ctx context.Context) {
	for i, service := range g.services {
		if err := service.Start(ctx); err != nil {
			log.Errorf("service %s starts error: %v", service.Name(), err)
			g.services = g.services[:i]
			g.appCancel()
			break
		} else {
			log.Infof("service %s starts successfully", service.Name())
		}
	}
}

// Signals registers monitor signals
func (g *GfSpBaseApp) Signals(sigs ...os.Signal) corelifecycle.Lifecycle {
	go g.signals(sigs...)
	return g
}

func (g *GfSpBaseApp) signals(sigs ...os.Signal) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, sigs...)
	for {
		select {
		case <-g.appCtx.Done():
			return
		case sig := <-sigCh:
			for _, j := range sigs {
				if j == sig {
					g.appCancel()
					return
				}
			}
		}
	}
}

// Wait blocks until context is done
func (g *GfSpBaseApp) Wait(ctx context.Context) {
	<-g.appCtx.Done()
	g.StopServices(ctx)
}

// StopServices stop services when context is done or timeout
func (g *GfSpBaseApp) StopServices(ctx context.Context) {
	gCtx, cancel := context.WithTimeout(context.Background(), DefaultStopTime)
	g.stopServices(ctx, cancel)

	<-gCtx.Done()
	if errors.Is(gCtx.Err(), context.Canceled) {
		log.Infow("services stop working", "stop service timeout", DefaultStopTime)
	} else if errors.Is(gCtx.Err(), context.DeadlineExceeded) {
		log.Error("timeout while stopping service, killing instance manually")
	}
}

func (g *GfSpBaseApp) stopServices(ctx context.Context, cancel context.CancelFunc) {
	for _, service := range g.services {
		if err := service.Stop(ctx); err != nil {
			log.Errorf("service %s stops failure: %v", service.Name(), err)
		} else {
			log.Infof("service %s stops successfully!", service.Name())
		}
	}
	cancel()
}

// Done check context is done
func (g *GfSpBaseApp) Done() <-chan struct{} {
	return g.appCtx.Done()
}
