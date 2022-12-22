package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Service provides abstract methods to control the lifecycle of a service
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Starter struct {
	context  context.Context
	cancel   context.CancelFunc
	services []Service
	fail     bool
}

func New() (*Starter, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Starter{
		context: ctx,
		cancel:  cancel,
	}, nil
}

func (s *Starter) RunServices(ctx context.Context, services ...Service) *Starter {
	if s.fail {
		return s
	}
	s.services = append(s.services, services...)
	for _, service := range services {
		select {
		case <-s.context.Done():
			log.Println("shutdown ...")
			s.fail = true
			return s
		default:
		}
		go s.start(ctx, service)
	}
	return s
}

func (s *Starter) start(ctx context.Context, service Service) {
	defer s.cancel()
	err := service.Start(ctx)
	if err != nil {
		log.Printf("service %s is done: %v", service.Name(), err)
	} else {
		log.Printf("service %s is done", service.Name())
	}
}

func (s *Starter) Signals(signals ...os.Signal) *Starter {
	if s.fail {
		return s
	}
	go s.signals(signals...)
	return s
}

func (s *Starter) signals(signals ...os.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(
		sigs,
		signals...,
	)
	for {
		select {
		case <-s.context.Done():
			return
		case event := <-sigs:
			for _, item := range signals {
				if item == event {
					s.cancel()
					return
				}
			}
		}
	}
}

func (s *Starter) Done() <-chan struct{} {
	return s.context.Done()
}

func (s *Starter) Wait(ctx context.Context) {
	<-s.context.Done()
	s.GracefulStop(ctx)
}

func (s *Starter) GracefulStop(conf context.Context) {
	log.Printf("Graceful shutdown ...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	go s.gracefulStop(conf, cancel)
	<-ctx.Done()
}

func (s *Starter) gracefulStop(conf context.Context, cancel context.CancelFunc) {
	wg := &sync.WaitGroup{}
	for _, service := range s.services {
		wg.Add(1)
		go s.stopService(conf, wg, service)
	}
	wg.Wait()
	cancel()
}

func (s *Starter) stopService(conf context.Context, wg *sync.WaitGroup, service Service) {
	defer wg.Done()
	err := service.Stop(conf)
	if err != nil {
		log.Printf("service %s is stopped: %v", service.Name(), err)
	} else {
		log.Printf("service %s is stopped", service.Name())
	}
}
