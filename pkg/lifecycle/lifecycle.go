package lifecycle

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// Service provides abstract methods to control the lifecycle of a service
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ServiceLifecycle is a lifecycle of one service
type ServiceLifecycle struct {
	innerCtx    context.Context
	innerCancel context.CancelFunc
	services    []Service
	failure     bool
	timeout     time.Duration
}

// NewService returns an initialized service lifecycle
func NewService(timeout time.Duration) *ServiceLifecycle {
	innerCtx, innerCancel := context.WithCancel(context.Background())
	return &ServiceLifecycle{
		innerCtx:    innerCtx,
		innerCancel: innerCancel,
		timeout:     timeout,
	}
}

// Init can be used to initialize components
func (s *ServiceLifecycle) Init(ctx context.Context, components ...Component) *ServiceLifecycle {
	if s.failure {
		return s
	}
	for _, c := range components {
		select {
		case <-s.innerCtx.Done():
			s.failure = true
			return s
		default:
		}
		if err := c.Init(ctx); err != nil {
			log.Panicf("Init %s error: %v", c.Name(), err)
			s.failure = true
		} else {
			log.Infof("Init %s successfully!", c.Name())
		}
	}
	return s
}

// StartServices starts running services
func (s *ServiceLifecycle) StartServices(ctx context.Context, services ...Service) *ServiceLifecycle {
	if s.failure {
		return s
	}
	s.services = append(s.services, services...)
	for _, service := range services {
		select {
		case <-s.innerCtx.Done():
			s.failure = true
			return s
		default:
		}
		go s.start(ctx, service)
	}
	return s
}

func (s *ServiceLifecycle) start(ctx context.Context, service Service) {
	defer s.innerCancel()
	if err := service.Start(ctx); err != nil {
		log.Panicf("Service %s starts error: %v", service.Name(), err)
	} else {
		log.Infof("Service %s starts successfully", service.Name())
	}
}

// Signals registers monitor signals
func (s *ServiceLifecycle) Signals(sigs ...os.Signal) *ServiceLifecycle {
	if s.failure {
		return s
	}
	go s.signals(sigs...)
	return s
}

func (s *ServiceLifecycle) signals(sigs ...os.Signal) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, sigs...)
	for {
		select {
		case <-s.innerCtx.Done():
			return
		case sig := <-sigCh:
			for _, j := range sigs {
				if j == sig {
					s.innerCancel()
					return
				}
			}
		}
	}
}

// Wait blocks until context is done
func (s *ServiceLifecycle) Wait(ctx context.Context) {
	<-s.innerCtx.Done()
	s.StopServices(ctx)
}

// StopServices can stop services when context is done or timeout
func (s *ServiceLifecycle) StopServices(ctx context.Context) {
	gCtx, cancel := context.WithTimeout(context.Background(), s.timeout)
	go s.stop(ctx, cancel)

	<-gCtx.Done()
	if errors.Is(gCtx.Err(), context.Canceled) {
		log.Infow("Services stop working", "service config timeout", s.timeout)
	} else if errors.Is(gCtx.Err(), context.DeadlineExceeded) {
		log.Panic("Timeout while stopping service, killing instance manually")
	}
}

func (s *ServiceLifecycle) stop(ctx context.Context, cancel context.CancelFunc) {
	var wg sync.WaitGroup
	for _, service := range s.services {
		wg.Add(1)
		go func(ctx context.Context, service Service) {
			defer wg.Done()
			if err := service.Stop(ctx); err != nil {
				log.Panicf("Service %s stops failure: %v", service.Name(), err)
			} else {
				log.Infof("Service %s stops successfully!", service.Name())
			}
		}(ctx, service)
	}
	wg.Wait()
	cancel()
}

// Done check context is done
func (s *ServiceLifecycle) Done() <-chan struct{} {
	return s.innerCtx.Done()
}
